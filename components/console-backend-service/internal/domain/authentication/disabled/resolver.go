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

// CreateIDPPresetMutation provides a failing mock function with given fields: ctx, name, issuer, jwksURI
func (_m *Resolver) CreateIDPPresetMutation(ctx context.Context, name string, issuer string, jwksURI string) (*gqlschema.IDPPreset, error) {
	var r0 *gqlschema.IDPPreset
	var r1 error
	r1 = _m.err

	return r0, r1
}

// DeleteIDPPresetMutation provides a failing mock function with given fields: ctx, name
func (_m *Resolver) DeleteIDPPresetMutation(ctx context.Context, name string) (*gqlschema.IDPPreset, error) {
	var r0 *gqlschema.IDPPreset
	var r1 error
	r1 = _m.err

	return r0, r1
}

// IDPPresetQuery provides a failing mock function with given fields: ctx, name
func (_m *Resolver) IDPPresetQuery(ctx context.Context, name string) (*gqlschema.IDPPreset, error) {
	var r0 *gqlschema.IDPPreset
	var r1 error
	r1 = _m.err

	return r0, r1
}

// IDPPresetsQuery provides a failing mock function with given fields: ctx, first, offset
func (_m *Resolver) IDPPresetsQuery(ctx context.Context, first *int, offset *int) ([]gqlschema.IDPPreset, error) {
	var r0 []gqlschema.IDPPreset
	var r1 error
	r1 = _m.err

	return r0, r1
}
