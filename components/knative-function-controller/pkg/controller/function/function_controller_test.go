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
	"crypto/sha256"
	"fmt"
	"testing"
	"time"

	duckv1alpha1 "github.com/knative/pkg/apis/duck/v1alpha1"

	buildv1alpha1 "github.com/knative/build/pkg/apis/build/v1alpha1"
	"github.com/knative/pkg/apis"
	duckv1beta1 "github.com/knative/pkg/apis/duck/v1beta1"
	servingv1alpha1 "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	runtimev1alpha1 "github.com/kyma-project/kyma/components/knative-function-controller/pkg/apis/runtime/v1alpha1"
	"github.com/onsi/gomega"
	"github.com/onsi/gomega/gstruct"
	"golang.org/x/net/context"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var c client.Client

const timeout = time.Second * 20

func TestReconcile(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	var depKey = types.NamespacedName{Name: "foo", Namespace: "default"}

	fnCreated := &runtimev1alpha1.Function{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "default",
		},
		Spec: runtimev1alpha1.FunctionSpec{
			Function:            "main() {asdfasdf}",
			FunctionContentType: "plaintext",
			Size:                "L",
			Runtime:             "nodejs6",
		},
	}

	expectedEnv := []corev1.EnvVar{
		{
			Name:  "FUNC_HANDLER",
			Value: "main",
		},
		{
			Name:  "MOD_NAME",
			Value: "handler",
		},
		{
			Name:  "FUNC_TIMEOUT",
			Value: "180",
		},
		{
			Name:  "FUNC_RUNTIME",
			Value: "nodejs8",
		},
		{
			Name:  "FUNC_MEMORY_LIMIT",
			Value: "128Mi",
		},
		{
			Name:  "FUNC_PORT",
			Value: "8080",
		},
		{
			Name:  "NODE_PATH",
			Value: "$(KUBELESS_INSTALL_VOLUME)/node_modules",
		},
	}

	fnConfig := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "fn-config",
			Namespace: "default",
		},
		Data: map[string]string{
			"dockerRegistry":     "test",
			"serviceAccountName": "build-bot",
			"runtimes": `[
				{
					"ID": "nodejs8",
					"DockerFileName": "dockerfile-nodejs8",
				}
			]`,
		},
	}

	// start manager
	mgr, err := manager.New(cfg, manager.Options{})
	g.Expect(err).NotTo(gomega.HaveOccurred())
	c = mgr.GetClient()
	recFn, requests, errors := SetupTestReconcile(newReconciler(mgr))
	g.Expect(add(mgr, recFn)).NotTo(gomega.HaveOccurred())
	stopMgr, mgrStopped := StartTestManager(mgr, g)
	defer func() {
		close(stopMgr)
		mgrStopped.Wait()
	}()

	// create configmap which holds settings required by the function controller
	g.Expect(c.Create(context.TODO(), fnConfig)).NotTo(gomega.HaveOccurred())
	// create the actual function
	g.Expect(c.Create(context.TODO(), fnCreated)).NotTo(gomega.HaveOccurred())
	defer func() {
		_ = c.Delete(context.TODO(), fnCreated)
		_ = c.Delete(context.TODO(), fnConfig)
	}()

	// call reconcile function
	g.Eventually(requests, timeout).Should(gomega.Receive(gomega.Equal(reconcile.Request{NamespacedName: types.NamespacedName{Name: "foo", Namespace: "default"}})))

	// get config map
	functionConfigMap := &corev1.ConfigMap{}
	g.Eventually(func() error { return c.Get(context.TODO(), depKey, functionConfigMap) }, timeout).
		Should(gomega.Succeed())
	g.Expect(functionConfigMap.Data["handler.js"]).To(gomega.Equal(fnCreated.Spec.Function))
	g.Expect(functionConfigMap.Data["package.json"]).To(gomega.Equal("{}"))

	// get service
	service := &servingv1alpha1.Service{}
	g.Eventually(func() error { return c.Get(context.TODO(), depKey, service) }, timeout).
		Should(gomega.Succeed())
	g.Expect(service.Namespace).To(gomega.Equal("default"))

	// ensure container environment variables are correct
	g.Expect(service.Spec.ConfigurationSpec.Template.Spec.RevisionSpec.PodSpec.Containers[0].Env).To(gomega.Equal(expectedEnv))

	// Unique Build name base on function sha
	hash := sha256.New()
	hash.Write([]byte(functionConfigMap.Data["handler.js"] + functionConfigMap.Data["package.json"]))
	functionSha := fmt.Sprintf("%x", hash.Sum(nil))
	shortSha := functionSha[0:10]
	buildName := fmt.Sprintf("%s-%s", fnCreated.Name, shortSha)

	// get the build object
	build := &buildv1alpha1.Build{}
	g.Eventually(func() error {
		return c.Get(context.TODO(), types.NamespacedName{Name: buildName, Namespace: "default"}, build)
	}, timeout).
		Should(gomega.Succeed())

	// get the build template
	buildTemplate := &buildv1alpha1.BuildTemplate{}
	g.Eventually(func() error {
		return c.Get(context.TODO(), types.NamespacedName{Name: "function-kaniko", Namespace: "default"}, buildTemplate)
	}, timeout).
		Should(gomega.Succeed())

	// ensure build template is correct
	// parameters are available
	g.Expect(buildTemplate.Spec.Parameters).To(gomega.ContainElement(
		gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
			"Name": gomega.BeEquivalentTo("IMAGE"),
		}),
	))
	g.Expect(buildTemplate.Spec.Parameters).To(gomega.ContainElement(
		gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
			"Name": gomega.BeEquivalentTo("DOCKERFILE"),
		}),
	))
	// TODO: ignore order
	// ensure build template references correct config map
	var configMapNameNodeJs6 = "dockerfile-nodejs-6"
	var configMapNameNodeJs8 = "dockerfile-nodejs-8"
	g.Expect(buildTemplate.Spec.Volumes[0].ConfigMap.LocalObjectReference.Name).To(gomega.BeEquivalentTo(configMapNameNodeJs6))
	g.Expect(buildTemplate.Spec.Volumes[1].ConfigMap.LocalObjectReference.Name).To(gomega.BeEquivalentTo(configMapNameNodeJs8))

	// g.Expect(service.Spec.RunLatest.Configuration.RevisionTemplate.Spec.Container.Image).To(gomega.HavePrefix("test/default-foo"))
	g.Expect(build.Spec.ServiceAccountName).To(gomega.Equal("build-bot"))
	// g.Expect(service.Spec.RunLatest.Configuration.RevisionTemplate.Spec.Container.Image).To(gomega.HavePrefix("test/default-foo"))

	// ensure that image name referenced in build is used in the service
	g.Expect(len(service.Spec.ConfigurationSpec.Template.Spec.RevisionSpec.PodSpec.Containers)).To(gomega.Equal(1))
	var imageNameService = service.Spec.ConfigurationSpec.Template.Spec.RevisionSpec.PodSpec.Containers[0].Image
	var imageNameBuild string
	for _, argumentSpec := range build.Spec.Template.Arguments {
		if argumentSpec.Name == "IMAGE" {
			imageNameBuild = argumentSpec.Value
			break
		}
	}
	g.Expect(imageNameBuild).To(gomega.Equal(imageNameService))

	// ensure build template has correct destination
	g.Expect(len(buildTemplate.Spec.Steps)).To(gomega.Equal(1))
	g.Expect(buildTemplate.Spec.Steps[0].Args[1]).To(gomega.Equal("--destination=${IMAGE}"))

	// ensure fetched function spec corresponds to created function spec
	fnUpdatedFetched := &runtimev1alpha1.Function{}
	g.Expect(c.Get(context.TODO(), depKey, fnUpdatedFetched)).NotTo(gomega.HaveOccurred())
	g.Expect(fnUpdatedFetched.Spec).To(gomega.Equal(fnCreated.Spec))

	// update function code and add dependencies
	fnUpdated := fnUpdatedFetched.DeepCopy()
	fnUpdated.Spec.Function = `main() {return "bla"}`
	fnUpdated.Spec.Deps = `dependencies`
	g.Expect(c.Update(context.TODO(), fnUpdated)).NotTo(gomega.HaveOccurred())

	// get the updated function and compare spec
	fnUpdatedFetched = &runtimev1alpha1.Function{}
	g.Expect(c.Get(context.TODO(), depKey, fnUpdatedFetched)).NotTo(gomega.HaveOccurred())
	g.Expect(fnUpdatedFetched.Spec).To(gomega.Equal(fnUpdated.Spec))
	// call reconcile function
	g.Eventually(requests, timeout).Should(gomega.Receive(gomega.Equal(reconcile.Request{NamespacedName: types.NamespacedName{Name: "foo", Namespace: "default"}})))

	// get updated function code and ensure it got updated
	cmUpdated := &corev1.ConfigMap{}
	g.Eventually(func() error {
		return c.Get(context.TODO(), depKey, cmUpdated)
	}).Should(gomega.Succeed())
	g.Eventually(func() string {
		c.Get(context.TODO(), depKey, cmUpdated)
		return cmUpdated.Data["handler.js"]
	}, timeout, 1*time.Second).Should(gomega.Equal(fnUpdated.Spec.Function))
	g.Eventually(func() string {
		c.Get(context.TODO(), depKey, cmUpdated)
		return cmUpdated.Data["package.json"]
	}, timeout, 1*time.Second).Should(gomega.Equal(`dependencies`))

	// ensure updated knative service has updated image
	ksvcUpdated := &servingv1alpha1.Service{}
	g.Expect(c.Get(context.TODO(), depKey, ksvcUpdated)).NotTo(gomega.HaveOccurred())
	hash = sha256.New()
	print("cmUpdated: %v", cmUpdated)
	hash.Write([]byte(cmUpdated.Data["handler.js"] + cmUpdated.Data["package.json"]))
	functionSha = fmt.Sprintf("%x", hash.Sum(nil))
	g.Expect(ksvcUpdated.Spec.ConfigurationSpec.Template.Spec.RevisionSpec.PodSpec.Containers[0].Image).
		To(gomega.Equal(fmt.Sprintf("test/%s-%s:%s", "default", "foo", functionSha)))

	g.Expect(fnUpdatedFetched.Status.Condition).To(gomega.Equal(runtimev1alpha1.FunctionConditionDeploying))

	// tests use a shared etcd, we need to clean up
	defer func() {
		_ = c.Delete(context.TODO(), fnCreated)
		_ = c.Delete(context.TODO(), ksvcUpdated)
		_ = c.Delete(context.TODO(), fnConfig)
	}()

	// ensure no errors occurred in reconciler
	g.Eventually(errors).ShouldNot(gomega.Receive(gomega.Succeed()))
}

// Test that deleting a function does not produce any errors
func TestReconcileDeleteFunction(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	objectName := "test-reconcile-delete-function"

	fnCreated := &runtimev1alpha1.Function{
		ObjectMeta: metav1.ObjectMeta{
			Name:      objectName,
			Namespace: "default",
		},
		Spec: runtimev1alpha1.FunctionSpec{
			Function:            "main() {asdfasdf}",
			FunctionContentType: "plaintext",
			Size:                "L",
			Runtime:             "nodejs6",
		},
	}

	fnConfig := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "fn-config",
			Namespace: "default",
		},
		Data: map[string]string{
			"dockerRegistry":     "test",
			"serviceAccountName": "build-bot",
			"runtimes": `[
				{
					"ID": "nodejs8",
					"DockerFileName": "dockerfile-nodejs8",
				}
			]`,
		},
	}

	// start manager
	mgr, err := manager.New(cfg, manager.Options{})
	g.Expect(err).NotTo(gomega.HaveOccurred())
	c = mgr.GetClient()
	recFn, requests, errors := SetupTestReconcile(newReconciler(mgr))
	g.Expect(add(mgr, recFn)).NotTo(gomega.HaveOccurred())
	stopMgr, mgrStopped := StartTestManager(mgr, g)
	defer func() {
		close(stopMgr)
		mgrStopped.Wait()
	}()

	// create configmap which holds settings required by the function controller
	g.Expect(c.Create(context.TODO(), fnConfig)).NotTo(gomega.HaveOccurred())
	// create the actual function
	g.Expect(c.Create(context.TODO(), fnCreated)).NotTo(gomega.HaveOccurred())
	defer func() {
		_ = c.Delete(context.TODO(), fnConfig)
		_ = c.Delete(context.TODO(), fnCreated)
	}()

	// call reconcile function
	fmt.Print("waiting for reconcile function to be called")
	g.Eventually(requests, timeout).Should(gomega.Receive(gomega.Equal(reconcile.Request{NamespacedName: types.NamespacedName{Name: objectName, Namespace: "default"}})))

	// delete the actual function
	g.Expect(c.Delete(context.TODO(), fnCreated)).NotTo(gomega.HaveOccurred())

	// call reconcile function
	fmt.Print("waiting for reconcile function to be called")
	g.Eventually(requests, timeout).Should(gomega.Receive(gomega.Equal(reconcile.Request{NamespacedName: types.NamespacedName{Name: objectName, Namespace: "default"}})))

	// ensure no errors occurred in reconciler
	g.Eventually(errors).ShouldNot(gomega.Receive(gomega.Succeed()))
}

// Test status of newly created function
func TestFunctionConditionNewFunction(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	objectName := "test-condition-new-function"
	mgr, err := manager.New(cfg, manager.Options{})
	g.Expect(err).NotTo(gomega.HaveOccurred())
	c := mgr.GetClient()

	function := runtimev1alpha1.Function{
		ObjectMeta: metav1.ObjectMeta{
			Name:      objectName,
			Namespace: "default",
		},
	}

	reconcileFunction := &ReconcileFunction{Client: c, scheme: scheme.Scheme}

	g.Expect(c.Create(context.TODO(), &function)).Should(gomega.Succeed())

	reconcileFunction.getFunctionCondition(&function)

	// no knative objects present => no function status
	g.Expect(fmt.Sprint(function.Status.Condition)).To(gomega.Equal(""))
}

// Test status of function with errored build
func TestFunctionConditionBuildError(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	objectName := "test-build-error"

	mgr, err := manager.New(cfg, manager.Options{})
	g.Expect(err).NotTo(gomega.HaveOccurred())
	c := mgr.GetClient()
	reconcileFunction := &ReconcileFunction{Client: c, scheme: scheme.Scheme}

	stopMgr, mgrStopped := StartTestManager(mgr, g)
	defer func() {
		close(stopMgr)
		mgrStopped.Wait()
	}()

	function := runtimev1alpha1.Function{
		ObjectMeta: metav1.ObjectMeta{
			Name:      objectName,
			Namespace: "default",
		},
	}

	// Example condition from an errored build (using wrong dockerhub credentials)
	// Conditions:
	//  Last Transition Time:  2019-06-26T15:57:31Z
	//  Message:               build step "build-step-build-and-push" exited with code 1 (image: "docker-pullable://gcr.io/kaniko-project/executor@sha256:d9fe474f80b73808dc12b54f45f5fc90f7856d9fc699d4a5e79d968a1aef1a72"); for logs run: kubectl -n default logs example-build-pod-ed7514 -c build-step-build-and-push
	//  Status:                False
	//  Type:                  Succeeded
	build := buildv1alpha1.Build{
		ObjectMeta: metav1.ObjectMeta{
			Name:      objectName,
			Namespace: "default",
		},
	}

	// create function and build
	g.Expect(c.Create(context.TODO(), &function)).Should(gomega.Succeed())
	g.Expect(c.Create(context.TODO(), &build)).Should(gomega.Succeed())
	defer func() {
		_ = c.Delete(context.TODO(), &function)
		_ = c.Delete(context.TODO(), &build)
	}()

	// get build and update status
	foundBuild := buildv1alpha1.Build{}
	g.Eventually(func() error {
		return c.Get(context.TODO(), types.NamespacedName{Name: objectName, Namespace: "default"}, &foundBuild)
	}).Should(gomega.Succeed())
	foundBuild.Status = buildv1alpha1.BuildStatus{
		Status: duckv1alpha1.Status{
			Conditions: []duckv1alpha1.Condition{
				{
					Type:   duckv1alpha1.ConditionSucceeded,
					Status: corev1.ConditionFalse,
				},
			},
		},
	}
	g.Expect(c.Status().Update(context.TODO(), &foundBuild)).Should(gomega.Succeed())

	g.Eventually(func() runtimev1alpha1.FunctionCondition {
		reconcileFunction.getFunctionCondition(&function)
		return function.Status.Condition
	}).Should(gomega.Equal(runtimev1alpha1.FunctionConditionError))
}

func TestFunctionConditionServiceSuccess(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	objectName := "test-service-success"

	mgr, err := manager.New(cfg, manager.Options{})
	g.Expect(err).NotTo(gomega.HaveOccurred())
	c := mgr.GetClient()
	reconcileFunction := &ReconcileFunction{Client: c, scheme: scheme.Scheme}

	stopMgr, mgrStopped := StartTestManager(mgr, g)
	defer func() {
		close(stopMgr)
		mgrStopped.Wait()
	}()

	function := runtimev1alpha1.Function{
		ObjectMeta: metav1.ObjectMeta{
			Name:      objectName,
			Namespace: "default",
		},
	}

	service := servingv1alpha1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      objectName,
			Namespace: "default",
		},
	}

	// create function and build
	g.Expect(c.Create(context.TODO(), &function)).Should(gomega.Succeed())
	g.Expect(c.Create(context.TODO(), &service)).Should(gomega.Succeed())
	defer func() {
		_ = c.Delete(context.TODO(), &function)
		_ = c.Delete(context.TODO(), &service)
	}()

	// get service and update status
	foundService := servingv1alpha1.Service{}
	g.Eventually(func() error {
		return c.Get(context.TODO(), types.NamespacedName{Name: objectName, Namespace: "default"}, &foundService)
	}).Should(gomega.Succeed())
	foundService.Status = servingv1alpha1.ServiceStatus{
		ConfigurationStatusFields: servingv1alpha1.ConfigurationStatusFields{
			LatestCreatedRevisionName: "foo",
			LatestReadyRevisionName:   "foo",
		},
		Status: duckv1beta1.Status{
			Conditions: []apis.Condition{
				{
					Type:   servingv1alpha1.ServiceConditionReady,
					Status: corev1.ConditionTrue,
				},
				{
					Type:   servingv1alpha1.RouteConditionReady,
					Status: corev1.ConditionTrue,
				},
				{
					Type:   servingv1alpha1.ConfigurationConditionReady,
					Status: corev1.ConditionTrue,
				},
			},
		},
	}
	g.Expect(c.Status().Update(context.TODO(), &foundService)).Should(gomega.Succeed())

	g.Eventually(func() runtimev1alpha1.FunctionCondition {
		reconcileFunction.getFunctionCondition(&function)
		return function.Status.Condition
	}).Should(gomega.Equal(runtimev1alpha1.FunctionConditionRunning))
}

func TestFunctionConditionServiceError(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	objectName := "test-service-error"

	mgr, err := manager.New(cfg, manager.Options{})
	g.Expect(err).NotTo(gomega.HaveOccurred())
	c := mgr.GetClient()
	reconcileFunction := &ReconcileFunction{Client: c, scheme: scheme.Scheme}

	stopMgr, mgrStopped := StartTestManager(mgr, g)
	defer func() {
		close(stopMgr)
		mgrStopped.Wait()
	}()

	function := runtimev1alpha1.Function{
		ObjectMeta: metav1.ObjectMeta{
			Name:      objectName,
			Namespace: "default",
		},
	}

	service := servingv1alpha1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      objectName,
			Namespace: "default",
		},
	}

	// create function and build
	g.Expect(c.Create(context.TODO(), &function)).Should(gomega.Succeed())
	g.Expect(c.Create(context.TODO(), &service)).Should(gomega.Succeed())
	defer func() {
		_ = c.Delete(context.TODO(), &function)
		_ = c.Delete(context.TODO(), &service)
	}()

	// get service and update status
	foundService := servingv1alpha1.Service{}
	g.Eventually(func() error {
		return c.Get(context.TODO(), types.NamespacedName{Name: objectName, Namespace: "default"}, &foundService)
	}).Should(gomega.Succeed())
	foundService.Status = servingv1alpha1.ServiceStatus{
		ConfigurationStatusFields: servingv1alpha1.ConfigurationStatusFields{
			LatestCreatedRevisionName: "foo",
			LatestReadyRevisionName:   "foo",
		},
		Status: duckv1beta1.Status{
			Conditions: []apis.Condition{
				{
					Type:   servingv1alpha1.ServiceConditionReady,
					Status: corev1.ConditionFalse,
				},
				{
					Type:   servingv1alpha1.RouteConditionReady,
					Status: corev1.ConditionFalse,
				},
				{
					Type:   servingv1alpha1.ConfigurationConditionReady,
					Status: corev1.ConditionFalse,
				},
			},
		},
	}
	g.Expect(c.Status().Update(context.TODO(), &foundService)).Should(gomega.Succeed())

	g.Eventually(func() runtimev1alpha1.FunctionCondition {
		reconcileFunction.getFunctionCondition(&function)
		return function.Status.Condition
	}).Should(gomega.Equal(runtimev1alpha1.FunctionConditionDeploying))
}

func TestCreateFunctionHandlerMap(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	functionCode := "some function code"
	functionDependecies := "some function dependencies"
	function := runtimev1alpha1.Function{Spec: runtimev1alpha1.FunctionSpec{
		Function: functionCode,
		Deps:     functionDependecies,
	},
	}
	functionHandlerMap := createFunctionHandlerMap(&function)

	mapx := map[string]string{
		"handler":      "handler.main",
		"handler.js":   functionCode,
		"package.json": functionDependecies,
	}
	g.Expect(functionHandlerMap).To(gomega.Equal(mapx))
}

func TestCreateFunctionHandlerMapNoDependencies(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	functionCode := "some function code"
	function := runtimev1alpha1.Function{Spec: runtimev1alpha1.FunctionSpec{
		Function: functionCode,
	},
	}
	functionHandlerMap := createFunctionHandlerMap(&function)

	mapx := map[string]string{
		"handler":      "handler.main",
		"handler.js":   functionCode,
		"package.json": "{}",
	}
	g.Expect(functionHandlerMap).To(gomega.Equal(mapx))
}
