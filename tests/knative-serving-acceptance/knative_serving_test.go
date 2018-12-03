package knative_serving_acceptance

import (
	"fmt"
	"github.com/avast/retry-go"
	"github.com/kyma-project/kyma/tests/tools/ingressgateway"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestKnativeServing_Acceptance(t *testing.T) {
	domainName := MustGetenv(t, "DOMAIN_NAME")
	target := MustGetenv(t, "TARGET")

	testServiceURL := fmt.Sprintf("https://test-service.knative-serving.%s", domainName)

	ingressClient, err := ingressgateway.Client()
	if err != nil {
		t.Fatalf("Unexpected error when creating ingressgateway client: %s", err)
	}

	err = retry.Do(func() error {
		t.Logf("Calling: %s", testServiceURL)
		resp, err := ingressClient.Get(testServiceURL)
		if err != nil {
			return err
		}

		bytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		msg := string(bytes)

		log.Println("Received %v: '%s'", resp.StatusCode, msg)

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("unexpected status code: %v", resp.StatusCode)
		}
		if msg != target {
			return fmt.Errorf("unexpected response: '%s'", msg)
		}

		return nil
	},
		retry.Attempts(10),
		retry.Delay(5*time.Second),
		retry.OnRetry(func(n uint, err error) {
			log.Printf("[%v] try failed: %s", n, err)
		}),
	)

	if err != nil {
		t.Fatalf("cannot get test service response: %s", err)
	}
}

func MustGetenv(t *testing.T, name string) string {
	env := os.Getenv(name)
	if env == "" {
		t.Fatalf("Missing '%s' variable", name)
	}
	return env
}
