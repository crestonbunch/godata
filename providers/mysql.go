package mysql

import (
	. "godata"
	//"database/sql"
	//"errors"
	//"github.com/go-sql-driver/mysql"
	//"strings"
)

const (
	XMLNamespace string = "http://docs.oasis-open.org/odata/ns/edmx"
	ODataVersion string = "4.0"
)

var MySQLPrepareMap = map[string]string{
	// wrap the input string in wildcard characters for LIKE clauses
	"contains":   "%?%",
	"endswith":   "%?",
	"startswith": "?%",
}

var MySQLNodeMap = map[string]string{
	"eq":  "(%s = %s)",
	"nq":  "(%s != %s)",
	"gt":  "(%s > %s)",
	"ge":  "(%s >= %s)",
	"lt":  "(%s < %s)",
	"le":  "(%s <= %s)",
	"and": "(%s AND %s)",
	"or":  "(%s OR %s)",
	"not": "(NOT %s)",
	// "has": ""
	//"add": ""
	//"sub": ""
	//"mul": ""
	//"div": ""
	//"mod": ""
	"contains":   "(%s LIKE %s)",
	"endswith":   "(%s LIKE %s)",
	"startswith": "(%s LIKE %s)",
	"length":     "LENGTH(%s)",
	"indexof":    "LOCATE(%s)",
	//"substring": "",
	"tolower":          "LOWER(%s)",
	"toupper":          "UPPER(%s)",
	"trim":             "TRIM(%s)",
	"concat":           "CONCAT(%s,%s)",
	"year":             "YEAR(%s)",
	"month":            "MONTH(%s)",
	"day":              "DAY(%s)",
	"hour":             "HOUR(%s)",
	"minute":           "MINUTE(%s)",
	"second":           "SECOND(%s)",
	"fractionalsecond": "MICROSECOND(%s)",
	"date":             "DATE(%s)",
	"time":             "TIME(%s)",
	//"totaloffsetminutes": "",
	"now": "NOW()",
	//"maxdatetime":"",
	//"mindatetime":"",
	//"totalseconds":"",
	"round":   "ROUND(%s)",
	"floor":   "FLOOR(%s)",
	"ceiling": "CEIL(%s)",
	//"isof": "",
	//"cast": "",
	//"geo.distance": "",
	//"geo.intersects": "",
	//"geo.length": "",
	//"any": "",
	//"all": "",
	"null": "NULL",
}

// Struct to hold MySQL connection parameters.
type MySQLConnectionParams struct {
	Database string
	Hostname string
	Port     string
	Username string
	Password string
}

// A provider for GoData using a MySQL backend. Reads requests, converts them
// to MySQL queries, and creates a response object to send back to the client.
type MySQLGoDataProvider struct {
	ConnectionParams *MySQLConnectionParams
	Namespace        string
	Entities         map[string]*MySQLGoDataEntity
	EntitySets       map[string]*MySQLGoDataEntitySet
	Actions          map[string]*GoDataAction
	Functions        map[string]*GoDataFunction
	Metadata         *GoDataMetadata
}

type MySQLGoDataEntity struct {
	TableName  string
	KeyType    string
	PropColMap map[string]string
	ColPropMap map[string]string
	EntityType *GoDataEntityType
}

type MySQLGoDataEntitySet struct {
	Entity    *MySQLGoDataEntity
	EntitySet *GoDataEntitySet
}

// Build an empty MySQL provider. Provide the connection parameters and the
// namespace name.
func BuildMySQLProvider(cxnParams *MySQLConnectionParams, namespace string) *MySQLGoDataProvider {
	return &MySQLGoDataProvider{
		ConnectionParams: cxnParams,
		Entities:         make(map[string]*MySQLGoDataEntity),
		EntitySets:       make(map[string]*MySQLGoDataEntitySet),
	}
}

func (p *MySQLGoDataProvider) BuildQuery(r *GoDataRequest) (string, error) {
	/*
		setName := r.FirstSegment.Name
		entitySet := p.EntitySets[setName]
		tableName := entitySet.Entity.TableName
		pKeyValue := r.FirstSegment.Identifier

		query := []byte{}
		params := []string{}

		selectClause, selectParams, selectErr := p.BuildSelectClause(r)
		if selectErr != nil {
			return nil, selectErr
		}
		query = append(query, selectClause)
		params = append(params, selectParams)

		fromClause, fromParams, fromErr := p.BuildFromClause(r)
		if fromErr != nil {
			return nil, fromErr
		}
		query = append(query, selectClause)
		params = append(params, selectParams)
	*/
	return "", NotImplementedError("not implemented")
}

// Build the select clause to begin the query, and also return the values to
// send to a prepared statement.
func (p *MySQLGoDataProvider) BuildSelectClause(r *GoDataRequest) ([]byte, []string, error) {
	return nil, nil, NotImplementedError("not implemented")
}

// Build the from clause in the query, and also return the values to send to
// the prepared statement.
func (p *MySQLGoDataProvider) BuildFromClause(r *GoDataRequest) ([]byte, []string, error) {
	return nil, nil, NotImplementedError("not implemented")
}

// Build a where clause that can be appended to an SQL query, and also return
// the values to send to a prepared statement.
func (p *MySQLGoDataProvider) BuildWhereClause(r *GoDataRequest) ([]byte, []string, error) {
	/*
		// Builds the WHERE clause recursively using DFS
		recursiveBuildWhere := func(n *ParseNode) ([]byte, []string, error) {
			if n.Token.Type == FilterTokenLiteral {
				// TODO: map to columns
				return []byte("?"), []byte(n.Token.Value), nil
			}

			if v, ok := MySQLNodeMap[n.Token.Value]; ok {
				params := string
				children := []byte{}
				// build each child first using DFS
				for _, child := range n.Children {
					q, o, err := recursiveBuildWhere(child)
					if err != nil {
						return nil, nil, err
					}
					children := append(children, q)
					// make the assumption that all params appear LTR and are never added
					// out of order
					params := append(params, o)
				}
				// merge together the children and the current node
				result := fmt.Sprintf(v, children...)
				return []byte(result), params, nil
			} else {
				return nil, nil, NotImplementedError(n.Token.Value + " is not implemented.")
			}
		}
	*/
	return nil, nil, NotImplementedError("not implemented")
}

// Respond to a GoDataRequest using the MySQL provider.
func (p *MySQLGoDataProvider) Response(r *GoDataRequest) *GoDataResponse {

	return nil
}

// Build the $metadata file from the entities in the builder. It creates a
// schema with the given namespace name.
func (builder *MySQLGoDataProvider) BuildMetadata() *GoDataMetadata {
	// convert maps to slices for the metadata
	entitySets := make([]*GoDataEntitySet, len(builder.EntitySets))
	for _, v := range builder.EntitySets {
		entitySets = append(entitySets, v.EntitySet)
	}
	entityTypes := make([]*GoDataEntityType, len(builder.Entities))
	for _, v := range builder.Entities {
		entityTypes = append(entityTypes, v.EntityType)
	}
	// build the schema
	container := GoDataEntityContainer{
		Name:       builder.ConnectionParams.Database,
		EntitySets: entitySets,
	}
	schema := GoDataSchema{
		Namespace:        builder.Namespace,
		EntityTypes:      entityTypes,
		EntityContainers: []*GoDataEntityContainer{&container},
	}
	services := GoDataServices{
		Schemas: []*GoDataSchema{&schema},
	}

	root := GoDataMetadata{
		XMLNamespace: XMLNamespace,
		Version:      ODataVersion,
		DataServices: &services,
	}

	return &root
}

// Expose a table in the MySQL database as an entity with the given name in the
// OData service.
func (builder *MySQLGoDataProvider) ExposeEntity(tblname, entityname string) *MySQLGoDataEntity {
	entitytype := GoDataEntityType{Name: entityname}
	myentity := &MySQLGoDataEntity{tblname, "", map[string]string{}, map[string]string{}, &entitytype}
	builder.Entities[entityname] = myentity
	return myentity
}

// Expose a queryable collection of entities
func (builder *MySQLGoDataProvider) ExposeEntitySet(entity *MySQLGoDataEntity, setname string) *MySQLGoDataEntitySet {
	entityset := &GoDataEntitySet{Name: setname, EntityType: builder.Namespace + "." + entity.EntityType.Name}
	myset := &MySQLGoDataEntitySet{entity, entityset}
	builder.EntitySets[setname] = myset
	return myset
}

// Adds the necessary NavigationProperty tags to entities to expose a one-to-one
// relationship from a (One) -> b (One). A column name and corresponding
// property name must be provided for each entity. The columns must be foreign
// keys corresponding to the primary key in the opposite table.
func (builder *MySQLGoDataProvider) ExposeOneToOne(a, b *MySQLGoDataEntity, acol, bcol, aprop, bprop string) {
	prop := GoDataNavigationProperty{Name: aprop, Type: builder.Namespace + "." + b.EntityType.Name, Partner: bprop}
	a.EntityType.NavigationProperties = append(a.EntityType.NavigationProperties, &prop)
	prop2 := GoDataNavigationProperty{Name: bprop, Type: builder.Namespace + "." + a.EntityType.Name, Partner: aprop}
	b.EntityType.NavigationProperties = append(b.EntityType.NavigationProperties, &prop2)
	a.PropColMap[aprop] = acol
	a.ColPropMap[acol] = aprop
	b.PropColMap[bprop] = bcol
	b.ColPropMap[bcol] = bprop
}

// Adds the necessary NavigationProperty tags to an entity to expose a many-to-one
// relationship from a (Many) -> b (One). A column name must be provided for entity a
// which will map to the key in entity b. A reverse property will be added to
// entity b to map back to entity a, and does not need an explicit column.
func (builder *MySQLGoDataProvider) ExposeManyToOne(a, b *MySQLGoDataEntity, acol, aprop, bprop string) {
	prop := GoDataNavigationProperty{Name: aprop, Type: builder.Namespace + "." + b.EntityType.Name, Partner: bprop}
	a.EntityType.NavigationProperties = append(a.EntityType.NavigationProperties, &prop)
	prop2 := GoDataNavigationProperty{Name: bprop, Type: "Collection(" + builder.Namespace + "." + a.EntityType.Name + ")", Partner: aprop}
	b.EntityType.NavigationProperties = append(b.EntityType.NavigationProperties, &prop2)
	// the property name corresponding to the column in a will be given the same name
	// as the key property in b so that it does not conflict with the property name
	// given by aprop. A referential constraint will be added to the NavigationProperty
	// in b that links back to this property in a.
	constrainedProp := b.EntityType.Key.PropertyRef.Name
	a.ExposeProperty(acol, constrainedProp, b.KeyType)
	constraint := GoDataReferentialConstraint{Property: constrainedProp, ReferencedProperty: constrainedProp}
	prop2.ReferentialConstraints = append(prop2.ReferentialConstraints, &constraint)
	a.PropColMap[aprop] = acol
	a.ColPropMap[acol] = aprop
}

// Adds the necessary NavigationProperty tags to entities to expose a many-to-many
// relationship from a (Many) -> b (Many). A third table must be provided that has
// foreign key mappings to the primary keys in both a & b. Both entities will
// be given a reference to each other with the given names in aprop & bprop.
func (builder *MySQLGoDataProvider) ExposeManyToMany(a, b *MySQLGoDataEntity, tblname, aprop, bprop string) {
	prop := GoDataNavigationProperty{Name: aprop, Type: "Collection(" + builder.Namespace + "." + b.EntityType.Name + ")", Partner: bprop}
	a.EntityType.NavigationProperties = append(a.EntityType.NavigationProperties, &prop)
	prop2 := GoDataNavigationProperty{Name: bprop, Type: "Collection(" + builder.Namespace + "." + a.EntityType.Name + ")", Partner: aprop}
	b.EntityType.NavigationProperties = append(b.EntityType.NavigationProperties, &prop2)
}

// Adds a NavigationPropertyBinding to two entity sets that are mapped together
// by a relationship. This SHOULD be done for any entity sets for whom their
// entities contain a NavigationProperty.
func (builder *MySQLGoDataProvider) BindProperty(a, b *MySQLGoDataEntitySet, apath, atarget, bpath, btarget string) {
	bind := GoDataNavigationPropertyBinding{Path: apath, Target: atarget}
	a.EntitySet.NavigationPropertyBindings = append(a.EntitySet.NavigationPropertyBindings, &bind)
	bind2 := GoDataNavigationPropertyBinding{Path: bpath, Target: btarget}
	b.EntitySet.NavigationPropertyBindings = append(b.EntitySet.NavigationPropertyBindings, &bind2)
}

// Expose a key on an entity returned by MySQLGoDataProvider.ExposeEntity.
// This is a necessary step for every entity. You must provide a column in the
// database to map to the property name in the OData entity, and the OData
// type.
func (entity *MySQLGoDataEntity) ExposeKey(colname, propname, t string) {
	entity.EntityType.Key = &GoDataKey{PropertyRef: &GoDataPropertyRef{Name: propname}}
	entity.KeyType = t
	entity.ExposePrimitive(colname, propname, t)
}

// Expose an OData primitive property on an entity. You must provide a
// corresponding table column in the database.
func (entity *MySQLGoDataEntity) ExposePrimitive(colname, propname, t string) {
	entity.ExposeProperty(colname, propname, t)
}

// Expose an OData property on an entity. You must provide a
// corresponding table column in the database.
func (entity *MySQLGoDataEntity) ExposeProperty(colname, propname, t string) {
	entity.EntityType.Properties = append(entity.EntityType.Properties, &GoDataProperty{Name: propname, Type: t})
	entity.PropColMap[propname] = colname
	entity.ColPropMap[colname] = propname
}
