package kubernetes

import (
	"testing"

	rbacv1 "k8s.io/api/rbac/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_roleBindingService_IsBase(t *testing.T) {
	baseNs := "base-ns"

	type args struct {
		role *rbacv1.RoleBinding
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "should correctly return if Role is base one",
			args: args{role: &rbacv1.RoleBinding{
				ObjectMeta: metav1.ObjectMeta{Namespace: baseNs, Labels: map[string]string{
					RbacLabel: RoleBindingLabelValue,
				}},
			}},
			want: true,
		},
		{
			name: "should correctly return false for Role in wrong ns",
			args: args{role: &rbacv1.RoleBinding{
				ObjectMeta: metav1.ObjectMeta{Namespace: "not-base-ns", Labels: map[string]string{
					RbacLabel: RoleBindingLabelValue,
				}},
			}},
			want: false,
		},
		{
			name: "should correctly return false for Role has wrong label value",
			args: args{role: &rbacv1.RoleBinding{
				ObjectMeta: metav1.ObjectMeta{Namespace: baseNs, Labels: map[string]string{
					RbacLabel: "some-random-value",
				}},
			}},
			want: false,
		},
		{
			name: "should correctly return false for Role with no labels",
			args: args{role: &rbacv1.RoleBinding{
				ObjectMeta: metav1.ObjectMeta{Namespace: baseNs},
			}},
			want: false,
		},
		{
			name: "should correctly return false for Role with no labels and in wrong namespace",
			args: args{role: &rbacv1.RoleBinding{
				ObjectMeta: metav1.ObjectMeta{Namespace: "not-base"},
			}},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &roleBindingService{
				config: Config{
					BaseNamespace: baseNs,
				},
			}
			if got := r.IsBase(tt.args.role); got != tt.want {
				t.Errorf("IsBase() = %v, want %v", got, tt.want)
			}
		})
	}
}
