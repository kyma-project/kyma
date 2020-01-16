package backupe2e

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/avast/retry-go"
	publishApi "github.com/kyma-project/kyma/components/event-bus/api/publish"
	subApis "github.com/kyma-project/kyma/components/event-bus/apis/eventing/v1alpha1"
	ebClientSet "github.com/kyma-project/kyma/components/event-bus/client/generated/clientset/internalclientset"

	"github.com/kyma-project/kyma/components/event-bus/test/util"
	"github.com/sirupsen/logrus"
	apiV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sClientSet "k8s.io/client-go/kubernetes"

	"github.com/kyma-project/kyma/tests/end-to-end/backup-restore-test/utils/config"

	"github.com/smartystreets/goconvey/convey"
)

const (
	eventType           = "test-e2e"
	subscriptionName    = "test-sub"
	eventActivationName = "test-ea"
	srcID               = "test.local"

	subscriberName           = "test-core-event-bus-subscriber"
	subscriberImage          = "eu.gcr.io/kyma-project/pr/event-bus-e2e-subscriber:PR-4893"
	publishEventEndpointURL  = "http://event-publish-service.kyma-system:8080/v1/events"
	publishStatusEndpointURL = "http://event-publish-service.kyma-system:8080/v1/status/ready"
)

var retryOptions = []retry.Option{
	retry.Attempts(13), // at max (100 * (1 << 13)) / 1000 = 819,2 sec
	retry.OnRetry(func(n uint, err error) {
		fmt.Printf(".")
	}),
}

// EventBusTest tests the Event Bus business logic after restoring Kyma from a backup
type EventBusTest struct {
	K8sInterface k8sClientSet.Interface
	EbInterface  ebClientSet.Interface
}

type eventBusFlow struct {
	namespace string
	log       logrus.FieldLogger

	k8sInterface k8sClientSet.Interface
	ebInterface  ebClientSet.Interface
}

// NewEventBusTest returns new instance of the EventBusTest
func NewEventBusTest() (*EventBusTest, error) {
	k8sConfig, err := config.NewRestClientConfig()
	if err != nil {
		return nil, err
	}

	k8sCli, err := k8sClientSet.NewForConfig(k8sConfig)
	if err != nil {
		return nil, err
	}

	ebCli, err := ebClientSet.NewForConfig(k8sConfig)
	if err != nil {
		return nil, err
	}

	return &EventBusTest{
		K8sInterface: k8sCli,
		EbInterface:  ebCli,
	}, nil
}

// CreateResources creates resources needed for e2e backup test
func (eb *EventBusTest) CreateResources(namespace string) {
	err := eb.newFlow(namespace).createResources()
	convey.So(err, convey.ShouldBeNil)
}

// TestResources tests resources after restoring from a backup
func (eb *EventBusTest) TestResources(namespace string) {
	err := eb.newFlow(namespace).testResources()
	convey.So(err, convey.ShouldBeNil)
}

func (eb *EventBusTest) newFlow(namespace string) *eventBusFlow {

	logger := logrus.New()
	// configure logger with text instead of json for easier reading in CI logs
	logger.Formatter = &logrus.TextFormatter{}
	// show file and line number
	logger.SetReportCaller(true)
	res := &eventBusFlow{
		namespace:    namespace,
		k8sInterface: eb.K8sInterface,
		ebInterface:  eb.EbInterface,
		log:          logger,
	}
	return res
}

func (f *eventBusFlow) createResources() error {
	// iterate over steps
	for _, fn := range []func() error{
		f.createEventActivation,
		f.createSubscription,
		f.createSubscriber,
		f.checkSubscriberStatus,
		f.checkPublisherStatus,
		f.checkSubscriptionReady,
		f.publishTestEvent,
		f.checkSubscriberReceivedEvent,
	} {
		if err := fn(); err != nil {
			return fmt.Errorf("CreateResources() failed with: %v", err)
		}
	}
	return nil
}

func (f *eventBusFlow) testResources() error {
	// iterate over steps
	for _, fn := range []func() error{
		f.checkSubscriberStatus,
		f.checkPublisherStatus,
		f.checkSubscriptionReady,
		f.publishTestEvent,
		f.checkSubscriberReceivedEvent,
	} {
		if err := fn(); err != nil {
			return fmt.Errorf("TestResources() failed with: %v", err)
		}
	}
	return nil
}

func (f *eventBusFlow) createSubscriber() error {
	if _, err := f.k8sInterface.AppsV1().Deployments(f.namespace).Get(subscriberName, metaV1.GetOptions{}); err != nil {
		if _, err := f.k8sInterface.AppsV1().Deployments(f.namespace).Create(util.NewSubscriberDeployment(subscriberImage)); err != nil {
			return fmt.Errorf("create Subscriber deployment: %v", err)
		}

		if _, err := f.k8sInterface.CoreV1().Services(f.namespace).Create(util.NewSubscriberService()); err != nil {
			return fmt.Errorf("create Subscriber service failed: %v", err)
		}
		return retry.Do(func() error {
			pods, err := f.k8sInterface.CoreV1().Pods(f.namespace).List(metaV1.ListOptions{LabelSelector: "app=" + subscriberName})
			if err != nil {
				return err
			}
			for _, pod := range pods.Items {
				if !isPodReady(&pod) {
					return fmt.Errorf("pod is not ready: %v", pod)
				}
			}
			return nil
		}, retryOptions...)
	}
	return nil
}

// Create an EventActivation
// Retry in case of any error except resource already exists
func (f *eventBusFlow) createEventActivation() error {
	eventActivation := util.NewEventActivation(eventActivationName, f.namespace, srcID)

	return retry.Do(func() error {
		_, err := f.ebInterface.ApplicationconnectorV1alpha1().EventActivations(f.namespace).Create(eventActivation)
		if err == nil {
			return nil
		}
		if !strings.Contains(err.Error(), "already exists") {
			return fmt.Errorf("waiting for event activation %v to exist", eventActivation)
		}
		return nil
	}, retryOptions...)
}

func (f *eventBusFlow) createSubscription() error {
	subscriberEventEndpointURL := "http://" + subscriberName + "." + f.namespace + ":9000/v1/events"
	return retry.Do(func() error {
		if _, err := f.ebInterface.EventingV1alpha1().Subscriptions(f.namespace).Create(util.NewSubscription(subscriptionName, f.namespace, subscriberEventEndpointURL, eventType, "v1", srcID)); err != nil {
			if !strings.Contains(err.Error(), "already exists") {
				return fmt.Errorf("error in creating subscription: %v", err)
			}
		}
		return nil
	}, retryOptions...)
}

// Check the subscriber status until the http get call succeeds and until status code is 200
func (f *eventBusFlow) checkSubscriberStatus() error {
	subscriberStatusEndpointURL := "http://" + subscriberName + "." + f.namespace + ":9000/v1/status"
	return retry.Do(func() error {
		return checkStatus(subscriberStatusEndpointURL)
	}, retryOptions...)
}

func (f *eventBusFlow) checkPublisherStatus() error {
	return retry.Do(func() error {
		return checkStatus(publishStatusEndpointURL)
	}, retryOptions...)
}

func (f *eventBusFlow) checkSubscriptionReady() error {
	activatedCondition := subApis.SubscriptionCondition{Type: subApis.Ready, Status: subApis.ConditionTrue}
	return retry.Do(func() error {
		kySub, err := f.ebInterface.EventingV1alpha1().Subscriptions(f.namespace).Get(subscriptionName, metaV1.GetOptions{})
		if err != nil {
			return fmt.Errorf("cannot get Kyma subscription, name: %v; namespace: %v", subscriptionName, f.namespace)
		}
		if kySub.HasCondition(activatedCondition) {
			return nil
		}
		return fmt.Errorf("subscription %v does not have condition %+v", kySub, activatedCondition)
	}, retryOptions...)
}

func (f *eventBusFlow) publishTestEvent() error {
	return retry.Do(func() error {
		_, err := f.publish(publishEventEndpointURL)
		return err
	}, retryOptions...)
}

func (f *eventBusFlow) publish(publishEventURL string) (*publishApi.Response, error) {
	payload := fmt.Sprintf(
		`{"source-id": "%s","event-type":"%s","event-type-version":"v1","event-time":"2018-11-02T22:08:41+00:00","data":"test-event-1"}`, srcID, eventType)
	res, err := http.Post(publishEventURL, "application/json", strings.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("post request failed: %v", err)
	}

	if err := verifyStatusCode(res, 200); err != nil {
		return nil, err
	}
	respObj := &publishApi.Response{}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			f.log.Error(err)
		}
	}()
	err = json.Unmarshal(body, &respObj)
	if err != nil {
		return nil, fmt.Errorf("unmarshal error: %v", err)
	}

	if len(respObj.EventID) == 0 {
		return nil, fmt.Errorf("empty respObj.EventID")
	}
	return respObj, err
}

func (f *eventBusFlow) checkSubscriberReceivedEvent() error {
	subscriberResultsEndpointURL := "http://" + subscriberName + "." + f.namespace + ":9000/v1/results"
	return retry.Do(func() error {
		res, err := http.Get(subscriberResultsEndpointURL)
		if err != nil {
			return fmt.Errorf("get request failed: %v", err)
		}

		if err := verifyStatusCode(res, 200); err != nil {
			return fmt.Errorf("get request failed: %v", err)
		}
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}
		var resp string
		err = json.Unmarshal(body, &resp)
		if err != nil {
			return err
		}

		err = res.Body.Close()
		if err != nil {
			return err
		}

		if len(resp) == 0 {
			return errors.New("no event received by subscriber")
		}
		if resp != "test-event-1" {
			return fmt.Errorf("wrong response: %s, want: %s", resp, "test-event-1")
		}
		return nil
	}, retryOptions...)
}

func checkStatus(statusEndpointURL string) error {
	res, err := http.Get(statusEndpointURL)
	if err != nil {
		return err
	}
	return verifyStatusCode(res, http.StatusOK)
}

// Verify that the http response has the given status code and return an error if not
func verifyStatusCode(res *http.Response, expectedStatusCode int) error {
	if res.StatusCode != expectedStatusCode {
		return fmt.Errorf("status code is wrong, have: %d, want: %d", res.StatusCode, expectedStatusCode)
	}
	return nil
}

func isPodReady(pod *apiV1.Pod) bool {
	for _, cs := range pod.Status.ContainerStatuses {
		if !cs.Ready {
			return false
		}
	}
	return true
}
