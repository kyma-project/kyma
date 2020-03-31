package migrator

import (
	"strconv"

	kubelessv1beta1 "github.com/kubeless/kubeless/pkg/apis/kubeless/v1beta1"
	"github.com/kyma-project/kyma/components/function-controller/migrator/pkg/resource/function"
	"github.com/kyma-project/kyma/components/function-controller/migrator/pkg/resource/kubeless"

	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
)

type KubelessFnOperator struct {
	Data   *kubelessv1beta1.Function
	ResCli kubeless.Function
}

func fromKubelessToServing(fn kubelessv1beta1.Function) (*function.Data, error) {
	var rawTimeout string
	switch fn.Spec.Timeout {
	case "":
		rawTimeout = "180"
	default:
		rawTimeout = fn.Spec.Timeout
	}

	timeout, err := strconv.Atoi(rawTimeout)
	if err != nil {
		return nil, errors.Wrap(err, "while converting timeout from string to int32")
	}

	var resources corev1.ResourceRequirements
	var envVars []corev1.EnvVar

	switch len(fn.Spec.Deployment.Spec.Template.Spec.Containers) {
	case 0:
		resources = corev1.ResourceRequirements{}
		envVars = []corev1.EnvVar{}
	default:
		resources = fn.Spec.Deployment.Spec.Template.Spec.Containers[0].Resources
		envVars = fn.Spec.Deployment.Spec.Template.Spec.Containers[0].Env
	}

	return &function.Data{
		Body:        fn.Spec.Function,
		Deps:        fn.Spec.Deps,
		Timeout:     int32(timeout),
		Labels:      fn.Labels,
		Annotations: fn.Annotations, // TODO discuss, this is probably bad idea
		Resources:   resources,
		EnvVar:      envVars,
	}, nil
}

func (m migrator) createServerlessFns() error {
	for _, fn := range m.kubelessFns {
		fnServing := function.New(m.dynamicCli, fn.Data.Name, fn.Data.Namespace, m.cfg.WaitTimeout, m.log.Info)

		servingFn, err := fromKubelessToServing(*fn.Data)
		if err != nil {
			return err
		}

		m.log.WithValues("Name", fn.Data.Name,
			"Namespace", fn.Data.Namespace,
			"GroupVersion", serverlessv1alpha1.GroupVersion.String(),
		).Info("Creating function")
		if err := fnServing.Create(servingFn); err != nil {
			return err
		}
	}
	return nil
}

func (m migrator) deleteKubelessFunctions() error {
	for _, fn := range m.kubelessFns {
		m.log.WithValues("name", fn.Data.Name,
			"namespace", fn.Data.Namespace,
			"GroupVersion", kubelessv1beta1.SchemeGroupVersion.String(),
		).Info("Deleting kubeless function")

		if err := fn.ResCli.Delete(); err != nil {
			return errors.Wrapf(err, "while deleting Kubeless function %s in namespace %s", fn.Data.Name, fn.Data.Namespace)
		}
	}
	return nil
}
