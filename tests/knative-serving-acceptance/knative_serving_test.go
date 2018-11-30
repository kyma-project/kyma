package knative_serving_acceptance

import (
	"fmt"
	"github.com/avast/retry-go"
	"github.com/kyma-project/kyma/tests/tools/ingressgateway"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
)

func TestKnativeServing_Acceptance(t *testing.T) {
	domainName := os.Getenv("DOMAIN_NAME")
	if domainName == "" {
		t.Fatal("Missing 'DOMAIN_NAME' variable")
	}

	testServiceURL := fmt.Sprintf("https://test-service.kantive-serving.%s", domainName)

	ingressClient, err := ingressgateway.Client()
	if err != nil {
		t.Fatalf("Unexpected error when creating ingressgateway client: %s", err)
	}

	err = retry.Do(func() error {
		resp, err := ingressClient.Get(testServiceURL)
		if err != nil {
			return err
		}

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("unexpected status code: %v", resp.StatusCode)
		}

		bytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		msg := string(bytes)
		if msg != "Test Target" {
			return fmt.Errorf("unexpected response: '%s'", msg)
		}

		return nil
	})

	if err != nil {
		t.Fatalf("cannot get test service response: %s", err)
	}
}
