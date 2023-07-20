package teststep

import (
	"encoding/json"
	"io"
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

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	out, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	trResponse := tracingResponse{}
	err = json.Unmarshal(out, &trResponse)
	if err != nil {
		return err
	}

	err = t.assertTracingResponse(trResponse)
	if err != nil {
		return err
	}
	t.log.Info("headers are okay")
	return nil
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
