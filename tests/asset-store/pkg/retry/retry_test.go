package retry

import (
	"errors"
	"fmt"
	"testing"

	"k8s.io/apimachinery/pkg/runtime/schema"

	. "github.com/onsi/gomega"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

var (
	errTest1 = errors.New("errorTest1")
)

func Test_errorFuncWithIgnore(t *testing.T) {
	tests := []struct {
		fn       func() error
		ignoreFn func(error) bool
		expected error
	}{
		{
			fn: func() error {
				return errTest1
			},
			expected: errTest1,
		},
		{
			fn: func() error {
				return errTest1
			},
			ignoreFn: func(err error) bool {
				return true
			},
		},
		{
			expected: ErrInvalidErrorFunc,
		},
		{
			ignoreFn: func(err error) bool {
				return err == ErrInvalidErrorFunc
			},
			expected: ErrInvalidErrorFunc,
		},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("test%d", i), func(t *testing.T) {
			g := NewGomegaWithT(t)
			actualErr := errorFuncWithIgnore(test.fn, test.ignoreFn)()
			switch {
			case test.expected == nil:
				g.Expect(actualErr).To(BeNil())
			default:
				g.Expect(actualErr).To(Equal(test.expected))
			}
		})
	}
}

func Test_errorFuncWithIgnore_callback(t *testing.T) {
	g := NewGomegaWithT(t)
	actual := false
	errorFuncWithIgnore(func() error {
		return errTest1
	}, func(err error) bool {
		return true
	}, func(args ...interface{}) {
		actual = true
	})()
	g.Expect(actual).To(Equal(true))
}

func Test_shouldRetry(t *testing.T) {
	tests := []struct {
		err      error
		expected bool
	}{
		{
			err:      apierrors.NewTimeoutError("test timeout error", 10),
			expected: true,
		},
		{
			err: apierrors.NewServerTimeout(schema.GroupResource{
				Group:    "test",
				Resource: "me",
			}, "GET", 10),
			expected: true,
		},
		{
			err:      apierrors.NewTooManyRequests("test error", 10),
			expected: true,
		},
		{
			err:      ErrInvalidErrorFunc,
			expected: false,
		},
		{
			expected: false,
		},
	}
	for i, test := range tests {
		retry := shouldRetry()
		t.Run(fmt.Sprintf("test%d", i), func(t *testing.T) {
			g := NewGomegaWithT(t)
			actual := retry(test.err)
			g.Expect(actual).To(Equal(test.expected))
		})
	}
}

func Test_shouldRetr_callback(t *testing.T) {
	g := NewGomegaWithT(t)
	var actual string
	cbk := func(data ...interface{}) {
		actual = data[0].(string)
	}
	shouldRetry(cbk)(apierrors.NewTimeoutError("test timeout error", 10))
	g.Expect(actual).To(Equal("retrying due to: Timeout: test timeout error"))
}
