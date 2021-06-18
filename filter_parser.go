package godata

const (
	FilterTokenOpenParen int = iota
	FilterTokenCloseParen
	FilterTokenWhitespace
	FilterTokenNav     // Property navigation
	FilterTokenColon   // Function arg separator for 'any(v:boolExpr)' and 'all(v:boolExpr)' lambda operators
	FilterTokenComma   // 5
	FilterTokenLogical // eq|ne|gt|ge|lt|le|and|or|not|has|in
	FilterTokenOp      // add|sub|mul|divby|div|mod and "/" token when used in lambda expression, e.g. tags/any()
	FilterTokenFunc
	FilterTokenLambda // any(), all() lambda functions
	FilterTokenNull   // 10
	FilterTokenIt
	FilterTokenRoot
	FilterTokenFloat
	FilterTokenInteger
	FilterTokenString // 15
	FilterTokenDate
	FilterTokenTime
	FilterTokenDateTime
	FilterTokenBoolean
	FilterTokenLiteral  // 20
	FilterTokenDuration // duration      = [ "duration" ] SQUOTE durationValue SQUOTE
	FilterTokenGuid
)

var GlobalFilterTokenizer = FilterTokenizer()
var GlobalFilterParser = FilterParser()

// ParseFilterString converts an input string from the $filter part of the URL into a parse
// tree that can be used by providers to create a response.
func ParseFilterString(filter string) (*GoDataFilterQuery, error) {
	tokens, err := GlobalFilterTokenizer.Tokenize(filter)
	if err != nil {
		return nil, err
	}
	// TODO: can we do this in one fell swoop?
	postfix, err := GlobalFilterParser.InfixToPostfix(tokens)
	if err != nil {
		return nil, err
	}
	tree, err := GlobalFilterParser.PostfixToTree(postfix)
	if err != nil {
		return nil, err
	}
	if tree == nil || tree.Token == nil ||
		(len(tree.Children) == 0 && tree.Token.Type != FilterTokenBoolean) {
		return nil, BadRequestError("Value must be a boolean expression")
	}
	return &GoDataFilterQuery{tree, filter}, nil
}

// FilterTokenDurationRe is a regex for a token of type duration.
// The token value is set to the ISO 8601 string inside the single quotes
// For example, if the input data is duration'PT2H', then the token value is set to PT2H without quotes.
const FilterTokenDurationRe = `^(duration)?'(?P<subtoken>-?P((([0-9]+Y([0-9]+M)?([0-9]+D)?|([0-9]+M)([0-9]+D)?|([0-9]+D))(T(([0-9]+H)([0-9]+M)?([0-9]+(\.[0-9]+)?S)?|([0-9]+M)([0-9]+(\.[0-9]+)?S)?|([0-9]+(\.[0-9]+)?S)))?)|(T(([0-9]+H)([0-9]+M)?([0-9]+(\.[0-9]+)?S)?|([0-9]+M)([0-9]+(\.[0-9]+)?S)?|([0-9]+(\.[0-9]+)?S)))))'`

// FilterTokenizer creates a tokenizer capable of tokenizing filter statements
// 4.01 Services MUST support case-insensitive operator names.
// See https://docs.oasis-open.org/odata/odata/v4.01/odata-v4.01-part2-url-conventions.html#_Toc31360955
func FilterTokenizer() *Tokenizer {
	t := Tokenizer{}
	// guidValue = 8HEXDIG "-" 4HEXDIG "-" 4HEXDIG "-" 4HEXDIG "-" 12HEXDIG
	t.Add(`^[[:xdigit:]]{8}-[[:xdigit:]]{4}-[[:xdigit:]]{4}-[[:xdigit:]]{4}-[[:xdigit:]]{12}`, FilterTokenGuid)
	// duration      = [ "duration" ] SQUOTE durationValue SQUOTE
	// durationValue = [ SIGN ] "P" [ 1*DIGIT "D" ] [ "T" [ 1*DIGIT "H" ] [ 1*DIGIT "M" ] [ 1*DIGIT [ "." 1*DIGIT ] "S" ] ]
	// Duration literals in OData 4.0 required prefixing with “duration”.
	// In OData 4.01, services MUST support duration and enumeration literals with or without the type prefix.
	// OData clients that want to operate across OData 4.0 and OData 4.01 services should always include the prefix for duration and enumeration types.
	t.Add(FilterTokenDurationRe, FilterTokenDuration)
	t.Add("^[0-9]{4,4}-[0-9]{2,2}-[0-9]{2,2}T[0-9]{2,2}:[0-9]{2,2}(:[0-9]{2,2}(.[0-9]+)?)?(Z|[+-][0-9]{2,2}:[0-9]{2,2})", FilterTokenDateTime)
	t.Add("^-?[0-9]{4,4}-[0-9]{2,2}-[0-9]{2,2}", FilterTokenDate)
	t.Add("^[0-9]{2,2}:[0-9]{2,2}(:[0-9]{2,2}(.[0-9]+)?)?", FilterTokenTime)
	t.Add("^\\(", FilterTokenOpenParen)
	t.Add("^\\)", FilterTokenCloseParen)
	t.Add("^(?P<token>/)(?i)(any|all)", FilterTokenOp)                                     // '/' as a token between a collection expression and a lambda function any() or all()
	t.Add("^/", FilterTokenNav)                                                            // '/' as a token for property navigation.
	t.AddWithSubstituteFunc("^:", FilterTokenColon, func(in string) string { return "," }) // Function arg separator for lambda functions (any, all)
	t.Add("^,", FilterTokenComma)                                                          // Default arg separator for functions
	// Per ODATA ABNF grammar, functions must be followed by a open parenthesis.
	// This implementation is a bit more lenient and allows space character between
	// the function name and the open parenthesis.
	// TODO: If we remove the optional space character, the function token will be
	// mistakenly interpreted as a literal.
	// E.g. ABNF for 'geo.distance':
	// distanceMethodCallExpr   = "geo.distance"   OPEN BWS commonExpr BWS COMMA BWS commonExpr BWS CLOSE
	t.Add("(?i)^(?P<token>(geo.distance|geo.intersects|geo.length))[\\s(]", FilterTokenFunc)
	// According to ODATA ABNF notation, functions must be followed by a open parenthesis with no space
	// between the function name and the open parenthesis.
	// However, we are leniently allowing space characters between the function and the open parenthesis.
	// TODO make leniency configurable.
	// E.g. ABNF for 'indexof':
	// indexOfMethodCallExpr    = "indexof"    OPEN BWS commonExpr BWS COMMA BWS commonExpr BWS CLOSE
	t.Add("(?i)^(?P<token>(substringof|substring|length|indexof|exists))[\\s(]", FilterTokenFunc)
	// Logical operators must be followed by a space character.
	// However, in practice user have written requests such as not(City eq 'Seattle')
	// We are leniently allowing space characters between the operator name and the open parenthesis.
	// TODO make leniency configurable.
	// Example:
	// notExpr = "not" RWS boolCommonExpr
	t.Add("(?i)^(?P<token>(eq|ne|gt|ge|lt|le|and|or|not|has|in))[\\s(]", FilterTokenLogical)
	// Arithmetic operators must be followed by a space character.
	t.Add("(?i)^(?P<token>(add|sub|mul|divby|div|mod))\\s", FilterTokenOp)
	// According to ODATA ABNF notation, functions must be followed by a open parenthesis with no space
	// between the function name and the open parenthesis.
	// However, we are leniently allowing space characters between the function and the open parenthesis.
	// TODO make leniency configurable.
	//
	// E.g. ABNF for 'contains':
	// containsMethodCallExpr   = "contains"   OPEN BWS commonExpr BWS COMMA BWS commonExpr BWS CLOSE
	t.Add("(?i)^(?P<token>(contains|endswith|startswith|tolower|toupper|"+
		"trim|concat|year|month|day|hour|minute|second|fractionalseconds|date|"+
		"time|totaloffsetminutes|now|maxdatetime|mindatetime|totalseconds|round|"+
		"floor|ceiling|isof|cast))[\\s(]", FilterTokenFunc)
	// anyExpr = "any" OPEN BWS [ lambdaVariableExpr BWS COLON BWS lambdaPredicateExpr ] BWS CLOSE
	// allExpr = "all" OPEN BWS   lambdaVariableExpr BWS COLON BWS lambdaPredicateExpr   BWS CLOSE
	t.Add("(?i)^(?P<token>(any|all))[\\s(]", FilterTokenLambda)
	t.Add("^null", FilterTokenNull)
	t.Add("^\\$it", FilterTokenIt)
	t.Add("^\\$root", FilterTokenRoot)
	t.Add("^-?[0-9]+\\.[0-9]+", FilterTokenFloat)
	t.Add("^-?[0-9]+", FilterTokenInteger)
	t.Add("^'(''|[^'])*'", FilterTokenString)
	t.Add("^(true|false)", FilterTokenBoolean)
	t.Add("^@*[a-zA-Z][a-zA-Z0-9_.]*", FilterTokenLiteral) // The optional '@' character is used to identify parameter aliases
	t.Ignore("^ ", FilterTokenWhitespace)

	return &t
}

func FilterParser() *Parser {
	parser := EmptyParser()
	parser.DefineOperator("/", 2, OpAssociationLeft, 8) // Note: '/' is used as a property navigator and between a collExpr and lambda function.
	parser.DefineOperator("has", 2, OpAssociationLeft, 8)
	// 'in' operator takes a literal list.
	// City in ('Seattle') needs to be interpreted as a list expression, not a paren expression.
	parser.DefineOperator("in", 2, OpAssociationLeft, 8).SetPreferListExpr(true)
	parser.DefineOperator("-", 1, OpAssociationNone, 7)
	parser.DefineOperator("not", 1, OpAssociationLeft, 7)
	parser.DefineOperator("cast", 2, OpAssociationNone, 7)
	parser.DefineOperator("mul", 2, OpAssociationNone, 6)
	parser.DefineOperator("div", 2, OpAssociationNone, 6)   // Division
	parser.DefineOperator("divby", 2, OpAssociationNone, 6) // Decimal Division
	parser.DefineOperator("mod", 2, OpAssociationNone, 6)
	parser.DefineOperator("add", 2, OpAssociationNone, 5)
	parser.DefineOperator("sub", 2, OpAssociationNone, 5)
	parser.DefineOperator("gt", 2, OpAssociationLeft, 4)
	parser.DefineOperator("ge", 2, OpAssociationLeft, 4)
	parser.DefineOperator("lt", 2, OpAssociationLeft, 4)
	parser.DefineOperator("le", 2, OpAssociationLeft, 4)
	parser.DefineOperator("eq", 2, OpAssociationLeft, 3)
	parser.DefineOperator("ne", 2, OpAssociationLeft, 3)
	parser.DefineOperator("and", 2, OpAssociationLeft, 2)
	parser.DefineOperator("or", 2, OpAssociationLeft, 1)
	parser.DefineFunction("contains", []int{2})
	parser.DefineFunction("endswith", []int{2})
	parser.DefineFunction("startswith", []int{2})
	parser.DefineFunction("exists", []int{2})
	parser.DefineFunction("length", []int{1})
	parser.DefineFunction("indexof", []int{2})
	parser.DefineFunction("substring", []int{2, 3})
	parser.DefineFunction("substringof", []int{2})
	parser.DefineFunction("tolower", []int{1})
	parser.DefineFunction("toupper", []int{1})
	parser.DefineFunction("trim", []int{1})
	parser.DefineFunction("concat", []int{2})
	parser.DefineFunction("year", []int{1})
	parser.DefineFunction("month", []int{1})
	parser.DefineFunction("day", []int{1})
	parser.DefineFunction("hour", []int{1})
	parser.DefineFunction("minute", []int{1})
	parser.DefineFunction("second", []int{1})
	parser.DefineFunction("fractionalseconds", []int{1})
	parser.DefineFunction("date", []int{1})
	parser.DefineFunction("time", []int{1})
	parser.DefineFunction("totaloffsetminutes", []int{1})
	parser.DefineFunction("now", []int{0})
	parser.DefineFunction("maxdatetime", []int{0})
	parser.DefineFunction("mindatetime", []int{0})
	parser.DefineFunction("totalseconds", []int{1})
	parser.DefineFunction("round", []int{1})
	parser.DefineFunction("floor", []int{1})
	parser.DefineFunction("ceiling", []int{1})
	parser.DefineFunction("isof", []int{1, 2}) // isof function can take one or two arguments.
	parser.DefineFunction("cast", []int{2})
	parser.DefineFunction("geo.distance", []int{2})
	parser.DefineFunction("geo.intersects", []int{2})
	parser.DefineFunction("geo.length", []int{1})
	parser.DefineFunction("any", []int{0, 2}) // 'any' can take either zero or one argument.
	parser.DefineFunction("all", []int{2})

	return parser
}

func SemanticizeFilterQuery(
	filter *GoDataFilterQuery,
	service *GoDataService,
	entity *GoDataEntityType,
) error {

	if filter == nil || filter.Tree == nil {
		return nil
	}

	var semanticizeFilterNode func(node *ParseNode) error
	semanticizeFilterNode = func(node *ParseNode) error {

		if node.Token.Type == FilterTokenLiteral {
			prop, ok := service.PropertyLookup[entity][node.Token.Value]
			if !ok {
				return BadRequestError("No property found " + node.Token.Value + " on entity " + entity.Name)
			}
			node.Token.SemanticType = SemanticTypeProperty
			node.Token.SemanticReference = prop
		} else {
			node.Token.SemanticType = SemanticTypePropertyValue
			node.Token.SemanticReference = &node.Token.Value
		}

		for _, child := range node.Children {
			err := semanticizeFilterNode(child)
			if err != nil {
				return err
			}
		}

		return nil
	}

	return semanticizeFilterNode(filter.Tree)
}
