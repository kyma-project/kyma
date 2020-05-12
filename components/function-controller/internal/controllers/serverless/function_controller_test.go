package serverless

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"

	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
)

var _ = ginkgo.Describe("updateConfigMap", func() {
	var (
		reconciler *FunctionReconciler
		request    ctrl.Request
	)

	ginkgo.BeforeEach(func() {
		function := newFixFunction("tutaj", "ah-tak-przeciez")
		request = ctrl.Request{NamespacedName: types.NamespacedName{Namespace: function.GetNamespace(), Name: function.GetName()}}
		gomega.Expect(resourceClient.Create(context.TODO(), function)).To(gomega.Succeed())

		reconciler = NewFunction(resourceClient, log.Log, config, record.NewFakeRecorder(100))
	})

	ginkgo.It("should update configmap", func() {
		result, err := reconciler.Reconcile(request)
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(result.Requeue).To(gomega.BeFalse())
		gomega.Expect(result.RequeueAfter).To(gomega.Equal(time.Second * 0))
	})
})

func newFixFunction(namespace, name string) *serverlessv1alpha1.Function {
	one := int32(1)
	two := int32(2)
	suffix := rand.Int()

	return &serverlessv1alpha1.Function{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%d", name, suffix),
			Namespace: namespace,
		},
		Spec: serverlessv1alpha1.FunctionSpec{
			Source: "module.exports = {main: function(event, context) {return 'Hello World. Epstein didnt kill himself.'}}",
			Deps:   "   ",
			Env: []corev1.EnvVar{
				{
					Name:  "TEST_1",
					Value: "VAL_1",
				},
				{
					Name:  "TEST_2",
					Value: "VAL_2",
				},
			},
			Resources:   corev1.ResourceRequirements{},
			MinReplicas: &one,
			MaxReplicas: &two,
			Labels: map[string]string{
				testBindingLabel1: "foobar",
				testBindingLabel2: testBindingLabelValue,
				"foo":             "bar",
			},
		},
	}
}

func TestFunctionReconciler_getConditionStatus(t *testing.T) {
	type args struct {
		conditions    []serverlessv1alpha1.Condition
		conditionType serverlessv1alpha1.ConditionType
	}
	tests := []struct {
		name string
		args args
		want corev1.ConditionStatus
	}{
		{
			name: "Should correctly return proper status given slice of conditions",
			args: args{
				conditions: []serverlessv1alpha1.Condition{
					{Type: serverlessv1alpha1.ConditionConfigurationReady, Status: corev1.ConditionFalse},
					{Type: serverlessv1alpha1.ConditionRunning, Status: corev1.ConditionTrue},
					{Type: serverlessv1alpha1.ConditionConfigurationReady, Status: corev1.ConditionFalse},
				},
				conditionType: serverlessv1alpha1.ConditionRunning,
			},
			want: corev1.ConditionTrue,
		},
		{
			name: "Should correctly return status unknown if there's no needed conditionType",
			args: args{
				conditions: []serverlessv1alpha1.Condition{
					{Type: serverlessv1alpha1.ConditionConfigurationReady, Status: corev1.ConditionFalse},
					{Type: serverlessv1alpha1.ConditionConfigurationReady, Status: corev1.ConditionFalse},
				},
				conditionType: serverlessv1alpha1.ConditionRunning,
			},
			want: corev1.ConditionUnknown,
		},
		{
			name: "Should correctly return status unknown if slice is empty",
			args: args{
				conditions:    []serverlessv1alpha1.Condition{},
				conditionType: serverlessv1alpha1.ConditionRunning,
			},
			want: corev1.ConditionUnknown,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewGomegaWithT(t)
			r := &FunctionReconciler{}
			got := r.getConditionStatus(tt.args.conditions, tt.args.conditionType)
			g.Expect(got).To(gomega.Equal(tt.want))
		})
	}
}

func TestFunctionReconciler_equalConditions(t *testing.T) {
	type args struct {
		existing []serverlessv1alpha1.Condition
		expected []serverlessv1alpha1.Condition
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "should work on the same slices",
			args: args{
				existing: []serverlessv1alpha1.Condition{
					{Type: serverlessv1alpha1.ConditionConfigurationReady, Status: corev1.ConditionFalse, Reason: serverlessv1alpha1.ConditionReasonConfigMapUpdated, Message: "msg"},
					{Type: serverlessv1alpha1.ConditionRunning, Status: corev1.ConditionTrue, Reason: serverlessv1alpha1.ConditionReasonServiceReady, Message: "some message"},
					{Type: serverlessv1alpha1.ConditionBuildReady, Status: corev1.ConditionFalse, Reason: serverlessv1alpha1.ConditionReasonJobFinished, Message: "blabla"}},
				expected: []serverlessv1alpha1.Condition{
					{Type: serverlessv1alpha1.ConditionConfigurationReady, Status: corev1.ConditionFalse, Reason: serverlessv1alpha1.ConditionReasonConfigMapUpdated, Message: "msg"},
					{Type: serverlessv1alpha1.ConditionRunning, Status: corev1.ConditionTrue, Reason: serverlessv1alpha1.ConditionReasonServiceReady, Message: "some message"},
					{Type: serverlessv1alpha1.ConditionBuildReady, Status: corev1.ConditionFalse, Reason: serverlessv1alpha1.ConditionReasonJobFinished, Message: "blabla"}},
			},
			want: true,
		},
		{
			name: "should return false on slices with different lengths",
			args: args{
				existing: []serverlessv1alpha1.Condition{
					{Type: serverlessv1alpha1.ConditionConfigurationReady, Status: corev1.ConditionFalse, Reason: serverlessv1alpha1.ConditionReasonConfigMapUpdated, Message: "msg"},
					{Type: serverlessv1alpha1.ConditionRunning, Status: corev1.ConditionTrue, Reason: serverlessv1alpha1.ConditionReasonServiceReady, Message: "some message"},
					{Type: serverlessv1alpha1.ConditionBuildReady, Status: corev1.ConditionFalse, Reason: serverlessv1alpha1.ConditionReasonJobFinished, Message: "blabla"}},
				expected: []serverlessv1alpha1.Condition{
					{Type: serverlessv1alpha1.ConditionConfigurationReady, Status: corev1.ConditionFalse, Reason: serverlessv1alpha1.ConditionReasonConfigMapUpdated, Message: "msg"},
					{Type: serverlessv1alpha1.ConditionRunning, Status: corev1.ConditionTrue, Reason: serverlessv1alpha1.ConditionReasonServiceReady, Message: "some message"},
				},
			},
			want: false,
		},
		{
			name: "should return false on different conditions",
			args: args{
				existing: []serverlessv1alpha1.Condition{
					{Type: serverlessv1alpha1.ConditionConfigurationReady, Status: corev1.ConditionFalse, Reason: serverlessv1alpha1.ConditionReasonConfigMapUpdated, Message: "msg"}},
				expected: []serverlessv1alpha1.Condition{
					{Type: serverlessv1alpha1.ConditionBuildReady, Status: corev1.ConditionFalse, Reason: serverlessv1alpha1.ConditionReasonConfigMapUpdated, Message: "msg"}},
			},
			want: false,
		},
		{
			name: "should return false on different Statuses",
			args: args{
				existing: []serverlessv1alpha1.Condition{
					{Type: serverlessv1alpha1.ConditionConfigurationReady, Status: corev1.ConditionFalse, Reason: serverlessv1alpha1.ConditionReasonConfigMapUpdated, Message: "msg"}},
				expected: []serverlessv1alpha1.Condition{
					{Type: serverlessv1alpha1.ConditionConfigurationReady, Status: corev1.ConditionUnknown, Reason: serverlessv1alpha1.ConditionReasonConfigMapUpdated, Message: "msg"}},
			},
			want: false,
		},
		{
			name: "should return false on different Reasons",
			args: args{
				existing: []serverlessv1alpha1.Condition{
					{Type: serverlessv1alpha1.ConditionConfigurationReady, Status: corev1.ConditionFalse, Reason: serverlessv1alpha1.ConditionReasonConfigMapUpdated, Message: "msg"}},
				expected: []serverlessv1alpha1.Condition{
					{Type: serverlessv1alpha1.ConditionConfigurationReady, Status: corev1.ConditionFalse, Reason: serverlessv1alpha1.ConditionReasonConfigMapCreated, Message: "msg"}},
			},
			want: false,
		},
		{
			name: "should return false on different messages",
			args: args{
				existing: []serverlessv1alpha1.Condition{
					{Type: serverlessv1alpha1.ConditionConfigurationReady, Status: corev1.ConditionFalse, Reason: serverlessv1alpha1.ConditionReasonConfigMapUpdated, Message: "msg"}},
				expected: []serverlessv1alpha1.Condition{
					{Type: serverlessv1alpha1.ConditionConfigurationReady, Status: corev1.ConditionFalse, Reason: serverlessv1alpha1.ConditionReasonConfigMapUpdated, Message: "msg-different"}},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewGomegaWithT(t)
			r := &FunctionReconciler{}
			got := r.equalConditions(tt.args.existing, tt.args.expected)
			g.Expect(got).To(gomega.Equal(tt.want))
		})
	}
}

func TestFunctionReconciler_mapsEqual(t *testing.T) {
	type args struct {
		existing map[string]string
		expected map[string]string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "two empty maps are the same",
			args: args{
				expected: map[string]string{},
				existing: map[string]string{},
			},
			want: true,
		},
		{
			name: "two maps with different len are different",
			args: args{
				expected: map[string]string{"some": "things"},
				existing: map[string]string{},
			},
			want: false,
		},
		{
			name: "two maps with same content are same",
			args: args{
				expected: map[string]string{"some": "things"},
				existing: map[string]string{"some": "things"},
			},
			want: true,
		},
		{
			name: "two maps with same content, but in different order, are same",
			args: args{
				expected: map[string]string{
					"some":       "things",
					"should not": "be seen",
				},
				existing: map[string]string{
					"should not": "be seen",
					"some":       "things",
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewGomegaWithT(t)
			r := &FunctionReconciler{}
			got := r.mapsEqual(tt.args.existing, tt.args.expected)
			g.Expect(got).To(gomega.Equal(tt.want))
		})
	}
}

func TestFunctionReconciler_equalJobs(t *testing.T) {
	type args struct {
		existing batchv1.Job
		expected batchv1.Job
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "two jobs with same container args are same",
			args: args{
				existing: batchv1.Job{Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{Args: []string{"arg1"}}},
						},
					},
				}},
				expected: batchv1.Job{Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{Args: []string{"arg1"}}},
						},
					},
				}},
			},
			want: true,
		},
		{
			name: "two jobs with same, multiple container args are same",
			args: args{
				existing: batchv1.Job{Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{Args: []string{"arg1", "arg2"}}},
						},
					},
				}},
				expected: batchv1.Job{Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{Args: []string{"arg1", "arg2"}}},
						},
					},
				}},
			},
			want: true,
		},
		{
			name: "two jobs with different length of args are not same",
			args: args{
				existing: batchv1.Job{Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{Args: []string{"arg1"}}},
						},
					},
				}},
				expected: batchv1.Job{Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{Args: []string{"arg1", "args2"}}},
						},
					},
				}},
			},
			want: false,
		},
		{
			name: "two jobs with different arg are not the same",
			args: args{
				existing: batchv1.Job{Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{Args: []string{"arg1"}}},
						},
					},
				}},
				expected: batchv1.Job{Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{Args: []string{"argument-number-1"}}},
						},
					},
				}},
			},
			want: false,
		},
		{
			name: "two jobs with second different argument are not the same",
			args: args{
				existing: batchv1.Job{Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{Args: []string{"arg1", "test-value-1"}}},
						},
					},
				}},
				expected: batchv1.Job{Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{Args: []string{"arg1", "test-value-30"}}},
						},
					},
				}},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewGomegaWithT(t)
			r := &FunctionReconciler{}
			got := r.equalJobs(tt.args.existing, tt.args.expected)
			g.Expect(got).To(gomega.Equal(tt.want))
		})
	}
}
