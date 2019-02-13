package asset

import (
	"context"
	"fmt"
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/apis/assetstore/v1alpha1"
	automock2 "github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/cleaner/automock"
	automock3 "github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/controller/asset/bucket/automock"
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/controller/asset/webhook"
	automock4 "github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/controller/asset/webhook/automock"
	automock5 "github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/loader/automock"
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/store"
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/uploader/automock"
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

const timeout = time.Hour * 5

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
	name := "the-tesciak"
	testData := newTestData(name)
	testData.asset.Status = v1alpha1.AssetStatus{LastHeartbeatTime: v1.Now()}

	mocks := newMocks()
	mocks.bucketLister.On("Get", testData.namespace, testData.bucketName).Return(testData.bucket, nil).Times(4)
	mocks.loader.On("Load", testData.asset.Spec.Source.Url, testData.assetName, testData.asset.Spec.Source.Mode, testData.asset.Spec.Source.Filter).Return(testData.tmpBaseDir, testData.files, nil).Once()
	mocks.loader.On("Clean", testData.tmpBaseDir).Return(nil).Once()
	mocks.validator.On("Validate", mock.Anything, testData.tmpBaseDir, testData.files, mock.Anything).Return(webhook.ValidationResult{Success: true}, nil).Once()
	mocks.mutator.On("Mutate", mock.Anything, testData.tmpBaseDir, testData.files, mock.Anything).Return(nil).Once()
	mocks.uploader.On("Upload", mock.Anything, testData.minioBucketName, testData.assetName, testData.tmpBaseDir, testData.files).Return(nil).Once()
	mocks.uploader.On("ContainsAll", testData.minioBucketName, testData.assetName, testData.files).Return(true, nil).Once()
	mocks.cleaner.On("Clean", mock.Anything, testData.minioBucketName, fmt.Sprintf("%s/", testData.assetName)).Return(nil).Once()
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
	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(testData.request)))

	asset := &v1alpha1.Asset{}
	err = c.Get(context.TODO(), testData.key, asset)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	g.Expect(asset.Status.Phase).To(gomega.Equal(v1alpha1.AssetReady))
	g.Expect(asset.Status.AssetRef.BaseUrl).To(gomega.Equal(testData.assetUrl))
	g.Expect(asset.Status.AssetRef.Assets).To(gomega.Equal(testData.files))
	g.Expect(asset.Finalizers).To(gomega.ContainElement(deleteAssetFinalizerName))
}

func TestReconcileAssetCreationSuccessMutationFailed(t *testing.T) {
	//Given
	name := "the-tesciak"
	testData := newTestData(name)
	testData.asset.Status = v1alpha1.AssetStatus{LastHeartbeatTime: v1.Now()}

	mocks := newMocks()
	mocks.bucketLister.On("Get", testData.namespace, testData.bucketName).Return(testData.bucket, nil).Times(3)
	mocks.loader.On("Load", testData.asset.Spec.Source.Url, testData.assetName, testData.asset.Spec.Source.Mode, testData.asset.Spec.Source.Filter).Return(testData.tmpBaseDir, testData.files, nil).Once()
	mocks.loader.On("Clean", testData.tmpBaseDir).Return(nil).Once()
	mocks.validator.On("Validate", mock.Anything, testData.tmpBaseDir, testData.files, mock.Anything).Return(webhook.ValidationResult{Success: true}, nil).Once()
	mocks.mutator.On("Mutate", mock.Anything, testData.tmpBaseDir, testData.files, mock.Anything).Return(fmt.Errorf("surprise!")).Once()
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

	asset := &v1alpha1.Asset{}
	err = c.Get(context.TODO(), testData.key, asset)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	g.Expect(asset.Status.Phase).To(gomega.Equal(v1alpha1.AssetFailed))
	g.Expect(asset.Status.Reason).To(gomega.Equal(string(ReasonMutationFailed)))
	g.Expect(asset.Finalizers).NotTo(gomega.ContainElement(deleteAssetFinalizerName))
}

func TestReconcileAssetCreationSuccessValidationFailed(t *testing.T) {
	//Given
	name := "the-tesciak"
	testData := newTestData(name)
	testData.asset.Status = v1alpha1.AssetStatus{LastHeartbeatTime: v1.Now()}

	mocks := newMocks()
	mocks.bucketLister.On("Get", testData.namespace, testData.bucketName).Return(testData.bucket, nil).Times(3)
	mocks.loader.On("Load", testData.asset.Spec.Source.Url, testData.assetName, testData.asset.Spec.Source.Mode, testData.asset.Spec.Source.Filter).Return(testData.tmpBaseDir, testData.files, nil).Once()
	mocks.loader.On("Clean", testData.tmpBaseDir).Return(nil).Once()
	mocks.validator.On("Validate", mock.Anything, testData.tmpBaseDir, testData.files, mock.Anything).Return(webhook.ValidationResult{Success: false}, nil).Once()
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

	asset := &v1alpha1.Asset{}
	err = c.Get(context.TODO(), testData.key, asset)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	g.Expect(asset.Status.Phase).To(gomega.Equal(v1alpha1.AssetFailed))
	g.Expect(asset.Status.Reason).To(gomega.Equal(string(ReasonValidationFailed)))
	g.Expect(asset.Finalizers).NotTo(gomega.ContainElement(deleteAssetFinalizerName))
}

func TestReconcileAssetCreationSuccessNoWebhooks(t *testing.T) {
	//Given
	name := "the-tesciak"
	testData := newTestData(name)
	testData.asset.Status = v1alpha1.AssetStatus{LastHeartbeatTime: v1.Now()}
	testData.asset.Spec.Source.MutationWebhookService = nil
	testData.asset.Spec.Source.ValidationWebhookService = nil

	mocks := newMocks()
	mocks.bucketLister.On("Get", testData.namespace, testData.bucketName).Return(testData.bucket, nil).Times(4)
	mocks.loader.On("Load", testData.asset.Spec.Source.Url, testData.assetName, testData.asset.Spec.Source.Mode, testData.asset.Spec.Source.Filter).Return(testData.tmpBaseDir, testData.files, nil).Once()
	mocks.loader.On("Clean", testData.tmpBaseDir).Return(nil).Once()
	mocks.uploader.On("Upload", mock.Anything, testData.minioBucketName, testData.assetName, testData.tmpBaseDir, testData.files).Return(nil).Once()
	mocks.uploader.On("ContainsAll", testData.minioBucketName, testData.assetName, testData.files).Return(true, nil).Once()
	mocks.cleaner.On("Clean", mock.Anything, testData.minioBucketName, fmt.Sprintf("%s/", testData.assetName)).Return(nil).Once()
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
	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(testData.request)))

	asset := &v1alpha1.Asset{}
	err = c.Get(context.TODO(), testData.key, asset)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	g.Expect(asset.Status.Phase).To(gomega.Equal(v1alpha1.AssetReady))
	g.Expect(asset.Status.AssetRef.BaseUrl).To(gomega.Equal(testData.assetUrl))
	g.Expect(asset.Status.AssetRef.Assets).To(gomega.Equal(testData.files))
	g.Expect(asset.Finalizers).To(gomega.ContainElement(deleteAssetFinalizerName))
}

func TestReconcileAssetCreationNoBucket(t *testing.T) {
	//Given
	name := "the-tesciak"
	testData := newTestData(name)
	testData.asset.Status = v1alpha1.AssetStatus{LastHeartbeatTime: v1.Now()}

	mocks := newMocks()
	mocks.bucketLister.On("Get", testData.namespace, testData.bucketName).Return(nil, nil).Times(3)
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

	asset := &v1alpha1.Asset{}
	err = c.Get(context.TODO(), testData.key, asset)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	g.Expect(asset.Status.Phase).To(gomega.Equal(v1alpha1.AssetPending))
	g.Expect(asset.Status.Reason).To(gomega.Equal(string(ReasonBucketNotReady)))
	g.Expect(asset.Finalizers).NotTo(gomega.ContainElement(deleteAssetFinalizerName))
}

func TestReconcileAssetReadyLostBucket(t *testing.T) {
	//Given
	name := "the-tesciak"
	testData := newTestData(name)

	mocks := newMocks()
	mocks.bucketLister.On("Get", testData.namespace, testData.bucketName).Return(nil, nil).Twice()
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

	asset := &v1alpha1.Asset{}
	err = c.Get(context.TODO(), testData.key, asset)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	g.Expect(asset.Status.Phase).To(gomega.Equal(v1alpha1.AssetPending))
	g.Expect(asset.Status.Reason).To(gomega.Equal(string(ReasonBucketNotReady)))
	g.Expect(asset.Finalizers).NotTo(gomega.ContainElement(deleteAssetFinalizerName))
}

func TestReconcileAssetReadyLostFiles(t *testing.T) {
	//Given
	name := "the-tesciak"
	testData := newTestData(name)
	testData.asset.Spec.Source.MutationWebhookService = nil
	testData.asset.Spec.Source.ValidationWebhookService = nil

	mocks := newMocks()
	mocks.bucketLister.On("Get", testData.namespace, testData.bucketName).Return(testData.bucket, nil).Times(4)
	mocks.loader.On("Load", testData.asset.Spec.Source.Url, testData.assetName, testData.asset.Spec.Source.Mode, testData.asset.Spec.Source.Filter).Return(testData.tmpBaseDir, testData.files, nil).Once()
	mocks.loader.On("Clean", testData.tmpBaseDir).Return(nil).Once()
	mocks.uploader.On("Upload", mock.Anything, testData.minioBucketName, testData.assetName, testData.tmpBaseDir, testData.files).Return(nil).Once()
	mocks.uploader.On("ContainsAll", testData.minioBucketName, testData.assetName, testData.files).Return(false, nil).Once()
	mocks.uploader.On("ContainsAll", testData.minioBucketName, testData.assetName, testData.files).Return(true, nil).Once()
	mocks.cleaner.On("Clean", mock.Anything, testData.minioBucketName, fmt.Sprintf("%s/", testData.assetName)).Return(nil).Once()

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

	asset := &v1alpha1.Asset{}
	err = c.Get(context.TODO(), testData.key, asset)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	g.Expect(asset.Status.Phase).To(gomega.Equal(v1alpha1.AssetPending))
	g.Expect(asset.Status.Reason).To(gomega.Equal(string(ReasonMissingFiles)))
	g.Expect(asset.Finalizers).NotTo(gomega.ContainElement(deleteAssetFinalizerName))

	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(testData.request)))

	asset = &v1alpha1.Asset{}
	err = c.Get(context.TODO(), testData.key, asset)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	g.Expect(asset.Status.Phase).To(gomega.Equal(v1alpha1.AssetReady))
	g.Expect(asset.Finalizers).To(gomega.ContainElement(deleteAssetFinalizerName))
}

type mocks struct {
	uploader     *automock.Uploader
	loader       *automock5.Loader
	cleaner      *automock2.Cleaner
	bucketLister *automock3.Lister
	validator    *automock4.Validator
	mutator      *automock4.Mutator
}

func newMocks() *mocks {
	return &mocks{
		uploader:     new(automock.Uploader),
		loader:       new(automock5.Loader),
		cleaner:      new(automock2.Cleaner),
		bucketLister: new(automock3.Lister),
		validator:    new(automock4.Validator),
		mutator:      new(automock4.Mutator),
	}
}

func (m *mocks) AssertExpetactions(t *testing.T) {
	m.validator.AssertExpectations(t)
	m.bucketLister.AssertExpectations(t)
	m.cleaner.AssertExpectations(t)
	m.uploader.AssertExpectations(t)
	m.loader.AssertExpectations(t)
	m.mutator.AssertExpectations(t)
}

type testData struct {
	namespace       string
	assetName       string
	assetUrl        string
	bucketName      string
	minioBucketName string
	bucketUrl       string
	key             types.NamespacedName

	tmpBaseDir string
	files      []string

	bucket *v1alpha1.Bucket
	asset  *v1alpha1.Asset

	request reconcile.Request
}

func newTestData(name string) *testData {
	namespace := fmt.Sprintf("%s-space", name)
	assetName := fmt.Sprintf("%s-asset", name)
	bucketName := fmt.Sprintf("%s-bucket", name)
	minioBucketName := store.BucketName(namespace, bucketName)
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

	bucket := &v1alpha1.Bucket{
		ObjectMeta: v1.ObjectMeta{
			Name:      bucketName,
			Namespace: namespace,
		},
		Status: v1alpha1.BucketStatus{
			Phase: v1alpha1.BucketReady,
			Url:   bucketUrl,
		},
	}

	asset := &v1alpha1.Asset{
		ObjectMeta: v1.ObjectMeta{
			Name:      assetName,
			Namespace: namespace,
		},
		Spec: v1alpha1.AssetSpec{
			BucketRef: v1alpha1.AssetBucketRef{Name: bucketName},
			Source: v1alpha1.AssetSource{
				Url:    sourceUrl,
				Mode:   v1alpha1.AssetSingle,
				Filter: "",
				ValidationWebhookService: []v1alpha1.AssetWebhookService{
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
				MutationWebhookService: []v1alpha1.AssetWebhookService{
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
		},
		Status: v1alpha1.AssetStatus{
			Phase: v1alpha1.AssetReady,
			AssetRef: v1alpha1.AssetStatusRef{
				BaseUrl: assetUrl,
				Assets:  files,
			},
			LastHeartbeatTime: v1.Now(),
		},
	}

	request := reconcile.Request{NamespacedName: types.NamespacedName{Name: assetName, Namespace: namespace}}

	return &testData{
		namespace:       namespace,
		assetName:       assetName,
		assetUrl:        assetUrl,
		bucketName:      bucketName,
		minioBucketName: minioBucketName,
		bucketUrl:       bucketUrl,
		tmpBaseDir:      tmpBaseDir,
		files:           files,
		key:             key,
		bucket:          bucket,
		asset:           asset,
		request:         request,
	}
}

func deleteAndExpectSuccess(cfg *testSuite, testData *testData) {
	g := cfg.g
	c := cfg.c
	err := c.Delete(context.TODO(), testData.asset)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Eventually(cfg.requests, timeout).Should(gomega.Receive(gomega.Equal(testData.request)))
	g.Eventually(func() bool {
		instance := &v1alpha1.Asset{}
		err := c.Get(context.TODO(), testData.key, instance)
		return apierrors.IsNotFound(err)
	}, timeout, 10*time.Millisecond).Should(gomega.BeTrue())
}
