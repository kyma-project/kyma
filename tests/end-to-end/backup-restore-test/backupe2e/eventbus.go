package backupe2e

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	apiV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	publishApi "github.com/kyma-project/kyma/components/event-bus/api/publish"
	subApis "github.com/kyma-project/kyma/components/event-bus/api/push/eventing.kyma-project.io/v1alpha1"
	eaClientSet "github.com/kyma-project/kyma/components/event-bus/generated/ea/clientset/versioned"
	subscriptionClientSet "github.com/kyma-project/kyma/components/event-bus/generated/push/clientset/versioned"
	"github.com/kyma-project/kyma/components/event-bus/test/util"
	"github.com/kyma-project/kyma/tests/end-to-end/backup-restore-test/utils/config"
	k8sClientSet "k8s.io/client-go/kubernetes"

	"github.com/smartystreets/goconvey/convey"
)

const (
	eventType           = "test-e2e"
	subscriptionName    = "test-sub"
	eventActivationName = "test-ea"
	srcID               = "test.local"

	noOfRetries = 20

	subscriberName           = "test-core-event-bus-subscriber"
	subscriberImage          = "eu.gcr.io/kyma-project/event-bus-e2e-subscriber:0.9.0"
	publishEventEndpointURL  = "http://event-bus-publish.kyma-system:8080/v1/events"
	publishStatusEndpointURL = "http://event-bus-publish.kyma-system:8080/v1/status/ready"
)

// EventBusTest tests the Event Bus business logic after restoring Kyma from a backup
type EventBusTest struct {
	K8sInterface  k8sClientSet.Interface
	EaInterface   eaClientSet.Interface
	SubsInterface subscriptionClientSet.Interface
}

type eventBusFlow struct {
	namespace string

	k8sInterface  k8sClientSet.Interface
	eaInterface   eaClientSet.Interface
	subsInterface subscriptionClientSet.Interface
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

	subCli, err := subscriptionClientSet.NewForConfig(k8sConfig)
	if err != nil {
		return nil, err
	}

	eaCli, err := eaClientSet.NewForConfig(k8sConfig)
	if err != nil {
		return nil, err
	}

	return &EventBusTest{
		K8sInterface:  k8sCli,
		EaInterface:   eaCli,
		SubsInterface: subCli,
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

// DeleteResources deletes resources before restoring from a backup
func (eb *EventBusTest) DeleteResources(namespace string) {
	err := eb.newFlow(namespace).deleteResources()
	convey.So(err, convey.ShouldBeNil)
}

func (eb *EventBusTest) newFlow(namespace string) *eventBusFlow {
	return &eventBusFlow{
		namespace:     namespace,
		k8sInterface:  eb.K8sInterface,
		eaInterface:   eb.EaInterface,
		subsInterface: eb.SubsInterface,
	}
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
		// f.cleanup,
	} {
		if err := fn(); err != nil {
			return fmt.Errorf("TestResources() failed with: %v", err)
		}
	}
	return nil
}

func (f *eventBusFlow) deleteResources() error {
	err := f.cleanup()
	if err != nil {
		return fmt.Errorf("DeleteResources() failed with: %v", err)
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
		time.Sleep(30 * time.Second)

		for i := 0; i < 60; i++ {
			var podReady bool
			if pods, err := f.k8sInterface.CoreV1().Pods(f.namespace).List(metaV1.ListOptions{LabelSelector: "app=" + subscriberName}); err == nil {
				for _, pod := range pods.Items {
					if podReady = isPodReady(&pod); !podReady {
						break
					}
				}
			}
			if podReady {
				break
			} else {
				time.Sleep(1 * time.Second)
			}
		}
	}
	return nil
}

func (f *eventBusFlow) createEventActivation() error {
	var err error
	for i := 0; i < noOfRetries; i++ {
		_, err = f.eaInterface.ApplicationconnectorV1alpha1().EventActivations(f.namespace).Create(util.NewEventActivation(eventActivationName, f.namespace, srcID))
		if err == nil {
			break
		}
		if !strings.Contains(err.Error(), "already exists") {
			time.Sleep(1 * time.Second)
		} else {
			break
		}
	}
	return err
}

func (f *eventBusFlow) createSubscription() error {
	subscriberEventEndpointURL := "http://" + subscriberName + "." + f.namespace + ":9000/v1/events"
	_, err := f.subsInterface.EventingV1alpha1().Subscriptions(f.namespace).Create(util.NewSubscription(subscriptionName, f.namespace, subscriberEventEndpointURL, eventType, "v1", srcID))
	if err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			return fmt.Errorf("error in creating subscription: %v", err)
		}
	}
	return err
}

func (f *eventBusFlow) checkSubscriberStatus() error {
	subscriberStatusEndpointURL := "http://" + subscriberName + "." + f.namespace + ":9000/v1/status"
	var err error
	for i := 0; i < noOfRetries; i++ {
		if res, err := http.Get(subscriberStatusEndpointURL); err != nil {
			time.Sleep(time.Duration(i) * time.Second)
		} else if !checkStatusCode(res, http.StatusOK) {
			time.Sleep(time.Duration(i) * time.Second)
		} else {
			break
		}
	}
	return err
}

func (f *eventBusFlow) checkPublisherStatus() error {
	var err error
	for i := 0; i < noOfRetries; i++ {
		if err = checkStatus(publishStatusEndpointURL); err != nil {
			time.Sleep(time.Duration(i) * time.Second)
		} else {
			break
		}
	}
	return err
}

func (f *eventBusFlow) checkSubscriptionReady() error {
	var err error
	activatedCondition := subApis.SubscriptionCondition{Type: subApis.Ready, Status: subApis.ConditionTrue}
	for i := 0; i < noOfRetries; i++ {
		kySub, err := f.subsInterface.EventingV1alpha1().Subscriptions(f.namespace).Get(subscriptionName, metaV1.GetOptions{})
		if err != nil {
			return fmt.Errorf("cannot get Kyma subscription, name: %v; namespace: %v", subscriptionName, f.namespace)
		}
		if kySub.HasCondition(activatedCondition) {
			return nil
		}

		time.Sleep(time.Duration(i) * time.Second)
	}
	return err
}

func (f *eventBusFlow) publishTestEvent() error {
	var eventSent bool
	var err error
	for i := 0; i < noOfRetries; i++ {
		if _, err = f.publish(publishEventEndpointURL); err != nil {
			time.Sleep(time.Duration(i) * time.Second)
		} else {
			eventSent = true
			break
		}
	}

	if !eventSent {
		return fmt.Errorf("cannot send test event: %v", err)
	}
	return nil
}

func (f *eventBusFlow) publish(publishEventURL string) (*publishApi.PublishResponse, error) {
	payload := fmt.Sprintf(
		`{"source-id": "%s","event-type":"%s","event-type-version":"v1","event-time":"2018-11-02T22:08:41+00:00","data":"test-event-1"}`, srcID, eventType)
	res, err := http.Post(publishEventURL, "application/json", strings.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("post request failed: %v", err)
	}

	if err := verifyStatusCode(res, 200); err != nil {
		return nil, err
	}
	respObj := &publishApi.PublishResponse{}
	body, err := ioutil.ReadAll(res.Body)
	defer func() {
		_ = res.Body.Close()
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
	for i := 0; i < noOfRetries; i++ {
		time.Sleep(time.Duration(i) * time.Second)
		res, err := http.Get(subscriberResultsEndpointURL)
		if err != nil {
			return fmt.Errorf("get request failed: %v", err)
		}

		if err := verifyStatusCode(res, 200); err != nil {
			return fmt.Errorf("get request failed: %v", err)
		}
		body, err := ioutil.ReadAll(res.Body)
		var resp string
		_ = json.Unmarshal(body, &resp)
		_ = res.Body.Close()
		if len(resp) == 0 { // no event received by subscriber
			continue
		}
		if resp != "test-event-1" {
			return fmt.Errorf("wrong response: %s, want: %s", resp, "test-event-1")
		}
		return nil
	}
	return errors.New("timeout for subscriber response")
}

func (f *eventBusFlow) cleanup() error {
	subscriberShutdownEndpointURL := "http://" + subscriberName + "." + f.namespace + ":9000/v1/shutdown"

	_, _ = http.Post(subscriberShutdownEndpointURL, "application/json", strings.NewReader(`{"shutdown": "true"}`))

	deletePolicy := metaV1.DeletePropagationForeground
	gracePeriodSeconds := int64(0)
	_ = f.k8sInterface.AppsV1().Deployments(f.namespace).Delete(subscriberName,
		&metaV1.DeleteOptions{GracePeriodSeconds: &gracePeriodSeconds, PropagationPolicy: &deletePolicy})

	_ = f.k8sInterface.CoreV1().Services(f.namespace).Delete(subscriberName,
		&metaV1.DeleteOptions{GracePeriodSeconds: &gracePeriodSeconds})

	_ = f.subsInterface.EventingV1alpha1().Subscriptions(f.namespace).Delete(subscriptionName, &metaV1.DeleteOptions{PropagationPolicy: &deletePolicy})

	_ = f.eaInterface.ApplicationconnectorV1alpha1().EventActivations(f.namespace).Delete(eventActivationName, &metaV1.DeleteOptions{PropagationPolicy: &deletePolicy})

	return nil
}

func checkStatus(statusEndpointURL string) error {
	res, err := http.Get(statusEndpointURL)
	if err != nil {
		return err
	}
	return verifyStatusCode(res, http.StatusOK)
}

func checkStatusCode(res *http.Response, expectedStatusCode int) bool {
	if res.StatusCode != expectedStatusCode {
		return false
	}
	return true
}

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
