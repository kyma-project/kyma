package gqlerror_test

import (
	"fmt"
	"testing"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/apierror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

type testKind int

const (
	someTestKind testKind = iota
)

func (k testKind) String() string {
	return "Test Kind"
}

func TestNew(t *testing.T) {
	var testCases = map[string]struct {
		kind      fmt.Stringer
		err       error
		validator func(error) bool
	}{
		"K8sNotFound":      {someTestKind, k8serrors.NewNotFound(schema.GroupResource{}, "test"), gqlerror.IsNotFound},
		"K8sAlreadyExists": {someTestKind, k8serrors.NewAlreadyExists(schema.GroupResource{}, "test"), gqlerror.IsAlreadyExists},
		"K8sInvalid":       {someTestKind, k8serrors.NewInvalid(schema.GroupKind{}, "test", field.ErrorList{}), gqlerror.IsInvalid},
		"K8sOther":         {someTestKind, k8serrors.NewBadRequest("test"), gqlerror.IsInternal},
		"APIInvalid":       {someTestKind, apierror.NewInvalid(pretty.Pod, apierror.ErrorFieldAggregate{}), gqlerror.IsInvalid},
		"Nested":           {someTestKind, errors.Wrap(k8serrors.NewNotFound(schema.GroupResource{}, "while test"), "test"), gqlerror.IsNotFound},
		"DoubleNested":     {someTestKind, errors.Wrap(errors.Wrap(k8serrors.NewNotFound(schema.GroupResource{}, "while test"), "while more"), "test"), gqlerror.IsNotFound},
		"Generic":          {someTestKind, errors.New("test"), gqlerror.IsInternal},
		"Nil":              {someTestKind, nil, nil},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			// when
			result := gqlerror.New(testCase.err, testCase.kind)

			// then
			if testCase.err != nil {
				require.NotNil(t, result)
				assert.True(t, testCase.validator(result))
				assert.NotEmpty(t, result.Error())
			} else {
				require.Nil(t, result)
			}
		})
	}
}

func TestNewAlreadyExists(t *testing.T) {
	var testCases = []struct {
		caseName string
		kind     fmt.Stringer
		opts     []gqlerror.Option
	}{
		{"AllParamsProvided", someTestKind, []gqlerror.Option{gqlerror.WithNamespace("production"), gqlerror.WithName("name"), gqlerror.WithDetails("details")}},
		{"NoKindNoOpts", nil, nil},
		{"NoKind", nil, []gqlerror.Option{gqlerror.WithNamespace("namespace"), gqlerror.WithName("name")}},
		{"NoOpts", someTestKind, nil},
	}

	for _, testCase := range testCases {
		t.Run(testCase.caseName, func(t *testing.T) {
			// when
			result := gqlerror.NewAlreadyExists(testCase.kind, testCase.opts...)

			// then
			require.NotNil(t, result)
			assert.True(t, gqlerror.IsAlreadyExists(result))
			assert.NotEmpty(t, result.Error())
		})
	}
}

func TestNewNotFound(t *testing.T) {
	var testCases = []struct {
		caseName string
		kind     fmt.Stringer
		opts     []gqlerror.Option
	}{
		{"AllParamsProvided", someTestKind, []gqlerror.Option{gqlerror.WithNamespace("namespace"), gqlerror.WithName("name"), gqlerror.WithDetails("some details")}},
		{"NoKindNoOpts", nil, nil},
		{"NoKind", nil, []gqlerror.Option{gqlerror.WithNamespace("namespace"), gqlerror.WithName("name")}},
		{"NoOpts", someTestKind, nil},
	}

	for _, testCase := range testCases {
		t.Run(testCase.caseName, func(t *testing.T) {
			// when
			result := gqlerror.NewNotFound(testCase.kind, testCase.opts...)

			// then
			require.NotNil(t, result)
			assert.True(t, gqlerror.IsNotFound(result))
			assert.NotEmpty(t, result.Error())
		})
	}
}

func TestNewInternal(t *testing.T) {
	// when
	result := gqlerror.NewInternal()

	// then
	require.NotNil(t, result)
	assert.True(t, gqlerror.IsInternal(result))
	assert.NotEmpty(t, result.Error())
}

func TestNewInvalid(t *testing.T) {
	fixErr := "fix"

	var testCases = []struct {
		caseName string
		err      string
		kind     fmt.Stringer
		opts     []gqlerror.Option
	}{
		{"AllParamsProvided", fixErr, someTestKind, []gqlerror.Option{gqlerror.WithNamespace("namespace"), gqlerror.WithName("name"), gqlerror.WithDetails("some details")}},
		{"NoKindNoOpts", fixErr, nil, nil},
		{"NoKind", fixErr, nil, []gqlerror.Option{gqlerror.WithNamespace("namespace"), gqlerror.WithName("name")}},
		{"NoOpts", fixErr, someTestKind, nil},
	}

	for _, testCase := range testCases {
		t.Run(testCase.caseName, func(t *testing.T) {
			// when
			result := gqlerror.NewInvalid(testCase.err, testCase.kind, testCase.opts...)

			// then
			require.NotNil(t, result)
			assert.True(t, gqlerror.IsInvalid(result))
			assert.NotEmpty(t, result.Error())
		})
	}
}

func TestIsAlreadyExists(t *testing.T) {
	var testCases = []struct {
		caseName string
		given    error
		expected bool
	}{
		{"AlreadyExists", gqlerror.NewAlreadyExists(nil), true},
		{"NotFound", gqlerror.NewNotFound(nil), false},
		{"Generic", errors.New("generic"), false},
		{"Nil", nil, false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.caseName, func(t *testing.T) {
			// when
			result := gqlerror.IsAlreadyExists(testCase.given)

			// then
			assert.Equal(t, testCase.expected, result)
		})
	}
}

func TestIsNotFound(t *testing.T) {
	var testCases = []struct {
		caseName string
		given    error
		expected bool
	}{
		{"AlreadyExists", gqlerror.NewAlreadyExists(nil), false},
		{"NotFound", gqlerror.NewNotFound(nil), true},
		{"Generic", errors.New("generic"), false},
		{"Nil", nil, false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.caseName, func(t *testing.T) {
			// when
			result := gqlerror.IsNotFound(testCase.given)

			// then
			assert.Equal(t, testCase.expected, result)
		})
	}
}

func TestIsInternalServer(t *testing.T) {
	var testCases = []struct {
		caseName string
		given    error
		expected bool
	}{
		{"Internal", gqlerror.NewInternal(), true},
		{"NotFound", gqlerror.NewNotFound(nil), false},
		{"Generic", errors.New("generic"), false},
		{"Nil", nil, false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.caseName, func(t *testing.T) {
			// when
			result := gqlerror.IsInternal(testCase.given)

			// then
			assert.Equal(t, testCase.expected, result)
		})
	}
}

func TestIsInvalid(t *testing.T) {
	var testCases = []struct {
		caseName string
		given    error
		expected bool
	}{
		{"Internal", gqlerror.NewInternal(), false},
		{"Invalid", gqlerror.NewInvalid("fix", nil, gqlerror.WithNamespace("namespace")), true},
		{"NotFound", gqlerror.NewNotFound(nil), false},
		{"Generic", errors.New("generic"), false},
		{"Nil", nil, false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.caseName, func(t *testing.T) {
			// when
			result := gqlerror.IsInvalid(testCase.given)

			// then
			assert.Equal(t, testCase.expected, result)
		})
	}
}

func TestReason_String_Unknown(t *testing.T) {
	var testCases = []struct {
		caseName string
		given    gqlerror.Status
		expected string
	}{
		{"Unknown", gqlerror.Unknown, "unknown"},
		{"NotDefined", -12, "unknown"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.caseName, func(t *testing.T) {
			// then
			assert.Equal(t, testCase.expected, fmt.Sprintf("%s", testCase.given))
		})
	}
}
