package controllers

import (
	"context"
	"time"

	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/apis/assetstore/v1alpha2"
	"github.com/kyma-project/kyma/components/cms-controller-manager/internal/source"
	"github.com/kyma-project/kyma/components/cms-controller-manager/pkg/apis/cms/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
)

var _ = Describe("Asset", func() {
	var (
		docstopic  *v1alpha1.DocsTopic
		reconciler *DocsTopicReconciler
		request    ctrl.Request
	)

	BeforeEach(func() {
		docstopic = newFixDocsTopic()
		Expect(k8sClient.Create(context.TODO(), docstopic)).To(Succeed())

		request = ctrl.Request{
			NamespacedName: types.NamespacedName{
				Name:      docstopic.Name,
				Namespace: docstopic.Namespace,
			},
		}

		assetService := newAssetService(k8sClient, scheme.Scheme)
		bucketService := newBucketService(k8sClient, scheme.Scheme, "us-east-1")

		reconciler = &DocsTopicReconciler{
			Client:           k8sClient,
			Log:              ctrl.Log,
			recorder:         record.NewFakeRecorder(100),
			relistInterval:   60 * time.Hour,
			assetSvc:         assetService,
			bucketSvc:        bucketService,
			webhookConfigSvc: webhookConfigSvc,
		}
	})

	It("should successfully create, update and delete DocsTopic", func() {
		By("creating the DocsTopic")
		result, err := reconciler.Reconcile(request)
		validateReconcilation(err, result)
		docstopic = &v1alpha1.DocsTopic{}
		Expect(k8sClient.Get(context.TODO(), request.NamespacedName, docstopic)).To(Succeed())
		validateDocsTopic(docstopic.Status.CommonDocsTopicStatus, docstopic.ObjectMeta, v1alpha1.DocsTopicPending, v1alpha1.DocsTopicWaitingForAssets)

		By("assets changes states to ready")
		assets := &v1alpha2.AssetList{}
		Expect(k8sClient.List(context.TODO(), assets)).To(Succeed())
		Expect(assets.Items).To(HaveLen(len(docstopic.Spec.Sources)))

		for _, asset := range assets.Items {
			asset.Status.Phase = v1alpha2.AssetReady
			asset.Status.LastHeartbeatTime = v1.Now()
			Expect(k8sClient.Status().Update(context.TODO(), &asset)).To(Succeed())

			if asset.Annotations["cms.kyma-project.io/asset-short-name"] == "source-one" {
				Expect(asset.Spec.Parameters).ToNot(BeNil())
				Expect(asset.Spec.Parameters).To(Equal(&fixParameters))
			} else {
				Expect(asset.Spec.Parameters).To(BeNil())
			}
		}

		result, err = reconciler.Reconcile(request)
		validateReconcilation(err, result)
		docstopic = &v1alpha1.DocsTopic{}
		Expect(k8sClient.Get(context.TODO(), request.NamespacedName, docstopic)).To(Succeed())
		validateDocsTopic(docstopic.Status.CommonDocsTopicStatus, docstopic.ObjectMeta, v1alpha1.DocsTopicReady, v1alpha1.DocsTopicAssetsReady)

		By("updating the DocsTopic")
		docstopic.Spec.Sources = source.FilterByType(docstopic.Spec.Sources, "dita")
		markdownIndex := source.IndexByType(docstopic.Spec.Sources, "markdown")
		Expect(markdownIndex).NotTo(Equal(-1))
		docstopic.Spec.Sources[markdownIndex].Filter = "zyx"
		Expect(k8sClient.Update(context.TODO(), docstopic)).To(Succeed())

		result, err = reconciler.Reconcile(request)
		validateReconcilation(err, result)
		docstopic = &v1alpha1.DocsTopic{}
		Expect(k8sClient.Get(context.TODO(), request.NamespacedName, docstopic)).To(Succeed())
		validateDocsTopic(docstopic.Status.CommonDocsTopicStatus, docstopic.ObjectMeta, v1alpha1.DocsTopicPending, v1alpha1.DocsTopicWaitingForAssets)

		assets = &v1alpha2.AssetList{}
		Expect(k8sClient.List(context.TODO(), assets)).To(Succeed())
		Expect(assets.Items).To(HaveLen(len(docstopic.Spec.Sources)))
		for _, a := range assets.Items {
			if a.Annotations["cms.kyma-project.io/asset-short-name"] != "source-two" {
				continue
			}
			Expect(a.Spec.Source.Filter).To(Equal("zyx"))
		}

		By("deleting the DocsTopic")
		Expect(k8sClient.Delete(context.TODO(), docstopic)).To(Succeed())

		_, err = reconciler.Reconcile(request)
		Expect(err).To(Succeed())

		docstopic = &v1alpha1.DocsTopic{}
		err = k8sClient.Get(context.TODO(), request.NamespacedName, docstopic)
		Expect(err).To(HaveOccurred())
		Expect(apiErrors.IsNotFound(err)).To(BeTrue())

	})
})

func newFixDocsTopic() *v1alpha1.DocsTopic {
	return &v1alpha1.DocsTopic{
		ObjectMeta: ctrl.ObjectMeta{
			Name:      string(uuid.NewUUID()),
			Namespace: "default",
		},
		Spec: v1alpha1.DocsTopicSpec{
			CommonDocsTopicSpec: v1alpha1.CommonDocsTopicSpec{
				Description: "Test topic, have fun",
				DisplayName: "Test Topic",
				Sources: []v1alpha1.Source{
					{
						Name:       "source-one",
						Type:       "openapi",
						Mode:       v1alpha1.DocsTopicSingle,
						URL:        "https://dummy.url/single",
						Parameters: &fixParameters,
					},
					{
						Name:   "source-two",
						Type:   "markdown",
						Filter: "xyz",
						Mode:   v1alpha1.DocsTopicPackage,
						URL:    "https://dummy.url/package",
					},
					{
						Name:   "source-three",
						Type:   "dita",
						Filter: "xyz",
						Mode:   v1alpha1.DocsTopicIndex,
						URL:    "https://dummy.url/index",
					},
					{
						Name: "source-four",
						Type: "openapi",
						Mode: v1alpha1.DocsTopicPackage,
						URL:  "https://dummy.url/single",
					},
				},
			},
		},
		Status: v1alpha1.DocsTopicStatus{CommonDocsTopicStatus: v1alpha1.CommonDocsTopicStatus{
			LastHeartbeatTime: v1.Now(),
		}},
	}
}
