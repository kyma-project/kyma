package migrator

import (
	"net"
	"net/url"
	"strings"

	"github.com/kyma-project/kyma/components/function-controller/migrator/pkg/resource/function"
	"github.com/kyma-project/kyma/components/function-controller/migrator/pkg/resource/trigger"
	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	eventingv1alpha1 "knative.dev/eventing/pkg/apis/eventing/v1alpha1"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var IncorrectSubscriberUri = errors.New("Incorrect subscriber uri format")

func (m migrator) setTriggerControllerReference(dependent *eventingv1alpha1.Trigger) error {
	fnName := dependent.Spec.Subscriber.Ref.Name
	fnNamespace := dependent.Spec.Subscriber.Ref.Namespace
	funCli := function.New(m.dynamicCli, fnName, fnNamespace, m.cfg.WaitTimeout, m.log.Info)
	fn, err := funCli.Get()
	if err != nil {
		return err
	}
	scheme := runtime.NewScheme()
	if err := serverlessv1alpha1.AddToScheme(scheme); err != nil {
		return errors.Wrap(err, "while adding serverless scheme to runtime scheme")
	}

	if err := eventingv1alpha1.AddToScheme(scheme); err != nil {
		return errors.Wrap(err, "while adding eventing scheme to runtime scheme")
	}

	if err := controllerutil.SetControllerReference(fn, dependent, scheme); err != nil {
		return errors.Wrap(err, "while setting controller reference")
	}
	return nil
}

func (m migrator) updateTriggers() error {
	triggers, err := m.getTriggerList()
	if err != nil {
		return err
	}

	for _, tr := range triggers {
		dest, err := migrateTriggerDestination(tr.Data)
		if err != nil {
			return err
		}
		tr.Data.Spec.Subscriber = dest

		// this has to be called before update
		if err := m.setTriggerControllerReference(tr.Data); err != nil {
			return err
		}

		m.log.WithValues("Name", tr.Data.Name,
			"Namespace", tr.Data.Namespace,
			"GroupVersion", eventingv1alpha1.SchemeGroupVersion.String(),
		).Info("Updating Trigger")

		if err := tr.ResCli.Update(tr.Data); err != nil {
			return err
		}
	}
	return nil
}

type TriggerOperator struct {
	Data   *eventingv1alpha1.Trigger
	ResCli *trigger.Trigger
}

func (m migrator) getTriggerList() ([]TriggerOperator, error) {
	triggerList, err := trigger.New(m.dynamicCli, "", "", m.cfg.WaitTimeout, m.log.Info).List()
	if err != nil {
		return nil, errors.Wrap(err, "while listing Knative Triggers")
	}

	var ret []TriggerOperator
	for _, tr := range triggerList {
		ret = append(ret, TriggerOperator{
			Data:   tr,
			ResCli: trigger.New(m.dynamicCli, tr.Name, tr.Namespace, m.cfg.WaitTimeout, m.log.Info),
		})
	}

	return ret, nil
}

func migrateTriggerDestination(tr *eventingv1alpha1.Trigger) (*duckv1.Destination, error) {
	subscriberUri, err := url.ParseRequestURI(tr.Spec.Subscriber.URI.String())
	if err != nil {
		return nil, err
	}

	hostWithoutPort, _, err := net.SplitHostPort(subscriberUri.Host)
	if err != nil {
		return nil, errors.Wrap(err, "while splitting host and port in subscriberUri")
	}

	splitHost := strings.Split(hostWithoutPort, ".")
	if len(splitHost) != 2 {
		return nil, IncorrectSubscriberUri
	}
	svcName, svcNamespace := splitHost[0], splitHost[1]

	apiVersion := servingv1.SchemeGroupVersion.Identifier()

	return &duckv1.Destination{
		Ref: &corev1.ObjectReference{
			Kind:       "Service",
			Namespace:  svcNamespace,
			Name:       svcName,
			APIVersion: apiVersion,
		},
		URI: nil,
	}, nil
}
