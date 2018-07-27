package application_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	api "github.com/kyma-project/kyma/components/event-bus/api/publish"
	publishapp "github.com/kyma-project/kyma/components/event-bus/cmd/event-bus-publish/application"
	pushapp "github.com/kyma-project/kyma/components/event-bus/cmd/event-bus-push/application"
	"github.com/kyma-project/kyma/components/event-bus/generated/push/clientset/versioned/fake"
	"github.com/kyma-project/kyma/components/event-bus/generated/push/informers/externalversions/eventing.kyma.cx/v1alpha1"
	"github.com/kyma-project/kyma/components/event-bus/internal/publish"
	"github.com/kyma-project/kyma/components/event-bus/internal/push/opts"
	"github.com/kyma-project/kyma/components/event-bus/test/util"
	"github.com/nats-io/nats-streaming-server/server"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
)

const (
	clusterID        = "kyma-nats-streaming"
	eventType        = "test-publish-push-success"
	eventTypeVersion = "v1"
	sourceType       = "ec"

	sourceEnvironmentV1 = "test"
	sourceNamespaceV1   = "local.kyma.commerce"
	eventDataV1         = "test-event-1"

	sourceEnvironmentV2 = "test.com"
	sourceNamespaceV2   = "sap.hybris"
	eventDataV2         = "test-event-2"

	publishServerStatusPath = "/v1/status/ready"
)

var (
	publishServer      *httptest.Server
	pushServer         *httptest.Server
	subscriberServerV1 *httptest.Server
	subscriberServerV2 *httptest.Server
)

func startNats() (*server.StanServer, error) {
	return server.RunServer(clusterID)
}

func stopNats(stanServer *server.StanServer) {
	stanServer.Shutdown()
}

func TestMain(m *testing.M) {

	stanServer, err := startNats()

	publishOpts := publish.DefaultOptions()
	println(publishOpts)
	publishApplication := publishapp.NewPublishApplication(publishOpts)
	publishServer = httptest.NewServer(util.Logger(publishApplication.ServerMux))

	subscriberServerV1 = util.NewSubscriberServerV1()
	subscriberServerV2 = util.NewSubscriberServerV2()

	pushOpts := opts.DefaultOptions
	pushOpts.NatsStreamingClusterID = clusterID
	pushApplication := pushapp.NewPushApplication(&pushOpts, newFakeInformer())
	subscriptionsSupervisor1 := pushApplication.SubscriptionsSupervisor
	subscriptionsSupervisor1.StartSubscriptionReq(
		util.NewSubscription(
			"test-sub",
			metav1.NamespaceDefault,
			subscriberServerV1.URL+util.SubServer1EventsPath,
			eventType,
			eventTypeVersion,
			sourceEnvironmentV1,
			sourceNamespaceV1,
			sourceType))

	subscriptionsSupervisor2 := pushApplication.SubscriptionsSupervisor
	subscriptionsSupervisor2.StartSubscriptionReq(
		util.NewSubscription(
			"test-sub",
			metav1.NamespaceDefault,
			subscriberServerV2.URL+util.SubServer2EventsPath,
			eventType,
			eventTypeVersion,
			sourceEnvironmentV2,
			sourceNamespaceV2,
			sourceType))

	pushServer = httptest.NewServer(util.Logger(pushApplication.ServerMux))

	if err != nil {
		panic(err)
	} else {
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

func makePayload(
	sourceNamespace string,
	sourceType string,
	sourceEnvironment string,
	eventType string,
	eventTypeVersion string,
	eventData string) string {
	return fmt.Sprintf(
		`{"source": {"source-namespace": "%s","source-type": "%s","source-environment": "%s"},
		"event-type": "%s","event-type-version": "%s","event-time": "2018-11-02T22:08:41+00:00","data": "%s"}`,
		sourceNamespace, sourceType, sourceEnvironment, eventType, eventTypeVersion, eventData)
}

func Test_Publish_Push_Request(t *testing.T) {
	{
		payloadV1 := makePayload(sourceNamespaceV1, sourceType, sourceEnvironmentV1, eventType, eventTypeVersion, eventDataV1)
		res, err := http.Post(publishServer.URL+"/v1/events", "application/json", strings.NewReader(payloadV1))
		checkIfError(err, t)
		verifyStatusCode(res, 200, t)
		log.Print(res)
		respObj := &api.PublishResponse{}
		body, err := ioutil.ReadAll(res.Body)
		defer res.Body.Close()
		err = json.Unmarshal(body, &respObj)
		assert.NotNil(t, respObj.EventID)
		assert.NotEmpty(t, respObj.EventID)
		log.Printf("%v", respObj)

		var ok bool
		for i := 0; i < 10; i++ {
			time.Sleep(1 * time.Second)
			res, err := http.Get(subscriberServerV1.URL + util.SubServer1ResultsPath)
			assert.Nil(t, err)
			body, err := ioutil.ReadAll(res.Body)
			var resp string
			json.Unmarshal(body, &resp)
			res.Body.Close()
			if len(resp) == 0 {
				continue
			}
			assert.Equal(t, eventDataV1, resp)
			ok = true
			break
		}
		assert.True(t, ok)
	}

	{
		payloadV2 := makePayload(sourceNamespaceV2, sourceType, sourceEnvironmentV2, eventType, eventTypeVersion, eventDataV2)
		res, err := http.Post(publishServer.URL+"/v1/events", "application/json", strings.NewReader(payloadV2))
		checkIfError(err, t)
		verifyStatusCode(res, 200, t)
		log.Print(res)
		respObj := &api.PublishResponse{}
		body, err := ioutil.ReadAll(res.Body)
		defer res.Body.Close()
		err = json.Unmarshal(body, &respObj)
		assert.NotNil(t, respObj.EventID)
		assert.NotEmpty(t, respObj.EventID)
		log.Printf("%v", respObj)

		var ok bool
		for i := 0; i < 10; i++ {
			time.Sleep(1 * time.Second)
			res, err := http.Get(subscriberServerV2.URL + util.SubServer2ResultsPath)
			assert.Nil(t, err)
			body, err := ioutil.ReadAll(res.Body)
			var resp string
			json.Unmarshal(body, &resp)
			res.Body.Close()
			if len(resp) == 0 {
				continue
			}
			assert.Equal(t, eventDataV2, resp)
			ok = true
			break
		}
		assert.True(t, ok)
	}
}

func newFakeInformer() cache.SharedIndexInformer {
	sub := util.NewSubscription(
		"test-sub",
		metav1.NamespaceDefault,
		subscriberServerV1.URL+util.SubServer1EventsPath,
		eventType,
		eventTypeVersion,
		sourceEnvironmentV1,
		sourceNamespaceV1,
		sourceType)
	clientSet := fake.NewSimpleClientset(sub)
	informer := v1alpha1.NewSubscriptionInformer(clientSet, metav1.NamespaceAll, 0, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	informer.GetIndexer().Add(sub)
	return informer
}

func verifyStatusCode(res *http.Response, expectedStatusCode int, t *testing.T) {
	if res.StatusCode != expectedStatusCode {
		t.Errorf("Status code is wrong, have: %d, want: %d", res.StatusCode, expectedStatusCode)
	}
}

func checkIfError(err error, t *testing.T) {
	if err != nil {
		t.Fatal(err)
	}
}
