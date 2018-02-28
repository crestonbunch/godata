package godata

const (
	FilterTokenOpenParen int = iota
	FilterTokenCloseParen
	FilterTokenWhitespace
	FilterTokenNav
	FilterTokenColon // for 'any' and 'all' lambda operators
	FilterTokenComma
	FilterTokenLogical
	FilterTokenOp
	FilterTokenFunc
	FilterTokenLambda
	FilterTokenNull
	FilterTokenIt
	FilterTokenRoot
	FilterTokenFloat
	FilterTokenInteger
	FilterTokenString
	FilterTokenDate
	FilterTokenTime
	FilterTokenDateTime
	FilterTokenBoolean
	FilterTokenLiteral
)

var GlobalFilterTokenizer = FilterTokenizer()
var GlobalFilterParser = FilterParser()

// Convert an input string from the $filter part of the URL into a parse
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
	return &GoDataFilterQuery{tree}, nil
}

// Create a tokenizer capable of tokenizing filter statements
func FilterTokenizer() *Tokenizer {
	t := Tokenizer{}
	t.Add("^[0-9]{4,4}-[0-9]{2,2}-[0-9]{2,2}T[0-9]{2,2}:[0-9]{2,2}(:[0-9]{2,2}(.[0-9]+)?)?(Z|[+-][0-9]{2,2}:[0-9]{2,2})", FilterTokenDateTime)
	t.Add("^-?[0-9]{4,4}-[0-9]{2,2}-[0-9]{2,2}", FilterTokenDate)
	t.Add("^[0-9]{2,2}:[0-9]{2,2}(:[0-9]{2,2}(.[0-9]+)?)?", FilterTokenTime)
	t.Add("^\\(", FilterTokenOpenParen)
	t.Add("^\\)", FilterTokenCloseParen)
	t.Add("^/", FilterTokenNav)
	t.Add("^:", FilterTokenColon)
	t.Add("^,", FilterTokenComma)
	t.Add("^(eq|ne|gt|ge|lt|le|and|or|not|has|in)", FilterTokenLogical)
	t.Add("^(add|sub|mul|divby|div|mod)", FilterTokenOp)
	t.Add("^(contains|endswith|startswith|length|indexof|substring|tolower|toupper|"+
		"trim|concat|year|month|day|hour|minute|second|fractionalseconds|date|"+
		"time|totaloffsetminutes|now|maxdatetime|mindatetime|totalseconds|round|"+
		"floor|ceiling|isof|cast|geo.distance|geo.intersects|geo.length)", FilterTokenFunc)
	t.Add("^(any|all)", FilterTokenLambda)
	t.Add("^null", FilterTokenNull)
	t.Add("^\\$it", FilterTokenIt)
	t.Add("^\\$root", FilterTokenRoot)
	t.Add("^-?[0-9]+\\.[0-9]+", FilterTokenFloat)
	t.Add("^-?[0-9]+", FilterTokenInteger)
	t.Add("^'(''|[^'])*'", FilterTokenString)
	t.Add("^(true|false)", FilterTokenBoolean)
	t.Add("^[a-zA-Z][a-zA-Z0-9_.]*", FilterTokenLiteral)
	t.Ignore("^ ", FilterTokenWhitespace)

	return &t
}

func FilterParser() *Parser {
	parser := EmptyParser()
	parser.DefineOperator("/", 2, OpAssociationLeft, 8)
	parser.DefineOperator("has", 2, OpAssociationLeft, 8)
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
	parser.DefineOperator("isof", 2, OpAssociationLeft, 4)
	parser.DefineOperator("eq", 2, OpAssociationLeft, 3)
	parser.DefineOperator("ne", 2, OpAssociationLeft, 3)
	parser.DefineOperator("in", -1, OpAssociationLeft, 3) // 'in' operator takes a literal list.
	parser.DefineOperator("and", 2, OpAssociationLeft, 2)
	parser.DefineOperator("or", 2, OpAssociationLeft, 1)
	parser.DefineOperator(":", 2, OpAssociationLeft, 1)
	parser.DefineFunction("contains", 2)
	parser.DefineFunction("endswith", 2)
	parser.DefineFunction("startswith", 2)
	parser.DefineFunction("length", 1)
	parser.DefineFunction("indexof", 2)
	parser.DefineFunction("substring", 2)
	parser.DefineFunction("tolower", 1)
	parser.DefineFunction("toupper", 1)
	parser.DefineFunction("trim", 1)
	parser.DefineFunction("concat", 2)
	parser.DefineFunction("year", 1)
	parser.DefineFunction("month", 1)
	parser.DefineFunction("day", 1)
	parser.DefineFunction("hour", 1)
	parser.DefineFunction("minute", 1)
	parser.DefineFunction("second", 1)
	parser.DefineFunction("fractionalseconds", 1)
	parser.DefineFunction("date", 1)
	parser.DefineFunction("time", 1)
	parser.DefineFunction("totaloffsetminutes", 1)
	parser.DefineFunction("now", 0)
	parser.DefineFunction("maxdatetime", 0)
	parser.DefineFunction("mindatetime", 0)
	parser.DefineFunction("totalseconds", 1)
	parser.DefineFunction("round", 1)
	parser.DefineFunction("floor", 1)
	parser.DefineFunction("ceiling", 1)
	parser.DefineFunction("isof", 2)
	parser.DefineFunction("cast", 2)
	parser.DefineFunction("geo.distance", 2)
	parser.DefineFunction("geo.intersects", 2)
	parser.DefineFunction("geo.length", 1)
	parser.DefineFunction("any", 1)
	parser.DefineFunction("all", 1)

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
