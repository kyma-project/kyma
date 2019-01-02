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
	errMessage := fmt.Sprintf("MODULE_DISABLED: The %s module is disabled.", e.ModuleName)
	return errMessage
}

func IsDisabledModuleError(err error) bool {
	_, ok := err.(*DisabledModuleError)
	return ok
}
