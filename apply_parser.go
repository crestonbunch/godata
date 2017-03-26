package godata

func ParseApplyString(apply string) (*GoDataApplyQuery, error) {
	result := GoDataApplyQuery(apply)
	return &result, nil
}
