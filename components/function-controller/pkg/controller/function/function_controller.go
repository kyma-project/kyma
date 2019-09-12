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

	buildv1alpha1 "github.com/knative/build/pkg/apis/build/v1alpha1"
	duckv1alpha1 "github.com/knative/pkg/apis/duck/v1alpha1"
	servingv1alpha1 "github.com/knative/serving/pkg/apis/serving/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"

	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	runtimeUtil "github.com/kyma-project/kyma/components/function-controller/pkg/utils"
)

var log = logf.Log.WithName("function_controller")

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

	// Watch for changes to Knative Build
	if err := c.Watch(&source.Kind{Type: &buildv1alpha1.Build{}}, functionAsOwner); err != nil {
		return err
	}

	return nil
}

var (
	// name of function config
	fnConfigName = getEnvDefault("CONTROLLER_CONFIGMAP", "fn-config")

	// namespace of function config
	fnConfigNamespace = getEnvDefault("CONTROLLER_CONFIGMAP_NS", "default")

	// name of build-template
	buildTemplateName = getEnvDefault("BUILD_TEMPLATE", "function-kaniko")

	// build and push step name
	buildAndPushStep = "build-step-build-and-push"
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

	// Synchronize Function BuildTemplate
	if err := r.syncFunctionBuildTemplate(fn, rnInfo); err != nil {
		if err := r.updateFunctionStatus(fn, serverlessv1alpha1.FunctionConditionError); err != nil {
			log.Error(err, "Error setting Function status", "namespace", fn.Namespace, "name", fn.Name)
		}

		log.Error(err, "Error during sync of the Function BuildTemplate", "namespace", fn.Namespace, "name", buildTemplateName)
		return reconcile.Result{}, err
	}

	// Generate build and image names
	fnSha, err := generateFunctionHash(fnCm)
	if err != nil {
		log.Error(err, "Error generating Function hash", "namespace", fn.Namespace, "name", fn.Name)
		return reconcile.Result{}, err
	}
	shortSha := fnSha[:10]

	imgName := fmt.Sprintf("%s/%s-%s:%s", rnInfo.RegistryInfo, fn.Namespace, fn.Name, fnSha)
	buildName := fmt.Sprintf("%s-%s", fn.Name, shortSha)
	log.Info("Build info", "namespace", fn.Namespace, "name", fn.Name, "buildName", buildName, "imageName", imgName)

	err = r.buildFunctionImage(rnInfo, fn, imgName, buildName)
	if err != nil {
		if err := r.updateFunctionStatus(fn, serverlessv1alpha1.FunctionConditionError); err != nil {
			log.Error(err, "Error setting Function status", "namespace", fn.Namespace, "name", fn.Name)
		}

		log.Error(err, "Error during Function build", "namespace", fn.Namespace, "name", buildName)
		return reconcile.Result{}, err
	}

	_, err = r.serveFunction(rnInfo, fnCm, fn, imgName)
	if err != nil {
		if err := r.updateFunctionStatus(fn, serverlessv1alpha1.FunctionConditionError); err != nil {
			log.Error(err, "Error setting Function status", "namespace", fn.Namespace, "name", fn.Name)
		}

		log.Error(err, "Error during Function serving", "namespace", fn.Namespace, "name", fn.Name)
		return reconcile.Result{}, err
	}

	r.setFunctionCondition(fn)

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

// syncFunctionBuildTemplate reconciles the current Function BuildTemplate with its
// desired state.
func (r *ReconcileFunction) syncFunctionBuildTemplate(fn *serverlessv1alpha1.Function, ri *runtimeUtil.RuntimeInfo) error {
	desiredBt := &buildv1alpha1.BuildTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      buildTemplateName,
			Namespace: fn.Namespace,
		},
		Spec: runtimeUtil.GetBuildTemplateSpec(ri),
	}

	if err := controllerutil.SetControllerReference(fn, desiredBt, r.scheme); err != nil {
		return fmt.Errorf("error setting controller reference: %s", err)
	}

	currentBt, err := r.getOrCreateFunctionBuildTemplate(desiredBt)
	if err != nil {
		if err := r.updateFunctionStatus(fn, serverlessv1alpha1.FunctionConditionError); err != nil {
			log.Error(err, "Error setting Function status", "namespace", fn.Namespace, "name", fn.Name)
		}

		log.Error(err, "Error during sync of the Function BuildTemplate", "namespace", desiredBt.Namespace, "name", desiredBt.Name)

		return err
	}

	if _, err = r.updateFunctionBuildTemplate(currentBt, desiredBt); err != nil {
		if err := r.updateFunctionStatus(fn, serverlessv1alpha1.FunctionConditionError); err != nil {
			log.Error(err, "Error setting Function status", "namespace", fn.Namespace, "name", fn.Name)
		}

		log.Error(err, "Error during sync of the Function BuildTemplate", "namespace", desiredBt.Namespace, "name", desiredBt.Name)
		return err
	}

	return nil
}

// getOrCreateFunctionBuildTemplate returns the existing Function BuildTemplate
// or creates it from the given desired state if it does not exist.
func (r *ReconcileFunction) getOrCreateFunctionBuildTemplate(desiredBt *buildv1alpha1.BuildTemplate) (*buildv1alpha1.BuildTemplate, error) {
	ctx := context.TODO()

	foundBt := &buildv1alpha1.BuildTemplate{}
	err := r.Get(ctx,
		client.ObjectKey{
			Name:      desiredBt.Name,
			Namespace: desiredBt.Namespace,
		},
		foundBt,
	)

	switch {
	case errors.IsNotFound(err):
		log.Info("Creating Function BuildTemplate", "namespace", desiredBt.Namespace, "name", desiredBt.Name)

		if err := r.Create(ctx, desiredBt); err != nil {
			return nil, err
		}
		return desiredBt, nil

	case err != nil:
		return nil, err
	}

	return foundBt, nil
}

// updateFunctionBuildTemplate reconciles the current Function BuildTemplate
// with its desired state.
func (r *ReconcileFunction) updateFunctionBuildTemplate(currentBt, desiredBt *buildv1alpha1.BuildTemplate) (*buildv1alpha1.BuildTemplate, error) {
	if reflect.DeepEqual(desiredBt.Spec, currentBt.Spec) {
		return currentBt, nil
	}

	log.Info("Updating Function BuildTemplate", "namespace", desiredBt.Namespace, "name", desiredBt.Name)

	newBt := &buildv1alpha1.BuildTemplate{
		ObjectMeta: desiredBt.ObjectMeta,
		Spec:       desiredBt.Spec,
	}
	newBt.ResourceVersion = currentBt.ResourceVersion

	if err := r.Update(context.TODO(), newBt); err != nil {
		return nil, err
	}
	return newBt, nil
}

// buildFunctionImage creates a container image build.
func (r *ReconcileFunction) buildFunctionImage(rnInfo *runtimeUtil.RuntimeInfo, fn *serverlessv1alpha1.Function, imageName string, buildName string) error {
	// Create a new Build data structure
	deployBuild := runtimeUtil.GetBuildResource(rnInfo, fn, imageName, buildName)

	if err := controllerutil.SetControllerReference(fn, deployBuild, r.scheme); err != nil {
		return err
	}

	// Check if the build object (building the function) already exists, if not create a new one.
	foundBuild := &buildv1alpha1.Build{}
	err := r.Get(context.TODO(), types.NamespacedName{Name: deployBuild.Name, Namespace: deployBuild.Namespace}, foundBuild)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating Knative Build", "namespace", deployBuild.Namespace, "name", deployBuild.Name)
		err = r.Create(context.TODO(), deployBuild)
		if err != nil {
			return err
		}

		err = r.updateFunctionStatus(fn, serverlessv1alpha1.FunctionConditionBuilding)
		if err != nil {
			if errors.IsNotFound(err) {
				return nil
			}
			return err
		}
		return nil

	} else if err != nil {
		log.Error(err, "Error while trying to create Knative Build", "namespace", deployBuild.Namespace, "name", deployBuild.Name)
		return err
	}

	// Update Build object
	if !reflect.DeepEqual(deployBuild.Spec, foundBuild.Spec) && !compareBuildImages(foundBuild, imageName) {
		err := r.updateFunctionStatus(fn, serverlessv1alpha1.FunctionConditionUpdating)
		if err != nil {
			if errors.IsNotFound(err) {
				return nil
			}
			return err
		}

		// create new Build with the new updated image
		log.Info("Creating new Knative Build", "namespace", deployBuild.Namespace, "name", deployBuild.Name)
		err = r.Create(context.TODO(), deployBuild)
		if err != nil {
			if errors.IsNotFound(err) {
				return nil
			}
			return err
		}

		log.Info("Updated Knative Build", "namespace", deployBuild.Namespace, "name", deployBuild.Name)

		return nil
	}

	return nil
}

// compareBuildImages returns whether two builds are equal.
func compareBuildImages(foundBuild *buildv1alpha1.Build, imageName string) bool {
	if foundBuild.Spec.Template != nil && len(foundBuild.Spec.Template.Arguments) > 0 {
		args := foundBuild.Spec.Template.Arguments
		for _, arg := range args {
			if arg.Name == "IMAGE" && arg.Value == imageName {
				return true
			}
		}
	}

	return false
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

	if reflect.DeepEqual(desiredKsvc.Spec, currentKsvc.Spec) {
		return currentKsvc, nil
	}

	// TODO(antoineco): a Kubernetes event would be more suitable than a log entry
	log.Info("Updating Knative Service", "namespace", desiredKsvc.Namespace, "name", desiredKsvc.Name)

	newKsvc := &servingv1alpha1.Service{
		ObjectMeta: desiredKsvc.ObjectMeta,
		Spec:       desiredKsvc.Spec,
	}
	newKsvc.ResourceVersion = currentKsvc.ResourceVersion

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
func (r *ReconcileFunction) setFunctionCondition(fn *serverlessv1alpha1.Function) {
	serviceReady := false
	configurationsReady := false
	routesReady := false

	// Get the status of the Build
	foundBuild := &buildv1alpha1.Build{}
	if err := r.Get(context.TODO(), types.NamespacedName{Name: fn.Name, Namespace: fn.Namespace}, foundBuild); ignoreNotFound(err) != nil {
		log.Error(err, "Error while trying to get the Knative Build for the Function Status", "namespace", fn.Namespace, "name", fn.Name)
		return
	}

	// if build show error, set function status to error too
	for _, condition := range foundBuild.Status.Conditions {
		if condition.Type == duckv1alpha1.ConditionSucceeded && condition.Status == corev1.ConditionFalse {
			err := r.updateFunctionStatus(fn, serverlessv1alpha1.FunctionConditionError)
			if err != nil {
				log.Error(err, "Error while trying to update the function Status", "namespace", fn.Namespace, "name", fn.Name)
			}
			return
		}
	}

	// Get Knative Service
	foundService := &servingv1alpha1.Service{}
	if err := r.Get(context.TODO(), types.NamespacedName{Name: fn.Name, Namespace: fn.Namespace}, foundService); err != nil {
		log.Error(err, "Error while trying to get the Knative Service for the function Status", "namespace", fn.Namespace, "name", fn.Name)
		return
	}

	// latest created and ready revisions share the same name.
	if foundService.Status.LatestCreatedRevisionName == foundService.Status.LatestReadyRevisionName {
		// Evaluates the status of the conditions
		if len(foundService.Status.Conditions) == 3 {
			conditions := foundService.Status.Conditions

			for _, cond := range conditions {
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
	}

	// Update the function status base on the ksvc status
	fnCondition := serverlessv1alpha1.FunctionConditionDeploying
	if configurationsReady && routesReady && serviceReady {
		fnCondition = serverlessv1alpha1.FunctionConditionRunning
	}

	log.Info(fmt.Sprintf("Function status: %s", fnCondition), "namespace", fn.Namespace, "name", fn.Name)

	r.updateFunctionStatus(fn, fnCondition)
}

func ignoreNotFound(err error) error {
	if errors.IsNotFound(err) {
		return nil
	}
	return err
}

// updateFunctionStatus updates the condition of a Function.
func (r *ReconcileFunction) updateFunctionStatus(fn *serverlessv1alpha1.Function, condition serverlessv1alpha1.FunctionCondition) error {
	fn.Status.Condition = condition
	return r.Status().Update(context.TODO(), fn)
}
