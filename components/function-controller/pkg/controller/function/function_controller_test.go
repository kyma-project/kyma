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

	"golang.org/x/net/context"

	"github.com/onsi/gomega"
	"github.com/onsi/gomega/gstruct"
	gomegatypes "github.com/onsi/gomega/types"

	tektonv1alpha1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	knapis "knative.dev/pkg/apis"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
	servingv1alpha1 "knative.dev/serving/pkg/apis/serving/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
)

var c client.Client

const timeout = time.Second * 60

func TestReconcile(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	var depKey = types.NamespacedName{Name: "foo", Namespace: "default"}

	fnCreated := &serverlessv1alpha1.Function{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "default",
		},
		Spec: serverlessv1alpha1.FunctionSpec{
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
			"defaults":           `{"size": "S", "runtime": "nodejs8", "timeOut": 180, "funcContentType": "plaintext"}`,
			"runtimes":           `[{"ID": "nodejs8", "dockerfileName": "dockerfile-nodejs8"}, {"ID": "nodejs6", "dockerfileName": "dockerfile-nodejs6"}]`,
			"funcSizes":          `[{"size": "S"}, {"size": "M"}, {"size": "L"}]`,
			"funcTypes":          `[{"type": "plaintext"}, {"type": "base64"}]`,
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

	// get Knative Service
	ksvc := &servingv1alpha1.Service{}
	g.Eventually(func() error { return c.Get(context.TODO(), depKey, ksvc) }, timeout).
		Should(gomega.Succeed())
	g.Expect(ksvc.Namespace).To(gomega.Equal("default"))

	// ensure container environment variables are correct
	g.Expect(ksvc.Spec.ConfigurationSpec.Template.Spec.RevisionSpec.PodSpec.Containers[0].Env).To(gomega.Equal(expectedEnv))

	// Unique Build name base on function sha
	hash := sha256.New()
	hash.Write([]byte(functionConfigMap.Data["handler.js"] + functionConfigMap.Data["package.json"]))
	functionSha := fmt.Sprintf("%x", hash.Sum(nil))
	shortSha := functionSha[:10]
	buildName := fmt.Sprintf("%s-%s", fnCreated.Name, shortSha)

	// get the Task
	task := &tektonv1alpha1.Task{}
	g.Eventually(func() error {
		return c.Get(context.TODO(), types.NamespacedName{Name: "function-build", Namespace: "default"}, task)
	}, timeout).
		Should(gomega.Succeed())

	// ensure Task input params are defined
	var idTaskInputParam gstruct.Identifier = func(element interface{}) string {
		return element.(tektonv1alpha1.ParamSpec).Name
	}
	var matchTaskInputParamType gomegatypes.GomegaMatcher = gstruct.MatchFields(gstruct.IgnoreExtras,
		gstruct.Fields{"Type": gomega.Equal(tektonv1alpha1.ParamTypeString)},
	)
	g.Expect(task.Spec.Inputs.Params).To(gstruct.MatchAllElements(idTaskInputParam, gstruct.Elements{
		"imageName":               matchTaskInputParamType,
		"dockerfileConfigmapName": matchTaskInputParamType,
		"sourceConfigmapName":     matchTaskInputParamType,
	}))

	// get the TaskRun object
	tr := &tektonv1alpha1.TaskRun{}
	g.Eventually(func() error {
		return c.Get(context.TODO(), types.NamespacedName{Name: buildName, Namespace: "default"}, tr)
	}, timeout).
		Should(gomega.Succeed())

	// ensure build template references all ConfigMaps (Dockerfiles, Fn source)
	var idTaskVolumes gstruct.Identifier = func(element interface{}) string {
		return element.(corev1.Volume).Name
	}
	var matchTaskVolumeSource gomegatypes.GomegaMatcher = gstruct.MatchFields(gstruct.IgnoreExtras,
		gstruct.Fields{"VolumeSource": gstruct.MatchFields(gstruct.IgnoreExtras,
			gstruct.Fields{"ConfigMap": gomega.Not(gomega.BeNil())},
		)},
	)
	g.Expect(task.Spec.Volumes).To(gstruct.MatchAllElements(idTaskVolumes, gstruct.Elements{
		"dockerfile-nodejs6": matchTaskVolumeSource,
		"dockerfile-nodejs8": matchTaskVolumeSource,
		"source":             matchTaskVolumeSource,
	}))

	g.Expect(tr.Spec.ServiceAccount).To(gomega.Equal("build-bot"))

	g.Expect(len(ksvc.Spec.ConfigurationSpec.Template.Spec.RevisionSpec.PodSpec.Containers)).To(gomega.Equal(1))

	// ensure that TaskRun params are set to their expected value
	imageNameService := ksvc.Spec.ConfigurationSpec.Template.Spec.RevisionSpec.PodSpec.Containers[0].Image

	var idTaskRunInputParam gstruct.Identifier = func(element interface{}) string {
		return element.(tektonv1alpha1.Param).Name
	}
	var matchTaskRunInputParamValue = func(val string) gomegatypes.GomegaMatcher {
		return gstruct.MatchFields(gstruct.IgnoreExtras,
			gstruct.Fields{"Value": gstruct.MatchFields(gstruct.IgnoreExtras,
				gstruct.Fields{"StringVal": gomega.Equal(val)},
			)},
		)
	}
	g.Expect(tr.Spec.Inputs.Params).To(gstruct.MatchAllElements(idTaskRunInputParam, gstruct.Elements{
		"imageName":               matchTaskRunInputParamValue(imageNameService),
		"dockerfileConfigmapName": matchTaskRunInputParamValue("dockerfile-" + fnCreated.Spec.Runtime),
		"sourceConfigmapName":     matchTaskRunInputParamValue(fnCreated.Name),
	}))

	g.Expect(len(task.Spec.Steps)).To(gomega.Equal(1))

	// ensure Task build step has correct args
	g.Expect(task.Spec.Steps[0].Args).To(gomega.ConsistOf(
		"--destination=$(inputs.params.imageName)",
	))

	// ensure fetched function spec corresponds to created function spec
	fnUpdatedFetched := &serverlessv1alpha1.Function{}
	g.Expect(c.Get(context.TODO(), depKey, fnUpdatedFetched)).NotTo(gomega.HaveOccurred())
	g.Expect(fnUpdatedFetched.Spec).To(gomega.Equal(fnCreated.Spec))

	// update function code and add dependencies
	fnUpdated := fnUpdatedFetched.DeepCopy()
	fnUpdated.Spec.Function = `main() {return "bla"}`
	fnUpdated.Spec.Deps = `dependencies`
	g.Expect(c.Update(context.TODO(), fnUpdated)).NotTo(gomega.HaveOccurred())

	// get the updated function and compare spec
	fnUpdatedFetched = &serverlessv1alpha1.Function{}
	g.Expect(c.Get(context.TODO(), depKey, fnUpdatedFetched)).NotTo(gomega.HaveOccurred())

	g.Eventually(func() string {
		c.Get(context.TODO(), depKey, fnUpdatedFetched)
		return fnUpdatedFetched.Spec.Function
	}, timeout, 10*time.Second).Should(gomega.Equal(fnUpdated.Spec.Function))

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
	}, timeout, 10*time.Second).Should(gomega.Equal(fnUpdated.Spec.Function))
	g.Eventually(func() string {
		c.Get(context.TODO(), depKey, cmUpdated)
		return cmUpdated.Data["package.json"]
	}, timeout, 10*time.Second).Should(gomega.Equal(`dependencies`))

	// ensure updated knative service has updated image
	ksvcUpdated := &servingv1alpha1.Service{}
	g.Expect(c.Get(context.TODO(), depKey, ksvcUpdated)).NotTo(gomega.HaveOccurred())

	fmt.Printf("cmUpdated: %v \n", cmUpdated)
	hash = sha256.New()
	hash.Write([]byte(cmUpdated.Data["handler.js"] + cmUpdated.Data["package.json"]))
	functionSha = fmt.Sprintf("%x", hash.Sum(nil))
	fmt.Printf("functionSha: %s \n", functionSha)
	fmt.Printf("ksvcUpdated: %v \n", ksvcUpdated)
	fmt.Printf("ksvcUpdated.Spec.ConfigurationSpec.Template.Spec.RevisionSpec.PodSpec.Containers[0].Image: %s \n", ksvcUpdated.Spec.ConfigurationSpec.Template.Spec.RevisionSpec.PodSpec.Containers[0].Image)
	ksvcUpdatedImage := ksvcUpdated.Spec.ConfigurationSpec.Template.Spec.RevisionSpec.PodSpec.Containers[0].Image
	functionShaImage := fmt.Sprintf("test/%s-%s:%s", "default", "foo", functionSha)

	g.Expect(ksvcUpdatedImage).To(gomega.Equal(functionShaImage))

	g.Expect(fnUpdatedFetched.Status.Condition).To(gomega.Equal(serverlessv1alpha1.FunctionConditionDeploying))

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

	fnCreated := &serverlessv1alpha1.Function{
		ObjectMeta: metav1.ObjectMeta{
			Name:      objectName,
			Namespace: "default",
		},
		Spec: serverlessv1alpha1.FunctionSpec{
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
			"defaults":           `{"size": "S", "runtime": "nodejs8", "timeOut": 180, "funcContentType": "plaintext"}`,
			"runtimes":           `[{"ID": "nodejs6", "dockerfileName": "dockerfile-nodejs6"}]`,
			"funcSizes":          `[{"size": "L"}]`,
			"funcTypes":          `[{"type": "plaintext"}]`,
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

	function := &serverlessv1alpha1.Function{
		ObjectMeta: metav1.ObjectMeta{
			Name:      objectName,
			Namespace: "default",
		},
	}

	reconcileFunction := &ReconcileFunction{Client: c, scheme: scheme.Scheme}

	g.Expect(c.Create(context.TODO(), function)).Should(gomega.Succeed())

	reconcileFunction.setFunctionCondition(function,
		&tektonv1alpha1.TaskRun{},
		&servingv1alpha1.Service{},
	)

	// no knative objects present => no function status update
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

	function := &serverlessv1alpha1.Function{
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
	tr := &tektonv1alpha1.TaskRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:      objectName,
			Namespace: "default",
		},
	}

	// create function and build
	g.Expect(c.Create(context.TODO(), function)).Should(gomega.Succeed())
	g.Expect(c.Create(context.TODO(), tr)).Should(gomega.Succeed())
	defer func() {
		_ = c.Delete(context.TODO(), function)
		_ = c.Delete(context.TODO(), tr)
	}()

	// get build and update status
	foundTr := &tektonv1alpha1.TaskRun{}
	g.Eventually(func() error {
		return c.Get(context.TODO(), types.NamespacedName{Name: objectName, Namespace: "default"}, foundTr)
	}).Should(gomega.Succeed())
	foundTr.Status = tektonv1alpha1.TaskRunStatus{
		Status: duckv1beta1.Status{
			Conditions: duckv1beta1.Conditions{
				{
					Type:   knapis.ConditionSucceeded,
					Status: corev1.ConditionFalse,
				},
			},
		},
	}
	g.Expect(c.Status().Update(context.TODO(), foundTr)).Should(gomega.Succeed())

	g.Eventually(func() serverlessv1alpha1.FunctionCondition {
		reconcileFunction.setFunctionCondition(function,
			foundTr,
			&servingv1alpha1.Service{},
		)
		return function.Status.Condition
	}).Should(gomega.Equal(serverlessv1alpha1.FunctionConditionError))
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

	function := &serverlessv1alpha1.Function{
		ObjectMeta: metav1.ObjectMeta{
			Name:      objectName,
			Namespace: "default",
		},
	}

	service := &servingv1alpha1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      objectName,
			Namespace: "default",
		},
	}

	// create Function and Knative Service
	g.Expect(c.Create(context.TODO(), function)).Should(gomega.Succeed())
	g.Expect(c.Create(context.TODO(), service)).Should(gomega.Succeed())
	defer func() {
		_ = c.Delete(context.TODO(), function)
		_ = c.Delete(context.TODO(), service)
	}()

	// get service and update status
	foundService := &servingv1alpha1.Service{}
	g.Eventually(func() error {
		return c.Get(context.TODO(), types.NamespacedName{Name: objectName, Namespace: "default"}, foundService)
	}).Should(gomega.Succeed())
	foundService.Status = servingv1alpha1.ServiceStatus{
		ConfigurationStatusFields: servingv1alpha1.ConfigurationStatusFields{
			LatestCreatedRevisionName: "foo",
			LatestReadyRevisionName:   "foo",
		},
		Status: duckv1beta1.Status{
			Conditions: duckv1beta1.Conditions{
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
	g.Expect(c.Status().Update(context.TODO(), foundService)).Should(gomega.Succeed())

	g.Eventually(func() serverlessv1alpha1.FunctionCondition {
		reconcileFunction.setFunctionCondition(function,
			&tektonv1alpha1.TaskRun{},
			foundService,
		)
		return function.Status.Condition
	}).Should(gomega.Equal(serverlessv1alpha1.FunctionConditionRunning))
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

	function := &serverlessv1alpha1.Function{
		ObjectMeta: metav1.ObjectMeta{
			Name:      objectName,
			Namespace: "default",
		},
	}

	service := &servingv1alpha1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      objectName,
			Namespace: "default",
		},
	}

	// create function and build
	g.Expect(c.Create(context.TODO(), function)).Should(gomega.Succeed())
	g.Expect(c.Create(context.TODO(), service)).Should(gomega.Succeed())
	defer func() {
		_ = c.Delete(context.TODO(), function)
		_ = c.Delete(context.TODO(), service)
	}()

	// get service and update status
	foundService := &servingv1alpha1.Service{}
	g.Eventually(func() error {
		return c.Get(context.TODO(), types.NamespacedName{Name: objectName, Namespace: "default"}, foundService)
	}).Should(gomega.Succeed())
	foundService.Status = servingv1alpha1.ServiceStatus{
		ConfigurationStatusFields: servingv1alpha1.ConfigurationStatusFields{
			LatestCreatedRevisionName: "foo",
			LatestReadyRevisionName:   "foo",
		},
		Status: duckv1beta1.Status{
			Conditions: duckv1beta1.Conditions{
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
	g.Expect(c.Status().Update(context.TODO(), foundService)).Should(gomega.Succeed())

	g.Eventually(func() serverlessv1alpha1.FunctionCondition {
		reconcileFunction.setFunctionCondition(function,
			&tektonv1alpha1.TaskRun{},
			foundService,
		)
		return function.Status.Condition
	}).Should(gomega.Equal(serverlessv1alpha1.FunctionConditionDeploying))
}

func TestCreateFunctionHandlerMap(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	functionCode := "some function code"
	functionDependecies := "some function dependencies"
	function := serverlessv1alpha1.Function{Spec: serverlessv1alpha1.FunctionSpec{
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
	function := serverlessv1alpha1.Function{Spec: serverlessv1alpha1.FunctionSpec{
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
