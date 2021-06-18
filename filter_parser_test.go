package godata

import (
	"fmt"
	"strings"
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
		// Previously, the unit test had no space character after 'gt'
		// E.g. 'CreateTime gt2011-08-29T21:58Z' was considered valid.
		// However the ABNF notation for ODATA logical operators is:
		//   gtExpr = RWS "gt" RWS commonExpr
		//   RWS = 1*( SP / HTAB / "%20" / "%09" )  ; "required" whitespace
		//
		// See http://docs.oasis-open.org/odata/odata/v4.01/csprd03/abnf/odata-abnf-construction-rules.txt
		input := "CreateTime gt " + tokenValue
		expect := []*Token{
			&Token{Value: "CreateTime", Type: FilterTokenLiteral},
			&Token{Value: "gt", Type: FilterTokenLogical},
			&Token{Value: tokenValue, Type: tokenType},
		}
		output, err := tokenizer.Tokenize(input)
		if err != nil {
			t.Errorf("Failed to tokenize input %s. Error: %v", input, err)
		}

		result, err := CompareTokens(expect, output)
		if !result {
			var a []string
			for _, t := range output {
				a = append(a, t.Value)
			}

			t.Errorf("Unexpected tokens for input '%s'. Tokens: %s Error: %v", input, strings.Join(a, ", "), err)
		}
	}
}

func TestFilterAnyArrayOfObjects(t *testing.T) {
	tokenizer := FilterTokenizer()
	input := "Tags/any(d:d/Key eq 'Site' and d/Value lt 10)"
	expect := []*Token{
		&Token{Value: "Tags", Type: FilterTokenLiteral},
		&Token{Value: "/", Type: FilterTokenOp},
		&Token{Value: "any", Type: FilterTokenLambda},
		&Token{Value: "(", Type: FilterTokenOpenParen},
		&Token{Value: "d", Type: FilterTokenLiteral},
		&Token{Value: ",", Type: FilterTokenColon}, // ':' is replaced by ',' which is the function argument separator.
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

func TestFilterAnyArrayOfPrimitiveTypes(t *testing.T) {
	tokenizer := FilterTokenizer()
	input := "Tags/any(d:d eq 'Site')"
	{
		expect := []*Token{
			&Token{Value: "Tags", Type: FilterTokenLiteral},
			&Token{Value: "/", Type: FilterTokenOp},
			&Token{Value: "any", Type: FilterTokenLambda},
			&Token{Value: "(", Type: FilterTokenOpenParen},
			&Token{Value: "d", Type: FilterTokenLiteral},
			&Token{Value: ",", Type: FilterTokenColon},
			&Token{Value: "d", Type: FilterTokenLiteral},
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
	q, err := ParseFilterString(input)
	if err != nil {
		t.Errorf("Error parsing query %s. Error: %s", input, err.Error())
		return
	}
	var expect []expectedParseNode = []expectedParseNode{
		{"/", 0},
		{"Tags", 1},
		{"any", 1},
		{"d", 2},
		{"eq", 2},
		{"d", 3},
		{"'Site'", 3},
	}
	pos := 0
	err = CompareTree(q.Tree, expect, &pos, 0)
	if err != nil {
		fmt.Printf("Got tree:\n%v\n", q.Tree.String())
		t.Errorf("Tree representation does not match expected value. error: %s", err.Error())
	}
}

// geographyPolygon   = geographyPrefix SQUOTE fullPolygonLiteral SQUOTE
// geographyPrefix = "geography"
// fullPolygonLiteral = sridLiteral polygonLiteral
// sridLiteral      = "SRID" EQ 1*5DIGIT SEMI
// polygonLiteral     = "Polygon" polygonData
// polygonData        = OPEN ringLiteral *( COMMA ringLiteral ) CLOSE
// positionLiteral  = doubleValue SP doubleValue  ; longitude, then latitude
/*
func TestFilterGeographyPolygon(t *testing.T) {
	input := "geo.intersects(location, geography'SRID=0;Polygon(-122.031577 47.578581, -122.031577 47.678581, -122.131577 47.678581, -122.031577 47.578581)')"
	q, err := ParseFilterString(input)
	if err != nil {
		t.Errorf("Error parsing query %s. Error: %s", input, err.Error())
		return
	}
	var expect []expectedParseNode = []expectedParseNode{
		{"geo.intersects", 0},
		{"location", 1},
		{"geography'SRID=0;Polygon(-122.031577 47.578581, -122.031577 47.678581, -122.131577 47.678581, -122.031577 47.578581)'", 1},
	}
	pos := 0
	err = CompareTree(q.Tree, expect, &pos, 0)
	if err != nil {
		fmt.Printf("Got tree:\n%v\n", q.Tree.String())
		t.Errorf("Tree representation does not match expected value. error: %s", err.Error())
	}
}
*/

// TestFilterAnyGeography matches documents where any of the geo coordinates in the locations field is within the given polygon.
/*
func TestFilterAnyGeography(t *testing.T) {
	input := "locations/any(loc: geo.intersects(loc, geography'Polygon((-122.031577 47.578581, -122.031577 47.678581, -122.131577 47.678581, -122.031577 47.578581))'))"
	q, err := ParseFilterString(input)
	if err != nil {
		t.Errorf("Error parsing query %s. Error: %s", input, err.Error())
		return
	}
	var expect []expectedParseNode = []expectedParseNode{
		{"/", 0},
		{"Tags", 1},
		{"any", 1},
		{"d", 2},
		{"or", 2},
		{"or", 3},
		{"or", 4},
		{"eq", 5},
		{"d", 6},
		{"'Site'", 6},
		{"eq", 5},
		{"'Environment'", 6},
		{"/", 6},
		{"d", 7},
		{"Key", 7},
		{"eq", 4},
		{"/", 5},
		{"/", 6},
		{"d", 7},
		{"d", 7},
		{"d", 6},
		{"123456", 5},
		{"eq", 3},
		{"concat", 4},
		{"/", 5},
		{"d", 6},
		{"FirstName", 6},
		{"/", 5},
		{"d", 6},
		{"LastName", 6},
		{"/", 4},
		{"$it", 5},
		{"FullName", 5},
	}
	pos := 0
	err = CompareTree(q.Tree, expect, &pos, 0)
	if err != nil {
		fmt.Printf("Got tree:\n%v\n", q.Tree.String())
		t.Errorf("Tree representation does not match expected value. error: %s", err.Error())
	}
}
*/

func TestFilterAnyMixedQuery(t *testing.T) {
	/*
		{
			"Tags": [
				"Site",
				{ "Key": "Environment" },
				{ "d" : { "d": 123456 }},
				{ "FirstName" : "Bob", "LastName": "Smith"}
			],
			"FullName": "BobSmith"
		}
	*/
	// The argument of a lambda operator is a case-sensitive lambda variable name followed by a colon (:) and a Boolean expression that
	// uses the lambda variable name to refer to properties of members of the collection identified by the navigation path.
	// If the name chosen for the lambda variable matches a property name of the current resource referenced by the resource path, the lambda variable takes precedence.
	// Clients can prefix properties of the current resource referenced by the resource path with $it.
	// Other path expressions in the Boolean expression neither prefixed with the lambda variable nor $it are evaluated in the scope of
	// the collection instances at the origin of the navigation path prepended to the lambda operator.
	input := "Tags/any(d:d eq 'Site' or 'Environment' eq d/Key or d/d/d eq 123456 or concat(d/FirstName, d/LastName) eq $it/FullName)"
	q, err := ParseFilterString(input)
	if err != nil {
		t.Errorf("Error parsing query %s. Error: %s", input, err.Error())
		return
	}
	var expect []expectedParseNode = []expectedParseNode{
		{"/", 0},
		{"Tags", 1},
		{"any", 1},
		{"d", 2},
		{"or", 2},
		{"or", 3},
		{"or", 4},
		{"eq", 5},
		{"d", 6},
		{"'Site'", 6},
		{"eq", 5},
		{"'Environment'", 6},
		{"/", 6},
		{"d", 7},
		{"Key", 7},
		{"eq", 4},
		{"/", 5},
		{"/", 6},
		{"d", 7},
		{"d", 7},
		{"d", 6},
		{"123456", 5},
		{"eq", 3},
		{"concat", 4},
		{"/", 5},
		{"d", 6},
		{"FirstName", 6},
		{"/", 5},
		{"d", 6},
		{"LastName", 6},
		{"/", 4},
		{"$it", 5},
		{"FullName", 5},
	}
	pos := 0
	err = CompareTree(q.Tree, expect, &pos, 0)
	if err != nil {
		fmt.Printf("Got tree:\n%v\n", q.Tree.String())
		t.Errorf("Tree representation does not match expected value. error: %s", err.Error())
	}
}

func TestFilterGuid(t *testing.T) {
	tokenizer := FilterTokenizer()
	input := "GuidValue eq 01234567-89ab-cdef-0123-456789abcdef"

	expect := []*Token{
		&Token{Value: "GuidValue", Type: FilterTokenLiteral},
		&Token{Value: "eq", Type: FilterTokenLogical},
		&Token{Value: "01234567-89ab-cdef-0123-456789abcdef", Type: FilterTokenGuid},
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

func TestFilterDurationWithType(t *testing.T) {
	tokenizer := FilterTokenizer()
	input := "Task eq duration'P12DT23H59M59.999999999999S'"

	expect := []*Token{
		&Token{Value: "Task", Type: FilterTokenLiteral},
		&Token{Value: "eq", Type: FilterTokenLogical},
		// Note the duration token is extracted.
		&Token{Value: "P12DT23H59M59.999999999999S", Type: FilterTokenDuration},
	}
	output, err := tokenizer.Tokenize(input)
	if err != nil {
		t.Error(err)
	}
	result, err := CompareTokens(expect, output)
	if !result {
		printTokens(output)
		t.Error(err)
	}
}

func TestFilterDurationWithoutType(t *testing.T) {
	tokenizer := FilterTokenizer()
	input := "Task eq 'P12DT23H59M59.999999999999S'"

	expect := []*Token{
		&Token{Value: "Task", Type: FilterTokenLiteral},
		&Token{Value: "eq", Type: FilterTokenLogical},
		&Token{Value: "P12DT23H59M59.999999999999S", Type: FilterTokenDuration},
	}
	output, err := tokenizer.Tokenize(input)
	if err != nil {
		t.Error(err)
	}
	result, err := CompareTokens(expect, output)
	if !result {
		printTokens(output)
		t.Error(err)
	}
}

func TestFilterAnyWithNoArgs(t *testing.T) {
	tokenizer := FilterTokenizer()
	input := "Tags/any()"
	{
		expect := []*Token{
			&Token{Value: "Tags", Type: FilterTokenLiteral},
			&Token{Value: "/", Type: FilterTokenOp},
			&Token{Value: "any", Type: FilterTokenLambda},
			&Token{Value: "(", Type: FilterTokenOpenParen},
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
	q, err := ParseFilterString(input)
	if err != nil {
		t.Errorf("Error parsing query %s. Error: %s", input, err.Error())
		return
	}
	var expect []expectedParseNode = []expectedParseNode{
		{"/", 0},
		{"Tags", 1},
		{"any", 1},
	}
	pos := 0
	err = CompareTree(q.Tree, expect, &pos, 0)
	if err != nil {
		fmt.Printf("Got tree:\n%v\n", q.Tree.String())
		t.Errorf("Tree representation does not match expected value. error: %s", err.Error())
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

func TestFilterNotBooleanProperty(t *testing.T) {
	tokenizer := FilterTokenizer()
	input := "not Enabled"
	{
		expect := []*Token{
			&Token{Value: "not", Type: FilterTokenLogical},
			&Token{Value: "Enabled", Type: FilterTokenLiteral},
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
	q, err := ParseFilterString(input)
	if err != nil {
		t.Errorf("Error parsing query %s. Error: %s", input, err.Error())
		return
	}
	var expect []expectedParseNode = []expectedParseNode{
		{"not", 0},
		{"Enabled", 1},
	}
	pos := 0
	err = CompareTree(q.Tree, expect, &pos, 0)
	if err != nil {
		fmt.Printf("Got tree:\n%v\n", q.Tree.String())
		t.Errorf("Tree representation does not match expected value. error: %s", err.Error())
	}

}

// Note: according to ODATA ABNF notation, there must be a space between not and open parenthesis.
// http://docs.oasis-open.org/odata/odata/v4.01/csprd03/abnf/odata-abnf-construction-rules.txt
func TestFilterNotWithNoSpace(t *testing.T) {
	tokenizer := FilterTokenizer()
	input := "not(City eq 'Seattle')"
	{
		expect := []*Token{
			&Token{Value: "not", Type: FilterTokenLogical},
			&Token{Value: "(", Type: FilterTokenOpenParen},
			&Token{Value: "City", Type: FilterTokenLiteral},
			&Token{Value: "eq", Type: FilterTokenLogical},
			&Token{Value: "'Seattle'", Type: FilterTokenString},
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

	q, err := ParseFilterString(input)
	if err != nil {
		t.Errorf("Error parsing query %s. Error: %s", input, err.Error())
		return
	}
	var expect []expectedParseNode = []expectedParseNode{
		{"not", 0},
		{"eq", 1},
		{"City", 2},
		{"'Seattle'", 2},
	}
	pos := 0
	err = CompareTree(q.Tree, expect, &pos, 0)
	if err != nil {
		fmt.Printf("Got tree:\n%v\n", q.Tree.String())
		t.Errorf("Tree representation does not match expected value. error: %s", err.Error())
	}
}

// TestFilterInOperator tests the "IN" operator with a comma-separated list of values.
func TestFilterInOperator(t *testing.T) {
	tokenizer := FilterTokenizer()
	input := "City in ( 'Seattle', 'Atlanta', 'Paris' )"

	expect := []*Token{
		&Token{Value: "City", Type: FilterTokenLiteral},
		&Token{Value: "in", Type: FilterTokenLogical},
		&Token{Value: "(", Type: FilterTokenOpenParen},
		&Token{Value: "'Seattle'", Type: FilterTokenString},
		&Token{Value: ",", Type: FilterTokenComma},
		&Token{Value: "'Atlanta'", Type: FilterTokenString},
		&Token{Value: ",", Type: FilterTokenComma},
		&Token{Value: "'Paris'", Type: FilterTokenString},
		&Token{Value: ")", Type: FilterTokenCloseParen},
	}
	tokens, err := tokenizer.Tokenize(input)
	if err != nil {
		t.Error(err)
	}
	result, err := CompareTokens(expect, tokens)
	if !result {
		t.Error(err)
	}
	var postfix *tokenQueue
	postfix, err = GlobalFilterParser.InfixToPostfix(tokens)
	if err != nil {
		t.Error(err)
	}
	expect = []*Token{
		&Token{Value: "City", Type: FilterTokenLiteral},
		&Token{Value: "'Seattle'", Type: FilterTokenString},
		&Token{Value: "'Atlanta'", Type: FilterTokenString},
		&Token{Value: "'Paris'", Type: FilterTokenString},
		&Token{Value: "3", Type: TokenTypeArgCount},
		&Token{Value: TokenListExpr, Type: TokenTypeListExpr},
		&Token{Value: "in", Type: FilterTokenLogical},
	}
	result, err = CompareQueue(expect, postfix)
	if !result {
		t.Error(err)
	}

	tree, err := GlobalFilterParser.PostfixToTree(postfix)
	if err != nil {
		t.Error(err)
	}

	var treeExpect []expectedParseNode = []expectedParseNode{
		{"in", 0},
		{"City", 1},
		{TokenListExpr, 1},
		{"'Seattle'", 2},
		{"'Atlanta'", 2},
		{"'Paris'", 2},
	}
	pos := 0
	err = CompareTree(tree, treeExpect, &pos, 0)
	if err != nil {
		printTree(tree)
		t.Errorf("Tree representation does not match expected value. error: %s", err.Error())
	}
}

// TestFilterInOperatorSingleValue tests the "IN" operator with a list containing a single value.
func TestFilterInOperatorSingleValue(t *testing.T) {
	tokenizer := FilterTokenizer()
	input := "City in ( 'Seattle' )"

	expect := []*Token{
		&Token{Value: "City", Type: FilterTokenLiteral},
		&Token{Value: "in", Type: FilterTokenLogical},
		&Token{Value: "(", Type: FilterTokenOpenParen},
		&Token{Value: "'Seattle'", Type: FilterTokenString},
		&Token{Value: ")", Type: FilterTokenCloseParen},
	}
	tokens, err := tokenizer.Tokenize(input)
	if err != nil {
		t.Error(err)
	}
	result, err := CompareTokens(expect, tokens)
	if !result {
		t.Error(err)
	}
	var postfix *tokenQueue
	postfix, err = GlobalFilterParser.InfixToPostfix(tokens)
	if err != nil {
		t.Error(err)
	}
	expect = []*Token{
		&Token{Value: "City", Type: FilterTokenLiteral},
		&Token{Value: "'Seattle'", Type: FilterTokenString},
		&Token{Value: "1", Type: TokenTypeArgCount},
		&Token{Value: TokenListExpr, Type: TokenTypeListExpr},
		&Token{Value: "in", Type: FilterTokenLogical},
	}
	result, err = CompareQueue(expect, postfix)
	if !result {
		t.Error(err)
	}

	tree, err := GlobalFilterParser.PostfixToTree(postfix)
	if err != nil {
		t.Error(err)
	}

	var treeExpect []expectedParseNode = []expectedParseNode{
		{"in", 0},
		{"City", 1},
		{TokenListExpr, 1},
		{"'Seattle'", 2},
	}
	pos := 0
	err = CompareTree(tree, treeExpect, &pos, 0)
	if err != nil {
		printTree(tree)
		t.Errorf("Tree representation does not match expected value. error: %s", err.Error())
	}
}

// TestFilterInOperatorEmptyList tests the "IN" operator with a list containing no value.
func TestFilterInOperatorEmptyList(t *testing.T) {
	tokenizer := FilterTokenizer()
	input := "City in ( )"

	expect := []*Token{
		&Token{Value: "City", Type: FilterTokenLiteral},
		&Token{Value: "in", Type: FilterTokenLogical},
		&Token{Value: "(", Type: FilterTokenOpenParen},
		&Token{Value: ")", Type: FilterTokenCloseParen},
	}
	tokens, err := tokenizer.Tokenize(input)
	if err != nil {
		t.Error(err)
	}
	result, err := CompareTokens(expect, tokens)
	if !result {
		t.Error(err)
	}
	var postfix *tokenQueue
	postfix, err = GlobalFilterParser.InfixToPostfix(tokens)
	if err != nil {
		t.Error(err)
	}
	expect = []*Token{
		&Token{Value: "City", Type: FilterTokenLiteral},
		&Token{Value: "0", Type: TokenTypeArgCount},
		&Token{Value: TokenListExpr, Type: TokenTypeListExpr},
		&Token{Value: "in", Type: FilterTokenLogical},
	}
	result, err = CompareQueue(expect, postfix)
	if !result {
		t.Error(err)
	}

	tree, err := GlobalFilterParser.PostfixToTree(postfix)
	if err != nil {
		t.Error(err)
	}

	var treeExpect []expectedParseNode = []expectedParseNode{
		{"in", 0},
		{"City", 1},
		{TokenListExpr, 1},
	}
	pos := 0
	err = CompareTree(tree, treeExpect, &pos, 0)
	if err != nil {
		printTree(tree)
		t.Errorf("Tree representation does not match expected value. error: %s", err.Error())
	}
}

// TestFilterInOperatorBothSides tests the "IN" operator.
// Use a listExpr on both sides of the IN operator.
//   listExpr  = OPEN BWS commonExpr BWS *( COMMA BWS commonExpr BWS ) CLOSE
// Validate if a list is within another list.
func TestFilterInOperatorBothSides(t *testing.T) {
	tokenizer := FilterTokenizer()
	input := "(1, 2) in ( ('ab', 'cd'), (1, 2), ('abc', 'def') )"

	expect := []*Token{
		&Token{Value: "(", Type: FilterTokenOpenParen},
		&Token{Value: "1", Type: FilterTokenInteger},
		&Token{Value: ",", Type: FilterTokenComma},
		&Token{Value: "2", Type: FilterTokenInteger},
		&Token{Value: ")", Type: FilterTokenCloseParen},
		&Token{Value: "in", Type: FilterTokenLogical},
		&Token{Value: "(", Type: FilterTokenOpenParen},

		&Token{Value: "(", Type: FilterTokenOpenParen},
		&Token{Value: "'ab'", Type: FilterTokenString},
		&Token{Value: ",", Type: FilterTokenComma},
		&Token{Value: "'cd'", Type: FilterTokenString},
		&Token{Value: ")", Type: FilterTokenCloseParen},
		&Token{Value: ",", Type: FilterTokenComma},

		&Token{Value: "(", Type: FilterTokenOpenParen},
		&Token{Value: "1", Type: FilterTokenInteger},
		&Token{Value: ",", Type: FilterTokenComma},
		&Token{Value: "2", Type: FilterTokenInteger},
		&Token{Value: ")", Type: FilterTokenCloseParen},
		&Token{Value: ",", Type: FilterTokenComma},

		&Token{Value: "(", Type: FilterTokenOpenParen},
		&Token{Value: "'abc'", Type: FilterTokenString},
		&Token{Value: ",", Type: FilterTokenComma},
		&Token{Value: "'def'", Type: FilterTokenString},
		&Token{Value: ")", Type: FilterTokenCloseParen},
		&Token{Value: ")", Type: FilterTokenCloseParen},
	}
	tokens, err := tokenizer.Tokenize(input)
	if err != nil {
		t.Error(err)
	}
	result, err := CompareTokens(expect, tokens)
	if !result {
		t.Error(err)
	}
	var postfix *tokenQueue
	postfix, err = GlobalFilterParser.InfixToPostfix(tokens)
	if err != nil {
		t.Error(err)
	}
	expect = []*Token{
		&Token{Value: "1", Type: FilterTokenInteger},
		&Token{Value: "2", Type: FilterTokenInteger},
		&Token{Value: "2", Type: TokenTypeArgCount},
		&Token{Value: TokenListExpr, Type: TokenTypeListExpr},

		&Token{Value: "'ab'", Type: FilterTokenString},
		&Token{Value: "'cd'", Type: FilterTokenString},
		&Token{Value: "2", Type: TokenTypeArgCount},
		&Token{Value: TokenListExpr, Type: TokenTypeListExpr},

		&Token{Value: "1", Type: FilterTokenInteger},
		&Token{Value: "2", Type: FilterTokenInteger},
		&Token{Value: "2", Type: TokenTypeArgCount},
		&Token{Value: TokenListExpr, Type: TokenTypeListExpr},

		&Token{Value: "'abc'", Type: FilterTokenString},
		&Token{Value: "'def'", Type: FilterTokenString},
		&Token{Value: "2", Type: TokenTypeArgCount},
		&Token{Value: TokenListExpr, Type: TokenTypeListExpr},

		&Token{Value: "3", Type: TokenTypeArgCount},
		&Token{Value: TokenListExpr, Type: TokenTypeListExpr},

		&Token{Value: "in", Type: FilterTokenLogical},
	}
	result, err = CompareQueue(expect, postfix)
	if !result {
		fmt.Printf("postfix notation: %s\n", postfix.String())
		t.Error(err)
	}

	tree, err := GlobalFilterParser.PostfixToTree(postfix)
	if err != nil {
		t.Error(err)
	}

	var treeExpect []expectedParseNode = []expectedParseNode{
		{"in", 0},
		{TokenListExpr, 1},
		{"1", 2},
		{"2", 2},
		//  ('ab', 'cd'), (1, 2), ('abc', 'def')
		{TokenListExpr, 1},
		{TokenListExpr, 2},
		{"'ab'", 3},
		{"'cd'", 3},
		{TokenListExpr, 2},
		{"1", 3},
		{"2", 3},
		{TokenListExpr, 2},
		{"'abc'", 3},
		{"'def'", 3},
	}
	pos := 0
	err = CompareTree(tree, treeExpect, &pos, 0)
	if err != nil {
		printTree(tree)
		t.Errorf("Tree representation does not match expected value. error: %s", err.Error())
	}
}

// TestFilterInOperatorWithFunc tests the "IN" operator with a comma-separated list
// of values, one of which is a function call which itself has a comma-separated list of values.
func TestFilterInOperatorWithFunc(t *testing.T) {
	tokenizer := FilterTokenizer()
	// 'Atlanta' is enclosed in a unecessary parenExpr to validate the expression is properly unwrapped.
	input := "City in ( 'Seattle', concat('San', 'Francisco'), ('Atlanta') )"

	{
		expect := []*Token{
			&Token{Value: "City", Type: FilterTokenLiteral},
			&Token{Value: "in", Type: FilterTokenLogical},
			&Token{Value: "(", Type: FilterTokenOpenParen},
			&Token{Value: "'Seattle'", Type: FilterTokenString},
			&Token{Value: ",", Type: FilterTokenComma},
			&Token{Value: "concat", Type: FilterTokenFunc},
			&Token{Value: "(", Type: FilterTokenOpenParen},
			&Token{Value: "'San'", Type: FilterTokenString},
			&Token{Value: ",", Type: FilterTokenComma},
			&Token{Value: "'Francisco'", Type: FilterTokenString},
			&Token{Value: ")", Type: FilterTokenCloseParen},
			&Token{Value: ",", Type: FilterTokenComma},
			&Token{Value: "(", Type: FilterTokenOpenParen},
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
	q, err := ParseFilterString(input)
	if err != nil {
		t.Errorf("Error parsing filter: %s", err.Error())
	}
	var expect []expectedParseNode = []expectedParseNode{
		{"in", 0},
		{"City", 1},
		{TokenListExpr, 1},
		{"'Seattle'", 2},
		{"concat", 2},
		{"'San'", 3},
		{"'Francisco'", 3},
		{"'Atlanta'", 2},
	}
	pos := 0
	err = CompareTree(q.Tree, expect, &pos, 0)
	if err != nil {
		printTree(q.Tree)
		t.Errorf("Tree representation does not match expected value. error: %s", err.Error())
	}
}

func TestFilterNotInListExpr(t *testing.T) {
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
			{TokenListExpr, 2},
			{"'Seattle'", 3},
			{"'Atlanta'", 3},
		}
		pos := 0
		err = CompareTree(tree, expect, &pos, 0)
		if err != nil {
			printTree(tree)
			t.Errorf("Tree representation does not match expected value. error: %s", err.Error())
		}

	}
}

func TestFilterAll(t *testing.T) {
	tokenizer := FilterTokenizer()
	input := "Tags/all(d:d/Key eq 'Site')"
	expect := []*Token{
		&Token{Value: "Tags", Type: FilterTokenLiteral},
		&Token{Value: "/", Type: FilterTokenOp},
		&Token{Value: "all", Type: FilterTokenLambda},
		&Token{Value: "(", Type: FilterTokenOpenParen},
		&Token{Value: "d", Type: FilterTokenLiteral},
		&Token{Value: ",", Type: FilterTokenColon},
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
		"substring(CompanyName,1,2) eq 'lf'", // substring with 3 arguments.
		// Bolean values
		"true",
		"false",
		"(true)",
		"((true))",
		"((true)) or false",
		"not true",
		"not false",
		"not (not true)",
		//"not not true", // TODO: I think this should work. 'not not true' is true
		// String functions
		"contains(CompanyName,'freds')",
		"endswith(CompanyName,'Futterkiste')",
		"startswith(CompanyName,'Alfr')",
		"length(CompanyName) eq 19",
		"indexof(CompanyName,'lfreds') eq 1",
		"substring(CompanyName,1) eq 'lfreds Futterkiste'", // substring() with 2 arguments.
		"'lfreds Futterkiste' eq substring(CompanyName,1)", // Same as above, but order of operands is reversed.
		"substring(CompanyName,1,2) eq 'lf'",               // substring() with 3 arguments.
		"'lf' eq substring(CompanyName,1,2) ",              // Same as above, but order of operands is reversed.
		"substringof('Alfreds', CompanyName) eq true",
		"tolower(CompanyName) eq 'alfreds futterkiste'",
		"toupper(CompanyName) eq 'ALFREDS FUTTERKISTE'",
		"trim(CompanyName) eq 'Alfreds Futterkiste'",
		"concat(concat(City,', '), Country) eq 'Berlin, Germany'",
		// GUID
		"GuidValue eq 01234567-89ab-cdef-0123-456789abcdef", // TODO According to ODATA ABNF notation, GUID values do not have quotes.
		// Date and Time functions
		"StartDate eq 2012-12-03",
		"DateTimeOffsetValue eq 2012-12-03T07:16:23Z",
		// duration      = [ "duration" ] SQUOTE durationValue SQUOTE
		// "DurationValue eq duration'P12DT23H59M59.999999999999S'", // TODO See ODATA ABNF notation
		"TimeOfDayValue eq 07:59:59.999",
		"year(BirthDate) eq 0",
		"month(BirthDate) eq 12",
		"day(StartTime) eq 8",
		"hour(StartTime) eq 1",
		"hour    (StartTime) eq 12",     // function followed by space characters
		"hour    ( StartTime   ) eq 15", // function followed by space characters
		"minute(StartTime) eq 0",
		"totaloffsetminutes(StartTime) eq 0",
		"second(StartTime) eq 0",
		"fractionalsecond(StartTime) lt 0.123456", // The fractionalseconds function returns the fractional seconds component of the
		// DateTimeOffset or TimeOfDay parameter value as a non-negative decimal value less than 1.
		"date(StartTime) ne date(EndTime)",
		"totaloffsetminutes(StartTime) eq 60",
		"StartTime eq mindatetime()",
		// "totalseconds(EndTime sub StartTime) lt duration'PT23H59'", // TODO The totalseconds function returns the duration of the value in total seconds, including fractional seconds.
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
		"GEO.INTERSECTS(Position,TargetArea)", // functions are case insensitive in ODATA 4.0.1
		// Logical operators
		"'Milk' eq 'Milk'",  // Compare two literals
		"'Water' ne 'Milk'", // Compare two literals
		"Name eq 'Milk'",
		"Name EQ 'Milk'", // operators are case insensitive in ODATA 4.0.1
		"Name ne 'Milk'",
		"Name NE 'Milk'",
		"Name gt 'Milk'",
		"Name ge 'Milk'",
		"Name lt 'Milk'",
		"Name le 'Milk'",
		"Name eq Name", // parameter equals to itself
		"Name eq 'Milk' and Price lt 2.55",
		"not endswith(Name,'ilk')",
		"Name eq 'Milk' or Price lt 2.55",
		"City eq 'Dallas' or City eq 'Houston'",
		// Nested properties
		"Product/Name eq 'Milk'",
		"Region/Product/Name eq 'Milk'",
		"Country/Region/Product/Name eq 'Milk'",
		//"style has Sales.Pattern'Yellow'", // TODO
		// Arithmetic operators
		"Price add 2.45 eq 5.00",
		"2.46 add Price eq 5.00",
		"Price add (2.47) eq 5.00",
		"(Price add (2.48)) eq 5.00",
		"Price ADD 2.49 eq 5.00", // 4.01 Services MUST support case-insensitive operator names.
		"Price sub 0.55 eq 2.00",
		"Price SUB 0.56 EQ 2.00", // 4.01 Services MUST support case-insensitive operator names.
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
		"not(S1 eq 'foo')",
		// Lambda operators
		"Tags/any()",                    // The any operator without an argument returns true if the collection is not empty
		"Tags/any(tag:tag eq 'London')", // 'Tags' is array of strings
		"Tags/any(tag:tag eq 'London' or tag eq 'Berlin')",          // 'Tags' is array of strings
		"Tags/any(var:var/Key eq 'Site' and var/Value eq 'London')", // 'Tags' is array of {"Key": "abc", "Value": "def"}
		"Tags/ANY(var:var/Key eq 'Site' AND var/Value eq 'London')",
		"Tags/any(var:var/Key eq 'Site' and var/Value eq 'London') and not (City in ('Dallas'))",
		"Tags/all(var:var/Key eq 'Site' and var/Value eq 'London')",
		"Price/any(t:not (12345 eq t))",
		// A long query.
		"Tags/any(var:var/Key eq 'Site' and var/Value eq 'London') or " +
			"Tags/any(var:var/Key eq 'Site' and var/Value eq 'Berlin') or " +
			"Tags/any(var:var/Key eq 'Site' and var/Value eq 'Paris') or " +
			"Tags/any(var:var/Key eq 'Site' and var/Value eq 'New York City') or " +
			"Tags/any(var:var/Key eq 'Site' and var/Value eq 'San Francisco')",
	}
	for _, input := range queries {
		q, err := ParseFilterString(input)
		if err != nil {
			t.Errorf("Error parsing query %s. Error: %s", input, err.Error())
			return
		} else if q.Tree == nil {
			t.Errorf("Error parsing query %s. Tree is nil", input)
		}
		if q.Tree.Token == nil {
			t.Errorf("Error parsing query %s. Root token is nil", input)
		}
		if q.Tree.Token.Type == FilterTokenLiteral {
			t.Errorf("Error parsing query %s. Unexpected root token type: %+v", input, q.Tree.Token)
		}
		//printTree(q.Tree)
	}
}

// The URLs below are not valid ODATA syntax, the parser should return an error.
func TestInvalidFilterSyntax(t *testing.T) {
	queries := []string{
		"()", // It's not a boolean expression
		"(TRUE)",
		"(City)",
		"(",
		"((((",
		")",
		"12345",                                // Number 12345 is not a boolean expression
		"0",                                    // Number 0 is not a boolean expression
		"'123'",                                // String '123' is not a boolean expression
		"TRUE",                                 // Should be 'true' lowercase
		"FALSE",                                // Should be 'false' lowercase
		"yes",                                  // yes is not a boolean expression
		"no",                                   // yes is not a boolean expression
		"",                                     // Empty string.
		"eq",                                   // Just a single logical operator
		"and",                                  // Just a single logical operator
		"add",                                  // Just a single arithmetic operator
		"add ",                                 // Just a single arithmetic operator
		"add 2",                                // Missing operands
		"add 2 3",                              // Missing operands
		"City",                                 // Just a single literal
		"City City City City",                  // Sequence of literals
		"City eq",                              // Missing operand
		"City eq (",                            // Wrong operand
		"City eq )",                            // Wrong operand
		"City equals 'Dallas'",                 // Unknown operator that starts with the same letters as a known operator
		"City near 'Dallas'",                   // Unknown operator that starts with the same letters as a known operator
		"City isNot 'Dallas'",                  // Unknown operator
		"not [City eq 'Dallas']",               // Wrong delimiter
		"not (City eq )",                       // Missing operand
		"not ((City eq 'Dallas'",               // Missing closing parenthesis
		"not (City eq 'Dallas'",                // Missing closing parenthesis
		"not (City eq 'Dallas'))",              // Extraneous closing parenthesis
		"not City eq 'Dallas')",                // Missing open parenthesis
		"City eq 'Dallas' orCity eq 'Houston'", // missing space between or and City
		// TODO: the query below should fail.
		//"Tags/any(var:var/Key eq 'Site') orTags/any(var:var/Key eq 'Site')",
		"not (City eq 'Dallas') and Name eq 'Houston')",
		"Tags/all()",                   // The all operator cannot be used without an argument expression.
		"LastName contains 'Smith'",    // Previously the godata library was not returning an error.
		"contains",                     // Function with missing parenthesis and arguments
		"contains()",                   // Function with missing arguments
		"contains LastName, 'Smith'",   // Missing parenthesis
		"contains(LastName)",           // Insufficent number of function arguments
		"contains(LastName, 'Smith'))", // Extraneous closing parenthesis
		"contains(LastName, 'Smith'",   // Missing closing parenthesis
		"contains LastName, 'Smith')",  // Missing open parenthesis
		"City eq 'Dallas' 'Houston'",   // extraneous string value
		//"contains(Name, 'a', 'b', 'c', 'd')", // Too many function arguments
	}
	for _, input := range queries {
		q, err := ParseFilterString(input)
		if err == nil {
			// The parser has incorrectly determined the syntax is valid.
			printTree(q.Tree)
			t.Errorf("The query '$filter=%s' is not valid ODATA syntax. The ODATA parser should return an error", input)
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
				tree.Children[0].Children[1].Children[0].Token.Value)
		}
		if tree.Children[0].Children[1].Children[1].Token.Value != TokenListExpr {
			t.Errorf("Unexpected operand for the 'in' operator. Expected 'list', got %s",
				tree.Children[0].Children[1].Children[1].Token.Value)
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
func CompareTokens(expected, actual []*Token) (bool, error) {
	if len(expected) != len(actual) {
		return false, fmt.Errorf("Different lengths. Expected %d, Got %d", len(expected), len(actual))
	}
	for i := range expected {
		if expected[i].Type != actual[i].Type {
			return false, fmt.Errorf("Different token types at index %d. Expected %v, Got %v. Value: %v",
				i, expected[i].Type, actual[i].Type, expected[i].Value)
		}
		if expected[i].Value != actual[i].Value {
			return false, fmt.Errorf("Different token values at index %d. Expected %v, Got %v",
				i, expected[i].Value, actual[i].Value)
		}
	}
	return true, nil
}

func CompareQueue(expect []*Token, b *tokenQueue) (bool, error) {
	bl := func() int {
		if b.Empty() {
			return 0
		}
		l := 1
		for node := b.Head; node != b.Tail; node = node.Next {
			l++
		}
		return l
	}()
	if len(expect) != bl {
		return false, fmt.Errorf("Different lengths. Got %d, expected %d", bl, len(expect))
	}
	node := b.Head
	for i := range expect {
		if expect[i].Type != node.Token.Type {
			return false, fmt.Errorf("Different token types at index %d. Got: %v, expected: %v. Expected value: %v",
				i, node.Token.Type, expect[i].Type, expect[i].Value)
		}
		if expect[i].Value != node.Token.Value {
			return false, fmt.Errorf("Different token values at index %d. Got: %v, expected: %v",
				i, node.Token.Value, expect[i].Value)
		}
		node = node.Next
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

func printTree(n *ParseNode) {
	fmt.Printf("Tree:\n%s\n", n.String())
}

func printTokens(tokens []*Token) {
	s := make([]string, len(tokens))
	for i := range tokens {
		s[i] = tokens[i].Value
	}
	fmt.Printf("TOKENS: %s\n", strings.Join(s, " "))
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

	var expect []expectedParseNode = []expectedParseNode{
		{"eq", 0},
		{"/", 1},
		{"Address", 2},
		{"City", 2},
		{"'Redmond'", 1},
	}
	pos := 0
	err = CompareTree(tree, expect, &pos, 0)
	if err != nil {
		printTree(tree)
		t.Errorf("Tree representation does not match expected value. error: %s", err.Error())
	}
}

func TestMultipleNestedPath(t *testing.T) {
	input := "Product/Address/City eq 'Redmond'"
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
		{"/", 1},
		{"/", 2},
		{"Product", 3},
		{"Address", 3},
		{"City", 2},
		{"'Redmond'", 1},
	}
	pos := 0
	err = CompareTree(tree, expect, &pos, 0)
	if err != nil {
		printTree(tree)
		t.Errorf("Tree representation does not match expected value. error: %s", err.Error())
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
			printTree(tree)
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
			printTree(tree)
			t.Errorf("Tree representation does not match expected value. error: %s", err.Error())
		}
	}
}

func TestSubstringofFunction(t *testing.T) {
	// Previously, the parser was incorrectly interpreting the 'substringof' function as the 'sub' operator.
	input := "substringof('Alfreds', CompanyName) eq true"
	tokens, err := GlobalFilterTokenizer.Tokenize(input)
	if err != nil {
		t.Error(err)
		return
	}
	{
		expect := []*Token{
			&Token{Value: "substringof", Type: FilterTokenFunc},
			&Token{Value: "(", Type: FilterTokenOpenParen},
			&Token{Value: "'Alfreds'", Type: FilterTokenString},
			&Token{Value: ",", Type: FilterTokenComma},
			&Token{Value: "CompanyName", Type: FilterTokenLiteral},
			&Token{Value: ")", Type: FilterTokenCloseParen},
			&Token{Value: "eq", Type: FilterTokenLogical},
			&Token{Value: "true", Type: FilterTokenBoolean},
		}
		result, err := CompareTokens(expect, tokens)
		if !result {
			t.Error(err)
		}
	}
	output, err := GlobalFilterParser.InfixToPostfix(tokens)
	if err != nil {
		t.Error(err)
		return
	}
	{
		expect := []*Token{
			&Token{Value: "'Alfreds'", Type: FilterTokenString},
			&Token{Value: "CompanyName", Type: FilterTokenLiteral},
			&Token{Value: "2", Type: TokenTypeArgCount}, // The number of function arguments.
			&Token{Value: "substringof", Type: FilterTokenFunc},
			&Token{Value: "true", Type: FilterTokenBoolean},
			&Token{Value: "eq", Type: FilterTokenLogical},
		}
		result, err := CompareQueue(expect, output)
		if !result {
			t.Error(err)
		}
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
		printTree(tree)
		t.Errorf("Tree representation does not match expected value. error: %s", err.Error())
	}
}

// TestSubstringNestedFunction tests the substring function with a nested call
// to substring, with the use of 2-argument and 3-argument substring.
func TestSubstringNestedFunction(t *testing.T) {
	// Previously, the parser was incorrectly interpreting the 'substringof' function as the 'sub' operator.
	input := "substring(substring('Francisco', 1), 3, 2) eq 'ci'"
	tokens, err := GlobalFilterTokenizer.Tokenize(input)
	if err != nil {
		t.Error(err)
		return
	}
	{
		expect := []*Token{
			&Token{Value: "substring", Type: FilterTokenFunc},
			&Token{Value: "(", Type: FilterTokenOpenParen},
			&Token{Value: "substring", Type: FilterTokenFunc},
			&Token{Value: "(", Type: FilterTokenOpenParen},
			&Token{Value: "'Francisco'", Type: FilterTokenString},
			&Token{Value: ",", Type: FilterTokenComma},
			&Token{Value: "1", Type: FilterTokenInteger},
			&Token{Value: ")", Type: FilterTokenCloseParen},
			&Token{Value: ",", Type: FilterTokenComma},
			&Token{Value: "3", Type: FilterTokenInteger},
			&Token{Value: ",", Type: FilterTokenComma},
			&Token{Value: "2", Type: FilterTokenInteger},
			&Token{Value: ")", Type: FilterTokenCloseParen},
			&Token{Value: "eq", Type: FilterTokenLogical},
			&Token{Value: "'ci'", Type: FilterTokenString},
		}
		result, err := CompareTokens(expect, tokens)
		if !result {
			t.Error(err)
		}
	}
	output, err := GlobalFilterParser.InfixToPostfix(tokens)
	if err != nil {
		t.Error(err)
		return
	}
	{
		expect := []*Token{
			&Token{Value: "'Francisco'", Type: FilterTokenString},
			&Token{Value: "1", Type: FilterTokenInteger},
			&Token{Value: "2", Type: TokenTypeArgCount}, // The number of function arguments.
			&Token{Value: "substring", Type: FilterTokenFunc},
			&Token{Value: "3", Type: FilterTokenInteger},
			&Token{Value: "2", Type: FilterTokenInteger},
			&Token{Value: "3", Type: TokenTypeArgCount}, // The number of function arguments.
			&Token{Value: "substring", Type: FilterTokenFunc},
			&Token{Value: "'ci'", Type: FilterTokenString},
			&Token{Value: "eq", Type: FilterTokenLogical},
		}
		result, err := CompareQueue(expect, output)
		if !result {
			t.Error(err)
		}
	}
	tree, err := GlobalFilterParser.PostfixToTree(output)
	if err != nil {
		t.Error(err)
		return
	}
	var expect []expectedParseNode = []expectedParseNode{
		{"eq", 0},
		{"substring", 1},
		{"substring", 2},
		{"'Francisco'", 3},
		{"1", 3},
		{"3", 2},
		{"2", 2},
		{"'ci'", 1},
	}
	pos := 0
	err = CompareTree(tree, expect, &pos, 0)
	if err != nil {
		printTree(tree)
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
		printTree(tree)
		t.Errorf("Tree representation does not match expected value. error: %s", err.Error())
	}
}

func TestLambdaAny(t *testing.T) {
	input := "Tags/any(var:var/Key eq 'Site')"
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
		{"var", 2},
		{"eq", 2},
		{"/", 3},
		{"var", 4},
		{"Key", 4},
		{"'Site'", 3},
	}
	pos := 0
	err = CompareTree(tree, expect, &pos, 0)
	if err != nil {
		printTree(tree)
		t.Errorf("Tree representation does not match expected value. error: %s", err.Error())
	}
}

func TestLambdaAnyNot(t *testing.T) {
	input := "Price/any(t:not (12345 eq t ))"

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
	{
		expect := []*Token{
			&Token{Value: "Price", Type: FilterTokenLiteral},
			&Token{Value: "t", Type: FilterTokenLiteral},
			&Token{Value: "12345", Type: FilterTokenInteger},
			&Token{Value: "t", Type: FilterTokenLiteral},
			&Token{Value: "eq", Type: FilterTokenLogical},
			&Token{Value: "not", Type: FilterTokenLogical},
			&Token{Value: "2", Type: TokenTypeArgCount},
			&Token{Value: "any", Type: FilterTokenLambda},
			&Token{Value: "/", Type: FilterTokenOp},
		}
		var result bool
		result, err = CompareQueue(expect, output)
		if !result {
			t.Error(err)
		}
	}
	tree, err := GlobalFilterParser.PostfixToTree(output)
	if err != nil {
		t.Error(err)
		return
	}
	var expect []expectedParseNode = []expectedParseNode{
		{"/", 0},
		{"Price", 1},
		{"any", 1},
		{"t", 2},
		{"not", 2},
		{"eq", 3},
		{"12345", 4},
		{"t", 4},
	}
	pos := 0
	err = CompareTree(tree, expect, &pos, 0)
	if err != nil {
		printTree(tree)
		t.Errorf("Tree representation does not match expected value. error: %s", err.Error())
	}
}

func TestLambdaAnyAnd(t *testing.T) {
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
		{"var", 2},
		{"and", 2},
		{"eq", 3},
		{"/", 4},
		{"var", 5},
		{"Key", 5},
		{"'Site'", 4},
		{"eq", 3},
		{"/", 4},
		{"var", 5},
		{"Value", 5},
		{"'London'", 4},
	}
	pos := 0
	err = CompareTree(tree, expect, &pos, 0)
	if err != nil {
		printTree(tree)
		t.Errorf("Tree representation does not match expected value. error: %s", err.Error())
	}
}

func TestLambdaNestedAny(t *testing.T) {
	input := "Enabled/any(t:t/Value eq Config/any(c:c/AdminState eq 'TRUE'))"
	q, err := ParseFilterString(input)
	if err != nil {
		t.Errorf("Error parsing query %s. Error: %s", input, err.Error())
		return
	}
	var expect []expectedParseNode = []expectedParseNode{
		{"/", 0},
		{"Enabled", 1},
		{"any", 1},
		{"t", 2},
		{"eq", 2},
		{"/", 3},
		{"t", 4},
		{"Value", 4},
		{"/", 3},
		{"Config", 4},
		{"any", 4},
		{"c", 5},
		{"eq", 5},
		{"/", 6},
		{"c", 7},
		{"AdminState", 7},
		{"'TRUE'", 6},
	}
	pos := 0
	err = CompareTree(q.Tree, expect, &pos, 0)
	if err != nil {
		printTree(q.Tree)
		t.Errorf("Tree representation does not match expected value. error: %s", err.Error())
	}
}

// TestLambdaAnyNested validates the any() lambda function with multiple nested properties.
func TestLambdaAnyNestedProperties(t *testing.T) {
	input := "Config/any(var:var/Config/Priority eq 123)"
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
		{"Config", 1},
		{"any", 1},
		{"var", 2},
		{"eq", 2},
		{"/", 3},
		{"/", 4},
		{"var", 5},
		{"Config", 5},
		{"Priority", 4},
		{"123", 3},
	}
	pos := 0
	err = CompareTree(tree, expect, &pos, 0)
	if err != nil {
		printTree(tree)
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
		{"var", 2},
		{"or", 2},
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
		{"gt", 3},
		{"Price", 4},
		{"1.0", 4},
	}
	pos := 0
	err = CompareTree(tree, expect, &pos, 0)
	if err != nil {
		printTree(tree)
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
		{"var", 2},
		{"or", 2},
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
		{"contains", 3},
		{"/", 4},
		{"var", 5},
		{"Value", 5},
		{"'Smith'", 4},
	}
	pos := 0
	err = CompareTree(tree, expect, &pos, 0)
	if err != nil {
		printTree(tree)
		t.Errorf("Tree representation does not match expected value. error: %s", err.Error())
	}
}

func TestFilterTokenizerExists(t *testing.T) {

	tokenizer := FilterTokenizer()
	input := "exists(Name,false)"
	expect := []*Token{
		&Token{Value: "exists", Type: FilterTokenFunc},
		&Token{Value: "(", Type: FilterTokenOpenParen},
		&Token{Value: "Name", Type: FilterTokenLiteral},
		&Token{Value: ",", Type: FilterTokenComma},
		&Token{Value: "false", Type: FilterTokenBoolean},
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
