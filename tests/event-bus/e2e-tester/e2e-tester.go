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

	"github.com/avast/retry-go"
	api "github.com/kyma-project/kyma/components/event-bus/api/publish"
	subApis "github.com/kyma-project/kyma/components/event-bus/api/push/eventing.kyma-project.io/v1alpha1"
	eaClientSet "github.com/kyma-project/kyma/components/event-bus/generated/ea/clientset/versioned"
	subscriptionClientSet "github.com/kyma-project/kyma/components/event-bus/generated/push/clientset/versioned"
	"github.com/kyma-project/kyma/components/event-bus/test/util"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	eventType               = "test-e2e"
	eventTypeVersion        = "v1"
	subscriptionName        = "test-sub"
	headersSubscriptionName = "headers-test-sub"
	eventActivationName     = "test-ea"
	srcID                   = "test.local"

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
)

var (
	clientK8S    *kubernetes.Clientset
	eaClient     *eaClientSet.Clientset
	subClient    *subscriptionClientSet.Clientset
	retryOptions = []retry.Option{
		retry.Attempts(13), // at max (100 * (1 << 13)) / 1000 = 819,2 sec
		retry.OnRetry(func(n uint, err error) {
			fmt.Printf(".")
		}),
	}
)

//Unexportable struct, encapsulates subscriber resource parameters
type testSubscriber struct {
	image                 string
	namespace             string
	eventEndpointV1URL    string
	resultsEndpointV1URL  string
	statusEndpointV1URL   string
	shutdownEndpointV1URL string
	eventEndpointV3URL    string
	resultsEndpointV3URL  string
	statusEndpointV3URL   string
}

//Unexportable struct, encapsulates publisher details
type publisherDetails struct {
	publishEventEndpointURL  string
	publishStatusEndpointURL string
}

func main() {
	// configure logger with text instead of json for easier reading in CI logs
	log.SetFormatter(&log.TextFormatter{})
	// show file and line number
	log.SetReportCaller(true)

	flags := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	var subscriber testSubscriber
	var pubDetails publisherDetails
	var logLevelString string
	var logLevel log.Level

	//Initialise publisher struct
	flags.StringVar(&pubDetails.publishEventEndpointURL, "publish-event-uri", "http://event-publish-service:8080/v1/events", "publish service events endpoint `URL`")
	flags.StringVar(&pubDetails.publishStatusEndpointURL, "publish-status-uri", "http://event-publish-service:8080/v1/status/ready", "publish service status endpoint `URL`")

	//Initialise subscriber
	flags.StringVar(&subscriber.image, "subscriber-image", "", "subscriber Docker `image` name")
	flags.StringVar(&subscriber.namespace, "subscriber-ns", "test-event-bus", "k8s `namespace` in which subscriber test app is running")
	flags.StringVar(&logLevelString, "log-level", "info", "logrus log level")

	if err := flags.Parse(os.Args[1:]); err != nil {
		panic(err)
	}

	// set log level
	var err error
	if logLevel, err = log.ParseLevel(logLevelString); err != nil {
		panic(err)
	}
	log.SetLevel(logLevel)

	initSubscriberUrls(&subscriber)

	if flags.NFlag() == 0 || subscriber.image == "" {

		if _, err := fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0]); err != nil {
			panic(err)
		}
		flags.PrintDefaults()
		os.Exit(1)
	}

	config, err := rest.InClusterConfig()
	if err != nil {
		log.WithField("error", err).Error("error in getting cluster config")
		shutdown(fail, &subscriber)
	}

	log.Info("create the clientK8S")
	clientK8S, err = kubernetes.NewForConfig(config)
	if err != nil {
		log.WithField("error", err).Error("failed to create a ClientSet")
		shutdown(fail, &subscriber)
	}

	err = createNamespace(subscriber.namespace)
	if err != nil {
		log.WithField("error", err).Error("cannot create namespace")
		shutdown(fail, &subscriber)
	}

	log.Info("create a test event activation")
	eaClient, err = eaClientSet.NewForConfig(config)
	if err != nil {
		log.WithField("error", err).Error("error in creating EventActivation client")
		shutdown(fail, &subscriber)
	}
	if err := createEventActivation(subscriber.namespace); err != nil {
		log.WithField("error", err).Error("cannot create the event activation")
		shutdown(fail, &subscriber)
	}

	log.Info("create a test Subscription")
	subClient, err = subscriptionClientSet.NewForConfig(config)
	if err != nil {
		log.WithField("error", err).Error("error in creating Subscription client")
		shutdown(fail, &subscriber)
	}
	if err := createSubscription(subscriber.namespace, subscriptionName, subscriber.eventEndpointV1URL); err != nil {
		log.WithField("error", err).Error("cannot create Kyma subscription")
		shutdown(fail, &subscriber)
	}

	log.Info("create a headers test subscription")
	subClient, err = subscriptionClientSet.NewForConfig(config)
	if err != nil {
		log.WithField("error", err).Error("error in creating headers subscription client")
		shutdown(fail, &subscriber)
	}
	if err := createSubscription(subscriber.namespace, headersSubscriptionName, subscriber.eventEndpointV3URL); err != nil {
		log.WithField("error", err).Error("cannot create Kyma headers subscription")
		shutdown(fail, &subscriber)
	}

	log.Info("create Subscriber")
	if err := createSubscriber(util.SubscriberName, subscriber.namespace, subscriber.image); err != nil {
		log.WithField("error", err).Error("create Subscriber failed")
	}

	log.Info("check Subscriber's v1 endpoint Status")
	if err := subscriber.checkSubscriberV1EndpointStatus(); err != nil {
		log.WithField("error", err).Error("cannot connect to Subscriber v1 endpoint")
		shutdown(fail, &subscriber)
	}

	log.Info("check Subscriber's v3 endpoint Status")
	if err := subscriber.checkSubscriberV3EndpointStatus(); err != nil {
		log.WithField("error", err).Info("cannot connect to Subscriber v3 endpoint")
		shutdown(fail, &subscriber)
	}

	log.Info("check Publisher Status")
	if err := pubDetails.checkPublisherStatus(); err != nil {
		log.WithField("error", err).Error("cannot connect to Publisher")
		shutdown(fail, &subscriber)
	}

	log.Info("check Kyma subscription ready Status")
	if err := subscriber.checkSubscriptionReady(subscriptionName); err != nil {
		log.WithField("error", err).Error("kyma Subscription not ready")
		shutdown(fail, &subscriber)
	}

	log.Info("check Kyma headers subscription ready Status")
	if err := subscriber.checkSubscriptionReady(headersSubscriptionName); err != nil {
		log.WithField("error", err).Error("kyma Subscription not ready")
		shutdown(fail, &subscriber)
	}

	log.Info("publish an event")
	err = retry.Do(func() error {
		_, err := publishTestEvent(pubDetails.publishEventEndpointURL)
		return err
	}, retryOptions...)
	if err != nil {
		log.WithField("error", err).Error("publish event failed")
		shutdown(fail, &subscriber)
	}

	log.Info("try to read the response from subscriber server")
	if err := subscriber.checkReceivedEvent(); err != nil {
		log.WithField("error", err).Error("cannot get the test event from subscriber")
		shutdown(fail, &subscriber)
	}

	log.Info("publish headers event")
	err = retry.Do(func() error {
		_, err := publishHeadersTestEvent(pubDetails.publishEventEndpointURL)
		return err
	}, retryOptions...)
	if err != nil {
		log.WithField("error", err).Error("publish for an event with headers failed")
		shutdown(fail, &subscriber)
	}

	log.Info("try to read the response from v3 endpoint of the subscriber")
	if err := subscriber.checkReceivedEventHeaders(); err != nil {
		log.WithField("error", err).Error("cannot get the test event from subscriber v3 endpoint")
		shutdown(fail, &subscriber)
	}

	log.Info("successfully finished")
	shutdown(success, &subscriber)
}

// Initialize subscriber urls
func initSubscriberUrls(subscriber *testSubscriber) {
	subscriber.eventEndpointV1URL = "http://" + util.SubscriberName + "." + subscriber.namespace + ":9000/v1/events"
	subscriber.resultsEndpointV1URL = "http://" + util.SubscriberName + "." + subscriber.namespace + ":9000/v1/results"
	subscriber.statusEndpointV1URL = "http://" + util.SubscriberName + "." + subscriber.namespace + ":9000/v1/status"
	subscriber.shutdownEndpointV1URL = "http://" + util.SubscriberName + "." + subscriber.namespace + ":9000/shutdown"
	subscriber.eventEndpointV3URL = "http://" + util.SubscriberName + "." + subscriber.namespace + ":9000/v3/events"
	subscriber.resultsEndpointV3URL = "http://" + util.SubscriberName + "." + subscriber.namespace + ":9000/v3/results"
	subscriber.statusEndpointV3URL = "http://" + util.SubscriberName + "." + subscriber.namespace + ":9000/v3/status"
}

func shutdown(code int, subscriber *testSubscriber) {
	log.Info("send shutdown request to Subscriber")
	if _, err := http.Post(subscriber.shutdownEndpointV1URL, "application/json", strings.NewReader(`{"shutdown": "true"}`)); err != nil {
		log.WithField("error", err).Warning("shutdown Subscriber failed")
	}
	log.Info("delete Subscriber deployment")
	deletePolicy := metav1.DeletePropagationForeground
	gracePeriodSeconds := int64(0)

	if err := clientK8S.AppsV1().Deployments(subscriber.namespace).Delete(util.SubscriberName,
		&metav1.DeleteOptions{GracePeriodSeconds: &gracePeriodSeconds, PropagationPolicy: &deletePolicy}); err != nil {
		log.WithField("error", err).Warn("delete Subscriber Deployment failed")
	}
	log.Info("delete Subscriber service")
	if err := clientK8S.CoreV1().Services(subscriber.namespace).Delete(util.SubscriberName,
		&metav1.DeleteOptions{GracePeriodSeconds: &gracePeriodSeconds}); err != nil {
		log.WithField("error", err).Warn("delete Subscriber Service failed")
	}
	if subClient != nil {
		log.WithField("subscription", subscriptionName).Info("delete test subscription")
		if err := subClient.EventingV1alpha1().Subscriptions(subscriber.namespace).Delete(subscriptionName,
			&metav1.DeleteOptions{PropagationPolicy: &deletePolicy}); err != nil {
			log.WithField("error", err).Warn("delete Subscription failed")
		}
	}
	if subClient != nil {
		log.WithField("subscription", subscriptionName).Info("delete headers test subscription")
		if err := subClient.EventingV1alpha1().Subscriptions(subscriber.namespace).Delete(headersSubscriptionName,
			&metav1.DeleteOptions{PropagationPolicy: &deletePolicy}); err != nil {
			log.WithField("error", err).Warn("delete Subscription failed")
		}
	}
	if eaClient != nil {
		log.WithField("event_activation", eventActivationName).Info("delete test event activation")
		if err := eaClient.ApplicationconnectorV1alpha1().EventActivations(subscriber.namespace).Delete(eventActivationName, &metav1.DeleteOptions{PropagationPolicy: &deletePolicy}); err != nil {
			log.WithField("error", err).Warn("delete Event Activation failed")
		}
	}

	log.WithField("namespace", subscriber.namespace).Info("delete test namespace")
	if err := clientK8S.Core().Namespaces().Delete(subscriber.namespace, &metav1.DeleteOptions{PropagationPolicy: &deletePolicy}); err != nil {
		log.WithField("error", err).Warn("delete Namespace failed")
	}
	os.Exit(code)
}

func publishTestEvent(publishEventURL string) (*api.Response, error) {
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

func publishHeadersTestEvent(publishEventURL string) (*api.Response, error) {
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

func (subscriber *testSubscriber) checkReceivedEvent() error {
	return retry.Do(func() error {
		res, err := http.Get(subscriber.resultsEndpointV1URL)
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
	}, retryOptions...)
}

func (subscriber *testSubscriber) checkReceivedEventHeaders() error {
	return retry.Do(func() error {
		res, err := http.Get(subscriber.resultsEndpointV3URL)
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
	}, retryOptions...)
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

func createSubscriber(subscriberName string, subscriberNamespace string, sbscrImg string) error {
	if _, err := clientK8S.AppsV1().Deployments(subscriberNamespace).Get(subscriberName, metav1.GetOptions{}); err != nil {
		log.Info("create Subscriber deployment")
		if _, err := clientK8S.AppsV1().Deployments(subscriberNamespace).Create(util.NewSubscriberDeployment(sbscrImg)); err != nil {
			log.WithField("error", err).Error("create Subscriber deployment failed")
			return err
		}
		log.Info("create Subscriber service")
		if _, err := clientK8S.CoreV1().Services(subscriberNamespace).Create(util.NewSubscriberService()); err != nil {
			log.WithField("error", err).Error("create Subscriber service failed")
			return err
		}

		// wait until pod is ready
		return retry.Do(func() error {
			var podReady bool
			var podNotReady error
			pods, err := clientK8S.CoreV1().Pods(subscriberNamespace).List(metav1.ListOptions{LabelSelector: "app=" + util.SubscriberName})
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
		}, retryOptions...)
	}
	return nil
}

// Create EventActivation and wait for successful creation
func createEventActivation(subscriberNamespace string) error {
	return retry.Do(func() error {
		_, err := eaClient.ApplicationconnectorV1alpha1().EventActivations(subscriberNamespace).Create(util.NewEventActivation(eventActivationName, subscriberNamespace, srcID))
		if err == nil {
			return nil
		}
		if !strings.Contains(err.Error(), "already exists") {
			return err
		}
		return nil
	})
}

func createNamespace(name string) error {

	ns := &apiv1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"test": "test-event-bus",
			},
		},
	}
	err := retry.Do(func() error {
		_, err := clientK8S.Core().Namespaces().Create(ns)
		return err
	}, retryOptions...)

	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return fmt.Errorf("namespace: %s could not be created: %v", name, err)
	}

	err = retry.Do(func() error {
		_, err = clientK8S.Core().Namespaces().Get(name, metav1.GetOptions{})
		return err
	}, retryOptions...)

	if err != nil {
		return fmt.Errorf("namespace: %s could not be fetched: %v", name, err)
	}
	log.WithField("namespace", name).Info("namespace is created")
	return nil
}

// Create Subscription and wait for successful creation
func createSubscription(subscriberNamespace string, subName string, subscriberEventEndpointURL string) error {
	return retry.Do(func() error {
		_, err := subClient.EventingV1alpha1().Subscriptions(subscriberNamespace).Create(util.NewSubscription(subName, subscriberNamespace, subscriberEventEndpointURL, eventType, "v1", srcID))
		if err == nil {
			return nil
		}
		if !strings.Contains(err.Error(), "already exists") {
			return err
		}
		return nil
	}, retryOptions...)
}

// Check that the subscriber endpoint is reachable and returns a 200
func (subscriber *testSubscriber) checkSubscriberV1EndpointStatus() error {
	return retry.Do(func() error {
		res, err := http.Get(subscriber.eventEndpointV1URL)
		if err != nil {
			return err
		}
		return verifyStatusCode(res, http.StatusOK)
	}, retryOptions...)
}

// Check that the subscriber3 endpoint is reachable and returns a 200
func (subscriber *testSubscriber) checkSubscriberV3EndpointStatus() error {
	return retry.Do(func() error {
		res, err := http.Get(subscriber.resultsEndpointV3URL)
		if err != nil {
			return err
		}
		return verifyStatusCode(res, http.StatusOK)
	}, retryOptions...)
}

// Check that the publisher endpoint is reachable and returns a 200
func (pubDetails *publisherDetails) checkPublisherStatus() error {
	return retry.Do(func() error {
		res, err := http.Get(pubDetails.publishStatusEndpointURL)
		if err != nil {
			return err
		}
		return verifyStatusCode(res, http.StatusOK)
	}, retryOptions...)
}

// Check that the subscription exists and has condition ready
func (subscriber *testSubscriber) checkSubscriptionReady(subscriptionName string) error {
	return retry.Do(func() error {
		var isReady bool
		activatedCondition := subApis.SubscriptionCondition{Type: subApis.Ready, Status: subApis.ConditionTrue}
		kySub, err := subClient.EventingV1alpha1().Subscriptions(subscriber.namespace).Get(subscriptionName, metav1.GetOptions{})
		if err != nil {
			return err
		}
		if isReady = kySub.HasCondition(activatedCondition); !isReady {
			return fmt.Errorf("subscription %v is not ready yet", subscriptionName)
		}
		return nil
	}, retryOptions...)
}
