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

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	"github.com/kyma-project/kyma/components/telemetry-operator/controller"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/configchecksum"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
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
