package controllers_test

import (
	"time"

	"github.com/google/uuid"
	function "github.com/kyma-project/kyma/components/function-controller/internal/controllers"
	serverless "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	tektonv1alpha1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

type fn = serverless.Function

type ct = client.Client

type rf = function.FunctionReconciler

type rq = ctrl.Request

const (
	testnamespace = "test-namespace"
	functionUID   = "1"
)

var (
	objmeta = &metav1.ObjectMeta{
		Name:      "function1",
		Namespace: testnamespace,
		UID:       functionUID,
		Labels: map[string]string{
			serverless.FnUUID: functionUID,
		},
	}
	namespacedname = namespacedName(objmeta)
)

func testWaitForCacheSync(stop <-chan struct{}) bool {
	return true
}

type testdata struct {
	desc           string
	fn             serverless.Function
	mocks          []runtime.Object
	expectedReason serverless.ConditionReason
	expectedStatus serverless.StatusPhase
}

func testTaskRun(fnUUID, imageTag string) *tektonv1alpha1.TaskRun {
	return &tektonv1alpha1.TaskRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:      uuid.New().String(),
			Namespace: testnamespace,
			Labels: map[string]string{
				"fnUUID":   fnUUID,
				"imageTag": imageTag,
			},
		},
	}
}

func namespacedName(obj *metav1.ObjectMeta) types.NamespacedName {
	return types.NamespacedName{
		Name:      obj.Name,
		Namespace: obj.Namespace,
	}
}

func fnOwnerReference(fn *fn) *metav1.OwnerReference {
	vtrue := true
	return &metav1.OwnerReference{
		Name:               fn.Name,
		APIVersion:         fn.APIVersion,
		Kind:               fn.Kind,
		UID:                fn.UID,
		Controller:         &vtrue,
		BlockOwnerDeletion: &vtrue,
	}
}

func prepareData(reqobj *metav1.ObjectMeta, obj ...runtime.Object) (ct, rf, rq) {
	ct := fake.NewFakeClientWithScheme(scheme.Scheme, obj...)
	cfg := function.Cfg{
		Client:            ct,
		Scheme:            scheme.Scheme,
		CacheSynchronizer: testWaitForCacheSync,
		EventRecorder:     record.NewFakeRecorder(1),
		Log:               zap.Logger(true),
	}
	rf := function.NewFunctionReconciler(
		&cfg,
		&function.FnReconcilerCfg{
			MaxConcurrentReconciles: 1,
			Limits:                  &corev1.ResourceList{},
			Requests:                &corev1.ResourceList{},
			RequeueDuration:         time.Hour,
		})
	req := req(reqobj)
	return ct, *rf, req
}

func req(obj *metav1.ObjectMeta) ctrl.Request {
	return ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      obj.Name,
			Namespace: obj.Namespace,
		},
	}
}

func init() {
	servingv1.AddToScheme(scheme.Scheme)
	serverless.AddToScheme(scheme.Scheme)
	corev1.AddToScheme(scheme.Scheme)
	tektonv1alpha1.AddToScheme(scheme.Scheme)
}
