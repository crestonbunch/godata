package godata

import (
	"testing"
)

func TestSimpleMetadata(t *testing.T) {

	entity1 := GoDataEntityType{
		Name: "TestEntity1",
		Key:  &GoDataKey{PropertyRef: &GoDataPropertyRef{Name: "Id"}},
		Properties: []*GoDataProperty{
			&GoDataProperty{Name: "Id", Type: "Edm.Int32"},
			&GoDataProperty{Name: "FirstName", Type: "Edm.String"},
			&GoDataProperty{Name: "LastName", Type: "Edm.String"},
		},
	}

	schema := GoDataSchema{
		Namespace:   "TestSchema",
		EntityTypes: []*GoDataEntityType{&entity1},
	}

	services := GoDataServices{
		Schemas: []*GoDataSchema{&schema},
	}

	root := GoDataMetadata{
		XMLNamespace: "http://docs.oasis-open.org/odata/ns/edmx",
		Version:      "4.0",
		DataServices: &services,
	}

	expected := "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n"
	expected += "<edmx:Edmx xmlns:edmx=\"http://docs.oasis-open.org/odata/ns/edmx\" Version=\"4.0\">\n"
	expected += "    <edmx:DataServices>\n"
	expected += "        <Schema Namespace=\"TestSchema\">\n"
	expected += "            <EntityType Name=\"TestEntity1\">\n"
	expected += "                <Key>\n"
	expected += "                    <PropertyRef Name=\"Id\"></PropertyRef>\n"
	expected += "                </Key>\n"
	expected += "                <Property Name=\"Id\" Type=\"Edm.Int32\"></Property>\n"
	expected += "                <Property Name=\"FirstName\" Type=\"Edm.String\"></Property>\n"
	expected += "                <Property Name=\"LastName\" Type=\"Edm.String\"></Property>\n"
	expected += "            </EntityType>\n"
	expected += "        </Schema>\n"
	expected += "    </edmx:DataServices>\n"
	expected += "</edmx:Edmx>"

	actual, err := root.Bytes()

	if err != nil {
		t.Error(err)
	}

	if string(actual) != expected {
		t.Error("Expected: \n"+expected, "\n\nGot: \n"+string(actual))
	}
}
