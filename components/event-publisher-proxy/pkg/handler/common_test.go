//go:build unit
// +build unit

package handler_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	handler "github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/handler"
)

func TestIsARequestWithLegacyEvent(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name         string
		givenURI     string
		wantedResult bool
	}{
		{
			name:         "is a legacy endpoint if application name, api version and events endpoint exist",
			givenURI:     "/app/v1/events",
			wantedResult: true,
		},
		{
			name:         "is not a legacy endpoint if application name, api version and events endpoint exist with multiple trailing slashes",
			givenURI:     "/app/v1/events//",
			wantedResult: false,
		},
		{
			name:         "is not a legacy endpoint if application name, api version and events endpoint exist after ",
			givenURI:     "/test/app/v1/events",
			wantedResult: false,
		},
		{
			name:         "is not a legacy endpoint if application name is missing",
			givenURI:     "/v1/events",
			wantedResult: false,
		},
		{
			name:         "is not a legacy endpoint if api version is missing",
			givenURI:     "/app/events",
			wantedResult: false,
		},
		{
			name:         "is not a legacy endpoint if events endpoint is missing",
			givenURI:     "/app/v1",
			wantedResult: false,
		},
		{
			name:         "is not a legacy endpoint if it ends with an endpoint other than events",
			givenURI:     "/app/v1/events/foo",
			wantedResult: false,
		},
		{
			name:         "is not a legacy endpoint if it contains multiple slashes in-between the api version and events endpoint",
			givenURI:     "/app/v1//events",
			wantedResult: false,
		},
		{
			name:         "is not a legacy endpoint if it contains any endpoint in-between the api version and events endpoint",
			givenURI:     "/app/v1/foo/events",
			wantedResult: false,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.wantedResult, handler.IsARequestWithLegacyEvent(tc.givenURI))
		})
	}
}
