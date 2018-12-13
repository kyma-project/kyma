package access_test

import (
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/application-broker/internal"
	"github.com/kyma-project/kyma/components/application-broker/internal/access"
	"github.com/kyma-project/kyma/components/application-broker/internal/access/automock"
	"github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned/fake"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestMappingExistsProvisionCheckerWhenProvisionAcceptable(t *testing.T) {
	// GIVEN
	re := fixRemoteEnv()
	rm := fixProdMapping()
	mockClientSet := fake.NewSimpleClientset(re, rm)

	mockREStorage := &automock.RemoteEnvironmentFinder{}
	defer mockREStorage.AssertExpectations(t)
	mockREStorage.On("FindOneByServiceID", fixRemoteServiceID()).Return(fixRemoteEnvModel(), nil)

	sut := access.NewMappingExistsProvisionChecker(mockREStorage, mockClientSet.ApplicationconnectorV1alpha1())
	// WHEN
	canProvisionOutput, err := sut.CanProvision(fixRemoteServiceID(), internal.Namespace(fixProdNs()), time.Nanosecond)
	// THEN
	assert.NoError(t, err)
	assert.True(t, canProvisionOutput.Allowed)
}

func TestMappingExistsProvisionCheckerWhenProvisionNotAcceptable(t *testing.T) {
	// GIVEN
	re := fixRemoteEnv()
	rm := fixProdMapping()
	mockClientSet := fake.NewSimpleClientset(re, rm)

	mockREStorage := &automock.RemoteEnvironmentFinder{}
	defer mockREStorage.AssertExpectations(t)
	mockREStorage.On("FindOneByServiceID", fixRemoteServiceID()).Return(fixRemoteEnvModel(), nil)

	sut := access.NewMappingExistsProvisionChecker(mockREStorage, mockClientSet.ApplicationconnectorV1alpha1())
	// WHEN
	canProvisionOutput, err := sut.CanProvision(fixRemoteServiceID(), internal.Namespace("stage"), time.Nanosecond)
	// THEN
	assert.NoError(t, err)
	assert.False(t, canProvisionOutput.Allowed)
	assert.Equal(t, "EnvironmentMapping does not exist in the [stage] namespace", canProvisionOutput.Reason)
}

func TestMappingExistsProvisionCheckerReturnsErrorWhenRENotFound(t *testing.T) {
	// GIVEN
	mockREStorage := &automock.RemoteEnvironmentFinder{}
	defer mockREStorage.AssertExpectations(t)
	mockREStorage.On("FindOneByServiceID", fixRemoteServiceID()).Return(nil, nil)

	sut := access.NewMappingExistsProvisionChecker(mockREStorage, nil)
	// WHEN
	_, err := sut.CanProvision(fixRemoteServiceID(), internal.Namespace("ns"), time.Nanosecond)
	// THEN
	assert.Error(t, err)
}

func fixRemoteEnv() *v1alpha1.RemoteEnvironment {
	return &v1alpha1.RemoteEnvironment{
		ObjectMeta: metav1.ObjectMeta{
			Name: fixREName(),
		},
		Spec: v1alpha1.RemoteEnvironmentSpec{
			Services: []v1alpha1.Service{
				{
					ID: "service-id",
				},
			},
		},
	}
}

func fixRemoteEnvModel() *internal.RemoteEnvironment {
	return &internal.RemoteEnvironment{
		Name: internal.RemoteEnvironmentName(fixREName()),
		Services: []internal.Service{
			{
				ID: internal.RemoteServiceID("service-id"),
			},
		},
	}
}

func fixRemoteServiceID() internal.RemoteServiceID {
	return internal.RemoteServiceID("service-id")
}

func fixProdMapping() *v1alpha1.EnvironmentMapping {
	return &v1alpha1.EnvironmentMapping{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fixREName(),
			Namespace: fixProdNs(),
		},
	}
}

func fixProdNs() string {
	return "production"
}

func fixREName() string {
	return "ec-prod"
}
