package doctor

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/kyma-project/kyma/components/nats-operator/logger"
	"github.com/kyma-project/kyma/components/nats-operator/pkg/client/natscluster"
	"github.com/kyma-project/kyma/components/nats-operator/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Doctor represents a nats-operator doctor.
type Doctor struct {
	k8sClient kubernetes.Interface
	interval  time.Duration
	log       *logrus.Logger
	state     *state
	stop      chan os.Signal
}

// New returns a new nats-operator doctor instance.
func New(k8sClient kubernetes.Interface, natsClient *natscluster.Client, interval time.Duration, log *logrus.Logger) *Doctor {
	return &Doctor{
		k8sClient: k8sClient,
		interval:  interval,
		log:       log,
		state:     newState(k8sClient, natsClient),
		stop:      newSignalHandler(),
	}
}

// Start starts nats-operator doctor which runs forever unless received a shutdown signal.
func (d *Doctor) Start(ctx context.Context) error {
	d.log.Infof("nats-cluster health-check will start after %v", d.interval)

	tick := time.NewTicker(d.interval)
	defer tick.Stop()
	defer close(d.stop)

	for {
		select {
		case <-d.stop:
			{
				d.log.WithField(logger.LogKeyReason, "shutdown signal received").Info("nats-cluster health-check stopped")
				return nil
			}
		case <-tick.C:
			{
				// make sure Eventing backend is NATS
				if natsBackend, err := d.state.isNatsBackend(ctx); err != nil {
					return err
				} else if !natsBackend {
					d.log.WithField(logger.LogKeyReason, "Eventing backend is not NATS").Infof("nats-cluster health-check skipped will retry after %v", d.interval)
					continue
				}

				// compute cluster state
				if err := d.state.compute(ctx); err != nil && !errors.IsRecoverable(err) {
					return err
				}

				// resolve issues
				if err := d.resolveIssuesIfAny(ctx); err != nil {
					d.stateLogger().WithField(logger.LogKeyReason, err).Infof("nats-cluster resolve issues failed will retry after %v", d.interval)
					continue
				}
			}
		}
	}
}

// resolveIssuesIfAny detects any issues with the cluster and tries to fix them.
func (d *Doctor) resolveIssuesIfAny(ctx context.Context) error {
	// make sure nats-operator deployment exists
	if d.state.natsOperatorDeployment == nil {
		return fmt.Errorf("nats-operator deployment not found in namespace %s", namespace)
	}

	// resolve issue if nats-operator deployment spec replicas is zero
	if *d.state.natsOperatorDeployment.Spec.Replicas == 0 {
		d.log.WithField(logger.LogKeySolution, "scale-up nats-operator replicas to 1").Info("nats-operator replicas is 0")
		replicas := int32(1)
		d.state.natsOperatorDeployment.Spec.Replicas = &replicas
		if _, err := d.k8sClient.AppsV1().Deployments(namespace).Update(ctx, d.state.natsOperatorDeployment, metav1.UpdateOptions{}); err != nil {
			return err
		}
		return nil
	}

	// make sure nats-operator pod exists
	if d.state.natsOperatorPod == nil {
		return fmt.Errorf("nats-operator pod not found in namespace %s", namespace)
	}

	// make sure nats-operator pod is running
	if d.state.natsOperatorPod.Status.Phase != v1.PodRunning {
		d.stateLogger().WithField(logger.LogKeySolution, "delete nats-operator pod").Info("nats-operator is not running")
		return d.deleteNatsOperatorPod(ctx)
	}

	// resolve issue if actual nats-servers running is less than desired
	if d.state.natsServersActual < d.state.natsServersDesired {
		d.stateLogger().WithField(logger.LogKeySolution, "delete nats-operator pod").Info("nats-servers running is less than desired")
		return d.deleteNatsOperatorPod(ctx)
	}

	d.stateLogger().Info("nats-cluster is healthy")
	return nil
}

// stateLogger returns a decorated logger from the state.
func (d *Doctor) stateLogger() *logrus.Entry {
	fields := make(logrus.Fields, 2)
	if d.state.natsOperatorPod != nil {
		fields["nats-servers-running"] = fmt.Sprintf("%d/%d", d.state.natsServersActual, d.state.natsServersDesired)
		fields["nats-operator"] = fmt.Sprintf("%s/%s", d.state.natsOperatorPod.Namespace, d.state.natsOperatorPod.Name)
	}
	return d.log.WithFields(fields)
}

// newSignalHandler returns a new channel that receives os interrupt signals.
func newSignalHandler() chan os.Signal {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	return stop
}

// deleteNatsOperatorPod deletes the nats-operator pod from the cluster.
func (d *Doctor) deleteNatsOperatorPod(ctx context.Context) error {
	if err := d.k8sClient.CoreV1().Pods(d.state.natsOperatorPod.Namespace).Delete(ctx, d.state.natsOperatorPod.Name, metav1.DeleteOptions{}); err != nil {
		return err
	}
	d.state.natsOperatorPod = nil
	return nil
}
