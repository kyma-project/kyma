package knative

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/onsi/ginkgo"
	gm "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"

	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
)

const (
	testCleanupLabelKey   = "clean-me-up-key"
	testCleanupLabelValue = "clean-me-up-value"
)

func Test_getNewestGeneration(t *testing.T) {
	type args struct {
		revisions []servingv1.Revision
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{
			name: "should return highest generation from label",
			args: args{
				revisions: []servingv1.Revision{
					{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{cfgGenerationLabel: "1"}}},
					{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{cfgGenerationLabel: "2"}}},
					{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{cfgGenerationLabel: "3"}}},
				},
			},
			want: 3,
		},
		{
			name: "should return error if even one revision lacks proper label",
			args: args{
				revisions: []servingv1.Revision{
					{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{cfgGenerationLabel: "1"}}},
					{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{cfgGenerationLabel: "2"}}},
					{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"random-label": "3"}}},
				},
			},
			want:    -1,
			wantErr: true,
		},
		{
			name: "should return error generation is not a number",
			args: args{
				revisions: []servingv1.Revision{
					{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{cfgGenerationLabel: "1"}}},
					{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{cfgGenerationLabel: "NaN"}}},
				},
			},
			want:    -1,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gm.NewGomegaWithT(t)

			got, err := getNewestGeneration(tt.args.revisions)
			g.Expect(err != nil).To(gm.Equal(tt.wantErr))
			g.Expect(got).To(gm.Equal(tt.want))
		})
	}
}

func TestFunctionReconciler_getOldRevisionSelector(t *testing.T) {
	type args struct {
		parentService string
		revisions     []servingv1.Revision
	}
	tests := []struct {
		name    string
		args    args
		want    labels.Selector
		wantErr bool
	}{
		{
			name: "properly parses revisions and service name",
			args: args{
				parentService: "testServiceName",
				revisions: []servingv1.Revision{
					{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{cfgGenerationLabel: "1"}}},
					{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{cfgGenerationLabel: "2"}}},
					{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{cfgGenerationLabel: "3"}}},
				},
			},
			wantErr: false,
			want: func() labels.Selector {
				lbl, _ := labels.Parse("serving.knative.dev/service=testServiceName,serving.knative.dev/configurationGeneration!=3")
				return lbl
			}(),
		},
		{
			name: "fails with incorrect labels from revisions",
			args: args{
				parentService: "testServiceName",
				revisions: []servingv1.Revision{
					{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{cfgGenerationLabel: "1"}}},
					{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{cfgGenerationLabel: "2"}}},
					{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{cfgGenerationLabel: "ups"}}},
				},
			},
			wantErr: true,
			want:    nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gm.NewGomegaWithT(t)

			r := &ServiceReconciler{}
			got, err := r.getOldRevisionSelector(tt.args.parentService, tt.args.revisions)

			g.Expect(err != nil).To(gm.Equal(tt.wantErr))
			if got != nil {
				g.Expect(got).To(gm.Equal(tt.want))
			} else {
				g.Expect(got).To(gm.BeNil())
			}
		})
	}
}

var _ = ginkgo.Describe("KService controller", func() {
	var (
		reconciler        *ServiceReconciler
		request           ctrl.Request
		ctx               = context.TODO()
		srvName           = "test-service"
		numberOfRevisions = 5
		namespace         = "serverless"
	)

	ginkgo.BeforeEach(func() {
		reconciler = NewServiceReconciler(resourceClient, log.Log, config)
		request = ctrl.Request{NamespacedName: types.NamespacedName{Namespace: namespace, Name: srvName}}
	})
	ginkgo.AfterEach(func() {
		sel := labels.SelectorFromSet(map[string]string{testCleanupLabelKey: testCleanupLabelValue})
		err := resourceClient.DeleteAllBySelector(ctx, &servingv1.Revision{}, namespace, sel)
		gm.Expect(err).To(gm.BeNil())
		err = resourceClient.DeleteAllBySelector(ctx, &servingv1.Service{}, namespace, sel)
		gm.Expect(err).To(gm.BeNil())
	})

	ginkgo.It("should cleanup old revisions, leaving newest one", func() {
		ginkgo.By("Creating test resources")
		srv := fixKservice(srvName, namespace, true)
		err := resourceClient.Create(ctx, &srv)
		gm.Expect(err).NotTo(gm.HaveOccurred(), "failed to create test KService resource")

		for _, rev := range fixRevisionList(srvName, namespace, numberOfRevisions) {
			pinnedRev := rev // pin
			gm.Expect(resourceClient.Create(ctx, &pinnedRev)).NotTo(gm.HaveOccurred())
			gm.Expect(resourceClient.Status().Update(ctx, &pinnedRev)).NotTo(gm.HaveOccurred())
		}

		ginkgo.By("should skip reconcile on service creation")
		_, err = reconciler.Reconcile(request)
		gm.Expect(err).NotTo(gm.HaveOccurred())

		initialRevList := &servingv1.RevisionList{}
		err = reconciler.client.ListByLabel(ctx, srv.GetNamespace(), map[string]string{serviceLabelKey: srv.GetName()}, initialRevList)
		gm.Expect(initialRevList.Items).To(gm.HaveLen(numberOfRevisions))

		ginkgo.By("Update service to be ready")
		srv.Status.Status.Conditions = duckv1.Conditions{{
			Type:   servingv1.ServiceConditionReady,
			Status: corev1.ConditionTrue,
		}}

		gm.Expect(resourceClient.Update(ctx, &srv)).Should(gm.Succeed())
		gm.Expect(resourceClient.Status().Update(ctx, &srv)).Should(gm.Succeed())

		ginkgo.By("waiting for controller to delete excess revisions")
		_, err = reconciler.Reconcile(request)
		gm.Expect(err).NotTo(gm.HaveOccurred())

		newRevisionList := &servingv1.RevisionList{}
		err = reconciler.client.ListByLabel(ctx, srv.GetNamespace(), map[string]string{serviceLabelKey: srv.GetName()}, newRevisionList)
		gm.Expect(newRevisionList.Items).To(gm.HaveLen(1))

		ginkgo.By("verify that the only revision left is the correct one")
		cfgLabelValue, ok := newRevisionList.Items[0].Labels[cfgGenerationLabel]
		gm.Expect(ok).To(gm.BeTrue())
		gm.Expect(cfgLabelValue).To(gm.Equal(strconv.Itoa(numberOfRevisions)))
	})
	ginkgo.It("should ignore service that has no proper labels", func() {
		ginkgo.By("Creating test resources")
		srv := fixKservice(srvName, namespace, false)
		err := resourceClient.Create(ctx, &srv)
		gm.Expect(err).NotTo(gm.HaveOccurred(), "failed to create test KService resource")

		for _, rev := range fixRevisionList(srvName, namespace, numberOfRevisions) {
			pinnedRev := rev // pin
			gm.Expect(resourceClient.Create(ctx, &pinnedRev)).NotTo(gm.HaveOccurred())
			gm.Expect(resourceClient.Status().Update(ctx, &pinnedRev)).NotTo(gm.HaveOccurred())
		}

		ginkgo.By("should skip reconcile on service creation")
		_, err = reconciler.Reconcile(request)
		gm.Expect(err).NotTo(gm.HaveOccurred())

		initialRevList := &servingv1.RevisionList{}
		err = reconciler.client.ListByLabel(ctx, srv.GetNamespace(), map[string]string{serviceLabelKey: srv.GetName()}, initialRevList)
		gm.Expect(initialRevList.Items).To(gm.HaveLen(numberOfRevisions))

		ginkgo.By("Update service to be ready")
		srv.Status.Status.Conditions = duckv1.Conditions{{
			Type:   servingv1.ServiceConditionReady,
			Status: corev1.ConditionTrue,
		}}
		gm.Expect(resourceClient.Update(ctx, &srv)).Should(gm.Succeed())
		gm.Expect(resourceClient.Status().Update(ctx, &srv)).Should(gm.Succeed())

		ginkgo.By("waiting for controller to ignore revisions for ksvc without proper labels")
		_, err = reconciler.Reconcile(request)
		gm.Expect(err).NotTo(gm.HaveOccurred())

		newRevisionList := &servingv1.RevisionList{}
		err = reconciler.client.ListByLabel(ctx, srv.GetNamespace(), map[string]string{serviceLabelKey: srv.GetName()}, newRevisionList)
		gm.Expect(newRevisionList.Items).To(gm.HaveLen(numberOfRevisions))
	})
})

func fixRevisionList(parentSvcName, namespace string, num int) []servingv1.Revision {
	revList := []servingv1.Revision{}
	for i := 0; i < num; i++ {
		revision := servingv1.Revision{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace,
				Name:      fmt.Sprintf("test-revision-%d", i+1),
				Labels: map[string]string{
					testCleanupLabelKey: testCleanupLabelValue,
					serviceLabelKey:     parentSvcName,
					cfgGenerationLabel:  strconv.Itoa(i + 1), // just like in real revision
				},
			},
		}

		revList = append(revList, revision)
	}
	return revList
}

func fixKservice(name, namespace string, addLables bool) servingv1.Service {
	svcLabels := map[string]string{
		testCleanupLabelKey: testCleanupLabelValue,
	}

	if addLables {
		svcLabels = map[string]string{
			serverlessv1alpha1.FunctionNameLabel:      "whatever-name",
			serverlessv1alpha1.FunctionUUIDLabel:      "whatever-uuid",
			serverlessv1alpha1.FunctionManagedByLabel: "whatever-managed-by",
			testCleanupLabelKey:                       testCleanupLabelValue,
		}
	}

	return servingv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    svcLabels,
		},
	}
}

func Test_hasCorrectLabels(t *testing.T) {
	tests := []struct {
		name string
		arg  servingv1.Service
		want bool
	}{
		{
			name: "pass on all required labels",
			want: true,
			arg: servingv1.Service{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{
					serverlessv1alpha1.FunctionNameLabel:      "whatever-name",
					serverlessv1alpha1.FunctionUUIDLabel:      "whatever-uuid",
					serverlessv1alpha1.FunctionManagedByLabel: "whatever-managed-by",
				}},
			},
		},
		{
			name: "pass on all required labels + extraneous one",
			want: true,
			arg: servingv1.Service{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{
					serverlessv1alpha1.FunctionNameLabel:      "whatever-name",
					serverlessv1alpha1.FunctionUUIDLabel:      "whatever-uuid",
					serverlessv1alpha1.FunctionManagedByLabel: "whatever-managed-by",
					"extra-label": "extra-value",
				}},
			},
		},
		{
			name: "one label missing",
			want: false,
			arg: servingv1.Service{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{
					serverlessv1alpha1.FunctionNameLabel:      "whatever-name",
					serverlessv1alpha1.FunctionManagedByLabel: "whatever-managed-by",
				}},
			},
		},
		{
			name: "all function-controller labels missing",
			want: false,
			arg:  servingv1.Service{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasCorrectLabels(tt.arg); got != tt.want {
				t.Errorf("hasCorrectLabels() = %v, want %v", got, tt.want)
			}
		})
	}
}
