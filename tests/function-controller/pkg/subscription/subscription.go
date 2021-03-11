package subscription

import (
	"context"
	"net/url"
	"time"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/helpers"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/resource"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/shared"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
)

type Subscription struct {
	resCli      *resource.Resource
	name        string
	namespace   string
	waitTimeout time.Duration
	log         *logrus.Entry
	verbose     bool
}

func New(name string, c shared.Container) *Subscription {
	gvr := schema.GroupVersionResource{
		Group:    "eventing.kyma-project.io",
		Version:  "v1alpha1",
		Resource: "subscriptions",
	}

	return &Subscription{
		resCli:      resource.New(c.DynamicCli, gvr, c.Namespace, c.Log, c.Verbose),
		name:        name,
		namespace:   c.Namespace,
		waitTimeout: c.WaitTimeout,
		log:         c.Log,
		verbose:     c.Verbose,
	}
}

func (s *Subscription) Create(funcName string, sinkUrl *url.URL) (string, error) {
	subscription := unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "eventing.kyma-project.io/v1alpha1",
			"kind":       "Subscription",
			"metadata": map[string]interface{}{
				"name":      s.name,
				"namespace": s.namespace,
			},
			"spec": map[string]interface{}{
				"id":       "id",
				"protocol": "NATS",
				"protocolsettings": map[string]interface{}{
					"contentMode":     "BINARY",
					"exemptHandshake": true,
					"qos":             "AT_LEAST_ONCE",
					"webhookAuth": map[string]interface{}{
						"type":         "oauth2",
						"grantType":    "client_credentials",
						"clientId":     "xxx",
						"clientSecret": "xxx",
						"tokenUrl":     "https://oauth2.xxx.com/oauth2/token",
						"scope":        []string{"guid-identifier"},
					},
				},
				"sink": sinkUrl.String(),
				"filter": map[string]interface{}{
					"dialect": "beb",
					"filters": []map[string]interface{}{
						{
							"eventSource": map[string]interface{}{
								"type":     "exact",
								"property": "source",
								"value":    "/default/kyma/id",
							},
							"eventType": map[string]interface{}{
								"type":     "exact",
								"property": "type",
								"value":    "sap.kyma.testapp1023.order.created.v1",
							},
						},
					},
				},
			},
		},
	}

	resourceVersion, err := s.resCli.Create(&subscription)
	if err != nil {
		return resourceVersion, errors.Wrapf(err, "while creating Subscription %s in namespace %s", s.name, s.namespace)
	}
	return resourceVersion, err
}

func (s *Subscription) Delete() error {
	err := s.resCli.Delete(s.name)
	if err != nil {
		return errors.Wrapf(err, "while deleting Subscription %s in namespace %s", s.name, s.namespace)
	}

	return nil
}

func (s *Subscription) Get() (*unstructured.Unstructured, error) {
	subscription, err := s.resCli.Get(s.name)
	if err != nil {
		return &unstructured.Unstructured{}, errors.Wrapf(err, "while getting Subscription %s in namespace %s", s.name, s.namespace)
	}

	return subscription, nil
}

func (s *Subscription) WaitForStatusRunning() error {
	subscription, err := s.Get()
	if err != nil {
		return err
	}

	if s.isStateReady(*subscription) {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.waitTimeout)
	defer cancel()

	condition := s.isSubscriptionReady()
	return resource.WaitUntilConditionSatisfied(ctx, s.resCli.ResCli, condition)
}

func (s *Subscription) LogResource() error {
	subscription, err := s.Get()
	if err != nil {
		return err
	}

	out, err := helpers.PrettyMarshall(subscription)
	if err != nil {
		return err
	}

	s.log.Infof("%s", out)
	return nil
}

func (s *Subscription) isSubscriptionReady() func(event watch.Event) (bool, error) {
	return func(event watch.Event) (bool, error) {
		subscription, ok := event.Object.(*unstructured.Unstructured)
		if !ok {
			return false, shared.ErrInvalidDataType
		}
		if subscription.GetName() != s.name {
			s.log.Infof("names mismatch, object's name %s, supplied %s", subscription.GetName(), s.name)
			return false, nil
		}

		return s.isStateReady(*subscription), nil
	}
}

func (s *Subscription) isStateReady(subscription unstructured.Unstructured) bool {
	correctCode := "OK"

	subscriptionStatusCode, subscriptionStatusCodeFound, err := unstructured.NestedString(subscription.Object, "status", "SubscriptionRuleStatus", "code")
	subscriptionStatus := err != nil || !subscriptionStatusCodeFound || subscriptionStatusCode != correctCode

	virtualServiceStatusCode, virtualServiceStatusCodeFound, err := unstructured.NestedString(subscription.Object, "status", "virtualServiceStatus", "code")
	virtualServiceStatus := err != nil || !virtualServiceStatusCodeFound || virtualServiceStatusCode != correctCode

	accessRuleStatusCode, accessRuleStatusCodeFound, err := unstructured.NestedString(subscription.Object, "status", "accessRuleStatus", "code")
	accessRuleStatus := err != nil || !accessRuleStatusCodeFound || accessRuleStatusCode != correctCode

	ready := subscriptionStatus && virtualServiceStatus && accessRuleStatus

	shared.LogReadiness(ready, s.verbose, s.name, s.log, subscription)

	return ready
}
