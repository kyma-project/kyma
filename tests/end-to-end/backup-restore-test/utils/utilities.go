package utils

import "fmt"

func need(needed int, expected []interface{}) string {
	if len(expected) != needed {
		return fmt.Sprintf("Function needs %v arguments but got %v", needed, len(expected))
	}
	return ""
}

func needBetween(needed int, maximum int, expected []interface{}) string {
	if len(expected) < needed {
		return fmt.Sprintf("Function needs %v arguments but got %v", needed, len(expected))
	}
	if len(expected) > maximum {
		return fmt.Sprintf("Function needs between %v and %v arguments but got %v", needed, maximum, len(expected))
	}
	return ""
}
