package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/rest"

	"github.com/avast/retry-go"
	api "github.com/kyma-project/kyma/components/event-bus/api/publish"
	subApis "github.com/kyma-project/kyma/components/event-bus/api/push/eventing.kyma-project.io/v1alpha1"
	eaClientSet "github.com/kyma-project/kyma/components/event-bus/generated/ea/clientset/versioned"
	subscriptionClientSet "github.com/kyma-project/kyma/components/event-bus/generated/push/clientset/versioned"
	"github.com/kyma-project/kyma/components/event-bus/test/util"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	port             = 9000
	eventType        = "test-e2e"
	eventTypeVersion = "v1"

	eventActivationName = "test-ea"
	srcID               = "test.local"

	success = 0
	fail    = 1

	idHeader               = "ce-id"
	timeHeader             = "ce-time"
	contentTypeHeader      = "content-type"
	sourceHeader           = "ce-source"
	eventTypeHeader        = "ce-type"
	eventTypeVersionHeader = "ce-eventtypeversion"
	customHeader           = "ce-xcustomheader"

	ceSourceIDHeaderValue         = "override-source-ID"
	ceEventTypeHeaderValue        = "override-event-type"
	contentTypeHeaderValue        = "application/json"
	ceEventTypeVersionHeaderValue = "override-event-type-version"
	customHeaderValue             = "Ce-X-custom-header-value"

	// subscribers
	subscriptionNameV1        = "test-sub-v1"
	headersSubscriptionNameV1 = "test-sub-with-headers-v1"
)

// Not exported struct, encapsulates the E2E test resources
type e2eTester struct {
	publisher1            publisher
	publisher2            publisher
	subscriber1           subscriber
	subscriber2           subscriber
	retryOts              []retry.Option
	k8sClient             *kubernetes.Clientset
	eventActivationClient *eaClientSet.Clientset
	subscriptionClient    *subscriptionClientSet.Clientset
}

// Not exported struct, encapsulates publisher details
type publisher struct {
	publishEventEndpointURLV1  string
	publishStatusEndpointURLV1 string
}

// Not exported struct, encapsulates subscriber resource parameters
type subscriber struct {
	image       string
	namespace   string
	eventsURL   string
	statusURL   string
	resultsURL  string
	shutdownURL string
}

type options struct {
	publishEventURLV1  string
	publishStatusURLV1 string
	image              string
	namespace          string
	logLevel           string
}

func init() {
	// configure logger with text instead of json for easier reading in CI logs
	log.SetFormatter(&log.TextFormatter{})

	// show file and line number
	log.SetReportCaller(true)
}

func main() {
	// init cli options
	opts := getDefaultOptions()
	opts.parseOrDie()

	// set log level
	setLogLevel(opts.logLevel)

	// run the E2E test scenarios
	tester := newE2ETester(opts, defaultRetryOptions())
	tester.initOrDie()
	tester.prepareTestResourcesOrDie()
	tester.checkTestResourcesReadyOrDie()
	tester.publishEventsOrDie()
	tester.checkEventsDeliveryOrDie()

	// test finished
	log.Info("test finished successfully")
	tester.shutdown(success, &tester.subscriber1)
}

// Init the E2E Tester or exit in case of errors
func (e *e2eTester) initOrDie() {
	log.Info("init cluster config")
	config, err := rest.InClusterConfig()
	if err != nil {
		log.WithField("error", err).Error("cannot init cluster config")
		e.shutdown(fail, &e.subscriber1)
	}

	log.Info("init k8s client")
	e.k8sClient, err = kubernetes.NewForConfig(config)
	if err != nil {
		log.WithField("error", err).Error("cannot init k8s ClientSet")
		e.shutdown(fail, &e.subscriber1)
	}

	log.Info("init EventActivation client")
	e.eventActivationClient, err = eaClientSet.NewForConfig(config)
	if err != nil {
		log.WithField("error", err).Error("cannot init EventActivation client")
		e.shutdown(fail, &e.subscriber1)
	}

	log.Info("init Subscription client")
	e.subscriptionClient, err = subscriptionClientSet.NewForConfig(config)
	if err != nil {
		log.WithField("error", err).Error("cannot init Subscription client")
		e.shutdown(fail, &e.subscriber1)
	}
}

func (e *e2eTester) prepareTestResourcesOrDie() {
	log.Info("create test namespace")
	err := e.createNamespace(e.subscriber1.namespace)
	if err != nil {
		log.WithField("error", err).Error("cannot create test namespace")
		e.shutdown(fail, &e.subscriber1)
	}

	log.Info("create EventActivation")
	if err := e.createEventActivation(e.subscriber1.namespace); err != nil {
		log.WithField("error", err).Error("cannot create EventActivation")
		e.shutdown(fail, &e.subscriber1)
	}

	log.Info("create Kyma Subscription-1")
	if err := e.createSubscription(e.subscriber1.namespace, subscriptionNameV1, e.subscriber1.eventsURL); err != nil {
		log.WithField("error", err).Error("cannot create Kyma subscription-1")
		e.shutdown(fail, &e.subscriber1)
	}

	log.Info("create a Kyma subscription-3")
	if err := e.createSubscription(e.subscriber1.namespace, headersSubscriptionNameV1, e.subscriber1.eventsURL); err != nil {
		log.WithField("error", err).Error("cannot create Kyma subscription-3")
		e.shutdown(fail, &e.subscriber1)
	}

	log.Info("create Subscriber")
	if err := e.createSubscriber(util.SubscriberName, e.subscriber1.namespace, e.subscriber1.image); err != nil {
		log.WithField("error", err).Error("cannot create Subscriber")
		e.shutdown(fail, &e.subscriber1)
	}
}

func (e *e2eTester) checkTestResourcesReadyOrDie() {
	log.Info("check Subscriber's v1 endpoint Status")
	if err := e.checkSubscriberV1EndpointStatus(); err != nil {
		log.WithField("error", err).Error("cannot connect to Subscriber v1 endpoint")
		e.shutdown(fail, &e.subscriber1)
	}

	log.Info("check Subscriber's v3 endpoint Status")
	if err := e.checkSubscriberV3EndpointStatus(); err != nil {
		log.WithField("error", err).Info("cannot connect to Subscriber v3 endpoint")
		e.shutdown(fail, &e.subscriber1)
	}

	log.Info("check Publisher Status")
	if err := e.checkPublisherStatus(); err != nil {
		log.WithField("error", err).Error("cannot connect to Publisher")
		e.shutdown(fail, &e.subscriber1)
	}

	log.Info("check Kyma subscription ready Status")
	if err := e.checkSubscriptionReady(subscriptionNameV1); err != nil {
		log.WithField("error", err).Error("kyma Subscription not ready")
		e.shutdown(fail, &e.subscriber1)
	}

	log.Info("check Kyma headers subscription ready Status")
	if err := e.checkSubscriptionReady(headersSubscriptionNameV1); err != nil {
		log.WithField("error", err).Error("kyma Subscription not ready")
		e.shutdown(fail, &e.subscriber1)
	}
}

func (e *e2eTester) publishEventsOrDie() {
	log.Info("publish an event")
	err := retry.Do(func() error {
		_, err := e.publishTestEvent(e.publisher1.publishEventEndpointURLV1)
		return err
	}, e.retryOts...)
	if err != nil {
		log.WithField("error", err).Error("cannot publish event failed")
		e.shutdown(fail, &e.subscriber1)
	}

	log.Info("publish event with headers")
	err = retry.Do(func() error {
		_, err := e.publishHeadersTestEvent(e.publisher1.publishEventEndpointURLV1)
		return err
	}, e.retryOts...)
	if err != nil {
		log.WithField("error", err).Error("cannot publish event with headers")
		e.shutdown(fail, &e.subscriber1)
	}
}

func (e *e2eTester) checkEventsDeliveryOrDie() {
	log.Info("try to read the response from subscriber1 server")
	if err := e.checkReceivedEvent(); err != nil {
		log.WithField("error", err).Error("cannot get the test event from subscriber1")
		e.shutdown(fail, &e.subscriber1)
	}

	log.Info("try to read the response from v3 endpoint of the subscriber1")
	if err := e.checkReceivedEventHeaders(); err != nil {
		log.WithField("error", err).Error("cannot get the test event from subscriber1 v3 endpoint")
		e.shutdown(fail, &e.subscriber1)
	}
}

func (e *e2eTester) shutdown(code int, subscriber *subscriber) {
	log.Info("send shutdown request to Subscriber")
	if _, err := http.Post(subscriber.shutdownURL, "application/json", strings.NewReader(`{"shutdown": "true"}`)); err != nil {
		log.WithField("error", err).Warning("shutdown Subscriber failed")
	}
	log.Info("delete Subscriber deployment")
	deletePolicy := metav1.DeletePropagationForeground
	gracePeriodSeconds := int64(0)

	if err := e.k8sClient.AppsV1().Deployments(subscriber.namespace).Delete(util.SubscriberName,
		&metav1.DeleteOptions{GracePeriodSeconds: &gracePeriodSeconds, PropagationPolicy: &deletePolicy}); err != nil {
		log.WithField("error", err).Warn("delete Subscriber Deployment failed")
	}
	log.Info("delete Subscriber service")
	if err := e.k8sClient.CoreV1().Services(subscriber.namespace).Delete(util.SubscriberName,
		&metav1.DeleteOptions{GracePeriodSeconds: &gracePeriodSeconds}); err != nil {
		log.WithField("error", err).Warn("delete Subscriber Service failed")
	}
	if e.subscriptionClient != nil {
		log.WithField("subscription", subscriptionNameV1).Info("delete test subscription")
		if err := e.subscriptionClient.EventingV1alpha1().Subscriptions(subscriber.namespace).Delete(subscriptionNameV1,
			&metav1.DeleteOptions{PropagationPolicy: &deletePolicy}); err != nil {
			log.WithField("error", err).Warn("delete Subscription failed")
		}
	}
	if e.subscriptionClient != nil {
		log.WithField("subscription", subscriptionNameV1).Info("delete headers test subscription")
		if err := e.subscriptionClient.EventingV1alpha1().Subscriptions(subscriber.namespace).Delete(headersSubscriptionNameV1,
			&metav1.DeleteOptions{PropagationPolicy: &deletePolicy}); err != nil {
			log.WithField("error", err).Warn("delete Subscription failed")
		}
	}
	if e.eventActivationClient != nil {
		log.WithField("event_activation", eventActivationName).Info("delete test event activation")
		if err := e.eventActivationClient.ApplicationconnectorV1alpha1().EventActivations(subscriber.namespace).Delete(eventActivationName, &metav1.DeleteOptions{PropagationPolicy: &deletePolicy}); err != nil {
			log.WithField("error", err).Warn("delete Event Activation failed")
		}
	}

	log.WithField("namespace", subscriber.namespace).Info("delete test namespace")
	if err := e.k8sClient.CoreV1().Namespaces().Delete(subscriber.namespace, &metav1.DeleteOptions{PropagationPolicy: &deletePolicy}); err != nil {
		log.WithField("error", err).Warn("delete Namespace failed")
	}
	os.Exit(code)
}

func (e *e2eTester) publishTestEvent(publishEventURL string) (*api.Response, error) {
	payload := fmt.Sprintf(
		`{"source-id": "%s","event-type":"%s","event-type-version":"%s","event-time":"2018-11-02T22:08:41+00:00","data":"test-event-1"}`, srcID, eventType, eventTypeVersion)
	log.WithField("event", payload).Info("event to be published")
	res, err := http.Post(publishEventURL, "application/json", strings.NewReader(payload))
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			log.Warn(err)
		}
	}()
	if _, err := httputil.DumpResponse(res, true); err != nil {
		return nil, err
	}
	if err := verifyStatusCode(res, 200); err != nil {
		return nil, err
	}
	respObj := &api.Response{}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(body, &respObj)
	if err != nil {
		return nil, err
	}
	log.WithField("response", string(body)).Info("publish response object")
	if len(respObj.EventID) == 0 {
		return nil, fmt.Errorf("empty respObj.EventID")
	}
	return respObj, err
}

func (e *e2eTester) publishHeadersTestEvent(publishEventURL string) (*api.Response, error) {
	payload := fmt.Sprintf(
		`{"source-id": "%s","event-type":"%s","event-type-version":"%s","event-time":"2018-11-02T22:08:41+00:00","data":"headers-test-event"}`, srcID, eventType, eventTypeVersion)
	log.WithField("event", payload).Info("event to be published")

	client := &http.Client{}
	req, err := http.NewRequest("POST", publishEventURL, strings.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Add(sourceHeader, ceSourceIDHeaderValue)
	req.Header.Add(eventTypeHeader, ceEventTypeHeaderValue)
	req.Header.Add(eventTypeVersionHeader, ceEventTypeVersionHeaderValue)
	req.Header.Add(customHeader, customHeaderValue)
	res, err := client.Do(req)
	if err != nil {
		log.WithField("error", err).Error("post request failed")
		return nil, err
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			log.Warn(err)
		}
	}()
	if _, err := httputil.DumpResponse(res, true); err != nil {
		return nil, err
	}
	if err := verifyStatusCode(res, 200); err != nil {
		return nil, err
	}
	respObj := &api.Response{}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(body, &respObj)
	if err != nil {
		return nil, err
	}
	log.WithField("response", string(body)).Info("publish response object")
	if len(respObj.EventID) == 0 {
		return nil, fmt.Errorf("empty respObj.EventID")
	}
	return respObj, err
}

func (e *e2eTester) checkReceivedEvent() error {
	return retry.Do(func() error {
		res, err := http.Get(e.subscriber1.resultsURL)
		if err != nil {
			return err
		}
		defer func() {
			if err := res.Body.Close(); err != nil {
				log.Warn(err)
			}
		}()
		if _, err := httputil.DumpResponse(res, true); err != nil {
			return err
		}
		if err := verifyStatusCode(res, 200); err != nil {
			return err
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
		log.WithField("response", resp).Info("subscriber response")
		if len(resp) == 0 {
			return errors.New("no event received by subscriber")
		}
		if resp != "test-event-1" {
			return fmt.Errorf("wrong response: %s, want: %s", resp, "test-event-1")
		}
		return nil
	}, e.retryOts...)
}

func (e *e2eTester) checkReceivedEventHeaders() error {
	return retry.Do(func() error {
		res, err := http.Get(e.subscriber1.resultsURL)
		if err != nil {
			return err
		}
		defer func() {
			if err := res.Body.Close(); err != nil {
				log.Warn(err)
			}
		}()
		if _, err := httputil.DumpResponse(res, true); err != nil {
			return err
		}
		if err := verifyStatusCode(res, 200); err != nil {
			return err
		}
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}
		var resp map[string][]string
		if err := json.Unmarshal(body, &resp); err != nil {
			return err
		}
		log.WithField("response", resp).Info("response for v3 endpoint of the subscriber")
		if len(resp) == 0 {
			return errors.New("no event received by subscriber")
		}

		lowerResponseHeaders := make(map[string][]string)
		for k := range resp {
			lowerResponseHeaders[strings.ToLower(k)] = resp[k]
		}

		var testDataSets = []struct {
			headerKey           string
			headerExpectedValue string
		}{
			{headerKey: sourceHeader, headerExpectedValue: srcID},
			{headerKey: eventTypeHeader, headerExpectedValue: eventType},
			{headerKey: eventTypeVersionHeader, headerExpectedValue: eventTypeVersion},
			{headerKey: contentTypeHeader, headerExpectedValue: contentTypeHeaderValue},
			{headerKey: customHeader, headerExpectedValue: customHeaderValue},
		}

		for _, testData := range testDataSets {
			if _, ok := lowerResponseHeaders[testData.headerKey]; !ok {
				return fmt.Errorf("map %v does not contain key %v", lowerResponseHeaders, testData.headerKey)
			}
			if lowerResponseHeaders[testData.headerKey][0] != testData.headerExpectedValue {
				return fmt.Errorf("wrong response: %s, want: %s", lowerResponseHeaders[testData.headerKey][0], testData.headerExpectedValue)
			}
		}

		if lowerResponseHeaders[idHeader][0] == "" {
			return fmt.Errorf("wrong response: %s, can't be empty", lowerResponseHeaders[idHeader][0])
		}
		if lowerResponseHeaders[timeHeader][0] == "" {
			return fmt.Errorf("wrong response: %s, can't be empty", lowerResponseHeaders[timeHeader][0])
		}
		return nil
	}, e.retryOts...)
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

func (e *e2eTester) createSubscriber(subscriberName string, subscriberNamespace string, subscriberImage string) error {
	if _, err := e.k8sClient.AppsV1().Deployments(subscriberNamespace).Get(subscriberName, metav1.GetOptions{}); err != nil {
		log.Info("create Subscriber deployment")
		if _, err := e.k8sClient.AppsV1().Deployments(subscriberNamespace).Create(util.NewSubscriberDeploymentWithName(subscriberName, subscriberImage)); err != nil {
			log.WithField("error", err).Error("create Subscriber deployment failed")
			return err
		}
		log.Info("create Subscriber service")
		if _, err := e.k8sClient.CoreV1().Services(subscriberNamespace).Create(util.NewSubscriberServiceWithName(subscriberName)); err != nil {
			log.WithField("error", err).Error("create Subscriber service failed")
			return err
		}

		// wait until pod is ready
		return retry.Do(func() error {
			var podReady bool
			var podNotReady error
			pods, err := e.k8sClient.CoreV1().Pods(subscriberNamespace).List(metav1.ListOptions{LabelSelector: "app=" + util.SubscriberName})
			if err != nil {
				return err
			}
			// check if pod is ready
			for _, pod := range pods.Items {
				if podReady = isPodReady(&pod); !podReady {
					podNotReady = fmt.Errorf("subscriber pod not ready: %+v", pod)
					break
				}
			}
			if !podReady {
				return podNotReady
			}
			log.Info("subscriber created")
			return nil
		}, e.retryOts...)
	}
	return nil
}

// Create EventActivation and wait for successful creation
func (e *e2eTester) createEventActivation(subscriberNamespace string) error {
	return retry.Do(func() error {
		_, err := e.eventActivationClient.ApplicationconnectorV1alpha1().EventActivations(subscriberNamespace).Create(util.NewEventActivation(eventActivationName, subscriberNamespace, srcID))
		if err == nil {
			return nil
		}
		if !strings.Contains(err.Error(), "already exists") {
			return err
		}
		return nil
	})
}

func (e *e2eTester) createNamespace(name string) error {

	ns := &apiv1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"test": "test-event-bus",
			},
		},
	}
	err := retry.Do(func() error {
		_, err := e.k8sClient.CoreV1().Namespaces().Create(ns)
		return err
	}, e.retryOts...)

	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return fmt.Errorf("namespace: %s could not be created: %v", name, err)
	}

	err = retry.Do(func() error {
		_, err = e.k8sClient.CoreV1().Namespaces().Get(name, metav1.GetOptions{})
		return err
	}, e.retryOts...)

	if err != nil {
		return fmt.Errorf("namespace: %s could not be fetched: %v", name, err)
	}
	log.WithField("namespace", name).Info("namespace is created")
	return nil
}

// Create Subscription and wait for successful creation
func (e *e2eTester) createSubscription(subscriberNamespace string, subName string, subscriberEventEndpointURL string) error {
	return retry.Do(func() error {
		_, err := e.subscriptionClient.EventingV1alpha1().Subscriptions(subscriberNamespace).Create(util.NewSubscription(subName, subscriberNamespace, subscriberEventEndpointURL, eventType, "v1", srcID))
		if err == nil {
			return nil
		}
		if !strings.Contains(err.Error(), "already exists") {
			return err
		}
		return nil
	}, e.retryOts...)
}

// Check that the subscriber endpoint is reachable and returns a 200
func (e *e2eTester) checkSubscriberV1EndpointStatus() error {
	return retry.Do(func() error {
		res, err := http.Get(e.subscriber1.eventsURL)
		if err != nil {
			return err
		}
		return verifyStatusCode(res, http.StatusOK)
	}, e.retryOts...)
}

// Check that the subscriber3 endpoint is reachable and returns a 200
func (e *e2eTester) checkSubscriberV3EndpointStatus() error {
	return retry.Do(func() error {
		res, err := http.Get(e.subscriber1.resultsURL)
		if err != nil {
			return err
		}
		return verifyStatusCode(res, http.StatusOK)
	}, e.retryOts...)
}

// Check that the publisher endpoint is reachable and returns a 200
func (e *e2eTester) checkPublisherStatus() error {
	return retry.Do(func() error {
		res, err := http.Get(e.publisher1.publishStatusEndpointURLV1)
		if err != nil {
			return err
		}
		return verifyStatusCode(res, http.StatusOK)
	}, e.retryOts...)
}

// Check that the subscription exists and has condition ready
func (e *e2eTester) checkSubscriptionReady(subscriptionName string) error {
	return retry.Do(func() error {
		var isReady bool
		activatedCondition := subApis.SubscriptionCondition{Type: subApis.Ready, Status: subApis.ConditionTrue}
		kySub, err := e.subscriptionClient.EventingV1alpha1().Subscriptions(e.subscriber1.namespace).Get(subscriptionName, metav1.GetOptions{})
		if err != nil {
			return err
		}
		if isReady = kySub.HasCondition(activatedCondition); !isReady {
			return fmt.Errorf("subscription %v is not ready yet", subscriptionName)
		}
		return nil
	}, e.retryOts...)
}

func defaultRetryOptions() *[]retry.Option {
	return &[]retry.Option{
		retry.Attempts(13), // at max (100 * (1 << 13)) / 1000 = 819,2 sec
		retry.OnRetry(func(n uint, err error) {
			fmt.Printf(".")
		}),
	}
}

func newE2ETester(opts *options, retryOptions *[]retry.Option) *e2eTester {
	e2eTester := &e2eTester{
		publisher1: publisher{
			publishEventEndpointURLV1:  opts.publishEventURLV1,
			publishStatusEndpointURLV1: opts.publishStatusURLV1,
		},
		subscriber1: subscriber{
			namespace:   opts.namespace,
			eventsURL:   fmt.Sprintf("http://%s.%s:%d/events", util.SubscriberName, opts.namespace, port),
			resultsURL:  fmt.Sprintf("http://%s.%s:%d/results", util.SubscriberName, opts.namespace, port),
			statusURL:   fmt.Sprintf("http://%s.%s:%d/status", util.SubscriberName, opts.namespace, port),
			shutdownURL: fmt.Sprintf("http://%s.%s:%d/shutdown", util.SubscriberName, opts.namespace, port),
		},
		retryOts: *retryOptions,
	}
	return e2eTester
}

func getDefaultOptions() *options {
	options := &options{}
	return options
}

func (o *options) parseOrDie() {
	flags := flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	flags.StringVar(&o.publishEventURLV1, "publish-event-uri", "http://event-publish-service:8080/v1/events", "publish service events endpoint `URL`")
	flags.StringVar(&o.publishStatusURLV1, "publish-status-uri", "http://event-publish-service:8080/v1/status/ready", "publish service status endpoint `URL`")
	flags.StringVar(&o.image, "subscriber-image", "", "subscriber Docker `image` name")
	flags.StringVar(&o.namespace, "subscriber-ns", "test-event-bus", "k8s `namespace` in which subscriber test app is running")
	flags.StringVar(&o.logLevel, "log-level", "info", "logrus log level")

	if err := flags.Parse(os.Args[1:]); err != nil {
		panic(err)
	}

	if flags.NFlag() == 0 || len(o.image) == 0 {
		if _, err := fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0]); err != nil {
			panic(err)
		}
		flags.PrintDefaults()
		os.Exit(1)
	}

	// print the effective options
	o.print()
}

func (o *options) print() {
	log.Info(strings.Repeat("-", 100))
	log.Info("publishEventURLV1: ", o.publishEventURLV1)
	log.Info("publishStatusURLV1: ", o.publishStatusURLV1)
	log.Info("image: ", o.image)
	log.Info("namespace: ", o.namespace)
	log.Info("logLevel: ", o.logLevel)
	log.Info(strings.Repeat("-", 100))
}

func setLogLevel(logLevel string) {
	if logLevel, err := log.ParseLevel(logLevel); err != nil {
		panic(err)
	} else {
		log.SetLevel(logLevel)
	}
}
