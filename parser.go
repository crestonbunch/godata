package godata

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

const (
	OpAssociationLeft int = iota
	OpAssociationRight
	OpAssociationNone
)

const (
	NodeTypeLiteral int = iota
	NodeTypeOp
	NodeTypeFunc
)

type Tokenizer struct {
	TokenMatchers  []*TokenMatcher
	IgnoreMatchers []*TokenMatcher
}

type TokenMatcher struct {
	Pattern         string         // The regular expression matching a ODATA query token, such as literal value, operator or function
	Re              *regexp.Regexp // The compiled regex
	Token           int            // The token identifier
	CaseInsentitive bool           // Regex is case-insensitive
}

type Token struct {
	Value string
	Type  int
	// Holds information about the semantic meaning of this token taken from the
	// context of the GoDataService.
	SemanticType      int
	SemanticReference interface{}
}

func (t *Tokenizer) Add(pattern string, token int) {
	rxp := regexp.MustCompile(pattern)
	matcher := &TokenMatcher{pattern, rxp, token, strings.Contains(pattern, "(?i)")}
	t.TokenMatchers = append(t.TokenMatchers, matcher)
}

func (t *Tokenizer) Ignore(pattern string, token int) {
	rxp := regexp.MustCompile(pattern)
	matcher := &TokenMatcher{pattern, rxp, token, strings.Contains(pattern, "(?i)")}
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
			if idx := m.Re.SubexpIndex("token"); idx > 0 {
				token = tokens[idx]
			} else {
				token = tokens[0]
			}
			target = target[len(token):] // remove the token from the input
			if !ignore {
				if m.CaseInsentitive {
					// In ODATA 4.0.1, operators and functions are case insensitive.
					parsed = Token{Value: strings.ToLower(string(token)), Type: m.Token}
				} else {
					parsed = Token{Value: string(token), Type: m.Token}
				}
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
	// Specifies whether the right-side operand is single value (by default) or multi-value.
	// For example, "City in ('San Jose', 'Chicago', 'Dallas')"
	// For left-side operand, the parser algorithm needs to change to support backtracking.
	MultiValueOperand bool
}

type Function struct {
	Token string
	// The number of parameters this function accepts
	Params []int
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
	return &Parser{make(map[string]*Operator, 0), make(map[string]*Function)}
}

// Add an operator to the language. Provide the token, the expected number of arguments,
// whether the operator is left, right, or not associative, and a precedence.
// The operandsCardinality specifies whether the operands are single-value or multi-value, which must
// be a comma-separated list enclosed in parenthesis.
func (p *Parser) DefineOperator(token string, operands, assoc, precedence int, multiValueOperand bool) {
	p.Operators[token] = &Operator{token, assoc, operands, precedence, multiValueOperand}
}

// Add a function to the language
func (p *Parser) DefineFunction(token string, params []int) {
	sort.Sort(sort.Reverse(sort.IntSlice(params)))
	p.Functions[token] = &Function{token, params}
}

// isGroupingOperator returns true if the specified token uses
// parenthesis as multi-value grouping delimiters.
func (p *Parser) isGroupingOperator(stack *tokenStack, token *Token) bool {
	if !stack.Empty() {
		if o1, ok := p.Operators[stack.Peek().Value]; ok {
			if o1.MultiValueOperand {
				// The '(' token is used as a grouping operator.
				return true
			}
		}
	}
	return false
}

// IsLiteral returns true if the specified token is considered a literal for this parser.
func (p *Parser) IsLiteral(token *Token) bool {
	if token.Value == "," || token.Value == "(" || token.Value == ")" {
		return false
	}
	if _, ok := p.Functions[token.Value]; ok {
		return false
	}
	if _, ok := p.Operators[token.Value]; ok {
		return false
	}
	return true
}

// InfixToPostfix parses the input string of tokens using the given definitions of operators
// and functions. (Everything else is assumed to be a literal.) Uses the
// Shunting-Yard algorithm.
func (p *Parser) InfixToPostfix(tokens []*Token) (*tokenQueue, error) {
	queue := tokenQueue{} // output queue in postfix
	stack := tokenStack{} // Operator stack

	prevIsLiteral := false
	for i := range tokens {
		isLiteral := p.IsLiteral(tokens[i])
		if isLiteral && prevIsLiteral {
			// Cannot have two consecutive literal values.
			return nil, BadRequestError(
				fmt.Sprintf("Request cannot include two consecutive literal values '%v' and '%v'",
					tokens[i-1].Value, tokens[i].Value))
		}
		prevIsLiteral = isLiteral
	}
	for len(tokens) > 0 {
		token := tokens[0]
		tokens = tokens[1:]
		if _, ok := p.Functions[token.Value]; ok {
			if len(tokens) == 0 || tokens[0].Value != "(" {
				// A function token must be followed by open parenthesis token.
				return nil, BadRequestError(fmt.Sprintf("Function '%s' must be followed by '('", token.Value))
			}
			// push functions onto the stack
			stack.Push(token)
		} else if token.Value == "," {
			// function parameter separator, pop off stack until we see a "("
			for !stack.Empty() && stack.Peek().Value != "(" {
				queue.Enqueue(stack.Pop())
			}
			// there was an error parsing
			if stack.Empty() {
				return nil, BadRequestError("Parse error")
			}
		} else if o1, ok := p.Operators[token.Value]; ok {
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
			stack.Push(token)
		} else if token.Value == "(" {
			// In OData, the parenthesis tokens can be used:
			// - To set explicit precedence order, such as "(a + 2) x b"
			//   These precedence tokens are removed while parsing the OData query.
			// - As a grouping operator for multi-value sets, such as "City in ('San Jose', 'Chicago', 'Dallas')"
			//   The grouping tokens are retained while parsing the OData query.
			if p.isGroupingOperator(&stack, token) {
				// The '(' token is used as a grouping operator.
				queue.Enqueue(token)
			} else {
				// The '(' token is used to set explicit precedence order. Do not enqueue
			}
			// push open parens onto the stack
			stack.Push(token)
		} else if token.Value == ")" {
			// if we find a close paren, pop things off the stack
			multiValueOperand := false
			for !stack.Empty() {
				if stack.Peek().Value == "(" {
					// Pop the stack so we can see the second element in the stack.
					t := stack.Pop()
					if !stack.Empty() && p.isGroupingOperator(&stack, stack.Peek()) {
						// The ')' token is used as the closing grouping operator.
						queue.Enqueue(token)       // The ')' closing parenthesis
						queue.Enqueue(stack.Pop()) // The operator that takes a multi-value operand.
						multiValueOperand = true
						// Continue dequeuing until we find a precedence delimiter
					} else {
						// In this case, '(' is a precedence token, we stop popping the operators
						// from the stack, and we need to re-add the last token back on the stack,
						// which was done to inspect whether the second operator in the stack takes
						// a multi-value operand.
						stack.Push(t)
						break
					}
				} else {
					queue.Enqueue(stack.Pop())
				}
			}
			// there was an error parsing
			if stack.Empty() {
				if !multiValueOperand {
					return nil, BadRequestError("Parse error. Mismatched parenthesis.")
				}
			} else {
				if !multiValueOperand {
					// pop off open paren
					stack.Pop()
				}
			}
			// if next token is a function, move it to the queue
			if !stack.Empty() {
				if _, ok := p.Functions[stack.Peek().Value]; ok {
					queue.Enqueue(stack.Pop())
				}
			}
		} else {
			// Token is a literal -- put it in the queue
			queue.Enqueue(token)
		}
	}

	// pop off the remaining operators onto the queue
	for !stack.Empty() {
		if stack.Peek().Value == "(" || stack.Peek().Value == ")" {
			return nil, BadRequestError("parse error. Mismatched parenthesis.")
		}
		queue.Enqueue(stack.Pop())
	}
	return &queue, nil
}

// Convert a Postfix token queue to a parse tree
func (p *Parser) PostfixToTree(queue *tokenQueue) (*ParseNode, error) {
	stack := &nodeStack{}
	currNode := &ParseNode{}

	t := queue.Head
	for t != nil {
		t = t.Next
	}
	for !queue.Empty() {
		// push the token onto the stack as a tree node
		currNode = &ParseNode{queue.Dequeue(), nil, make([]*ParseNode, 0)}
		stack.Push(currNode)

		if _, ok := p.Functions[stack.Peek().Token.Value]; ok {
			// if the top of the stack is a function
			node := stack.Pop()
			f := p.Functions[node.Token.Value]
			// pop off function parameters
			idx := 0
			argCount := f.Params[idx]
			for i := 0; i < argCount; i++ {
				if stack.Empty() {
					// Some functions, e.g. substring, can take a variable number of arguments.
					foundMatch := false
					for idx < (len(f.Params) - 1) {
						idx++
						argCount = f.Params[idx]
						if i == argCount {
							foundMatch = true
							break
						}
					}
					if !foundMatch {
						return nil, fmt.Errorf("Insufficient number of operands for function '%s'", node.Token.Value)
					} else {
						break
					}
				}
				// prepend children so they get added in the right order
				c := stack.Pop()
				c.Parent = node
				node.Children = append([]*ParseNode{c}, node.Children...)
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
		} else if stack.Peek().Token.Value == ")" {
			// This is a multi-value operand, such as the 'in' operator.
			// In this case, the parenthesis is not used as a precedence delimiter.
			// Pop the close parenthesis.
			stack.Pop()
			var children []*ParseNode
			for !stack.Empty() && stack.Peek().Token.Value != "(" {
				// prepend children so they get added in the right order
				children = append([]*ParseNode{stack.Pop()}, children...)
			}
			// Pop the open parenthesis
			node := stack.Pop()
			for _, v := range children {
				v.Parent = node
			}
			node.Children = children
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
	Token *Token
	Prev  *tokenStackNode
}

func (s *tokenStack) Push(t *Token) {
	node := tokenStackNode{t, s.Head}
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
	result := ""
	node := q.Head
	for node != nil {
		result += node.Token.Value
		node = node.Next
	}
	return result
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
	output := ""
	currNode := s.Head
	for currNode != nil {
		output += " " + currNode.ParseNode.Token.Value
		currNode = currNode.Prev
	}
	return output
}
