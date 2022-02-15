package controller

import (
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func SetupTargetKindReconciler(cli client.Client, worker TargetKindWorker, log logrus.FieldLogger, scheme *runtime.Scheme) *TargetKindReconciler {
	return NewTargetKindReconciler(
		cli,
		worker,
		log.WithField("reconciler", "TargetKind"),
		scheme)
}
