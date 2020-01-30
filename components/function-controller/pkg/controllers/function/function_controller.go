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

package function

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"strings"

	"crypto/sha256"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"

	tektonv1alpha2 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha2"
	knapis "knative.dev/pkg/apis"
	servingapis "knative.dev/serving/pkg/apis/serving"
	servingv1alpha1 "knative.dev/serving/pkg/apis/serving/v1alpha1"

	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	runtimeUtil "github.com/kyma-project/kyma/components/function-controller/pkg/utils"
)

var log = logf.Log.WithName("function_controller")

// List of annotations set on Knative Serving objects by the Knative Serving admission webhook.
var knativeServingAnnotations = []string{
	servingapis.GroupName + knapis.CreatorAnnotationSuffix,
	servingapis.GroupName + knapis.UpdaterAnnotationSuffix,
}

// Add creates a new Function Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileFunction{Client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("function-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to Function
	err = c.Watch(&source.Kind{Type: &serverlessv1alpha1.Function{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	functionAsOwner := &handler.EnqueueRequestForOwner{
		OwnerType:    &serverlessv1alpha1.Function{},
		IsController: true,
	}

	// Watch for changes to Knative Service
	if err := c.Watch(&source.Kind{Type: &servingv1alpha1.Service{}}, functionAsOwner); err != nil {
		return err
	}

	// Watch for changes to Tekton TaskRun
	if err := c.Watch(&source.Kind{Type: &tektonv1alpha2.TaskRun{}}, functionAsOwner); err != nil {
		return err
	}

	return nil
}

var (
	// name of function config
	fnConfigName = getEnvDefault("CONTROLLER_CONFIGMAP", "fn-config")

	// namespace of function config
	fnConfigNamespace = getEnvDefault("CONTROLLER_CONFIGMAP_NS", "default")
)

// ReconcileFunction is the controller.Reconciler implementation for Function objects.
type ReconcileFunction struct {
	client.Client
	scheme *runtime.Scheme
}

func getEnvDefault(envName, defaultValue string) string {
	// use default value if environment variable is empty
	if value := os.Getenv(envName); value != "" {
		return value
	}
	return defaultValue
}

// Reconcile reads that state of the cluster for a Function object and makes changes based on the state read
// and what is in the Function.Spec
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;watch;update;list
// +kubebuilder:rbac:groups="apps;extensions",resources=deployments,verbs=create;get;watch;update;delete;list;update;patch
// +kubebuilder:rbac:groups="serverless.kyma-project.io",resources=functions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="serverless.kyma-project.io",resources=functions/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="admissionregistration.k8s.io",resources=mutatingwebhookconfigurations;validatingwebhookconfigurations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="serving.knative.dev",resources=services;routes;configurations;revisions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="build.knative.dev",resources=builds;buildtemplates;clusterbuildtemplates;services,verbs=get;list;create;update;delete;patch;watch
// +kubebuilder:rbac:groups="tekton.dev",resources=tasks;taskruns,verbs=get;list;watch;create;update;patch;delete

func (r *ReconcileFunction) Reconcile(req reconcile.Request) (reconcile.Result, error) {
	// Get Function instance
	fn, err := r.getFunctionInstance(req)
	switch {
	case errors.IsNotFound(err):
		// Function was deleted, skip reconciliation
		return reconcile.Result{}, nil

	case err != nil:
		if err := r.updateFunctionStatus(fn, serverlessv1alpha1.FunctionConditionError); err != nil {
			log.Error(err, "Error setting Function status", "namespace", req.Namespace, "name", req.Name)
		}

		log.Error(err, "Error reading Function instance", "namespace", req.Namespace, "name", req.Name)
		return reconcile.Result{}, err
	}

	log.Info("Function instance found", "namespace", fn.Namespace, "name", fn.Name)

	// Initialize Function condition
	if fn.Status.Condition == "" {
		if err = r.updateFunctionStatus(fn, serverlessv1alpha1.FunctionConditionUnknown); err != nil {
			log.Error(err, "Error setting Function status", "namespace", fn.Namespace, "name", fn.Name)
			return reconcile.Result{}, err
		}
	}

	// Get Function Controller Configuration
	fnConfig, err := r.getRuntimeConfig()
	if err != nil {
		log.Error(err, "Error reading controller configuration", "namespace", fnConfigNamespace, "name", fnConfigName)
		return reconcile.Result{}, err
	}

	// Get a new *RuntimeInfo
	// TODO(antoineco): this validation should happen ahead of the reconciliation
	rnInfo, err := runtimeUtil.New(fnConfig)
	if err != nil {
		log.Error(err, "Error creating RuntimeInfo", "namespace", fnConfig.Namespace, "name", fnConfig.Name)
		return reconcile.Result{}, err
	}

	// Synchronize Function ConfigMap
	fnCm, err := r.syncFunctionConfigMap(fn)
	if err != nil {
		if err := r.updateFunctionStatus(fn, serverlessv1alpha1.FunctionConditionError); err != nil {
			log.Error(err, "Error setting Function status", "namespace", fn.Namespace, "name", fn.Name)
		}

		log.Error(err, "Error during sync of the Function ConfigMap", "namespace", fn.Namespace, "name", fn.Name)
		return reconcile.Result{}, err
	}

	// Generate build and image names
	fnSha, err := generateFunctionHash(fnCm)
	if err != nil {
		log.Error(err, "Error generating Function hash", "namespace", fn.Namespace, "name", fn.Name)
		return reconcile.Result{}, err
	}

	// length of the suffix appended to the function name to generate a build name
	const buildNameSuffixLen = 10

	// note: we generate these to ensure a new TaskRun is created every
	// time the Function spec is updated.
	imgName := fmt.Sprintf("%s/%s-%s:%s", rnInfo.RegistryInfo, fn.Namespace, fn.Name, fnSha)
	buildName := fmt.Sprintf("%s-%s", fn.Name, fnSha[:buildNameSuffixLen])
	log.Info("Build info", "namespace", fn.Namespace, "name", fn.Name, "buildName", buildName, "imageName", imgName)

	// Run Function build (Tekton TaskRun)
	fnTr, err := r.buildFunctionImage(rnInfo, fn, imgName, buildName)
	if err != nil {
		if err := r.updateFunctionStatus(fn, serverlessv1alpha1.FunctionConditionError); err != nil {
			log.Error(err, "Error setting Function status", "namespace", fn.Namespace, "name", fn.Name)
		}

		log.Error(err, "Error during Function build", "namespace", fn.Namespace, "name", buildName)
		return reconcile.Result{}, err
	}

	// Serve Function (Knative Service)
	fnKsvc, err := r.serveFunction(rnInfo, fnCm, fn, imgName)
	if err != nil {
		if err := r.updateFunctionStatus(fn, serverlessv1alpha1.FunctionConditionError); err != nil {
			log.Error(err, "Error setting Function status", "namespace", fn.Namespace, "name", fn.Name)
		}

		log.Error(err, "Error during Function serving", "namespace", fn.Namespace, "name", fn.Name)
		return reconcile.Result{}, err
	}

	// Set Function status
	if err := r.setFunctionCondition(fn, fnTr, fnKsvc); err != nil {
		log.Error(err, "Error setting Function status", "namespace", fn.Namespace, "name", fn.Name)
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

// getRuntimeConfig returns the Function Controller ConfigMap from the cluster.
// TODO(antoineco): func duplicated in pkg/webhook/default_server/function/mutating
func (r *ReconcileFunction) getRuntimeConfig() (*corev1.ConfigMap, error) {
	cm := &corev1.ConfigMap{}

	err := r.Get(context.TODO(),
		client.ObjectKey{
			Name:      fnConfigName,
			Namespace: fnConfigNamespace,
		},
		cm,
	)
	if err != nil {
		return nil, err
	}

	return cm, nil
}

// getFunctionInstance gets the Function instance from the cluster.
func (r *ReconcileFunction) getFunctionInstance(req reconcile.Request) (*serverlessv1alpha1.Function, error) {
	fn := &serverlessv1alpha1.Function{}
	if err := r.Get(context.TODO(), req.NamespacedName, fn); err != nil {
		return nil, err
	}

	return fn, nil
}

// createFunctionHandlerMap converts a Function spec to a map that can be
// mounted as individual files inside a runtime container.
func createFunctionHandlerMap(fn *serverlessv1alpha1.Function) map[string]string {
	data := make(map[string]string)
	data["handler"] = "handler.main"
	data["handler.js"] = fn.Spec.Function

	data["package.json"] = "{}"
	if strings.Trim(fn.Spec.Deps, " ") != "" {
		data["package.json"] = fn.Spec.Deps
	}

	return data
}

// syncFunctionConfigMap reconciles the current Function ConfigMap with its
// desired state.
func (r *ReconcileFunction) syncFunctionConfigMap(fn *serverlessv1alpha1.Function) (*corev1.ConfigMap, error) {
	desiredCm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Labels:    fn.Labels,
			Namespace: fn.Namespace,
			Name:      fn.Name,
		},
		Data: createFunctionHandlerMap(fn),
	}

	if err := controllerutil.SetControllerReference(fn, desiredCm, r.scheme); err != nil {
		return nil, fmt.Errorf("error setting controller reference: %s", err)
	}

	currentCm, err := r.getOrCreateFunctionConfigMap(desiredCm)
	if err != nil {
		return nil, err
	}

	return r.updateFunctionConfigMap(currentCm, desiredCm)
}

// getOrCreateFunctionConfigMap returns the existing Function ConfigMap or
// creates it from the given desired state if it does not exist.
func (r *ReconcileFunction) getOrCreateFunctionConfigMap(desiredCm *corev1.ConfigMap) (*corev1.ConfigMap, error) {
	ctx := context.TODO()

	currentCm := &corev1.ConfigMap{}
	err := r.Get(ctx,
		client.ObjectKey{
			Name:      desiredCm.Name,
			Namespace: desiredCm.Namespace,
		},
		currentCm,
	)

	switch {
	case errors.IsNotFound(err):
		// TODO(antoineco): a Kubernetes event would be more suitable than a log entry
		log.Info("Creating Function ConfigMap", "namespace", desiredCm.Namespace, "name", desiredCm.Name)

		if err := r.Create(ctx, desiredCm); err != nil {
			return nil, err
		}
		return desiredCm, nil

	case err != nil:
		return nil, err
	}

	return currentCm, nil
}

// updateFunctionConfigMap reconciles the current Function ConfigMap with its
// desired state.
func (r *ReconcileFunction) updateFunctionConfigMap(currentCm, desiredCm *corev1.ConfigMap) (*corev1.ConfigMap, error) {
	if reflect.DeepEqual(desiredCm.Data, currentCm.Data) {
		return currentCm, nil
	}

	// TODO(antoineco): a Kubernetes event would be more suitable than a log entry
	log.Info("Updating Function ConfigMap", "namespace", desiredCm.Namespace, "name", desiredCm.Name)

	newCm := &corev1.ConfigMap{
		ObjectMeta: desiredCm.ObjectMeta,
		Data:       desiredCm.Data,
	}
	newCm.ResourceVersion = currentCm.ResourceVersion

	if err := r.Update(context.TODO(), newCm); err != nil {
		return nil, err
	}
	return newCm, nil
}

// generateFunctionHash generates a hash from the Function config.
func generateFunctionHash(fnCm *corev1.ConfigMap) (string, error) {
	hash := sha256.New()
	if _, err := hash.Write([]byte(fnCm.Data["handler.js"] + fnCm.Data["package.json"])); err != nil {
		return "", fmt.Errorf("error writing to hash: %s", err)
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// buildFunctionImage creates a container image build.
func (r *ReconcileFunction) buildFunctionImage(rnInfo *runtimeUtil.RuntimeInfo, fn *serverlessv1alpha1.Function,
	imageName, buildName string) (*tektonv1alpha2.TaskRun, error) {

	desiredTr := &tektonv1alpha2.TaskRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:      buildName,
			Namespace: fn.Namespace,
			Labels:    fn.Labels,
		},
		Spec: *runtimeUtil.GetBuildTaskRunSpec(rnInfo, fn, imageName),
	}

	if err := controllerutil.SetControllerReference(fn, desiredTr, r.scheme); err != nil {
		return nil, fmt.Errorf("error setting controller reference: %s", err)
	}

	// note: a TaskRun must be recreated upon changes, so we only support
	// the creation of new TaskRuns, not the update after creation.
	return r.getOrCreateFunctionBuildTaskRun(desiredTr, fn)
}

// getOrCreateFunctionBuildTaskRun returns the existing Function build TaskRun
// or creates it from the given desired state if it does not exist.
func (r *ReconcileFunction) getOrCreateFunctionBuildTaskRun(desiredTr *tektonv1alpha2.TaskRun,
	fn *serverlessv1alpha1.Function) (*tektonv1alpha2.TaskRun, error) {

	ctx := context.TODO()

	currentTr := &tektonv1alpha2.TaskRun{}
	err := r.Get(ctx,
		client.ObjectKey{
			Name:      desiredTr.Name,
			Namespace: desiredTr.Namespace,
		},
		currentTr,
	)

	switch {
	case errors.IsNotFound(err):
		// TODO(antoineco): a Kubernetes event would be more suitable than a log entry
		log.Info("Creating Function build TaskRun", "namespace", desiredTr.Namespace, "name", desiredTr.Name)

		if err := r.Create(ctx, desiredTr); err != nil {
			return nil, err
		}

		// TODO(antoineco): it would be enough to compute the status ONCE, at the end of the sync
		if err := r.updateFunctionStatus(fn, serverlessv1alpha1.FunctionConditionBuilding); err != nil {
			log.Error(err, "Error setting Function status", "namespace", fn.Namespace, "name", fn.Name)
		}

		return desiredTr, nil

	case err != nil:
		return nil, err
	}

	return currentTr, nil
}

// serveFunction creates a Knative Service for serving a Function.
func (r *ReconcileFunction) serveFunction(rnInfo *runtimeUtil.RuntimeInfo, foundCm *corev1.ConfigMap,
	fn *serverlessv1alpha1.Function, imageName string) (*servingv1alpha1.Service, error) {

	ctx := context.TODO()

	desiredKsvc := &servingv1alpha1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Labels:    fn.Labels,
			Namespace: fn.Namespace,
			Name:      fn.Name,
		},
		Spec: runtimeUtil.GetServiceSpec(imageName, *fn, rnInfo),
	}

	if err := controllerutil.SetControllerReference(fn, desiredKsvc, r.scheme); err != nil {
		return nil, err
	}

	// Check if the Serving object (serving the function) already exists, if not create a new one.
	currentKsvc := &servingv1alpha1.Service{}
	err := r.Get(ctx,
		client.ObjectKey{
			Name:      desiredKsvc.Name,
			Namespace: desiredKsvc.Namespace,
		},
		currentKsvc,
	)

	switch {
	case errors.IsNotFound(err):
		// TODO(antoineco): a Kubernetes event would be more suitable than a log entry
		log.Info("Creating Knative Service", "namespace", desiredKsvc.Namespace, "name", desiredKsvc.Name)

		if err := r.Create(ctx, desiredKsvc); err != nil {
			return nil, err
		}

		// TODO(antoineco): it would be enough to compute the status ONCE, at the end of the sync
		if err = r.updateFunctionStatus(fn, serverlessv1alpha1.FunctionConditionDeploying); err != nil {
			log.Error(err, "Error setting Function status", "namespace", fn.Namespace, "name", fn.Name)
		}

		return desiredKsvc, nil

	case err != nil:
		return nil, err
	}

	if runtimeUtil.Semantic.DeepEqual(desiredKsvc, currentKsvc) {
		return currentKsvc, nil
	}

	// TODO(antoineco): a Kubernetes event would be more suitable than a log entry
	log.Info("Updating Knative Service", "namespace", desiredKsvc.Namespace, "name", desiredKsvc.Name)

	newKsvc := &servingv1alpha1.Service{
		ObjectMeta: desiredKsvc.ObjectMeta,
		Spec:       desiredKsvc.Spec,
	}
	newKsvc.ResourceVersion = currentKsvc.ResourceVersion
	// immutable Knative annotations must be preserved
	for _, ann := range knativeServingAnnotations {
		metav1.SetMetaDataAnnotation(&newKsvc.ObjectMeta, ann, currentKsvc.Annotations[ann])
	}

	if err := r.Update(ctx, newKsvc); err != nil {
		return nil, err
	}

	// TODO(antoineco): it would be enough to compute the status ONCE, at the end of the sync
	if err = r.updateFunctionStatus(fn, serverlessv1alpha1.FunctionConditionDeploying); err != nil {
		log.Error(err, "Error setting Function status", "namespace", fn.Namespace, "name", fn.Name)
	}

	return newKsvc, nil
}

// setFunctionCondition sets the Function condition based on the status of the Knative service.
// A function is running is if the Status of the Knative service has:
// - the last created revision and the last ready revision are the same.
// - the conditions service, route and configuration should have status true and type ready.
// Update the status of the function base on the defined function condition.
// For a function get the status error either the creation or update of the knative service or build must have failed.
func (r *ReconcileFunction) setFunctionCondition(fn *serverlessv1alpha1.Function, tr *tektonv1alpha2.TaskRun,
	ksvc *servingv1alpha1.Service) error {

	// set Function status to error if the TaskRun failed
	for _, c := range tr.Status.Conditions {
		if c.Type == knapis.ConditionSucceeded && c.Status == corev1.ConditionFalse {
			if err := r.updateFunctionStatus(fn, serverlessv1alpha1.FunctionConditionError); err != nil {
				log.Error(err, "Error setting Function status", "namespace", fn.Namespace, "name", fn.Name)
			}

			return nil
		}
	}

	// skip status update if Knative Service has not reported status yet
	if len(ksvc.Status.Conditions) == 0 {
		return nil
	}

	var serviceReady bool
	var configurationsReady bool
	var routesReady bool

	// latest created and ready revisions share the same name.
	if ksvc.Status.LatestCreatedRevisionName == ksvc.Status.LatestReadyRevisionName {
		for _, cond := range ksvc.Status.Conditions {
			if cond.Status == corev1.ConditionTrue {
				if cond.Type == servingv1alpha1.ServiceConditionReady {
					serviceReady = true
				}
				if cond.Type == servingv1alpha1.RouteConditionReady {
					routesReady = true
				}
				if cond.Type == servingv1alpha1.ConfigurationConditionReady {
					configurationsReady = true
				}
			}
		}
	}

	// Update the function status base on the ksvc status
	fnCondition := serverlessv1alpha1.FunctionConditionDeploying
	if configurationsReady && routesReady && serviceReady {
		fnCondition = serverlessv1alpha1.FunctionConditionRunning
	}

	log.Info(fmt.Sprintf("Function status: %s", fnCondition), "namespace", fn.Namespace, "name", fn.Name)

	return r.updateFunctionStatus(fn, fnCondition)
}

// updateFunctionStatus updates the condition of a Function.
func (r *ReconcileFunction) updateFunctionStatus(fn *serverlessv1alpha1.Function, condition serverlessv1alpha1.FunctionCondition) error {
	fn.Status.Condition = condition
	return r.Status().Update(context.TODO(), fn)
}
