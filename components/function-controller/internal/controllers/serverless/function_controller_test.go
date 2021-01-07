package serverless

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
)

func newFixFunction(namespace, name string, minReplicas, maxReplicas int) *serverlessv1alpha1.Function {
	one := int32(minReplicas)
	two := int32(maxReplicas)
	suffix := rand.Int()

	return &serverlessv1alpha1.Function{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%d", name, suffix),
			Namespace: namespace,
		},
		Spec: serverlessv1alpha1.FunctionSpec{
			Source:  "module.exports = {main: function(event, context) {return 'Hello World.'}}",
			Deps:    "   ",
			Runtime: serverlessv1alpha1.Nodejs12,
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
					{Type: serverlessv1alpha1.ConditionRunning, Status: corev1.ConditionTrue, Reason: serverlessv1alpha1.ConditionReasonServiceCreated, Message: "some message"},
					{Type: serverlessv1alpha1.ConditionBuildReady, Status: corev1.ConditionFalse, Reason: serverlessv1alpha1.ConditionReasonJobFinished, Message: "blabla"}},
				expected: []serverlessv1alpha1.Condition{
					{Type: serverlessv1alpha1.ConditionConfigurationReady, Status: corev1.ConditionFalse, Reason: serverlessv1alpha1.ConditionReasonConfigMapUpdated, Message: "msg"},
					{Type: serverlessv1alpha1.ConditionRunning, Status: corev1.ConditionTrue, Reason: serverlessv1alpha1.ConditionReasonServiceCreated, Message: "some message"},
					{Type: serverlessv1alpha1.ConditionBuildReady, Status: corev1.ConditionFalse, Reason: serverlessv1alpha1.ConditionReasonJobFinished, Message: "blabla"}},
			},
			want: true,
		},
		{
			name: "should return false on slices with different lengths",
			args: args{
				existing: []serverlessv1alpha1.Condition{
					{Type: serverlessv1alpha1.ConditionConfigurationReady, Status: corev1.ConditionFalse, Reason: serverlessv1alpha1.ConditionReasonConfigMapUpdated, Message: "msg"},
					{Type: serverlessv1alpha1.ConditionRunning, Status: corev1.ConditionTrue, Reason: serverlessv1alpha1.ConditionReasonServiceCreated, Message: "some message"},
					{Type: serverlessv1alpha1.ConditionBuildReady, Status: corev1.ConditionFalse, Reason: serverlessv1alpha1.ConditionReasonJobFinished, Message: "blabla"}},
				expected: []serverlessv1alpha1.Condition{
					{Type: serverlessv1alpha1.ConditionConfigurationReady, Status: corev1.ConditionFalse, Reason: serverlessv1alpha1.ConditionReasonConfigMapUpdated, Message: "msg"},
					{Type: serverlessv1alpha1.ConditionRunning, Status: corev1.ConditionTrue, Reason: serverlessv1alpha1.ConditionReasonServiceCreated, Message: "some message"},
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

func TestFunctionReconciler_envsEqual(t *testing.T) {
	envVarSrc := &corev1.EnvVarSource{
		ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{
				Name: "some-name",
			},
			Key:      "some-key",
			Optional: nil,
		},
	}

	envVarSrc2 := &corev1.EnvVarSource{
		ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{
				Name: "some-name",
			},
			Key:      "some-key",
			Optional: nil,
		},
	}

	differentEnvVarSrc := &corev1.EnvVarSource{
		ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{
				Name: "some-name",
			},
			Key:      "some-key-that-is-different",
			Optional: nil,
		},
	}

	type args struct {
		existing []corev1.EnvVar
		expected []corev1.EnvVar
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "simple case",
			args: args{
				existing: []corev1.EnvVar{{Name: "env1", Value: "val1"}, {Name: "env2", Value: "val2"}},
				expected: []corev1.EnvVar{{Name: "env1", Value: "val1"}, {Name: "env2", Value: "val2"}},
			},
			want: true,
		},
		{
			name: "different length case",
			args: args{
				existing: []corev1.EnvVar{{Name: "env1", Value: "val1"}},
				expected: []corev1.EnvVar{{Name: "env1", Value: "val1"}, {Name: "env2", Value: "val2"}},
			},
			want: false,
		},
		{
			name: "different length case",
			args: args{
				existing: []corev1.EnvVar{{Name: "env1", Value: "val1"}},
				expected: []corev1.EnvVar{{Name: "env1", Value: "val1"}, {Name: "env2", Value: "val2"}},
			},
			want: false,
		},
		{
			name: "different value in one env",
			args: args{
				existing: []corev1.EnvVar{{Name: "env1", Value: "val1"}, {Name: "env2", Value: "different-value"}},
				expected: []corev1.EnvVar{{Name: "env1", Value: "val1"}, {Name: "env2", Value: "val2"}},
			},
			want: false,
		},
		{
			name: "same valueFrom in one env - same reference",
			args: args{
				existing: []corev1.EnvVar{{Name: "env1", Value: "val1"}, {Name: "env2", ValueFrom: envVarSrc}},
				expected: []corev1.EnvVar{{Name: "env1", Value: "val1"}, {Name: "env2", ValueFrom: envVarSrc}},
			},
			want: true,
		},
		{
			name: "same valueFrom in one env - same object, different reference",
			args: args{
				existing: []corev1.EnvVar{{Name: "env1", Value: "val1"}, {Name: "env2", ValueFrom: envVarSrc}},
				expected: []corev1.EnvVar{{Name: "env1", Value: "val1"}, {Name: "env2", ValueFrom: envVarSrc2}},
			},
			want: true,
		},
		{
			name: "different valueFrom in one env",
			args: args{
				existing: []corev1.EnvVar{{Name: "env1", Value: "val1"}, {Name: "env2", ValueFrom: envVarSrc}},
				expected: []corev1.EnvVar{{Name: "env1", Value: "val1"}, {Name: "env2", ValueFrom: differentEnvVarSrc}},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewGomegaWithT(t)
			r := &FunctionReconciler{}
			got := r.envsEqual(tt.args.existing, tt.args.expected)
			g.Expect(got).To(gomega.Equal(tt.want))
		})
	}
}

func TestFunctionReconciler_getConditionReason(t *testing.T) {
	type args struct {
		conditions    []serverlessv1alpha1.Condition
		conditionType serverlessv1alpha1.ConditionType
	}
	tests := []struct {
		name string
		args args
		want serverlessv1alpha1.ConditionReason
	}{
		{
			name: "Should correctly return proper status given slice of conditions",
			args: args{
				conditions: []serverlessv1alpha1.Condition{
					{Type: serverlessv1alpha1.ConditionConfigurationReady, Status: corev1.ConditionFalse, Reason: serverlessv1alpha1.ConditionReasonConfigMapCreated},
					{Type: serverlessv1alpha1.ConditionRunning, Status: corev1.ConditionTrue, Reason: serverlessv1alpha1.ConditionReasonServiceCreated},
					{Type: serverlessv1alpha1.ConditionConfigurationReady, Status: corev1.ConditionFalse, Reason: serverlessv1alpha1.ConditionReasonDeploymentWaiting},
				},
				conditionType: serverlessv1alpha1.ConditionRunning,
			},
			want: serverlessv1alpha1.ConditionReasonServiceCreated,
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
			want: "",
		},
		{
			name: "Should correctly return status unknown if slice is empty",
			args: args{
				conditions:    []serverlessv1alpha1.Condition{},
				conditionType: serverlessv1alpha1.ConditionRunning,
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewGomegaWithT(t)
			r := &FunctionReconciler{}
			got := r.getConditionReason(tt.args.conditions, tt.args.conditionType)
			g.Expect(got).To(gomega.Equal(tt.want))
		})
	}
}
