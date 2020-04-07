/*
Copyright 2019 The Kyma Authors.

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
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	tektonv1alpha1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"

	serverless "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	funcerr "github.com/kyma-project/kyma/components/function-controller/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	utilrand "k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	knapis "knative.dev/pkg/apis"
	servingapis "knative.dev/serving/pkg/apis/serving"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	rtClient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	sourceVolName       = "source"
	dockerfileVolName   = "dockerfile"
	kanikoExecutorImage = "gcr.io/kaniko-project/executor:v0.19.0"
	// https://github.com/tektoncd/pipeline/blob/master/docs/auth.md#least-privilege
	tektonDockerVolume = "/tekton/home/.docker/"
	cmEntryChanged     = "'%s' changed, image rebuild is required"
)

// List of annotations set on Knative Serving objects by the Knative Serving admission webhook.
var (
	knativeServingAnnotations = []string{
		servingapis.GroupName + knapis.CreatorAnnotationSuffix,
		servingapis.GroupName + knapis.UpdaterAnnotationSuffix,
	}
	taskrunEnvs = []corev1.EnvVar{
		{
			Name:  "DOCKER_CONFIG",
			Value: tektonDockerVolume,
		},
	}
	containerName                            = "lambda"
	DefaultImagePullAccount                  = "function-controller"
	DefaultTektonRequestsCPU                 = resource.MustParse("350m")
	DefaultTektonRequestsMem                 = resource.MustParse("750Mi")
	DefaultTektonLimitsCPU                   = resource.MustParse("400m")
	DefaultTektonLimitsMem                   = resource.MustParse("1G")
	DefaultDockerRegistryPort            int = 5000
	DefaultDockerRegistryFqdn                = "function-controller-docker-registry.kyma-system.svc.cluster.local"
	DefaultDockerRegistryExternalAddress     = "https://registry.kyma.local"
	DefaultImagePullSecretName               = "regcred"
	DefaultRuntimeConfigmapName              = "fn-ctrl-runtime"
	DefaultRequeueDuration                   = time.Minute * 5
)

type DockerCfg struct {
	DockerRegistryPort            int
	DockerRegistryFqdn            string
	DockerRegistryExternalAddress string
	DockerRegistryName            string
}
type FnReconcilerCfg struct {
	MaxConcurrentReconciles int
	Limits, Requests        *corev1.ResourceList
	RuntimeConfigmap        string
	ImagePullSecretName     string
	ImagePullAccount        string
	RequeueDuration         time.Duration
	DockerCfg
}

type CacheSynchronizer = func(stop <-chan struct{}) bool

type Cfg struct {
	rtClient.Client
	*runtime.Scheme
	CacheSynchronizer
	Log logr.Logger
	record.EventRecorder
}

func NewFunctionReconciler(
	cfg *Cfg,
	fnCfg *FnReconcilerCfg) *FunctionReconciler {
	return &FunctionReconciler{
		scheme:                  cfg.Scheme,
		log:                     cfg.Log,
		recorder:                cfg.EventRecorder,
		maxConcurrentReconciles: fnCfg.MaxConcurrentReconciles,
		cacheSynchronizer:       cfg.CacheSynchronizer,
		requeue:                 fnCfg.RequeueDuration,
		cmHelper:                newConfigMapHelper(cfg.Client),
		fnHelper: &fnHelper{
			client:           cfg.Client,
			runtimeConfigmap: fnCfg.RuntimeConfigmap,
		},
		svcHelper: newServiceHelper(cfg.Client),
		regHelper: &registryHelper{
			dockerRegistryFQDN:            fnCfg.DockerRegistryFqdn,
			dockerRegistryPort:            fnCfg.DockerRegistryPort,
			dockerRegistryExternalAddress: fnCfg.DockerRegistryExternalAddress,
			dockerRegistryName:            fnCfg.DockerRegistryName,
		},
		rtHelper: &rtHelper{
			secret:           fnCfg.ImagePullSecretName,
			serviceAccount:   fnCfg.ImagePullAccount,
			runtimeConfigmap: fnCfg.RuntimeConfigmap,
		},
		trHelper: newTaskRunHelper(
			cfg.Client,
			fnCfg.Limits,
			fnCfg.Requests),
	}
}

// FunctionReconciler is the controller.Reconciler implementation for Function objects.
type FunctionReconciler struct {
	scheme                  *runtime.Scheme
	maxConcurrentReconciles int
	cacheSynchronizer       func(stop <-chan struct{}) bool
	log                     logr.Logger
	recorder                record.EventRecorder
	requeue                 time.Duration
	regHelper               RegistryHelper
	rtHelper                RuntimeHelper
	cmHelper                ConfigMapHelper
	fnHelper                FunctionHelper
	trHelper                TaskRunHelper
	svcHelper               ServiceHelper
}

// Reconcile reads that state of the cluster for a Function object and makes changes based on the state read
// and what is in the Function.Spec
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;watch;update;list
// +kubebuilder:rbac:groups="apps;extensions",resources=deployments,verbs=create;get;watch;update;delete;list;update;patch
// +kubebuilder:rbac:groups="serverless.kyma-project.io",resources=functions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="serverless.kyma-project.io",resources=functions/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="serving.knative.dev",resources=services;routes;configurations;revisions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="tekton.dev",resources=tasks;taskruns,verbs=get;list;watch;create;update;patch;delete;deletecollection
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch

func (r *FunctionReconciler) Reconcile(
	req reconcile.Request) (reconcile.Result, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fn := &serverless.Function{}
	err := r.fnHelper.Get(ctx, req.NamespacedName, fn)

	if apierrors.IsNotFound(err) {
		return reconcile.Result{}, nil
	}

	if err != nil {
		return reconcile.Result{}, err
	}

	newStatus := r.handle(ctx, fn)
	// do nothing if status was not changed
	if newStatus == nil {
		return reconcile.Result{}, nil
	}
	// requeue will be done automatically
	err = r.updateStatus(ctx, fn, newStatus)
	if err != nil {
		return reconcile.Result{}, err
	}

	// r.recordPhaseChange(fn, Normal, newStatus.Phase)

	return reconcile.Result{}, nil
}

func fnRecLogger(log logr.Logger, fn *serverless.Function) logr.Logger {
	return log.WithValues(
		"kind", fn.GetObjectKind().GroupVersionKind().Kind,
		"name", fn.GetName(),
		"namespace", fn.GetNamespace(),
		"imageTag", fn.Status.ImageTag,
		"fnUuid", fmt.Sprintf("%s", fn.UID),
	)
}

func (r *FunctionReconciler) handle(
	ctx context.Context,
	fn *serverless.Function) *serverless.FunctionStatus {
	log := fnRecLogger(r.log, fn)

	if fn.Status.ObservedGeneration != fn.GetGeneration() {
		log.Info("created or updated")
		return fn.FunctionStatusInitializing()
	}

	if fn.Status.Phase == serverless.FunctionPhaseInitializing {
		log.Info("in initializing phase")
		return r.handleInitializing(ctx, fn, log)
	}

	if fn.Status.Phase == serverless.FunctionPhaseBuilding {
		log.Info("in building phase")
		return r.handleBuilding(ctx, fn, log)
	}

	if fn.Status.Phase == serverless.FunctionPhaseDeploying {
		log.Info("in deploying phase")
		return r.handleDeploying(ctx, fn, log)
	}

	return nil
}

// function is in Initializing phase - this means:
// - either: it's just have been created (the config map will not be found)
// - or: it's been updated; in this case:
// 		- either: it requires image rebuild
//		- or: redeployment
func (r *FunctionReconciler) handleInitializing(
	ctx context.Context,
	fn *serverless.Function,
	log logr.Logger) *serverless.FunctionStatus {
	// config map containing configuration of the lambda function
	var cm corev1.ConfigMap
	err := r.cmHelper.Fetch(ctx, fn.LabelSelector(), &cm)
	newImageTag := uuid.New().String()

	if err == nil {
		log.WithValues(
			"newImageTag", newImageTag,
			"configmapName", cm.GetName()).
			Info("updating function")
		return r.handleInitializingUpdateFunction(ctx, fn, log, &cm, newImageTag)
	}

	if !apierrors.IsNotFound(err) {
		log.Error(err, "error while getting function's config map")
		return fn.FunctionStatusGetConfigFailed(err)
	}

	return r.handleInitializingNewFunction(ctx, fn, log, newImageTag)
}

const (
	maxNameLength          = 63
	randomLength           = 5
	maxGeneratedNameLength = maxNameLength - randomLength
)

func generateName(base string) string {
	if len(base) > maxGeneratedNameLength {
		base = base[:maxGeneratedNameLength]
	}
	return fmt.Sprintf("%s%s", base, utilrand.String(randomLength))
}

func (r *FunctionReconciler) handleInitializingNewFunction(
	ctx context.Context,
	fn *serverless.Function,
	log logr.Logger,
	imageTag string) *serverless.FunctionStatus {
	cm := configMap(fn)
	imageName := r.regHelper.BuildImageName(fn.Name, fn.Namespace, imageTag)
	limits := r.trHelper.Limits()
	requests := r.trHelper.Requests()
	serviceAccount := r.rtHelper.ServiceAccount()
	configmap := r.rtHelper.RuntimeConfigmap()

	tr := newTaskRunFromFn(fn, imageName, imageTag, serviceAccount, cm.Name, configmap, limits, requests)

	log.Info("setting owner reference of config map")
	// the owner reference of config map needs to be set to this function
	err := controllerutil.SetControllerReference(fn, cm, r.scheme)
	if err != nil {
		log.Error(err, "while setting owner reference of config map")
		return fn.ConditionReasonCreateConfigFailed(err)
	}

	log.Info("creating config map")
	// create config map
	err = r.cmHelper.Create(ctx, cm)
	if err != nil {
		log.Error(err, "error creating config map")
		return fn.ConditionReasonCreateConfigFailed(err)
	}

	log.Info("setting owner reference of task run")
	// the owner reference of task run needs to be set to this function
	err = controllerutil.SetControllerReference(fn, tr, r.scheme)
	if err != nil {
		log.Error(err, "setting owner reference of task run failed")
		return fn.ConditionReasonCreateConfigFailed(err)
	}

	// the task run needs to be created
	log.Info("creating task run")
	err = r.trHelper.Create(ctx, tr)
	if err != nil {
		log.Error(err, "creating task run failed")
		return fn.ConditionReasonCreateConfigFailed(err)
	}

	return fn.ConditionReasonCreateConfigSucceeded(imageTag)
}

type rebuildImg bool

const (
	rebuildImgRequired    rebuildImg = true
	rebuildImgInessential rebuildImg = false
)

func synchronizeMapEntry(
	cm *corev1.ConfigMap,
	key, expected string) rebuildImg {
	if cm.Data == nil {
		cm.Data = map[string]string{
			key: expected,
		}
		return rebuildImgRequired
	}

	actual, found := cm.Data[key]
	if !found {
		cm.Data[key] = expected
		return rebuildImgRequired
	}

	if strings.EqualFold(actual, expected) {
		return rebuildImgInessential
	}

	cm.Data[key] = expected
	return rebuildImgRequired
}

func checkImgRebuild(
	fn *serverless.Function,
	cm *corev1.ConfigMap) rebuildImg {
	rebuildImage := rebuildImgInessential

	for key, value := range map[string]string{
		ConfigMapFunction: fn.Spec.Function,
		ConfigMapDeps:     fn.GetSanitizedDeps(),
	} {
		result := synchronizeMapEntry(cm, key, value)
		if result == rebuildImgRequired {
			rebuildImage = rebuildImgRequired
		}
	}
	rebuildImage = rebuildImage || conditionCheck(fn.Status.Conditions)
	return rebuildImage
}

func (r *FunctionReconciler) handleInitializingUpdateFunction(
	ctx context.Context,
	fn *serverless.Function,
	log logr.Logger,
	cm *corev1.ConfigMap,
	newImageTag string) *serverless.FunctionStatus {
	cmCopy := cm.DeepCopy()

	// determine if controller needs to rebuild the docker image
	rebuildImage := checkImgRebuild(fn, cmCopy)

	if rebuildImage == rebuildImgInessential {
		// the changes do not have impact on docker image;
		// the lambda must be redeployed but the tusk run may
		// still be building; function has to be in building phase;
		log.Info("function update does not require image rebuild")
		return fn.FunctionStatusUpdateRuntimeConfig()
	}

	log.Info("updating function config map")
	err := r.cmHelper.Update(ctx, cmCopy)
	if err != nil {
		log.WithValues("configmapName", cmCopy.Name).
			Error(err, "update function config map failed")

		return fn.FunctionPhaseFailed(
			serverless.ConditionReasonUpdateConfigFailed,
			err.Error(),
		)
	}

	log.Info("delete all task runs associated with the function")
	err = r.trHelper.DeleteAll(ctx, fn.Namespace, fn.LabelSelector())
	if err != nil {
		log.Error(err, "failed to delete task runs associated with the function")
		return fn.FunctionPhaseFailed(
			serverless.ConditionReasonUpdateConfigFailed,
			err.Error())
	}

	log.Info("create TaskRun associated with the function")
	imageName := r.regHelper.BuildImageName(fn.Name, fn.Namespace, newImageTag)

	limits, requests := r.trHelper.Limits(), r.trHelper.Requests()
	srcConfigmap, rtmConfigmap := cm.Name, r.rtHelper.RuntimeConfigmap()

	tr := newTaskRunFromFn(
		fn,
		imageName,
		newImageTag,
		r.rtHelper.ServiceAccount(),
		srcConfigmap,
		rtmConfigmap,
		limits, requests)

	err = controllerutil.SetControllerReference(fn, tr, r.scheme)
	if err != nil {
		log.Error(err, "create TaskRun associated with the function failed")
		return fn.FunctionStatusUpdateConfigFailed(err)
	}

	err = r.trHelper.Create(ctx, tr)
	if err != nil {
		log.Error(err, "create TaskRun associated with the function failed")
		return fn.FunctionStatusUpdateConfigFailed(err)
	}

	return fn.FunctionStatusUpdateConfigSucceeded(newImageTag)
}

func (r *FunctionReconciler) handleBuilding(
	ctx context.Context,
	fn *serverless.Function,
	log logr.Logger) *serverless.FunctionStatus {
	tr, err := r.trHelper.Fetch(ctx, fn.ImgLabelSelector())
	if err != nil {
		log.Error(err, "fetching TaskRun failed")
		return fn.FunctionPhaseFailed(
			serverless.ConditionReasonBuildFailed,
			err.Error(),
		)
	}

	if tr == nil {
		err := funcerr.NewInvalidState("invalid state - unable to associate image to function")
		log.Error(err, "fetching task run failed")
		return fn.FunctionPhaseFailed(
			serverless.ConditionReasonBuildFailed,
			err.Error(),
		)
	}

	taskRunCondition := getTaskRunCondition(tr)

	if taskRunCondition == TaskRunConditionSucceeded {
		log.Info("task run succeeded")
		return fn.FunctionStatusBuildSucceed()
	}

	if taskRunCondition == TaskRunConditionRunning {
		log.Info("task run is running")
		return nil
	}

	err = funcerr.NewInvalidState(fmt.Sprintf("task run: %s.%s failed", tr.Namespace, tr.Name))
	log.Error(err, "build failed")
	return fn.FunctionPhaseFailed(
		serverless.ConditionReasonBuildFailed,
		taskRunCondition.String(),
	)
}

// updates status of given function; to not operate directly on
// cache object, function needs to be copied
func (r *FunctionReconciler) updateStatus(
	ctx context.Context,
	fn *serverless.Function,
	newStatus *serverless.FunctionStatus) error {
	fnCopy := fn.DeepCopy()
	fnCopy.Status = *newStatus
	err := retry.RetryOnConflict(
		retry.DefaultRetry,
		func() error {
			instance := &serverless.Function{}
			err := r.fnHelper.Get(ctx, fn.NamespacedName(), instance)
			if err != nil {
				if apierrors.IsNotFound(err) {
					return nil
				}
				// Error reading the object - requeue the request.
				return err
			}
			err = r.fnHelper.UpdateStatus(ctx, fnCopy)
			if err != nil && apierrors.IsConflict(err) {
				r.cacheSynchronizer(ctx.Done())
			}
			return err
		})
	return err
}

func (r *FunctionReconciler) handleDeploying(
	ctx context.Context,
	fn *serverless.Function,
	log logr.Logger) *serverless.FunctionStatus {
	svc, err := r.svcHelper.Fetch(ctx, fn.LabelSelector())
	if err != nil {
		log.Error(err, "fetching knative service failed")
		return fn.FunctionPhaseFailed(serverless.ConditionReasonDeployFailed, err.Error())
	}

	// create new knative serving
	if svc == nil {
		log.Info("creating new knative service")
		imgName := r.regHelper.ServiceImageName(fn.Name, fn.Namespace, fn.Status.ImageTag)
		return r.handleDeployingNewService(ctx, fn, log, imgName)
	}

	updateServing, err := shouldUpdateServing(r.log, svc, append(envVarsForRevision, fn.Spec.Env...), fn.Status.ImageTag)
	if err != nil {
		log.Error(err, "unable to detect if knative service should be updated")
		return fn.FunctionPhaseFailed(serverless.ConditionReasonDeployFailed, err.Error())
	}

	if updateServing {
		log.Info("updating knative service")
		return r.handleDeployingUpdateService(ctx, fn, log, svc)
	}

	servingCondition := getSvcConditionStatus(svc)

	if servingCondition == ConditionStatusFailed {
		err := funcerr.NewInvalidState(fmt.Sprintf("knative service: %s.%s failed", svc.Namespace, svc.Name))
		log.Error(err, "deployment failed")
		return fn.FunctionPhaseFailed(serverless.ConditionReasonDeployFailed, err.Error())
	}

	if servingCondition == ConditionStatusSucceeded {
		log.Info("knative service depoyed")
		return fn.FunctionStatusDeploySucceeded()
	}

	log.Info("waiting knative service to be deployed")
	return nil
}

func (r *FunctionReconciler) handleDeployingNewService(
	ctx context.Context,
	fn *serverless.Function,
	log logr.Logger,
	image string) *serverless.FunctionStatus {
	podSpec := newPodSpec(
		image,
		r.rtHelper.Secret(),
		r.rtHelper.ServiceAccount(),
		fn.Spec.Env...,
	)

	fnLabels := applyLabels(string(fn.UID), fn.Status.ImageTag, fn.Labels)

	svc := newService(fn.Name, fn.Namespace, podSpec, fnLabels)

	err := controllerutil.SetControllerReference(fn, svc, r.scheme)
	if err != nil {
		return fn.FunctionPhaseFailed(
			serverless.ConditionReasonDeployFailed,
			err.Error(),
		)
	}

	err = r.svcHelper.Create(ctx, svc)
	if err != nil {
		return fn.FunctionPhaseFailed(
			serverless.ConditionReasonDeployFailed,
			err.Error(),
		)
	}

	return fn.FunctionStatusDeploying(
		serverless.ConditionReasonCreateServiceSucceeded,
		"",
	)
}

func (r *FunctionReconciler) handleDeployingUpdateService(
	ctx context.Context,
	fn *serverless.Function,
	log logr.Logger,
	svc *servingv1.Service) *serverless.FunctionStatus {
	if len(svc.Spec.Template.Spec.PodSpec.Containers) != 1 {
		log.Error(errInvalidPodSpec, "invalid pod specification")
		return fn.FunctionPhaseFailed(
			serverless.ConditionReasonDeployFailed,
			errInvalidPodSpec.Error(),
		)
	}

	svcCopy := svc.DeepCopy()

	// update image tag
	svcCopy.Labels["imageTag"] = fn.Status.ImageTag

	// set up environmental variables
	svcCopy.Spec.Template.Spec.Containers[0].Env = append(envVarsForRevision, fn.Spec.Env...)

	if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		err := r.svcHelper.Update(ctx, svcCopy)
		if err != nil && apierrors.IsConflict(err) {
			r.cacheSynchronizer(ctx.Done())
		}
		return err
	}); err != nil {
		return fn.FunctionPhaseFailed(
			serverless.ConditionReasonDeployFailed,
			err.Error(),
		)
	}

	return fn.ConditionReasonUpdateServiceSucceeded("")
}

func (r *FunctionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&serverless.Function{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&tektonv1alpha1.TaskRun{}).
		Owns(&servingv1.Service{}).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: r.maxConcurrentReconciles,
		}).
		Complete(r)
}
