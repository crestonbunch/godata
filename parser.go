package godata

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const (
	OpAssociationLeft int = iota
	OpAssociationRight
	OpAssociationNone
)

// TokenTypeArgCount is used to specify the number of arguments of a function or listExpr
// This is used to handle variadic functions and listExpr.
const TokenTypeArgCount int = -1

// TokenTypeListExpr represents a parent node for a variadic listExpr.
// "list"
//   "item1"
//   "item2"
//   ...
const TokenTypeListExpr int = -2

const TokenListExpr = "list"

// TokenComma is the default separator for function arguments.
const (
	TokenComma      = ","
	TokenOpenParen  = "("
	TokenCloseParen = ")"
)

type Tokenizer struct {
	TokenMatchers  []*TokenMatcher
	IgnoreMatchers []*TokenMatcher
}

type TokenMatcher struct {
	Pattern         string                 // The regular expression matching a ODATA query token, such as literal value, operator or function
	Re              *regexp.Regexp         // The compiled regex
	Token           int                    // The token identifier
	CaseInsensitive bool                   // Regex is case-insensitive
	Subst           func(in string) string // A function that substitutes the raw input token with another representation. By default it is the identity.
}

type Token struct {
	Value string
	Type  int
	// Holds information about the semantic meaning of this token taken from the
	// context of the GoDataService.
	SemanticType      SemanticType
	SemanticReference interface{}
}

func (t *Tokenizer) Add(pattern string, token int) {
	t.AddWithSubstituteFunc(pattern, token, func(in string) string { return in })
}

func (t *Tokenizer) AddWithSubstituteFunc(pattern string, token int, subst func(string) string) {
	rxp := regexp.MustCompile(pattern)
	matcher := &TokenMatcher{
		Pattern:         pattern,
		Re:              rxp,
		Token:           token,
		CaseInsensitive: strings.Contains(pattern, "(?i)"),
		Subst:           subst,
	}
	t.TokenMatchers = append(t.TokenMatchers, matcher)
}

func (t *Tokenizer) Ignore(pattern string, token int) {
	rxp := regexp.MustCompile(pattern)
	matcher := &TokenMatcher{
		Pattern:         pattern,
		Re:              rxp,
		Token:           token,
		CaseInsensitive: strings.Contains(pattern, "(?i)"),
		Subst:           func(in string) string { return in },
	}
	t.IgnoreMatchers = append(t.IgnoreMatchers, matcher)
}

func (t *Tokenizer) TokenizeBytes(target []byte) ([]*Token, error) {
	result := make([]*Token, 0)
	match := true // false when no match is found
	for len(target) > 0 && match {
		match = false
		ignore := false
		var tokens [][]byte
		var m *TokenMatcher
		for _, m = range t.TokenMatchers {
			tokens = m.Re.FindSubmatch(target)
			if len(tokens) > 0 {
				match = true
				break
			}
		}
		if len(tokens) == 0 {
			for _, m = range t.IgnoreMatchers {
				tokens = m.Re.FindSubmatch(target)
				if len(tokens) > 0 {
					ignore = true
					break
				}
			}
		}
		if len(tokens) > 0 {
			match = true
			var parsed Token
			var token []byte
			// If the regex includes a named group and the name of that group is "token"
			// then the value of the token is set to the subgroup. Other characters are
			// not consumed by the tokenization process.
			// For example, the regex:
			//    ^(?P<token>(eq|ne|gt|ge|lt|le|and|or|not|has|in))\\s
			// has a group named 'token' and the group is followed by a mandatory space character.
			// If the input data is `Name eq 'Bob'`, the token is correctly set to 'eq' and
			// the space after eq is not consumed, because the space character itself is supposed
			// to be the next token.
			//
			// If Token.Value needs to be a sub-regex but the entire token needs to be consumed,
			// use 'subtoken'
			//    ^(duration)?'(?P<subtoken>[0-9])'
			l := 0
			if idx := m.Re.SubexpIndex("token"); idx > 0 {
				token = tokens[idx]
				l = len(token)
			} else if idx := m.Re.SubexpIndex("subtoken"); idx > 0 {
				token = tokens[idx]
				l = len(tokens[0])
			} else {
				token = tokens[0]
				l = len(token)
			}
			target = target[l:] // remove the token from the input
			if !ignore {
				var v string
				if m.CaseInsensitive {
					// In ODATA 4.0.1, operators and functions are case insensitive.
					v = strings.ToLower(string(token))
				} else {
					v = string(token)
				}
				parsed = Token{Value: m.Subst(v), Type: m.Token}
				result = append(result, &parsed)
			}
		}
	}

	if len(target) > 0 && !match {
		return result, BadRequestError("No matching token for " + string(target))
	}
	if len(result) < 1 {
		return result, BadRequestError("Empty query parameter")
	}
	return result, nil
}

func (t *Tokenizer) Tokenize(target string) ([]*Token, error) {
	return t.TokenizeBytes([]byte(target))
}

type Parser struct {
	// Map from string inputs to operator types
	Operators map[string]*Operator
	// Map from string inputs to function types
	Functions map[string]*Function
}

type Operator struct {
	Token string
	// Whether the operator is left/right/or not associative
	Association int
	// The number of operands this operator operates on
	Operands int
	// Rank of precedence
	Precedence int
	// Determine if the operands should be interpreted as a ListExpr or parenExpr according
	// to ODATA ABNF grammar.
	// This is only used when a listExpr has zero or one items.
	// When a listExpr has 2 or more items, there is no ambiguity between listExpr and parenExpr.
	// For example:
	//    2 + (3) ==> the right operand is a parenExpr
	//    City IN ('Seattle', 'Atlanta') ==> the right operand is unambiguously a listExpr
	//    City IN ('Seattle') ==> the right operand should be a listExpr
	PreferListExpr bool
}

func (o *Operator) SetPreferListExpr(v bool) {
	o.PreferListExpr = v
}

type Function struct {
	Token  string // The function token
	Params []int  // The number of parameters this function accepts
}

type ParseNode struct {
	Token    *Token
	Parent   *ParseNode
	Children []*ParseNode
}

func (p *ParseNode) String() string {
	var sb strings.Builder
	var treePrinter func(n *ParseNode, sb *strings.Builder, level int)

	treePrinter = func(n *ParseNode, s *strings.Builder, level int) {
		if n == nil || n.Token == nil {
			s.WriteRune('\n')
			return
		}
		s.WriteString(fmt.Sprintf("[%2d]", n.Token.Type))
		s.WriteString(strings.Repeat("  ", level))
		s.WriteString(n.Token.Value)
		s.WriteRune('\n')
		for _, v := range n.Children {
			treePrinter(v, s, level+1)
		}
	}
	treePrinter(p, &sb, 0)
	return sb.String()
}

func EmptyParser() *Parser {
	return &Parser{
		Operators: make(map[string]*Operator, 0),
		Functions: make(map[string]*Function),
	}
}

// DefineOperator adds an operator to the language. Provide the token, the expected number of arguments,
// whether the operator is left, right, or not associative, and a precedence.
func (p *Parser) DefineOperator(token string, operands, assoc, precedence int) *Operator {
	op := &Operator{
		Token:       token,
		Association: assoc,
		Operands:    operands,
		Precedence:  precedence,
	}
	p.Operators[token] = op
	return op
}

// DefineFunction adds a function to the language
// params is the number of parameters this function accepts
func (p *Parser) DefineFunction(token string, params []int) {
	sort.Sort(sort.Reverse(sort.IntSlice(params)))
	p.Functions[token] = &Function{token, params}
}

// InfixToPostfix parses the input string of tokens using the given definitions of operators
// and functions.
// Everything else is assumed to be a literal.
// Uses the Shunting-Yard algorithm.
//
// Infix notation for variadic functions and operators: f ( a, b, c, d )
// Postfix notation with wall notation:                 | a b c d f
// Postfix notation with count notation:                a b c d 4 f
//
func (p *Parser) InfixToPostfix(tokens []*Token) (*tokenQueue, error) {
	queue := tokenQueue{} // output queue in postfix
	stack := tokenStack{} // Operator stack

	isCurrentTokenLiteral := false
	for len(tokens) > 0 {
		token := tokens[0]
		tokens = tokens[1:]
		if _, ok := p.Functions[token.Value]; ok {
			isCurrentTokenLiteral = false
			if len(tokens) == 0 || tokens[0].Value != TokenOpenParen {
				// A function token must be followed by open parenthesis token.
				return nil, BadRequestError(fmt.Sprintf("Function '%s' must be followed by '('", token.Value))
			}
			stack.incrementListArgCount()
			// push functions onto the stack
			stack.Push(token)
		} else if o1, ok := p.Operators[token.Value]; ok {
			isCurrentTokenLiteral = false
			// push operators onto stack according to precedence
			if !stack.Empty() {
				for o2, ok := p.Operators[stack.Peek().Value]; ok &&
					(o1.Association == OpAssociationLeft && o1.Precedence <= o2.Precedence) ||
					(o1.Association == OpAssociationRight && o1.Precedence < o2.Precedence); {
					queue.Enqueue(stack.Pop())

					if stack.Empty() {
						break
					}
					o2, ok = p.Operators[stack.Peek().Value]
				}
			}
			if o1.Operands == 1 { // not, -
				stack.incrementListArgCount()
			}
			stack.Push(token)
		} else {
			switch token.Value {
			case TokenOpenParen:
				isCurrentTokenLiteral = false
				// In OData, the parenthesis tokens can be used:
				// - As a parenExpr to set explicit precedence order, such as "(a + 2) x b"
				//   These precedence tokens are removed while parsing the OData query.
				// - As a listExpr for multi-value sets, such as "City in ('San Jose', 'Chicago', 'Dallas')"
				//   The list tokens are retained while parsing the OData query.
				//   ABNF grammar:
				//   listExpr  = OPEN BWS commonExpr BWS *( COMMA BWS commonExpr BWS ) CLOSE
				stack.incrementListArgCount()
				// Push open parens onto the stack
				stack.Push(token)
			case TokenCloseParen:
				isCurrentTokenLiteral = false
				// if we find a close paren, pop things off the stack
				for !stack.Empty() {
					if stack.Peek().Value == TokenOpenParen {
						break
					} else {
						queue.Enqueue(stack.Pop())
					}
				}
				if stack.Empty() {
					// there was an error parsing
					return nil, BadRequestError("Parse error. Mismatched parenthesis.")
				}
				// Get the argument count associated with the open paren.
				// Example: (a, b, c) is a listExpr with three arguments.
				argCount := stack.getArgCount()
				// pop off open paren
				stack.Pop()

				// Determine if the parenthesis delimiters are:
				// - A listExpr, possibly an empty list or single element.
				//   Note a listExpr may be on the left-side or right-side of operators,
				//   or it may be a list of function arguments.
				// - A parenExpr, which is used as a precedence delimiter.
				//
				// (1, 2, 3) is a listExpr, there is no ambiguity.
				// (1) matches both listExpr or parenExpr.
				// parenExpr takes precedence over listExpr.
				//
				// For example:
				//   1 IN (1, 2)  ==> parenthesis is used to create a list of two elements.
				//   (1) + (2)    ==> parenthesis is a precedence delimiter, i.e. parenExpr.
				var isFunc, isListExpr bool
				if !stack.Empty() {
					_, isFunc = p.Functions[stack.Peek().Value]
				}
				switch {
				case isFunc:
					isListExpr = true
				case argCount <= 1:
					// When a listExpr contains a single item, it is ambiguous whether it is a listExpr or parenExpr.
					if !stack.Empty() {
						if o1, ok := p.Operators[stack.Peek().Value]; ok {
							if o1.PreferListExpr {
								// The expression is the right operand of an operator that has a preference for listExpr vs parenExpr.
								isListExpr = true
							}
						}
					}
					if !isListExpr && len(tokens) > 0 {
						if o1, ok := p.Operators[tokens[0].Value]; ok {
							// The expression is the left operand of an operator that has a preference for listExpr vs parenExpr.
							if o1.PreferListExpr {
								isListExpr = true
							}
						}
					}
				case argCount > 1:
					isListExpr = true
				}
				if isListExpr {
					// The open parenthesis was a delimiter for a listExpr.
					// Add a token indicating the number of arguments in the list.
					queue.Enqueue(&Token{
						Value: strconv.Itoa(argCount),
						Type:  TokenTypeArgCount,
					})
					// Enqueue a 'list' token if we are processing a ListExpr.
					if !isFunc {
						queue.Enqueue(&Token{
							Value: TokenListExpr,
							Type:  TokenTypeListExpr,
						})
					}
				}
				// if next token is a function or multi-value operator, move it to the queue
				if isFunc {
					queue.Enqueue(stack.Pop())
				}
			case TokenComma:
				// Function argument separator (",")
				isCurrentTokenLiteral = false
				// Comma may be used as:
				// 1. Separator of function parameters,
				// 2. Separator for listExpr such as "City IN ('Seattle', 'San Francisco')"
				//
				// Pop off stack until we see a TokenOpenParen
				for !stack.Empty() && stack.Peek().Value != TokenOpenParen {
					// This happens when the previous function argument is an expression composed
					// of multiple tokens, as opposed to a single token. For example:
					//     max(sin( 5 mul pi ) add 3, sin( 5 ))
					queue.Enqueue(stack.Pop())
				}
				if stack.Empty() {
					// there was an error parsing. The top of the stack must be open parenthesis
					return nil, BadRequestError("Parse error")
				}
				if stack.Peek().Value != TokenOpenParen {
					panic("unexpected token")
				}
			default:
				if isCurrentTokenLiteral {
					// In most cases, it is invalid to have two consecutive literal values.
					// TODO: The exception is for a positionLiteral:
					//    positionLiteral  = doubleValue SP doubleValue  ; longitude, then latitude
					return nil, BadRequestError("Request cannot include two consecutive literal values")
				}
				isCurrentTokenLiteral = true
				// Token is a literal -- put it in the queue
				stack.incrementListArgCount()
				queue.Enqueue(token)
			}
		}
	}

	// pop off the remaining operators onto the queue
	for !stack.Empty() {
		if stack.Peek().Value == TokenOpenParen || stack.Peek().Value == TokenCloseParen {
			return nil, BadRequestError("parse error. Mismatched parenthesis.")
		}
		queue.Enqueue(stack.Pop())
	}
	return &queue, nil
}

// PostfixToTree converts a Postfix token queue to a parse tree
func (p *Parser) PostfixToTree(queue *tokenQueue) (*ParseNode, error) {
	stack := &nodeStack{}
	currNode := &ParseNode{}

	t := queue.Head
	for t != nil {
		t = t.Next
	}
	processVariadicArgs := func(parent *ParseNode) (int, error) {
		if stack.Empty() {
			return 0, fmt.Errorf("No argCount token found. '%s'", parent.Token.Value)
		}
		n := stack.Pop()
		if n.Token.Type != TokenTypeArgCount {
			return 0, fmt.Errorf("No argCount token found. '%s'", parent.Token.Value)
		}
		argCount, err := strconv.Atoi(n.Token.Value)
		if err != nil {
			return 0, err
		}
		for i := 0; i < argCount; i++ {
			if stack.Empty() {
				return 0, fmt.Errorf("Missing argument found. '%s'", parent.Token.Value)
			}
			c := stack.Pop()
			// Attach the operand to its parent node which represents the function/operator
			c.Parent = parent
			// prepend children so they get added in the right order
			parent.Children = append([]*ParseNode{c}, parent.Children...)
		}
		return argCount, nil
	}
	for !queue.Empty() {
		// push the token onto the stack as a tree node
		currToken := queue.Dequeue()
		currNode = &ParseNode{currToken, nil, make([]*ParseNode, 0)}
		stack.Push(currNode)

		if _, ok := p.Functions[stack.Peek().Token.Value]; ok {
			// if the top of the stack is a function
			node := stack.Pop()
			f := p.Functions[node.Token.Value]
			// Pop off function parameters
			// The first token is the number of function arguments, which is useful
			// when parsing variadic functions.
			// Some functions, e.g. substring, can take a variable number of arguments.
			if argCount, err := processVariadicArgs(node); err != nil {
				return nil, err
			} else {
				foundMatch := false
				for _, expectedArgCount := range f.Params {
					if argCount == expectedArgCount {
						foundMatch = true
						break
					}
				}
				if !foundMatch {
					return nil, fmt.Errorf("Invalid number of arguments for function '%s'. Got %d argument",
						node.Token.Value, argCount)
				}
			}
			stack.Push(node)
		} else if _, ok := p.Operators[stack.Peek().Token.Value]; ok {
			// if the top of the stack is an operator
			node := stack.Pop()
			o := p.Operators[node.Token.Value]
			// pop off operands
			for i := 0; i < o.Operands; i++ {
				if stack.Empty() {
					return nil, fmt.Errorf("Insufficient number of operands for operator '%s'", node.Token.Value)
				}
				// prepend children so they get added in the right order
				c := stack.Pop()
				c.Parent = node
				node.Children = append([]*ParseNode{c}, node.Children...)
			}
			stack.Push(node)
		} else if stack.Peek().Token.Type == TokenTypeListExpr {
			// ListExpr: List of items
			node := stack.Pop()
			if _, err := processVariadicArgs(node); err != nil {
				return nil, err
			}
			stack.Push(node)

		}
	}
	return currNode, nil
}

type tokenStack struct {
	Head *tokenStackNode
	Size int
}

type tokenStackNode struct {
	Token      *Token          // The token value.
	Prev       *tokenStackNode // The previous node in the stack.
	tokenCount int             // The number of arguments in a listExpr.
}

func (s *tokenStack) Push(t *Token) {
	node := tokenStackNode{Token: t, Prev: s.Head}
	//fmt.Println("Pushed:", t.Value, "->", s.String())
	s.Head = &node
	s.Size++
}

func (s *tokenStack) Pop() *Token {
	node := s.Head
	s.Head = node.Prev
	s.Size--
	//fmt.Println("Popped:", node.Token.Value, "<-", s.String())
	return node.Token
}

func (s *tokenStack) Peek() *Token {
	return s.Head.Token
}

func (s *tokenStack) Empty() bool {
	return s.Head == nil
}

func (s *tokenStack) incrementListArgCount() {
	if !s.Empty() && s.Head.Token.Value == TokenOpenParen {
		s.Head.tokenCount++
	}
}

func (s *tokenStack) getArgCount() int {
	return s.Head.tokenCount
}

func (s *tokenStack) String() string {
	output := ""
	currNode := s.Head
	for currNode != nil {
		output += " " + currNode.Token.Value
		currNode = currNode.Prev
	}
	return output
}

type tokenQueue struct {
	Head *tokenQueueNode
	Tail *tokenQueueNode
}

type tokenQueueNode struct {
	Token *Token
	Prev  *tokenQueueNode
	Next  *tokenQueueNode
}

// Enqueue adds the specified token at the tail of the queue.
func (q *tokenQueue) Enqueue(t *Token) {
	node := tokenQueueNode{t, q.Tail, nil}
	//fmt.Println(t.Value)

	if q.Tail == nil {
		q.Head = &node
	} else {
		q.Tail.Next = &node
	}

	q.Tail = &node
}

// Dequeue removes the token at the head of the queue and returns the token.
func (q *tokenQueue) Dequeue() *Token {
	node := q.Head
	if node.Next != nil {
		node.Next.Prev = nil
	}
	q.Head = node.Next
	if q.Head == nil {
		q.Tail = nil
	}
	return node.Token
}

func (q *tokenQueue) Empty() bool {
	return q.Head == nil && q.Tail == nil
}

func (q *tokenQueue) String() string {
	var sb strings.Builder
	node := q.Head
	for node != nil {
		sb.WriteString(node.Token.Value)
		node = node.Next
		if node != nil {
			sb.WriteRune(' ')
		}
	}
	return sb.String()
}

func (q *tokenQueue) GetValue() string {
	var sb strings.Builder
	node := q.Head
	for node != nil {
		sb.WriteString(node.Token.Value)
		node = node.Next
	}
	return sb.String()
}

type nodeStack struct {
	Head *nodeStackNode
}

type nodeStackNode struct {
	ParseNode *ParseNode
	Prev      *nodeStackNode
}

func (s *nodeStack) Push(n *ParseNode) {
	node := nodeStackNode{n, s.Head}
	s.Head = &node
}

func (s *nodeStack) Pop() *ParseNode {
	node := s.Head
	s.Head = node.Prev
	return node.ParseNode
}

func (s *nodeStack) Peek() *ParseNode {
	return s.Head.ParseNode
}

func (s *nodeStack) Empty() bool {
	return s.Head == nil
}

func (s *nodeStack) String() string {
	var sb strings.Builder
	currNode := s.Head
	for currNode != nil {
		sb.WriteRune(' ')
		sb.WriteString(currNode.ParseNode.Token.Value)
		currNode = currNode.Prev
	}
	return sb.String()
}
