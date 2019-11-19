package eventactivation

import (
	"context"

	pkgerrors "github.com/pkg/errors"
	"go.uber.org/zap"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/cache"

	"knative.dev/eventing/pkg/logging"
	"knative.dev/eventing/pkg/reconciler"
	"knative.dev/pkg/controller"

	applicationconnectorv1alpha1 "github.com/kyma-project/kyma/components/event-bus/apis/applicationconnector/v1alpha1"
	applicationconnectorclientv1alpha1 "github.com/kyma-project/kyma/components/event-bus/client/generated/clientset/internalclientset/typed/applicationconnector/v1alpha1"
	eventingclientv1alpha1 "github.com/kyma-project/kyma/components/event-bus/client/generated/clientset/internalclientset/typed/eventing/v1alpha1"
	applicationconnectorlistersv1alpha1 "github.com/kyma-project/kyma/components/event-bus/client/generated/lister/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/event-bus/internal/knative/util"
)

const (
	finalizerName = "eventactivation.finalizers.kyma-project.io"
)

type Reconciler struct {
	// wrapper for core controller components (clients, logger, ...)
	*reconciler.Base

	// listers index properties about resources
	eventActivationLister applicationconnectorlistersv1alpha1.EventActivationLister

	// clients allow interactions with API objects
	applicationconnectorClient applicationconnectorclientv1alpha1.ApplicationconnectorV1alpha1Interface
	eventingClient eventingclientv1alpha1.EventingV1alpha1Interface

	time util.CurrentTime
}

// Reconcile reconciles a EventActivation object
func (r *Reconciler) Reconcile(ctx context.Context, key string) error {
	ea, err := eventActivationByKey(key, r.eventActivationLister)
	if err != nil {
		return err
	}

	// Modify a copy, not the original.
	ea = ea.DeepCopy()

	// Reconcile this copy of the EventActivation and then write back any status
	// updates regardless of whether the reconcile error out.
	requeue, reconcileErr := r.reconcile(ctx, ea)
	if reconcileErr != nil {
		r.Recorder.Eventf(ea, corev1.EventTypeWarning, "EventactivationReconcileFailed",
			"Eventactivation reconciliation failed: %v", reconcileErr)
	}

	if updateStatusErr := util.UpdateEventActivation(r.applicationconnectorClient, ea); updateStatusErr != nil {
		r.Recorder.Eventf(ea, corev1.EventTypeWarning, "EventactivationReconcileFailed", "Updating EventActivation status failed: %v", updateStatusErr)
		return updateStatusErr
	}

	if !requeue && reconcileErr == nil {
		r.Recorder.Eventf(ea, corev1.EventTypeNormal, "EventactivationReconciled", "EventActivation reconciled, name: %q; namespace: %q", ea.Name, ea.Namespace)
	}

	return reconcileErr
}

// eventActivationByKey retrieves a EventActivation object from a lister by ns/name key.
func eventActivationByKey(key string, l applicationconnectorlistersv1alpha1.EventActivationLister) (*applicationconnectorv1alpha1.EventActivation, error) {
	ns, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return nil, controller.NewPermanentError(pkgerrors.Wrap(err, "invalid object key"))
	}

	src, err := l.EventActivations(ns).Get(name)
	switch {
	case apierrors.IsNotFound(err):
		return nil, controller.NewPermanentError(pkgerrors.Wrap(err, "object no longer exists"))
	case err != nil:
		return nil, err
	}

	return src, nil
}

func (r *Reconciler) reconcile(ctx context.Context, ea *applicationconnectorv1alpha1.EventActivation) (bool, error) {
	log := logging.FromContext(ctx)
	// delete or add finalizers
	if !ea.DeletionTimestamp.IsZero() {
		// deactivate all Kyma subscriptions related to this ea
		subs, _ := util.GetSubscriptionsForEventActivation(r.eventingClient, ea)
		util.DeactivateSubscriptions(r.eventingClient, subs, log, r.time)

		// remove the finalizer from the list
		ea.ObjectMeta.Finalizers = util.RemoveString(&ea.ObjectMeta.Finalizers, finalizerName)
		log.Info("Finalizer removed", zap.String("Finalizer name", finalizerName))
		return false, nil
	}

	// If we are adding the finalizer for the first time, then ensure that finalizer is persisted
	if !util.ContainsString(&ea.ObjectMeta.Finalizers, finalizerName) {
		//Finalizer is not added, let's add it
		ea.ObjectMeta.Finalizers = append(ea.ObjectMeta.Finalizers, finalizerName)
		log.Info("Finalizer added", zap.String("Finalizer name", finalizerName))
		return true, nil
	}

	// check and activate, if necessary, all the subscriptions
	if subs, err := util.GetSubscriptionsForEventActivation(r.eventingClient, ea); err != nil {
		log.Error("GetSubscriptionsForEventActivation() failed", zap.Error(err))
	} else {
		log.Info("Kyma subscriptions found: ", zap.Any("subs", subs))
		// activate all subscriptions
		if err := util.ActivateSubscriptions(r.eventingClient, subs, log, r.time); err != nil {
			log.Error("ActivateSubscriptions() failed", zap.Error(err))
			return false, err
		}
	}
	return false, nil
}
