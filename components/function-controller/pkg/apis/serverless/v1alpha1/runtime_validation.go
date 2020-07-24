package v1alpha1

import (
	"errors"
	"fmt"
	"strings"
)

func ValidateDependencies(runtime Runtime, dependencies string) error {
	switch runtime {
	case Nodejs10, Nodejs12:
		return validateNodeJSDependecies(dependencies)
	case Python37:
		return nil
	}
	return fmt.Errorf("Cannot find runtime: %s", runtime)

}

func validateNodeJSDependecies(dependencies string) error {
	if deps := strings.TrimSpace(dependencies); deps != "" && (deps[0] != '{' || deps[len(deps)-1] != '}') {
		return errors.New("deps should start with '{' and end with '}'")
	}
	return nil
}
