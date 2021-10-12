package access_test

import (
	"github.com/kyma-project/kyma/components/application-broker/internal/access"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/application-broker/internal"
	"github.com/kyma-project/kyma/components/application-broker/internal/access/automock"
	mappingTypes "github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned/fake"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestMappingExistsProvisionCheckerWhenProvisionAcceptable(t *testing.T) {
	// GIVEN
	rm := fixProdMapping()
	mockClientSet := fake.NewSimpleClientset(rm)

	mockAppStorage := &automock.ApplicationFinder{}
	defer mockAppStorage.AssertExpectations(t)
	mockAppStorage.On("FindOneByServiceID", fixApplicationServiceID()).Return(fixApplicationModel(), nil)

	sut := access.NewMappingExistsProvisionChecker(mockAppStorage, mockClientSet.ApplicationconnectorV1alpha1())
	// WHEN
	canProvisionOutput, err := sut.CanProvision(fixApplicationServiceID(), internal.Namespace(fixProdNs()), time.Nanosecond)
	// THEN
	assert.NoError(t, err)
	assert.True(t, canProvisionOutput.Allowed)
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

func fixApplicationModel() *internal.Application {
	return &internal.Application{
		Name: internal.ApplicationName(fixApName()),
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

func fixProdMapping() *mappingTypes.ApplicationMapping {
	return &mappingTypes.ApplicationMapping{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fixApName(),
			Namespace: fixProdNs(),
		},
	}
}

func fixProdNs() string {
	return "production"
}

func fixApName() string {
	return "ec-prod"
}
