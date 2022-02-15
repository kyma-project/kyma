package teststep

import (
	"net/url"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/kyma-project/kyma/tests/function-controller/testsuite"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/apirule"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/servicebinding"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/servicebindingusage"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/serviceinstance"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/step"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/subscription"
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
	createdObjects  []Object
}

type Object interface {
	Delete() error
	GetName() string
	LogResource() error
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
	if err := f.createAPIRule(); err != nil {
		return err
	}

	if err := f.createSvcInstance(); err != nil {
		return err
	}

	if err := f.createSvcBinding(); err != nil {
		return err
	}

	if err := f.createSvcBindingUsage(); err != nil {
		return err
	}

	if err := f.createSubscription(); err != nil {
		return err
	}

	return nil
}

func (f ConfigureFunction) createAPIRule() error {
	f.log.Infof("Creating APIRule...")
	_, err := f.apiRule.Create(f.fnName, f.apiRuleURL, f.domainPort)
	if err != nil {
		return errors.Wrap(err, "while creating api rule")
	}
	f.createdObjects = append(f.createdObjects, f.apiRule)

	f.log.Infof("Waiting for APIRule to have ready phase...")
	err = f.apiRule.WaitForStatusRunning()
	if err != nil {
		return errors.Wrap(err, "while checking api rule")
	}

	return nil
}

func (f ConfigureFunction) createSvcInstance() error {
	f.log.Infof("Creating service instance...")
	err := f.svcInstance.Create(testsuite.ServiceClassExternalName, testsuite.ServicePlanExternalName)
	if err != nil {
		return errors.Wrap(err, "while creating service instance")
	}
	f.createdObjects = append(f.createdObjects, f.svcInstance)

	f.log.Infof("Waiting for service instance to have ready phase...")
	err = f.svcInstance.WaitForStatusRunning()
	if err != nil {
		return errors.Wrap(err, "while checking service instance")
	}

	return nil
}

func (f ConfigureFunction) createSvcBinding() error {
	f.log.Infof("Creating service binding...")
	err := f.svcBinding.Create(f.svcInstance.GetName())
	if err != nil {
		return errors.Wrap(err, "while creating service binding")
	}
	f.createdObjects = append(f.createdObjects, f.svcBinding)

	f.log.Infof("Waiting for service binding to have ready phase...")
	err = f.svcBinding.WaitForStatusRunning()
	if err != nil {
		return errors.Wrap(err, "while waiting for service binding")
	}

	return nil
}

func (f ConfigureFunction) createSvcBindingUsage() error {
	f.log.Infof("Creating service binding usage...")
	// we are deliberately creating Servicebindingusage HERE, to test how it behaves after function update
	err := f.svcBindingUsage.Create(f.svcBinding.GetName(), f.fnName, testsuite.RedisEnvPrefix)
	if err != nil {
		return errors.Wrap(err, "while creating service binding usage")
	}
	f.createdObjects = append(f.createdObjects, f.svcBindingUsage)

	f.log.Infof("Waiting for service binding usage to have ready phase...")
	err = f.svcBindingUsage.WaitForStatusRunning()
	if err != nil {
		return errors.Wrap(err, "while waiting for service binding usage")
	}

	return nil
}

func (f ConfigureFunction) createSubscription() error {
	f.log.Infof("Creating a subscription...")
	_, err := f.subscription.Create(f.sinkURL)
	if err != nil {
		return errors.Wrap(err, "while creating subscription")
	}
	f.createdObjects = append(f.createdObjects, f.subscription)

	f.log.Infof("Waiting for subscription to be ready...")
	err = f.subscription.WaitForStatusRunning()
	if err != nil {
		return errors.Wrap(err, "while waiting for subscription ready")
	}

	return nil
}

func (f ConfigureFunction) OnError() error {
	var errAll *multierror.Error

	for _, object := range f.createdObjects {
		if err := object.LogResource(); err != nil {
			errAll = multierror.Append(errAll, errors.Wrapf(err, "while getting %s", object.GetName()))
		}
	}

	return errAll.ErrorOrNil()
}

func (f ConfigureFunction) Cleanup() error {
	var errAll *multierror.Error

	for i := len(f.createdObjects) - 1; i >= 0; i-- {
		if err := f.createdObjects[i].Delete(); err != nil {
			errAll = multierror.Append(errAll, errors.Wrapf(err, "while deleting %s", f.createdObjects[i].GetName()))
		}
	}

	return errAll.ErrorOrNil()
}
