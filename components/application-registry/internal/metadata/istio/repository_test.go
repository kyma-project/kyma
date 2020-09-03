package istio

import (
	"errors"
	"testing"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/kyma-project/kyma/components/application-registry/internal/k8sconsts"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/istio/mocks"
	"github.com/kyma-project/kyma/components/application-registry/pkg/apis/istio/v1alpha2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const applicationUID = types.UID("appUID")

var config = RepositoryConfig{Namespace: "testns"}

func TestRepository_Create(t *testing.T) {

	t.Run("should create authorization policy", func(t *testing.T) {
		// given
		expected := &v1alpha2.AuthorizationPolicy{
			ObjectMeta: v1.ObjectMeta{
				Name: "app-test-uuid1",
				Labels: map[string]string{
					k8sconsts.LabelApplication: "app",
					k8sconsts.LabelServiceId:   "sid",
				},
				OwnerReferences: k8sconsts.CreateOwnerReferenceForApplication("app", applicationUID),
			},
			Spec: &v1alpha2.AuthorizationPolicySpec{
				Selector: &v1alpha2.WorkloadSelector{
					MatchLabels: map[string]string{
						"app-test-uuid1": "true",
					},
				},
				Action: v1alpha2.Allow,
				Rules: []v1alpha2.Rule{
					{
						To: []v1alpha2.To{
							{
								Operation: v1alpha2.Operation{
									Hosts: []string{
										"app-test-uuid1.testns.svc.cluster.local",
									},
								},
							},
						},
					},
				},
			},
		}

		authorizationPolicyInterface := new(mocks.AuthorizationPolicyInterface)
		authorizationPolicyInterface.On("Create", expected).Return(nil, nil)

		repository := NewRepository(authorizationPolicyInterface, config)

		// when
		err := repository.CreateAuthorizationPolicy("app", "appUID", "sid", "app-test-uuid1")

		// then
		assert.NoError(t, err)
		authorizationPolicyInterface.AssertExpectations(t)
	})

	t.Run("should handle error when creating authorization policy", func(t *testing.T) {
		// given
		authorizationPolicyInterface := new(mocks.AuthorizationPolicyInterface)
		authorizationPolicyInterface.On("Create", mock.AnythingOfType("*v1alpha2.AuthorizationPolicy")).
			Return(nil, errors.New("some error"))

		repository := NewRepository(authorizationPolicyInterface, config)

		// when
		err := repository.CreateAuthorizationPolicy("app", "appUID", "sid", "app-test-uuid1")

		// then
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "some error")
	})
}

func TestRepository_Upsert(t *testing.T) {

	t.Run("should upsert authorization policy", func(t *testing.T) {
		// given
		expected := &v1alpha2.AuthorizationPolicy{
			ObjectMeta: v1.ObjectMeta{
				Name: "app-test-uuid1",
				Labels: map[string]string{
					k8sconsts.LabelApplication: "app",
					k8sconsts.LabelServiceId:   "sid",
				},
				OwnerReferences: k8sconsts.CreateOwnerReferenceForApplication("app", applicationUID),
			},
			Spec: &v1alpha2.AuthorizationPolicySpec{
				Selector: &v1alpha2.WorkloadSelector{
					MatchLabels: map[string]string{
						"app-test-uuid1": "true",
					},
				},
				Action: v1alpha2.Allow,
				Rules: []v1alpha2.Rule{
					{
						To: []v1alpha2.To{
							{
								Operation: v1alpha2.Operation{
									Hosts: []string{
										"app-test-uuid1.testns.svc.cluster.local",
									},
								},
							},
						},
					},
				},
			},
		}

		authorizationPolicyInterface := new(mocks.AuthorizationPolicyInterface)
		authorizationPolicyInterface.On("Create", expected).Return(nil, nil)

		repository := NewRepository(authorizationPolicyInterface, config)

		// when
		err := repository.UpsertAuthorizationPolicy("app", "appUID", "sid", "app-test-uuid1")

		// then
		assert.NoError(t, err)
		authorizationPolicyInterface.AssertExpectations(t)
	})

	t.Run("should handle already exists error when upserting authorization policy", func(t *testing.T) {
		// given
		authorizationPolicyInterface := new(mocks.AuthorizationPolicyInterface)
		authorizationPolicyInterface.On("Create", mock.AnythingOfType("*v1alpha2.AuthorizationPolicy")).
			Return(nil, k8serrors.NewAlreadyExists(schema.GroupResource{}, ""))

		repository := NewRepository(authorizationPolicyInterface, config)

		// when
		err := repository.UpsertAuthorizationPolicy("app", "appUID", "sid", "app-test-uuid1")

		// then
		assert.NoError(t, err)
		authorizationPolicyInterface.AssertExpectations(t)
	})

	t.Run("should handle error when upserting authorization policy", func(t *testing.T) {
		// given
		authorizationPolicyInterface := new(mocks.AuthorizationPolicyInterface)
		authorizationPolicyInterface.On("Create", mock.AnythingOfType("*v1alpha2.AuthorizationPolicy")).
			Return(nil, errors.New("some error"))

		repository := NewRepository(authorizationPolicyInterface, config)

		// when
		err := repository.UpsertAuthorizationPolicy("app", "appUID", "sid", "app-test-uuid1")

		// then
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "some error")
	})
}

func TestRepository_Delete(t *testing.T) {

	t.Run("should delete authorization policy", func(t *testing.T) {
		// given
		authorizationPolicyInterface := new(mocks.AuthorizationPolicyInterface)
		authorizationPolicyInterface.On("Delete", "app-test-uuid1", (*v1.DeleteOptions)(nil)).Return(nil)

		repository := NewRepository(authorizationPolicyInterface, config)

		// when
		err := repository.DeleteAuthorizationPolicy("app-test-uuid1")

		// then
		assert.NoError(t, err)
		authorizationPolicyInterface.AssertExpectations(t)
	})

	t.Run("should handle error when deleting authorization policy", func(t *testing.T) {
		// given
		authorizationPolicyInterface := new(mocks.AuthorizationPolicyInterface)
		authorizationPolicyInterface.On("Delete", "app-test-uuid1", (*v1.DeleteOptions)(nil)).
			Return(errors.New("some error"))

		repository := NewRepository(authorizationPolicyInterface, config)

		// when
		err := repository.DeleteAuthorizationPolicy("app-test-uuid1")

		// then
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "some error")
	})

	t.Run("should ignore not found error when deleting authorization policy", func(t *testing.T) {
		// given
		authorizationPolicyInterface := new(mocks.AuthorizationPolicyInterface)
		authorizationPolicyInterface.On("Delete", "app-test-uuid1", (*v1.DeleteOptions)(nil)).
			Return(k8serrors.NewNotFound(schema.GroupResource{}, ""))

		repository := NewRepository(authorizationPolicyInterface, config)

		// when
		err := repository.DeleteAuthorizationPolicy("app-test-uuid1")

		// then
		assert.NoError(t, err)
	})
}
