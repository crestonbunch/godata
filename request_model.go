package godata

type GoDataIdentifier map[string]string

const (
	RequestKindUnknown int = iota
	RequestKindMetadata
	RequestKindService
	RequestKindEntity
	RequestKindCollection
	RequestKindSingleton
	RequestKindProperty
	RequestKindPropertyValue
	RequestKindRef
	RequestKindCount
)

const (
	SemanticTypeUnknown int = iota
	SemanticTypeEntity
	SemanticTypeEntitySet
	SemanticTypeDerivedEntity
	SemanticTypeAction
	SemanticTypeFunction
	SemanticTypeProperty
	SemanticTypePropertyValue
	SemanticTypeRef
	SemanticTypeCount
	SemanticTypeMetadata
)

type GoDataRequest struct {
	FirstSegment *GoDataSegment
	LastSegment  *GoDataSegment
	Query        *GoDataQuery
	RequestKind  int
}

// Represents a segment (slash-separated) part of the URI path. Each segment
// has a link to the next segment (the last segment precedes nil).
type GoDataSegment struct {
	// The raw segment parsed from the URI
	RawValue string

	// The kind of resource being pointed at by this segment
	SemanticType int

	// A pointer to the metadata type this object represents, as defined by a
	// particular service
	SemanticReference interface{}

	// The name of the entity, type, collection, etc.
	Name string

	// map[string]string of identifiers passed to this segment. If the identifier
	// is not key/value pair(s), then all values will be nil. If there is no
	// identifier, it will be nil.
	Identifier *GoDataIdentifier

	// The next segment in the path.
	Next *GoDataSegment
	// The previous segment in the path.
	Prev *GoDataSegment
}

type GoDataQuery struct {
	Filter      *GoDataFilterQuery
	At          *GoDataFilterQuery
	Apply       *GoDataApplyQuery
	Expand      *GoDataExpandQuery
	Select      *GoDataSelectQuery
	OrderBy     *GoDataOrderByQuery
	Top         *GoDataTopQuery
	Skip        *GoDataSkipQuery
	Count       *GoDataCountQuery
	InlineCount *GoDataInlineCountQuery
	Search      *GoDataSearchQuery
	Format      *GoDataFormatQuery
}

// Stores a parsed version of the filter query string. Can be used by
// providers to apply the filter based on their own implementation. The filter
// is stored as a parse tree that can be traversed.
type GoDataFilterQuery struct {
	Tree *ParseNode
	// The raw filter string
	RawValue string
}

type GoDataApplyQuery string

type GoDataExpandQuery struct {
	ExpandItems []*ExpandItem
}

type GoDataSelectQuery struct {
	SelectItems []*SelectItem
	// The raw select string
	RawValue string
}

type GoDataOrderByQuery struct {
	OrderByItems []*OrderByItem
	// The raw orderby string
	RawValue string
}

type GoDataTopQuery int

type GoDataSkipQuery int

type GoDataCountQuery bool

type GoDataInlineCountQuery string

type GoDataSearchQuery struct {
	Tree *ParseNode
	// The raw search string
	RawValue string
}

type GoDataFormatQuery struct {
}

// Check if this identifier has more than one key/value pair.
func (id *GoDataIdentifier) HasMultiple() bool {
	count := 0
	for range map[string]string(*id) {
		count++
	}
	return count > 1
}

// Return the first key in the map. This is how you should get the identifier
// for single values, e.g. when the path is Employee(1), etc.
func (id *GoDataIdentifier) Get() string {
	for k, _ := range map[string]string(*id) {
		return k
	}
	return ""
}

// Return a specific value for a specific key.
func (id *GoDataIdentifier) GetKey(key string) (string, bool) {
	v, ok := map[string]string(*id)[key]
	return v, ok
}
