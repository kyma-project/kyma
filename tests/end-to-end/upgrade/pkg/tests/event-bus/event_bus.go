package eventbus

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/avast/retry-go"
	api "github.com/kyma-project/kyma/components/event-bus/api/publish"
	subApis "github.com/kyma-project/kyma/components/event-bus/api/push/eventing.kyma-project.io/v1alpha1"
	eaClientSet "github.com/kyma-project/kyma/components/event-bus/generated/ea/clientset/versioned"
	subscriptionClientSet "github.com/kyma-project/kyma/components/event-bus/generated/push/clientset/versioned"
	eventBusTestUtil "github.com/kyma-project/kyma/components/event-bus/test/util"
	"github.com/sirupsen/logrus"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	// events
	eventActivationName = "test-ea"
	sourceID            = "test.local"
	eventType           = "test-e2e"
	version1            = "v1"
	version2            = "v2"
	eventData1          = "test-data-1"
	eventData2          = "test-data-2"

	// publishers
	publishEventEndpointURLV1 = "http://event-publish-service.kyma-system:8080/v1/events"
	publishEventEndpointURLV2 = "http://event-publish-service.kyma-system:8080/v2/events"
	publishStatusEndpointURL  = "http://event-publish-service.kyma-system:8080/v1/status/ready"

	// subscribers
	port               = 9000
	subscriberNameV1   = "test-subscriber-v1"
	subscriberNameV2   = "test-subscriber-v2"
	subscriptionNameV1 = "test-subscription-v1"
	subscriptionNameV2 = "test-subscription-v2"
	subscriberImage    = "eu.gcr.io/kyma-project/pr/event-bus-e2e-subscriber:PR-4893"
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
	retryOptions  []retry.Option
}

func defaultRetryOptions() *[]retry.Option {
	return &[]retry.Option{
		retry.Attempts(13), // at max (100 * (1 << 13)) / 1000 = 819,2 sec
		retry.OnRetry(func(n uint, err error) {
			fmt.Printf(".")
		}),
	}
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
	return ut.newFlow(stop, log, namespace, defaultRetryOptions()).createResources()
}

// TestResources tests resources after upgrade phase
func (ut *UpgradeTest) TestResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	return ut.newFlow(stop, log, namespace, defaultRetryOptions()).testResources()
}

func (ut *UpgradeTest) newFlow(stop <-chan struct{}, log logrus.FieldLogger, namespace string, retryOptions *[]retry.Option) *eventBusFlow {
	return &eventBusFlow{
		log:       log,
		stop:      stop,
		namespace: namespace,

		k8sInterface:  ut.K8sInterface,
		eaInterface:   ut.EaInterface,
		subsInterface: ut.SubsInterface,
		retryOptions:  *retryOptions,
	}
}

func (f *eventBusFlow) createResources() error {
	// iterate over steps
	for _, fn := range []func() error{
		// create resources
		f.createEventActivation,
		f.createSubscriptionV1,
		f.createSubscriptionV2,
		f.createSubscriberV1,
		f.createSubscriberV2,
		// check resources status
		f.checkPublisherStatus,
		f.checkSubscriberStatusV1,
		f.checkSubscriberStatusV2,
		f.checkSubscriptionReadyV1,
		f.checkSubscriptionReadyV2,
		// publish test events
		f.publishTestEventV1,
		f.publishTestEventV2,
		// check event delivery
		f.checkSubscriberReceivedEventV1,
		f.checkSubscriberReceivedEventV2,
	} {
		if err := fn(); err != nil {
			f.log.WithField("error", err).Error("createResources() failed")
			return err
		}
	}
	return nil
}

func (f *eventBusFlow) testResources() error {
	// iterate over steps
	for _, fn := range []func() error{
		// check resources status
		f.checkPublisherStatus,
		f.checkSubscriberStatusV1,
		f.checkSubscriberStatusV2,
		f.checkSubscriptionReadyV1,
		f.checkSubscriptionReadyV2,
		// publish tes events
		f.publishTestEventV1,
		f.publishTestEventV2,
		// check event delivery
		f.checkSubscriberReceivedEventV1,
		f.checkSubscriberReceivedEventV2,
		// cleanup test resources
		f.cleanup,
	} {
		if err := fn(); err != nil {
			f.log.WithField("error", err).Error("testResources() failed")
			return err
		}
	}
	return nil
}

func (f *eventBusFlow) createSubscriberV1() error {
	return f.createSubscriber(subscriberNameV1)
}

func (f *eventBusFlow) createSubscriberV2() error {
	return f.createSubscriber(subscriberNameV2)
}

func (f *eventBusFlow) createSubscriber(subscriberName string) error {
	if _, err := f.k8sInterface.AppsV1().Deployments(f.namespace).Get(subscriberName, metav1.GetOptions{}); err != nil {
		f.log.Info("create Subscriber deployment")
		if _, err := f.k8sInterface.AppsV1().Deployments(f.namespace).Create(eventBusTestUtil.NewSubscriberDeploymentWithName(subscriberName, subscriberImage)); err != nil {
			f.log.WithField("error", err).Error("create Subscriber deployment")
			return err
		}
		f.log.Info("create Subscriber service")
		if _, err := f.k8sInterface.CoreV1().Services(f.namespace).Create(eventBusTestUtil.NewSubscriberServiceWithName(subscriberName)); err != nil {
			f.log.WithField("error", err).Error("create Subscriber service failed")
		}
		err := retry.Do(func() error {
			var podReady bool
			if pods, err := f.k8sInterface.CoreV1().Pods(f.namespace).List(metav1.ListOptions{LabelSelector: "app=" + subscriberName}); err != nil {
				f.log.WithField("error", err).Error("list Pods failed")
			} else {
				for _, pod := range pods.Items {
					if podReady = isPodReady(&pod); !podReady {
						f.log.WithField("pod", pod).Info("pod not ready")
						break
					}
				}
			}
			if podReady {
				return nil
			}
			return fmt.Errorf("subscriber Pod not ready")
		}, f.retryOptions...)
		if err == nil {
			f.log.Info("subscriber created")
		}
		return err
	}
	return nil
}

func (f *eventBusFlow) createEventActivation() error {
	f.log.Info("create Event Activation")
	return retry.Do(func() error {
		_, err := f.eaInterface.ApplicationconnectorV1alpha1().EventActivations(f.namespace).Create(eventBusTestUtil.NewEventActivation(eventActivationName, f.namespace, sourceID))
		if err == nil {
			return nil
		}
		if !strings.Contains(err.Error(), "already exists") {
			return fmt.Errorf("error in creating EventActivation: %v", err)
		}
		return nil
	}, f.retryOptions...)
}

func (f *eventBusFlow) createSubscriptionV1() error {
	return f.createSubscription(subscriptionNameV1, version1)
}

func (f *eventBusFlow) createSubscriptionV2() error {
	return f.createSubscription(subscriptionNameV2, version2)
}

func (f *eventBusFlow) createSubscription(subscriptionName, version string) error {
	f.log.Info("create Subscription")
	eventsEndpoint := getSubscriberEventsURL(subscriptionName, f.namespace, version)
	_, err := f.subsInterface.EventingV1alpha1().Subscriptions(f.namespace).Create(eventBusTestUtil.NewSubscription(subscriptionName, f.namespace, eventsEndpoint, eventType, version, sourceID))
	if err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			f.log.WithField("error", err).Error("error in creating subscription")
			return err
		}
	}
	return err
}

func (f *eventBusFlow) checkSubscriberStatusV1() error {
	return f.checkSubscriberStatus(subscriberNameV1, version1)
}

func (f *eventBusFlow) checkSubscriberStatusV2() error {
	return f.checkSubscriberStatus(subscriberNameV2, version2)
}

func (f *eventBusFlow) checkSubscriberStatus(subscriberName, version string) error {
	f.log.Info("check Subscriber status")
	endpoint := getSubscriberStatusURL(subscriberName, f.namespace, version)
	return retry.Do(func() error {
		if res, err := http.Get(endpoint); err != nil {
			return fmt.Errorf("subscriber Status request failed: %v", err)
		} else if !f.checkStatusCode(res, http.StatusOK) {
			return fmt.Errorf("subscriber Server Status request returns: %v", res)
		}
		return nil
	}, f.retryOptions...)
}

func (f *eventBusFlow) checkPublisherStatus() error {
	f.log.Info("check Publisher status")
	return retry.Do(func() error {
		if err := checkStatus(publishStatusEndpointURL); err != nil {
			return fmt.Errorf("publisher not ready: %v", err)
		}
		return nil
	}, f.retryOptions...)
}

func (f *eventBusFlow) checkSubscriptionReadyV1() error {
	return f.checkSubscriptionReady(subscriptionNameV1)
}

func (f *eventBusFlow) checkSubscriptionReadyV2() error {
	return f.checkSubscriptionReady(subscriptionNameV2)
}

func (f *eventBusFlow) checkSubscriptionReady(subscriptionName string) error {
	f.log.Info("check Subscription ready status")
	activatedCondition := subApis.SubscriptionCondition{Type: subApis.Ready, Status: subApis.ConditionTrue}
	return retry.Do(func() error {
		kySub, err := f.subsInterface.EventingV1alpha1().Subscriptions(f.namespace).Get(subscriptionName, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("cannot get Kyma subscription: %v; name: %v; namespace: %v", err, subscriptionName, f.namespace)
		}
		if kySub.HasCondition(activatedCondition) {
			return nil
		}
		return fmt.Errorf("kyma subscription %+v does not have condition: %v", kySub, activatedCondition)
	}, f.retryOptions...)
}

func (f *eventBusFlow) publishTestEventV1() error {
	return f.publishTestEvent(version1, eventData1, publishEventEndpointURLV1)
}

func (f *eventBusFlow) publishTestEventV2() error {
	return f.publishTestEvent(version2, eventData2, publishEventEndpointURLV2)
}

func (f *eventBusFlow) publishTestEvent(version, eventData, publishUrl string) error {
	f.log.WithField("value", eventData).Debug("publish data")

	return retry.Do(func() error {
		if _, err := f.publish(version, eventData, publishUrl); err != nil {
			return fmt.Errorf("publish event failed: %v", err)
		}
		return nil
	}, f.retryOptions...)
}

func (f *eventBusFlow) publish(eventTypeVersion, eventData, publishEventURL string) (*api.PublishResponse, error) {
	payload := fmt.Sprintf(
		`{"source-id": "%s","event-type":"%s","event-type-version":"%s","event-time":"2018-11-02T22:08:41+00:00","data":"%s"}`,
		sourceID, eventType, eventTypeVersion, eventData)
	f.log.WithField("event", payload).Info("event to be published")
	res, err := http.Post(publishEventURL, "application/json", strings.NewReader(payload))
	if err != nil {
		f.log.WithField("error", err).Error("post request failed")
		return nil, err
	}
	f.dumpResponse(res)
	if err := verifyStatusCode(res, 200); err != nil {
		return nil, err
	}
	respObj := &api.PublishResponse{}
	var body []byte
	if body, err = ioutil.ReadAll(res.Body); err != nil {
		f.log.WithField("error", err).Error("unmarshal error")
		return nil, err
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			f.log.Error(err)
		}
	}()
	err = json.Unmarshal(body, &respObj)
	f.log.WithField("response", string(body)).Info("publish response object")
	if len(respObj.EventID) == 0 {
		return nil, fmt.Errorf("empty respObj.EventID")
	}
	return respObj, err
}

func (f *eventBusFlow) checkSubscriberReceivedEventV1() error {
	return f.checkSubscriberReceivedEvent(subscriberNameV1, version1, eventData1)
}

func (f *eventBusFlow) checkSubscriberReceivedEventV2() error {
	return f.checkSubscriberReceivedEvent(subscriberNameV2, version2, eventData2)
}

func (f *eventBusFlow) checkSubscriberReceivedEvent(subscriberName, version, expectedData string) error {
	endpoint := getSubscriberResultsURL(subscriberName, f.namespace, version)
	return retry.Do(func() error {
		res, err := http.Get(endpoint)
		if err != nil {
			return fmt.Errorf("get request failed: %v", err)
		}
		f.dumpResponse(res)
		if err := verifyStatusCode(res, 200); err != nil {
			return fmt.Errorf("get request failed: %v", err)
		}
		var body []byte
		if body, err = ioutil.ReadAll(res.Body); err != nil {
			return err
		}
		var resp string
		if err := json.Unmarshal(body, &resp); err != nil {
			return err
		}
		f.log.WithField("response", resp).Info("subscriber response")
		defer func() {
			if err := res.Body.Close(); err != nil {
				f.log.Error(err)
			}
		}()
		if len(resp) == 0 { // no event received by subscriber
			return fmt.Errorf("no event received by subscriber: %v", resp)
		}
		f.log.WithFields(logrus.Fields{
			"expected": expectedData,
			"actual":   resp,
		}).Debug("subscriber response")
		if resp != expectedData {
			return fmt.Errorf("wrong response: %s, want: %s", resp, expectedData)
		}
		return nil
	}, f.retryOptions...)
}

func (f *eventBusFlow) cleanup() error {
	// delete policy
	deletePolicy := metav1.DeletePropagationForeground

	// cleanup subscribers
	f.cleanupSubscriber(subscriberNameV1, version1, deletePolicy)
	f.cleanupSubscriber(subscriberNameV2, version2, deletePolicy)

	// cleanup subscriptions
	f.cleanupSubscription(subscriptionNameV1, deletePolicy)
	f.cleanupSubscription(subscriptionNameV2, deletePolicy)

	// cleanup event activation
	f.log.WithField("event_activation", eventActivationName).Info("delete test EventActivation")
	if err := f.eaInterface.ApplicationconnectorV1alpha1().EventActivations(f.namespace).Delete(eventActivationName, &metav1.DeleteOptions{PropagationPolicy: &deletePolicy}); err != nil {
		f.log.WithField("error", err).Warn("delete EventActivation failed")
	}

	return nil
}

func (f *eventBusFlow) cleanupSubscriber(subscriberName, version string, deletePolicy metav1.DeletionPropagation) {
	f.log.Info("send shutdown request to Subscriber")
	subscriberShutdownEndpointURL := getSubscriberShutdownURL(subscriberName, f.namespace, version)
	if _, err := http.Post(subscriberShutdownEndpointURL, "application/json", strings.NewReader(`{"shutdown": "true"}`)); err != nil {
		f.log.WithField("error", err).Warn("shutdown Subscriber failed")
	}

	f.log.Info("delete Subscriber deployment")
	gracePeriodSeconds := int64(0)
	if err := f.k8sInterface.AppsV1().Deployments(f.namespace).Delete(subscriberName,
		&metav1.DeleteOptions{GracePeriodSeconds: &gracePeriodSeconds, PropagationPolicy: &deletePolicy}); err != nil {
		f.log.WithField("error", err).Warn("delete Subscriber Deployment failed")
	}

	f.log.Info("delete Subscriber service")
	if err := f.k8sInterface.CoreV1().Services(f.namespace).Delete(subscriberName,
		&metav1.DeleteOptions{GracePeriodSeconds: &gracePeriodSeconds}); err != nil {
		f.log.WithField("error", err).Warn("delete Subscriber Service failed")
	}
}

func (f *eventBusFlow) cleanupSubscription(subscriptionName string, deletePolicy metav1.DeletionPropagation) {
	f.log.WithField("subscription", subscriptionName).Info("delete test Subscription")
	if err := f.subsInterface.EventingV1alpha1().Subscriptions(f.namespace).Delete(subscriptionName, &metav1.DeleteOptions{PropagationPolicy: &deletePolicy}); err != nil {
		f.log.WithField("error", err).Warn("delete Subscription failed", err)
	}
}

func getSubscriberEventsURL(name, namespace, version string) string {
	return fmt.Sprintf("http://%s.%s:%d/%s/events", name, namespace, port, version)
}

func getSubscriberResultsURL(name, namespace, version string) string {
	return fmt.Sprintf("http://%s.%s:%d/%s/results", name, namespace, port, version)
}

func getSubscriberStatusURL(name, namespace, version string) string {
	return fmt.Sprintf("http://%s.%s:%d/%s/status", name, namespace, port, version)
}

func getSubscriberShutdownURL(name, namespace, version string) string {
	return fmt.Sprintf("http://%s.%s:%d/%s/shutdown", name, namespace, port, version)
}

func checkStatus(statusEndpointURL string) error {
	res, err := http.Get(statusEndpointURL)
	if err != nil {
		return err
	}
	return verifyStatusCode(res, http.StatusOK)
}

func (f *eventBusFlow) dumpResponse(resp *http.Response) {
	defer func() {
		if err := resp.Body.Close(); err != nil {
			f.log.Error(err)
		}
	}()
	dump, err := httputil.DumpResponse(resp, true)
	if err != nil {
		f.log.Error(err)
	}
	f.log.WithField("value", dump).Debug("dump response")
}

func (f *eventBusFlow) checkStatusCode(res *http.Response, expectedStatusCode int) bool {
	if res.StatusCode != expectedStatusCode {
		f.log.WithFields(logrus.Fields{
			"actual":   res.StatusCode,
			"expected": expectedStatusCode,
		}).Warn("status code is wrong")
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
