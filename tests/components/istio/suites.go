package istio

import (
	"github.com/kyma-project/kyma/tests/components/istio/helpers"
	corev1 "k8s.io/api/core/v1"
)

type istioInstalledCase struct {
	pilotPods     *corev1.PodList
	ingressGwPods *corev1.PodList
}

type istioReconcilationCase struct {
	command     helpers.Command
	doneChannel chan struct{}
}
