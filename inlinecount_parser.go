package godata

const (
	ALLPAGES = "allpages"
	NONE     = "none"
)

func ParseInlineCountString(inlinecount string) (*GoDataInlineCountQuery, error) {
	result := GoDataInlineCountQuery(inlinecount)
	if inlinecount == ALLPAGES {
		return &result, nil
	} else if inlinecount == NONE {
		return &result, nil
	} else {
		return nil, BadRequestError("Invalid inlinecount query.")
	}
}
