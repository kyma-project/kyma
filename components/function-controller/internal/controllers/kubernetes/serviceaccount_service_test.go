package kubernetes

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/kyma-project/kyma/components/function-controller/internal/resource"
	"github.com/kyma-project/kyma/components/function-controller/internal/resource/automock"
)

func Test_serviceAccountService_extractSecretTokens(t *testing.T) {
	type args struct {
		serviceAccount *corev1.ServiceAccount
	}
	tests := []struct {
		name string
		args args
		want []corev1.ObjectReference
	}{
		{
			name: "should return secret with the same prefix",
			args: args{serviceAccount: &corev1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{Name: "test-name"},
				Secrets:    []corev1.ObjectReference{{Name: "test-name-token-123"}},
			}},
			want: []corev1.ObjectReference{{Name: "test-name-token-123"}},
		},
		{
			name: "should not return secret with improper prefix",
			args: args{serviceAccount: &corev1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{Name: "test-name"},
				Secrets:    []corev1.ObjectReference{{Name: "super-secret-secret"}},
			}},
			want: []corev1.ObjectReference{},
		},
		{
			name: "should return multiple correct secrets",
			args: args{serviceAccount: &corev1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{Name: "test-name"},
				Secrets:    []corev1.ObjectReference{{Name: "test-name-token-123-blah1"}, {Name: "test-name-token-123-blah2"}, {Name: "random-one"}},
			}},
			want: []corev1.ObjectReference{{Name: "test-name-token-123-blah1"}, {Name: "test-name-token-123-blah2"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			se := &serviceAccountService{}
			if got := se.extractSecretTokens(tt.args.serviceAccount); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("extractSecretTokens() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_serviceAccountService_shiftSecretTokens(t *testing.T) {
	type args struct {
		serviceAccount *corev1.ServiceAccount
	}
	tests := []struct {
		name string
		args args
		want []corev1.ObjectReference
	}{
		{
			name: "should not return secret with the same prefix",
			args: args{serviceAccount: &corev1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{Name: "test-name"},
				Secrets:    []corev1.ObjectReference{{Name: "test-name-token-123"}},
			}},
			want: []corev1.ObjectReference{},
		},
		{
			name: "should return secret with prefix that doesn't match",
			args: args{serviceAccount: &corev1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{Name: "test-name"},
				Secrets:    []corev1.ObjectReference{{Name: "super-secret-secret"}},
			}},
			want: []corev1.ObjectReference{{Name: "super-secret-secret"}},
		},

		{
			name: "should return multiple correct secrets",
			args: args{serviceAccount: &corev1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{Name: "test-name"},
				Secrets:    []corev1.ObjectReference{{Name: "test-name-token-123-blah1"}, {Name: "test-name-token-123-blah2"}, {Name: "random-one"}, {Name: "random-two"}},
			}},
			want: []corev1.ObjectReference{{Name: "random-one"}, {Name: "random-two"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			se := &serviceAccountService{}
			if got := se.shiftSecretTokens(tt.args.serviceAccount); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("shiftSecretTokens() = %v, want %v", got, tt.want)
			}
		})
	}
}

var _ = ginkgo.Describe("updateServiceAccount method", func() {
	var (
		ctx = context.TODO()
	)

	ginkgo.It("should update serviceAccount merging two svcAcc together", func() {
		client := new(automock.Client)

		var obj resource.Object

		client.On("Update", mock.Anything, mock.Anything).Return(nil).Once().Run(func(args mock.Arguments) {
			obj = args.Get(1).(resource.Object)
		})
		defer client.AssertExpectations(ginkgo.GinkgoT())

		r := &serviceAccountService{client: client}

		truethy := true

		base := &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name: "base",
				Labels: map[string]string{
					"base-label-1": "label-1",
					"base-label-2": "label-2",
				},
				Annotations: map[string]string{
					"base-anno-key-1": "anno-1",
				},
			},
			Secrets:                      []corev1.ObjectReference{{Name: "base-secret-1"}, {Name: "base-secret-1"}},
			ImagePullSecrets:             []corev1.LocalObjectReference{{Name: "base-image-pull-secret"}, {Name: "base-image-pull-secret-2"}},
			AutomountServiceAccountToken: &truethy,
		}

		falsy := false

		instance := &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name: "instance",
				Labels: map[string]string{
					"instance-label-1": "label-1",
					"instance-label-2": "label-2",
				},
				Annotations: map[string]string{
					"instance-anno-key-1": "anno-1",
				},
			},
			Secrets:                      []corev1.ObjectReference{{Name: "instance-secret-1"}, {Name: "instance-secret-1"}},
			ImagePullSecrets:             []corev1.LocalObjectReference{{Name: "instance-image-pull-secret"}, {Name: "instance-image-pull-secret-2"}},
			AutomountServiceAccountToken: &falsy,
		}

		err := r.updateServiceAccount(ctx, log.Log, instance, base)
		gomega.Expect(err).To(gomega.Succeed())

		gomega.Expect(obj).NotTo(gomega.BeNil())

		updatedServiceAcc := obj.(*corev1.ServiceAccount)

		// inherited from instance
		gomega.Expect(updatedServiceAcc.Name).To(gomega.Equal(instance.GetName()))

		// inherited from base
		gomega.Expect(updatedServiceAcc.Annotations).To(gomega.And(gomega.Equal(base.GetAnnotations()), gomega.Not(gomega.BeNil())))
		gomega.Expect(updatedServiceAcc.Labels).To(gomega.And(gomega.Equal(base.GetLabels()), gomega.Not(gomega.BeNil())))
		gomega.Expect(updatedServiceAcc.ImagePullSecrets).To(gomega.And(gomega.Equal(base.ImagePullSecrets), gomega.Not(gomega.BeNil())))
		gomega.Expect(updatedServiceAcc.AutomountServiceAccountToken).To(gomega.And(gomega.Equal(base.AutomountServiceAccountToken), gomega.Not(gomega.BeNil())))
		gomega.Expect(updatedServiceAcc.Secrets).To(gomega.And(gomega.Equal(base.Secrets), gomega.Not(gomega.BeNil())))
		gomega.Expect(updatedServiceAcc.Secrets).NotTo(gomega.Equal(instance.Secrets))

		gomega.Expect(updatedServiceAcc.AutomountServiceAccountToken).NotTo(gomega.Equal(instance.AutomountServiceAccountToken))
	})
	ginkgo.It("should correctly extract token and normal secrets", func() {
		client := new(automock.Client)

		var obj resource.Object

		client.On("Update", mock.Anything, mock.Anything).Return(nil).Once().Run(func(args mock.Arguments) {
			obj = args.Get(1).(resource.Object)
		})

		defer client.AssertExpectations(ginkgo.GinkgoT())

		r := &serviceAccountService{client: client}

		base := &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{Name: "base"},
			Secrets:    []corev1.ObjectReference{{Name: "base-secret-1"}, {Name: "base-secret-2"}},
		}

		instance := &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{Name: "instance"},
			Secrets:    []corev1.ObjectReference{{Name: "instance-token-1"}, {Name: "instance-token-2"}},
		}

		err := r.updateServiceAccount(ctx, log.Log, instance, base)
		gomega.Expect(err).To(gomega.Succeed())

		gomega.Expect(obj).NotTo(gomega.BeNil())

		updatedServiceAcc := obj.(*corev1.ServiceAccount)

		gomega.Expect(updatedServiceAcc.Name).To(gomega.Equal(instance.GetName()))
		gomega.Expect(updatedServiceAcc.Secrets).To(gomega.Equal([]corev1.ObjectReference{
			{Name: "base-secret-1"},
			{Name: "base-secret-2"},
			{Name: "instance-token-1"},
			{Name: "instance-token-2"},
		}))
	})
	ginkgo.It("should return error on update error", func() {
		client := new(automock.Client)

		client.On("Update", mock.Anything, mock.Anything).Return(errors.New("update err")).Once()

		r := &serviceAccountService{client: client}

		base := &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{Name: "base"},
			Secrets:    []corev1.ObjectReference{{Name: "base-secret-1"}, {Name: "base-secret-2"}},
		}

		instance := &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{Name: "instance"},
			Secrets:    []corev1.ObjectReference{{Name: "instance-token-1"}, {Name: "instance-token-2"}},
		}

		err := r.updateServiceAccount(ctx, log.Log, instance, base)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})
})

func Test_serviceAccountService_IsBase(t *testing.T) {
	baseNs := "base-ns"

	type args struct {
		serviceAccount *corev1.ServiceAccount
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "should correctly return if service account is base one",
			args: args{serviceAccount: &corev1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{Namespace: baseNs, Labels: map[string]string{
					ConfigLabel: ServiceAccountLabelValue,
				}},
			}},
			want: true,
		},
		{
			name: "should correctly return false for service account in wrong ns",
			args: args{serviceAccount: &corev1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{Namespace: "not-base-ns", Labels: map[string]string{
					ConfigLabel: ServiceAccountLabelValue,
				}},
			}},
			want: false,
		},
		{
			name: "should correctly return false for service account has wrong label value",
			args: args{serviceAccount: &corev1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{Namespace: baseNs, Labels: map[string]string{
					ConfigLabel: "some-random-value",
				}},
			}},
			want: false,
		},
		{
			name: "should correctly return false for service account with no labels",
			args: args{serviceAccount: &corev1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{Namespace: baseNs},
			}},
			want: false,
		},
		{
			name: "should correctly return false for service account with no labels and in wrong namespace",
			args: args{serviceAccount: &corev1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{Namespace: "not-base"},
			}},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &serviceAccountService{
				config: Config{
					BaseNamespace: baseNs,
				},
			}
			if got := r.IsBase(tt.args.serviceAccount); got != tt.want {
				t.Errorf("IsBase() = %v, want %v", got, tt.want)
			}
		})
	}
}

var _ = ginkgo.Describe("createServiceAccount method", func() {
	var (
		ctx = context.TODO()
	)

	ginkgo.It("should create configmap correctly", func() {
		client := new(automock.Client)

		var obj resource.Object

		client.On("Create", mock.Anything, mock.Anything).Return(nil).Once().Run(func(args mock.Arguments) {
			obj = args.Get(1).(resource.Object)
		})
		defer client.AssertExpectations(ginkgo.GinkgoT())

		r := &serviceAccountService{client: client}

		truethy := true

		base := &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "base",
				Namespace: "original-ns",
				Labels: map[string]string{
					"base-label-1": "label-1",
					"base-label-2": "label-2",
				},
				Annotations: map[string]string{
					"base-anno-key-1": "anno-1",
				},
			},
			Secrets:                      []corev1.ObjectReference{{Name: "base-secret-1"}, {Name: "base-secret-1"}, {Name: "base-token-1"}, {Name: "base-token-2"}, {Name: "base-token-3"}},
			ImagePullSecrets:             []corev1.LocalObjectReference{{Name: "base-image-pull-secret"}, {Name: "base-image-pull-secret-2"}},
			AutomountServiceAccountToken: &truethy,
		}

		namespace := "some-ns"

		err := r.createServiceAccount(ctx, log.Log, namespace, base)
		gomega.Expect(err).To(gomega.Succeed())

		gomega.Expect(obj).NotTo(gomega.BeNil())

		createdServiceAcc := obj.(*corev1.ServiceAccount)

		gomega.Expect(createdServiceAcc.Name).To(gomega.Equal(base.GetName()))
		gomega.Expect(createdServiceAcc.Namespace).To(gomega.Equal(namespace))
		gomega.Expect(createdServiceAcc.Namespace).NotTo(gomega.Equal(base.Namespace))
		gomega.Expect(createdServiceAcc.Annotations).To(gomega.And(gomega.Equal(base.GetAnnotations()), gomega.Not(gomega.BeNil())))
		gomega.Expect(createdServiceAcc.Labels).To(gomega.And(gomega.Equal(base.GetLabels()), gomega.Not(gomega.BeNil())))
		gomega.Expect(createdServiceAcc.ImagePullSecrets).To(gomega.And(gomega.Equal(base.ImagePullSecrets), gomega.Not(gomega.BeNil())))
		gomega.Expect(createdServiceAcc.AutomountServiceAccountToken).To(gomega.And(gomega.Equal(base.AutomountServiceAccountToken), gomega.Not(gomega.BeNil())))
		gomega.Expect(createdServiceAcc.Secrets).NotTo(gomega.BeNil())
		gomega.Expect(createdServiceAcc.Secrets).NotTo(gomega.Equal(base.Secrets), "there should be not tokens here, as they are autogenerated by k8s")
		gomega.Expect(createdServiceAcc.Secrets).To(gomega.Equal([]corev1.ObjectReference{{Name: "base-secret-1"}, {Name: "base-secret-1"}}))
	})
	ginkgo.It("should return error on update error", func() {
		client := new(automock.Client)

		client.On("Create", mock.Anything, mock.Anything).Return(errors.New("update err")).Once()

		r := &serviceAccountService{client: client}

		base := &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{Name: "base"},
			Secrets:    []corev1.ObjectReference{{Name: "base-secret-1"}, {Name: "base-secret-2"}},
		}

		err := r.createServiceAccount(ctx, log.Log, "random-ns", base)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})
})
