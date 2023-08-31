package serverless

import (
	"context"
	"testing"

	"go.uber.org/zap"

	"github.com/kyma-project/kyma/components/function-controller/internal/resource/automock"
	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	serverlessv1alpha2 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha2"
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
			got := equalServices(tt.args.existing, tt.args.expected)
			g.Expect(got).To(gomega.Equal(tt.want))
		})
	}
}

func TestFunctionReconciler_deleteExcessServices(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		g := gomega.NewGomegaWithT(t)

		instance := &serverlessv1alpha2.Function{
			ObjectMeta: metav1.ObjectMeta{Name: "fn-name"},
		}

		services := []corev1.Service{
			{ObjectMeta: metav1.ObjectMeta{Name: "fn-name"}},
			{ObjectMeta: metav1.ObjectMeta{Name: "fn-some-other-name"}},
		}

		client := new(automock.Client)
		client.On("Delete", context.TODO(), &services[1]).Return(nil).Once()
		defer client.AssertExpectations(t)

		s := systemState{
			instance: *instance,
			services: corev1.ServiceList{
				Items: services,
			},
		}

		r := reconciler{
			log: zap.NewNop().Sugar(),
			k8s: k8s{
				client: client,
			},
		}

		_, err := stateFnDeleteServices(context.TODO(), &r, &s)

		g.Expect(err).To(gomega.Succeed())
		g.Expect(client.Calls).To(gomega.HaveLen(1), "delete should happen only for service which has different name than it's parent fn")
	})

	t.Run("should delete both svc that have different name than fn", func(t *testing.T) {
		g := gomega.NewGomegaWithT(t)

		instance := &serverlessv1alpha2.Function{
			ObjectMeta: metav1.ObjectMeta{Name: "fn-name"},
		}

		services := []corev1.Service{
			{ObjectMeta: metav1.ObjectMeta{Name: "fn-other-name"}},
			{ObjectMeta: metav1.ObjectMeta{Name: "fn-some-other-name"}},
		}

		client := new(automock.Client)
		client.On("Delete", context.TODO(), &services[0]).Return(nil).Once()
		client.On("Delete", context.TODO(), &services[1]).Return(nil).Once()
		defer client.AssertExpectations(t)

		s := systemState{
			instance: *instance,
			services: corev1.ServiceList{
				Items: services,
			},
		}

		r := reconciler{
			log: zap.NewNop().Sugar(),
			k8s: k8s{
				client: client,
			},
		}

		_, err := stateFnDeleteServices(context.TODO(), &r, &s)

		g.Expect(err).To(gomega.Succeed())
		g.Expect(client.Calls).To(gomega.HaveLen(2), "delete should happen only for service which has different name than it's parent fn")
	})
}
