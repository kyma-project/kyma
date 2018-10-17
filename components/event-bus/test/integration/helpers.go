package integration

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"testing"
	"time"

	api "github.com/kyma-project/kyma/components/event-bus/api/publish"
	"github.com/kyma-project/kyma/components/event-bus/generated/push/clientset/versioned/fake"
	"github.com/kyma-project/kyma/components/event-bus/generated/push/informers/externalversions/eventing.kyma.cx/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
	"context"
	"github.com/nats-io/nats-streaming-server/server"
)

const (
	clusterID               = "kyma-nats-streaming"
	eventType               = "test-publish-push-success"
	eventTypeVersion        = "v1"
	sourceIDV1              = "test.local.kyma.commerce.ec"
	eventDataV1             = "test-event-1"
	sourceIDV2              = "test.local.kyma.commerce.ec"
	eventDataV2             = "test-event-2"
	publishServerStatusPath = "/v1/status/ready"
	headerKymaTopic         = "kyma-topic"
)

func startNats() (*server.StanServer, error) {
	return server.RunServer(clusterID)
}

func stopNats(stanServer *server.StanServer) {
	stanServer.Shutdown()
}

func makePayload(sourceID, eventType, eventTypeVersion, eventData string) string {
	return fmt.Sprintf(`{"source-id": "%s", "event-type": "%s","event-type-version": "%s","event-time": "2018-11-02T22:08:41+00:00","data": "%s"}`,
		sourceID, eventType, eventTypeVersion, eventData)
}

//func newFakeInformer() cache.SharedIndexInformer {
//	sub := util.NewSubscription(
//		"test-sub",
//		metav1.NamespaceDefault,
//		subscriberServerV1.URL+util.SubServer1EventsPath,
//		eventType,
//		eventTypeVersion,
//		sourceIDV1)
//	clientSet := fake.NewSimpleClientset(sub)
//	informer := v1alpha1.NewSubscriptionInformer(clientSet, metav1.NamespaceAll, 0, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
//	informer.GetIndexer().Add(sub)
//	return informer
//}

func newFakeInformer2() cache.SharedIndexInformer {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

	client := fake.NewSimpleClientset()

	informers := v1alpha1.NewSubscriptionInformer(client, metav1.NamespaceAll, 1*time.Minute, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})

	informers.Run(ctx.Done())

	informer := v1alpha1.NewSubscriptionInformer(client, metav1.NamespaceAll, 0, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})

	if !informer.HasSynced() {
		time.Sleep(10 * time.Millisecond)
	}

	//client.EventingV1alpha1().Subscriptions("").Create()

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

func publishEvent(t *testing.T, publishServerURL string, payload string) {
	res, err := http.Post(publishServerURL+"/v1/events", "application/json", strings.NewReader(payload))
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
}

func verifyEndpointReceivedEvent(t *testing.T, endpoint, data string) {
	var ok bool
	for i := 0; i < 20; i++ {
		time.Sleep(1 * time.Second)
		res, err := http.Get(endpoint)
		assert.Nil(t, err)
		body, err := ioutil.ReadAll(res.Body)
		var resp string
		json.Unmarshal(body, &resp)
		res.Body.Close()
		if len(resp) == 0 {
			continue
		}
		assert.Equal(t, data, resp)
		ok = true
		break
	}
	assert.True(t, ok)
}
