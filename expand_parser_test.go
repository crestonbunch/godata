package godata

import (
	"testing"
)

func TestTrivialExpand(t *testing.T) {
	input := "Products/Categories"

	output, err := ParseExpandString(input)

	if err != nil {
		t.Error(err)
		return
	}

	if output.ExpandItems[0].Path[0].Value != "Products" {
		t.Error("First path item is not 'Products'")
		return
	}
	if output.ExpandItems[0].Path[1].Value != "Categories" {
		t.Error("Second path item is not 'Categories'")
		return
	}
}

func TestSimpleExpand(t *testing.T) {
	input := "Products($filter=DiscontinuedDate eq null)"

	output, err := ParseExpandString(input)

	if err != nil {
		t.Error(err)
		return
	}

	if output.ExpandItems[0].Path[0].Value != "Products" {
		t.Error("First path item is not 'Products'")
		return
	}

	if output.ExpandItems[0].Filter == nil {
		t.Error("Filter not parsed")
		return
	}

	if output.ExpandItems[0].Filter.Tree == nil {
		t.Error("Filter tree is null")
		return
	}

	if output.ExpandItems[0].Filter.Tree.Token.Value != "eq" {
		t.Error("Filter not parsed correctly")
		return
	}
}

func TestExpandNestedCommas(t *testing.T) {
	input := "DirectReports($select=FirstName,LastName;$levels=4)"

	output, err := ParseExpandString(input)

	if err != nil {
		t.Error(err)
		return
	}

	if output.ExpandItems[0].Path[0].Value != "DirectReports" {
		t.Error("First path item is not 'DirectReports'")
		return
	}

	if output.ExpandItems[0].Select.SelectItems[0].Segments[0].Value != "FirstName" {
		actual := output.ExpandItems[0].Select.SelectItems[0].Segments[0]
		t.Error("First select segment is '" + actual.Value + "', expected 'FirstName'")
		return
	}

	if output.ExpandItems[0].Select.SelectItems[1].Segments[0].Value != "LastName" {
		actual := output.ExpandItems[0].Select.SelectItems[1].Segments[0]
		t.Error("First select segment is '" + actual.Value + "', expected 'LastName'")
		return
	}

	if output.ExpandItems[0].Levels != 4 {
		t.Error("Levels does not equal 4")
		return
	}

}

func TestExpandNestedParens(t *testing.T) {
	input := "Products($filter=not (DiscontinuedDate eq null))"

	output, err := ParseExpandString(input)

	if err != nil {
		t.Error(err)
		return
	}

	if output.ExpandItems[0].Path[0].Value != "Products" {
		t.Error("First path item is not 'Products'")
		return
	}

	if output.ExpandItems[0].Filter == nil {
		t.Error("Filter not parsed")
		return
	}

	if output.ExpandItems[0].Filter.Tree == nil {
		t.Error("Filter tree is null")
		return
	}

	if output.ExpandItems[0].Filter.Tree.Token.Value != "not" {
		actual := output.ExpandItems[0].Filter.Tree.Token.Value
		t.Error("Root filter value is '" + actual + "', expected 'not'")
		return
	}
}
