package knative_eventing_kafka_channel

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

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

	// create a Kafka channel
	if _, err := kafkaClient.Create(newKafkaChannel(name, namespace)); err != nil {
		t.Fatalf("cannot create a Kafka channel: %s: error: %v", name, err)
	} else {
		t.Logf("created Kafka channel: %s", name)
	}

	// assert the Kafka channel status to be ready
	if err := checkChannelReadyWithRetry(t, interrupted, kafkaClient, name, 5*time.Second, 10, retry.FixedDelay); err != nil {
		t.Fatalf("test failed with error: %s", err)
	} else {
		t.Logf("test finished successfully")
	}

	// TODO(marcobebway) extend the test to assert event delivery also works using the Kafka channel.
}

// loadConfigOrDie loads the cluster config or exits the test with failure if it did not succeed.
func loadConfigOrDie(t *testing.T) *rest.Config {
	t.Helper()

	if _, err := os.Stat(clientcmd.RecommendedHomeFile); os.IsNotExist(err) {
		config, err := rest.InClusterConfig()
		if err != nil {
			t.Fatalf("cannot create in-cluster config: %v", err)
		}
		return config
	}

	config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		t.Fatalf("cannot read config: %s", err)
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

// checkChannelReadyWithRetry gets the Kafka channel given its name and checks its status to be ready in a retry loop.
func checkChannelReadyWithRetry(t *testing.T, interrupted chan bool,
	kafkaClient kafkaclientset.KafkaChannelInterface, name string,
	duration time.Duration, attempts uint, delayType retry.DelayTypeFunc) error {
	t.Helper()

	return retry.Do(func() error {
		select {
		case <-interrupted:
			t.Fatal("cannot continue, test was interrupted")
			return nil
		default:
			kafkaChannel, err := kafkaClient.Get(name, v1.GetOptions{})
			if err != nil {
				return err
			}
			t.Logf("found Kafka channel: %s with ready status: %v", name, kafkaChannel.Status.IsReady())
			if !kafkaChannel.Status.IsReady() {
				return fmt.Errorf("the Kafka channel is not ready")
			}
			return nil
		}
	},
		retry.Delay(duration),
		retry.Attempts(attempts),
		retry.DelayType(delayType),
		retry.OnRetry(func(n uint, err error) { t.Logf("[%v] try failed: %s", n, err) }),
	)
}

// deleteChannel deletes the Kafka channel given its name if it was not already deleted.
func deleteChannel(t *testing.T, kafkaClient kafkaclientset.KafkaChannelInterface, name string) {
	t.Helper()

	err := kafkaClient.Delete(name, &v1.DeleteOptions{})
	switch {
	case errors.IsNotFound(err):
		// the Kafka channel was already deleted
	case err != nil:
		t.Fatalf("cannot delete Kafka channel %v, Error: %v", name, err)
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

		t.Log("interrupt signal received, cleanup started")
		cleanup()
		t.Log("cleanup finished")

		interrupted <- true
	}()

	return interrupted
}
