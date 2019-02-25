package asset

import (
	"context"
	"fmt"
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/apis/assetstore/v1alpha2"
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/assethook/engine"
	engineMock "github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/assethook/engine/automock"
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/handler/asset/pretty"
	loaderMock "github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/loader/automock"
	storeMock "github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/store/automock"
	"github.com/onsi/gomega"
	"github.com/pkg/errors"
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

func TestReconcileAssetCreationSuccess(t *testing.T) {
	//Given
	name := "the-tesciak-success"
	testData := newTestData(name)

	mocks := newMocks()
	mocks.loader.On("Load", testData.asset.Spec.Source.Url, testData.assetName, testData.asset.Spec.Source.Mode, testData.asset.Spec.Source.Filter).Return(testData.tmpBaseDir, testData.files, nil).Once()
	mocks.loader.On("Clean", testData.tmpBaseDir).Return(nil).Once()
	mocks.validator.On("Validate", mock.Anything, mock.Anything, testData.tmpBaseDir, testData.files, testData.asset.Spec.Source.ValidationWebhookService).Return(engine.ValidationResult{Success: true}, nil).Once()
	mocks.mutator.On("Mutate", mock.Anything, mock.Anything, testData.tmpBaseDir, testData.files, testData.asset.Spec.Source.MutationWebhookService).Return(nil).Once()
	mocks.store.On("PutObjects", mock.Anything, testData.bucketName, testData.assetName, testData.tmpBaseDir, testData.files).Return(nil).Once()
	mocks.store.On("DeleteObjects", mock.Anything, testData.bucketName, fmt.Sprintf("/%s", testData.assetName)).Return(nil).Once()
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

	//When
	err = c.Create(context.TODO(), testData.asset)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	defer deleteAndExpectSuccess(cfg, testData)

	//Then
	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(testData.request)))
	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(testData.request)))

	asset := &v1alpha2.Asset{}
	err = c.Get(context.TODO(), testData.key, asset)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	g.Expect(asset.Status.Phase).To(gomega.Equal(v1alpha2.AssetReady))
	g.Expect(asset.Status.AssetRef.BaseUrl).To(gomega.Equal(testData.assetUrl))
	g.Expect(asset.Status.AssetRef.Assets).To(gomega.Equal(testData.files))
	g.Expect(asset.Finalizers).To(gomega.ContainElement(deleteAssetFinalizerName))
}

func TestReconcileAssetCreationSuccessMutationFailed(t *testing.T) {
	//Given
	name := "the-tesciak-mutation-fail"
	testData := newTestData(name)

	mocks := newMocks()
	mocks.loader.On("Load", testData.asset.Spec.Source.Url, testData.assetName, testData.asset.Spec.Source.Mode, testData.asset.Spec.Source.Filter).Return(testData.tmpBaseDir, testData.files, nil).Once()
	mocks.loader.On("Clean", testData.tmpBaseDir).Return(nil).Once()
	mocks.mutator.On("Mutate", mock.Anything, mock.Anything, testData.tmpBaseDir, testData.files, testData.asset.Spec.Source.MutationWebhookService).Return(errors.New("surprise!")).Once()
	mocks.store.On("DeleteObjects", mock.Anything, testData.bucketName, fmt.Sprintf("/%s", testData.assetName)).Return(nil).Once()
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

	//When
	err = c.Create(context.TODO(), testData.asset)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	defer deleteAndExpectSuccess(cfg, testData)

	//Then
	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(testData.request)))
	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(testData.request)))

	asset := &v1alpha2.Asset{}
	err = c.Get(context.TODO(), testData.key, asset)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	g.Expect(asset.Status.Phase).To(gomega.Equal(v1alpha2.AssetFailed))
	g.Expect(asset.Status.Reason).To(gomega.Equal(pretty.MutationFailed.String()))
}

func TestReconcileAssetCreationSuccessValidationFailed(t *testing.T) {
	//Given
	name := "the-tesciak-validation-fail"
	testData := newTestData(name)

	mocks := newMocks()
	mocks.loader.On("Load", testData.asset.Spec.Source.Url, testData.assetName, testData.asset.Spec.Source.Mode, testData.asset.Spec.Source.Filter).Return(testData.tmpBaseDir, testData.files, nil).Once()
	mocks.loader.On("Clean", testData.tmpBaseDir).Return(nil).Once()
	mocks.validator.On("Validate", mock.Anything, mock.Anything, testData.tmpBaseDir, testData.files, testData.asset.Spec.Source.ValidationWebhookService).Return(engine.ValidationResult{Success: false}, nil).Once()
	mocks.mutator.On("Mutate", mock.Anything, mock.Anything, testData.tmpBaseDir, testData.files, testData.asset.Spec.Source.MutationWebhookService).Return(nil).Once()
	mocks.store.On("DeleteObjects", mock.Anything, testData.bucketName, fmt.Sprintf("/%s", testData.assetName)).Return(nil).Once()
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

	//When
	err = c.Create(context.TODO(), testData.asset)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	defer deleteAndExpectSuccess(cfg, testData)

	//Then
	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(testData.request)))
	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(testData.request)))

	asset := &v1alpha2.Asset{}
	err = c.Get(context.TODO(), testData.key, asset)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	g.Expect(asset.Status.Phase).To(gomega.Equal(v1alpha2.AssetFailed))
	g.Expect(asset.Status.Reason).To(gomega.Equal(pretty.ValidationFailed.String()))
}

func TestReconcileAssetCreationSuccessNoWebhooks(t *testing.T) {
	//Given
	name := "the-tesciak-no-hooks"
	testData := newTestData(name)
	testData.asset.Spec.Source.MutationWebhookService = nil
	testData.asset.Spec.Source.ValidationWebhookService = nil

	mocks := newMocks()
	mocks.loader.On("Load", testData.asset.Spec.Source.Url, testData.assetName, testData.asset.Spec.Source.Mode, testData.asset.Spec.Source.Filter).Return(testData.tmpBaseDir, testData.files, nil).Once()
	mocks.loader.On("Clean", testData.tmpBaseDir).Return(nil).Once()
	mocks.store.On("PutObjects", mock.Anything, testData.bucketName, testData.assetName, testData.tmpBaseDir, testData.files).Return(nil).Once()
	mocks.store.On("DeleteObjects", mock.Anything, testData.bucketName, fmt.Sprintf("/%s", testData.assetName)).Return(nil).Once()
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

	//When
	err = c.Create(context.TODO(), testData.asset)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	defer deleteAndExpectSuccess(cfg, testData)

	//Then
	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(testData.request)))
	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(testData.request)))

	asset := &v1alpha2.Asset{}
	err = c.Get(context.TODO(), testData.key, asset)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	g.Expect(asset.Status.Phase).To(gomega.Equal(v1alpha2.AssetReady))
	g.Expect(asset.Status.AssetRef.BaseUrl).To(gomega.Equal(testData.assetUrl))
	g.Expect(asset.Status.AssetRef.Assets).To(gomega.Equal(testData.files))
	g.Expect(asset.Finalizers).To(gomega.ContainElement(deleteAssetFinalizerName))
}

func TestReconcileAssetCreationNoBucket(t *testing.T) {
	//Given
	name := "the-tesciak-no-bucket"
	testData := newTestData(name)

	mocks := newMocks()
	defer mocks.AssertExpetactions(t)

	cfg := prepareReconcilerTest(t, mocks)
	g := cfg.g
	c := cfg.c
	defer cfg.finishTest()

	//When
	err := c.Create(context.TODO(), testData.asset)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	defer deleteAndExpectSuccess(cfg, testData)

	//Then
	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(testData.request)))
	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(testData.request)))

	asset := &v1alpha2.Asset{}
	err = c.Get(context.TODO(), testData.key, asset)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	g.Expect(asset.Status.Phase).To(gomega.Equal(v1alpha2.AssetPending))
	g.Expect(asset.Status.Reason).To(gomega.Equal(pretty.BucketNotReady.String()))
}

func TestReconcileAssetReadyLostBucket(t *testing.T) {
	//Given
	name := "the-tesciak-lost-bucket"
	testData := newTestData(name)
	testData.asset.Status.LastHeartbeatTime = v1.NewTime(time.Now().Add(-365 * time.Hour))

	mocks := newMocks()
	defer mocks.AssertExpetactions(t)

	cfg := prepareReconcilerTest(t, mocks)
	g := cfg.g
	c := cfg.c
	defer cfg.finishTest()

	//When
	err := c.Create(context.TODO(), testData.asset)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	defer deleteAndExpectSuccess(cfg, testData)
	err = c.Status().Update(context.TODO(), testData.asset)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	//Then
	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(testData.request)))
	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(testData.request)))

	asset := &v1alpha2.Asset{}
	err = c.Get(context.TODO(), testData.key, asset)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	g.Expect(asset.Status.Phase).To(gomega.Equal(v1alpha2.AssetPending))
	g.Expect(asset.Status.Reason).To(gomega.Equal(pretty.BucketNotReady.String()))
}

func TestReconcileAssetReadyLostFiles(t *testing.T) {
	//Given
	name := "the-tesciak-lost-files"
	testData := newTestData(name)
	testData.asset.Spec.Source.MutationWebhookService = nil
	testData.asset.Spec.Source.ValidationWebhookService = nil
	testData.asset.Status.LastHeartbeatTime = v1.NewTime(time.Now().Add(-365 * time.Hour))

	mocks := newMocks()
	mocks.loader.On("Load", testData.asset.Spec.Source.Url, testData.assetName, testData.asset.Spec.Source.Mode, testData.asset.Spec.Source.Filter).Return(testData.tmpBaseDir, testData.files, nil).Once()
	mocks.loader.On("Clean", testData.tmpBaseDir).Return(nil).Once()
	mocks.store.On("PutObjects", mock.Anything, testData.bucketName, testData.assetName, testData.tmpBaseDir, testData.files).Return(nil).Once()
	mocks.store.On("ContainsAllObjects", mock.Anything, testData.bucketName, testData.assetName, testData.files).Return(false, nil).Once()
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

	//When
	err = c.Create(context.TODO(), testData.asset)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	defer deleteAndExpectSuccess(cfg, testData)
	err = c.Status().Update(context.TODO(), testData.asset)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	//Then
	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(testData.request)))

	asset := &v1alpha2.Asset{}
	err = c.Get(context.TODO(), testData.key, asset)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	g.Expect(asset.Status.Phase).To(gomega.Equal(v1alpha2.AssetFailed))
	g.Expect(asset.Status.Reason).To(gomega.Equal(pretty.MissingContent.String()))

	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(testData.request)))

	asset = &v1alpha2.Asset{}
	err = c.Get(context.TODO(), testData.key, asset)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	g.Expect(asset.Status.Phase).To(gomega.Equal(v1alpha2.AssetReady))
}

type mocks struct {
	store     *storeMock.Store
	loader    *loaderMock.Loader
	validator *engineMock.Validator
	mutator   *engineMock.Mutator
}

func newMocks() *mocks {
	return &mocks{
		store:     new(storeMock.Store),
		loader:    new(loaderMock.Loader),
		validator: new(engineMock.Validator),
		mutator:   new(engineMock.Mutator),
	}
}

func (m *mocks) AssertExpetactions(t *testing.T) {
	m.store.AssertExpectations(t)
	m.validator.AssertExpectations(t)
	m.loader.AssertExpectations(t)
	m.mutator.AssertExpectations(t)
}

type testData struct {
	namespace  string
	assetName  string
	assetUrl   string
	bucketName string
	bucketUrl  string
	key        types.NamespacedName

	tmpBaseDir string
	files      []string

	bucket *v1alpha2.Bucket
	asset  *v1alpha2.Asset

	request reconcile.Request
}

func newTestData(name string) *testData {
	namespace := fmt.Sprintf("%s-space", name)
	assetName := fmt.Sprintf("%s-asset", name)
	bucketName := fmt.Sprintf("%s-bucket", name)
	bucketUrl := fmt.Sprintf("https://minio.%s.local/%s", name, bucketName)
	assetUrl := fmt.Sprintf("%s/%s", bucketUrl, assetName)
	sourceUrl := fmt.Sprintf("https://source.%s.local/file.zip", name)
	key := types.NamespacedName{Name: assetName, Namespace: namespace}
	tmpBaseDir := fmt.Sprintf("/tmp/%s", assetName)

	files := []string{
		fmt.Sprintf("%s.zip", name),
		fmt.Sprintf("%s.json", name),
		fmt.Sprintf("%s.yaml", name),
		fmt.Sprintf("%s.md", name),
		fmt.Sprintf("lvl/%s.md", name),
		fmt.Sprintf("lvl/lvl/%s.md", name),
	}

	bucket := &v1alpha2.Bucket{
		ObjectMeta: v1.ObjectMeta{
			Name:      bucketName,
			Namespace: namespace,
		},
		Status: v1alpha2.BucketStatus{CommonBucketStatus: v1alpha2.CommonBucketStatus{
			Phase:             v1alpha2.BucketReady,
			RemoteName:        bucketName,
			Url:               bucketUrl,
			LastHeartbeatTime: v1.Now(),
		}},
	}

	asset := &v1alpha2.Asset{
		ObjectMeta: v1.ObjectMeta{
			Name:      assetName,
			Namespace: namespace,
		},
		Spec: v1alpha2.AssetSpec{CommonAssetSpec: v1alpha2.CommonAssetSpec{
			BucketRef: v1alpha2.AssetBucketRef{Name: bucketName},
			Source: v1alpha2.AssetSource{
				Url:  sourceUrl,
				Mode: v1alpha2.AssetSingle,
				ValidationWebhookService: []v1alpha2.AssetWebhookService{
					{
						Namespace: "test",
						Name:      "test",
						Endpoint:  "/test",
					},
					{
						Namespace: "test",
						Name:      "test",
						Endpoint:  "/test",
					},
				},
				MutationWebhookService: []v1alpha2.AssetWebhookService{
					{
						Namespace: "test",
						Name:      "test",
						Endpoint:  "/test",
					},
					{
						Namespace: "test",
						Name:      "test",
						Endpoint:  "/test",
					},
				},
			},
		}},
		Status: v1alpha2.AssetStatus{CommonAssetStatus: v1alpha2.CommonAssetStatus{
			ObservedGeneration: int64(1),
			Phase:              v1alpha2.AssetReady,
			AssetRef: v1alpha2.AssetStatusRef{
				BaseUrl: assetUrl,
				Assets:  files,
			},
			LastHeartbeatTime: v1.Now(),
		}},
	}

	request := reconcile.Request{NamespacedName: types.NamespacedName{Name: assetName, Namespace: namespace}}

	return &testData{
		namespace:  namespace,
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

func deleteBucket(cfg *testSuite, bucket *v1alpha2.Bucket) {
	g := cfg.g
	c := cfg.c
	err := c.Delete(context.TODO(), bucket)
	g.Expect(err).NotTo(gomega.HaveOccurred())
}

func deleteAndExpectSuccess(cfg *testSuite, testData *testData) {
	g := cfg.g
	c := cfg.c
	err := c.Delete(context.TODO(), testData.asset)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(testData.request)))
	g.Eventually(func() bool {
		instance := &v1alpha2.Asset{}
		err := c.Get(context.TODO(), testData.key, instance)
		return apierrors.IsNotFound(err)
	}, timeout, 10*time.Millisecond).Should(gomega.BeTrue())
}
