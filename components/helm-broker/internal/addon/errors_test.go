package addon_test

import (
	"errors"
	"testing"

	"github.com/kyma-project/kyma/components/helm-broker/internal/addon"
	"github.com/stretchr/testify/assert"
)

func TestLoadingError(t *testing.T) {
	tests := map[string]struct {
		givenErr          error
		expToBeLoadingErr bool
	}{
		"Should report true for Loading error": {
			givenErr:          addon.NewLoadingError(errors.New("fix err")),
			expToBeLoadingErr: true,
		},
		"Should report false for generic error": {
			givenErr:          errors.New("fix err"),
			expToBeLoadingErr: false,
		},
		"Should report false for Fetching error": {
			givenErr:          addon.NewFetchingError(errors.New("fix err")),
			expToBeLoadingErr: false,
		},
	}
	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			assert.Equal(t, tc.expToBeLoadingErr, addon.IsLoadingError(tc.givenErr))
		})
	}
}

func TestFetchingError(t *testing.T) {
	tests := map[string]struct {
		givenErr          error
		expToBeLoadingErr bool
	}{
		"Should report true for Fetching error": {
			givenErr:          addon.NewFetchingError(errors.New("fix err")),
			expToBeLoadingErr: true,
		},
		"Should report false for generic error": {
			givenErr:          errors.New("fix err"),
			expToBeLoadingErr: false,
		},
		"Should report false for Loading error": {
			givenErr:          addon.NewLoadingError(errors.New("fix err")),
			expToBeLoadingErr: false,
		},
	}
	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			assert.Equal(t, tc.expToBeLoadingErr, addon.IsFetchingError(tc.givenErr))
		})
	}
}
