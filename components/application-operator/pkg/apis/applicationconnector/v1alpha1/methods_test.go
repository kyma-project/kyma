package v1alpha1_test

import (
	"testing"

	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestApplication_HasFinalizer(t *testing.T) {
	testCases := []struct {
		finalizers []string
		searched   string
		result     bool
	}{
		{
			finalizers: []string{"finalizer.test", "finalizer.test2", "finalizer.test3"},
			searched:   "finalizer.test",
			result:     true,
		},
		{
			finalizers: []string{"finalizer.test", "finalizer.test2", "finalizer.test3"},
			searched:   "finalizer.not.present",
			result:     false,
		},
		{
			finalizers: nil,
			searched:   "finalizer",
			result:     false,
		},
	}

	t.Run("test has finalizer", func(t *testing.T) {
		for _, test := range testCases {
			app := v1alpha1.Application{
				ObjectMeta: v1.ObjectMeta{
					Name:       "test",
					Finalizers: test.finalizers,
				},
			}

			assert.Equal(t, test.result, app.HasFinalizer(test.searched))
		}
	})
}

func TestApplication_RemoveFinalizer(t *testing.T) {
	testCases := []struct {
		finalizers []string
		removed    string
		result     []string
	}{
		{
			finalizers: []string{"finalizer.test", "finalizer.test2", "finalizer.test3"},
			removed:    "finalizer.test",
			result:     []string{"finalizer.test2", "finalizer.test3"},
		},
		{
			finalizers: []string{"finalizer.test", "finalizer.test2", "finalizer.test3"},
			removed:    "finalizer.not.present",
			result:     []string{"finalizer.test", "finalizer.test2", "finalizer.test3"},
		},
		{
			finalizers: nil,
			removed:    "finalizer",
			result:     nil,
		},
	}

	t.Run("test has finalizer", func(t *testing.T) {
		for _, test := range testCases {
			app := v1alpha1.Application{
				ObjectMeta: v1.ObjectMeta{
					Name:       "test",
					Finalizers: test.finalizers,
				},
			}

			app.RemoveFinalizer(test.removed)

			assert.Equal(t, test.result, app.Finalizers)
		}
	})
}

func TestApplication_SetFinalizer(t *testing.T) {
	testCases := []struct {
		finalizers []string
		new        string
		result     []string
	}{
		{
			finalizers: []string{"finalizer.test", "finalizer.test2", "finalizer.test3"},
			new:        "finalizer.test4",
			result:     []string{"finalizer.test", "finalizer.test2", "finalizer.test3", "finalizer.test4"},
		},
		{
			finalizers: []string{"finalizer.test", "finalizer.test2", "finalizer.test3"},
			new:        "finalizer.test",
			result:     []string{"finalizer.test", "finalizer.test2", "finalizer.test3"},
		},
		{
			finalizers: nil,
			new:        "finalizer",
			result:     []string{"finalizer"},
		},
	}

	t.Run("test has finalizer", func(t *testing.T) {
		for _, test := range testCases {
			app := v1alpha1.Application{
				ObjectMeta: v1.ObjectMeta{
					Name:       "test",
					Finalizers: test.finalizers,
				},
			}

			app.SetFinalizer(test.new)

			assert.Equal(t, test.result, app.Finalizers)
		}
	})
}
