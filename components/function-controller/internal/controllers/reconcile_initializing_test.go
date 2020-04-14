package controllers_test

import (
	"context"
	"testing"

	serverless "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"github.com/onsi/gomega"
	tektonv1alpha1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func TestReconcileInitializingNewFunction(t *testing.T) {
	helloWorldFn := serverless.Function{
		Status: serverless.FunctionStatus{
			Phase: serverless.FunctionPhaseInitializing,
		},
		ObjectMeta: *objmeta,
		Spec: serverless.FunctionSpec{
			Function: `module.exports = {
				main: function(event, context) {
				  return 'Hello World'
				}
			  }`,
			Deps: `{
				"name": "hellowithdeps",
				"version": "0.0.1",
				"dependencies": {
				  "end-of-stream": "^1.4.1",
				  "from2": "^2.3.0",
				  "lodash": "^4.17.5"
				}
			  }`,
		},
	}

	testCases := []testdata{
		{
			desc:           "new function",
			fn:             helloWorldFn,
			expectedStatus: serverless.FunctionPhaseBuilding,
			expectedReason: serverless.ConditionReasonCreateConfigSucceeded,
		},
		{
			desc: "updated function",
			fn:   helloWorldFn,
			mocks: []runtime.Object{
				mustCfgMapWithOwner(&helloWorldFn),
			},
			expectedStatus: serverless.FunctionPhaseBuilding,
			expectedReason: serverless.ConditionReasonUpdateConfigSucceeded,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			g := gomega.NewWithT(t)

			// workaround
			// tektonv1alpha1.AddToScheme(scheme.Scheme)

			allmocks := append(tC.mocks, &tC.fn)
			ct, rf, req := prepareData(objmeta, allmocks...)

			_, err := rf.Reconcile(req)
			g.Expect(err).ShouldNot(gomega.HaveOccurred())

			g.Expect(ct.Get(context.TODO(), namespacedname, &tC.fn)).ShouldNot(gomega.HaveOccurred())
			g.Expect(tC.fn.Status.ObservedGeneration).Should(gomega.Equal(tC.fn.ObjectMeta.Generation))
			g.Expect(tC.fn.Status.Phase).To(gomega.Equal(tC.expectedStatus))
			g.Expect(tC.fn.Status.Conditions).ShouldNot(gomega.HaveLen(0))

			length := len(tC.fn.Status.Conditions)
			g.Expect(tC.fn.Status.Conditions[length-1].Reason).To(gomega.Equal(tC.expectedReason))

			var cms corev1.ConfigMapList
			err = ct.List(context.TODO(), &cms, &client.ListOptions{
				LabelSelector: tC.fn.LabelSelector(),
			})
			g.Expect(err).ShouldNot(gomega.HaveOccurred())
			g.Expect(cms.Items).Should(gomega.HaveLen(1))

			cm := &cms.Items[0]

			fnRef := fnOwnerReference(&tC.fn)

			g.Expect(cm.ObjectMeta.OwnerReferences).Should(gomega.ContainElement(*fnRef))

			var trs tektonv1alpha1.TaskRunList
			g.Expect(ct.List(context.TODO(), &trs, &client.ListOptions{
				LabelSelector: labels.SelectorFromSet(map[string]string{
					"fnUUID": string(tC.fn.UID),
				}),
			})).ShouldNot(gomega.HaveOccurred())
			g.Expect(len(trs.Items)).To(gomega.Equal(1))
		})
	}
}

func mustCfgMapWithOwner(owner metav1.Object) *corev1.ConfigMap {
	cm := corev1.ConfigMap{
		ObjectMeta: *objmeta,
		Data: map[string]string{
			"handler":      "handler.main",
			"handler.js":   "",
			"package.json": "{}",
		},
	}
	err := controllerutil.SetControllerReference(owner, &cm, scheme.Scheme)
	if err != nil {
		panic("unable to set controller reference to config map")
	}
	return &cm
}
