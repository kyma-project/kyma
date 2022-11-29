/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package logparser

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	"github.com/kyma-project/kyma/components/telemetry-operator/controller"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/configchecksum"
)

const checksumAnnotationKey = "checksum/logparser-config"

type Config struct {
	ParsersConfigMap types.NamespacedName
	DaemonSet        types.NamespacedName
}

//go:generate mockery --name DaemonSetProber --filename daemon_set_prober.go
type DaemonSetProber interface {
	IsReady(ctx context.Context, name types.NamespacedName) (bool, error)
}

//go:generate mockery --name DaemonSetAnnotator --filename daemon_set_annotator.go
type DaemonSetAnnotator interface {
	SetAnnotation(ctx context.Context, name types.NamespacedName, key, value string) error
}

type Reconciler struct {
	client.Client
	config    Config
	prober    DaemonSetProber
	annotator DaemonSetAnnotator
	syncer    syncer
}

func NewReconciler(client client.Client, config Config, prober DaemonSetProber, annotator DaemonSetAnnotator) *Reconciler {
	var r Reconciler

	r.Client = client
	r.config = config
	r.prober = prober
	r.annotator = annotator
	r.syncer = syncer{client, config}

	return &r
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&telemetryv1alpha1.LogParser{}).
		Watches(
			&source.Kind{Type: &appsv1.DaemonSet{}},
			handler.EnqueueRequestsFromMapFunc(r.mapDaemonSets),
			builder.WithPredicates(onlyUpdate()),
		).
		Complete(r)
}

func onlyUpdate() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc:  func(event event.CreateEvent) bool { return false },
		DeleteFunc:  func(deleteEvent event.DeleteEvent) bool { return false },
		UpdateFunc:  func(updateEvent event.UpdateEvent) bool { return true },
		GenericFunc: func(genericEvent event.GenericEvent) bool { return false },
	}
}

func (r *Reconciler) mapDaemonSets(object client.Object) []reconcile.Request {
	daemonSet := object.(*appsv1.DaemonSet)

	var requests []reconcile.Request
	if daemonSet.Name != r.config.DaemonSet.Name || daemonSet.Namespace != r.config.DaemonSet.Namespace {
		return requests
	}

	var allParsers telemetryv1alpha1.LogParserList
	if err := r.List(context.Background(), &allParsers); err != nil {
		ctrl.Log.Error(err, "DamonSet UpdateEvent: fetching LogParserList failed!", err.Error())
		return requests
	}

	for _, parser := range allParsers.Items {
		requests = append(requests, reconcile.Request{NamespacedName: types.NamespacedName{Name: parser.Name}})
	}
	ctrl.Log.V(1).Info(fmt.Sprintf("DaemonSet changed event handling done: Created %d new reconciliation requests.\n", len(requests)))
	return requests
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (reconcileResult ctrl.Result, reconcileErr error) {
	log := logf.FromContext(ctx)
	log.V(1).Info("Reconciliation triggered")

	var parser telemetryv1alpha1.LogParser
	if err := r.Get(ctx, req.NamespacedName, &parser); err != nil {
		log.V(1).Info("Ignoring deleted LogParser")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	defer func() {
		if err := r.updateStatus(ctx, parser.Name); err != nil {
			reconcileResult = ctrl.Result{Requeue: controller.ShouldRetryOn(err)}
			reconcileErr = fmt.Errorf("failed to update LogPipeline status: %v", err)
		}
	}()

	err := ensureFinalizer(ctx, r.Client, &parser)
	if err != nil {
		return ctrl.Result{Requeue: controller.ShouldRetryOn(err)}, nil
	}

	if err = r.syncer.syncFluentBitConfig(ctx); err != nil {
		return ctrl.Result{Requeue: controller.ShouldRetryOn(err)}, nil
	}

	err = cleanupFinalizerIfNeeded(ctx, r.Client, &parser)
	if err != nil {
		return ctrl.Result{Requeue: controller.ShouldRetryOn(err)}, nil
	}

	checksum, err := r.calculateConfigChecksum(ctx)
	if err != nil {
		return ctrl.Result{Requeue: controller.ShouldRetryOn(err)}, nil
	}

	if err = r.annotator.SetAnnotation(ctx, r.config.DaemonSet, checksumAnnotationKey, checksum); err != nil {
		return ctrl.Result{Requeue: controller.ShouldRetryOn(err)}, nil
	}

	return reconcileResult, reconcileErr
}

func (r *Reconciler) calculateConfigChecksum(ctx context.Context) (string, error) {
	var cm corev1.ConfigMap
	if err := r.Get(ctx, r.config.ParsersConfigMap, &cm); err != nil {
		return "", fmt.Errorf("failed to get %s/%s ConfigMap: %v", r.config.ParsersConfigMap.Namespace, r.config.ParsersConfigMap.Name, err)
	}
	return configchecksum.Calculate([]corev1.ConfigMap{cm}, nil), nil
}
