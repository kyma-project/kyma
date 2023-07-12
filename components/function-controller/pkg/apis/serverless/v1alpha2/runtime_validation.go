package v1alpha2

import (
	"errors"
	"fmt"
	"strings"
)

func ValidateDependencies(runtime Runtime, dependencies string) error {
	switch runtime {
	case NodeJs16, NodeJs18:
		return validateNodeJSDependencies(dependencies)
	case Python39:
		return nil
	}
	return fmt.Errorf("cannot find runtime: %s", runtime)
}

func validateNodeJSDependencies(dependencies string) error {
	if deps := strings.TrimSpace(dependencies); deps != "" && (deps[0] != '{' || deps[len(deps)-1] != '}') {
		return errors.New("deps should start with '{' and end with '}'")
	}
	return nil
}
