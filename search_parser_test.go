package godata

import (
	"testing"
)

func TestSearchQuery(t *testing.T) {
	tokenizer := SearchTokenizer()
	input := "mountain OR (\"red bikes\" AND avocados)"

	expect := []*Token{
		&Token{Value: "mountain", Type: SearchTokenLiteral},
		&Token{Value: "OR", Type: SearchTokenOp},
		&Token{Value: "(", Type: SearchTokenOpenParen},
		&Token{Value: "\"red bikes\"", Type: SearchTokenLiteral},
		&Token{Value: "AND", Type: SearchTokenOp},
		&Token{Value: "avocados", Type: SearchTokenLiteral},
		&Token{Value: ")", Type: SearchTokenCloseParen},
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
