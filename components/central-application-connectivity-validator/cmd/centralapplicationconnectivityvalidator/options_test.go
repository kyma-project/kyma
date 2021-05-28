package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOptionsValidation(t *testing.T) {
	tests := []struct {
		name  string
		args  args
		valid bool
	}{
		{
			name:  "default arguments",
			valid: true,
			args: args{
				appNamePlaceholder:       "%%APP_NAME%%",
				eventServicePathPrefixV1: "/%%APP_NAME%%/v1/events",
				eventServicePathPrefixV2: "/%%APP_NAME%%/v2/events",
				eventMeshPathPrefix:      "/%%APP_NAME%%/events",
				appRegistryPathPrefix:    "/%%APP_NAME%%/v1/metadata",
			},
		},
		{
			name:  "skip validation when appNamePlaceholder is empty",
			valid: true,
			args: args{
				appNamePlaceholder:       "",
				eventServicePathPrefixV1: "/app1/v1/events",
				eventServicePathPrefixV2: "/app1/v2/events",
				eventMeshPathPrefix:      "//events",
				appRegistryPathPrefix:    "/app2/v1/metadata",
			},
		},
		{
			name:  "missing app name prefix in eventServicePathPrefixV1",
			valid: false,
			args: args{
				appNamePlaceholder:       "%%APP_NAME%%",
				eventServicePathPrefixV1: "/v1/events",
				eventServicePathPrefixV2: "/%%APP_NAME%%/v2/events",
				eventMeshPathPrefix:      "/%%APP_NAME%%/events",
				appRegistryPathPrefix:    "/%%APP_NAME%%/v1/metadata",
			},
		},
		{
			name:  "missing app name prefix in eventServicePathPrefixV2",
			valid: false,
			args: args{
				appNamePlaceholder:       "%%APP_NAME%%",
				eventServicePathPrefixV1: "/%%APP_NAME%%/v1/events",
				eventServicePathPrefixV2: "//v2/events",
				eventMeshPathPrefix:      "/%%APP_NAME%%/events",
				appRegistryPathPrefix:    "/%%APP_NAME%%/v1/metadata",
			},
		},
		{
			name:  "missing app name prefix in eventMeshPathPrefix",
			valid: false,
			args: args{
				appNamePlaceholder:       "%%APP_NAME%%",
				eventServicePathPrefixV1: "/%%APP_NAME%%/v1/events",
				eventServicePathPrefixV2: "/%%APP_NAME%%/v2/events",
				eventMeshPathPrefix:      "//events",
				appRegistryPathPrefix:    "/%%APP_NAME%%/v1/metadata",
			},
		},
		{
			name:  "missing app name prefix in appRegistryPathPrefix",
			valid: false,
			args: args{
				appNamePlaceholder:       "%%APP_NAME%%",
				eventServicePathPrefixV1: "/%%APP_NAME%%/v1/events",
				eventServicePathPrefixV2: "/%%APP_NAME%%/v2/events",
				eventMeshPathPrefix:      "/%%APP_NAME%%/events",
				appRegistryPathPrefix:    "//v1/metadata",
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			opts := options{
				args:   tc.args,
				config: config{},
			}
			err := opts.validate()
			b := (err == nil && tc.valid) || (err != nil && !tc.valid)
			assert.Truef(t, b, "Parsing validation error: %v, valid: %v", err, tc.valid)
		})
	}
}
