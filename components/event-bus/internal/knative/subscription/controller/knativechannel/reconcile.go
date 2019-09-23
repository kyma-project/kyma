package knativechannel

import (
	"context"

	"knative.dev/pkg/apis"

	messagingV1Alpha1 "github.com/knative/eventing/pkg/apis/messaging/v1alpha1"
	subApis "github.com/kyma-project/kyma/components/event-bus/api/push/eventing.kyma-project.io/v1alpha1"
	"github.com/kyma-project/kyma/components/event-bus/internal/knative/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	finalizerName = "channel.finalizers.kyma-project.io"
)

type reconciler struct {
	client   client.Client
	recorder record.EventRecorder
	time     util.CurrentTime
}

// Verify the struct implements reconcile.Reconciler
var _ reconcile.Reconciler = &reconciler{}

func (r *reconciler) InjectClient(c client.Client) error {
	r.client = c
	return nil
}

// Reconcile reconciles a Kn Channel object
func (r *reconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	log.Info("Reconcile: ", "request", request)

	ctx := context.TODO()
	channel := &messagingV1Alpha1.Channel{}
	err := r.client.Get(ctx, request.NamespacedName, channel)

	// The Channel may have been deleted since it was added to the work queue. If so, there is
	// nothing to be done.
	if errors.IsNotFound(err) {
		log.Info("Could not find Channel: ", "err", err)
		return reconcile.Result{}, nil
	}

	// Any other error should be retrieved in another reconciliation. ???
	if err != nil {
		log.Error(err, "Unable to Get channel object")
		return reconcile.Result{}, err
	}

	log.Info("Reconciling Channel", "UID", string(channel.ObjectMeta.UID))

	// Modify a copy, not the original.
	channel = channel.DeepCopy()

	// Reconcile this copy of the EventActivation and then write back any status
	// updates regardless of whether the reconcile error out.
	requeue, reconcileErr := r.reconcile(ctx, channel)
	if reconcileErr != nil {
		log.Error(reconcileErr, "Reconciling Channel")
		r.recorder.Eventf(channel, corev1.EventTypeWarning, "ChannelReconcileFailed", "Channel reconciliation failed: %v", reconcileErr)
	}

	if updateStatusErr := util.UpdateKnativeChannel(ctx, r.client, channel); updateStatusErr != nil {
		log.Error(updateStatusErr, "Updating Knative Channel status")
		r.recorder.Eventf(channel, corev1.EventTypeWarning, "ChannelReconcileFailed", "Updating Kn Channel status failed: %v", updateStatusErr)
		return reconcile.Result{}, updateStatusErr
	}

	if !requeue && reconcileErr == nil {
		log.Info("Channel reconciled")
		r.recorder.Eventf(channel, corev1.EventTypeNormal, "ChannelReconciled", "Channel reconciled, name: %q; namespace: %q", channel.Name, channel.Namespace)
	}
	return reconcile.Result{
		Requeue: requeue,
	}, reconcileErr
}

func (r *reconciler) reconcile(ctx context.Context, ch *messagingV1Alpha1.Channel) (bool, error) {
	var isChReady, isChannelReadyInSub bool

	sub, err := util.GetSubscriptionForChannel(ctx, r.client, ch)
	if err != nil {
		log.Error(err, "GetSubscriptionsForChannel() failed")
		return false, err
	}
	if sub == nil {
		log.Info("No matching subscription found for channel: " + ch.Namespace + "/" + ch.Name)
		return false, nil
	}

	// Delete or add finalizers
	if !ch.DeletionTimestamp.IsZero() {
		util.DeactivateSubscriptionForChannel(ctx, r.client, sub, log, r.time)

		ch.ObjectMeta.Finalizers = util.RemoveString(&ch.ObjectMeta.Finalizers, finalizerName)
		log.Info("Finalizer removed for Knative Channel", "Finalizer name", finalizerName)
		return false, nil
	}

	// If we are adding the finalizer for the first time, then ensure that finalizer is persisted
	if !util.ContainsString(&ch.ObjectMeta.Finalizers, finalizerName) {
		ch.ObjectMeta.Finalizers = append(ch.ObjectMeta.Finalizers, finalizerName)
		log.Info("Finalizer added for Knative Channel", "Finalizer name", finalizerName)
		return true, nil
	}

	for _, condition := range ch.Status.Conditions {
		if condition.Type == apis.ConditionReady && condition.Status == corev1.ConditionTrue {
			isChReady = true
			break
		}
	}
	for _, cond := range sub.Status.Conditions {
		if cond.Type == subApis.ChannelReady && cond.Status == subApis.ConditionTrue {
			isChannelReadyInSub = true
			break
		}
	}
	if isChReady && !isChannelReadyInSub {
		if err := util.ActivateSubscriptionForChannel(ctx, r.client, sub, log, r.time); err != nil {
			log.Error(err, "ActivateSubscriptionForChannel() failed")
			return false, err
		}
	}

	if !isChReady && isChannelReadyInSub {
		if err := util.DeactivateSubscriptionForChannel(ctx, r.client, sub, log, r.time); err != nil {
			log.Error(err, "DeactivateSubscriptionForChannel() failed")
			return false, err
		}
	}
	return false, nil
}
