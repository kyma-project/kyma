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
	eaClientSet "github.com/kyma-project/kyma/components/event-bus/generated/ea/clientset/versioned"
	subscriptionClientSet "github.com/kyma-project/kyma/components/event-bus/generated/push/clientset/versioned"
	"github.com/kyma-project/kyma/components/event-bus/test/util"
	_ "github.com/satori/go.uuid"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	eventType           = "test-e2e-event-bus"
	testerName          = "test-core-event-bus-tester"
	subscriptionName    = "test-sub"
	eventActivationName = "test-ea"
	srcNamespace        = "local.kyma.commerce"
	srcType             = "commerce"
	srcEnv              = "test"

	SUCCESS = 0
	FAIL    = 1
	RETRIES = 20
)

var (
	subscriberEventEndpointURL    *string
	subscriberResultsEndpointURL  *string
	subscriberStatusEndpointURL   *string
	subscriberShutdownEndpointURL *string
	clientK8S                     *kubernetes.Clientset
	eaClient                      *eaClientSet.Clientset
	subClient                     *subscriptionClientSet.Clientset
	namespace                     *string
	publishEventEndpointURL       *string
	publishStatusEndpointURL      *string
)

func main() {
	flags := flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	publishEventEndpointURL = flags.String("publish-event-uri", "http://core-publish:8080/v1/events", "publish service events endpoint `URL`")
	publishStatusEndpointURL = flags.String("publish-status-uri", "http://core-publish:8080/v1/status/ready", "publish service status endpoint `URL`")
	namespace = flags.String("ns", "kyma-system", "k8s `namespace` in which test app is running")
	subscriberImage := flags.String("subscriber-image", "", "subscriber Docker `image` name")
	subscriberEventEndpointURL = flags.String("subscriber-events-uri", "http://"+util.SubscriberName+":9000/v1/events", "subscriber service events endpoint `URL`")
	subscriberResultsEndpointURL = flags.String("subscriber-results-uri", "http://"+util.SubscriberName+":9000/v1/results", "subscriber service results endpoint `URL`")
	subscriberStatusEndpointURL = flags.String("subscriber-status-uri", "http://"+util.SubscriberName+":9000/v1/status", "subscriber service status endpoint `URL`")
	subscriberShutdownEndpointURL = flags.String("subscriber-shutdown-uri", "http://"+util.SubscriberName+":9000/v1/shutdown", "subscriber service shutdown endpoint `URL`")

	flags.Parse(os.Args[1:])

	if flags.NFlag() == 0 || *subscriberImage == "" {

		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flags.PrintDefaults()
		os.Exit(1)
	}

	config, err := rest.InClusterConfig()
	if err != nil {
		log.Printf("Error in getting cluster config: %v\n", err)
		shutdown(FAIL)
	}

	log.Println("Create the clientK8S")
	clientK8S, err = kubernetes.NewForConfig(config)
	if err != nil {
		log.Printf("Failed to create a ClientSet: %v\n", err)
		shutdown(FAIL)
	}

	log.Println("Create an event activation")
	eaClient, err = eaClientSet.NewForConfig(config)
	if err != nil {
		log.Printf("Error in creating event activation client: %v\n", err)
		shutdown(FAIL)
	}

	if !createEventActivation(namespace, RETRIES) {
		log.Println("Error: Cannot create the event activation")
		shutdown(FAIL)
	}
	time.Sleep(5 * time.Second)

	log.Println("Create a subscription")
	subClient, err = subscriptionClientSet.NewForConfig(config)
	if err != nil {
		log.Printf("Error in creating subscription client: %v\n", err)
		shutdown(FAIL)
	}
	if _, err = subClient.EventingV1alpha1().Subscriptions(*namespace).Create(util.NewSubscription(
		subscriptionName,
		*namespace,
		*subscriberEventEndpointURL,
		eventType,
		"v1",
		srcEnv,
		srcNamespace,
		srcType)); err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			log.Printf("Error in creating subscription: %v\n", err)
			shutdown(FAIL)
		}
	}
	time.Sleep(5 * time.Second)

	log.Println("Create Subscriber")
	if err := createSubscriber(namespace, util.SubscriberName, *subscriberImage); err != nil {
		log.Printf("Create Subscriber failed: %v\n", err)
	}

	log.Println("Check Subscriber Status")
	if !checkSubscriberStatus(RETRIES) {
		log.Println("Error: Cannot connect to Subscriber")
		shutdown(FAIL)
	}

	log.Println("Check Publisher Status")
	if !checkPublisherStatus(RETRIES) {
		log.Println("Error: Cannot connect to Publisher")
		shutdown(FAIL)
	}

	log.Println("Publish an event")

	var eventSent bool
	for i := 0; i < RETRIES; i++ {
		if _, err := publishTestEvent(*publishEventEndpointURL); err != nil {
			log.Printf("Publish event failed: %v; Retrying (%d/%d)", err, i, RETRIES)
			time.Sleep(time.Duration(i) * time.Second)
		} else {
			eventSent = true
			break
		}
	}
	if !eventSent {
		log.Println("Error: Cannot send test event")
		shutdown(FAIL)
	}

	log.Println("Try to read the response from subscriber server")
	if err := checkSubscriberReceivedEvent(); err != nil {
		log.Printf("Error: Cannot get the test event from subscriber: %v\n", err)
		shutdown(FAIL)
	}

	log.Println("Successfully finished")
	shutdown(SUCCESS)
}

func shutdown(code int) {
	log.Println("Send shutdown request to Subscriber")
	if _, err := http.Post(*subscriberShutdownEndpointURL, "application/json", strings.NewReader(`{"shutdown": "true"}`)); err != nil {
		log.Printf("Warning: Shutdown Subscriber falied: %v", err)
	}
	log.Println("Delete Subscriber deployment")
	deletePolicy := metav1.DeletePropagationForeground
	gracePeriodSeconds := int64(0)
	if err := clientK8S.AppsV1().Deployments(*namespace).Delete(util.SubscriberName,
		&metav1.DeleteOptions{GracePeriodSeconds: &gracePeriodSeconds, PropagationPolicy: &deletePolicy}); err != nil {
		log.Printf("Warning: Delete Subscriber Deployment falied: %v", err)
	}
	log.Println("Delete Subscriber service")
	if err := clientK8S.CoreV1().Services(*namespace).Delete(util.SubscriberName,
		&metav1.DeleteOptions{GracePeriodSeconds: &gracePeriodSeconds}); err != nil {
		log.Printf("Warning: Delete Subscriber Service falied: %v", err)
	}
	if subClient != nil {
		log.Printf("Delete test subscription: %v\n", subscriptionName)
		if err := subClient.EventingV1alpha1().Subscriptions(*namespace).Delete(subscriptionName, &metav1.DeleteOptions{PropagationPolicy: &deletePolicy}); err != nil {
			log.Printf("Warning: Delete Subscription falied: %v", err)
		}
	}
	if eaClient != nil {
		log.Printf("Delete test event activation: %v\n", eventActivationName)
		if err := eaClient.RemoteenvironmentV1alpha1().EventActivations(*namespace).Delete(eventActivationName, &metav1.DeleteOptions{PropagationPolicy: &deletePolicy}); err != nil {
			log.Printf("Warning: Delete Event Activation falied: %v", err)
		}
	}
	//needed if Istio injection will be enabled for tester cause the kyma script which runs the tests checks the pod status
	/*
		if _, err := clientK8S.CoreV1().Pods("kyma-system").Get(testerName, metav1.GetOptions{}); err != nil {
			log.Printf("Cannot get my pod: %v", err)
		} else {
			if code == SUCCESS {
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
		`{"source": {"source-namespace": "%s","source-type": "%s","source-environment": "%s"},
	"event-type": "%s","event-type-version": "v1","event-time": "2018-11-02T22:08:41+00:00","data": "test-event-1"}`,
		srcNamespace, srcType, srcEnv, eventType)
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

func checkSubscriberReceivedEvent() error {

	for i := 0; i < RETRIES; i++ {
		time.Sleep(time.Duration(i) * time.Second)
		log.Printf("Get subscriber response (%d/%d)\n", i, RETRIES)
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

func createSubscriber(namespace *string, subscriberName string, sbscrImg string) error {
	if _, err := clientK8S.AppsV1().Deployments(*namespace).Get(subscriberName, metav1.GetOptions{}); err != nil {
		log.Println("Create Subscriber deployment")
		if _, err := clientK8S.AppsV1().Deployments(*namespace).Create(util.NewSubscriberDeployment(sbscrImg)); err != nil {
			log.Printf("Create Subscriber deployment: %v\n", err)
			return err
		}
		time.Sleep(30 * time.Second)

		for i := 0; i < 60; i++ {
			var podReady bool
			if pods, err := clientK8S.CoreV1().Pods(*namespace).List(metav1.ListOptions{LabelSelector: "app=" + util.SubscriberName}); err != nil {
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

		log.Println("Create Subscriber service")
		if _, err := clientK8S.CoreV1().Services(*namespace).Create(util.NewSubscriberService()); err != nil {
			log.Printf("Create Subscriber service failed: %v\n", err)
		}
		time.Sleep(30 * time.Second)

		log.Println("Subscriber recreated")
	}
	return nil
}

func createEventActivation(namespace *string, noOfRetries int) bool {
	var eventActivationOK bool
	var err error
	for i := 0; i < noOfRetries; i++ {
		if _, err = eaClient.RemoteenvironmentV1alpha1().EventActivations(*namespace).Create(util.NewEventActivation(
			eventActivationName,
			*namespace,
			srcEnv,
			srcNamespace,
			srcType)); err == nil {
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

func checkSubscriberStatus(noOfRetries int) bool {
	var subscriberOK bool
	for i := 0; i < noOfRetries; i++ {
		if res, err := http.Get(*subscriberStatusEndpointURL); err != nil {
			log.Printf("Subscriber Status request failed: %v; Retrying (%d/%d)", err, i, noOfRetries)
			time.Sleep(time.Duration(i) * time.Second)
		} else if !checkStatusCode(res, http.StatusOK) {
			log.Printf("Subscriber Server Status request returns: %v; Retrying (%d/%d)\n", res, i, noOfRetries)
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
