package godata

import (
	"net/http"
	"strings"
)

// The basic interface for a GoData provider. All providers must implement
// these functions.
type GoDataProvider interface {
	Response(*GoDataRequest) *GoDataResponse
	BuildMetadata() *GoDataMetadata
}

// A GoDataService will spawn an HTTP listener, which will connect GoData
// requests with a backend provider given to it.
type GoDataService struct {
	Provider GoDataProvider
	Metadata *GoDataMetadata
	// A mapping from schema names to schema references
	SchemaLookup map[string]*GoDataSchema
	// A bottom-up mapping from entity type names to schema namespaces to
	// the entity type reference
	EntityTypeLookup map[string]map[string]*GoDataEntityType
	// A bottom-up mapping from entity container names to schema namespaces to
	// the entity container reference
	EntityContainerLookup map[string]map[string]*GoDataEntityContainer
	// A bottom-up mapping from entity set names to entity collecton names to
	// schema namespaces to the entity set reference
	EntitySetLookup map[string]map[string]map[string]*GoDataEntitySet
	// A lookup for entity properties if an entity type is given, lookup
	// properties by name
	PropertyLookup map[*GoDataEntityType]map[string]*GoDataProperty
	// A lookup for navigational properties if an entity type is given,
	// lookup navigational properties by name
	NavigationPropertyLookup map[*GoDataEntityType]map[string]*GoDataNavigationProperty
}

// Create a new service from a given provider. This step builds lookups for
// all parts of the data model, so constant time lookups can be performed. This
// step only happens once when the server starts up, so the overall cost is
// minimal
func BuildService(provider GoDataProvider) *GoDataService {
	metadata := provider.BuildMetadata()

	// build the lookups from the metadata
	schemaLookup := map[string]*GoDataSchema{}
	entityLookup := map[string]map[string]*GoDataEntityType{}
	containerLookup := map[string]map[string]*GoDataEntityContainer{}
	entitySetLookup := map[string]map[string]map[string]*GoDataEntitySet{}
	propertyLookup := map[*GoDataEntityType]map[string]*GoDataProperty{}
	navPropLookup := map[*GoDataEntityType]map[string]*GoDataNavigationProperty{}

	for _, schema := range metadata.DataServices.Schemas {
		schemaLookup[schema.Namespace] = schema

		for _, entity := range schema.EntityTypes {
			if _, ok := entityLookup[entity.Name]; !ok {
				entityLookup[entity.Name] = map[string]*GoDataEntityType{}
			}
			if _, ok := propertyLookup[entity]; !ok {
				propertyLookup[entity] = map[string]*GoDataProperty{}
			}
			if _, ok := navPropLookup[entity]; !ok {
				navPropLookup[entity] = map[string]*GoDataNavigationProperty{}
			}
			entityLookup[entity.Name][schema.Namespace] = entity

			for _, prop := range entity.Properties {
				propertyLookup[entity][prop.Name] = prop
			}
			for _, prop := range entity.NavigationProperties {
				navPropLookup[entity][prop.Name] = prop
			}
		}

		for _, container := range schema.EntityContainers {
			if _, ok := containerLookup[container.Name]; !ok {
				containerLookup[container.Name] = map[string]*GoDataEntityContainer{}
			}
			containerLookup[container.Name][schema.Namespace] = container

			for _, set := range container.EntitySets {
				if _, ok := entitySetLookup[set.Name]; !ok {
					entitySetLookup[set.Name] = map[string]map[string]*GoDataEntitySet{}
				}
				if _, ok := entitySetLookup[set.Name][container.Name]; !ok {
					entitySetLookup[set.Name][container.Name] = map[string]*GoDataEntitySet{}
				}
				entitySetLookup[set.Name][container.Name][schema.Namespace] = set
			}
		}
	}

	return &GoDataService{
		provider,
		provider.BuildMetadata(),
		schemaLookup,
		entityLookup,
		containerLookup,
		entitySetLookup,
		propertyLookup,
		navPropLookup,
	}
}

// The default handler for parsing requests as GoDataRequests, passing them
// to a GoData provider, and then building a response.
func (service *GoDataService) GoDataHTTPHandler(w http.ResponseWriter, r *http.Request) {

	request, err := ParseRequest(r.URL.Path, r.URL.Query())

	if err != nil {
		panic(err) // TODO: return proper error
	}

	// Semanticize all tokens in the request, connecting them with their
	// corresponding types in the service
	err = SemanticizeRequest(request, service)

	if err != nil {
		panic(err) // TODO: return proper error
	}

	// TODO: differentiate GET and POST requests
	var response []byte = []byte{}
	if request.RequestKind == RequestKindMetadata {
		response, err = service.buildMetadataResponse(request)
	} else if request.RequestKind == RequestKindService {
		response, err = service.buildServiceResponse(request)
	} else if request.RequestKind == RequestKindCollection {
		response, err = service.buildCollectionResponse(request)
	} else if request.RequestKind == RequestKindEntity {
		response, err = service.buildEntityResponse(request)
	} else if request.RequestKind == RequestKindProperty {
		response, err = service.buildPropertyResponse(request)
	} else if request.RequestKind == RequestKindPropertyValue {
		response, err = service.buildPropertyValueResponse(request)
	} else if request.RequestKind == RequestKindCount {
		response, err = service.buildCountResponse(request)
	} else if request.RequestKind == RequestKindRef {
		response, err = service.buildRefResponse(request)
	} else {
		err = NotImplementedError("Request type not understood.")
	}

	if err != nil {
		panic(err) // TODO: return proper error
	}

	w.Write(response)
}

func (service *GoDataService) buildMetadataResponse(request *GoDataRequest) ([]byte, error) {
	return service.Metadata.Bytes()
}

func (service *GoDataService) buildServiceResponse(request *GoDataRequest) ([]byte, error) {
	return nil, NotImplementedError("Service responses are not implemented yet.")
}

func (service *GoDataService) buildCollectionResponse(request *GoDataRequest) ([]byte, error) {
	return nil, NotImplementedError("Collection responses are not implemented yet.")
}

func (service *GoDataService) buildEntityResponse(request *GoDataRequest) ([]byte, error) {
	return nil, NotImplementedError("Entity responses are not implemented yet.")
}

func (service *GoDataService) buildPropertyResponse(request *GoDataRequest) ([]byte, error) {
	return nil, NotImplementedError("Property responses are not implemented yet.")
}

func (service *GoDataService) buildPropertyValueResponse(request *GoDataRequest) ([]byte, error) {
	return nil, NotImplementedError("Property value responses are not implemented yet.")
}

func (service *GoDataService) buildCountResponse(request *GoDataRequest) ([]byte, error) {
	return nil, NotImplementedError("Count responses are not implemented yet.")
}

func (service *GoDataService) buildRefResponse(request *GoDataRequest) ([]byte, error) {
	return nil, NotImplementedError("Ref responses are not implemented yet.")
}

// Start the service listening on the given address.
func (service *GoDataService) ListenAndServe(addr string) {
	http.HandleFunc("/", service.GoDataHTTPHandler)
	http.ListenAndServe(addr, nil)
}

// Lookup an entity type from the service metadata. Accepts a fully qualified
// name, e.g., ODataService.EntityTypeName or, if unambiguous, accepts a
// simple identifier, e.g., EntityTypeName.
func (service *GoDataService) LookupEntityType(name string) (*GoDataEntityType, error) {
	// strip "Collection()" and just return the raw entity type
	// The provider should keep track of what are collections and what aren't
	if strings.Contains(name, "(") && strings.Contains(name, ")") {
		name = name[strings.Index(name, "(")+1 : strings.LastIndex(name, ")")]
	}

	parts := strings.Split(name, ".")
	entityName := parts[len(parts)-1]
	// remove entity from the list of parts
	parts = parts[:len(parts)-1]

	schemas, ok := service.EntityTypeLookup[entityName]
	if !ok {
		return nil, BadRequestError("Entity " + name + " does not exist.")
	}

	if len(parts) > 0 {
		// namespace is provided
		entity, ok := schemas[parts[len(parts)-1]]
		if !ok {
			return nil, BadRequestError("Entity " + name + " not found in given namespace.")
		}
		return entity, nil
	} else {
		// no namespace, just return the first one

		// throw error if ambiguous
		if len(schemas) > 1 {
			return nil, BadRequestError("Entity " + name + " is ambiguous. Please provide a namespace.")
		}

		for _, v := range schemas {
			return v, nil
		}
	}

	// If this happens, that's very bad
	return nil, BadRequestError("No schema lookup found for entity " + name)
}

// Lookup an entity set from the service metadata. Accepts a fully qualified
// name, e.g., ODataService.ContainerName.EntitySetName,
// ContainerName.EntitySetName or, if unambiguous, accepts a  simple identifier,
// e.g., EntitySetName.
func (service *GoDataService) LookupEntitySet(name string) (*GoDataEntitySet, error) {
	parts := strings.Split(name, ".")
	setName := parts[len(parts)-1]
	// remove entity set from the list of parts
	parts = parts[:len(parts)-1]

	containers, ok := service.EntitySetLookup[setName]
	if !ok {
		return nil, BadRequestError("Entity set " + name + " does not exist.")
	}

	if len(parts) > 0 {
		// container is provided
		schemas, ok := containers[parts[len(parts)-1]]
		if !ok {
			return nil, BadRequestError("Container " + name + " not found.")
		}

		// remove container name from the list of parts
		parts = parts[:len(parts)-1]

		if len(parts) > 0 {
			// schema is provided
			set, ok := schemas[parts[len(parts)-1]]
			if !ok {
				return nil, BadRequestError("Entity set " + name + " not found.")
			}
			return set, nil
		} else {
			// no schema is provided

			if len(schemas) > 1 {
				// container is ambiguous
				return nil, BadRequestError("Entity set " + name + " is ambiguous. Please provide fully qualified name.")
			}

			// there should be one schema, if not then something went horribly wrong
			for _, set := range schemas {
				return set, nil
			}
		}

	} else {
		// no container is provided

		// return error if entity set is ambiguous
		if len(containers) > 1 {
			return nil, BadRequestError("Entity set " + name + " is ambiguous. Please provide fully qualified name.")
		}

		// find the first schema, it will be the only one
		for _, schemas := range containers {
			if len(schemas) > 1 {
				// container is ambiguous
				return nil, BadRequestError("Entity set " + name + " is ambiguous. Please provide fully qualified name.")
			}

			// there should be one schema, if not then something went horribly wrong
			for _, set := range schemas {
				return set, nil
			}
		}
	}
	return nil, BadRequestError("Entity set " + name + " not found.")
}
