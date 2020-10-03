package godata

import (
	"net/url"
	"testing"
)

type DummyProvider struct {
}

func (*DummyProvider) GetEntity(*GoDataRequest) (*GoDataResponseField, error) {
	return nil, NotImplementedError("Dummy provider implements nothing.")
}

func (*DummyProvider) GetEntityCollection(*GoDataRequest) (*GoDataResponseField, error) {
	return nil, NotImplementedError("Dummy provider implements nothing.")
}

func (*DummyProvider) GetCount(*GoDataRequest) (int, error) {
	return 0, NotImplementedError("Dummy provider implements nothing.")
}

func (*DummyProvider) GetMetadata() *GoDataMetadata {

	metadata := &GoDataMetadata{
		DataServices: &GoDataServices{
			Schemas: []*GoDataSchema{
				&GoDataSchema{
					Namespace: "Store",
					EntityTypes: []*GoDataEntityType{
						&GoDataEntityType{
							Name: "Customer",
							Properties: []*GoDataProperty{
								&GoDataProperty{
									Name: "Name",
									Type: GoDataString,
								},
								&GoDataProperty{
									Name: "Age",
									Type: GoDataInt32,
								},
							},
							NavigationProperties: []*GoDataNavigationProperty{
								&GoDataNavigationProperty{
									Name:    "Orders",
									Type:    "Collection(Store.Order)",
									Partner: "Customer",
								},
							},
						},
						&GoDataEntityType{
							Name: "Order",
							Properties: []*GoDataProperty{
								&GoDataProperty{
									Name: "Id",
									Type: GoDataString,
								},
							},
							NavigationProperties: []*GoDataNavigationProperty{
								&GoDataNavigationProperty{
									Name:    "Customer",
									Type:    "Store.Customer",
									Partner: "Orders",
								},
							},
						},
					},
					EntityContainers: []*GoDataEntityContainer{
						&GoDataEntityContainer{
							Name: "Collections",
							EntitySets: []*GoDataEntitySet{
								&GoDataEntitySet{
									Name:       "Customers",
									EntityType: "Store.Customer",
									NavigationPropertyBindings: []*GoDataNavigationPropertyBinding{
										&GoDataNavigationPropertyBinding{
											Path:   "Orders",
											Target: "Orders",
										},
									},
								},
								&GoDataEntitySet{
									Name:       "Orders",
									EntityType: "Store.Order",
									NavigationPropertyBindings: []*GoDataNavigationPropertyBinding{
										&GoDataNavigationPropertyBinding{
											Path:   "Customer",
											Target: "Customer",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	return metadata
}

func TestSemanticizeRequest(t *testing.T) {
	provider := &DummyProvider{}

	testUrl := "Customers?$expand=Orders($expand=*($levels=1))&$filter=Name eq 'Bob'"

	url, err := url.Parse(testUrl)

	if err != nil {
		t.Error(err)
		return
	}

	req, err := ParseRequest(url.Path, url.Query(), false)

	if err != nil {
		t.Error(err)
		return
	}

	service, err := BuildService(provider, "http://localhost")

	if err != nil {
		t.Error(err)
		return
	}

	err = SemanticizeRequest(req, service)

	if err != nil {
		t.Error(err)
		return
	}

	if req.LastSegment.SemanticType != SemanticTypeEntitySet {
		t.Error("Request last segment semantic type is not SemanticTypeEntitySet")
		return
	}

	target := req.LastSegment.SemanticReference.(*GoDataEntitySet).Name
	if target != "Customers" {
		t.Error("Request last segment semantic reference name is '" + target + "' not 'Customers'")
		return
	}

	if req.Query.Expand.ExpandItems[0].Path[0].SemanticType != SemanticTypeEntity {
		t.Error("Request expand last segment is not SemanticTypeEntity")
		return
	}

	if req.Query.Expand.ExpandItems[0].Expand == nil {
		t.Error("Request second-level expand is nil!")
		return
	}

	target = req.Query.Expand.ExpandItems[0].Expand.ExpandItems[0].Path[0].Value
	if target != "Customer" {
		t.Error("Request second-level expand last segment is '" + target + "' not 'Customer'")
		return
	}

	target = req.Query.Expand.ExpandItems[0].Expand.ExpandItems[0].Expand.ExpandItems[0].Path[0].Value
	if target != "Orders" {
		t.Error("Request third-level expand last segment is '" + target + "' not 'Orders'")
		return
	}

}

func TestSemanticizeRequestWildcard(t *testing.T) {
	provider := &DummyProvider{}

	testUrl := "Customers?$expand=*($levels=2)&$filter=Name eq 'Bob'"

	url, err := url.Parse(testUrl)

	if err != nil {
		t.Error(err)
		return
	}

	req, err := ParseRequest(url.Path, url.Query(), false)

	if err != nil {
		t.Error(err)
		return
	}

	service, err := BuildService(provider, "http://localhost")

	if err != nil {
		t.Error(err)
		return
	}

	err = SemanticizeRequest(req, service)

	if err != nil {
		t.Errorf("Failed to semanticize request. Error: %v", err)
		return
	}

	if req.LastSegment.SemanticType != SemanticTypeEntitySet {
		t.Error("Request last segment semantic type is not SemanticTypeEntitySet")
		return
	}

	target := req.LastSegment.SemanticReference.(*GoDataEntitySet).Name
	if target != "Customers" {
		t.Error("Request last segment semantic reference name is '" + target + "' not 'Customers'")
		return
	}

	if req.Query.Expand.ExpandItems[0].Path[0].SemanticType != SemanticTypeEntity {
		t.Error("Request expand last segment is not SemanticTypeEntity")
		return
	}

	if req.Query.Expand.ExpandItems[0].Expand == nil {
		t.Error("Request second-level expand is nil!")
		return
	}

	target = req.Query.Expand.ExpandItems[0].Expand.ExpandItems[0].Path[0].Value
	if target != "Customer" {
		t.Error("Request second-level expand last segment is '" + target + "' not 'Customer'")
		return
	}

	target = req.Query.Expand.ExpandItems[0].Expand.ExpandItems[0].Expand.ExpandItems[0].Path[0].Value
	if target != "Orders" {
		t.Error("Request third-level expand last segment is '" + target + "' not 'Orders'")
		return
	}

}

func BenchmarkBuildProvider(b *testing.B) {
	for n := 0; n < b.N; n++ {
		provider := &DummyProvider{}

		_, err := BuildService(provider, "http://localhost")

		if err != nil {
			b.Error(err)
			return
		}
	}
}

func BenchmarkTypicalParseSemanticizeRequest(b *testing.B) {
	provider := &DummyProvider{}

	testUrl := "Customers?$expand=Orders&$filter=Name eq 'Bob'"

	url, err := url.Parse(testUrl)

	if err != nil {
		b.Error(err)
		return
	}

	for n := 0; n < b.N; n++ {

		req, err := ParseRequest(url.Path, url.Query(), false)

		if err != nil {
			b.Error(err)
			return
		}

		service, err := BuildService(provider, "http://localhost")

		if err != nil {
			b.Error(err)
			return
		}

		err = SemanticizeRequest(req, service)

		if err != nil {
			b.Error(err)
			return
		}
	}
}

func BenchmarkWildcardParseSemanticizeRequest(b *testing.B) {
	provider := &DummyProvider{}

	testUrl := "Customers?$expand=*($levels=2)&$filter=Name eq 'Bob'"

	url, err := url.Parse(testUrl)

	if err != nil {
		b.Error(err)
		return
	}

	for n := 0; n < b.N; n++ {

		req, err := ParseRequest(url.Path, url.Query(), false)

		if err != nil {
			b.Error(err)
			return
		}

		service, err := BuildService(provider, "http://localhost")

		if err != nil {
			b.Error(err)
			return
		}

		err = SemanticizeRequest(req, service)

		if err != nil {
			b.Error(err)
			return
		}
	}
}
