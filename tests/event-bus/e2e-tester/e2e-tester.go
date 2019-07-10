package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"time"

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
	retries = 20

	idHeader               = "ce-id"
	timeHeader             = "ce-time"
	contentTypeHeader      = "content-type"
	sourceHeader           = "ce-source"
	eventTypeHeader        = "ce-type"
	eventTypeVersionHeader = "ce-eventtypeversion"
	customHeader           = "ce-xcustomheader"
	specVersionHeader      = "ce-specversion"

	ceSourceIDHeaderValue         = "override-source-ID"
	ceEventTypeHeaderValue        = "override-event-type"
	ceSpecVersionHeaderValue      = "0.3"
	contentTypeHeaderValue        = "application/json"
	ceEventTypeVersionHeaderValue = "override-event-type-version"
	customHeaderValue             = "Ce-X-custom-header-value"
)

var (
	clientK8S *kubernetes.Clientset
	eaClient  *eaClientSet.Clientset
	subClient *subscriptionClientSet.Clientset
)

//Unexportable struct, encapsulates subscriber resource parameters
type subscriberDetails struct {
	subscriberImage               string
	subscriberName                string
	subscriberNamespace           string
	subscriberEventEndpointURL    string
	subscriberResultsEndpointURL  string
	subscriberStatusEndpointURL   string
	subscriberShutdownEndpointURL string
	subscriber3EventEndpointURL   string
	subscriber3ResultsEndpointURL string
	subscriber3StatusEndpointURL  string
}

//Unexportable struct, encapsulates publisher details
type publisherDetails struct {
	publishEventEndpointURL  string
	publishStatusEndpointURL string
}

func main() {
	flags := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	var _subscriberDetails subscriberDetails
	var _publisherDetails publisherDetails

	//Initialise publisher struct
	flags.StringVar(&_publisherDetails.publishEventEndpointURL, "publish-event-uri", "http://event-bus-publish:8080/v1/events", "publish service events endpoint `URL`")
	flags.StringVar(&_publisherDetails.publishStatusEndpointURL, "publish-status-uri", "http://event-bus-publish:8080/v1/status/ready", "publish service status endpoint `URL`")

	//Initialise subscriber
	flags.StringVar(&_subscriberDetails.subscriberImage, "subscriber-image", "", "subscriber Docker `image` name")
	flags.StringVar(&_subscriberDetails.subscriberNamespace, "subscriber-ns", "default", "k8s `namespace` in which subscriber test app is running")
	flags.StringVar(&_subscriberDetails.subscriberName, "subscriber-domain", util.SubscriberName, "hostname(without http**) of the deployed subscriber service")
	flags.Parse(os.Args[1:])

	initSubscriberUrls(&_subscriberDetails)

	if flags.NFlag() == 0 || _subscriberDetails.subscriberImage == "" {

		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flags.PrintDefaults()
		os.Exit(1)
	}

	config, err := rest.InClusterConfig()
	if err != nil {
		log.Printf("Error in getting cluster config: %v\n", err)
		shutdown(fail, &_subscriberDetails)
	}

	log.Println("Create the clientK8S")
	clientK8S, err = kubernetes.NewForConfig(config)
	if err != nil {
		log.Printf("Failed to create a ClientSet: %v\n", err)
		shutdown(fail, &_subscriberDetails)
	}

	log.Println("Create a test event activation")
	eaClient, err = eaClientSet.NewForConfig(config)
	if err != nil {
		log.Printf("Error in creating event activation client: %v\n", err)
		shutdown(fail, &_subscriberDetails)
	}
	if !createEventActivation(_subscriberDetails.subscriberNamespace, retries) {
		log.Println("Error: Cannot create the event activation")
		shutdown(fail, &_subscriberDetails)
	}
	time.Sleep(5 * time.Second)

	log.Println("Create a test subscriptions")
	subClient, err = subscriptionClientSet.NewForConfig(config)
	if err != nil {
		log.Printf("Error in creating subscription client: %v\n", err)
		shutdown(fail, &_subscriberDetails)
	}
	if !_subscriberDetails.createSubscription(subscriptionName) {
		log.Println("Error: Cannot create Kyma subscription")
		shutdown(fail, &_subscriberDetails)
	}
	time.Sleep(5 * time.Second)

	log.Println("Create a headers test subscription")
	subClient, err = subscriptionClientSet.NewForConfig(config)
	if err != nil {
		log.Printf("Error in creating headers subscription client: %v\n", err)
		shutdown(fail, &_subscriberDetails)
	}
	if !_subscriberDetails.createSubscription(headersSubscriptionName) {
		log.Println("Error: Cannot create Kyma headers subscription")
		shutdown(fail, &_subscriberDetails)
	}
	time.Sleep(5 * time.Second)

	log.Println("Create Subscriber")
	if err := _subscriberDetails.createSubscriber(); err != nil {
		log.Printf("Create Subscriber failed: %v\n", err)
	}

	log.Println("Check Subscriber Status")
	if !_subscriberDetails.checkSubscriberStatus(retries) {
		log.Println("Error: Cannot connect to Subscriber")
		shutdown(fail, &_subscriberDetails)
	}

	log.Println("Check Subscriber 3 Status")
	if !_subscriberDetails.checkSubscriber3Status(retries) {
		log.Println("Error: Cannot connect to Subscriber 3")
		shutdown(fail, &_subscriberDetails)
	}

	log.Println("Check Publisher Status")
	if !_publisherDetails.checkPublisherStatus(retries) {
		log.Println("Error: Cannot connect to Publisher")
		shutdown(fail, &_subscriberDetails)
	}

	log.Println("Check Kyma subscription ready Status")
	if !_subscriberDetails.checkSubscriptionReady(subscriptionName, retries) {
		log.Println("Error: Kyma Subscription not ready")
		shutdown(fail, &_subscriberDetails)
	}

	log.Println("Check Kyma headers subscription ready Status")
	if !_subscriberDetails.checkSubscriptionReady(headersSubscriptionName, retries) {
		log.Println("Error: Kyma Subscription not ready")
		shutdown(fail, &_subscriberDetails)
	}

	log.Println("Publish an event")
	var eventSent bool
	for i := 0; i < retries; i++ {
		if _, err := publishTestEvent(_publisherDetails.publishEventEndpointURL); err != nil {
			log.Printf("Publish event failed: %v; Retrying (%d/%d)", err, i, retries)
			time.Sleep(time.Duration(i) * time.Second)
		} else {
			eventSent = true
			break
		}
	}
	if !eventSent {
		log.Println("Error: Cannot send test event")
		shutdown(fail, &_subscriberDetails)
	}

	log.Println("Try to read the response from subscriber server")
	if err := _subscriberDetails.checkSubscriberReceivedEvent(); err != nil {
		log.Printf("Error: Cannot get the test event from subscriber: %v\n", err)
		shutdown(fail, &_subscriberDetails)
	}

	log.Println("Publish headers event")
	var headersEventSent bool
	for i := 0; i < retries; i++ {
		if _, err := publishHeadersTestEvent(_publisherDetails.publishEventEndpointURL); err != nil {
			log.Printf("Publish headers event failed: %v; Retrying (%d/%d)", err, i, retries)
			time.Sleep(time.Duration(i) * time.Second)
		} else {
			headersEventSent = true
			break
		}
	}
	if !headersEventSent {
		log.Println("Error: Cannot send test event")
		shutdown(fail, &_subscriberDetails)
	}

	log.Println("Try to read the response from subscriber 3 server")
	if err := _subscriberDetails.checkSubscriberReceivedEventHeaders(); err != nil {
		log.Printf("Error: Cannot get the test event from subscriber 3: %v\n", err)
		shutdown(fail, &_subscriberDetails)
	}

	log.Println("Successfully finished")
	shutdown(success, &_subscriberDetails)
}

// Initialize subscriber urls
func initSubscriberUrls(_subscriberDetails *subscriberDetails) {
	_subscriberDetails.subscriberEventEndpointURL = "http://" + _subscriberDetails.subscriberName + "." + _subscriberDetails.subscriberNamespace + ":9000/v1/events"
	_subscriberDetails.subscriberResultsEndpointURL = "http://" + _subscriberDetails.subscriberName + "." + _subscriberDetails.subscriberNamespace + ":9000/v1/results"
	_subscriberDetails.subscriberStatusEndpointURL = "http://" + _subscriberDetails.subscriberName + "." + _subscriberDetails.subscriberNamespace + ":9000/v1/status"
	_subscriberDetails.subscriberShutdownEndpointURL = "http://" + _subscriberDetails.subscriberName + "." + _subscriberDetails.subscriberNamespace + ":9000/shutdown"
	_subscriberDetails.subscriber3EventEndpointURL = "http://" + _subscriberDetails.subscriberName + "." + _subscriberDetails.subscriberNamespace + ":9000/v3/events"
	_subscriberDetails.subscriber3ResultsEndpointURL = "http://" + _subscriberDetails.subscriberName + "." + _subscriberDetails.subscriberNamespace + ":9000/v3/results"
	_subscriberDetails.subscriber3StatusEndpointURL = "http://" + _subscriberDetails.subscriberName + "." + _subscriberDetails.subscriberNamespace + ":9000/v3/status"
}

func shutdown(code int, _subscriberDetails *subscriberDetails) {
	log.Println("Send shutdown request to Subscriber")
	if _, err := http.Post(_subscriberDetails.subscriberShutdownEndpointURL, "application/json", strings.NewReader(`{"shutdown": "true"}`)); err != nil {
		log.Printf("Warning: Shutdown Subscriber falied: %v", err)
	}
	log.Println("Delete Subscriber deployment")
	deletePolicy := metav1.DeletePropagationForeground
	gracePeriodSeconds := int64(0)

	if err := clientK8S.AppsV1().Deployments(_subscriberDetails.subscriberNamespace).Delete(_subscriberDetails.subscriberName,
		&metav1.DeleteOptions{GracePeriodSeconds: &gracePeriodSeconds, PropagationPolicy: &deletePolicy}); err != nil {
		log.Printf("Warning: Delete Subscriber Deployment falied: %v", err)
	}
	log.Println("Delete Subscriber service")
	if err := clientK8S.CoreV1().Services(_subscriberDetails.subscriberNamespace).Delete(_subscriberDetails.subscriberName,
		&metav1.DeleteOptions{GracePeriodSeconds: &gracePeriodSeconds}); err != nil {
		log.Printf("Warning: Delete Subscriber Service falied: %v", err)
	}
	if subClient != nil {
		log.Printf("Delete test subscription: %v\n", subscriptionName)
		if err := subClient.EventingV1alpha1().Subscriptions(_subscriberDetails.subscriberNamespace).Delete(subscriptionName,
			&metav1.DeleteOptions{PropagationPolicy: &deletePolicy}); err != nil {
			log.Printf("Warning: Delete Subscription falied: %v", err)
		}
	}
	if subClient != nil {
		log.Printf("Delete headers test subscription: %v\n", subscriptionName)
		if err := subClient.EventingV1alpha1().Subscriptions(_subscriberDetails.subscriberNamespace).Delete(headersSubscriptionName,
			&metav1.DeleteOptions{PropagationPolicy: &deletePolicy}); err != nil {
			log.Printf("Warning: Delete Subscription falied: %v", err)
		}
	}
	if eaClient != nil {
		log.Printf("Delete test event activation: %v\n", eventActivationName)
		if err := eaClient.ApplicationconnectorV1alpha1().EventActivations(_subscriberDetails.subscriberNamespace).Delete(eventActivationName, &metav1.DeleteOptions{PropagationPolicy: &deletePolicy}); err != nil {
			log.Printf("Warning: Delete Event Activation falied: %v", err)
		}
	}
	//needed if Istio injection will be enabled for tester cause the kyma script which runs the tests checks the pod status
	/*
		if _, err := clientK8S.CoreV1().Pods("kyma-system").Get(testerName, metav1.GetOptions{}); err != nil {
			log.Printf("Cannot get my pod: %v", err)
		} else {
			if code == success {
				body := "{\"status\":{\"phase\":\"" + v1.PodSucceeded + "\"}}"
				_, err = clientK8S.CoreV1().Pods("kyma-system").Patch(testerName, types.MergePatchType, []byte(body), "status")
			} else {
				body := "{\"status\":{\"phase\":\"" + v1.PodFailed + "\"}}"
				_, err = clientK8S.CoreV1().Pods("kyma-system").Patch(testerName, types.MergePatchType, []byte(body), "status")
			}
			if err != nil {
				log.Printf("Cannot set status: %v", err)
			}
		}
	*/
	os.Exit(code)
}

func publishTestEvent(publishEventURL string) (*api.Response, error) {
	payload := fmt.Sprintf(
		`{"source-id": "%s","event-type":"%s","event-type-version":"%s","event-time":"2018-11-02T22:08:41+00:00","data":"test-event-1"}`, srcID, eventType, eventTypeVersion)
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
	respObj := &api.Response{}
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

func publishHeadersTestEvent(publishEventURL string) (*api.Response, error) {
	payload := fmt.Sprintf(
		`{"source-id": "%s","event-type":"%s","event-type-version":"%s","event-time":"2018-11-02T22:08:41+00:00","data":"headers-test-event"}`, srcID, eventType, eventTypeVersion)
	log.Printf("event to be published: %v\n", payload)

	client := &http.Client{}
	req, err := http.NewRequest("POST", publishEventURL, strings.NewReader(payload))
	req.Header.Add(sourceHeader, ceSourceIDHeaderValue)
	req.Header.Add(eventTypeHeader, ceEventTypeHeaderValue)
	req.Header.Add(eventTypeVersionHeader, ceEventTypeVersionHeaderValue)
	req.Header.Add(customHeader, customHeaderValue)
	res, err := client.Do(req)

	if err != nil {
		log.Printf("Post request failed: %v\n", err)
		return nil, err
	}
	dumpResponse(res)
	if err := verifyStatusCode(res, 200); err != nil {
		return nil, err
	}
	respObj := &api.Response{}
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

func (_subscriberDetails *subscriberDetails) checkSubscriberReceivedEvent() error {
	for i := 0; i < retries; i++ {
		time.Sleep(time.Duration(i) * time.Second)
		log.Printf("Get subscriber response (%d/%d)\n", i, retries)
		res, err := http.Get(_subscriberDetails.subscriberResultsEndpointURL)
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

func (_subscriberDetails *subscriberDetails) checkSubscriberReceivedEventHeaders() error {
	for i := 0; i < retries; i++ {
		time.Sleep(time.Duration(i) * time.Second)
		log.Printf("Get subscriber 3 response (%d/%d)\n", i, retries)
		res, err := http.Get(_subscriberDetails.subscriber3ResultsEndpointURL)
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
		var resp map[string][]string
		json.Unmarshal(body, &resp)
		log.Printf("Subscriber 3 response: %v\n", resp)
		res.Body.Close()
		if len(resp) == 0 { // no event received by subscriber
			continue
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
			{headerKey: specVersionHeader, headerExpectedValue: ceSpecVersionHeaderValue},
			{headerKey: customHeader, headerExpectedValue: customHeaderValue},
		}

		for _, testData := range testDataSets {
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
	}
	return errors.New("timeout for subscriber response")
}

func (_subscriberDetails *subscriberDetails) createSubscriber() error {
	if _, err := clientK8S.AppsV1().Deployments(_subscriberDetails.subscriberNamespace).Get(_subscriberDetails.subscriberName, metav1.GetOptions{}); err != nil {
		log.Println("Create Subscriber deployment")
		if _, err := clientK8S.AppsV1().Deployments(_subscriberDetails.subscriberNamespace).Create(util.NewSubscriberDeployment(_subscriberDetails.subscriberImage)); err != nil {
			log.Printf("Create Subscriber deployment: %v\n", err)
			return err
		}
		log.Println("Create Subscriber service")
		if _, err := clientK8S.CoreV1().Services(_subscriberDetails.subscriberNamespace).Create(util.NewSubscriberService()); err != nil {
			log.Printf("Create Subscriber service failed: %v\n", err)
			return err
		}
		time.Sleep(30 * time.Second)

		for i := 0; i < 60; i++ {
			var podReady bool
			if pods, err := clientK8S.CoreV1().Pods(_subscriberDetails.subscriberNamespace).List(metav1.ListOptions{LabelSelector: "app=" + _subscriberDetails.subscriberName}); err != nil {
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

func (_subscriberDetails *subscriberDetails) createSubscription(subName string) bool {
	_, err := subClient.EventingV1alpha1().Subscriptions(_subscriberDetails.subscriberNamespace).Create(util.NewSubscription(subName, _subscriberDetails.subscriberNamespace, _subscriberDetails.subscriberEventEndpointURL, eventType, "v1", srcID))
	if err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			log.Printf("Error in creating subscription: %v\n", err)
			return false
		}
	}
	return true
}

func (_subscriberDetails *subscriberDetails) checkSubscriberStatus(noOfRetries int) bool {
	var subscriberOK bool
	for i := 0; i < noOfRetries; i++ {
		if res, err := http.Get(_subscriberDetails.subscriberStatusEndpointURL); err != nil {
			log.Printf("Subscriber Status request failed: %v; Retrying (%d/%d)", err, i, noOfRetries)
			time.Sleep(time.Duration(i) * time.Second)
		} else if !checkStatusCode(res, http.StatusOK) {
			log.Printf("Subscriber Server Status request returns: %v; Retrying (%d/%d)\n", res, i, noOfRetries)
			time.Sleep(time.Duration(i) * time.Second)
		} else {
			subscriberOK = true
			break
		}
	}
	return subscriberOK
}

func (_subscriberDetails *subscriberDetails) checkSubscriber3Status(noOfRetries int) bool {
	var subscriberOK bool
	for i := 0; i < noOfRetries; i++ {
		if res, err := http.Get(_subscriberDetails.subscriber3StatusEndpointURL); err != nil {
			log.Printf("Subscriber 3 Status request failed: %v; Retrying (%d/%d)", err, i, noOfRetries)
			time.Sleep(time.Duration(i) * time.Second)
		} else if !checkStatusCode(res, http.StatusOK) {
			log.Printf("Subscriber 3 Server Status request returns: %v; Retrying (%d/%d)\n", res, i, noOfRetries)
			time.Sleep(time.Duration(i) * time.Second)
		} else {
			subscriberOK = true
			break
		}
	}
	return subscriberOK
}

func (_subscriberDetails *subscriberDetails) checkSubscriptionReady(subscriptionName string, noOfRetries int) bool {
	var isReady bool
	activatedCondition := subApis.SubscriptionCondition{Type: subApis.Ready, Status: subApis.ConditionTrue}
	for i := 0; i < noOfRetries && !isReady; i++ {
		kySub, err := subClient.EventingV1alpha1().Subscriptions(_subscriberDetails.subscriberNamespace).Get(subscriptionName, metav1.GetOptions{})
		if err != nil {
			log.Printf("Cannot get Kyma subscription, name: %v; namespace: %v", subscriptionName, _subscriberDetails.subscriberNamespace)
			break
		} else {
			if isReady = kySub.HasCondition(activatedCondition); !isReady {
				time.Sleep(1 * time.Second)
			}
		}
	}
	return isReady
}

func (_publisherDetails *publisherDetails) checkPublisherStatus(noOfRetries int) bool {
	var publishOK bool
	for i := 0; i < noOfRetries; i++ {
		if err := checkPublishStatus(_publisherDetails.publishStatusEndpointURL); err != nil {
			log.Printf("Publisher not ready: %v", err)
			time.Sleep(time.Duration(i) * time.Second)
		} else {
			publishOK = true
			break
		}
	}
	return publishOK
}

func createEventActivation(subscriberNamespace string, noOfRetries int) bool {
	var eventActivationOK bool
	var err error
	for i := 0; i < noOfRetries; i++ {
		_, err = eaClient.ApplicationconnectorV1alpha1().EventActivations(subscriberNamespace).Create(util.NewEventActivation(eventActivationName, subscriberNamespace, srcID))
		if err == nil {
			eventActivationOK = true
			break
		}
		if !strings.Contains(err.Error(), "already exists") {
			log.Printf("Error in creating event activation - %v; Retrying (%d/%d)\n", err, i, noOfRetries)
			time.Sleep(1 * time.Second)
		} else {
			eventActivationOK = true
			break
		}
	}
	return eventActivationOK
}

func dumpResponse(resp *http.Response) {
	defer resp.Body.Close()
	_, err := httputil.DumpResponse(resp, true)
	if err != nil {
		log.Fatal(err)
	}
}

func checkStatusCode(res *http.Response, expectedStatusCode int) bool {
	if res.StatusCode != expectedStatusCode {
		log.Printf("Status code is wrong, have: %d, want: %d\n", res.StatusCode, expectedStatusCode)
		return false
	}
	return true
}

func checkPublishStatus(statusEndpointURL string) error {
	res, err := http.Get(statusEndpointURL)
	if err != nil {
		return err
	}
	return verifyStatusCode(res, http.StatusOK)
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
