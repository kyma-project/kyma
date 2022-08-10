package serverless

import (
	"testing"

	serverlessv1alpha2 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha2"
	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

func TestFunctionReconciler_getConditionStatus(t *testing.T) {
	type args struct {
		conditions    []serverlessv1alpha2.Condition
		conditionType serverlessv1alpha2.ConditionType
	}
	tests := []struct {
		name string
		args args
		want corev1.ConditionStatus
	}{
		{
			name: "Should correctly return proper status given slice of conditions",
			args: args{
				conditions: []serverlessv1alpha2.Condition{
					{Type: serverlessv1alpha2.ConditionConfigurationReady, Status: corev1.ConditionFalse},
					{Type: serverlessv1alpha2.ConditionRunning, Status: corev1.ConditionTrue},
					{Type: serverlessv1alpha2.ConditionConfigurationReady, Status: corev1.ConditionFalse},
				},
				conditionType: serverlessv1alpha2.ConditionRunning,
			},
			want: corev1.ConditionTrue,
		},
		{
			name: "Should correctly return status unknown if there's no needed conditionType",
			args: args{
				conditions: []serverlessv1alpha2.Condition{
					{Type: serverlessv1alpha2.ConditionConfigurationReady, Status: corev1.ConditionFalse},
					{Type: serverlessv1alpha2.ConditionConfigurationReady, Status: corev1.ConditionFalse},
				},
				conditionType: serverlessv1alpha2.ConditionRunning,
			},
			want: corev1.ConditionUnknown,
		},
		{
			name: "Should correctly return status unknown if slice is empty",
			args: args{
				conditions:    []serverlessv1alpha2.Condition{},
				conditionType: serverlessv1alpha2.ConditionRunning,
			},
			want: corev1.ConditionUnknown,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewGomegaWithT(t)
			got := getConditionStatus(tt.args.conditions, tt.args.conditionType)
			g.Expect(got).To(gomega.Equal(tt.want))
		})
	}
}

func TestFunctionReconciler_equalConditions(t *testing.T) {
	type args struct {
		existing []serverlessv1alpha2.Condition
		expected []serverlessv1alpha2.Condition
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "should work on the same slices",
			args: args{
				existing: []serverlessv1alpha2.Condition{
					{Type: serverlessv1alpha2.ConditionConfigurationReady, Status: corev1.ConditionFalse, Reason: serverlessv1alpha2.ConditionReasonConfigMapUpdated, Message: "msg"},
					{Type: serverlessv1alpha2.ConditionRunning, Status: corev1.ConditionTrue, Reason: serverlessv1alpha2.ConditionReasonServiceCreated, Message: "some message"},
					{Type: serverlessv1alpha2.ConditionBuildReady, Status: corev1.ConditionFalse, Reason: serverlessv1alpha2.ConditionReasonJobFinished, Message: "blabla"}},
				expected: []serverlessv1alpha2.Condition{
					{Type: serverlessv1alpha2.ConditionConfigurationReady, Status: corev1.ConditionFalse, Reason: serverlessv1alpha2.ConditionReasonConfigMapUpdated, Message: "msg"},
					{Type: serverlessv1alpha2.ConditionRunning, Status: corev1.ConditionTrue, Reason: serverlessv1alpha2.ConditionReasonServiceCreated, Message: "some message"},
					{Type: serverlessv1alpha2.ConditionBuildReady, Status: corev1.ConditionFalse, Reason: serverlessv1alpha2.ConditionReasonJobFinished, Message: "blabla"}},
			},
			want: true,
		},
		{
			name: "should return false on slices with different lengths",
			args: args{
				existing: []serverlessv1alpha2.Condition{
					{Type: serverlessv1alpha2.ConditionConfigurationReady, Status: corev1.ConditionFalse, Reason: serverlessv1alpha2.ConditionReasonConfigMapUpdated, Message: "msg"},
					{Type: serverlessv1alpha2.ConditionRunning, Status: corev1.ConditionTrue, Reason: serverlessv1alpha2.ConditionReasonServiceCreated, Message: "some message"},
					{Type: serverlessv1alpha2.ConditionBuildReady, Status: corev1.ConditionFalse, Reason: serverlessv1alpha2.ConditionReasonJobFinished, Message: "blabla"}},
				expected: []serverlessv1alpha2.Condition{
					{Type: serverlessv1alpha2.ConditionConfigurationReady, Status: corev1.ConditionFalse, Reason: serverlessv1alpha2.ConditionReasonConfigMapUpdated, Message: "msg"},
					{Type: serverlessv1alpha2.ConditionRunning, Status: corev1.ConditionTrue, Reason: serverlessv1alpha2.ConditionReasonServiceCreated, Message: "some message"},
				},
			},
			want: false,
		},
		{
			name: "should return false on different conditions",
			args: args{
				existing: []serverlessv1alpha2.Condition{
					{Type: serverlessv1alpha2.ConditionConfigurationReady, Status: corev1.ConditionFalse, Reason: serverlessv1alpha2.ConditionReasonConfigMapUpdated, Message: "msg"}},
				expected: []serverlessv1alpha2.Condition{
					{Type: serverlessv1alpha2.ConditionBuildReady, Status: corev1.ConditionFalse, Reason: serverlessv1alpha2.ConditionReasonConfigMapUpdated, Message: "msg"}},
			},
			want: false,
		},
		{
			name: "should return false on different Statuses",
			args: args{
				existing: []serverlessv1alpha2.Condition{
					{Type: serverlessv1alpha2.ConditionConfigurationReady, Status: corev1.ConditionFalse, Reason: serverlessv1alpha2.ConditionReasonConfigMapUpdated, Message: "msg"}},
				expected: []serverlessv1alpha2.Condition{
					{Type: serverlessv1alpha2.ConditionConfigurationReady, Status: corev1.ConditionUnknown, Reason: serverlessv1alpha2.ConditionReasonConfigMapUpdated, Message: "msg"}},
			},
			want: false,
		},
		{
			name: "should return false on different Reasons",
			args: args{
				existing: []serverlessv1alpha2.Condition{
					{Type: serverlessv1alpha2.ConditionConfigurationReady, Status: corev1.ConditionFalse, Reason: serverlessv1alpha2.ConditionReasonConfigMapUpdated, Message: "msg"}},
				expected: []serverlessv1alpha2.Condition{
					{Type: serverlessv1alpha2.ConditionConfigurationReady, Status: corev1.ConditionFalse, Reason: serverlessv1alpha2.ConditionReasonConfigMapCreated, Message: "msg"}},
			},
			want: false,
		},
		{
			name: "should return false on different messages",
			args: args{
				existing: []serverlessv1alpha2.Condition{
					{Type: serverlessv1alpha2.ConditionConfigurationReady, Status: corev1.ConditionFalse, Reason: serverlessv1alpha2.ConditionReasonConfigMapUpdated, Message: "msg"}},
				expected: []serverlessv1alpha2.Condition{
					{Type: serverlessv1alpha2.ConditionConfigurationReady, Status: corev1.ConditionFalse, Reason: serverlessv1alpha2.ConditionReasonConfigMapUpdated, Message: "msg-different"}},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewGomegaWithT(t)
			got := equalConditions(tt.args.existing, tt.args.expected)
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
			got := mapsEqual(tt.args.existing, tt.args.expected)
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
			got := envsEqual(tt.args.existing, tt.args.expected)
			g.Expect(got).To(gomega.Equal(tt.want))
		})
	}
}

func TestFunctionReconciler_getConditionReason(t *testing.T) {
	type args struct {
		conditions    []serverlessv1alpha2.Condition
		conditionType serverlessv1alpha2.ConditionType
	}
	tests := []struct {
		name string
		args args
		want serverlessv1alpha2.ConditionReason
	}{
		{
			name: "Should correctly return proper status given slice of conditions",
			args: args{
				conditions: []serverlessv1alpha2.Condition{
					{Type: serverlessv1alpha2.ConditionConfigurationReady, Status: corev1.ConditionFalse, Reason: serverlessv1alpha2.ConditionReasonConfigMapCreated},
					{Type: serverlessv1alpha2.ConditionRunning, Status: corev1.ConditionTrue, Reason: serverlessv1alpha2.ConditionReasonServiceCreated},
					{Type: serverlessv1alpha2.ConditionConfigurationReady, Status: corev1.ConditionFalse, Reason: serverlessv1alpha2.ConditionReasonDeploymentWaiting},
				},
				conditionType: serverlessv1alpha2.ConditionRunning,
			},
			want: serverlessv1alpha2.ConditionReasonServiceCreated,
		},
		{
			name: "Should correctly return status unknown if there's no needed conditionType",
			args: args{
				conditions: []serverlessv1alpha2.Condition{
					{Type: serverlessv1alpha2.ConditionConfigurationReady, Status: corev1.ConditionFalse},
					{Type: serverlessv1alpha2.ConditionConfigurationReady, Status: corev1.ConditionFalse},
				},
				conditionType: serverlessv1alpha2.ConditionRunning,
			},
			want: "",
		},
		{
			name: "Should correctly return status unknown if slice is empty",
			args: args{
				conditions:    []serverlessv1alpha2.Condition{},
				conditionType: serverlessv1alpha2.ConditionRunning,
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewGomegaWithT(t)
			got := getConditionReason(tt.args.conditions, tt.args.conditionType)
			g.Expect(got).To(gomega.Equal(tt.want))
		})
	}
}
