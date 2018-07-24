package syncer_test

import (
	"context"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/remote-environment-broker/internal/syncer"
	"github.com/kyma-project/kyma/components/remote-environment-broker/internal/syncer/automock"
	"github.com/kyma-project/kyma/components/remote-environment-broker/platform/logger/spy"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRelistRequesterRequestRelistSuccessSingleTrigger(t *testing.T) {
	// given
	fixBrokerName := "fix-broker-name"
	fixRelistDuration := time.Microsecond

	logSink := newLogSinkForErrors()

	syncCalled := make(chan struct{})
	fulfillExpectation := func(mock.Arguments) {
		close(syncCalled)
	}

	scSyncerMock := &automock.ServiceCatalogSyncer{}
	defer scSyncerMock.AssertExpectations(t)
	scSyncerMock.On("Sync", fixBrokerName, mock.AnythingOfType("int")).
		Run(fulfillExpectation).
		Return(nil).
		Once()

	relister := syncer.NewRelistRequester(scSyncerMock, fixBrokerName, fixRelistDuration, logSink.Logger)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	go relister.Run(ctx.Done())

	// when
	relister.RequestRelist()

	// then
	awaitForChanAtMost(t, syncCalled, time.Second)
	assert.Empty(t, logSink.DumpAll())
}

func TestRelistRequesterRequestRelistSuccessMultipleTrigger(t *testing.T) {
	// given
	fixBrokerName := "fix-broker-name"
	fixRelistDuration := time.Second

	syncCalled := make(chan struct{})
	fulfillExpectation := func(mock.Arguments) {
		close(syncCalled)
	}

	scSyncerMock := &automock.ServiceCatalogSyncer{}
	defer scSyncerMock.AssertExpectations(t)
	scSyncerMock.On("Sync", fixBrokerName, mock.AnythingOfType("int")).
		Run(fulfillExpectation).
		Return(nil)

	afterChan := make(chan time.Time, 1)
	afterTimeMock := func(d time.Duration) <-chan time.Time {
		return afterChan
	}

	relister := syncer.NewRelistRequester(scSyncerMock, fixBrokerName, fixRelistDuration, spy.NewLogDummy()).
		WithTimeAfter(afterTimeMock)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	go relister.Run(ctx.Done())

	// when
	for i := 0; i < 10; i++ { // trigger request multiple times, no one should be blocking
		relister.RequestRelist()
	}

	afterChan <- time.Now() // simulate that time has expired

	// then
	awaitForChanAtMost(t, syncCalled, time.Second)
	scSyncerMock.AssertNumberOfCalls(t, "Sync", 1)
	assert.Empty(t, afterChan, "timeAfter was not called when it should be")
}

func newLogSinkForErrors() *spy.LogSink {
	logSink := spy.NewLogSink()
	logSink.Logger.Logger.Level = logrus.ErrorLevel
	return logSink
}

func awaitForChanAtMost(t *testing.T, ch <-chan struct{}, timeout time.Duration) {
	select {
	case <-ch:
	case <-time.After(timeout):
		t.Fatalf("timeout occurred when waiting for channel")
	}
}
