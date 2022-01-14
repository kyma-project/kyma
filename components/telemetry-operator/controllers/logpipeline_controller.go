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

package controllers

import (
	"bytes"
	"context"
	"strings"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/api/v1alpha1"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/fluentbit"
)

const (
	sectionsConfigMapFinalizer = "FLUENT_BIT_SECTIONS_CONFIG_MAP"
	parserConfigMapFinalizer   = "FLUENT_BIT_PARSERS_CONFIG_MAP"
	//nolint:gosec
	secretRefsFinalizer = "FLUENT_BIT_SECRETS"
	filesFinalizer      = "FLUENT_BIT_FILES"
)

// LogPipelineReconciler reconciles a LogPipeline object
type LogPipelineReconciler struct {
	client.Client
	Scheme                     *runtime.Scheme
	FluentBitSectionsConfigMap types.NamespacedName
	FluentBitParsersConfigMap  types.NamespacedName
	FluentBitDaemonSet         types.NamespacedName
	FluentBitEnvSecret         types.NamespacedName
	FluentBitFilesConfigMap    types.NamespacedName
}

//+kubebuilder:rbac:groups=telemetry.kyma-project.io,resources=logpipelines,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=telemetry.kyma-project.io,resources=logpipelines/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=telemetry.kyma-project.io,resources=logpipelines/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=pods,verbs=delete;list;watch
//+kubebuilder:rbac:groups="apps",resources=daemonsets,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *LogPipelineReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	var logPipeline telemetryv1alpha1.LogPipeline
	if err := r.Get(ctx, req.NamespacedName, &logPipeline); err != nil {
		log.Info("Ignoring deleted LogPipeline")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	updatedSectionsCm, err := r.syncSectionsConfigMap(ctx, &logPipeline)
	if err != nil {
		log.Error(err, "Failed to sync Sections ConfigMap")
		return ctrl.Result{}, err
	}

	updatedParsersCm, err := r.syncParsersConfigMap(ctx, &logPipeline)
	if err != nil {
		log.Error(err, "Failed to sync Parsers ConfigMap")
		return ctrl.Result{}, err
	}

	updatedFilesCm, err := r.syncFilesConfigMap(ctx, &logPipeline)
	if err != nil {
		log.Error(err, "Failed to sync mounted files")
		return ctrl.Result{}, err
	}

	updatedEnv, err := r.syncSecretRefs(ctx, &logPipeline)
	if err != nil {
		log.Error(err, "Failed to sync secret references")
		return ctrl.Result{}, err
	}

	if updatedSectionsCm || updatedParsersCm || updatedFilesCm || updatedEnv {
		log.Info("Updated fluent bit configuration")
		if err := r.Update(ctx, &logPipeline); err != nil {
			log.Error(err, "Cannot update LogPipeline")
			return ctrl.Result{}, err
		}

		if err := r.deleteFluentBitPods(ctx, log); err != nil {
			log.Error(err, "Cannot delete fluent bit pods")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// Get ConfigMap from Kubernetes API or create new one if not existing.
func (r *LogPipelineReconciler) getOrCreateConfigMap(ctx context.Context, name types.NamespacedName) (corev1.ConfigMap, error) {
	var cm corev1.ConfigMap
	if err := r.Get(ctx, name, &cm); err != nil {
		if errors.IsNotFound(err) {
			cm = corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name.Name,
					Namespace: name.Namespace,
				},
			}
			if err := r.Create(ctx, &cm); err != nil {
				return cm, err
			}
		} else {
			return cm, err
		}
	}
	return cm, nil
}

// Synchronize LogPipeline with ConfigMap of FluentBit sections (Input, Filter and Output).
func (r *LogPipelineReconciler) syncSectionsConfigMap(ctx context.Context, logPipeline *telemetryv1alpha1.LogPipeline) (bool, error) {
	log := log.FromContext(ctx)
	cm, err := r.getOrCreateConfigMap(ctx, r.FluentBitSectionsConfigMap)
	if err != nil {
		return false, err
	}
	cmKey := logPipeline.Name + ".conf"

	// Add or remove Fluent Bit configuration sections
	if logPipeline.DeletionTimestamp != nil {
		if cm.Data != nil && controllerutil.ContainsFinalizer(logPipeline, sectionsConfigMapFinalizer) {
			log.Info("Deleting fluent bit config")
			delete(cm.Data, cmKey)
			controllerutil.RemoveFinalizer(logPipeline, sectionsConfigMapFinalizer)
		}
	} else {
		fluentBitConfig := mergeFluentBitConfig(logPipeline)
		if cm.Data == nil {
			data := make(map[string]string)
			data[cmKey] = fluentBitConfig
			cm.Data = data
		} else {
			if oldConfig, hasKey := cm.Data[cmKey]; hasKey && oldConfig == fluentBitConfig {
				// Nothing changed
				return false, nil
			}
			cm.Data[cmKey] = fluentBitConfig
		}
		controllerutil.AddFinalizer(logPipeline, sectionsConfigMapFinalizer)
	}

	// Update ConfigMap
	if err := r.Update(ctx, &cm); err != nil {
		return false, err
	}

	return true, nil
}

// Synchronize LogPipeline with ConfigMap of FluentBit parsers (Parser and MultiLineParser).
func (r *LogPipelineReconciler) syncParsersConfigMap(ctx context.Context, logPipeline *telemetryv1alpha1.LogPipeline) (bool, error) {
	log := log.FromContext(ctx)
	cm, err := r.getOrCreateConfigMap(ctx, r.FluentBitParsersConfigMap)
	if err != nil {
		return false, err
	}
	cmKey := "parsers.conf"

	// Add or remove Fluent Bit configuration parsers
	if logPipeline.DeletionTimestamp != nil {
		if cm.Data != nil && controllerutil.ContainsFinalizer(logPipeline, parserConfigMapFinalizer) {
			log.Info("Deleting fluent bit parsers config")
			delete(cm.Data, cmKey)
			controllerutil.RemoveFinalizer(logPipeline, parserConfigMapFinalizer)

			var logPipelines telemetryv1alpha1.LogPipelineList
			err = r.List(ctx, &logPipelines)
			if err != nil {
				return false, err
			}
			fluentBitParsersConfig := mergeFluentBitParsersConfig(&logPipelines)
			if fluentBitParsersConfig == "" {
				cm.Data = nil
			} else {
				data := make(map[string]string)
				data[cmKey] = fluentBitParsersConfig
				cm.Data = data
			}
		}
	} else {
		var logPipelines telemetryv1alpha1.LogPipelineList
		err = r.List(ctx, &logPipelines)
		if err != nil {
			return false, err
		}

		fluentBitParsersConfig := mergeFluentBitParsersConfig(&logPipelines)
		if fluentBitParsersConfig == "" {
			cm.Data = nil
		} else if cm.Data == nil {
			data := make(map[string]string)
			data[cmKey] = fluentBitParsersConfig
			cm.Data = data
		} else {
			if oldConfig, hasKey := cm.Data[cmKey]; hasKey && oldConfig == fluentBitParsersConfig {
				// Nothing changed
				return false, nil
			}
			cm.Data[cmKey] = fluentBitParsersConfig
		}
		controllerutil.AddFinalizer(logPipeline, parserConfigMapFinalizer)
	}

	// Update ConfigMap
	if err := r.Update(ctx, &cm); err != nil {
		return false, err
	}

	return true, nil
}

// Synchronize file references with Fluent Bit files ConfigMap.
func (r *LogPipelineReconciler) syncFilesConfigMap(ctx context.Context, logPipeline *telemetryv1alpha1.LogPipeline) (bool, error) {
	changed := false
	cm, err := r.getOrCreateConfigMap(ctx, r.FluentBitFilesConfigMap)
	if err != nil {
		return false, err
	}

	// Sync files from every section
	for _, file := range logPipeline.Spec.Files {
		if logPipeline.DeletionTimestamp != nil {
			if _, hasKey := cm.Data[file.Name]; hasKey {
				delete(cm.Data, file.Name)
				controllerutil.RemoveFinalizer(logPipeline, filesFinalizer)
				changed = true
			}
		} else {
			if cm.Data == nil {
				data := make(map[string]string)
				data[file.Name] = file.Content
				cm.Data = data
				controllerutil.AddFinalizer(logPipeline, filesFinalizer)
				changed = true
			} else {
				if oldContent, hasKey := cm.Data[file.Name]; hasKey && oldContent == file.Content {
					// Nothing changed
					continue
				}
				cm.Data[file.Name] = file.Content
				controllerutil.AddFinalizer(logPipeline, filesFinalizer)
				changed = true
			}
		}
	}

	// Update ConfigMap
	if err := r.Update(ctx, &cm); err != nil {
		return false, err
	}

	return changed, nil
}

// Copy referenced secrets to global Fluent Bit environment secret.
func (r *LogPipelineReconciler) syncSecretRefs(ctx context.Context, logPipeline *telemetryv1alpha1.LogPipeline) (bool, error) {
	log := log.FromContext(ctx)
	var secret corev1.Secret
	changed := false

	// Get or create Secret
	if err := r.Get(ctx, r.FluentBitEnvSecret, &secret); err != nil {
		if errors.IsNotFound(err) {
			secret = corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      r.FluentBitEnvSecret.Name,
					Namespace: r.FluentBitEnvSecret.Namespace,
				},
			}
			if err := r.Create(ctx, &secret); err != nil {
				return false, err
			}
			changed = true
		} else {
			return false, err
		}
	}

	//Sync environment from referenced Secrets to Fluent Bit Secret
	for _, secretRef := range logPipeline.Spec.SecretRefs {
		var referencedSecret corev1.Secret
		if err := r.Get(ctx, types.NamespacedName{Name: secretRef.Name, Namespace: secretRef.Namespace}, &referencedSecret); err != nil {
			log.Error(err, "Cannot read secret %s from namespace %s", secretRef.Name, secretRef.Namespace)
			continue
		}
		for k, v := range referencedSecret.Data {
			if logPipeline.DeletionTimestamp != nil {
				if _, hasKey := secret.Data[k]; hasKey {
					delete(secret.Data, k)
					controllerutil.RemoveFinalizer(logPipeline, secretRefsFinalizer)
					//nolint:ineffassign
					changed = true
				}
			} else {
				if secret.Data == nil {
					data := make(map[string][]byte)
					data[k] = v
					secret.Data = data
					controllerutil.AddFinalizer(logPipeline, secretRefsFinalizer)
					//nolint:ineffassign
					changed = true
				} else {
					if oldEnv, hasKey := secret.Data[k]; hasKey && bytes.Equal(oldEnv, v) {
						continue
					}
					secret.Data[k] = v
					controllerutil.AddFinalizer(logPipeline, secretRefsFinalizer)
					//nolint:ineffassign
					changed = true
				}
			}
			changed = true
		}
	}

	// Update Fluent Bit Secret
	if err := r.Update(ctx, &secret); err != nil {
		return false, err
	}

	return changed, nil
}

// Delete all Fluent Bit pods to apply new configuration.
func (r *LogPipelineReconciler) deleteFluentBitPods(ctx context.Context, log logr.Logger) error {
	var fluentBitDs appsv1.DaemonSet
	if err := r.Get(ctx, r.FluentBitDaemonSet, &fluentBitDs); err != nil {
		log.Error(err, "Cannot get Fluent Bit DaemonSet")
	}

	var fluentBitPods corev1.PodList
	if err := r.List(ctx, &fluentBitPods, client.InNamespace(r.FluentBitDaemonSet.Namespace), client.MatchingLabels(fluentBitDs.Spec.Template.Labels)); err != nil {
		log.Error(err, "Cannot list fluent bit pods")
		return err
	}

	log.Info("Restarting Fluent Bit pods")
	for i := range fluentBitPods.Items {
		if err := r.Delete(ctx, &fluentBitPods.Items[i]); err != nil {
			log.Error(err, "Cannot delete pod "+fluentBitPods.Items[i].Name)
		}
	}
	return nil
}

// Merge FluentBit filters and outputs to single FluentBit configuration.
func mergeFluentBitConfig(logPipeline *telemetryv1alpha1.LogPipeline) string {
	var sb strings.Builder
	for _, filter := range logPipeline.Spec.Filters {
		sb.WriteString(fluentbit.BuildConfigSection(fluentbit.FilterConfigHeader, filter.Content))
	}
	for _, output := range logPipeline.Spec.Outputs {
		sb.WriteString(fluentbit.BuildConfigSection(fluentbit.OutputConfigHeader, output.Content))
	}
	return sb.String()
}

// Merge FluentBit parsers and multiLine parsers to single FluentBit configuration.
func mergeFluentBitParsersConfig(logPipelines *telemetryv1alpha1.LogPipelineList) string {
	var sb strings.Builder
	for _, logPipeline := range logPipelines.Items {
		if logPipeline.DeletionTimestamp == nil {
			for _, parser := range logPipeline.Spec.Parsers {
				sb.WriteString(fluentbit.BuildConfigSection(fluentbit.ParserConfigHeader, parser.Content))
			}
			for _, multiLineParser := range logPipeline.Spec.MultiLineParsers {
				sb.WriteString(fluentbit.BuildConfigSection(fluentbit.MultiLineParserConfigHeader, multiLineParser.Content))
			}
		}
	}
	return sb.String()
}

// SetupWithManager sets up the controller with the Manager.
func (r *LogPipelineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&telemetryv1alpha1.LogPipeline{}).
		Complete(r)
}
