package apierror_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/apierror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testKind int

const (
	someTestKind testKind = iota
)

func (k testKind) String() string {
	return "Test Kind"
}

func TestNewInvalid(t *testing.T) {
	var testCases = []struct {
		caseName  string
		kind      fmt.Stringer
		aggregate apierror.ErrorFieldAggregate
	}{
		{"EmptyAggregate", someTestKind, apierror.ErrorFieldAggregate{}},
		{"NotEmptyAggregate", someTestKind, apierror.ErrorFieldAggregate{
			apierror.NewMissingField("apiVersion"),
			apierror.NewInvalidField("kind", "", ""),
		}},
		{"EmptyResource", nil, apierror.ErrorFieldAggregate{}},
	}

	for _, testCase := range testCases {
		t.Run(testCase.caseName, func(t *testing.T) {
			result := apierror.NewInvalid(testCase.kind, testCase.aggregate)

			require.NotNil(t, result)
			assert.True(t, apierror.IsInvalid(result))
			assert.NotEmpty(t, result.Error())
		})
	}
}

func TestNewInvalidField(t *testing.T) {
	var testCases = []struct {
		caseName string
		path     string
		value    string
		detail   string
	}{
		{"AllParamsProvided", "kind", "Pood", "Resource name shouldn't be changed"},
		{"NoPath", "", "Pood", "Resource name shouldn't be changed"},
		{"NoValue", "kind", "", "Resource name shouldn't be changed"},
		{"NoDetail", "kind", "Pood", ""},
		{"NoArgs", "", "", ""},
	}

	for _, testCase := range testCases {
		t.Run(testCase.caseName, func(t *testing.T) {
			result := apierror.NewInvalidField(testCase.path, testCase.value, testCase.detail)

			require.NotNil(t, result)
			assert.NotEmpty(t, result)
		})
	}
}

func TestNewMissingField(t *testing.T) {
	var testCases = []struct {
		caseName string
		path     string
	}{
		{"WithPath", "kind"},
		{"NoPath", ""},
	}

	for _, testCase := range testCases {
		t.Run(testCase.caseName, func(t *testing.T) {
			result := apierror.NewMissingField(testCase.path)

			require.NotNil(t, result)
			assert.NotEmpty(t, result)
		})
	}
}

func TestErrorFieldAggregate_String(t *testing.T) {
	var testCases = []struct {
		caseName    string
		aggregate   apierror.ErrorFieldAggregate
		expectEmpty bool
	}{
		{"Empty", apierror.ErrorFieldAggregate{}, true},
		{"Single", apierror.ErrorFieldAggregate{
			apierror.NewMissingField("kind"),
		}, false},
		{"Multiple", apierror.ErrorFieldAggregate{
			apierror.NewMissingField("kind"),
			apierror.NewMissingField("apiVersion"),
		}, false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.caseName, func(t *testing.T) {
			result := testCase.aggregate.String()

			assert.Equal(t, testCase.expectEmpty, len(result) == 0)
		})
	}
}

func TestIsInvalid(t *testing.T) {
	var testCases = []struct {
		caseName string
		given    error
		expected bool
	}{
		{"Internal", apierror.NewInvalid(someTestKind, nil), true},
		{"Generic", errors.New("generic"), false},
		{"Nil", nil, false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.caseName, func(t *testing.T) {
			result := apierror.IsInvalid(testCase.given)

			assert.Equal(t, testCase.expected, result)
		})
	}
}
