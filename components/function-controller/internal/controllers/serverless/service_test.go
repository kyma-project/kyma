package serverless

import (
	"testing"

	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestFunctionReconciler_equalServices(t *testing.T) {
	type args struct {
		existing corev1.Service
		expected corev1.Service
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "simple case",
			args: args{
				existing: corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "svc-name",
						Namespace: "svc-ns",
						Labels: map[string]string{
							"label1": "label1",
						},
					},
					Spec: corev1.ServiceSpec{
						Ports: []corev1.ServicePort{{
							Name:       "http",
							Port:       80,
							TargetPort: intstr.FromInt(80)},
						},
						Selector: map[string]string{
							"selector": "sel-value",
						},
					},
				},
				expected: corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "svc-name",
						Namespace: "svc-ns",
						Labels: map[string]string{
							"label1": "label1",
						},
					},
					Spec: corev1.ServiceSpec{
						Ports: []corev1.ServicePort{{
							Name:       "http",
							Port:       80,
							TargetPort: intstr.FromInt(80)},
						},
						Selector: map[string]string{
							"selector": "sel-value",
						},
					},
				},
			},
			want: true,
		},
		{
			name: "fails on different labels",
			args: args{
				existing: corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "svc-name",
						Namespace: "svc-ns",
						Labels: map[string]string{
							"label1": "label1",
						},
					},
					Spec: corev1.ServiceSpec{
						Ports: []corev1.ServicePort{{
							Name:       "http",
							Port:       80,
							TargetPort: intstr.FromInt(80)},
						},
						Selector: map[string]string{
							"selector": "sel-value",
						},
					},
				},
				expected: corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "svc-name",
						Namespace: "svc-ns",
						Labels: map[string]string{
							"different": "label-different",
						},
					},
					Spec: corev1.ServiceSpec{
						Ports: []corev1.ServicePort{{
							Name:       "http",
							Port:       80,
							TargetPort: intstr.FromInt(80)},
						},
						Selector: map[string]string{
							"selector": "sel-value",
						},
					},
				},
			},
			want: false,
		},
		{
			name: "fails on different port",
			args: args{
				existing: corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "svc-name",
						Namespace: "svc-ns",
						Labels: map[string]string{
							"label1": "label1",
						},
					},
					Spec: corev1.ServiceSpec{
						Ports: []corev1.ServicePort{{
							Name:       "http",
							Port:       8000,
							TargetPort: intstr.FromInt(80)},
						},
						Selector: map[string]string{
							"selector": "sel-value",
						},
					},
				},
				expected: corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "svc-name",
						Namespace: "svc-ns",
						Labels: map[string]string{
							"label1": "label1",
						},
					},
					Spec: corev1.ServiceSpec{
						Ports: []corev1.ServicePort{{
							Name:       "http",
							Port:       80,
							TargetPort: intstr.FromInt(80)},
						},
						Selector: map[string]string{
							"selector": "sel-value",
						},
					},
				},
			},
			want: false,
		},
		{
			name: "fails on different port name",
			args: args{
				existing: corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "svc-name",
						Namespace: "svc-ns",
						Labels: map[string]string{
							"label1": "label1",
						},
					},
					Spec: corev1.ServiceSpec{
						Ports: []corev1.ServicePort{{
							Name:       "http",
							Port:       80,
							TargetPort: intstr.FromInt(80)},
						},
						Selector: map[string]string{
							"selector": "sel-value",
						},
					},
				},
				expected: corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "svc-name",
						Namespace: "svc-ns",
						Labels: map[string]string{
							"label1": "label1",
						},
					},
					Spec: corev1.ServiceSpec{
						Ports: []corev1.ServicePort{{
							Name:       "httpzzzz",
							Port:       80,
							TargetPort: intstr.FromInt(80)},
						},
						Selector: map[string]string{
							"selector": "sel-value",
						},
					},
				},
			},
			want: false,
		},
		{
			name: "fails on different targetPort",
			args: args{
				existing: corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "svc-name",
						Namespace: "svc-ns",
						Labels: map[string]string{
							"label1": "label1",
						},
					},
					Spec: corev1.ServiceSpec{
						Ports: []corev1.ServicePort{{
							Name:       "http",
							Port:       80,
							TargetPort: intstr.FromInt(80)},
						},
						Selector: map[string]string{
							"selector": "sel-value",
						},
					},
				},
				expected: corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "svc-name",
						Namespace: "svc-ns",
						Labels: map[string]string{
							"label1": "label1",
						},
					},
					Spec: corev1.ServiceSpec{
						Ports: []corev1.ServicePort{{
							Name:       "http",
							Port:       80,
							TargetPort: intstr.FromInt(666)},
						},
						Selector: map[string]string{
							"selector": "sel-value",
						},
					},
				},
			},
			want: false,
		},
		{
			name: "fails on different selector",
			args: args{
				existing: corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "svc-name",
						Namespace: "svc-ns",
						Labels: map[string]string{
							"label1": "label1",
						},
					},
					Spec: corev1.ServiceSpec{
						Ports: []corev1.ServicePort{{
							Name:       "http",
							Port:       80,
							TargetPort: intstr.FromInt(80)},
						},
						Selector: map[string]string{
							"selector": "sel-value",
						},
					},
				},
				expected: corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "svc-name",
						Namespace: "svc-ns",
						Labels: map[string]string{
							"label1": "label1",
						},
					},
					Spec: corev1.ServiceSpec{
						Ports: []corev1.ServicePort{{
							Name:       "http",
							Port:       80,
							TargetPort: intstr.FromInt(80)},
						},
						Selector: map[string]string{
							"selector": "sel-value-DIFFERENT",
						},
					},
				},
			},
			want: false,
		},
		{
			name: "fails if there is 0 ports in existing",
			args: args{
				existing: corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "svc-name",
						Namespace: "svc-ns",
						Labels: map[string]string{
							"label1": "label1",
						},
					},
					Spec: corev1.ServiceSpec{
						Ports: []corev1.ServicePort{},
						Selector: map[string]string{
							"selector": "sel-value",
						},
					},
				},
				expected: corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "svc-name",
						Namespace: "svc-ns",
						Labels: map[string]string{
							"label1": "label1",
						},
					},
					Spec: corev1.ServiceSpec{
						Ports: []corev1.ServicePort{{
							Name:       "http",
							Port:       80,
							TargetPort: intstr.FromInt(80)},
						},
						Selector: map[string]string{
							"selector": "sel-value",
						},
					},
				},
			},
			want: false,
		},
		{
			name: "fails if there is 0 ports in expected",
			args: args{
				existing: corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "svc-name",
						Namespace: "svc-ns",
						Labels: map[string]string{
							"label1": "label1",
						},
					},
					Spec: corev1.ServiceSpec{
						Ports: []corev1.ServicePort{{
							Name: "test",
						}},
						Selector: map[string]string{
							"selector": "sel-value",
						},
					},
				},
				expected: corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "svc-name",
						Namespace: "svc-ns",
						Labels: map[string]string{
							"label1": "label1",
						},
					},
					Spec: corev1.ServiceSpec{
						Ports: []corev1.ServicePort{},
						Selector: map[string]string{
							"selector": "sel-value",
						},
					},
				},
			},
			want: false,
		},
		{
			name: "fails if there is 0 ports in either case",
			args: args{
				existing: corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "svc-name",
						Namespace: "svc-ns",
						Labels: map[string]string{
							"label1": "label1",
						},
					},
					Spec: corev1.ServiceSpec{
						Selector: map[string]string{
							"selector": "sel-value",
						},
					},
				},
				expected: corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "svc-name",
						Namespace: "svc-ns",
						Labels: map[string]string{
							"label1": "label1",
						},
					},
					Spec: corev1.ServiceSpec{
						Selector: map[string]string{
							"selector": "sel-value",
						},
					},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewGomegaWithT(t)
			r := &FunctionReconciler{}
			got := r.equalServices(tt.args.existing, tt.args.expected)
			g.Expect(got).To(gomega.Equal(tt.want))
		})
	}
}
