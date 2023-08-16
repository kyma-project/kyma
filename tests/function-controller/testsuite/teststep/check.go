package teststep

import (
	"context"
	"encoding/json"
	goerrors "errors"
	"fmt"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
	"io"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"log"
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
	resourceLabel     = "serverless.kyma-project.io/resource"
	functionNameLabel = "serverless.kyma-project.io/function-name"
	manageByLabel     = "serverless.kyma-project.io/managed-by"
	uuidLabel         = "serverless.kyma-project.io/uuid"
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

func (d APIGatewayFunctionCheck) Cleanup() error {
	return nil
}

func (d APIGatewayFunctionCheck) OnError() error {
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

/*
*
https://github.com/cloudevents/spec/blob/v1.0.2/cloudevents/spec.md
*/
type cloudEventResponse struct {
	// Required
	CeType string `json:"ce-type"`
	// Required
	CeSource string `json:"ce-source"`
	// Required
	CeSpecVersion string `json:"ce-specversion"`
	// Required
	CeID string `json:"ce-id"`
	// Optional
	CeTime string `json:"ce-time"`
	// Optional
	CeDataContentType string `json:"ce-datacontenttype"`
	// Extension field, optional.
	CeEventTypeVersion string         `json:"ce-eventtypeversion"`
	Data               cloudEventData `json:"data"`
}

type cloudEventData struct {
	Hello string `json:"hello"`
}

var _ step.Step = &CloudEventCheck{}

type CloudEventCheck struct {
	//TODO: review and refactor
	name     string
	log      *logrus.Entry
	endpoint string
	encoding cloudevents.Encoding
}

func NewCloudEventCheck(log *logrus.Entry, name string, encoding cloudevents.Encoding, target *url.URL) *CloudEventCheck {
	return &CloudEventCheck{
		encoding: encoding,
		name:     name,
		log:      log.WithField(step.LogStepKey, name),
		endpoint: target.String(),
	}
}

func (ce CloudEventCheck) Name() string {
	return ce.name
}

func (ce CloudEventCheck) Run() error {
	c, err := cloudevents.NewClientHTTP()
	if err != nil {
		return errors.Wrap(err, "while creating cloud event client")
	}

	event := cloudevents.NewEvent()

	data := cloudEventData{}
	ctx := cloudevents.ContextWithTarget(context.Background(), ce.endpoint)
	switch ce.encoding {
	case cloudevents.EncodingStructured:
		ctx = cloudevents.WithEncodingStructured(ctx)
		data.Hello = "Structured"
	case cloudevents.EncodingBinary:
		data.Hello = "Binary"
		ctx = cloudevents.WithEncodingBinary(ctx)
	default:
		return errors.Errorf("Encoding not supported: %s", ce.encoding)
	}

	err = event.SetData(cloudevents.ApplicationJSON, data)
	if err != nil {
		return errors.Wrap(err, "while setting cloud event data")
	}
	expResp := cloudEventResponse{
		CeType:             fmt.Sprintf("test-%s", ce.encoding),
		CeSource:           "contract-test",
		CeSpecVersion:      cloudevents.VersionV1,
		CeEventTypeVersion: "v1alpha2",
		CeDataContentType:  "",
		Data:               data,
	}
	event.SetSource(expResp.CeSource)
	event.SetType(expResp.CeType)
	event.SetSpecVersion(expResp.CeSpecVersion)
	event.SetExtension("eventtypeversion", expResp.CeEventTypeVersion)

	result := c.Send(ctx, event)
	if cloudevents.IsUndelivered(result) {
		return errors.Wrap(result, "while sending cloud event")
	}
	log.Printf("sent: %v", event)
	log.Printf("result: %v", result)

	ceResp, err := ce.getCloudEventFromFunction()
	if err != nil {
		return errors.Wrap(err, "while fetching cloud event from function")
	}
	err = ce.assertResponse(ceResp, expResp)
	if err != nil {
		return errors.Wrapf(err, "while validating cloud event: %s", out)
	}
	ce.log.Info("cloud event is okay")
	return nil
}

func (ce CloudEventCheck) getCloudEventFromFunction() (cloudEventResponse, error) {
	req := &http.Request{}
	fnURL, err := url.Parse(ce.endpoint)
	if err != nil {
		return cloudEventResponse{}, errors.Wrap(err, "while parsing function url")
	}
	req.URL = fnURL

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return cloudEventResponse{}, errors.Wrap(err, "while doing GET request to function")
	}
	out, err := io.ReadAll(res.Body)
	if err != nil {
		return cloudEventResponse{}, errors.Wrap(err, "while reading response body")
	}
	fmt.Println("GET response:\n", string(out))

	ceResp := cloudEventResponse{}
	err = json.Unmarshal(out, &ceResp)
	if err != nil {
		return cloudEventResponse{}, errors.Wrap(err, "while unmarshalling response")
	}
	return ceResp, nil
}

func (ce CloudEventCheck) Cleanup() error {
	return nil
}

func (ce CloudEventCheck) OnError() error {
	return nil
}

func (ce CloudEventCheck) assertResponse(response cloudEventResponse, expectedResponse cloudEventResponse) error {
	var errJoined error

	if expectedResponse.Data.Hello != response.Data.Hello {
		err := errors.Errorf("Expected %s, got %s in cloud event data", expectedResponse.Data.Hello, response.Data.Hello)
		errJoined = goerrors.Join(err)
	}

	_, err := time.Parse(time.RFC3339, response.CeTime)
	if err != nil {
		errJoined = goerrors.Join(errors.Wrap(err, "while validating date"))
	}

	if response.CeSource != expectedResponse.CeSource {
		errJoined = goerrors.Join(errors.Errorf("expected source %s, got: %s", expectedResponse.CeSource, response.CeSource))
	}

	if response.CeType != expectedResponse.CeType {
		errJoined = goerrors.Join(errors.Errorf("expected type %s, got: %s", expectedResponse.CeType, response.CeType))
	}

	if response.CeSpecVersion != expectedResponse.CeSpecVersion {
		errJoined = goerrors.Join(errors.Errorf("expected spec version %s, got: %s", expectedResponse.CeSpecVersion, response.CeSpecVersion))
	}

	if response.CeEventTypeVersion != expectedResponse.CeEventTypeVersion {
		errJoined = goerrors.Join(errors.Errorf("expected event type version %s, got: %s", expectedResponse.CeEventTypeVersion, response.CeEventTypeVersion))
	}

	_, err = uuid.Parse(response.CeID)
	if err != nil {
		errJoined = goerrors.Join(errors.Errorf("expected UUID, got: %s", response.CeID))
	}
	return errJoined
}
