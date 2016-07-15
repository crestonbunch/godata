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

	request, err := ParseRequest(parsedUrl.Path, parsedUrl.Query())

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
