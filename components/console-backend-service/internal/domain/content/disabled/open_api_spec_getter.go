// Code generated by failery v1.0.0. DO NOT EDIT.

package disabled

import storage "github.com/kyma-project/kyma/components/console-backend-service/internal/domain/content/storage"

// openApiSpecGetter is an autogenerated failing mock type for the openApiSpecGetter type
type openApiSpecGetter struct {
	err error
}

// NewOpenApiSpecGetter creates a new openApiSpecGetter type instance
func NewOpenApiSpecGetter(err error) *openApiSpecGetter {
	return &openApiSpecGetter{err: err}
}

// Find provides a failing mock function with given fields: kind, id
func (_m *openApiSpecGetter) Find(kind string, id string) (*storage.OpenApiSpec, error) {
	var r0 *storage.OpenApiSpec
	var r1 error
	r1 = _m.err

	return r0, r1
}
