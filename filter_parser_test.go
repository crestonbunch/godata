package godata

import (
	"fmt"
	"testing"
)

func TestFilterDateTime(t *testing.T) {
	tokenizer := FilterTokenizer()
	tokens := map[string]int{
		"2011-08-29T21:58Z":             FilterTokenDateTime,
		"2011-08-29T21:58:33Z":          FilterTokenDateTime,
		"2011-08-29T21:58:33.123Z":      FilterTokenDateTime,
		"2011-08-29T21:58+11:23":        FilterTokenDateTime,
		"2011-08-29T21:58:33+11:23":     FilterTokenDateTime,
		"2011-08-29T21:58:33.123+11:23": FilterTokenDateTime,
		"2011-08-29T21:58:33-11:23":     FilterTokenDateTime,
		"2011-08-29":                    FilterTokenDate,
		"21:58:33":                      FilterTokenTime,
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

func TestFilterDivby(t *testing.T) {
	{
		tokenizer := FilterTokenizer()
		input := "Price div 2 gt 3.5"
		expect := []*Token{
			&Token{Value: "Price", Type: FilterTokenLiteral},
			&Token{Value: "div", Type: FilterTokenOp},
			&Token{Value: "2", Type: FilterTokenInteger},
			&Token{Value: "gt", Type: FilterTokenLogical},
			&Token{Value: "3.5", Type: FilterTokenFloat},
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
	{
		tokenizer := FilterTokenizer()
		input := "Price divby 2 gt 3.5"
		expect := []*Token{
			&Token{Value: "Price", Type: FilterTokenLiteral},
			&Token{Value: "divby", Type: FilterTokenOp},
			&Token{Value: "2", Type: FilterTokenInteger},
			&Token{Value: "gt", Type: FilterTokenLogical},
			&Token{Value: "3.5", Type: FilterTokenFloat},
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

func TestFunction(t *testing.T) {
	tokenizer := FilterTokenizer()
	// The syntax for ODATA functions follows the inline parameter syntax. The function name must be followed
	// by an opening parenthesis, followed by a comma-separated list of parameters, followed by a closing parenthesis.
	// For example:
	// GET serviceRoot/Airports?$filter=contains(Location/Address, 'San Francisco')
	input := "contains(LastName, 'Smith') and FirstName eq 'John' and City eq 'Houston'"
	expect := []*Token{
		&Token{Value: "contains", Type: FilterTokenFunc},
		&Token{Value: "(", Type: FilterTokenOpenParen},
		&Token{Value: "LastName", Type: FilterTokenLiteral},
		&Token{Value: ",", Type: FilterTokenComma},
		&Token{Value: "'Smith'", Type: FilterTokenString},
		&Token{Value: ")", Type: FilterTokenCloseParen},
		&Token{Value: "and", Type: FilterTokenLogical},
		&Token{Value: "FirstName", Type: FilterTokenLiteral},
		&Token{Value: "eq", Type: FilterTokenLogical},
		&Token{Value: "'John'", Type: FilterTokenString},
		&Token{Value: "and", Type: FilterTokenLogical},
		&Token{Value: "City", Type: FilterTokenLiteral},
		&Token{Value: "eq", Type: FilterTokenLogical},
		&Token{Value: "'Houston'", Type: FilterTokenString},
	}
	{
		output, err := tokenizer.Tokenize(input)
		if err != nil {
			t.Error(err)
		}
		result, err := CompareTokens(expect, output)
		if !result {
			t.Error(err)
		}
	}
	{
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
		if tree.Token.Value != "and" {
			t.Errorf("Root is '%v', not 'and'", tree.Token.Value)
		}
		if len(tree.Children) != 2 {
			t.Errorf("Unexpected number of operators. Expected 2, got %d", len(tree.Children))
		}
		if tree.Children[0].Token.Value != "and" {
			t.Errorf("First child is '%v', not 'and'", tree.Children[0].Token.Value)
		}
		if len(tree.Children[0].Children) != 2 {
			t.Errorf("Unexpected number of operators. Expected 2, got %d", len(tree.Children))
		}
		if tree.Children[0].Children[0].Token.Value != "contains" {
			t.Errorf("First child is '%v', not 'contains'", tree.Children[0].Children[0].Token.Value)
		}
		if tree.Children[1].Token.Value != "eq" {
			t.Errorf("First child is '%v', not 'eq'", tree.Children[1].Token.Value)
		}
	}
}

func TestNestedFunction(t *testing.T) {
	tokenizer := FilterTokenizer()
	// Test ODATA syntax with nested function calls
	input := "contains(LastName, toupper('Smith')) or FirstName eq 'John'"
	expect := []*Token{
		&Token{Value: "contains", Type: FilterTokenFunc},
		&Token{Value: "(", Type: FilterTokenOpenParen},
		&Token{Value: "LastName", Type: FilterTokenLiteral},
		&Token{Value: ",", Type: FilterTokenComma},
		&Token{Value: "toupper", Type: FilterTokenFunc},
		&Token{Value: "(", Type: FilterTokenOpenParen},
		&Token{Value: "'Smith'", Type: FilterTokenString},
		&Token{Value: ")", Type: FilterTokenCloseParen},
		&Token{Value: ")", Type: FilterTokenCloseParen},
		&Token{Value: "or", Type: FilterTokenLogical},
		&Token{Value: "FirstName", Type: FilterTokenLiteral},
		&Token{Value: "eq", Type: FilterTokenLogical},
		&Token{Value: "'John'", Type: FilterTokenString},
	}
	{
		output, err := tokenizer.Tokenize(input)
		if err != nil {
			t.Error(err)
		}
		result, err := CompareTokens(expect, output)
		if !result {
			t.Error(err)
		}
	}
	{
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
		if tree.Token.Value != "or" {
			t.Errorf("Root is '%v', not 'or'", tree.Token.Value)
		}
		if len(tree.Children) != 2 {
			t.Errorf("Unexpected number of operators. Expected 2, got %d", len(tree.Children))
		}
		if tree.Children[0].Token.Value != "contains" {
			t.Errorf("First child is '%v', not 'contains'", tree.Children[0].Token.Value)
		}
		if len(tree.Children[0].Children) != 2 {
			t.Errorf("Unexpected number of nested children. Expected 2, got %d", len(tree.Children[0].Children))
		}
		if tree.Children[0].Children[1].Token.Value != "toupper" {
			t.Errorf("First child is '%v', not 'toupper'", tree.Children[0].Children[1].Token.Value)
		}
		if tree.Children[1].Token.Value != "eq" {
			t.Errorf("First child is '%v', not 'eq'", tree.Children[1].Token.Value)
		}
	}
}

func TestInvalidFunctionSyntax(t *testing.T) {
	// The URLs below are not valid ODATA syntax, it should return an error.
	// Note that previously the godata library was not returning an error.
	// GET serviceRoot/Airports?$filter=Location/Address contains 'San Francisco'
	{
		input := "LastName contains 'Smith'"
		tokens, err := GlobalFilterTokenizer.Tokenize(input)
		if err != nil {
			t.Error(err)
			return
		}
		_, err = GlobalFilterParser.InfixToPostfix(tokens)
		if err == nil {
			t.Errorf("The query '%s' is not valid ODATA syntax. The ODATA parser should return an error", input)
			return
		}
	}
	{
		// Extraneous closing parenthesis
		input := "contains(LastName, 'Smith'))"
		tokens, err := GlobalFilterTokenizer.Tokenize(input)
		if err != nil {
			t.Error(err)
			return
		}
		_, err = GlobalFilterParser.InfixToPostfix(tokens)
		if err == nil {
			t.Errorf("The query '%s' is not valid ODATA syntax. The ODATA parser should return an error", input)
			return
		}
	}
	{
		// Missing closing parenthesis
		input := "contains(LastName, 'Smith'"
		tokens, err := GlobalFilterTokenizer.Tokenize(input)
		if err != nil {
			t.Error(err)
			return
		}
		_, err = GlobalFilterParser.InfixToPostfix(tokens)
		if err == nil {
			t.Errorf("The query '%s' is not valid ODATA syntax. The ODATA parser should return an error", input)
			return
		}
	}
	{
		// Missing open parenthesis
		input := "contains LastName, 'Smith')"
		tokens, err := GlobalFilterTokenizer.Tokenize(input)
		if err != nil {
			t.Error(err)
			return
		}
		_, err = GlobalFilterParser.InfixToPostfix(tokens)
		if err == nil {
			t.Errorf("The query '%s' is not valid ODATA syntax. The ODATA parser should return an error", input)
			return
		}
	}
}

// See http://docs.oasis-open.org/odata/odata/v4.01/csprd02/part1-protocol/odata-v4.01-csprd02-part1-protocol.html#_Toc486263411
// Test 'in', which is the 'Is a member of' operator.

func TestFilterIn(t *testing.T) {
	tokenizer := FilterTokenizer()
	input := "contains(LastName, 'Smith') and Site in ('London', 'Paris', 'San Francisco', 'Dallas') and FirstName eq 'John'"
	expect := []*Token{
		&Token{Value: "contains", Type: FilterTokenFunc},
		&Token{Value: "(", Type: FilterTokenOpenParen},
		&Token{Value: "LastName", Type: FilterTokenLiteral},
		&Token{Value: ",", Type: FilterTokenComma},
		&Token{Value: "'Smith'", Type: FilterTokenString},
		&Token{Value: ")", Type: FilterTokenCloseParen},
		&Token{Value: "and", Type: FilterTokenLogical},
		&Token{Value: "Site", Type: FilterTokenLiteral},
		&Token{Value: "in", Type: FilterTokenLogical},
		&Token{Value: "(", Type: FilterTokenOpenParen},
		&Token{Value: "'London'", Type: FilterTokenString},
		&Token{Value: ",", Type: FilterTokenComma},
		&Token{Value: "'Paris'", Type: FilterTokenString},
		&Token{Value: ",", Type: FilterTokenComma},
		&Token{Value: "'San Francisco'", Type: FilterTokenString},
		&Token{Value: ",", Type: FilterTokenComma},
		&Token{Value: "'Dallas'", Type: FilterTokenString},
		&Token{Value: ")", Type: FilterTokenCloseParen},
		&Token{Value: "and", Type: FilterTokenLogical},
		&Token{Value: "FirstName", Type: FilterTokenLiteral},
		&Token{Value: "eq", Type: FilterTokenLogical},
		&Token{Value: "'John'", Type: FilterTokenString},
	}
	{
		output, err := tokenizer.Tokenize(input)
		if err != nil {
			t.Error(err)
		}
		result, err := CompareTokens(expect, output)
		if !result {
			t.Error(err)
		}
	}
	{
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

		/*
		  The expected tree is:
		  and        6
		   and        6
		     contains   8
		       LastName   20
		       'Smith'    15
		     in         6
		       Site       20
		       (          0
		         'London'   15
		         'Paris'    15
		         'San Francisco' 15
		         'Dallas'   15
		   eq         6
		     FirstName  20
		     'John'     15

		*/
		if tree.Token.Value != "and" {
			t.Errorf("Root is '%v', not 'and'", tree.Token.Value)
		}
		if len(tree.Children) != 2 {
			t.Errorf("Unexpected number of operators. Expected 2, got %d", len(tree.Children))
		}
		if tree.Children[0].Token.Value != "and" {
			t.Errorf("First child is '%v', not 'and'", tree.Children[0].Token.Value)
		}
		if len(tree.Children[0].Children) != 2 {
			t.Errorf("Unexpected number of operators. Expected 2, got %d", len(tree.Children))
		}
		if tree.Children[0].Children[0].Token.Value != "contains" {
			t.Errorf("First child is '%v', not 'contains'", tree.Children[0].Children[0].Token.Value)
		}
		if tree.Children[0].Children[1].Token.Value != "in" {
			t.Errorf("First child is '%v', not 'in'", tree.Children[0].Children[1].Token.Value)
		}
		if len(tree.Children[0].Children[1].Children) != 2 {
			t.Errorf("Unexpected number of operands for the 'in' operator. Expected 2, got %d",
				len(tree.Children[0].Children[1].Children))
		}
		if tree.Children[0].Children[1].Children[0].Token.Value != "Site" {
			t.Errorf("Unexpected operand for the 'in' operator. Expected 'Site', got %s",
				len(tree.Children[0].Children[1].Children[0].Token.Value))
		}
		if tree.Children[0].Children[1].Children[1].Token.Value != "(" {
			t.Errorf("Unexpected operand for the 'in' operator. Expected '(', got %s",
				len(tree.Children[0].Children[1].Children[1].Token.Value))
		}
		if len(tree.Children[0].Children[1].Children[1].Children) != 4 {
			t.Errorf("Unexpected number of operands for the 'in' operator. Expected 4, got %d",
				len(tree.Children[0].Children[1].Children[1].Token.Value))
		}
		if tree.Children[1].Token.Value != "eq" {
			t.Errorf("First child is '%v', not 'eq'", tree.Children[1].Token.Value)
		}
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
		return false, fmt.Errorf("Different lengths. %d != %d", len(a), len(b))
	}
	for i, _ := range a {
		if a[i].Type != b[i].Type {
			return false, fmt.Errorf("Different token types at index %d. Type: %v != %v. Value: %v",
				i, a[i].Type, b[i].Type, a[i].Value)
		}
		if a[i].Value != b[i].Value {
			return false, fmt.Errorf("Different token values at index %d. Value: %v != %v",
				i, a[i].Value, b[i].Value)
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
