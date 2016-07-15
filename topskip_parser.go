package godata

import (
	"strconv"
)

func ParseTopString(top string) (*GoDataTopQuery, error) {
	i, err := strconv.Atoi(top)
	result := GoDataTopQuery(i)
	return &result, err
}

func ParseSkipString(skip string) (*GoDataSkipQuery, error) {
	i, err := strconv.Atoi(skip)
	result := GoDataSkipQuery(i)
	return &result, err
}
