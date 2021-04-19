package ptr

import "strconv"

func String(val string) *string {
	return &val
}

func BoolFromString(val string) *bool {
	b, err := strconv.ParseBool(val)
	if err != nil {
		return nil
	}

	return &b
}
