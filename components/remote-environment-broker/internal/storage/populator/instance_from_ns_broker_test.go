package populator_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/fake"
	"github.com/kyma-project/kyma/components/remote-environment-broker/internal"
	"github.com/kyma-project/kyma/components/remote-environment-broker/internal/storage/populator"
	"github.com/kyma-project/kyma/components/remote-environment-broker/internal/storage/populator/automock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNsBrokerPopulateInstances(t *testing.T) {
	// GIVEN
	mockClientSet := fake.NewSimpleClientset(fixAllSCObjects()...)

	mockInserter := &automock.InstanceInserter{}
	defer mockInserter.AssertExpectations(t)

	insertExpectations := newSiInsertExpectations(expectedServiceInstanceFromNsBroker())
	defer insertExpectations.AssertExpectations(t)

	mockInserter.On("Insert", mock.MatchedBy(insertExpectations.OnInsertArgsChecker)).
		Run(func(args mock.Arguments) {
			actualInstance := args.Get(0).(*internal.Instance)
			fmt.Println(actualInstance)
			insertExpectations.ReportInsertingInstance(actualInstance)
		}).
		Return(nil).Once()

	mockConverter := &automock.InstanceConverter{}
	defer mockConverter.AssertExpectations(t)

	mockConverter.On("MapServiceInstance", fixRebServiceInstanceFromNsSC()).Return(expectedServiceInstanceFromNsBroker())
	sut := populator.NewInstancesFromNsBrokers(mockClientSet, mockInserter, mockConverter)
	// WHEN
	actualErr := sut.Do(context.Background())
	// THEN
	assert.NoError(t, actualErr)
}

func TestNsBrokerPopulateErrorOnInsert(t *testing.T) {
	// GIVEN
	mockClientSet := fake.NewSimpleClientset(fixAllSCObjects()...)

	mockInserter := &automock.InstanceInserter{}
	defer mockInserter.AssertExpectations(t)
	mockInserter.On("Insert", mock.Anything).Return(errors.New("some error"))

	mockConverter := &automock.InstanceConverter{}
	defer mockConverter.AssertExpectations(t)

	mockConverter.On("MapServiceInstance", mock.Anything).Return(&internal.Instance{})
	sut := populator.NewInstancesFromNsBrokers(mockClientSet, mockInserter, mockConverter)
	// WHEN
	actualErr := sut.Do(context.Background())
	// THEN
	assert.EqualError(t, actualErr, "while inserting service instance: some error")

}
