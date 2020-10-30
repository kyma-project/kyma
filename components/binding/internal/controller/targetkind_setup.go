package controller

import (
	"github.com/kyma-project/kyma/components/binding/internal/worker"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func SetupTargetKindReconciler(cli client.Client, dynamicClient dynamic.Interface, log logrus.FieldLogger, targetKindStorage worker.TargetKindStorage, scheme *runtime.Scheme) *TargetKindReconciler {
	return NewTargetKindReconciler(
		cli,
		worker.NewTargetKindWorker(targetKindStorage, dynamicClient),
		log.WithField("reconciler", "TargetKind"),
		scheme)
}
