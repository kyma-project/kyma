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

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-controller/api/v1alpha1"
)

const (
	configMapFinalizer   = "FLUENT_BIT_CONFIG_MAP"
	environmentFinalizer = "FLUENT_BIT_ENV"
	filesFinalizer       = "FLUENT_BIT_FILES"
)

// LoggingConfigurationReconciler reconciles a LoggingConfiguration object
type LoggingConfigurationReconciler struct {
	client.Client
	Scheme                  *runtime.Scheme
	FluentBitConfigMap      types.NamespacedName
	FluentBitDaemonSet      types.NamespacedName
	FluentBitEnvSecret      types.NamespacedName
	FluentBitFilesConfigMap types.NamespacedName
}

//+kubebuilder:rbac:groups=telemetry.kyma-project.io,resources=loggingconfigurations,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=telemetry.kyma-project.io,resources=loggingconfigurations/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=telemetry.kyma-project.io,resources=loggingconfigurations/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=pods,verbs=delete;list;watch
//+kubebuilder:rbac:groups="apps",resources=daemonsets,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *LoggingConfigurationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	var config telemetryv1alpha1.LoggingConfiguration
	if err := r.Get(ctx, req.NamespacedName, &config); err != nil {
		log.Error(err, "unable to fetch LoggingConfiguration")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	updatedCm, err := r.syncConfigMap(ctx, &config)
	if err != nil {
		log.Error(err, "failed to sync ConfigMap")
		return ctrl.Result{}, err
	}

	updatedFiles, err := r.syncFiles(ctx, &config)
	if err != nil {
		log.Error(err, "failed to sync mounted files")
		return ctrl.Result{}, err
	}

	updatedEnv, err := r.syncEnvironment(ctx, &config)
	if err != nil {
		log.Error(err, "failed to sync environment variables")
		return ctrl.Result{}, err
	}

	if updatedCm || updatedFiles || updatedEnv {
		if err := r.Update(ctx, &config); err != nil {
			log.Error(err, "cannot update LoggingConfiguration")
			return ctrl.Result{}, err
		}

		if err := r.deleteFluentBitPods(ctx, log); err != nil {
			log.Error(err, "cannot delete fluent bit pods")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *LoggingConfigurationReconciler) syncConfigMap(ctx context.Context, config *telemetryv1alpha1.LoggingConfiguration) (bool, error) {
	var cm corev1.ConfigMap

	// Get or create ConfigMap
	if err := r.Get(ctx, r.FluentBitConfigMap, &cm); err != nil {
		if errors.IsNotFound(err) {
			cm = corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      r.FluentBitConfigMap.Name,
					Namespace: r.FluentBitConfigMap.Namespace,
				},
			}
			if err := r.Create(ctx, &cm); err != nil {
				return false, err
			}
		} else {
			return false, err
		}
	}
	cmKey := config.Name + ".conf"

	// Add or remove Fluent Bit configuration sections
	if config.DeletionTimestamp != nil {
		if cm.Data != nil && controllerutil.ContainsFinalizer(config, configMapFinalizer) {
			delete(cm.Data, cmKey)
			controllerutil.RemoveFinalizer(config, configMapFinalizer)
		}
	} else {
		fluentBitConfig := generateFluentBitConfig(config.Spec.Sections)
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
		controllerutil.AddFinalizer(config, configMapFinalizer)
	}

	// Update ConfigMap
	if err := r.Update(ctx, &cm); err != nil {
		return false, err
	}

	return true, nil
}

func (r *LoggingConfigurationReconciler) syncFiles(ctx context.Context, config *telemetryv1alpha1.LoggingConfiguration) (bool, error) {
	var cm corev1.ConfigMap
	changed := false

	// Get or create ConfigMap
	if err := r.Get(ctx, r.FluentBitFilesConfigMap, &cm); err != nil {
		if errors.IsNotFound(err) {
			cm = corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      r.FluentBitFilesConfigMap.Name,
					Namespace: r.FluentBitFilesConfigMap.Namespace,
				},
			}
			if err := r.Create(ctx, &cm); err != nil {
				return false, err
			}
			changed = true
		} else {
			return false, err
		}
	}

	// Sync files from every section
	for _, section := range config.Spec.Sections {
		for _, file := range section.Files {
			if config.DeletionTimestamp != nil {
				if _, hasKey := cm.Data[file.Name]; hasKey {
					delete(cm.Data, file.Name)
					controllerutil.RemoveFinalizer(config, filesFinalizer)
					changed = true
				}
			} else {
				if cm.Data == nil {
					data := make(map[string]string)
					data[file.Name] = file.Content
					cm.Data = data
					controllerutil.AddFinalizer(config, filesFinalizer)
					changed = true
				} else {
					if oldContent, hasKey := cm.Data[file.Name]; hasKey && oldContent == file.Content {
						// Nothing changed
						continue
					}
					cm.Data[file.Name] = file.Content
					controllerutil.AddFinalizer(config, filesFinalizer)
					changed = true
				}
			}
		}
	}

	// Update ConfigMap
	if err := r.Update(ctx, &cm); err != nil {
		return false, err
	}

	return changed, nil
}

func (r *LoggingConfigurationReconciler) syncEnvironment(ctx context.Context, config *telemetryv1alpha1.LoggingConfiguration) (bool, error) {
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
	for _, section := range config.Spec.Sections {
		for _, secretRef := range section.Environment {
			var referencedSecret corev1.Secret
			if err := r.Get(ctx, types.NamespacedName{Name: secretRef.Name, Namespace: secretRef.Namespace}, &referencedSecret); err != nil {
				log.Error(err, "cannot read secret %s from namespace %s", secretRef.Name, secretRef.Namespace)
				continue
			}
			for k, v := range referencedSecret.Data {
				if config.DeletionTimestamp != nil {
					if _, hasKey := secret.Data[k]; hasKey {
						delete(secret.Data, k)
						controllerutil.RemoveFinalizer(config, environmentFinalizer)
						//nolint:ineffassign
						changed = true
					}
				} else {
					if secret.Data == nil {
						data := make(map[string][]byte)
						data[k] = v
						secret.Data = data
						controllerutil.AddFinalizer(config, environmentFinalizer)
						//nolint:ineffassign
						changed = true
					} else {
						if oldEnv, hasKey := secret.Data[k]; hasKey && bytes.Equal(oldEnv, v) {
							continue
						}
						secret.Data[k] = v
						controllerutil.AddFinalizer(config, environmentFinalizer)
						//nolint:ineffassign
						changed = true
					}
				}
				changed = true
			}
		}
	}

	// Update Fluent Bit Secret
	if err := r.Update(ctx, &secret); err != nil {
		return false, err
	}

	return changed, nil
}

func (r *LoggingConfigurationReconciler) deleteFluentBitPods(ctx context.Context, log logr.Logger) error {
	var fluentBitDs appsv1.DaemonSet
	if err := r.Get(ctx, r.FluentBitDaemonSet, &fluentBitDs); err != nil {
		log.Error(err, "cannot get Fluent Bit DaemonSet")
	}

	var fluentBitPods corev1.PodList
	if err := r.List(ctx, &fluentBitPods, client.InNamespace(r.FluentBitDaemonSet.Namespace), client.MatchingLabels(fluentBitDs.Spec.Template.Labels)); err != nil {
		log.Error(err, "cannot list fluent bit pods")
		return err
	}

	log.Info("restarting Fluent Bit pods")
	for i := range fluentBitPods.Items {
		if err := r.Delete(ctx, &fluentBitPods.Items[i]); err != nil {
			log.Error(err, "cannot delete pod "+fluentBitPods.Items[i].Name)
		}
	}
	return nil
}

func generateFluentBitConfig(sections []telemetryv1alpha1.Section) string {
	var result string
	for _, section := range sections {
		result += section.Content + "\n\n"
	}
	return result
}

// SetupWithManager sets up the controller with the Manager.
func (r *LoggingConfigurationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&telemetryv1alpha1.LoggingConfiguration{}).
		Complete(r)
}
