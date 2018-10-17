package integration

import (
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	publishapp "github.com/kyma-project/kyma/components/event-bus/cmd/event-bus-publish/application"
	pushapp "github.com/kyma-project/kyma/components/event-bus/cmd/event-bus-push/application"
	"github.com/kyma-project/kyma/components/event-bus/internal/common"
	"github.com/kyma-project/kyma/components/event-bus/internal/publish"
	"github.com/kyma-project/kyma/components/event-bus/internal/push/actors"
	"github.com/kyma-project/kyma/components/event-bus/internal/push/opts"
	"github.com/kyma-project/kyma/components/event-bus/test/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	publishServer            *httptest.Server
	pushServer               *httptest.Server
	subscriberServerV1       *httptest.Server
	subscriberServerV2       *httptest.Server
	subscriptionsSupervisor1 *actors.SubscriptionsSupervisor
	subscriptionsSupervisor2 *actors.SubscriptionsSupervisor
)

func TestMain(m *testing.M) {

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
	pushOpts.NatsStreamingClusterID = clusterID
	pushApplication := pushapp.NewPushApplication(&pushOpts, newFakeInformer2())
	subscriptionsSupervisor1 = pushApplication.SubscriptionsSupervisor
	subscription1 := util.NewSubscription("test-sub", metav1.NamespaceDefault, subscriberServerV1.URL+util.SubServer1EventsPath, eventType, eventTypeVersion,
		sourceIDV1)
	subscriptionsSupervisor1.StartSubscriptionReq(subscription1, common.DefaultRequestProvider)

	subscriptionsSupervisor2 = pushApplication.SubscriptionsSupervisor
	subscription2 := util.NewSubscription("test-sub", metav1.NamespaceDefault, subscriberServerV2.URL+util.SubServer2EventsPath, eventType, eventTypeVersion,
		sourceIDV2)
	subscriptionsSupervisor2.StartSubscriptionReq(subscription2, common.DefaultRequestProvider)

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
	{
		payloadV1 := makePayload(sourceIDV1, eventType, eventTypeVersion, eventDataV1)
		publishEvent(t, publishServer.URL, payloadV1)
		verifyEndpointReceivedEvent(t, subscriberServerV1.URL+util.SubServer1ResultsPath, eventDataV1)
	}

	{
		payloadV2 := makePayload(sourceIDV2, eventType, eventTypeVersion, eventDataV2)
		publishEvent(t, publishServer.URL, payloadV2)
		verifyEndpointReceivedEvent(t, subscriberServerV2.URL+util.SubServer2ResultsPath, eventDataV2)

	}
}

func Test_pushRequestShouldNotIncludeKymaTopicHeader(t *testing.T) {
	var pushRequest *http.Request
	requestProvider := common.RequestProvider(func(method, url string, body io.Reader) (*http.Request, error) {
		var err error
		pushRequest, err = http.NewRequest(method, url, body)
		return pushRequest, err
	})

	subscription1 := util.NewSubscription("test-sub-1", metav1.NamespaceDefault, subscriberServerV1.URL+util.SubServer1EventsPath, eventType, eventTypeVersion,
		sourceIDV1)
	subscriptionsSupervisor1.StartSubscriptionReq(subscription1, requestProvider)

	payloadV1 := makePayload(sourceIDV1, eventType, eventTypeVersion, eventDataV1)
	publishEvent(t, publishServer.URL, payloadV1)

	verifyEndpointReceivedEvent(t, subscriberServerV1.URL+util.SubServer1ResultsPath, eventDataV1)

	if pushRequest == nil {
		t.Fatal("push request should not be nil")
	}

	if header := pushRequest.Header.Get(headerKymaTopic); len(header) > 0 {
		t.Fatalf("request to endpoint should not include the header %s", headerKymaTopic)
	}
}

func Test_sameSubjectSubscribersInDifferentNamespacesShouldReceiveEventsOfThatSubject(t *testing.T) {
	// create two subscriptions with the same name in two different namespaces to the same subject
	name := "test-subscription"
	subscription1 := util.NewSubscription(name, "namespace-1", subscriberServerV1.URL+util.SubServer1EventsPath, eventType, eventTypeVersion, sourceIDV1)
	subscription2 := util.NewSubscription(name, "namespace-2", subscriberServerV2.URL+util.SubServer2EventsPath, eventType, eventTypeVersion, sourceIDV1)

	// handle the two subscriptions
	subscriptionsSupervisor1.StartSubscriptionReq(subscription1, common.DefaultRequestProvider)
	subscriptionsSupervisor2.StartSubscriptionReq(subscription2, common.DefaultRequestProvider)

	// publish one event
	payload := makePayload(sourceIDV1, eventType, eventTypeVersion, eventDataV1)
	publishEvent(t, publishServer.URL, payload)

	// verify that both subscribers received the event
	verifyEndpointReceivedEvent(t, subscriberServerV1.URL+util.SubServer1ResultsPath, eventDataV1)
	verifyEndpointReceivedEvent(t, subscriberServerV2.URL+util.SubServer2ResultsPath, eventDataV1)
}
