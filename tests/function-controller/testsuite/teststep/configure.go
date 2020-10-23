package teststep

import (
	"net/url"
	"strings"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/apirule"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/broker"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/servicebinding"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/servicebindingusage"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/serviceinstance"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/step"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/trigger"
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
	svcInstance     *serviceinstance.ServiceInstance
	svcBinding      *servicebinding.ServiceBinding
	svcBindingUsage *servicebindingusage.ServiceBindingUsage
	broker          *broker.Broker
	trigger         *trigger.Trigger
	domainPort      uint32
}

func NewConfigureFunction(log *logrus.Entry, name string, fnName string, apiRule *apirule.APIRule, apiruleURL *url.URL,
	serviceInstance *serviceinstance.ServiceInstance, binding *servicebinding.ServiceBinding, usage *servicebindingusage.ServiceBindingUsage, broker *broker.Broker, trigger *trigger.Trigger,
	domainPort uint32) step.Step {

	apiruleURLWithoutScheme := strings.Trim(apiruleURL.String(), apiruleURL.Scheme)
	apiruleURLWithoutScheme = strings.Trim(apiruleURLWithoutScheme, "://")

	return &ConfigureFunction{
		log:             log.WithField(step.LogStepKey, name),
		name:            name,
		fnName:          fnName,
		apiRule:         apiRule,
		apiRuleURL:      apiruleURLWithoutScheme,
		svcInstance:     serviceInstance,
		svcBinding:      binding,
		svcBindingUsage: usage,
		broker:          broker,
		trigger:         trigger,
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

	f.log.Infof("Waiting for broker to have ready phase...")
	err = f.broker.WaitForStatusRunning()
	if err != nil {
		return errors.Wrap(err, "while waiting for broker")
	}
	// Trigger needs to be created after broker, as it depends on it
	// watch out for a situation where broker is not created yet!
	f.log.Infof("Creating Trigger...")
	err = f.trigger.Create(f.fnName)
	if err != nil {
		return errors.Wrap(err, "while creating trigger")
	}

	f.log.Infof("Waiting for Trigger to have ready phase...")
	err = f.trigger.WaitForStatusRunning()
	if err != nil {
		return errors.Wrap(err, "while waiting for trigger")
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

	if err := f.broker.LogResource(); err != nil {
		return errors.Wrap(err, "while getting broker")
	}

	if err := f.trigger.LogResource(); err != nil {
		return errors.Wrap(err, "while getting trigger")
	}

	return nil
}

func (f ConfigureFunction) Cleanup() error {
	err := f.trigger.Delete()
	if err != nil {
		return errors.Wrap(err, "while deleting trigger")
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
