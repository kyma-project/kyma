package docstopic

import (
	"context"
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/apis/assetstore/v1alpha2"
	"github.com/kyma-project/kyma/components/cms-controller-manager/pkg/apis/cms/v1alpha1"
	"github.com/kyma-project/kyma/components/cms-controller-manager/pkg/handler/docstopic/pretty"
	"github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var c client.Client

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

	r := &ReconcileDocsTopic{
		relistInterval: time.Hour,
		Client:         c,
		scheme:         scheme,
		recorder:       mgr.GetRecorder("docstopic-controller"),
		assetSvc:       assetService,
		bucketSvc:      bucketService,
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
	delete(currentTopic.Spec.Sources, "dita")
	markdown := currentTopic.Spec.Sources["markdown"]
	markdown.Filter = "zyx"
	currentTopic.Spec.Sources["markdown"] = markdown
	err = c.Update(context.TODO(), currentTopic)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	// Then
	// Update Assets
	g.Eventually(requests, timeout).Should(gomega.Receive(gomega.Equal(request)))
	g.Eventually(requests, timeout).Should(gomega.Receive(gomega.Equal(request)))

	assets = &v1alpha2.AssetList{}
	err = c.List(context.TODO(), nil, assets)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(assets.Items).To(gomega.HaveLen(len(currentTopic.Spec.Sources)))
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
				Sources: map[string]v1alpha1.Source{
					"openapi": v1alpha1.Source{
						Mode: v1alpha1.DocsTopicSingle,
						URL:  "https://dummy.url/single",
					},
					"markdown": v1alpha1.Source{
						Filter: "xyz",
						Mode:   v1alpha1.DocsTopicPackage,
						URL:    "https://dummy.url/package",
					},
					"dita": v1alpha1.Source{
						Filter: "xyz",
						Mode:   v1alpha1.DocsTopicIndex,
						URL:    "https://dummy.url/index",
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
