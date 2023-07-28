package teststep

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/function"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/poller"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/step"
)

type HTTPCheck struct {
	name        string
	log         *logrus.Entry
	endpoint    string
	expectedMsg string
	poll        poller.Poller
}

var _ step.Step = HTTPCheck{}

const (
	publisherURL = "http://eventing-event-publisher-proxy.kyma-system/publish"
)

func NewHTTPCheck(log *logrus.Entry, name string, url *url.URL, poller poller.Poller, expectedMsg string) *HTTPCheck {
	return &HTTPCheck{
		name:        name,
		log:         log.WithField(step.LogStepKey, name),
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

var _ step.Step = DefaultedFunctionCheck{}

type DefaultedFunctionCheck struct {
	name string
	fn   *function.Function
}

func NewDefaultedFunctionCheck(name string, fn *function.Function) step.Step {
	return &DefaultedFunctionCheck{
		name: name,
		fn:   fn,
	}
}

func (d DefaultedFunctionCheck) Name() string {
	return d.name
}

func (d DefaultedFunctionCheck) Run() error {
	fn, err := d.fn.Get()
	if err != nil {
		return err
	}

	if fn == nil {
		return errors.New("function can't be nil")
	}

	spec := fn.Spec
	if spec.Replicas == nil {
		return errors.New("replicas equal nil")
	} else if spec.ResourceConfiguration.Function.Resources.Requests.Memory().IsZero() || spec.ResourceConfiguration.Function.Resources.Requests.Cpu().IsZero() {
		return errors.New("requests equal zero")
	} else if spec.ResourceConfiguration.Function.Resources.Limits.Memory().IsZero() || spec.ResourceConfiguration.Function.Resources.Limits.Cpu().IsZero() {
		return errors.New("limits equal zero")
	}
	return nil
}

func (d DefaultedFunctionCheck) Cleanup() error {
	return nil
}

func (d DefaultedFunctionCheck) OnError() error {
	return nil
}

var _ step.Step = &TracingHTTPCheck{}

type TracingHTTPCheck struct {
	name     string
	log      *logrus.Entry
	endpoint string
	poll     poller.Poller
}

type tracingResponse struct {
	TraceParent string `json:"traceparent"`
	TraceID     string `json:"x-b3-traceid"`
	SpanID      string `json:"x-b3-spanid"`
}

func NewTracingHTTPCheck(log *logrus.Entry, name string, url *url.URL, poller poller.Poller) *TracingHTTPCheck {
	return &TracingHTTPCheck{
		name:     name,
		log:      log.WithField(step.LogStepKey, name),
		endpoint: url.String(),
		poll:     poller.WithLogger(log),
	}

}

func (t TracingHTTPCheck) Name() string {
	return t.name
}

func (t TracingHTTPCheck) Run() error {
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

func (t TracingHTTPCheck) doRetrievableHttpCall(req *http.Request, retries int) (*http.Response, error) {
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

func (t TracingHTTPCheck) Cleanup() error {
	return nil
}

func (t TracingHTTPCheck) OnError() error {
	return nil
}

func (t TracingHTTPCheck) assertTracingResponse(response tracingResponse) error {
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
	resourceLabel     = "serverless.kyma-project.io/resource=deployment"
	functionNameLabel = "serverless.kyma-project.io/function-name="
)

type APIGatewayFunctionCheck struct {
	name      string
	fn        *function.Function
	client    *v1.CoreV1Client
	namespace string
	runtime   string
}

func NewAPIGatewayFunctionCheck(name string, fn *function.Function, coreV1 *v1.CoreV1Client, ns string, rt string) *APIGatewayFunctionCheck {
	return &APIGatewayFunctionCheck{
		name:      name,
		fn:        fn,
		client:    coreV1,
		namespace: ns,
		runtime:   rt,
	}
}

func (d APIGatewayFunctionCheck) Name() string {
	return d.name
}

func (d APIGatewayFunctionCheck) Run() error {

	svc, err := d.client.Services(d.namespace).Get(context.Background(), d.name, metav1.GetOptions{})
	if err != nil {
		return errors.Wrap(err, "while trying to get service")
	}

	pod, err := d.client.Pods(d.namespace).List(context.Background(), metav1.ListOptions{LabelSelector: fmt.Sprintf("%s,%s=%s", resourceLabel, functionNameLabel, d.name)})
	if err != nil {
		return errors.Wrap(err, "while trying to get pod")
	}

	for k, v := range pod.Items[0].ObjectMeta.Labels {
		if val, exists := svc.Spec.Selector[k]; exists {
			if val == v {
				delete(svc.Spec.Selector, k)
			} else {
				return errors.Errorf("Expected %s but got %s", v, val)
			}
		}
	}

	if len(svc.Spec.Selector) != 0 {
		return errors.New("The labels are not matching")
	}

	return nil
}

func (d APIGatewayFunctionCheck) Cleanup() error {
	return nil
}

func (d APIGatewayFunctionCheck) OnError() error {
	return nil
}
