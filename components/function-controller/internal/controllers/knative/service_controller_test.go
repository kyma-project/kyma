package knative

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/onsi/ginkgo"
	gm "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const timeout = time.Second * 15

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
	ctx := context.TODO()
	srvName := "test-service"
	ns := SetupTest(ctx)
	numberOfRevisions := 5

	ginkgo.It("should cleanup old revisions, leaving newest one", func() {
		ginkgo.By("Creating test resources")
		srv := fixKservice(srvName, ns.Name)
		err := k8sClient.Create(ctx, &srv)
		gm.Expect(err).NotTo(gm.HaveOccurred(), "failed to create test KService resource")

		revisionList := fixRevisionList(srvName, ns.Name, numberOfRevisions)
		for _, rev := range revisionList {
			pinnedRev := rev // pin
			gm.Expect(k8sClient.Create(context.TODO(), &pinnedRev)).NotTo(gm.HaveOccurred())
			gm.Expect(k8sClient.Status().Update(context.TODO(), &pinnedRev)).NotTo(gm.HaveOccurred())
		}

		ginkgo.By("Update service to be ready")
		srv.Status.Status.Conditions = duckv1.Conditions{{
			Type:   servingv1.ServiceConditionReady,
			Status: corev1.ConditionTrue,
		}}

		gm.Expect(k8sClient.Update(context.TODO(), &srv)).Should(gm.Succeed())
		gm.Expect(k8sClient.Status().Update(context.TODO(), &srv)).Should(gm.Succeed())

		ginkgo.By("waiting for controller to delete excess revisions")
		revList := servingv1.RevisionList{}
		gm.Eventually(func() int {
			_ = k8sClient.List(ctx, &revList, &client.ListOptions{
				LabelSelector: labels.SelectorFromSet(map[string]string{
					serviceLabelKey: srvName,
				}),
				Namespace: ns.Name,
			})
			return len(revList.Items)
		}, 5*time.Second, 500*time.Millisecond).Should(gm.Equal(1))

		ginkgo.By("verify that the only revision left is the correct one")
		cfgLabelValue, ok := revList.Items[0].Labels[cfgGenerationLabel]
		gm.Expect(ok).To(gm.BeTrue())
		gm.Expect(cfgLabelValue).To(gm.Equal(strconv.Itoa(numberOfRevisions)))
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
					serviceLabelKey:    parentSvcName,
					cfgGenerationLabel: strconv.Itoa(i + 1),
				},
			},
			Spec: servingv1.RevisionSpec{},
		}

		revList = append(revList, revision)
	}
	return revList
}

func fixKservice(name, namespace string) servingv1.Service {
	return servingv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    nil,
		},
		Spec: servingv1.ServiceSpec{
			ConfigurationSpec: servingv1.ConfigurationSpec{
				Template: servingv1.RevisionTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{},
					Spec: servingv1.RevisionSpec{
						PodSpec: corev1.PodSpec{},
					},
				},
			},
		},
	}
}
