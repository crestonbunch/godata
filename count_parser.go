package godata

import (
	"strconv"
)

func ParseCountString(count string) (*GoDataCountQuery, error) {
	i, err := strconv.ParseBool(count)
	if err != nil {
		return nil, err
	}

	result := GoDataCountQuery(i)

	return &result, nil
}
