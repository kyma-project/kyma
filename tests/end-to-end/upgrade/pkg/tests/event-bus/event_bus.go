package eventbus

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/kyma-project/kyma/components/event-bus/api/publish"
	subApis "github.com/kyma-project/kyma/components/event-bus/api/push/eventing.kyma-project.io/v1alpha1"
	eaClientSet "github.com/kyma-project/kyma/components/event-bus/generated/ea/clientset/versioned"
	subscriptionClientSet "github.com/kyma-project/kyma/components/event-bus/generated/push/clientset/versioned"
	"github.com/kyma-project/kyma/components/event-bus/test/util"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
)

const (
	eventType           = "test-e2e"
	subscriptionName    = "test-sub"
	eventActivationName = "test-ea"
	srcID               = "test.local"

	success     = 0
	fail        = 1
	noOfRetries = 20

	subscriberName           = "test-core-event-bus-subscriber"
	subscriberImage          = "eu.gcr.io/kyma-project/event-bus-e2e-subscriber:0.9.0"
	publishEventEndpointURL  = "http://event-bus-publish:8080/v1/events"
	publishStatusEndpointURL = "http://event-bus-publish:8080/v1/status/ready"
)

// UpgradeTest tests the Event Bus business logic after Kyma upgrade phase
type UpgradeTest struct {
	K8sInterface  kubernetes.Interface
	EaInterface   eaClientSet.Interface
	SubsInterface subscriptionClientSet.Interface
}

type eventBusFlow struct {
	namespace string
	log       logrus.FieldLogger
	stop      <-chan struct{}

	k8sInterface  kubernetes.Interface
	eaInterface   eaClientSet.Interface
	subsInterface subscriptionClientSet.Interface
}

// NewEventBusUpgradeTest returns new instance of the UpgradeTest
func NewEventBusUpgradeTest(k8sCli kubernetes.Interface, eaInterface eaClientSet.Interface, subsCli subscriptionClientSet.Interface) *UpgradeTest {
	return &UpgradeTest{
		K8sInterface:  k8sCli,
		EaInterface:   eaInterface,
		SubsInterface: subsCli,
	}
}

// CreateResources creates resources needed for e2e upgrade test
func (ut *UpgradeTest) CreateResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	return ut.newFlow(stop, log, namespace).createResources()
}

// TestResources tests resources after upgrade phase
func (ut *UpgradeTest) TestResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	return ut.newFlow(stop, log, namespace).testResources()
}

func (ut *UpgradeTest) newFlow(stop <-chan struct{}, log logrus.FieldLogger, namespace string) *eventBusFlow {
	return &eventBusFlow{
		log:       log,
		stop:      stop,
		namespace: namespace,

		k8sInterface:  ut.K8sInterface,
		eaInterface:   ut.EaInterface,
		subsInterface: ut.SubsInterface,
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
		err := fn()
		if err != nil {
			log.Printf("CreateResources() failed with: %v", err)
			return err
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
		f.cleanup,
	} {
		err := fn()
		if err != nil {
			log.Printf("TestResources() failed with: %v", err)
			return err
		}
	}
	return nil
}

func (f *eventBusFlow) createSubscriber() error {
	if _, err := f.k8sInterface.AppsV1().Deployments(f.namespace).Get(subscriberName, metav1.GetOptions{}); err != nil {
		log.Println("Create Subscriber deployment")
		if _, err := f.k8sInterface.AppsV1().Deployments(f.namespace).Create(util.NewSubscriberDeployment(subscriberImage)); err != nil {
			log.Printf("Create Subscriber deployment: %v\n", err)
			return err
		}
		log.Println("Create Subscriber service")
		if _, err := f.k8sInterface.CoreV1().Services(f.namespace).Create(util.NewSubscriberService()); err != nil {
			log.Printf("Create Subscriber service failed: %v\n", err)
		}
		time.Sleep(30 * time.Second)

		for i := 0; i < 60; i++ {
			var podReady bool
			if pods, err := f.k8sInterface.CoreV1().Pods(f.namespace).List(metav1.ListOptions{LabelSelector: "app=" + subscriberName}); err != nil {
				log.Printf("List Pods failed: %v\n", err)
			} else {
				for _, pod := range pods.Items {
					if podReady = isPodReady(&pod); !podReady {
						log.Printf("Pod not ready: %+v\n;", pod)
						break
					}
				}
			}
			if podReady {
				break
			} else {
				log.Printf("Subscriber Pod not ready, retrying (%d/%d)", i, 60)
				time.Sleep(1 * time.Second)
			}
		}
		log.Println("Subscriber created")
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
			log.Printf("Error in creating event activation - %v; Retrying (%d/%d)\n", err, i, noOfRetries)
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
			log.Printf("Error in creating subscription: %v\n", err)
			return err
		}
	}
	return err
}

func (f *eventBusFlow) checkSubscriberStatus() error {
	subscriberStatusEndpointURL := "http://" + subscriberName + "." + f.namespace + ":9000/v1/status"
	var err error
	for i := 0; i < noOfRetries; i++ {
		if res, err := http.Get(subscriberStatusEndpointURL); err != nil {
			log.Printf("Subscriber Status request failed: %v; Retrying (%d/%d)", err, i, noOfRetries)
			time.Sleep(time.Duration(i) * time.Second)
		} else if !checkStatusCode(res, http.StatusOK) {
			log.Printf("Subscriber Server Status request returns: %v; Retrying (%d/%d)\n", res, i, noOfRetries)
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
			log.Printf("Publisher not ready: %v", err)
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
		kySub, err := f.subsInterface.EventingV1alpha1().Subscriptions(f.namespace).Get(subscriptionName, metav1.GetOptions{})
		if err != nil {
			log.Printf("Cannot get Kyma subscription, name: %v; namespace: %v", subscriptionName, f.namespace)
			return err
		}
		if kySub.HasCondition(activatedCondition) {
			return nil
		}

		time.Sleep(1 * time.Second)
	}
	return err
}

func (f *eventBusFlow) publishTestEvent() error {
	var eventSent bool
	var err error
	for i := 0; i < noOfRetries; i++ {
		if _, err = publish(publishEventEndpointURL); err != nil {
			log.Printf("Publish event failed: %v; Retrying (%d/%d)", err, i, noOfRetries)
			time.Sleep(time.Duration(i) * time.Second)
		} else {
			eventSent = true
			break
		}
	}

	if !eventSent {
		log.Println("Error: Cannot send test event")
		return err
	}
	return nil
}

func publish(publishEventURL string) (*api.PublishResponse, error) {
	payload := fmt.Sprintf(
		`{"source-id": "%s","event-type":"%s","event-type-version":"v1","event-time":"2018-11-02T22:08:41+00:00","data":"test-event-1"}`, srcID, eventType)
	log.Printf("event to be published: %v\n", payload)
	res, err := http.Post(publishEventURL, "application/json", strings.NewReader(payload))
	if err != nil {
		log.Printf("Post request failed: %v\n", err)
		return nil, err
	}
	dumpResponse(res)
	if err := verifyStatusCode(res, 200); err != nil {
		return nil, err
	}
	respObj := &api.PublishResponse{}
	body, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	err = json.Unmarshal(body, &respObj)
	if err != nil {
		log.Printf("Unmarshal error: %v", err)
		return nil, err
	}
	log.Printf("Publish response object: %+v", respObj)
	if len(respObj.EventID) == 0 {
		return nil, fmt.Errorf("empty respObj.EventID")
	}
	return respObj, err
}

func (f *eventBusFlow) checkSubscriberReceivedEvent() error {
	subscriberResultsEndpointURL := "http://" + subscriberName + "." + f.namespace + ":9000/v1/results"
	for i := 0; i < noOfRetries; i++ {
		time.Sleep(time.Duration(i) * time.Second)
		log.Printf("Get subscriber response (%d/%d)\n", i, noOfRetries)
		res, err := http.Get(subscriberResultsEndpointURL)
		if err != nil {
			log.Printf("Get request failed: %v\n", err)
			return err
		}
		dumpResponse(res)
		if err := verifyStatusCode(res, 200); err != nil {
			log.Printf("Get request failed: %v", err)
			return err
		}
		body, err := ioutil.ReadAll(res.Body)
		var resp string
		json.Unmarshal(body, &resp)
		log.Printf("Subscriber response: %s\n", resp)
		res.Body.Close()
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
	log.Println("Send shutdown request to Subscriber")
	if _, err := http.Post(subscriberShutdownEndpointURL, "application/json", strings.NewReader(`{"shutdown": "true"}`)); err != nil {
		log.Printf("Warning: Shutdown Subscriber failed: %v", err)
	}
	log.Println("Delete Subscriber deployment")
	deletePolicy := metav1.DeletePropagationForeground
	gracePeriodSeconds := int64(0)
	if err := f.k8sInterface.AppsV1().Deployments(f.namespace).Delete(subscriberName,
		&metav1.DeleteOptions{GracePeriodSeconds: &gracePeriodSeconds, PropagationPolicy: &deletePolicy}); err != nil {
		log.Printf("Warning: Delete Subscriber Deployment falied: %v", err)
	}
	log.Println("Delete Subscriber service")
	if err := f.k8sInterface.CoreV1().Services(f.namespace).Delete(subscriberName,
		&metav1.DeleteOptions{GracePeriodSeconds: &gracePeriodSeconds}); err != nil {
		log.Printf("Warning: Delete Subscriber Service falied: %v", err)
	}

	log.Printf("Delete test subscription: %v\n", subscriptionName)
	if err := f.subsInterface.EventingV1alpha1().Subscriptions(f.namespace).Delete(subscriptionName, &metav1.DeleteOptions{PropagationPolicy: &deletePolicy}); err != nil {
		log.Printf("Warning: Delete Subscription falied: %v", err)
	}

	log.Printf("Delete test event activation: %v\n", eventActivationName)
	if err := f.eaInterface.ApplicationconnectorV1alpha1().EventActivations(f.namespace).Delete(eventActivationName, &metav1.DeleteOptions{PropagationPolicy: &deletePolicy}); err != nil {
		log.Printf("Warning: Delete Event Activation falied: %v", err)
	}

	return nil
}

func checkStatus(statusEndpointURL string) error {
	res, err := http.Get(statusEndpointURL)
	if err != nil {
		return err
	}
	return verifyStatusCode(res, http.StatusOK)
}

func dumpResponse(resp *http.Response) {
	defer resp.Body.Close()
	dump, err := httputil.DumpResponse(resp, true)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%q", dump)
}

func checkStatusCode(res *http.Response, expectedStatusCode int) bool {
	if res.StatusCode != expectedStatusCode {
		log.Printf("Status code is wrong, have: %d, want: %d\n", res.StatusCode, expectedStatusCode)
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

func isPodReady(pod *apiv1.Pod) bool {
	for _, cs := range pod.Status.ContainerStatuses {
		if !cs.Ready {
			return false
		}
	}
	return true
}
