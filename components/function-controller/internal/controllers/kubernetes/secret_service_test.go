package kubernetes

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_secretService_IsBase(t *testing.T) {
	baseNs := "base-ns"

	tests := []struct {
		name string

		secret *corev1.Secret
		want   bool
	}{
		{
			name: "should properly detect base secret",
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: baseNs,
					Labels: map[string]string{
						ConfigLabel: CredentialsLabelValue,
					}},
			},
			want: true,
		},
		{
			name: "should return false for secret without needed labels",
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: baseNs,
				}},
			want: false,
		},
		{
			name: "should return false for secret in wrong namespace",
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "some-other-ns",
					Labels: map[string]string{
						ConfigLabel: CredentialsLabelValue,
					}},
			},
			want: false,
		},
		{
			name: "should return false for secret in some other namespace and with no labels",
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "blabla-other-ns",
				}},
			want: false,
		},
		{
			name: "should return false for secret with wrong label value",
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: baseNs,
					Labels: map[string]string{
						ConfigLabel: "wrong-label-value",
					}},
			},
			want: false,
		},
		{
			name: "should return false for secret with wrong label key",
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: baseNs,
					Labels: map[string]string{
						"some-weird-label-key": CredentialsLabelValue,
					}},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &secretService{
				config: Config{
					BaseNamespace: baseNs,
				},
			}
			if got := r.IsBase(tt.secret); got != tt.want {
				t.Errorf("IsBase() = %v, want %v", got, tt.want)
			}
		})
	}
}
