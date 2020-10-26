package controller

import (
	"github.com/kyma-project/kyma/components/binding/internal/target"
	"github.com/kyma-project/kyma/components/binding/internal/worker"

	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func SetupBindingReconciler(cli client.Client, log logrus.FieldLogger, scheme *runtime.Scheme) *BindingReconciler {
	bindingWorker := worker.NewBindingWorker(target.NewHandler(cli))

	return NewBindingReconciler(
		cli,
		bindingWorker,
		log.WithField("reconciler", "Binding"),
		scheme)
}
