package v1alpha2

import (
	"fmt"
	"strings"
)

func returnAllErrs(msg string, allErrs []string) error {
	if len(allErrs) == 0 {
		return nil
	}

	if len(msg) > 0 {
		return fmt.Errorf("%s: %v", msg, allErrs)
	}

	return fmt.Errorf("%v", allErrs)
}

func validateIfMissingFields(properties ...property) error {
	var allErrs []string
	for _, item := range properties {
		if strings.TrimSpace(item.value) != "" {
			continue
		}
		allErrs = append(allErrs, fmt.Sprintf("%s is required", item.name))
	}

	return returnAllErrs("missing required fields", allErrs)
}
