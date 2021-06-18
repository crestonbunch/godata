package godata

import (
	"net/url"
	"testing"
)

func TestUrlParser(t *testing.T) {
	testUrl := "Employees(1)/Sales.Manager?$expand=DirectReports%28$select%3DFirstName%2CLastName%3B$levels%3D4%29"
	parsedUrl, err := url.Parse(testUrl)

	if err != nil {
		t.Error(err)
		return
	}

	request, err := ParseRequest(parsedUrl.Path, parsedUrl.Query(), false)

	if err != nil {
		t.Error(err)
		return
	}

	if request.FirstSegment.Name != "Employees" {
		t.Error("First segment is '" + request.FirstSegment.Name + "' not Employees")
		return
	}
	if request.FirstSegment.Identifier.Get() != "1" {
		t.Error("Employee identifier not found")
		return
	}
	if request.FirstSegment.Next.Name != "Sales.Manager" {
		t.Error("Second segment is not Sales.Manager")
		return
	}
}

func TestUrlParserStrictValidation(t *testing.T) {
	testUrl := "Employees(1)/Sales.Manager?$expand=DirectReports%28$select%3DFirstName%2CLastName%3B$levels%3D4%29"
	parsedUrl, err := url.Parse(testUrl)
	if err != nil {
		t.Error(err)
		return
	}
	_, err = ParseRequest(parsedUrl.Path, parsedUrl.Query(), false)
	if err != nil {
		t.Error(err)
		return
	}

	testUrl = "Employees(1)/Sales.Manager?$filter=FirstName eq 'Bob'"
	parsedUrl, err = url.Parse(testUrl)
	if err != nil {
		t.Error(err)
		return
	}
	_, err = ParseRequest(parsedUrl.Path, parsedUrl.Query(), false)
	if err != nil {
		t.Error(err)
		return
	}

	// Wrong filter with an extraneous single quote
	testUrl = "Employees(1)/Sales.Manager?$filter=FirstName eq' 'Bob'"
	parsedUrl, err = url.Parse(testUrl)
	if err != nil {
		t.Error(err)
		return
	}
	_, err = ParseRequest(parsedUrl.Path, parsedUrl.Query(), false)
	if err == nil {
		t.Errorf("Parser should have returned invalid filter error: %s", testUrl)
		return
	}

	// Valid query with two parameters:
	// $filter=FirstName eq 'Bob'
	// at=Version eq '123'
	testUrl = "Employees(1)/Sales.Manager?$filter=FirstName eq 'Bob'&at=Version eq '123'"
	parsedUrl, err = url.Parse(testUrl)
	if err != nil {
		t.Error(err)
		return
	}
	_, err = ParseRequest(parsedUrl.Path, parsedUrl.Query(), false)
	if err != nil {
		t.Error(err)
		return
	}

	// Invalid query:
	// $filter=FirstName eq' 'Bob' has extraneous single quote.
	// at=Version eq '123'         is valid
	testUrl = "Employees(1)/Sales.Manager?$filter=FirstName eq' 'Bob'&at=Version eq '123'"
	parsedUrl, err = url.Parse(testUrl)
	if err != nil {
		t.Error(err)
		return
	}
	_, err = ParseRequest(parsedUrl.Path, parsedUrl.Query(), false)
	if err == nil {
		t.Errorf("Parser should have returned invalid filter error: %s", testUrl)
		return
	}

	testUrl = "Employees(1)/Sales.Manager?$select=3DFirstName"
	parsedUrl, err = url.Parse(testUrl)
	if err != nil {
		t.Error(err)
		return
	}
	_, err = ParseRequest(parsedUrl.Path, parsedUrl.Query(), false)
	if err != nil {
		t.Error(err)
		return
	}

	testUrl = "Employees(1)/Sales.Manager?$filter=Name in ('Bob','Alice')&$select=Name,Address%3B$expand=Address($select=City)"
	parsedUrl, err = url.Parse(testUrl)
	if err != nil {
		t.Error(err)
		return
	}
	_, err = ParseRequest(parsedUrl.Path, parsedUrl.Query(), false /*strict*/)
	if err != nil {
		t.Errorf("Unexpected parsing error: %v", err)
		return
	}

	// A $select option cannot be wrapped with parenthesis. This is not legal ODATA.

	/*
		 queryOptions = queryOption *( "&" queryOption )
		 queryOption  = systemQueryOption
				/ aliasAndValue
				/ nameAndValue
				/ customQueryOption
		 systemQueryOption = compute
				/ deltatoken
				/ expand
				/ filter
				/ format
				/ id
				/ inlinecount
				/ orderby
				/ schemaversion
				/ search
				/ select
				/ skip
				/ skiptoken
				/ top
				/ index
		  select = ( "$select" / "select" ) EQ selectItem *( COMMA selectItem )
	*/
	testUrl = "Employees(1)/Sales.Manager?$filter=Name in ('Bob','Alice')&($select=Name,Address%3B$expand=Address($select=City))"
	parsedUrl, err = url.Parse(testUrl)
	if err != nil {
		t.Error(err)
		return
	}
	_, err = ParseRequest(parsedUrl.Path, parsedUrl.Query(), false /*strict*/)
	if err == nil {
		t.Errorf("Parser should have raised error")
		return
	}

	// Duplicate keyword: '$select' is present twice.
	testUrl = "Employees(1)/Sales.Manager?$select=3DFirstName&$select=3DFirstName"
	parsedUrl, err = url.Parse(testUrl)
	if err != nil {
		t.Error(err)
		return
	}
	// In lenient mode, do not return an error when there is a duplicate keyword.
	_, err = ParseRequest(parsedUrl.Path, parsedUrl.Query(), true /*lenient*/)
	if err != nil {
		t.Error(err)
		return
	}
	// In strict mode, return an error when there is a duplicate keyword.
	_, err = ParseRequest(parsedUrl.Path, parsedUrl.Query(), false /*strict*/)
	if err == nil {
		t.Error("Parser should have returned duplicate keyword error")
		return
	}

	// Unsupported keywords
	testUrl = "Employees(1)/Sales.Manager?orderby=FirstName"
	parsedUrl, err = url.Parse(testUrl)
	if err != nil {
		t.Error(err)
		return
	}
	_, err = ParseRequest(parsedUrl.Path, parsedUrl.Query(), true /*lenient*/)
	if err != nil {
		t.Error(err)
		return
	}
	_, err = ParseRequest(parsedUrl.Path, parsedUrl.Query(), false /*strict*/)
	if err == nil {
		t.Error("Parser should have returned unsupported keyword error")
		return
	}

	testUrl = "Employees(1)/Sales.Manager?$select=LastName&$expand=Address"
	parsedUrl, err = url.Parse(testUrl)
	if err != nil {
		t.Error(err)
		return
	}
	_, err = ParseRequest(parsedUrl.Path, parsedUrl.Query(), false)
	if err != nil {
		t.Error(err)
		return
	}

	testUrl = "Employees(1)/Sales.Manager?$select=FirstName,LastName&$expand=Address"
	parsedUrl, err = url.Parse(testUrl)
	if err != nil {
		t.Error(err)
		return
	}
	_, err = ParseRequest(parsedUrl.Path, parsedUrl.Query(), false)
	if err != nil {
		t.Error(err)
		return
	}

}
