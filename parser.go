package godata

import (
	"regexp"
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
	Pattern string
	Re      *regexp.Regexp
	Token   int
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
	matcher := &TokenMatcher{pattern, rxp, token}
	t.TokenMatchers = append(t.TokenMatchers, matcher)
}

func (t *Tokenizer) Ignore(pattern string, token int) {
	rxp := regexp.MustCompile(pattern)
	matcher := &TokenMatcher{pattern, rxp, token}
	t.IgnoreMatchers = append(t.IgnoreMatchers, matcher)
}

func (t *Tokenizer) TokenizeBytes(target []byte) ([]*Token, error) {
	result := make([]*Token, 0)
	match := true // false when no match is found
	for len(target) > 0 && match {
		match = false
		for _, m := range t.TokenMatchers {
			token := m.Re.Find(target)
			if len(token) > 0 {
				parsed := Token{Value: string(token), Type: m.Token}
				result = append(result, &parsed)
				target = target[len(token):] // remove the token from the input
				match = true
				break
			}
		}
		for _, m := range t.IgnoreMatchers {
			token := m.Re.Find(target)
			if len(token) > 0 {
				match = true
				target = target[len(token):] // remove the token from the input
				break
			}
		}
	}

	if len(target) > 0 && !match {
		return result, BadRequestError("No matching token for " + string(target))
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
}

type Function struct {
	Token string
	// The number of parameters this function accepts
	Params int
}

type ParseNode struct {
	Token    *Token
	Parent   *ParseNode
	Children []*ParseNode
}

func EmptyParser() *Parser {
	return &Parser{make(map[string]*Operator, 0), make(map[string]*Function)}
}

// Add an operator to the language. Provide the token, a precedence, and
// whether the operator is left, right, or not associative.
func (p *Parser) DefineOperator(token string, operands, assoc, precedence int) {
	p.Operators[token] = &Operator{token, assoc, operands, precedence}
}

// Add a function to the language
func (p *Parser) DefineFunction(token string, params int) {
	p.Functions[token] = &Function{token, params}
}

// Parse the input string of tokens using the given definitions of operators
// and functions. (Everything else is assumed to be a literal.) Uses the
// Shunting-Yard algorithm.
func (p *Parser) InfixToPostfix(tokens []*Token) (*tokenQueue, error) {
	queue := tokenQueue{}
	stack := tokenStack{}

	for len(tokens) > 0 {
		token := tokens[0]
		tokens = tokens[1:]

		if _, ok := p.Functions[token.Value]; ok {
			// push functions onto the stack
			stack.Push(token)
		} else if token.Value == "," {
			// function parameter separator, pop off stack until we see a "("
			for stack.Peek().Value != "(" || stack.Empty() {
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
			// push open parens onto the stack
			stack.Push(token)
		} else if token.Value == ")" {
			// if we find a close paren, pop things off the stack
			for !stack.Empty() && stack.Peek().Value != "(" {
				queue.Enqueue(stack.Pop())
			}
			// there was an error parsing
			if stack.Empty() {
				return nil, BadRequestError("Parse error. Mismatched parenthesis.")
			}
			// pop off open paren
			stack.Pop()
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
			for i := 0; i < f.Params; i++ {
				// prepend children so they get added in the right order
				node.Children = append([]*ParseNode{stack.Pop()}, node.Children...)
			}
			stack.Push(node)
		} else if _, ok := p.Operators[stack.Peek().Token.Value]; ok {
			// if the top of the stack is an operator
			node := stack.Pop()
			o := p.Operators[node.Token.Value]
			// pop off operands
			for i := 0; i < o.Operands; i++ {
				// prepend children so they get added in the right order
				node.Children = append([]*ParseNode{stack.Pop()}, node.Children...)
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
