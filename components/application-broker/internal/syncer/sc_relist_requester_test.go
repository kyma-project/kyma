package syncer_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/application-broker/internal/syncer"
	"github.com/kyma-project/kyma/components/application-broker/internal/syncer/automock"
	"github.com/kyma-project/kyma/components/application-broker/platform/logger/spy"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewRelistRequesterSuccess(t *testing.T) {
	// given
	fixBrokerLabelKey := "app"
	fixBrokerLabelValue := "label"
	labelSelector := fmt.Sprintf("%s=%s", fixBrokerLabelKey, fixBrokerLabelValue)
	fixRelistDuration := time.Microsecond

	logSink := newLogSinkForErrors()

	syncCalled := make(chan struct{})
	fulfillExpectation := func(mock.Arguments) {
		close(syncCalled)
	}

	brokerSyncer := &automock.BrokerSyncer{}
	brokerSyncer.On("Sync", labelSelector, 5).
		Run(fulfillExpectation).Return(nil)
	defer brokerSyncer.AssertExpectations(t)

	relister := syncer.NewRelistRequester(brokerSyncer, fixRelistDuration, fixBrokerLabelKey, fixBrokerLabelValue, logSink.Logger)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	go relister.Run(ctx.Done())

	// when
	relister.RequestRelist()

	// then
	awaitForChanAtMost(t, syncCalled, time.Second)
	assert.Empty(t, logSink.DumpAll())
}

func TestNewRelistRequesterError(t *testing.T) {
	// given
	fixBrokerLabelKey := "app"
	fixBrokerLabelValue := "label"
	labelSelector := fmt.Sprintf("%s=%s", fixBrokerLabelKey, fixBrokerLabelValue)
	fixRelistDuration := time.Microsecond
	maxRetries := 5

	logSink := newLogSinkForErrors()

	syncCalled := make(chan struct{})
	fulfillExpectation := func(mock.Arguments) {
		close(syncCalled)
	}

	brokerSyncer := &automock.BrokerSyncer{}
	brokerSyncer.On("Sync", labelSelector, maxRetries).
		Run(fulfillExpectation).Return(errors.New("fix"))
	defer brokerSyncer.AssertExpectations(t)

	relister := syncer.NewRelistRequester(brokerSyncer, fixRelistDuration, fixBrokerLabelKey, fixBrokerLabelValue, logSink.Logger)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	go relister.Run(ctx.Done())

	// when
	relister.RequestRelist()

	// then
	awaitForChanAtMost(t, syncCalled, time.Second)
	syncer.WaitAtMost(func() (bool, error) {
		if len(logSink.DumpAll()) > 0 {
			return true, nil
		}
		return false, nil
	}, time.Second)

	logSink.AssertLogged(t, logrus.ErrorLevel, fmt.Sprintf("Error occurred when synchronizing ServiceBrokers [labelSelector: %s][maxRetires=%d]: %v", labelSelector, maxRetries, fixError()))
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

func fixError() error {
	return errors.New("fix")
}
