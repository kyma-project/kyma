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

func Test_fnWithIgnore(t *testing.T) {
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
			expected: ErrInvalidFunc,
		},
		{
			ignoreFn: func(err error) bool {
				return err == ErrInvalidFunc
			},
			expected: ErrInvalidFunc,
		},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("test%d", i), func(t *testing.T) {
			g := NewGomegaWithT(t)
			actualErr := fnWithIgnore(test.fn, test.ignoreFn)()
			switch {
			case test.expected == nil:
				g.Expect(actualErr).To(BeNil())
			default:
				g.Expect(actualErr).To(Equal(test.expected))
			}
		})
	}
}

func Test_fnWithIgnore_callback(t *testing.T) {
	g := NewGomegaWithT(t)
	actual := false
	err := fnWithIgnore(func() error {
		return errTest1
	}, func(err error) bool {
		return true
	}, func(args ...interface{}) {
		actual = true
	})()
	g.Expect(err).To(BeNil())
	g.Expect(actual).To(Equal(true))
}

func Test_errorFn(t *testing.T) {
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
			err:      ErrInvalidFunc,
			expected: false,
		},
		{
			expected: false,
		},
	}
	for i, test := range tests {
		retry := errorFn()
		t.Run(fmt.Sprintf("test%d", i), func(t *testing.T) {
			g := NewGomegaWithT(t)
			actual := retry(test.err)
			g.Expect(actual).To(Equal(test.expected))
		})
	}
}

func Test_errorFn_callback(t *testing.T) {
	g := NewGomegaWithT(t)
	var actual string
	var ok bool
	cbk := func(data ...interface{}) {
		actual, ok = data[0].(string)
	}
	errorFn(cbk)(apierrors.NewTimeoutError("test timeout error", 10))
	g.Expect(ok).To(BeTrue())
	g.Expect(actual).To(Equal("retrying due to: Timeout: test timeout error"))
}
