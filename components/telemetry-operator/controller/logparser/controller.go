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
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/configchecksum"
	corev1 "k8s.io/api/core/v1"

	controllermetrics "github.com/kyma-project/kyma/components/telemetry-operator/controller/metrics"

	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	"github.com/kyma-project/kyma/components/telemetry-operator/controller"
)

const checksumAnnotationKey = "checksum/logparser-config"

//go:generate mockery --name DaemonSetProber --filename daemon_set_prober.go
type DaemonSetProber interface {
	IsReady(ctx context.Context, name types.NamespacedName) (bool, error)
}

//go:generate mockery --name DaemonSetAnnotator --filename daemon_set_annotator.go
type DaemonSetAnnotator interface {
	SetAnnotation(ctx context.Context, name types.NamespacedName, key, value string) error
}

type Config struct {
	ParsersConfigMap types.NamespacedName
	DaemonSet        types.NamespacedName
}

// Reconciler reconciles a LogParser object
type Reconciler struct {
	client.Client
	config    Config
	syncer    *syncer
	prober    DaemonSetProber
	annotator DaemonSetAnnotator
}

func NewReconciler(client client.Client, config Config, prober DaemonSetProber, annotator DaemonSetAnnotator) *Reconciler {
	var r Reconciler

	r.Client = client
	r.config = config
	r.prober = prober
	r.annotator = annotator
	r.syncer = newSyncer(client, config)

	controllermetrics.RegisterMetrics()

	return &r
}

// SetupWithManager sets up the controller with the Manager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&telemetryv1alpha1.LogParser{}).
		Complete(r)
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)
	var logParser telemetryv1alpha1.LogParser
	if err := r.Get(ctx, req.NamespacedName, &logParser); err != nil {
		log.Info("Ignoring deleted LogParser")
		// Ignore not-found errors since we can get them on deleted requests
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	var changed, err = r.syncer.SyncParsersConfigMap(ctx, &logParser)
	if err != nil {
		return ctrl.Result{Requeue: controller.ShouldRetryOn(err)}, nil
	}

	if changed {
		log.V(1).Info("Fluent Bit parser configuration was updated. Restarting the DaemonSet")

		if err = r.Update(ctx, &logParser); err != nil {
			log.Error(err, "Failed to update LogParser")
			return ctrl.Result{Requeue: controller.ShouldRetryOn(err)}, err
		}

		checksum, err := r.calculateConfigChecksum(ctx)
		if err != nil {
			return ctrl.Result{Requeue: controller.ShouldRetryOn(err)}, nil
		}

		if err := r.annotator.SetAnnotation(ctx, r.config.DaemonSet, checksumAnnotationKey, checksum); err != nil {
			return ctrl.Result{Requeue: controller.ShouldRetryOn(err)}, nil
		}

		condition := telemetryv1alpha1.NewLogParserCondition(
			telemetryv1alpha1.FluentBitDSRestartedReason,
			telemetryv1alpha1.LogParserPending,
		)
		if err = r.updateLogParserStatus(ctx, req.NamespacedName, condition); err != nil {
			return ctrl.Result{Requeue: controller.ShouldRetryOn(err)}, err
		}
	}

	if logParser.Status.GetCondition(telemetryv1alpha1.LogParserRunning) == nil {
		var ready bool
		ready, err = r.prober.IsReady(ctx, r.config.DaemonSet)
		if err != nil {
			log.Error(err, "Failed to check Fluent Bit readiness")
			return ctrl.Result{RequeueAfter: controller.RequeueTime}, err
		}
		if !ready {
			log.V(1).Info(fmt.Sprintf("Checked %s - not yet ready. Requeueing...", req.NamespacedName.Name))
			return ctrl.Result{RequeueAfter: controller.RequeueTime}, nil
		}
		log.V(1).Info(fmt.Sprintf("Checked %s - ready", req.NamespacedName.Name))

		condition := telemetryv1alpha1.NewLogParserCondition(
			telemetryv1alpha1.FluentBitDSRestartCompletedReason,
			telemetryv1alpha1.LogParserRunning,
		)

		if err = r.updateLogParserStatus(ctx, req.NamespacedName, condition); err != nil {
			return ctrl.Result{RequeueAfter: controller.RequeueTime}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *Reconciler) updateLogParserStatus(ctx context.Context, name types.NamespacedName, condition *telemetryv1alpha1.LogParserCondition) error {
	log := logf.FromContext(ctx)

	var logParser telemetryv1alpha1.LogParser
	if err := r.Get(ctx, name, &logParser); err != nil {
		log.Error(err, "Failed to get LogParser")
		return err
	}

	// Do not update status if the log parser is being deleted
	if logParser.DeletionTimestamp != nil {
		return nil
	}

	// If the log parser had a running condition and then was modified, all conditions are removed.
	// In this case, condition tracking starts off from the beginning.
	if logParser.Status.GetCondition(telemetryv1alpha1.LogParserRunning) != nil &&
		condition.Type == telemetryv1alpha1.LogParserPending {
		log.V(1).Info(fmt.Sprintf("Updating the status of %s to %s. Resetting previous conditions", name.Name, condition.Type))
		logParser.Status.Conditions = []telemetryv1alpha1.LogParserCondition{}
	} else {
		log.V(1).Info(fmt.Sprintf("Updating the status of %s to %s", name.Name, condition.Type))
	}

	logParser.Status.SetCondition(*condition)

	if err := r.Status().Update(ctx, &logParser); err != nil {
		log.Error(err, fmt.Sprintf("Failed to update LogParser status to %s", condition.Type))
		return err
	}
	return nil
}

func (r *Reconciler) calculateConfigChecksum(ctx context.Context) (string, error) {
	var cm corev1.ConfigMap
	if err := r.Get(ctx, r.config.ParsersConfigMap, &cm); err != nil {
		return "", fmt.Errorf("failed to get %s/%s ConfigMap: %v", r.config.ParsersConfigMap.Namespace, r.config.ParsersConfigMap.Name, err)
	}
	return configchecksum.Calculate([]corev1.ConfigMap{cm}, nil), nil
}
