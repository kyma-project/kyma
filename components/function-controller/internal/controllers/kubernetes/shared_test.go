package kubernetes

import (
	"context"
	"testing"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"github.com/vrischmann/envconfig"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func Test_isExcludedNamespace(t *testing.T) {
	type args struct {
		name     string
		base     string
		excluded []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "should exclude base namespace",
			args: args{
				name:     "the-same-as-base",
				base:     "the-same-as-base",
				excluded: nil,
			},
			want: true,
		},
		{
			name: "should exclude namespace if it's in excluded list",
			args: args{
				name:     "not-the-same-as-base",
				base:     "the-same-as-base",
				excluded: []string{"data", "tada", "not-the-same-as-base"},
			},
			want: true,
		},
		{
			name: "should exclude namespace if it's in excluded list more than 1 time",
			args: args{
				name:     "not-the-same-as-base",
				base:     "the-same-as-base",
				excluded: []string{"data", "tada", "not-the-same-as-base", "not-the-same-as-base", "not-the-same-as-base"},
			},
			want: true,
		},
		{
			name: "should not exclude otherwise",
			args: args{
				name:     "some-random-value",
				base:     "the-same-as-base",
				excluded: []string{"data", "tada", "not-the-same-as-base"},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isExcludedNamespace(tt.args.name, tt.args.base, tt.args.excluded); got != tt.want {
				t.Errorf("isExcludedNamespace() = %v, want %v", got, tt.want)
			}
		})
	}
}

var _ = ginkgo.Describe("ConfigStruct", func() {
	ginkgo.It("Excluded namespaces should not have length of 1", func() {
		// this test is just to be secure if someone used "," instead of ";"
		// I assume that we have more than 1 namespace we need to exclude by default

		cfg := Config{}
		err := envconfig.Init(&cfg)
		gomega.Expect(err).To(gomega.Succeed())

		gomega.Expect(cfg.ExcludedNamespaces).NotTo(gomega.HaveLen(1))
	})
})

var _ = ginkgo.Describe("getNamespaces", func() {
	var (
		baseNs             = corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "baseNs"}}
		excludedNamespaces = []string{"excluded1", "excluded2"}
		excludedNs1        = corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "excluded1"}}
		ctx                = context.TODO()
		ns1                = corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "test1"}}
		ns2                = corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "test2"}}
		nsTerminating      = corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "terminating"}, Status: corev1.NamespaceStatus{
			Phase: corev1.NamespaceTerminating,
		}}
	)

	ginkgo.It("should successfully return non base, non excluded namespace", func() {
		fakeClient := fake.NewFakeClientWithScheme(scheme.Scheme, &ns1)
		namespaces, err := getNamespaces(ctx, fakeClient, baseNs.Name, excludedNamespaces)
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(namespaces).To(gomega.HaveLen(1))
		gomega.Expect(namespaces[0]).To(gomega.Equal("test1"))
	})
	ginkgo.It("should successfully omit terminating namespace", func() {
		fakeClient := fake.NewFakeClientWithScheme(scheme.Scheme, &ns1, &ns2, &nsTerminating)
		namespaces, err := getNamespaces(ctx, fakeClient, baseNs.Name, excludedNamespaces)
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(namespaces).To(gomega.HaveLen(2))
		gomega.Expect(namespaces).To(gomega.ConsistOf("test1", "test2"))
		gomega.Expect(namespaces).NotTo(gomega.ConsistOf("terminating"))
	})
	ginkgo.It("should successfully omit base namespace", func() {
		fakeClient := fake.NewFakeClientWithScheme(scheme.Scheme, &baseNs)
		namespaces, err := getNamespaces(ctx, fakeClient, baseNs.Name, []string{"random"})
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(namespaces).To(gomega.HaveLen(0))
	})
	ginkgo.It("should successfully excluded namespace", func() {
		fakeClient := fake.NewFakeClientWithScheme(scheme.Scheme, &excludedNs1)
		namespaces, err := getNamespaces(ctx, fakeClient, baseNs.Name, []string{excludedNs1.Name})
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(namespaces).To(gomega.HaveLen(0))
	})
})
