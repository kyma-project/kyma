package docstopic

import (
	"context"
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/apis/assetstore/v1alpha2"
	"github.com/kyma-project/kyma/components/cms-controller-manager/pkg/apis/cms/v1alpha1"
	"github.com/kyma-project/kyma/components/cms-controller-manager/pkg/handler/docstopic/pretty"
	"github.com/kyma-project/kyma/components/cms-controller-manager/pkg/source"
	"github.com/kyma-project/kyma/components/cms-controller-manager/pkg/webhookconfig"
	"github.com/onsi/gomega"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var (
	webhookCfgMapName      = "test"
	webhookCfgMapNamespace = "test"
)

const timeout = time.Second * 5

func TestReconcile(t *testing.T) {
	// Given
	g := gomega.NewGomegaWithT(t)

	mgr, err := manager.New(cfg, manager.Options{})
	g.Expect(err).ToNot(gomega.HaveOccurred())
	c := mgr.GetClient()
	scheme := mgr.GetScheme()
	assetService := newAssetService(c, scheme)
	bucketService := newBucketService(c, scheme, "")
	informer, err := mgr.GetCache().GetInformer(&coreV1.ConfigMap{})
	g.Expect(err).To(gomega.BeNil())
	assetWhsConfigService := webhookconfig.New(informer.GetIndexer(), webhookCfgMapName, webhookCfgMapNamespace)

	r := &ReconcileDocsTopic{
		Client:           c,
		scheme:           scheme,
		relistInterval:   time.Hour,
		recorder:         mgr.GetRecorder("docstopic-controller"),
		assetSvc:         assetService,
		bucketSvc:        bucketService,
		webhookConfigSvc: assetWhsConfigService,
	}

	recFn, requests := SetupTestReconcile(r)
	err = add(mgr, recFn)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	stopMgr, mgrStopped := StartTestManager(mgr, g)
	defer func() {
		close(stopMgr)
		mgrStopped.Wait()
	}()

	topic := fixDocsTopic()
	request := fixRequest(topic)

	// When
	err = c.Create(context.TODO(), topic)

	// Then
	g.Expect(err).ToNot(gomega.HaveOccurred())
	defer c.Delete(context.TODO(), topic)

	// DocsTopic creation
	g.Eventually(requests, timeout).Should(gomega.Receive(gomega.Equal(request)))
	// Update DocsTopic status after asstes creation
	g.Eventually(requests, timeout).Should(gomega.Receive(gomega.Equal(request)))

	currentTopic := &v1alpha1.DocsTopic{}
	err = c.Get(context.TODO(), request.NamespacedName, currentTopic)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(currentTopic.Status.Phase).To(gomega.Equal(v1alpha1.DocsTopicPending))
	g.Expect(currentTopic.Status.Reason).To(gomega.Equal(pretty.WaitingForAssets.String()))

	assets := &v1alpha2.AssetList{}
	err = c.List(context.TODO(), nil, assets)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(assets.Items).To(gomega.HaveLen(len(topic.Spec.Sources)))

	// Updating assets statuses
	// When
	for _, asset := range assets.Items {
		asset.Status.Phase = v1alpha2.AssetReady
		asset.Status.LastHeartbeatTime = v1.Now()
		err = c.Status().Update(context.TODO(), &asset)
		g.Expect(err).ToNot(gomega.HaveOccurred())
		g.Eventually(requests, timeout).Should(gomega.Receive(gomega.Equal(request)))
	}

	// Update DocsTopic status
	// Then
	g.Eventually(requests, timeout).Should(gomega.Receive(gomega.Equal(request)))

	currentTopic = &v1alpha1.DocsTopic{}
	err = c.Get(context.TODO(), request.NamespacedName, currentTopic)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(currentTopic.Status.Phase).To(gomega.Equal(v1alpha1.DocsTopicReady))
	g.Expect(currentTopic.Status.Reason).To(gomega.Equal(pretty.AssetsReady.String()))

	// Update DocsTopic spec
	// When
	currentTopic.Spec.Sources = source.FilterByType(currentTopic.Spec.Sources, "dita")
	markdownIndex := source.IndexByType(currentTopic.Spec.Sources, "markdown")
	g.Expect(markdownIndex).NotTo(gomega.Equal(-1))
	currentTopic.Spec.Sources[markdownIndex].Filter = "zyx"
	err = c.Update(context.TODO(), currentTopic)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	// Then
	// Update Assets
	g.Eventually(requests, timeout).Should(gomega.Receive(gomega.Equal(request)))

	assets = &v1alpha2.AssetList{}
	err = c.List(context.TODO(), nil, assets)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(assets.Items).To(gomega.HaveLen(len(currentTopic.Spec.Sources)))
	for _, a := range assets.Items {
		if a.Annotations["assetshortname.cms.kyma-project.io"] != "source-two" {
			continue
		}
		g.Expect(a.Spec.Source.Filter).To(gomega.Equal("zyx"))
	}

}

func fixDocsTopic() *v1alpha1.DocsTopic {
	return &v1alpha1.DocsTopic{
		ObjectMeta: v1.ObjectMeta{
			Name:      "test-name",
			Namespace: "test",
		},
		Spec: v1alpha1.DocsTopicSpec{
			CommonDocsTopicSpec: v1alpha1.CommonDocsTopicSpec{
				Description: "Test topic, have fun",
				DisplayName: "Test Topic",
				Sources: []v1alpha1.Source{
					{
						Name: "source-one",
						Type: "openapi",
						Mode: v1alpha1.DocsTopicSingle,
						URL:  "https://dummy.url/single",
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
	}
}

func fixRequest(topic *v1alpha1.DocsTopic) reconcile.Request {
	return reconcile.Request{
		NamespacedName: types.NamespacedName{Name: topic.Name, Namespace: topic.Namespace},
	}
}
