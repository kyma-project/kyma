package tracepipeline

import (
	"context"
	"fmt"
	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	newReconciler := ctrl.NewControllerManagedBy(mgr).
		For(&telemetryv1alpha1.TracePipeline{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Secret{}).
		Owns(&corev1.Service{}).
		Watches(
			&source.Kind{Type: &corev1.Secret{}},
			handler.EnqueueRequestsFromMapFunc(r.mapSecret),
			builder.WithPredicates(onlyUpdate()),
		).
		Watches(
			&source.Kind{Type: &appsv1.Deployment{}},
			handler.EnqueueRequestsFromMapFunc(r.mapDeployments),
			builder.WithPredicates(onlyUpdate()),
		)

	if r.config.CreateServiceMonitor {
		newReconciler.Owns(&monitoringv1.ServiceMonitor{})
	}

	return newReconciler.Complete(r)
}

func onlyUpdate() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc:  func(event event.CreateEvent) bool { return false },
		DeleteFunc:  func(deleteEvent event.DeleteEvent) bool { return false },
		UpdateFunc:  func(updateEvent event.UpdateEvent) bool { return true },
		GenericFunc: func(genericEvent event.GenericEvent) bool { return false },
	}
}

func (r *Reconciler) mapSecret(object client.Object) []reconcile.Request {
	secret := object.(*corev1.Secret)
	var pipelines telemetryv1alpha1.TracePipelineList
	var requests []reconcile.Request
	err := r.List(context.Background(), &pipelines)
	if err != nil {
		ctrl.Log.Error(err, "Secret UpdateEvent: fetching TracePipelineList failed!", err.Error())
		return requests
	}

	ctrl.Log.V(1).Info(fmt.Sprintf("Secret UpdateEvent: handling Secret: %s", secret.Name))
	for i := range pipelines.Items {
		var pipeline = pipelines.Items[i]
		if containsAnyRefToSecret(&pipeline, secret) {
			requests = append(requests, reconcile.Request{NamespacedName: types.NamespacedName{Name: pipeline.Name}})
			ctrl.Log.V(1).Info(fmt.Sprintf("Secret UpdateEvent: added reconcile request for pipeline: %s", pipeline.Name))
		}
	}
	return requests
}

func (r *Reconciler) mapDeployments(object client.Object) []reconcile.Request {
	deployment := object.(*appsv1.Deployment)
	var requests []reconcile.Request
	if deployment.Name != r.config.ResourceName || deployment.Namespace != r.config.CollectorNamespace {
		return requests
	}

	var allPipelines telemetryv1alpha1.TracePipelineList
	if err := r.List(context.Background(), &allPipelines); err != nil {
		ctrl.Log.Error(err, "Deployment UpdateEvent: fetching TracePipelineList failed!", err.Error())
		return requests
	}

	for range allPipelines.Items {
		requests = append(requests, reconcile.Request{NamespacedName: types.NamespacedName{Name: r.config.ResourceName, Namespace: r.config.CollectorNamespace}})
	}
	ctrl.Log.V(1).Info(fmt.Sprintf("Deployment changed event handling done: Created %d new reconciliation requests.\n", len(requests)))
	return requests
}

func containsAnyRefToSecret(pipeline *telemetryv1alpha1.TracePipeline, secret *corev1.Secret) bool {
	secretName := types.NamespacedName{Namespace: secret.Namespace, Name: secret.Name}
	if pipeline.Spec.Output.Otlp.Endpoint.IsDefined() &&
		referencesSecret(pipeline.Spec.Output.Otlp.Endpoint, secretName) {
		return true
	}

	if pipeline.Spec.Output.Otlp == nil ||
		pipeline.Spec.Output.Otlp.Authentication == nil ||
		pipeline.Spec.Output.Otlp.Authentication.Basic == nil ||
		!pipeline.Spec.Output.Otlp.Authentication.Basic.IsDefined() {
		return false
	}

	auth := pipeline.Spec.Output.Otlp.Authentication.Basic

	return referencesSecret(auth.User, secretName) || referencesSecret(auth.Password, secretName)
}

func referencesSecret(valueType telemetryv1alpha1.ValueType, secretName types.NamespacedName) bool {
	if valueType.Value == "" && valueType.ValueFrom != nil && valueType.ValueFrom.IsSecretKeyRef() {
		return valueType.ValueFrom.SecretKeyRef.NamespacedName() == secretName
	}

	return false
}
