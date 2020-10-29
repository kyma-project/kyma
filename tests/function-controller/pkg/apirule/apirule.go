package apirule

import (
	"context"
	"time"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/helpers"

	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	watchtools "k8s.io/client-go/tools/watch"

	apiruleTypes "github.com/kyma-project/kyma/tests/function-controller/pkg/apirule/types"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/resource"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/shared"

	"github.com/pkg/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type APIRule struct {
	resCli      *resource.Resource
	name        string
	namespace   string
	waitTimeout time.Duration
	log         *logrus.Entry
	verbose     bool
}

func New(name string, c shared.Container) *APIRule {
	return &APIRule{
		resCli:      resource.New(c.DynamicCli, apiruleTypes.GroupVersion.WithResource("apirules"), c.Namespace, c.Log, c.Verbose),
		name:        name,
		namespace:   c.Namespace,
		waitTimeout: c.WaitTimeout,
		log:         c.Log,
		verbose:     c.Verbose,
	}
}

func (a *APIRule) Create(serviceName, host string, port uint32) (string, error) {
	gateway := "kyma-gateway.kyma-system.svc.cluster.local"

	rule := &apiruleTypes.APIRule{
		TypeMeta: v1.TypeMeta{
			Kind:       "APIRule",
			APIVersion: apiruleTypes.GroupVersion.String(),
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      a.name,
			Namespace: a.namespace,
		},
		Spec: apiruleTypes.APIRuleSpec{
			Gateway: &gateway,
			Rules: []apiruleTypes.Rule{
				{
					Path:    "/.*",
					Methods: []string{"GET"},
					AccessStrategies: []*apiruleTypes.Authenticator{
						{
							Handler: &apiruleTypes.Handler{
								Name: "allow",
							},
						},
					},
				},
			},
			Service: &apiruleTypes.Service{
				Name: &serviceName,
				Port: &port,
				Host: &host,
			},
		},
	}

	resourceVersion, err := a.resCli.Create(rule)
	if err != nil {
		return resourceVersion, errors.Wrapf(err, "while creating APIRule %s in namespace %s", a.name, a.namespace)
	}
	return resourceVersion, err
}

func (a *APIRule) Delete() error {
	err := a.resCli.Delete(a.name, a.waitTimeout)
	if err != nil {
		return errors.Wrapf(err, "while deleting APIRule %s in namespace %s", a.name, a.namespace)
	}

	return nil
}

func (a *APIRule) Get() (*apiruleTypes.APIRule, error) {
	u, err := a.resCli.Get(a.name)
	if err != nil {
		return &apiruleTypes.APIRule{}, errors.Wrapf(err, "while getting ApiRule %s in namespace %s", a.name, a.namespace)
	}

	apirule, err := convertFromUnstructuredToAPIRule(u)
	if err != nil {
		return &apiruleTypes.APIRule{}, err
	}

	return &apirule, nil
}

func (a *APIRule) WaitForStatusRunning() error {
	apirule, err := a.Get()
	if err != nil {
		return err
	}

	if a.isStateReady(*apirule) {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), a.waitTimeout)
	defer cancel()

	condition := a.isApiRuleReady(a.name)
	_, err = watchtools.Until(ctx, apirule.GetResourceVersion(), a.resCli.ResCli, condition)
	if err != nil {
		return err
	}
	return nil
}

func (a *APIRule) LogResource() error {
	apiRule, err := a.Get()
	if err != nil {
		return err
	}

	out, err := helpers.PrettyMarshall(apiRule)
	if err != nil {
		return err
	}

	a.log.Infof("%s", out)
	return nil
}

func (a *APIRule) isApiRuleReady(name string) func(event watch.Event) (bool, error) {
	return func(event watch.Event) (bool, error) {
		u, ok := event.Object.(*unstructured.Unstructured)
		if !ok {
			return false, shared.ErrInvalidDataType
		}
		if u.GetName() != name {
			a.log.Infof("names mismatch, object's name %s, supplied %s", u.GetName(), name)
			return false, nil
		}

		apirule, err := convertFromUnstructuredToAPIRule(u)
		if err != nil {
			return false, err
		}

		return a.isStateReady(apirule), nil
	}
}

func (a *APIRule) isStateReady(apirule apiruleTypes.APIRule) bool {
	status := apirule.Status
	if status.VirtualServiceStatus == nil || status.APIRuleStatus == nil || status.AccessRuleStatus == nil {
		return false
	}

	ready := status.AccessRuleStatus.Code == apiruleTypes.StatusOK &&
		status.APIRuleStatus.Code == apiruleTypes.StatusOK &&
		status.VirtualServiceStatus.Code == apiruleTypes.StatusOK

	shared.LogReadiness(ready, a.verbose, a.name, a.log, apirule)

	return ready
}

func convertFromUnstructuredToAPIRule(u *unstructured.Unstructured) (apiruleTypes.APIRule, error) {
	apirule := apiruleTypes.APIRule{}
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &apirule)
	return apirule, err
}
