package controllers_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	serverless "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"github.com/onsi/gomega"
	tektonv1alpha1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestReconcileBuilding(t *testing.T) {
	emptyReason := serverless.ConditionReason("")
	testCases := []testdata{
		{
			desc: "ConditionReasonNoImageTag",
			fn: serverless.Function{
				ObjectMeta: *objmeta,
				Status: serverless.FunctionStatus{
					Phase: serverless.FunctionPhaseBuilding,
				},
			},
			expectedReason: serverless.ConditionReasonBuildFailed,
			expectedStatus: serverless.FunctionPhaseFailed,
		},
		{
			desc: "ConditionReasonTasRunNotFound",
			fn: serverless.Function{
				ObjectMeta: *objmeta,
				Status: serverless.FunctionStatus{
					Phase:    serverless.FunctionPhaseBuilding,
					ImageTag: "1",
				},
			},
			expectedReason: serverless.ConditionReasonBuildFailed,
			expectedStatus: serverless.FunctionPhaseFailed,
		},
		{
			desc: "ConditionReasonUnknown",
			fn: serverless.Function{
				ObjectMeta: *objmeta,
				Status: serverless.FunctionStatus{
					Phase:    serverless.FunctionPhaseBuilding,
					ImageTag: "1",
				},
			},
			mocks: []runtime.Object{
				testTaskRun(functionUID, "1"),
				testTaskRun(functionUID, "1"),
			},
			expectedReason: serverless.ConditionReasonBuildFailed,
			expectedStatus: serverless.FunctionPhaseFailed,
		},
		{
			desc: "no condition reason - task run running",
			fn: serverless.Function{
				ObjectMeta: *objmeta,
				Status: serverless.FunctionStatus{
					Phase:    serverless.FunctionPhaseBuilding,
					ImageTag: "1",
				},
			},
			mocks: []runtime.Object{
				testTaskRun(functionUID, "1"),
			},
			expectedStatus: serverless.FunctionPhaseBuilding,
			expectedReason: emptyReason,
		},
		{
			desc: "TaskRunCancelled",
			fn: serverless.Function{
				ObjectMeta: *objmeta,
				Status: serverless.FunctionStatus{
					Phase:    serverless.FunctionPhaseBuilding,
					ImageTag: "1",
				},
			},
			mocks: []runtime.Object{
				&tektonv1alpha1.TaskRun{
					ObjectMeta: v1.ObjectMeta{
						Name:      uuid.New().String(),
						Namespace: testnamespace,
						Labels: map[string]string{
							"fnUUID":   functionUID,
							"imageTag": "1",
						},
					},
					Spec: tektonv1alpha1.TaskRunSpec{
						Status: "TaskRunCancelled",
					},
				},
			},
			expectedStatus: serverless.FunctionPhaseFailed,
			expectedReason: serverless.ConditionReasonBuildFailed,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			g := gomega.NewWithT(t)
			allmocks := append(tC.mocks, &tC.fn)
			ct, rf, req := prepareData(objmeta, allmocks...)
			_, err := rf.Reconcile(req)
			g.Expect(err).ShouldNot(gomega.HaveOccurred())
			g.Expect(ct.Get(context.TODO(), namespacedname, &tC.fn)).ShouldNot(gomega.HaveOccurred())
			g.Expect(tC.fn.Status.Phase).To(gomega.Equal(tC.expectedStatus))

			if tC.expectedReason == emptyReason {
				g.Expect(tC.fn.Status.Conditions).Should(gomega.HaveLen(0))
				return
			}
			g.Expect(tC.fn.Status.Conditions).Should(gomega.HaveLen(1))
			g.Expect(tC.fn.Status.Conditions[0].Reason).To(gomega.Equal(tC.expectedReason))
		})
	}
}
