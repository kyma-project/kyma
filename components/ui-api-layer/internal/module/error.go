package module

import "fmt"

type DisabledModuleError struct {
	ModuleName string
}

func NewDisabledModuleError(moduleName string) *DisabledModuleError {
	return &DisabledModuleError{
		ModuleName: moduleName,
	}
}

func (e *DisabledModuleError) Error() string {
	errMessage := fmt.Sprintf("The %s module is disabled.", e.ModuleName)
	return errMessage
}
