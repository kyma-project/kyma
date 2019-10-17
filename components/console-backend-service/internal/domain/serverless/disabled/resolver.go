// Code generated by failery v1.0.0. DO NOT EDIT.

package disabled

import context "context"
import gqlschema "github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"

// Resolver is an autogenerated failing mock type for the Resolver type
type Resolver struct {
	err error
}

// NewResolver creates a new Resolver type instance
func NewResolver(err error) *Resolver {
	return &Resolver{err: err}
}

// CreateFunction provides a failing mock function with given fields: ctx, name, namespace, labels, size, runtime
func (_m *Resolver) CreateFunction(ctx context.Context, name string, namespace string, labels gqlschema.Labels, size string, runtime string) (gqlschema.Function, error) {
	var r0 gqlschema.Function
	var r1 error
	r1 = _m.err

	return r0, r1
}

// DeleteFunction provides a failing mock function with given fields: ctx, name, namespace
func (_m *Resolver) DeleteFunction(ctx context.Context, name string, namespace string) (gqlschema.FunctionMutationOutput, error) {
	var r0 gqlschema.FunctionMutationOutput
	var r1 error
	r1 = _m.err

	return r0, r1
}

// FunctionQuery provides a failing mock function with given fields: ctx, name, namespace
func (_m *Resolver) FunctionQuery(ctx context.Context, name string, namespace string) (*gqlschema.Function, error) {
	var r0 *gqlschema.Function
	var r1 error
	r1 = _m.err

	return r0, r1
}

// FunctionsQuery provides a failing mock function with given fields: ctx, namespace
func (_m *Resolver) FunctionsQuery(ctx context.Context, namespace string) ([]gqlschema.Function, error) {
	var r0 []gqlschema.Function
	var r1 error
	r1 = _m.err

	return r0, r1
}

// UpdateFunction provides a failing mock function with given fields: ctx, name, namespace, params
func (_m *Resolver) UpdateFunction(ctx context.Context, name string, namespace string, params gqlschema.FunctionUpdateInput) (gqlschema.Function, error) {
	var r0 gqlschema.Function
	var r1 error
	r1 = _m.err

	return r0, r1
}
