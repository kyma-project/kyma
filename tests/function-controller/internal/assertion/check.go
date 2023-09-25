package assertion

import (
	"context"
	"encoding/json"
	goerrors "errors"
	"fmt"
	"github.com/kyma-project/kyma/tests/function-controller/internal/executor"
	"github.com/kyma-project/kyma/tests/function-controller/internal/resources/function"
	"github.com/kyma-project/kyma/tests/function-controller/internal/utils"
	"io"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"
)

type HTTPCheck struct {
	name        string
	log         *logrus.Entry
	endpoint    string
	expectedMsg string
	poll        utils.Poller
}

var _ executor.Step = HTTPCheck{}

const (
	publisherURL = "http://eventing-event-publisher-proxy.kyma-system/publish"
)

func NewHTTPCheck(log *logrus.Entry, name string, url *url.URL, poller utils.Poller, expectedMsg string) *HTTPCheck {
	return &HTTPCheck{
		name:        name,
		log:         log.WithField(executor.LogStepKey, name),
		endpoint:    url.String(),
		expectedMsg: expectedMsg,
		poll:        poller.WithLogger(log),
	}

}

func (h HTTPCheck) Name() string {
	return h.name
}

func (h HTTPCheck) Run() error {
	// backoff is needed because even tough the deployment may be ready
	// the language specific server may not start yet
	// there may also be some problems with istio sidecars etc
	backoff := wait.Backoff{
		Steps:    10,
		Duration: 500 * time.Millisecond,
		Factor:   2.0,
		Jitter:   0.1,
	}
	return retry.OnError(backoff, func(err error) bool {
		return true
	}, func() error {
		err := errors.Wrap(h.poll.PollForAnswer(h.endpoint, "", h.expectedMsg), "while checking connection to function")
		if err != nil {
			h.log.Warn(err)
		}
		return err
	})
}

func (h HTTPCheck) Cleanup() error {
	return nil
}

func (h HTTPCheck) OnError() error {
	return nil
}

var _ executor.Step = &tracingHTTPCheck{}

type tracingHTTPCheck struct {
	name     string
	log      *logrus.Entry
	endpoint string
	poll     utils.Poller
}

type tracingResponse struct {
	TraceParent string `json:"traceparent"`
	TraceID     string `json:"x-b3-traceid"`
	SpanID      string `json:"x-b3-spanid"`
}

func TracingHTTPCheck(log *logrus.Entry, name string, url *url.URL, poller utils.Poller) *tracingHTTPCheck {
	return &tracingHTTPCheck{
		name:     name,
		log:      log.WithField(executor.LogStepKey, name),
		endpoint: url.String(),
		poll:     poller.WithLogger(log),
	}

}

func (t tracingHTTPCheck) Name() string {
	return t.name
}

func (t tracingHTTPCheck) Run() error {
	req, err := http.NewRequest(http.MethodGet, t.endpoint, nil)
	if err != nil {
		return err
	}

	req.Header.Add("X-B3-Sampled", "1")

	resp, err := t.doRetrievableHttpCall(req, 5)
	if err != nil {
		return errors.Wrap(err, "while doing http call")
	}

	out, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "while reading response body")
	}

	trResponse := tracingResponse{}
	err = json.Unmarshal(out, &trResponse)
	if err != nil {
		return errors.Wrapf(err, "while unmarshalling response to json")
	}

	err = t.assertTracingResponse(trResponse)
	if err != nil {
		return errors.Wrapf(err, "Got following headers: %s", out)
	}
	t.log.Info("headers are okay")
	return nil
}

func (t tracingHTTPCheck) doRetrievableHttpCall(req *http.Request, retries int) (*http.Response, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	var finalResp *http.Response = nil

	var backoff = wait.Backoff{
		Steps:    retries,
		Duration: 2 * time.Second,
		Factor:   1.0,
	}

	err := retry.OnError(backoff, func(err error) bool {
		t.log.Warnf("Got error: %s", err.Error())
		return true
	}, func() error {
		resp, err := client.Do(req)
		if err != nil {
			return err
		}

		if resp.StatusCode != http.StatusOK {
			return errors.Errorf(" expected status code: %d, got: %d", http.StatusOK, resp.StatusCode)
		}
		finalResp = resp
		return nil
	})
	return finalResp, errors.Wrap(err, "while trying to call function")
}

func (t tracingHTTPCheck) Cleanup() error {
	return nil
}

func (t tracingHTTPCheck) OnError() error {
	return nil
}

func (t tracingHTTPCheck) assertTracingResponse(response tracingResponse) error {
	if response.TraceID == "" {
		return errors.New("No trace ID")
	}
	if response.TraceParent == "" {
		return errors.New("No TraceParent")
	}
	if response.SpanID == "" {
		return errors.New("No span ID")
	}

	return nil
}

const (
	resourceLabel     = "serverless.kyma-project.io/resource"
	functionNameLabel = "serverless.kyma-project.io/function-name"
	manageByLabel     = "serverless.kyma-project.io/managed-by"
	uuidLabel         = "serverless.kyma-project.io/uuid"
)

type apiGatewayFunctionCheck struct {
	name      string
	fn        *function.Function
	client    *v1.CoreV1Client
	namespace string
	runtime   string
}

func APIGatewayFunctionCheck(name string, fn *function.Function, coreV1 *v1.CoreV1Client, ns string, rt string) *apiGatewayFunctionCheck {
	return &apiGatewayFunctionCheck{
		name:      name,
		fn:        fn,
		client:    coreV1,
		namespace: ns,
		runtime:   rt,
	}
}

func (d apiGatewayFunctionCheck) Name() string {
	return d.name
}

func (d apiGatewayFunctionCheck) Run() error {

	svc, err := d.client.Services(d.namespace).Get(context.Background(), d.name, metav1.GetOptions{})
	if err != nil {
		return errors.Wrap(err, "while trying to get service")
	}

	pod, err := d.client.Pods(d.namespace).List(context.Background(), metav1.ListOptions{LabelSelector: fmt.Sprintf("%s=deployment,%s=%s", resourceLabel, functionNameLabel, d.name)})
	if err != nil {
		return errors.Wrap(err, "while trying to get pod")
	}

	err = checkIfRequiredLabelsExists(svc.Spec.Selector, true)
	if err != nil {
		return errors.Wrap(err, " while checking the service labels")
	}
	err = checkIfRequiredLabelsExists(pod.Items[0].ObjectMeta.Labels, false)
	if err != nil {
		return errors.Wrap(err, " while checking the pod labels")
	}

	err = checkIfContractIsFulfilled(pod.Items[0], *svc)
	if err != nil {
		return errors.Wrap(err, " while checking labels")
	}

	if len(svc.Spec.Selector) != 0 {
		return errors.New("The labels are not matching")
	}

	return nil
}

func (d apiGatewayFunctionCheck) Cleanup() error {
	return nil
}

func (d apiGatewayFunctionCheck) OnError() error {
	return nil
}

func checkIfContractIsFulfilled(pod corev1.Pod, service corev1.Service) error {
	var errJoined error

	for k, v := range pod.Labels {
		if val, exists := service.Spec.Selector[k]; exists {
			if val == v {
				delete(service.Spec.Selector, k)
			} else {
				err := errors.Errorf("Expected %s but got %s", v, val)
				errJoined = goerrors.Join(err)
			}
		}
	}
	return errJoined
}

func checkIfRequiredLabelsExists(labels map[string]string, isService bool) error {
	requiredLabels := []string{resourceLabel, functionNameLabel, manageByLabel, uuidLabel}

	if isService {
		if len(labels) != 4 {
			return errors.New(fmt.Sprintf("Service has got %d instead of 4 labels", len(labels)))
		}
	}

	var errJoined error

	for _, label := range requiredLabels {
		if _, exists := labels[label]; !exists {
			err := errors.New(fmt.Sprintf("Label %s is missing", label))
			errJoined = goerrors.Join(err)
		}
	}
	return errJoined
}
