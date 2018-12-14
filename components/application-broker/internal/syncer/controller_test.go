package syncer_test

import (
	"context"
	"io/ioutil"
	"sync"
	"testing"
	"time"

	"github.com/ghodss/yaml"
	"github.com/kyma-project/kyma/components/application-broker/internal"
	"github.com/kyma-project/kyma/components/application-broker/internal/syncer"
	"github.com/kyma-project/kyma/components/application-broker/internal/syncer/automock"
	"github.com/kyma-project/kyma/components/application-broker/platform/logger/spy"
	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned/fake"
	"github.com/kyma-project/kyma/components/application-operator/pkg/client/informers/externalversions"
	"github.com/stretchr/testify/mock"
)

func TestControllerRunSuccess(t *testing.T) {
	// given
	reCR := mustLoadCRFix("testdata/app-CR-valid.input.yaml")
	reDM := internal.Application{
		Name: "mapped",
	}

	client := fake.NewSimpleClientset(&reCR)

	informerFactory := externalversions.NewSharedInformerFactory(client, 0)
	serviceCatalogSharedInformers := informerFactory.Applicationconnector().V1alpha1()
	reInformer := serviceCatalogSharedInformers.Applications()

	expectations := &sync.WaitGroup{}
	expectations.Add(4)
	fulfillExpectation := func(mock.Arguments) {
		expectations.Done()
	}

	validatorMock := &automock.ApplicationCRValidator{}
	defer validatorMock.AssertExpectations(t)
	validatorMock.ExpectOnValidate(&reCR).Run(fulfillExpectation).Once()

	mapperMock := &automock.ApplicationCRMapper{}
	defer mapperMock.AssertExpectations(t)
	mapperMock.ExpectOnToModel(&reCR, &reDM).Run(fulfillExpectation).Once()

	upserterMock := &automock.ApplicationUpserter{}
	defer upserterMock.AssertExpectations(t)
	upserterMock.ExpectOnUpsert(&reDM).Run(fulfillExpectation).Once()

	relistRequesterMock := &automock.SCRelistRequester{}
	defer relistRequesterMock.AssertExpectations(t)
	relistRequesterMock.ExpectOnRequestRelist().Run(fulfillExpectation).Once()

	syncJob := syncer.New(reInformer, upserterMock, nil, relistRequesterMock, spy.NewLogDummy()).
		WithCRValidator(validatorMock).
		WithCRMapper(mapperMock)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stopCh := make(chan struct{})
	defer close(stopCh)
	informerFactory.Start(stopCh)

	// when
	go syncJob.Run(ctx.Done())

	// then
	awaitForSyncGroupAtMost(t, expectations, 2*time.Second)
}

func awaitForSyncGroupAtMost(t *testing.T, wg *sync.WaitGroup, timeout time.Duration) {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()

	select {
	case <-c:
	case <-time.After(timeout):
		t.Fatalf("timeout occurred when waiting for sync group")
	}
}

func mustLoadCRFix(path string) v1alpha1.Application {
	in, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	var application v1alpha1.Application
	err = yaml.Unmarshal(in, &application)
	if err != nil {
		panic(err)
	}

	return application
}
