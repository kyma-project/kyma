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

	sourceIDHeader         = "Ce-Source-Id"
	eventTypeHeader        = "Ce-Event-Type"
	eventTypeVersionHeader = "Ce-Event-Type-Version"
	customHeader           = "Ce-X-Custom-Header"

	ceSourceIDHeaderValue         = "override-source-ID"
	ceEventTypeHeaderValue        = "override-event-type"
	ceEventTypeVersionHeaderValue = "override-event-type-version"
	customHeaderValue             = "Ce-X-custom-header-value"
)

var (
	subscriberNamespace           *string
	subscriberEventEndpointURL    *string
	subscriberResultsEndpointURL  *string
	subscriberStatusEndpointURL   *string
	subscriberShutdownEndpointURL *string
	subscriber3EventEndpointURL   *string
	subscriber3ResultsEndpointURL *string
	subscriber3StatusEndpointURL  *string
	clientK8S                     *kubernetes.Clientset
	eaClient                      *eaClientSet.Clientset
	subClient                     *subscriptionClientSet.Clientset
	publishEventEndpointURL       *string
	publishStatusEndpointURL      *string
)

func main() {
	flags := flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	publishEventEndpointURL = flags.String("publish-event-uri", "http://event-bus-publish:8080/v1/events", "publish service events endpoint `URL`")
	publishStatusEndpointURL = flags.String("publish-status-uri", "http://event-bus-publish:8080/v1/status/ready", "publish service status endpoint `URL`")
	subscriberImage := flags.String("subscriber-image", "", "subscriber Docker `image` name")
	subscriberNamespace = flags.String("subscriber-ns", "default", "k8s `namespace` in which subscriber test app is running")
	subscriberEventEndpointURL = flags.String("subscriber-events-uri",
		"http://"+util.SubscriberName+"."+*subscriberNamespace+":9000/v1/events", "subscriber service events endpoint `URL`")
	subscriberResultsEndpointURL = flags.String("subscriber-results-uri",
		"http://"+util.SubscriberName+"."+*subscriberNamespace+":9000/v1/results", "subscriber service results endpoint `URL`")
	subscriberStatusEndpointURL = flags.String("subscriber-status-uri",
		"http://"+util.SubscriberName+"."+*subscriberNamespace+":9000/v1/status", "subscriber service status endpoint `URL`")
	subscriberShutdownEndpointURL = flags.String("subscriber-shutdown-uri",
		"http://"+util.SubscriberName+"."+*subscriberNamespace+":9000/shutdown", "subscriber service shutdown endpoint `URL`")

	subscriber3EventEndpointURL = flags.String("subscriber3-events-uri",
		"http://"+util.SubscriberName+"."+*subscriberNamespace+":9000/v3/events", "subscriber 3 service events endpoint `URL`")
	subscriber3ResultsEndpointURL = flags.String("subscriber3-results-uri",
		"http://"+util.SubscriberName+"."+*subscriberNamespace+":9000/v3/results", "subscriber 3 service results endpoint `URL`")
	subscriber3StatusEndpointURL = flags.String("subscriber3-status-uri",
		"http://"+util.SubscriberName+"."+*subscriberNamespace+":9000/v3/status", "subscriber 3 service status endpoint `URL`")

	flags.Parse(os.Args[1:])

	if flags.NFlag() == 0 || *subscriberImage == "" {

		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flags.PrintDefaults()
		os.Exit(1)
	}

	config, err := rest.InClusterConfig()
	if err != nil {
		log.Printf("Error in getting cluster config: %v\n", err)
		shutdown(fail)
	}

	log.Println("Create the clientK8S")
	clientK8S, err = kubernetes.NewForConfig(config)
	if err != nil {
		log.Printf("Failed to create a ClientSet: %v\n", err)
		shutdown(fail)
	}

	log.Println("Create a test event activation")
	eaClient, err = eaClientSet.NewForConfig(config)
	if err != nil {
		log.Printf("Error in creating event activation client: %v\n", err)
		shutdown(fail)
	}
	if !createEventActivation(*subscriberNamespace, retries) {
		log.Println("Error: Cannot create the event activation")
		shutdown(fail)
	}
	time.Sleep(5 * time.Second)

	log.Println("Create a test subscriptions")
	subClient, err = subscriptionClientSet.NewForConfig(config)
	if err != nil {
		log.Printf("Error in creating subscription client: %v\n", err)
		shutdown(fail)
	}
	if !createSubscription(*subscriberNamespace, subscriptionName, *subscriberEventEndpointURL) {
		log.Println("Error: Cannot create Kyma subscription")
		shutdown(fail)
	}
	time.Sleep(5 * time.Second)

	log.Println("Create a headers test subscription")
	subClient, err = subscriptionClientSet.NewForConfig(config)
	if err != nil {
		log.Printf("Error in creating headers subscription client: %v\n", err)
		shutdown(fail)
	}
	if !createSubscription(*subscriberNamespace, headersSubscriptionName, *subscriber3EventEndpointURL) {
		log.Println("Error: Cannot create Kyma headers subscription")
		shutdown(fail)
	}
	time.Sleep(5 * time.Second)

	log.Println("Create Subscriber")
	if err := createSubscriber(util.SubscriberName, *subscriberNamespace, *subscriberImage); err != nil {
		log.Printf("Create Subscriber failed: %v\n", err)
	}

	log.Println("Check Subscriber Status")
	if !checkSubscriberStatus(retries) {
		log.Println("Error: Cannot connect to Subscriber")
		shutdown(fail)
	}

	log.Println("Check Subscriber 3 Status")
	if !checkSubscriber3Status(retries) {
		log.Println("Error: Cannot connect to Subscriber 3")
		shutdown(fail)
	}

	log.Println("Check Publisher Status")
	if !checkPublisherStatus(retries) {
		log.Println("Error: Cannot connect to Publisher")
		shutdown(fail)
	}

	log.Println("Check Kyma subscription ready Status")
	if !checkSubscriptionReady(*subscriberNamespace, subscriptionName, retries) {
		log.Println("Error: Kyma Subscription not ready")
		shutdown(fail)
	}

	log.Println("Check Kyma headers subscription ready Status")
	if !checkSubscriptionReady(*subscriberNamespace, headersSubscriptionName, retries) {
		log.Println("Error: Kyma Subscription not ready")
		shutdown(fail)
	}

	log.Println("Publish an event")
	var eventSent bool
	for i := 0; i < retries; i++ {
		if _, err := publishTestEvent(*publishEventEndpointURL); err != nil {
			log.Printf("Publish event failed: %v; Retrying (%d/%d)", err, i, retries)
			time.Sleep(time.Duration(i) * time.Second)
		} else {
			eventSent = true
			break
		}
	}
	if !eventSent {
		log.Println("Error: Cannot send test event")
		shutdown(fail)
	}

	log.Println("Try to read the response from subscriber server")
	if err := checkSubscriberReceivedEvent(); err != nil {
		log.Printf("Error: Cannot get the test event from subscriber: %v\n", err)
		shutdown(fail)
	}

	log.Println("Publish headers event")
	var headersEventSent bool
	for i := 0; i < retries; i++ {
		if _, err := publishHeadersTestEvent(*publishEventEndpointURL); err != nil {
			log.Printf("Publish headers event failed: %v; Retrying (%d/%d)", err, i, retries)
			time.Sleep(time.Duration(i) * time.Second)
		} else {
			headersEventSent = true
			break
		}
	}
	if !headersEventSent {
		log.Println("Error: Cannot send test event")
		shutdown(fail)
	}

	log.Println("Try to read the response from subscriber 3 server")
	if err := checkSubscriberReceivedEventHeaders(); err != nil {
		log.Printf("Error: Cannot get the test event from subscriber 3: %v\n", err)
		shutdown(fail)
	}

	log.Println("Successfully finished")
	shutdown(success)
}

func shutdown(code int) {
	log.Println("Send shutdown request to Subscriber")
	if _, err := http.Post(*subscriberShutdownEndpointURL, "application/json", strings.NewReader(`{"shutdown": "true"}`)); err != nil {
		log.Printf("Warning: Shutdown Subscriber falied: %v", err)
	}
	log.Println("Delete Subscriber deployment")
	deletePolicy := metav1.DeletePropagationForeground
	gracePeriodSeconds := int64(0)
	if err := clientK8S.AppsV1().Deployments(*subscriberNamespace).Delete(util.SubscriberName,
		&metav1.DeleteOptions{GracePeriodSeconds: &gracePeriodSeconds, PropagationPolicy: &deletePolicy}); err != nil {
		log.Printf("Warning: Delete Subscriber Deployment falied: %v", err)
	}
	log.Println("Delete Subscriber service")
	if err := clientK8S.CoreV1().Services(*subscriberNamespace).Delete(util.SubscriberName,
		&metav1.DeleteOptions{GracePeriodSeconds: &gracePeriodSeconds}); err != nil {
		log.Printf("Warning: Delete Subscriber Service falied: %v", err)
	}
	if subClient != nil {
		log.Printf("Delete test subscription: %v\n", subscriptionName)
		if err := subClient.EventingV1alpha1().Subscriptions(*subscriberNamespace).Delete(subscriptionName,
			&metav1.DeleteOptions{PropagationPolicy: &deletePolicy}); err != nil {
			log.Printf("Warning: Delete Subscription falied: %v", err)
		}
	}
	if subClient != nil {
		log.Printf("Delete headers test subscription: %v\n", subscriptionName)
		if err := subClient.EventingV1alpha1().Subscriptions(*subscriberNamespace).Delete(headersSubscriptionName,
			&metav1.DeleteOptions{PropagationPolicy: &deletePolicy}); err != nil {
			log.Printf("Warning: Delete Subscription falied: %v", err)
		}
	}
	if eaClient != nil {
		log.Printf("Delete test event activation: %v\n", eventActivationName)
		if err := eaClient.ApplicationconnectorV1alpha1().EventActivations(*subscriberNamespace).Delete(eventActivationName, &metav1.DeleteOptions{PropagationPolicy: &deletePolicy}); err != nil {
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

func publishTestEvent(publishEventURL string) (*api.PublishResponse, error) {
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

func publishHeadersTestEvent(publishEventURL string) (*api.PublishResponse, error) {
	payload := fmt.Sprintf(
		`{"source-id": "%s","event-type":"%s","event-type-version":"%s","event-time":"2018-11-02T22:08:41+00:00","data":"headers-test-event"}`, srcID, eventType, eventTypeVersion)
	log.Printf("event to be published: %v\n", payload)

	client := &http.Client{}
	req, err := http.NewRequest("POST", publishEventURL, strings.NewReader(payload))
	req.Header.Add(sourceIDHeader, ceSourceIDHeaderValue)
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

func checkSubscriberReceivedEvent() error {
	for i := 0; i < retries; i++ {
		time.Sleep(time.Duration(i) * time.Second)
		log.Printf("Get subscriber response (%d/%d)\n", i, retries)
		res, err := http.Get(*subscriberResultsEndpointURL)
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

func checkSubscriberReceivedEventHeaders() error {
	for i := 0; i < retries; i++ {
		time.Sleep(time.Duration(i) * time.Second)
		log.Printf("Get subscriber 3 response (%d/%d)\n", i, retries)
		res, err := http.Get(*subscriber3ResultsEndpointURL)
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

		if resp[sourceIDHeader][0] != srcID {
			return fmt.Errorf("wrong response: %s, want: %s", resp[sourceIDHeader][0], srcID)
		}
		if resp[eventTypeHeader][0] != eventType {
			return fmt.Errorf("wrong response: %s, want: %s", resp[eventTypeHeader][0], eventType)
		}
		if resp[eventTypeVersionHeader][0] != eventTypeVersion {
			return fmt.Errorf("wrong response: %s, want: %s", resp[eventTypeVersionHeader][0], eventTypeVersion)
		}
		if resp[customHeader][0] != customHeaderValue {
			return fmt.Errorf("wrong response: %s, want: %s", resp[customHeader][0], customHeaderValue)
		}
		return nil
	}
	return errors.New("timeout for subscriber response")
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

func createSubscriber(subscriberName string, subscriberNamespace string, sbscrImg string) error {
	if _, err := clientK8S.AppsV1().Deployments(subscriberNamespace).Get(subscriberName, metav1.GetOptions{}); err != nil {
		log.Println("Create Subscriber deployment")
		if _, err := clientK8S.AppsV1().Deployments(subscriberNamespace).Create(util.NewSubscriberDeployment(sbscrImg)); err != nil {
			log.Printf("Create Subscriber deployment: %v\n", err)
			return err
		}
		log.Println("Create Subscriber service")
		if _, err := clientK8S.CoreV1().Services(subscriberNamespace).Create(util.NewSubscriberService()); err != nil {
			log.Printf("Create Subscriber service failed: %v\n", err)
			return err
		}
		time.Sleep(30 * time.Second)

		for i := 0; i < 60; i++ {
			var podReady bool
			if pods, err := clientK8S.CoreV1().Pods(subscriberNamespace).List(metav1.ListOptions{LabelSelector: "app=" + util.SubscriberName}); err != nil {
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

func createSubscription(subscriberNamespace string, subName string, subscriberEventEndpointURL string) bool {
	_, err := subClient.EventingV1alpha1().Subscriptions(subscriberNamespace).Create(util.NewSubscription(subName, subscriberNamespace, subscriberEventEndpointURL, eventType, "v1", srcID))
	if err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			log.Printf("Error in creating subscription: %v\n", err)
			return false
		}
	}
	return true
}

func checkSubscriberStatus(noOfRetries int) bool {
	var subscriberOK bool
	for i := 0; i < noOfRetries; i++ {
		if res, err := http.Get(*subscriberStatusEndpointURL); err != nil {
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

func checkSubscriber3Status(noOfRetries int) bool {
	var subscriberOK bool
	for i := 0; i < noOfRetries; i++ {
		if res, err := http.Get(*subscriber3StatusEndpointURL); err != nil {
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

func checkPublisherStatus(noOfRetries int) bool {
	var publishOK bool
	for i := 0; i < noOfRetries; i++ {
		if err := checkPublishStatus(*publishStatusEndpointURL); err != nil {
			log.Printf("Publisher not ready: %v", err)
			time.Sleep(time.Duration(i) * time.Second)
		} else {
			publishOK = true
			break
		}
	}
	return publishOK
}

func checkSubscriptionReady(subscriberNamespace string, subscriptionName string, noOfRetries int) bool {
	var isReady bool
	activatedCondition := subApis.SubscriptionCondition{Type: subApis.Ready, Status: subApis.ConditionTrue}
	for i := 0; i < noOfRetries && !isReady; i++ {
		kySub, err := subClient.EventingV1alpha1().Subscriptions(subscriberNamespace).Get(subscriptionName, metav1.GetOptions{})
		if err != nil {
			log.Printf("Cannot get Kyma subscription, name: %v; namespace: %v", subscriptionName, subscriberNamespace)
			break
		} else {
			if isReady = kySub.HasCondition(activatedCondition); !isReady {
				time.Sleep(1 * time.Second)
			}
		}
	}
	return isReady
}
