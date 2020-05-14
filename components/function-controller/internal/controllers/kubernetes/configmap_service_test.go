package kubernetes

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_configMapService_IsBase(t *testing.T) {
	baseNs := "base-ns"

	type args struct {
		configmap *corev1.ConfigMap
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "should correctly return if ConfigMap is base one",
			args: args{configmap: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{Namespace: baseNs, Labels: map[string]string{
					ConfigLabel: RuntimeLabelValue,
				}},
			}},
			want: true,
		},
		{
			name: "should correctly return false for ConfigMap in wrong ns",
			args: args{configmap: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{Namespace: "not-base-ns", Labels: map[string]string{
					ConfigLabel: RuntimeLabelValue,
				}},
			}},
			want: false,
		},
		{
			name: "should correctly return false for ConfigMap has wrong label value",
			args: args{configmap: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{Namespace: baseNs, Labels: map[string]string{
					ConfigLabel: "some-random-value",
				}},
			}},
			want: false,
		},
		{
			name: "should correctly return false for ConfigMap with no labels",
			args: args{configmap: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{Namespace: baseNs},
			}},
			want: false,
		},
		{
			name: "should correctly return false for ConfigMap with no labels and in wrong namespace",
			args: args{configmap: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{Namespace: "not-base"},
			}},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &configMapService{
				config: Config{
					BaseNamespace: baseNs,
				},
			}
			if got := r.IsBase(tt.args.configmap); got != tt.want {
				t.Errorf("IsBase() = %v, want %v", got, tt.want)
			}
		})
	}
}
