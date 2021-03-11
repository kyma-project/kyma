package teststep

import (
	"net/url"
	"strings"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/apirule"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/servicebinding"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/servicebindingusage"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/serviceinstance"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/step"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/subscription"
	"github.com/kyma-project/kyma/tests/function-controller/testsuite"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type ConfigureFunction struct {
	log             *logrus.Entry
	name            string
	fnName          string
	apiRule         *apirule.APIRule
	apiRuleURL      string
	sinkURL         *url.URL
	subscription    *subscription.Subscription
	svcInstance     *serviceinstance.ServiceInstance
	svcBinding      *servicebinding.ServiceBinding
	svcBindingUsage *servicebindingusage.ServiceBindingUsage
	domainPort      uint32
}

func NewConfigureFunction(log *logrus.Entry, name string, fnName string, apiRule *apirule.APIRule, apiruleURL *url.URL, sinkURL *url.URL, subscription *subscription.Subscription,
	serviceInstance *serviceinstance.ServiceInstance, binding *servicebinding.ServiceBinding, usage *servicebindingusage.ServiceBindingUsage,
	domainPort uint32) step.Step {

	apiruleURLWithoutScheme := strings.Trim(apiruleURL.String(), apiruleURL.Scheme)
	apiruleURLWithoutScheme = strings.Trim(apiruleURLWithoutScheme, "://")

	return &ConfigureFunction{
		log:             log.WithField(step.LogStepKey, name),
		name:            name,
		fnName:          fnName,
		apiRule:         apiRule,
		apiRuleURL:      apiruleURLWithoutScheme,
		sinkURL:         sinkURL,
		subscription:    subscription,
		svcInstance:     serviceInstance,
		svcBinding:      binding,
		svcBindingUsage: usage,
		domainPort:      domainPort,
	}
}

func (f ConfigureFunction) Name() string {
	return f.name
}

func (f ConfigureFunction) Run() error {
	f.log.Infof("Creating APIRule...")
	_, err := f.apiRule.Create(f.fnName, f.apiRuleURL, f.domainPort)
	if err != nil {
		return errors.Wrap(err, "while creating api rule")
	}

	f.log.Infof("Waiting for APIRule to have ready phase...")
	err = f.apiRule.WaitForStatusRunning()
	if err != nil {
		return errors.Wrap(err, "while checking api rule")
	}

	f.log.Infof("Creating service instance...")
	err = f.svcInstance.Create(testsuite.ServiceClassExternalName, testsuite.ServicePlanExternalName)
	if err != nil {
		return errors.Wrap(err, "while creating service instance")
	}

	f.log.Infof("Waiting for service instance to have ready phase...")
	err = f.svcInstance.WaitForStatusRunning()
	if err != nil {
		return errors.Wrap(err, "while checking service instance")
	}

	f.log.Infof("Creating service binding...")
	err = f.svcBinding.Create(f.svcInstance.GetName())
	if err != nil {
		return errors.Wrap(err, "while creating service binding")
	}

	f.log.Infof("Waiting for service binding to have ready phase...")
	err = f.svcBinding.WaitForStatusRunning()
	if err != nil {
		return errors.Wrap(err, "while waiting for service binding")
	}

	f.log.Infof("Creating service binding usage...")
	// we are deliberately creating Servicebindingusage HERE, to test how it behaves after function update
	err = f.svcBindingUsage.Create(f.svcBinding.GetName(), f.fnName, testsuite.RedisEnvPrefix)
	if err != nil {
		return errors.Wrap(err, "while creating service binding usage")
	}

	f.log.Infof("Waiting for service binding usage to have ready phase...")
	err = f.svcBindingUsage.WaitForStatusRunning()
	if err != nil {
		return errors.Wrap(err, "while waiting for service binding usage")
	}

	f.log.Infof("Creating the Subscirption...")
	subscription, err := f.subscription.Create(f.sinkURL)
	if err != nil {
		return errors.Wrap(err, "while creating subscription")
	}

	f.log.Infof("Waiting for Subscription to have ready phase...")
	err = f.subscription.WaitForStatusRunning(subscription)
	if err != nil {
		return errors.Wrap(err, "while waiting for subscription ready")
	}

	return nil
}

func (f ConfigureFunction) OnError() error {
	if err := f.apiRule.LogResource(); err != nil {
		return errors.Wrap(err, "while getting apirule")
	}

	if err := f.svcInstance.LogResource(); err != nil {
		return errors.Wrap(err, "while getting service instance")
	}

	if err := f.svcBinding.LogResource(); err != nil {
		return errors.Wrap(err, "while getting service binding")
	}

	if err := f.svcBindingUsage.LogResource(); err != nil {
		return errors.Wrap(err, "while getting service binding usage")
	}

	if err := f.subscription.LogResource(); err != nil {
		return errors.Wrap(err, "while getting subscription")
	}

	return nil
}

func (f ConfigureFunction) Cleanup() error {
	err := f.subscription.Delete()
	if err != nil {
		return errors.Wrap(err, "while deleting subscription")
	}

	err = f.svcBindingUsage.Delete()
	if err != nil {
		return errors.Wrap(err, "while deleting service binding usage")
	}

	err = f.svcBinding.Delete()
	if err != nil {
		return errors.Wrap(err, "while deleting service binding")
	}

	err = f.svcInstance.Delete()
	if err != nil {
		return errors.Wrap(err, "while deleting service instance")
	}

	err = f.apiRule.Delete()
	if err != nil {
		return errors.Wrap(err, "while deleting api rule")
	}

	return nil
}
