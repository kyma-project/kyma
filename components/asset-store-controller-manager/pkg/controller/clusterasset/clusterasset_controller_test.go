package clusterasset

import (
	"context"
	"fmt"
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/apis/assetstore/v1alpha2"
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/assethook/engine"
	engineMock "github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/assethook/engine/automock"
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/handler/asset/pretty"
	loaderMock "github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/loader/automock"
	storeMock "github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/store/automock"
	"github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var c client.Client

const timeout = time.Second * 10

func TestAdd(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		//Given
		g := gomega.NewGomegaWithT(t)

		accessKeyName := "APP_STORE_ACCESS_KEY"
		secretKeyName := "APP_STORE_SECRET_KEY"
		originalAccessKey := os.Getenv(accessKeyName)
		originalSecretKey := os.Getenv(secretKeyName)

		defer func() {
			err := os.Setenv(accessKeyName, originalAccessKey)
			g.Expect(err).ShouldNot(gomega.HaveOccurred())
			err = os.Setenv(secretKeyName, originalSecretKey)
			g.Expect(err).ShouldNot(gomega.HaveOccurred())
		}()

		err := os.Setenv(accessKeyName, "test")
		g.Expect(err).NotTo(gomega.HaveOccurred())
		err = os.Setenv(secretKeyName, "test")

		g.Expect(err).NotTo(gomega.HaveOccurred())

		mgr, err := manager.New(cfg, manager.Options{})
		g.Expect(err).NotTo(gomega.HaveOccurred())

		//When
		err = Add(mgr)

		//Then
		g.Expect(err).NotTo(gomega.HaveOccurred())
	})

	t.Run("Error", func(t *testing.T) {
		g := gomega.NewGomegaWithT(t)
		err := Add(nil)
		g.Expect(err).To(gomega.HaveOccurred())
	})
}

func TestReconcileClusterAsset_Reconcile(t *testing.T) {
	//Given
	name := "test-asset"
	newFilter := "abcd"
	testData := newTestData(name)

	mocks := newMocks()
	mocks.loader.On("Load", testData.asset.Spec.Source.URL, testData.assetName, testData.asset.Spec.Source.Mode, testData.asset.Spec.Source.Filter).Return(testData.tmpBaseDir, testData.files, nil).Once()
	mocks.loader.On("Load", testData.asset.Spec.Source.URL, testData.assetName, testData.asset.Spec.Source.Mode, newFilter).Return(testData.tmpBaseDir, testData.files, nil).Once()
	mocks.loader.On("Clean", testData.tmpBaseDir).Return(nil).Twice()
	mocks.validator.On("Validate", mock.Anything, mock.Anything, testData.tmpBaseDir, testData.files, testData.asset.Spec.Source.ValidationWebhookService).Return(engine.ValidationResult{Success: true}, nil).Twice()
	mocks.mutator.On("Mutate", mock.Anything, mock.Anything, testData.tmpBaseDir, testData.files, testData.asset.Spec.Source.MutationWebhookService).Return(nil).Twice()
	mocks.store.On("PutObjects", mock.Anything, testData.bucketName, testData.assetName, testData.tmpBaseDir, testData.files).Return(nil).Twice()
	mocks.store.On("ListObjects", mock.Anything, testData.bucketName, fmt.Sprintf("/%s", testData.assetName)).Return(nil, nil).Times(3)
	defer mocks.AssertExpetactions(t)

	cfg := prepareReconcilerTest(t, mocks)
	g := cfg.g
	c := cfg.c
	defer cfg.finishTest()

	err := c.Create(context.TODO(), testData.bucket)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	defer deleteBucket(cfg, testData.bucket)
	err = c.Status().Update(context.TODO(), testData.bucket)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	//When - Create
	err = c.Create(context.TODO(), testData.asset)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	//Then

	// Created
	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(testData.request)))

	asset := &v1alpha2.ClusterAsset{}
	err = c.Get(context.TODO(), testData.key, asset)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	validateClusterAsset(cfg, asset, v1alpha2.AssetPending, pretty.Scheduled, "", nil)

	// Ready
	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(testData.request)))
	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(testData.request)))

	asset = &v1alpha2.ClusterAsset{}
	err = c.Get(context.TODO(), testData.key, asset)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	validateClusterAsset(cfg, asset, v1alpha2.AssetReady, pretty.Uploaded, testData.assetUrl, testData.files)

	// When - update
	updatedAsset := asset.DeepCopy()
	updatedAsset.Spec.Source.Filter = newFilter
	err = c.Update(context.TODO(), updatedAsset)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	// Updated
	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(testData.request)))

	asset = &v1alpha2.ClusterAsset{}
	err = c.Get(context.TODO(), testData.key, asset)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	validateClusterAsset(cfg, asset, v1alpha2.AssetPending, pretty.Scheduled, "", nil)

	// Ready
	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(testData.request)))
	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(testData.request)))

	asset = &v1alpha2.ClusterAsset{}
	err = c.Get(context.TODO(), testData.key, asset)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	validateClusterAsset(cfg, asset, v1alpha2.AssetReady, pretty.Uploaded, testData.assetUrl, testData.files)

	// When - delete
	err = c.Delete(context.TODO(), asset)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	// Then
	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(testData.request)))
	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(testData.request)))

	asset = &v1alpha2.ClusterAsset{}
	err = c.Get(context.TODO(), testData.key, asset)
	g.Expect(err).To(gomega.HaveOccurred())
	g.Expect(apierrors.IsNotFound(err)).To(gomega.BeTrue())
}

func validateClusterAsset(cfg *testSuite, instance *v1alpha2.ClusterAsset, phase v1alpha2.AssetPhase, reason pretty.Reason, baseUrl string, files []string) {
	g := cfg.g

	g.Expect(instance.Status.Phase).To(gomega.Equal(phase))
	g.Expect(instance.Status.Reason).To(gomega.Equal(reason.String()))
	g.Expect(instance.Status.AssetRef.BaseURL).To(gomega.Equal(baseUrl))
	g.Expect(instance.Status.AssetRef.Files).To(gomega.HaveLen(len(files)))
	g.Expect(instance.Finalizers).To(gomega.ContainElement(deleteClusterAssetFinalizerName))
}

type mocks struct {
	store             *storeMock.Store
	loader            *loaderMock.Loader
	validator         *engineMock.Validator
	mutator           *engineMock.Mutator
	metadataExtractor *engineMock.MetadataExtractor
}

func newMocks() *mocks {
	return &mocks{
		store:             new(storeMock.Store),
		loader:            new(loaderMock.Loader),
		validator:         new(engineMock.Validator),
		mutator:           new(engineMock.Mutator),
		metadataExtractor: new(engineMock.MetadataExtractor),
	}
}

func (m *mocks) AssertExpetactions(t *testing.T) {
	m.store.AssertExpectations(t)
	m.validator.AssertExpectations(t)
	m.loader.AssertExpectations(t)
	m.mutator.AssertExpectations(t)
	m.metadataExtractor.AssertExpectations(t)
}

type testData struct {
	assetName  string
	assetUrl   string
	bucketName string
	bucketUrl  string
	key        types.NamespacedName

	tmpBaseDir string
	files      []string

	bucket *v1alpha2.ClusterBucket
	asset  *v1alpha2.ClusterAsset

	request reconcile.Request
}

func newTestData(name string) *testData {
	assetName := fmt.Sprintf("%s-asset", name)
	bucketName := fmt.Sprintf("%s-bucket", name)
	bucketUrl := fmt.Sprintf("https://minio.%s.local/%s", name, bucketName)
	assetUrl := fmt.Sprintf("%s/%s", bucketUrl, assetName)
	sourceUrl := fmt.Sprintf("https://source.%s.local/file.zip", name)
	key := types.NamespacedName{Name: assetName, Namespace: ""}
	tmpBaseDir := fmt.Sprintf("/tmp/%s", assetName)

	files := []string{
		fmt.Sprintf("%s.zip", name),
		fmt.Sprintf("%s.json", name),
		fmt.Sprintf("%s.yaml", name),
		fmt.Sprintf("%s.md", name),
		fmt.Sprintf("lvl/%s.md", name),
		fmt.Sprintf("lvl/lvl/%s.md", name),
	}

	bucket := &v1alpha2.ClusterBucket{
		ObjectMeta: v1.ObjectMeta{
			Name: bucketName,
		},
		Status: v1alpha2.ClusterBucketStatus{CommonBucketStatus: v1alpha2.CommonBucketStatus{
			Phase:             v1alpha2.BucketReady,
			RemoteName:        bucketName,
			URL:               bucketUrl,
			LastHeartbeatTime: v1.Now(),
		}},
	}

	asset := &v1alpha2.ClusterAsset{
		ObjectMeta: v1.ObjectMeta{
			Name: assetName,
		},
		Spec: v1alpha2.ClusterAssetSpec{CommonAssetSpec: v1alpha2.CommonAssetSpec{
			BucketRef: v1alpha2.AssetBucketRef{Name: bucketName},
			Source: v1alpha2.AssetSource{
				URL:  sourceUrl,
				Mode: v1alpha2.AssetSingle,
				ValidationWebhookService: []v1alpha2.AssetWebhookService{
					{
						WebhookService: v1alpha2.WebhookService{
							Namespace: "test",
							Name:      "test",
							Endpoint:  "/test",
						},
					},
					{WebhookService: v1alpha2.WebhookService{
						Namespace: "test",
						Name:      "test",
						Endpoint:  "/test",
					},
					},
				},
				MutationWebhookService: []v1alpha2.AssetWebhookService{
					{
						WebhookService: v1alpha2.WebhookService{
							Namespace: "test",
							Name:      "test",
							Endpoint:  "/test",
						},
					},
					{
						WebhookService: v1alpha2.WebhookService{
							Namespace: "test",
							Name:      "test",
							Endpoint:  "/test",
						},
					},
				},
			},
		}},
	}

	request := reconcile.Request{NamespacedName: types.NamespacedName{Name: assetName, Namespace: ""}}

	return &testData{
		assetName:  assetName,
		assetUrl:   assetUrl,
		bucketName: bucketName,
		bucketUrl:  bucketUrl,
		tmpBaseDir: tmpBaseDir,
		files:      files,
		key:        key,
		bucket:     bucket,
		asset:      asset,
		request:    request,
	}
}

func deleteBucket(cfg *testSuite, bucket *v1alpha2.ClusterBucket) {
	g := cfg.g
	c := cfg.c
	err := c.Delete(context.TODO(), bucket)
	g.Expect(err).NotTo(gomega.HaveOccurred())
}
