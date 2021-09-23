package doctor

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/kyma-project/kyma/components/nats-operator/pkg/client/natscluster"
	"github.com/kyma-project/kyma/components/nats-operator/pkg/errors"
)

const (
	// namespace is used as a value for the Kubernetes namespace which contains Eventing resources.
	namespace = "kyma-system"

	// labelKeyEventingBackend is used as a named key for Eventing backend label.
	labelKeyEventingBackend = "kyma-project.io/eventing-backend"
	// labelValEventingBackendBeb is used as a value for Eventing backend label.
	labelValEventingBackendBeb = "beb"

	// labelKeyNatsOperator is used as a named key for NATS operator label.
	labelKeyNatsOperator = "name"
	// labelValNatsOperator is used as a value for NATS operator label.
	labelValNatsOperator = "nats-operator"

	// labelKeyNatsCluster is used as a named key for NATS server Pod label.
	labelKeyNatsCluster = "nats_cluster"
	// labelValNatsCluster is used as a value for NATS server Pod label.
	labelValNatsCluster = "eventing-nats"

	// natsClusterName is used as a named key for NatsCluster CustomResource.
	natsClusterName = "eventing-nats"
	// natsOperatorDeploymentName is used as a value for NatsCluster CustomResource.
	natsOperatorDeploymentName = "nats-operator"
)

// state represents the state of a NATS cluster.
type state struct {
	natsOperatorDeployment *appsv1.Deployment
	natsOperatorPod        *v1.Pod
	natsServersActual      int
	natsServersDesired     int

	k8sClient  kubernetes.Interface
	natsClient *natscluster.Client
}

// newState returns a new state instance.
func newState(k8sClient kubernetes.Interface, natsClient *natscluster.Client) *state {
	return &state{k8sClient: k8sClient, natsClient: natsClient}
}

// compute updates the state internal objects.
func (s *state) compute(ctx context.Context) error {
	if err := s.computeNatsServersDesired(ctx); err != nil {
		return err
	}
	if err := s.computeNatsOperatorDeployment(ctx); err != nil {
		return err
	}
	if err := s.computeNatsOperatorPod(ctx); err != nil {
		return err
	}
	if err := s.computeNatsServersActual(ctx); err != nil {
		return err
	}
	return nil
}

// isNatsBackend returns true if the Eventing backend is NATS, otherwise returns false.
func (s *state) isNatsBackend(ctx context.Context) (bool, error) {
	secretList, err := s.k8sClient.CoreV1().Secrets("").List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", labelKeyEventingBackend, labelValEventingBackendBeb),
	})
	if err != nil {
		return false, err
	}
	return len(secretList.Items) == 0, nil
}

// computeNatsServersDesired updates the state natsServersDesired count.
func (s *state) computeNatsServersDesired(ctx context.Context) error {
	s.natsServersDesired = 0
	natsCluster, err := s.natsClient.Get(ctx, namespace, natsClusterName)
	if err == nil {
		s.natsServersDesired = natsCluster.Spec.Size
		return nil
	}
	if k8serrors.IsNotFound(err) {
		return errors.Recoverable(err)
	}
	return err
}

// computeNatsOperatorDeployment updates the state natsOperatorDeployment object.
func (s *state) computeNatsOperatorDeployment(ctx context.Context) error {
	s.natsOperatorDeployment = nil
	natsOperatorDeployment, err := s.k8sClient.AppsV1().Deployments(namespace).Get(ctx, natsOperatorDeploymentName, metav1.GetOptions{})
	if err == nil {
		s.natsOperatorDeployment = natsOperatorDeployment
		if s.natsOperatorDeployment.Spec.Replicas != nil && *s.natsOperatorDeployment.Spec.Replicas == 0 {
			return errors.Recoverable(fmt.Errorf("nats-operator deployment spec replicas is 0"))
		}
		return nil
	}
	if k8serrors.IsNotFound(err) {
		return errors.Recoverable(err)
	}
	return err
}

// computeNatsOperatorPod updates the state natsOperatorPod object.
func (s *state) computeNatsOperatorPod(ctx context.Context) error {
	s.natsOperatorPod = nil
	podList, err := s.k8sClient.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", labelKeyNatsOperator, labelValNatsOperator),
	})
	if err != nil {
		return err
	}
	if isPodListEmpty(podList) {
		return errors.Recoverable(fmt.Errorf("nats-operator pod not found in namespace %s", namespace))
	}
	s.natsOperatorPod = &podList.Items[0]
	if s.natsOperatorPod.Status.Phase != v1.PodRunning {
		return errors.Recoverable(fmt.Errorf("nats-operator pod in namespace %s is not running", namespace))
	}
	return nil
}

// computeNatsServersActual updates the state natsServersActual count.
func (s *state) computeNatsServersActual(ctx context.Context) error {
	s.natsServersActual = 0
	podList, err := s.k8sClient.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", labelKeyNatsCluster, labelValNatsCluster),
	})
	if err != nil {
		return err
	}
	if isPodListEmpty(podList) {
		return errors.Recoverable(fmt.Errorf("nats-server pod(s) not found in namespace %s", namespace))
	}
	for _, pod := range podList.Items {
		if pod.Status.Phase == v1.PodRunning {
			s.natsServersActual++
		}
	}
	return nil
}

// isPodListEmpty returns true if the given pod list is nil or has no items, otherwise returns false.
func isPodListEmpty(podList *v1.PodList) bool {
	return podList == nil || len(podList.Items) == 0
}
