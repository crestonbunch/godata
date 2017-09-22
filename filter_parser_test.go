package godata

import (
	"errors"
	"fmt"
	"strconv"
	"testing"
)

func TestFilterDateTime(t *testing.T) {
	tokenizer := FilterTokenizer()
	tokens := map[string]int{
		"2011-08-29T21:58Z": FilterTokenDateTime,
		"2011-08-29T21:58:33Z": FilterTokenDateTime,
		"2011-08-29T21:58:33.123Z": FilterTokenDateTime,
		"2011-08-29T21:58+11:23": FilterTokenDateTime,
		"2011-08-29T21:58:33+11:23": FilterTokenDateTime,
		"2011-08-29T21:58:33.123+11:23": FilterTokenDateTime,
		"2011-08-29T21:58:33-11:23": FilterTokenDateTime,
		"2011-08-29": FilterTokenDate,
		"21:58:33": FilterTokenTime,
	}
	for tokenValue, tokenType := range tokens {
		input := "CreateTime gt" + tokenValue
		expect := []*Token{
			&Token{Value: "CreateTime", Type: FilterTokenLiteral},
			&Token{Value: "gt", Type: FilterTokenLogical},
			&Token{Value: tokenValue, Type: tokenType},
		}
		output, err := tokenizer.Tokenize(input)
		if err != nil {
			t.Error(err)
		}

		result, err := CompareTokens(expect, output)
		if !result {
			t.Error(err)
		}
	}
}

func TestFilterAny(t *testing.T) {
	tokenizer := FilterTokenizer()
	input := "Tags/any(d:d/Key eq 'Site' and d/Value lt 10)"
	expect := []*Token{
		&Token{Value: "Tags", Type: FilterTokenLiteral},
		&Token{Value: "/", Type: FilterTokenNav},
		&Token{Value: "any", Type: FilterTokenLambda},
		&Token{Value: "(", Type: FilterTokenOpenParen},
		&Token{Value: "d", Type: FilterTokenLiteral},
		&Token{Value: ":", Type: FilterTokenColon},
		&Token{Value: "d", Type: FilterTokenLiteral},
		&Token{Value: "/", Type: FilterTokenNav},
		&Token{Value: "Key", Type: FilterTokenLiteral},
		&Token{Value: "eq", Type: FilterTokenLogical},
		&Token{Value: "'Site'", Type: FilterTokenString},
		&Token{Value: "and", Type: FilterTokenLogical},
		&Token{Value: "d", Type: FilterTokenLiteral},
		&Token{Value: "/", Type: FilterTokenNav},
		&Token{Value: "Value", Type: FilterTokenLiteral},
		&Token{Value: "lt", Type: FilterTokenLogical},
		&Token{Value: "10", Type: FilterTokenInteger},
		&Token{Value: ")", Type: FilterTokenCloseParen},
	}
	output, err := tokenizer.Tokenize(input)
	if err != nil {
		t.Error(err)
	}

	result, err := CompareTokens(expect, output)
	if !result {
		t.Error(err)
	}
}

func TestFilterAll(t *testing.T) {
	tokenizer := FilterTokenizer()
	input := "Tags/all(d:d/Key eq 'Site')"
	expect := []*Token{
		&Token{Value: "Tags", Type: FilterTokenLiteral},
		&Token{Value: "/", Type: FilterTokenNav},
		&Token{Value: "all", Type: FilterTokenLambda},
		&Token{Value: "(", Type: FilterTokenOpenParen},
		&Token{Value: "d", Type: FilterTokenLiteral},
		&Token{Value: ":", Type: FilterTokenColon},
		&Token{Value: "d", Type: FilterTokenLiteral},
		&Token{Value: "/", Type: FilterTokenNav},
		&Token{Value: "Key", Type: FilterTokenLiteral},
		&Token{Value: "eq", Type: FilterTokenLogical},
		&Token{Value: "'Site'", Type: FilterTokenString},
		&Token{Value: ")", Type: FilterTokenCloseParen},
	}
	output, err := tokenizer.Tokenize(input)
	if err != nil {
		t.Error(err)
	}

	result, err := CompareTokens(expect, output)
	if !result {
		t.Error(err)
	}
}

func TestFilterTokenizer(t *testing.T) {

	tokenizer := FilterTokenizer()
	input := "Name eq 'Milk' and Price lt 2.55"
	expect := []*Token{
		&Token{Value: "Name", Type: FilterTokenLiteral},
		&Token{Value: "eq", Type: FilterTokenLogical},
		&Token{Value: "'Milk'", Type: FilterTokenString},
		&Token{Value: "and", Type: FilterTokenLogical},
		&Token{Value: "Price", Type: FilterTokenLiteral},
		&Token{Value: "lt", Type: FilterTokenLogical},
		&Token{Value: "2.55", Type: FilterTokenFloat},
	}
	output, err := tokenizer.Tokenize(input)
	if err != nil {
		t.Error(err)
	}

	result, err := CompareTokens(expect, output)
	if !result {
		t.Error(err)
	}
}

func TestFilterTokenizerFunc(t *testing.T) {

	tokenizer := FilterTokenizer()
	input := "not endswith(Name,'ilk')"
	expect := []*Token{
		&Token{Value: "not", Type: FilterTokenLogical},
		&Token{Value: "endswith", Type: FilterTokenFunc},
		&Token{Value: "(", Type: FilterTokenOpenParen},
		&Token{Value: "Name", Type: FilterTokenLiteral},
		&Token{Value: ",", Type: FilterTokenComma},
		&Token{Value: "'ilk'", Type: FilterTokenString},
		&Token{Value: ")", Type: FilterTokenCloseParen},
	}
	output, err := tokenizer.Tokenize(input)
	if err != nil {
		t.Error(err)
	}

	result, err := CompareTokens(expect, output)
	if !result {
		t.Error(err)
	}
}

func BenchmarkFilterTokenizer(b *testing.B) {
	t := FilterTokenizer()
	for i := 0; i < b.N; i++ {
		input := "Name eq 'Milk' and Price lt 2.55"
		t.Tokenize(input)
	}
}

// Check if two slices of tokens are the same.
func CompareTokens(a, b []*Token) (bool, error) {
	if len(a) != len(b) {
		return false, errors.New("Different lengths")
	}
	for i, _ := range a {
		if a[i].Value != b[i].Value || a[i].Type != b[i].Type {
			return false, errors.New("Different at index " + strconv.Itoa(i) + " " +
				a[i].Value + " != " + b[i].Value + " or types are different.")
		}
	}
	return true, nil
}

func TestFilterParserTree(t *testing.T) {

	input := "not (A eq B)"

	tokens, err := GlobalFilterTokenizer.Tokenize(input)
	if err != nil {
		t.Error(err)
		return
	}
	output, err := GlobalFilterParser.InfixToPostfix(tokens)

	if err != nil {
		t.Error(err)
		return
	}

	tree, err := GlobalFilterParser.PostfixToTree(output)

	if err != nil {
		t.Error(err)
		return
	}

	if tree.Token.Value != "not" {
		t.Error("Root is '" + tree.Token.Value + "' not 'not'")
	}
	if tree.Children[0].Token.Value != "eq" {
		t.Error("First child is '" + tree.Children[1].Token.Value + "' not 'eq'")
	}

}

func printTree(n *ParseNode, level int) {
	indent := ""
	for i := 0; i < level; i++ {
		indent += "  "
	}
	fmt.Printf("%s %-10s %-10d\n", indent, n.Token.Value, n.Token.Type)
	for _, v := range n.Children {
		printTree(v, level+1)
	}
}

func TestNestedPath(t *testing.T) {
	input := "Address/City eq 'Redmond'"
	tokens, err := GlobalFilterTokenizer.Tokenize(input)
	if err != nil {
		t.Error(err)
		return
	}
	output, err := GlobalFilterParser.InfixToPostfix(tokens)
	if err != nil {
		t.Error(err)
		return
	}

	tree, err := GlobalFilterParser.PostfixToTree(output)
	if err != nil {
		t.Error(err)
		return
	}
	//printTree(tree, 0)
	if tree.Token.Value != "eq" {
		t.Error("Root is '" + tree.Token.Value + "' not 'eq'")
	}
	if tree.Children[0].Token.Value != "/" {
		t.Error("First child is \"" + tree.Children[0].Token.Value + "\", not '/'")
	}
	if tree.Children[1].Token.Value != "'Redmond'" {
		t.Error("First child is \"" + tree.Children[1].Token.Value + "\", not 'Redmond'")
	}
}

func TestLambda(t *testing.T) {
	input := "Tags/any(var:var/Key eq 'Site' and var/Value eq 'London')"
	tokens, err := GlobalFilterTokenizer.Tokenize(input)
	if err != nil {
		t.Error(err)
		return
	}
	output, err := GlobalFilterParser.InfixToPostfix(tokens)
	if err != nil {
		t.Error(err)
		return
	}

	tree, err := GlobalFilterParser.PostfixToTree(output)
	if err != nil {
		t.Error(err)
		return
	}
	//printTree(tree, 0)

	if tree.Token.Value != "/" {
		t.Error("Root is '" + tree.Token.Value + "' not '/'")
	}
	if tree.Children[0].Token.Value != "Tags" {
		t.Error("First child is '" + tree.Children[0].Token.Value + "' not 'Tags'")
	}
	if tree.Children[1].Token.Value != "any" {
		t.Error("First child is '" + tree.Children[1].Token.Value + "' not 'any'")
	}
	if tree.Children[1].Children[0].Token.Value != ":" {
		t.Error("First child is '" + tree.Children[1].Children[0].Token.Value + "' not ':'")
	}
}
