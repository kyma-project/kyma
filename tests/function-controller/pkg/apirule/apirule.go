package apirule

import (
	"context"
	"time"

	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/helpers"

	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/watch"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/resource"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/shared"

	"github.com/pkg/errors"
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
	gvr := schema.GroupVersionResource{
		Group:    "gateway.kyma-project.io",
		Version:  "v1alpha1",
		Resource: "apirules",
	}

	return &APIRule{
		resCli:      resource.New(c.DynamicCli, gvr, c.Namespace, c.Log, c.Verbose),
		name:        name,
		namespace:   c.Namespace,
		waitTimeout: c.WaitTimeout,
		log:         c.Log,
		verbose:     c.Verbose,
	}
}

func (a *APIRule) GetName() string {
	return a.name
}

func (a *APIRule) Create(serviceName, host string, port uint32) (string, error) {
	gateway := "kyma-gateway.kyma-system.svc.cluster.local"

	apirule := unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "gateway.kyma-project.io/v1alpha1",
			"kind":       "APIRule",
			"metadata": map[string]interface{}{
				"name":      a.name,
				"namespace": a.namespace,
			},
			"spec": map[string]interface{}{
				"gateway": gateway,
				"service": map[string]interface{}{
					"name": serviceName,
					"port": port,
					"host": host,
				},
				"rules": []map[string]interface{}{
					{
						"accessStrategies": []map[string]interface{}{
							{
								"config":  map[string]interface{}{},
								"handler": "allow",
							},
						},
						"methods": []interface{}{
							"GET",
						},
						"path": "/.*",
					},
				},
			},
		},
	}

	resourceVersion, err := a.resCli.Create(&apirule)
	if err != nil {
		return resourceVersion, errors.Wrapf(err, "while creating APIRule %s in namespace %s", a.name, a.namespace)
	}
	return resourceVersion, err
}

func (a *APIRule) Delete() error {
	err := a.resCli.Delete(a.name)
	if err != nil {
		return errors.Wrapf(err, "while deleting APIRule %s in namespace %s", a.name, a.namespace)
	}

	return nil
}

func (a *APIRule) Get() (*unstructured.Unstructured, error) {
	apirule, err := a.resCli.Get(a.name)
	if err != nil {
		return &unstructured.Unstructured{}, errors.Wrapf(err, "while getting ApiRule %s in namespace %s", a.name, a.namespace)
	}
	return apirule, nil
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
	return resource.WaitUntilConditionSatisfied(ctx, a.resCli.ResCli, condition)
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
		apirule, ok := event.Object.(*unstructured.Unstructured)
		if !ok {
			return false, shared.ErrInvalidDataType
		}
		if apirule.GetName() != name {
			a.log.Infof("names mismatch, object's name %s, supplied %s", apirule.GetName(), name)
			return false, nil
		}

		return a.isStateReady(*apirule), nil
	}
}

func (a *APIRule) isStateReady(apirule unstructured.Unstructured) bool {
	correctCode := "OK"

	apiruleStatusCode, apiruleStatusCodeFound, err := unstructured.NestedString(apirule.Object, "status", "APIRuleStatus", "code")
	apiruleStatus := err != nil || !apiruleStatusCodeFound || apiruleStatusCode != correctCode

	virtualServiceStatusCode, virtualServiceStatusCodeFound, err := unstructured.NestedString(apirule.Object, "status", "virtualServiceStatus", "code")
	virtualServiceStatus := err != nil || !virtualServiceStatusCodeFound || virtualServiceStatusCode != correctCode

	accessRuleStatusCode, accessRuleStatusCodeFound, err := unstructured.NestedString(apirule.Object, "status", "accessRuleStatus", "code")
	accessRuleStatus := err != nil || !accessRuleStatusCodeFound || accessRuleStatusCode != correctCode

	ready := apiruleStatus && virtualServiceStatus && accessRuleStatus

	shared.LogReadiness(ready, a.verbose, a.name, a.log, apirule)

	return ready
}
