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

	// Duplicate keywords
	testUrl = "Employees(1)/Sales.Manager?$select=3DFirstName&$select=3DFirstName"
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

}
