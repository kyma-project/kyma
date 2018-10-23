package integration

import (
	"context"
	publishapp "github.com/kyma-project/kyma/components/event-bus/cmd/event-bus-publish/application"
	pushapp "github.com/kyma-project/kyma/components/event-bus/cmd/event-bus-push/application"
	"github.com/kyma-project/kyma/components/event-bus/internal/publish"
	"github.com/kyma-project/kyma/components/event-bus/internal/push/opts"
	"github.com/kyma-project/kyma/components/event-bus/test/util"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

var (
	publishServer      *httptest.Server
	pushServer         *httptest.Server
	subscriberServerV1 *httptest.Server
	subscriberServerV2 *httptest.Server
)

const waitForSubscriptionToStart = 3 * time.Second

func TestMain(m *testing.M) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	stanServer, err := startNats()
	log.Printf("StanServer %v, error : %v \n", stanServer, err)
	if err != nil {
		panic(err)
	}

	publishOpts := publish.DefaultOptions()
	println(publishOpts)
	publishApplication := publishapp.NewPublishApplication(publishOpts)
	publishServer = httptest.NewServer(util.Logger(publishApplication.ServerMux))

	subscriberServerV1 = util.NewSubscriberServerV1()
	subscriberServerV2 = util.NewSubscriberServerV2()

	pushOpts := opts.DefaultOptions
	pushOpts.CheckEventsActivation = true
	pushOpts.NatsStreamingClusterID = clusterID
	pushApplication := pushapp.NewPushApplication(&pushOpts, newFakeInformer(ctx))

	pushServer = httptest.NewServer(util.Logger(pushApplication.ServerMux))

	retCode := m.Run()

	publishServer.Close()
	publishApplication.Stop()

	pushServer.Close()
	pushApplication.Stop()

	stopNats(stanServer)

	subscriberServerV1.Close()
	subscriberServerV2.Close()

	os.Exit(retCode)

}

func Test_Publish_Status(t *testing.T) {
	res, err := http.Get(publishServer.URL + publishServerStatusPath)
	checkIfError(err, t)
	verifyStatusCode(res, http.StatusOK, t)
}

func Test_Push_Status(t *testing.T) {
	res, err := http.Get(pushServer.URL + publishServerStatusPath)
	checkIfError(err, t)
	verifyStatusCode(res, http.StatusOK, t)
}

func Test_Subscriber_Status(t *testing.T) {
	res1, err1 := http.Get(subscriberServerV1.URL + util.SubServer1StatusPath)
	checkIfError(err1, t)
	verifyStatusCode(res1, http.StatusOK, t)
	res2, err2 := http.Get(subscriberServerV2.URL + util.SubServer2StatusPath)
	checkIfError(err2, t)
	verifyStatusCode(res2, http.StatusOK, t)
}

func Test_Publish_Push_Request(t *testing.T) {
	verifyPublishPushFlow("Test_Publish_Push_Request_1", "sub-1", "namespace-1", t)

	verifyPublishPushFlow("Test_Publish_Push_Request_2", "sub-2", "namespace-1", t)
}

func Test_sameSubjectSubscribersInDifferentNamespacesShouldReceiveEventsOfThatSubject(t *testing.T) {
	// create two subscriptions with the same name in two different namespaces to the same subject
	name := "test-subscription"
	ns1 := "namespace-1"
	ns2 := "namespace-2"
	eventData := "Test_sameSubjectSubscribersInDifferentNamespacesShouldReceiveEventsOfThatSubject"

	triggerCreateSub(name, ns1, subscriberServerV1.URL+util.SubServer1EventsPath, eventType, eventTypeVersion, sourceIDV1, t)

	triggerCreateSub(name, ns2, subscriberServerV2.URL+util.SubServer2EventsPath, eventType, eventTypeVersion, sourceIDV1, t)

	time.Sleep(waitForSubscriptionToStart)

	// publish one event
	payload := makePayload(sourceIDV1, eventType, eventTypeVersion, eventData)
	publishEvent(t, publishServer.URL, payload)

	// verify that both subscribers received the event
	verifyEndpointReceivedEvent(t, subscriberServerV1.URL+util.SubServer1ResultsPath, eventData)
	verifyEndpointReceivedEvent(t, subscriberServerV2.URL+util.SubServer2ResultsPath, eventData)

	triggerDeleteSub(name, ns1, t)
	triggerDeleteSub(name, ns2, t)
}

func Test_UpdateSubscriptionURL(t *testing.T) {
	name := "test-update"
	ns := "namespace-1"

	preUpdateEvent := "test-pre-update"
	triggerCreateSub(name, ns, subscriberServerV1.URL+util.SubServer1EventsPath, eventType, eventTypeVersion, sourceIDV1, t)
	time.Sleep(waitForSubscriptionToStart)

	publishEvent(t, publishServer.URL, makePayload(sourceIDV1, eventType, eventTypeVersion, preUpdateEvent))
	verifyEndpointReceivedEvent(t, subscriberServerV1.URL+util.SubServer1ResultsPath, preUpdateEvent)

	triggerUpdateSub(name, ns, subscriberServerV2.URL+util.SubServer2EventsPath, eventType, eventTypeVersion, sourceIDV1, t)
	time.Sleep(waitForSubscriptionToStart)

	postUpdateEvent := "test-post-update"
	publishEvent(t, publishServer.URL, makePayload(sourceIDV1, eventType, eventTypeVersion, postUpdateEvent))
	verifyEndpointReceivedEvent(t, subscriberServerV2.URL+util.SubServer2ResultsPath, postUpdateEvent)

	triggerDeleteSub(name, ns, t)

}

func verifyPublishPushFlow(eventData string, subscriptionName string, namespace string, t *testing.T) {
	payloadV1 := makePayload(sourceIDV1, eventType, eventTypeVersion, eventData)

	triggerCreateSub(subscriptionName, namespace, subscriberServerV1.URL+util.SubServer1EventsPath, eventType, eventTypeVersion, sourceIDV1, t)

	time.Sleep(waitForSubscriptionToStart)

	publishEvent(t, publishServer.URL, payloadV1)
	verifyEndpointReceivedEvent(t, subscriberServerV1.URL+util.SubServer1ResultsPath, eventData)

	triggerDeleteSub(subscriptionName, namespace, t)

}

func triggerCreateSub(subscriptionName string, namespace string, subscriberEventEndpointURL string, eventType string, eventTypeVersion string,
	sourceID string, t *testing.T) {
	_, err := createNewSubscription(subscriptionName, namespace, subscriberEventEndpointURL, eventType, eventTypeVersion, sourceID)
	checkIfError(err, t)
}

func triggerDeleteSub(subscriptionName string, namespace string, t *testing.T) {
	err := deleteSubscription(subscriptionName, namespace)
	checkIfError(err, t)
}

func triggerUpdateSub(subscriptionName string, namespace string, subscriberEventEndpointURL string, eventType string, eventTypeVersion string,
	sourceID string, t *testing.T) {
	_, err := updateSubscription(subscriptionName, namespace, subscriberEventEndpointURL, eventType, eventTypeVersion, sourceID)
	checkIfError(err, t)
}
