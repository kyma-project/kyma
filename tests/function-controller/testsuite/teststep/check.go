package teststep

import (
	"fmt"
	"net/url"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/function"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/poller"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/step"
	"github.com/kyma-project/kyma/tests/function-controller/testsuite"
)

type HTTPCheck struct {
	name        string
	log         *logrus.Entry
	endpoint    string
	expectedMsg string
	poll        poller.Poller
}

var _ step.Step = HTTPCheck{}

func NewHTTPCheck(log *logrus.Entry, name string, url *url.URL, poller poller.Poller, expectedMsg string) *HTTPCheck {
	return &HTTPCheck{
		name:        name,
		log:         log.WithField(step.LogStepKey, name),
		endpoint:    url.String(),
		expectedMsg: expectedMsg,
		poll:        poller,
	}

}

func (h HTTPCheck) Name() string {
	return h.name
}

func (h HTTPCheck) Run() error {
	return errors.Wrap(h.poll.PollForAnswer(h.endpoint, "", h.expectedMsg), "while checking function through the gateway")
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
	if spec.MinReplicas == nil {
		return errors.New("minReplicas equal nil")
	} else if spec.MaxReplicas == nil {
		return errors.New("maxReplicas equal nil")
	} else if spec.Resources.Requests.Memory().IsZero() || spec.Resources.Requests.Cpu().IsZero() {
		return errors.New("requests equal zero")
	} else if spec.Resources.Limits.Memory().IsZero() || spec.Resources.Limits.Cpu().IsZero() {
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

type E2EFunctionCheck struct {
	log          *logrus.Entry
	name         string
	inClusterURL string
	fnGatewayURL string
	brokerURL    string
	poller       poller.Poller
}

func NewE2EFunctionCheck(log *logrus.Entry, name string, inClusterURL, fnGatewayURL, brokerURL *url.URL, poller poller.Poller) E2EFunctionCheck {
	return E2EFunctionCheck{
		log:          log.WithField(step.LogStepKey, name),
		name:         name,
		inClusterURL: inClusterURL.String(),
		fnGatewayURL: fnGatewayURL.String(),
		brokerURL:    brokerURL.String(),
		poller:       poller,
	}
}

func (c E2EFunctionCheck) Name() string {
	return c.name
}

func (c E2EFunctionCheck) Run() error {
	c.log.Infof("Testing local connection through the service to updated Function")
	err := c.poller.PollForAnswer(c.inClusterURL, testsuite.HappyMsg, fmt.Sprintf("Hello %s world 1", testsuite.HappyMsg))
	if err != nil {
		return errors.Wrap(err, "while testing local connection through the service")
	}

	c.log.Infof("Step: %s, Testing connection through the Gateway", c.Name())
	c.log.Infof("Gateway URL: %s", c.fnGatewayURL)
	err = c.poller.PollForAnswer(c.fnGatewayURL, testsuite.HappyMsg, fmt.Sprintf("Hello %s world 2", testsuite.HappyMsg))
	if err != nil {
		return errors.Wrap(err, "while testing connection throight gateway")
	}

	c.log.Infof("Step: %s, Testing connection to event-mesh via Trigger", c.Name())
	// https://knative.dev/v0.12-docs/eventing/broker-trigger/
	err = testsuite.CreateEvent(c.brokerURL) // pinging the broker ingress sends an event to function via trigger
	if err != nil {
		return errors.Wrap(err, "while testing connection to event-mesh via Trigger")
	}

	c.log.Infof("Step: %s, Check if event has come to the function through the service", c.Name())
	err = c.poller.PollForAnswer(c.inClusterURL, "", testsuite.GotEventMsg)
	if err != nil {
		return errors.Wrap(err, "while local connection through the service")
	}

	c.log.Infof("Step: %s, Testing injection of env variables via incluster url", c.Name())
	err = c.poller.PollForAnswer(c.inClusterURL, testsuite.RedisEnvPing, testsuite.AnswerForEnvPing)
	if err != nil {
		return errors.Wrap(err, "while injection of env variables via incluster url")
	}

	c.log.Infof("Step: %s, Testing injection of env variables via gateway", c.Name())
	err = c.poller.PollForAnswer(c.fnGatewayURL, testsuite.RedisEnvPing, testsuite.AnswerForEnvPing)
	if err != nil {
		return errors.Wrap(err, "while injection of env variables via gateway")
	}
	return nil
}

func (c E2EFunctionCheck) Cleanup() error {
	return nil
}

func (e E2EFunctionCheck) OnError() error {
	return nil
}
