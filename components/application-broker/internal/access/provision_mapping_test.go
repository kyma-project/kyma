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
	app := fixRemoteEnv()
	rm := fixProdMapping()
	mockClientSet := fake.NewSimpleClientset(app, rm)

	mockREStorage := &automock.ApplicationFinder{}
	defer mockREStorage.AssertExpectations(t)
	mockREStorage.On("FindOneByServiceID", fixApplicationServiceID()).Return(fixRemoteEnvModel(), nil)

	sut := access.NewMappingExistsProvisionChecker(mockREStorage, mockClientSet.ApplicationconnectorV1alpha1())
	// WHEN
	canProvisionOutput, err := sut.CanProvision(fixApplicationServiceID(), internal.Namespace(fixProdNs()), time.Nanosecond)
	// THEN
	assert.NoError(t, err)
	assert.True(t, canProvisionOutput.Allowed)
}

func TestMappingExistsProvisionCheckerWhenProvisionNotAcceptable(t *testing.T) {
	// GIVEN
	app := fixRemoteEnv()
	rm := fixProdMapping()
	mockClientSet := fake.NewSimpleClientset(app, rm)

	mockREStorage := &automock.ApplicationFinder{}
	defer mockREStorage.AssertExpectations(t)
	mockREStorage.On("FindOneByServiceID", fixApplicationServiceID()).Return(fixRemoteEnvModel(), nil)

	sut := access.NewMappingExistsProvisionChecker(mockREStorage, mockClientSet.ApplicationconnectorV1alpha1())
	// WHEN
	canProvisionOutput, err := sut.CanProvision(fixApplicationServiceID(), internal.Namespace("stage"), time.Nanosecond)
	// THEN
	assert.NoError(t, err)
	assert.False(t, canProvisionOutput.Allowed)
	assert.Equal(t, "ApplicationMapping does not exist in the [stage] namespace", canProvisionOutput.Reason)
}

func TestMappingExistsProvisionCheckerReturnsErrorWhenRENotFound(t *testing.T) {
	// GIVEN
	mockREStorage := &automock.ApplicationFinder{}
	defer mockREStorage.AssertExpectations(t)
	mockREStorage.On("FindOneByServiceID", fixApplicationServiceID()).Return(nil, nil)

	sut := access.NewMappingExistsProvisionChecker(mockREStorage, nil)
	// WHEN
	_, err := sut.CanProvision(fixApplicationServiceID(), internal.Namespace("ns"), time.Nanosecond)
	// THEN
	assert.Error(t, err)
}

func fixRemoteEnv() *v1alpha1.Application {
	return &v1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name: fixREName(),
		},
		Spec: v1alpha1.ApplicationSpec{
			Services: []v1alpha1.Service{
				{
					ID: "service-id",
				},
			},
		},
	}
}

func fixRemoteEnvModel() *internal.Application {
	return &internal.Application{
		Name: internal.ApplicationName(fixREName()),
		Services: []internal.Service{
			{
				ID: internal.ApplicationServiceID("service-id"),
			},
		},
	}
}

func fixApplicationServiceID() internal.ApplicationServiceID {
	return internal.ApplicationServiceID("service-id")
}

func fixProdMapping() *v1alpha1.ApplicationMapping {
	return &v1alpha1.ApplicationMapping{
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
