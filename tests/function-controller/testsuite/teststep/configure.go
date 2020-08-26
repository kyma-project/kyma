package teststep

import (
	"net/url"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/json"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/apirule"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/broker"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/servicebinding"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/servicebindingusage"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/serviceinstance"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/step"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/trigger"
	"github.com/kyma-project/kyma/tests/function-controller/testsuite"
)

type ConfigreFunction struct {
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

func ConfigureFunction(log *logrus.Entry, name string, fnName string, apiRule *apirule.APIRule, apiruleURL *url.URL,
	serviceInstance *serviceinstance.ServiceInstance, binding *servicebinding.ServiceBinding, usage *servicebindingusage.ServiceBindingUsage, broker *broker.Broker, trigger *trigger.Trigger,
	domainPort uint32) step.Step {

	apiruleURLWithoutScheme := strings.Trim(apiruleURL.String(), apiruleURL.Scheme)
	apiruleURLWithoutScheme = strings.Trim(apiruleURLWithoutScheme, "://")

	return &ConfigreFunction{
		log:             log,
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

func (f ConfigreFunction) Name() string {
	return f.name
}

func (f ConfigreFunction) Run() error {
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
		return errors.Wrap(err, "while waitiing for service binding")
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

type RuntimeObjectPrinter interface {
	Print(object runtime.Object)
}

func stringifyToYaml(r interface{}) (string, error) {
	out, err := json.Marshal(r)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func (f ConfigreFunction) prettyPrint(obj interface{}) {
	out, err := stringifyToYaml(obj)
	if err != nil {
		f.log.Warnf("error: %s", err)
	} else {
		f.log.Infof("%s", out)
	}
}

func (f ConfigreFunction) LogResources() error {
	tr, err := f.trigger.Get()
	if err != nil {
		return err
	}
	f.prettyPrint(tr)

	sbu, err := f.svcBindingUsage.Get()
	if err != nil {
		return err
	}
	f.prettyPrint(sbu)
	sb, err := f.svcBinding.Get()
	if err != nil {
		return err
	}
	f.prettyPrint(sb)

	si, err := f.svcInstance.Get()
	if err != nil {
		return err
	}
	f.prettyPrint(si)

	ar, err := f.apiRule.Get()
	if err != nil {
		return err
	}
	f.prettyPrint(ar)

	return nil
}

func (f ConfigreFunction) Cleanup() error {
	if err := f.LogResources(); err != nil {
		return errors.Wrapf(err, "while logging resources before cleanup")
	}

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
