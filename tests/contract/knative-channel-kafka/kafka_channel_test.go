package knative_eventing_kafka_channel

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/client"
	"knative.dev/pkg/apis"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	// allow client authentication against GKE clusters
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	"github.com/avast/retry-go"
	kafkav1alpha1 "github.com/kyma-incubator/knative-kafka/components/controller/pkg/apis/knativekafka/v1alpha1"
	kafkaclientset "github.com/kyma-incubator/knative-kafka/components/controller/pkg/client/clientset/versioned/typed/knativekafka/v1alpha1"
)

var (
	// interrupt signals to be handled for graceful cleanup.
	interruptSignals = []os.Signal{syscall.SIGTERM, syscall.SIGINT}
)

// TestKnativeEventingKafkaChannelAcceptance creates a test Kafka channel and asserts its status to be ready.
func TestKnativeEventingKafkaChannelAcceptance(t *testing.T) {
	// test meta for the Kafka channel
	name := "test-kafka-channel"
	namespace := "knative-eventing"

	// load cluster config
	config := loadConfigOrDie(t)

	// prepare a Kafka client
	kafkaClient := kafkaclientset.NewForConfigOrDie(config).KafkaChannels(namespace)

	// cleanup test resources gracefully when an interrupt signal is received
	interrupted := cleanupOnInterrupt(t, interruptSignals, func() { deleteChannel(t, kafkaClient, name) })
	defer close(interrupted)

	// cleanup the Kafka channel when the test is finished
	defer deleteChannel(t, kafkaClient, name)

	// delete the Kafka channel if existed before to make sure that
	// the new channel to be created has the correct structure and data
	if err := deleteChannelIfExistsAndWaitUntilDeleted(t, interrupted, kafkaClient, name, 5*time.Second, 10, retry.FixedDelay); err != nil {
		log.Printf("test failed with error: %s", err)
		t.FailNow()
	}

	// create a Kafka channel
	if _, err := kafkaClient.Create(newKafkaChannel(name, namespace)); err != nil {
		log.Printf("cannot create a Kafka channel: %s: error: %v", name, err)
		t.FailNow()
	} else {
		log.Printf("created Kafka channel: %s", name)
	}

	// assert the Kafka channel status to be ready
	readyKafkaChannel, err := checkChannelReadyWithRetry(t, interrupted, kafkaClient, name, 5*time.Second, 10, retry.FixedDelay)
	if err != nil {
		log.Printf("test failed with error: %s", err)
		t.FailNow()
	}

	target := readyKafkaChannel.Status.Address.URL
	ceClient := createCloudEventsClient(t, target)
	event := createCloudEvent(t)

	err = retry.Do(func() error {
		// send an CE event to Kafka channel
		log.Printf("sending cloudevent to Kafka channel: %q", target)
		rctx, _, err := ceClient.Send(context.Background(), event)
		if err != nil {
			return err
		}
		rtctx := cloudevents.HTTPTransportContextFrom(rctx)
		log.Printf("received status code: %d", rtctx.StatusCode)
		if !is2XXStatusCode(rtctx.StatusCode) {
			return fmt.Errorf("received non 2xx status code: %d", rtctx.StatusCode)
		}
		return nil
	}, retry.Attempts(24), retry.Delay(time.Second*5), // 120=24*5 seconds
		retry.OnRetry(func(n uint, err error) { log.Printf("[%v] try failed: %s", n, err) }),
	)

	if err != nil {
		log.Printf("could not send cloudevent %+v to %q: %v", event, target, err)
		t.FailNow()
	}

	log.Printf("test finished successfully")
}

// is2XXStatusCode checks whether status code is a 2XX status code
func is2XXStatusCode(statusCode int) bool {
	return statusCode >= http.StatusOK && statusCode < http.StatusMultipleChoices
}

func createCloudEvent(t *testing.T) cloudevents.Event {
	event := cloudevents.NewEvent()
	data := "hello kafka"
	if err := event.SetData(data); err != nil {
		log.Printf("could not set cloudevent data %q: %v", data, err)
		t.FailNow()
	}
	event.SetType("com.example.testing")
	event.SetID("A234-1234-1234")
	event.SetSource("kafka channel test")
	return event
}

func createCloudEventsClient(t *testing.T, target *apis.URL) client.Client {
	transport, err := cloudevents.NewHTTPTransport(
		cloudevents.WithTarget(target.String()),
	)
	if err != nil {
		log.Printf("could not create cloudevents http transport: %v", err)
		t.FailNow()
	}
	ceClient, err := cloudevents.NewClient(transport)
	if err != nil {
		log.Printf("could not create cloudevents client: %v", err)
		t.FailNow()
	}
	return ceClient
}

// loadConfigOrDie loads the cluster config or exits the test with failure if it did not succeed.
func loadConfigOrDie(t *testing.T) *rest.Config {
	t.Helper()

	if _, err := os.Stat(clientcmd.RecommendedHomeFile); os.IsNotExist(err) {
		config, err := rest.InClusterConfig()
		if err != nil {
			log.Printf("cannot create in-cluster config: %v", err)
			t.FailNow()
		}
		return config
	}

	config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		log.Printf("cannot read config: %s", err)
		t.FailNow()
	}
	return config
}

// newKafkaChannel returns a new instance of a Kafka channel type.
func newKafkaChannel(name, namespace string) *kafkav1alpha1.KafkaChannel {
	return &kafkav1alpha1.KafkaChannel{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
}

// deleteChannelIfExistsAndWaitUntilDeleted deletes the Kafka channel if it exists and waits until it is deleted.
func deleteChannelIfExistsAndWaitUntilDeleted(t *testing.T, interrupted chan bool,
	kafkaClient kafkaclientset.KafkaChannelInterface, name string,
	duration time.Duration, attempts uint, delayType retry.DelayTypeFunc) error {
	t.Helper()

	if _, err := kafkaClient.Get(name, v1.GetOptions{}); err == nil {
		log.Printf("delete the old Kafka channel: %s", name)
		deleteChannel(t, kafkaClient, name)

		// wait for the old Kafka channel to be deleted
		return retry.Do(func() error {
			select {
			case <-interrupted:
				log.Printf("cannot continue, test was interrupted")
				t.FailNow()
				return nil
			default:
				if _, err := kafkaClient.Get(name, v1.GetOptions{}); err == nil {
					return fmt.Errorf("the old Kafka channel is not deleted yet")
				}
				return nil
			}
		},
			retry.Delay(duration),
			retry.Attempts(attempts),
			retry.DelayType(delayType),
			retry.OnRetry(func(n uint, err error) { log.Printf("[%v] try failed: %s", n, err) }),
		)
	}

	return nil
}

// checkChannelReadyWithRetry gets the Kafka channel given its name and checks its status to be ready in a retry loop.
func checkChannelReadyWithRetry(t *testing.T, interrupted chan bool,
	kafkaClient kafkaclientset.KafkaChannelInterface, name string,
	duration time.Duration, attempts uint, delayType retry.DelayTypeFunc) (*kafkav1alpha1.KafkaChannel, error) {
	t.Helper()
	var kafkaChannel *kafkav1alpha1.KafkaChannel

	err := retry.Do(func() error {
		var err error
		select {
		case <-interrupted:
			log.Printf("cannot continue, test was interrupted")
			t.FailNow()
			return nil
		default:
			kafkaChannel, err = kafkaClient.Get(name, v1.GetOptions{})
			if err != nil {
				return err
			}
			log.Printf("found Kafka channel: %s with ready status: %v", name, kafkaChannel.Status.IsReady())
			if !kafkaChannel.Status.IsReady() {
				return fmt.Errorf("the Kafka channel is not ready")
			}
			return nil
		}
	},
		retry.Delay(duration),
		retry.Attempts(attempts),
		retry.DelayType(delayType),
		retry.OnRetry(func(n uint, err error) { log.Printf("[%v] try failed: %s", n, err) }),
	)

	return kafkaChannel, err
}

// deleteChannel deletes the Kafka channel given its name if it was not already deleted.
func deleteChannel(t *testing.T, kafkaClient kafkaclientset.KafkaChannelInterface, name string) {
	t.Helper()

	err := kafkaClient.Delete(name, &v1.DeleteOptions{})
	switch {
	case errors.IsGone(err):
	case errors.IsNotFound(err):
		log.Printf("tried to delete Kafka channel: %s but it was already deleted", name)
	case err != nil:
		log.Printf("cannot delete Kafka channel %v, Error: %v", name, err)
		t.FailNow()
	default:
		log.Printf("deleted Kafka channel: %s", name)
	}
}

// cleanupOnInterrupt executes the cleanup function in a goroutine if any of the interrupt signals was received.
func cleanupOnInterrupt(t *testing.T, interruptSignals []os.Signal, cleanup func()) chan bool {
	t.Helper()

	// to notify the callers if an interrupt signal was received
	interrupted := make(chan bool, 1)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, interruptSignals...)

	go func() {
		defer close(ch)
		<-ch
		interrupted <- true

		t.Log("interrupt signal received, cleanup started")
		cleanup()
		t.Log("cleanup finished")
	}()

	return interrupted
}
