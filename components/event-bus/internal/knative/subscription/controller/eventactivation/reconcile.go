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
	finalizerName                   = "eventactivation.finalizers.kyma-project.io"
	eventactivationreconciled       = "EventactivationReconciled"
	eventactivationreconciledfailed = "EventactivationReconcileFailed"
)

//Reconciler EventActivation reconciler
type Reconciler struct {
	// wrapper for core controller components (clients, logger, ...)
	*reconciler.Base

	// listers index properties about resources
	eventActivationLister applicationconnectorlistersv1alpha1.EventActivationLister

	// clients allow interactions with API objects
	applicationconnectorClient applicationconnectorclientv1alpha1.ApplicationconnectorV1alpha1Interface
	kymaEventingClient         eventingclientv1alpha1.EventingV1alpha1Interface

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
	reconcileErr := r.reconcile(ctx, ea)
	if reconcileErr != nil {
		r.Recorder.Eventf(ea, corev1.EventTypeWarning, eventactivationreconciledfailed,
			"Eventactivation reconciliation failed: %v", reconcileErr)
	}

	if err := util.UpdateEventActivation(r.applicationconnectorClient, ea); err != nil {
		r.Recorder.Eventf(ea, corev1.EventTypeWarning, eventactivationreconciledfailed, "Updating EventActivation status failed: %v", err)
		return err
	}

	if reconcileErr == nil {
		r.Recorder.Eventf(ea, corev1.EventTypeNormal, eventactivationreconciled, "EventActivation reconciled, name: %q; namespace: %q", ea.Name, ea.Namespace)
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

func (r *Reconciler) reconcile(ctx context.Context, ea *applicationconnectorv1alpha1.EventActivation) error {
	log := logging.FromContext(ctx)
	// delete or add finalizers
	if !ea.DeletionTimestamp.IsZero() {
		// deactivate all Kyma subscriptions related to this ea
		subs, err := util.GetSubscriptionsForEventActivation(r.kymaEventingClient, ea)
		if err != nil {
			return err
		}
		error := util.DeactivateSubscriptions(r.kymaEventingClient, subs, log, r.time)
		if error != nil {
			return error
		}
		// remove the finalizer from the list
		ea.ObjectMeta.Finalizers = util.RemoveString(&ea.ObjectMeta.Finalizers, finalizerName)
		log.Info("Finalizer removed", zap.String("Finalizer name", finalizerName))
		return nil
	}

	// If we are adding the finalizer for the first time, then ensure that finalizer is persisted
	if !util.ContainsString(&ea.ObjectMeta.Finalizers, finalizerName) {
		//Finalizer is not added, let's add it
		ea.ObjectMeta.Finalizers = append(ea.ObjectMeta.Finalizers, finalizerName)
		log.Info("Finalizer added", zap.String("Finalizer name", finalizerName))
		return nil
	}

	// check and activate, if necessary, all the subscriptions
	subs, err := util.GetSubscriptionsForEventActivation(r.kymaEventingClient, ea)
	if err != nil {
		log.Error("GetSubscriptionsForEventActivation() failed", zap.Error(err))
		return err
	}
	log.Info("Kyma subscriptions found: ", zap.Any("subs", subs))
	// activate all subscriptions
	if err := util.ActivateSubscriptions(r.kymaEventingClient, subs, log, r.time); err != nil {
		log.Error("ActivateSubscriptions() failed", zap.Error(err))
		return err
	}
	return nil
}
