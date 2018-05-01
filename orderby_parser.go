package godata

import (
	"strings"
)

const (
	ASC  = "asc"
	DESC = "desc"
)

type OrderByItem struct {
	Field *Token
	Order string
}

func ParseOrderByString(orderby string) (*GoDataOrderByQuery, error) {
	items := strings.Split(orderby, ",")

	result := make([]*OrderByItem, 0)

	for _, v := range items {
		parts := strings.Split(v, " ")
		field := &Token{Value: parts[0]}
		var order string = ASC
		if len(parts) > 1 {
			if strings.ToLower(parts[1]) == ASC {
				order = ASC
			} else if strings.ToLower(parts[1]) == DESC {
				order = DESC
			} else {
				return nil, BadRequestError("Could not parse orderby query.")
			}
		}
		result = append(result, &OrderByItem{field, order})
	}

	return &GoDataOrderByQuery{result, orderby}, nil
}

func SemanticizeOrderByQuery(orderby *GoDataOrderByQuery, service *GoDataService, entity *GoDataEntityType) error {
	if orderby == nil {
		return nil
	}

	for _, item := range orderby.OrderByItems {
		if prop, ok := service.PropertyLookup[entity][item.Field.Value]; ok {
			item.Field.SemanticType = SemanticTypeProperty
			item.Field.SemanticReference = prop
		} else {
			return BadRequestError("No property " + item.Field.Value + " for entity " + entity.Name)
		}
	}

	return nil
}
