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

func TestFilterNot(t *testing.T) {
	tokenizer := FilterTokenizer()
	input := "not ( City in ( 'Seattle', 'Atlanta' ) )"

	{
		expect := []*Token{
			&Token{Value: "not", Type: FilterTokenLogical},
			&Token{Value: "(", Type: FilterTokenOpenParen},
			&Token{Value: "City", Type: FilterTokenLiteral},
			&Token{Value: "in", Type: FilterTokenLogical},
			&Token{Value: "(", Type: FilterTokenOpenParen},
			&Token{Value: "'Seattle'", Type: FilterTokenString},
			&Token{Value: ",", Type: FilterTokenComma},
			&Token{Value: "'Atlanta'", Type: FilterTokenString},
			&Token{Value: ")", Type: FilterTokenCloseParen},
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
		var expect []expectedParseNode = []expectedParseNode{
			{"not", 0},
			{"in", 1},
			{"City", 2},
			{"(", 2},
			{"'Seattle'", 3},
			{"'Atlanta'", 3},
		}
		pos := 0
		err = CompareTree(tree, expect, &pos, 0)
		if err != nil {
			printTree(tree, 0)
			t.Errorf("Tree representation does not match expected value. error: %s", err.Error())
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

func TestValidFilterSyntax(t *testing.T) {
	queries := []string{
		// String functions
		"contains(CompanyName,'freds')",
		"endswith(CompanyName,'Futterkiste')",
		"startswith(CompanyName,'Alfr')",
		"length(CompanyName) eq 19",
		"indexof(CompanyName,'lfreds') eq 1",
		"substring(CompanyName,1) eq 'lfreds Futterkiste'",
		"substring(CompanyName,1,2) eq 'lf'", // substring with 3 arguments.
		"substringof('Alfreds', CompanyName) eq true",
		"tolower(CompanyName) eq 'alfreds futterkiste'",
		"toupper(CompanyName) eq 'ALFREDS FUTTERKISTE'",
		"trim(CompanyName) eq 'Alfreds Futterkiste'",
		"concat(concat(City,', '), Country) eq 'Berlin, Germany'",
		// Date and Time functions
		"year(BirthDate) eq 0",
		"month(BirthDate) eq 12",
		"day(StartTime) eq 8",
		"hour(StartTime) eq 1",
		"minute(StartTime) eq 0",
		"second(StartTime) eq 0",
		"date(StartTime) ne date(EndTime)",
		"totaloffsetminutes(StartTime) eq 60",
		"StartTime eq mindatetime()",
		"EndTime eq maxdatetime()",
		"time(StartTime) le StartOfDay",
		"time('2015-10-14T23:30:00.104+02:00') lt now()",
		"time(2015-10-14T23:30:00.104+02:00) lt now()",
		// Math functions
		"round(Freight) eq 32",
		"floor(Freight) eq 32",
		"ceiling(Freight) eq 33",
		"Rating mod 5 eq 0",
		"Price div 2 eq 3",
		// Type functions
		"isof(ShipCountry,Edm.String)",
		"isof(NorthwindModel.BigOrder)",
		"cast(ShipCountry,Edm.String)",
		// Parameter aliases
		// See http://docs.oasis-open.org/odata/odata/v4.0/errata03/os/complete/part1-protocol/odata-v4.0-errata03-os-part1-protocol-complete.html#_Toc453752288
		"Region eq @p1", // Aliases start with @
		// Geo functions
		"geo.distance(CurrentPosition,TargetPosition)",
		"geo.length(DirectRoute)",
		"geo.intersects(Position,TargetArea)",
		// Logical operators
		"Name eq 'Milk'",
		"Name ne 'Milk'",
		"Name gt 'Milk'",
		"Name ge 'Milk'",
		"Name lt 'Milk'",
		"Name le 'Milk'",
		"Name eq 'Milk' and Price lt 2.55",
		"not endswith(Name,'ilk')",
		"Name eq 'Milk' or Price lt 2.55",
		//"style has Sales.Pattern'Yellow'", // TODO
		// Arithmetic operators
		"Price add 2.45 eq 5.00",
		"Price sub 0.55 eq 2.00",
		"Price mul 2.0 eq 5.10",
		"Price div 2.55 eq 1",
		"Rating div 2 eq 2",
		"Rating mod 5 eq 0",
		// Grouping
		"(4 add 5) mod (4 sub 1) eq 0",
		"not (City eq 'Dallas') or Name in ('a', 'b', 'c') and not (State eq 'California')",
		// Nested functions
		"length(trim(CompanyName)) eq length(CompanyName)",
		"concat(concat(City, ', '), Country) eq 'Berlin, Germany'",
		// Various parenthesis combinations
		"City eq 'Dallas'",
		"City eq ('Dallas')",
		"'Dallas' eq City",
		"not (City eq 'Dallas')",
		"City in ('Dallas')",
		"(City in ('Dallas'))",
		"(City in ('Dallas', 'Houston'))",
		"not (City in ('Dallas'))",
		"not (City in ('Dallas', 'Houston'))",
		"not (((City eq 'Dallas')))",
	}
	for _, input := range queries {
		tokens, err := GlobalFilterTokenizer.Tokenize(input)
		if err != nil {
			t.Errorf("Error parsing query %s. Error: %s", input, err.Error())
			return
		}
		output, err := GlobalFilterParser.InfixToPostfix(tokens)
		if err != nil {
			t.Errorf("Error parsing query %s. Error: %s", input, err.Error())
			return
		}
		tree, err := GlobalFilterParser.PostfixToTree(output)
		if err != nil {
			t.Errorf("Error parsing query %s. Error: %s", input, err.Error())
			return
		} else if tree != nil {
			//printTree(tree, 0)
		}
	}
}

// The URLs below are not valid ODATA syntax, the parser should return an error.
func TestInvalidFilterSyntax(t *testing.T) {
	queries := []string{
		"City eq",                 // Missing operand
		"City eq (",               // Wrong operand
		"City eq )",               // Wrong operand
		"not [City eq 'Dallas']",  // Wrong delimiter
		"not (City eq )",          // Missing operand
		"not ((City eq 'Dallas'",  // Missing closing parenthesis
		"not (City eq 'Dallas'",   // Missing closing parenthesis
		"not (City eq 'Dallas'))", // Extraneous closing parenthesis
		"not City eq 'Dallas')",   // Missing open parenthesis
		"not (City eq 'Dallas') and Name eq 'Houston')",
		"LastName contains 'Smith'",    // Previously the godata library was not returning an error.
		"contains",                     // Function with missing parenthesis and arguments
		"contains()",                   // Function with missing arguments
		"contains LastName, 'Smith'",   // Missing parenthesis
		"contains(LastName)",           // Insufficent number of function arguments
		"contains(LastName, 'Smith'))", // Extraneous closing parenthesis
		"contains(LastName, 'Smith'",   // Missing closing parenthesis
		"contains LastName, 'Smith')",  // Missing open parenthesis
		//"City eq 'Dallas' 'Houston'",   // extraneous string value
		//"contains(Name, 'a', 'b', 'c', 'd')", // Too many function arguments
	}
	for _, input := range queries {
		tokens, err := GlobalFilterTokenizer.Tokenize(input)
		if err == nil {
			var output *tokenQueue
			output, err = GlobalFilterParser.InfixToPostfix(tokens)
			if err == nil {
				var tree *ParseNode
				tree, err = GlobalFilterParser.PostfixToTree(output)
				if err == nil {
					// The parser has incorrectly determined the syntax is valid.
					printTree(tree, 0)
				}
			}
		}
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

func TestSubstringFunction(t *testing.T) {
	// substring can take 2 or 3 arguments.
	{
		input := "substring(CompanyName,1) eq 'Foo'"
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
		var expect []expectedParseNode = []expectedParseNode{
			{"eq", 0},
			{"substring", 1},
			{"CompanyName", 2},
			{"1", 2},
			{"'Foo'", 1},
		}
		pos := 0
		err = CompareTree(tree, expect, &pos, 0)
		if err != nil {
			printTree(tree, 0)
			t.Errorf("Tree representation does not match expected value. error: %s", err.Error())
		}
	}
	{
		input := "substring(CompanyName,1,2) eq 'lf'"
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
		var expect []expectedParseNode = []expectedParseNode{
			{"eq", 0},
			{"substring", 1},
			{"CompanyName", 2},
			{"1", 2},
			{"2", 2},
			{"'lf'", 1},
		}
		pos := 0
		err = CompareTree(tree, expect, &pos, 0)
		if err != nil {
			printTree(tree, 0)
			t.Errorf("Tree representation does not match expected value. error: %s", err.Error())
		}
	}
}

func TestSubstringofFunction(t *testing.T) {
	// Previously, the parser was incorrectly interpreting the 'substring' function as the 'sub' operator.
	input := "substringof('Alfreds', CompanyName) eq true"
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
	var expect []expectedParseNode = []expectedParseNode{
		{"eq", 0},
		{"substringof", 1},
		{"'Alfreds'", 2},
		{"CompanyName", 2},
		{"true", 1},
	}
	pos := 0
	err = CompareTree(tree, expect, &pos, 0)
	if err != nil {
		printTree(tree, 0)
		t.Errorf("Tree representation does not match expected value. error: %s", err.Error())
	}
}

func TestGeoFunctions(t *testing.T) {
	// Previously, the parser was incorrectly interpreting the 'geo.xxx' functions as the 'ge' operator.
	input := "geo.distance(CurrentPosition,TargetPosition)"
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
	var expect []expectedParseNode = []expectedParseNode{
		{"geo.distance", 0},
		{"CurrentPosition", 1},
		{"TargetPosition", 1},
	}
	pos := 0
	err = CompareTree(tree, expect, &pos, 0)
	if err != nil {
		printTree(tree, 0)
		t.Errorf("Tree representation does not match expected value. error: %s", err.Error())
	}
}

func TestLambda1(t *testing.T) {
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

	var expect []expectedParseNode = []expectedParseNode{
		{"/", 0},
		{"Tags", 1},
		{"any", 1},
		{":", 2},
		{"var", 3},
		{"and", 3},
		{"eq", 4},
		{"/", 5},
		{"var", 6},
		{"Key", 6},
		{"'Site'", 5},
		{"eq", 4},
		{"/", 5},
		{"var", 6},
		{"Value", 6},
		{"'London'", 5},
	}
	pos := 0
	err = CompareTree(tree, expect, &pos, 0)
	if err != nil {
		printTree(tree, 0)
		t.Errorf("Tree representation does not match expected value. error: %s", err.Error())
	}
}
func TestLambda2(t *testing.T) {
	input := "Tags/any(var:var/Key eq 'Site' and var/Value eq 'London' or Price gt 1.0)"
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
	var expect []expectedParseNode = []expectedParseNode{
		{"/", 0},
		{"Tags", 1},
		{"any", 1},
		{":", 2},
		{"var", 3},
		{"or", 3},
		{"and", 4},
		{"eq", 5},
		{"/", 6},
		{"var", 7},
		{"Key", 7},
		{"'Site'", 6},
		{"eq", 5},
		{"/", 6},
		{"var", 7},
		{"Value", 7},
		{"'London'", 6},
		{"gt", 4},
		{"Price", 5},
		{"1.0", 5},
	}
	pos := 0
	err = CompareTree(tree, expect, &pos, 0)
	if err != nil {
		printTree(tree, 0)
		t.Errorf("Tree representation does not match expected value. error: %s", err.Error())
	}
}

func TestLambda3(t *testing.T) {
	input := "Tags/any(var:var/Key eq 'Site' and var/Value eq 'London' or Price gt 1.0 or contains(var/Value, 'Smith'))"
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
	var expect []expectedParseNode = []expectedParseNode{
		{"/", 0},
		{"Tags", 1},
		{"any", 1},
		{":", 2},
		{"var", 3},
		{"or", 3},
		{"or", 4},
		{"and", 5},
		{"eq", 6},
		{"/", 7},
		{"var", 8},
		{"Key", 8},
		{"'Site'", 7},
		{"eq", 6},
		{"/", 7},
		{"var", 8},
		{"Value", 8},
		{"'London'", 7},
		{"gt", 5},
		{"Price", 6},
		{"1.0", 6},
		{"contains", 4},
		{"/", 5},
		{"var", 6},
		{"Value", 6},
		{"'Smith'", 5},
	}
	pos := 0
	err = CompareTree(tree, expect, &pos, 0)
	if err != nil {
		printTree(tree, 0)
		t.Errorf("Tree representation does not match expected value. error: %s", err.Error())
	}
}

// CompareTree compares a tree representing a ODATA filter with the expected results.
// The expected values are a slice of nodes in breadth-first traversal.
func CompareTree(node *ParseNode, expect []expectedParseNode, pos *int, level int) error {
	if *pos >= len(expect) {
		return fmt.Errorf("Unexpected token. Got %s, expected no value", node.Token.Value)
	}
	if node.Token.Value != expect[*pos].Value {
		return fmt.Errorf("Unexpected token. Got %s -> %d, expected: %s -> %d", node.Token.Value, level, expect[*pos].Value, expect[*pos].Level)
	}
	if level != expect[*pos].Level {
		return fmt.Errorf("Unexpected level. Got %s -> %d, expected: %s -> %d", node.Token.Value, level, expect[*pos].Value, expect[*pos].Level)
	}
	for _, v := range node.Children {
		*pos++
		if err := CompareTree(v, expect, pos, level+1); err != nil {
			return err
		}
	}
	if level == 0 && *pos+1 != len(expect) {
		return fmt.Errorf("Expected number of tokens: %d, got %d", len(expect), *pos+1)
	}
	return nil
}

type expectedParseNode struct {
	Value string
	Level int
}
