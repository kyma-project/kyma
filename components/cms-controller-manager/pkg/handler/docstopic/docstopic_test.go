package docstopic_test

import (
	"context"
	"fmt"
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/apis/assetstore/v1alpha2"
	"github.com/kyma-project/kyma/components/cms-controller-manager/pkg/apis/cms/v1alpha1"
	"github.com/kyma-project/kyma/components/cms-controller-manager/pkg/handler/docstopic"
	"github.com/kyma-project/kyma/components/cms-controller-manager/pkg/handler/docstopic/automock"
	"github.com/kyma-project/kyma/components/cms-controller-manager/pkg/handler/docstopic/pretty"
	"github.com/onsi/gomega"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"testing"
)

var log = logf.Log.WithName("docstopic-test")

func TestDocstopicHandler_Handle_AddOrUpdate(t *testing.T) {
	t.Run("Create", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		ctx := context.TODO()
		sources := map[string]v1alpha1.Source{"swag": testSource("https://dummy.url", v1alpha1.DocsTopicSingle)}
		testData := testData("halo", sources)

		assetSvc := new(automock.AssetService)
		defer assetSvc.AssertExpectations(t)
		bucketSvc := new(automock.BucketService)
		defer bucketSvc.AssertExpectations(t)

		bucketSvc.On("List", ctx, testData.Namespace, map[string]string{"access.cms.kyma-project.io": "public"}).Return([]string{"test-bucket"}, nil).Once()
		assetSvc.On("List", ctx, testData.Namespace, map[string]string{"docstopic.cms.kyma-project.io": testData.Name}).Return(nil, nil).Once()
		assetSvc.On("Create", ctx, testData, mock.Anything).Return(nil).Once()

		handler := docstopic.New(log, fakeRecorder(), assetSvc, bucketSvc)

		// When
		status, err := handler.Handle(ctx, testData, testData.Spec.CommonDocsTopicSpec, testData.Status.CommonDocsTopicStatus)

		// Then
		g.Expect(err).ToNot(gomega.HaveOccurred())
		g.Expect(status).ToNot(gomega.BeNil())
		g.Expect(status.Phase).To(gomega.Equal(v1alpha1.DocsTopicPending))
		g.Expect(status.Reason).To(gomega.Equal(pretty.WaitingForAssets.String()))
	})

	t.Run("CreateError", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		ctx := context.TODO()
		sources := map[string]v1alpha1.Source{"swag": testSource("https://dummy.url", v1alpha1.DocsTopicSingle)}
		testData := testData("halo", sources)

		assetSvc := new(automock.AssetService)
		defer assetSvc.AssertExpectations(t)
		bucketSvc := new(automock.BucketService)
		defer bucketSvc.AssertExpectations(t)

		bucketSvc.On("List", ctx, testData.Namespace, map[string]string{"access.cms.kyma-project.io": "public"}).Return([]string{"test-bucket"}, nil).Once()
		assetSvc.On("List", ctx, testData.Namespace, map[string]string{"docstopic.cms.kyma-project.io": testData.Name}).Return(nil, nil).Once()
		assetSvc.On("Create", ctx, testData, mock.Anything).Return(errors.New("test-data")).Once()

		handler := docstopic.New(log, fakeRecorder(), assetSvc, bucketSvc)

		// When
		status, err := handler.Handle(ctx, testData, testData.Spec.CommonDocsTopicSpec, testData.Status.CommonDocsTopicStatus)

		// Then
		g.Expect(err).To(gomega.HaveOccurred())
		g.Expect(status).ToNot(gomega.BeNil())
		g.Expect(status.Phase).To(gomega.Equal(v1alpha1.DocsTopicFailed))
		g.Expect(status.Reason).To(gomega.Equal(pretty.AssetsCreationFailed.String()))
	})

	t.Run("Update", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		ctx := context.TODO()
		bucketName := "test-bucket"
		sourceName := "swag"
		sources := map[string]v1alpha1.Source{"swag": testSource("https://dummy.url", v1alpha1.DocsTopicSingle)}
		testData := testData("halo", sources)
		existingAsset := commonAsset(sourceName, testData.Name, bucketName, sources[sourceName], v1alpha2.AssetPending)
		existingAsset.Spec.Source.Filter = "xyz"
		existingAssets := []docstopic.CommonAsset{existingAsset}

		assetSvc := new(automock.AssetService)
		defer assetSvc.AssertExpectations(t)
		bucketSvc := new(automock.BucketService)
		defer bucketSvc.AssertExpectations(t)

		bucketSvc.On("List", ctx, testData.Namespace, map[string]string{"access.cms.kyma-project.io": "public"}).Return([]string{bucketName}, nil).Once()
		assetSvc.On("List", ctx, testData.Namespace, map[string]string{"docstopic.cms.kyma-project.io": testData.Name}).Return(existingAssets, nil).Once()
		assetSvc.On("Update", ctx, mock.Anything).Return(nil).Once()

		handler := docstopic.New(log, fakeRecorder(), assetSvc, bucketSvc)

		// When
		status, err := handler.Handle(ctx, testData, testData.Spec.CommonDocsTopicSpec, testData.Status.CommonDocsTopicStatus)

		// Then
		g.Expect(err).ToNot(gomega.HaveOccurred())
		g.Expect(status).ToNot(gomega.BeNil())
		g.Expect(status.Phase).To(gomega.Equal(v1alpha1.DocsTopicPending))
		g.Expect(status.Reason).To(gomega.Equal(pretty.WaitingForAssets.String()))
	})

	t.Run("UpdateError", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		ctx := context.TODO()
		bucketName := "test-bucket"
		sourceName := "swag"
		sources := map[string]v1alpha1.Source{"swag": testSource("https://dummy.url", v1alpha1.DocsTopicSingle)}
		testData := testData("halo", sources)
		existingAsset := commonAsset(sourceName, testData.Name, bucketName, sources[sourceName], v1alpha2.AssetPending)
		existingAsset.Spec.Source.Filter = "xyz"
		existingAssets := []docstopic.CommonAsset{existingAsset}

		assetSvc := new(automock.AssetService)
		defer assetSvc.AssertExpectations(t)
		bucketSvc := new(automock.BucketService)
		defer bucketSvc.AssertExpectations(t)

		bucketSvc.On("List", ctx, testData.Namespace, map[string]string{"access.cms.kyma-project.io": "public"}).Return([]string{bucketName}, nil).Once()
		assetSvc.On("List", ctx, testData.Namespace, map[string]string{"docstopic.cms.kyma-project.io": testData.Name}).Return(existingAssets, nil).Once()
		assetSvc.On("Update", ctx, mock.Anything).Return(errors.New("test-error")).Once()

		handler := docstopic.New(log, fakeRecorder(), assetSvc, bucketSvc)

		// When
		status, err := handler.Handle(ctx, testData, testData.Spec.CommonDocsTopicSpec, testData.Status.CommonDocsTopicStatus)

		// Then
		g.Expect(err).To(gomega.HaveOccurred())
		g.Expect(status).ToNot(gomega.BeNil())
		g.Expect(status.Phase).To(gomega.Equal(v1alpha1.DocsTopicFailed))
		g.Expect(status.Reason).To(gomega.Equal(pretty.AssetsUpdateFailed.String()))
	})

	t.Run("Delete", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		ctx := context.TODO()
		bucketName := "test-bucket"
		sourceName := "swag"
		sources := map[string]v1alpha1.Source{"swag": testSource("https://dummy.url", v1alpha1.DocsTopicSingle)}
		testData := testData("halo", sources)
		existingAsset := commonAsset(sourceName, testData.Name, bucketName, sources[sourceName], v1alpha2.AssetPending)
		toRemove := commonAsset("papa", testData.Name, bucketName, sources[sourceName], v1alpha2.AssetPending)
		existingAssets := []docstopic.CommonAsset{existingAsset, toRemove}

		assetSvc := new(automock.AssetService)
		defer assetSvc.AssertExpectations(t)
		bucketSvc := new(automock.BucketService)
		defer bucketSvc.AssertExpectations(t)

		bucketSvc.On("List", ctx, testData.Namespace, map[string]string{"access.cms.kyma-project.io": "public"}).Return([]string{bucketName}, nil).Once()
		assetSvc.On("List", ctx, testData.Namespace, map[string]string{"docstopic.cms.kyma-project.io": testData.Name}).Return(existingAssets, nil).Once()
		assetSvc.On("Delete", ctx, toRemove).Return(nil).Once()

		handler := docstopic.New(log, fakeRecorder(), assetSvc, bucketSvc)

		// When
		status, err := handler.Handle(ctx, testData, testData.Spec.CommonDocsTopicSpec, testData.Status.CommonDocsTopicStatus)

		// Then
		g.Expect(err).ToNot(gomega.HaveOccurred())
		g.Expect(status).ToNot(gomega.BeNil())
		g.Expect(status.Phase).To(gomega.Equal(v1alpha1.DocsTopicPending))
		g.Expect(status.Reason).To(gomega.Equal(pretty.WaitingForAssets.String()))
	})

	t.Run("DeleteError", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		ctx := context.TODO()
		bucketName := "test-bucket"
		sourceName := "swag"
		sources := map[string]v1alpha1.Source{"swag": testSource("https://dummy.url", v1alpha1.DocsTopicSingle)}
		testData := testData("halo", sources)
		existingAsset := commonAsset(sourceName, testData.Name, bucketName, sources[sourceName], v1alpha2.AssetPending)
		toRemove := commonAsset("papa", testData.Name, bucketName, sources[sourceName], v1alpha2.AssetPending)
		existingAssets := []docstopic.CommonAsset{existingAsset, toRemove}

		assetSvc := new(automock.AssetService)
		defer assetSvc.AssertExpectations(t)
		bucketSvc := new(automock.BucketService)
		defer bucketSvc.AssertExpectations(t)

		bucketSvc.On("List", ctx, testData.Namespace, map[string]string{"access.cms.kyma-project.io": "public"}).Return([]string{bucketName}, nil).Once()
		assetSvc.On("List", ctx, testData.Namespace, map[string]string{"docstopic.cms.kyma-project.io": testData.Name}).Return(existingAssets, nil).Once()
		assetSvc.On("Delete", ctx, toRemove).Return(errors.New("test-error")).Once()

		handler := docstopic.New(log, fakeRecorder(), assetSvc, bucketSvc)

		// When
		status, err := handler.Handle(ctx, testData, testData.Spec.CommonDocsTopicSpec, testData.Status.CommonDocsTopicStatus)

		// Then
		g.Expect(err).To(gomega.HaveOccurred())
		g.Expect(status).ToNot(gomega.BeNil())
		g.Expect(status.Phase).To(gomega.Equal(v1alpha1.DocsTopicFailed))
		g.Expect(status.Reason).To(gomega.Equal(pretty.AssetsDeletionFailed.String()))
	})

	t.Run("CreateWithBucket", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		ctx := context.TODO()
		sources := map[string]v1alpha1.Source{"swag": testSource("https://dummy.url", v1alpha1.DocsTopicSingle)}
		testData := testData("halo", sources)

		assetSvc := new(automock.AssetService)
		defer assetSvc.AssertExpectations(t)
		bucketSvc := new(automock.BucketService)
		defer bucketSvc.AssertExpectations(t)

		bucketSvc.On("List", ctx, testData.Namespace, map[string]string{"access.cms.kyma-project.io": "public"}).Return(nil, nil).Once()
		bucketSvc.On("Create", ctx, mock.Anything, false, map[string]string{"access.cms.kyma-project.io": "public"}).Return(nil).Once()
		assetSvc.On("List", ctx, testData.Namespace, map[string]string{"docstopic.cms.kyma-project.io": testData.Name}).Return(nil, nil).Once()
		assetSvc.On("Create", ctx, testData, mock.Anything).Return(nil).Once()

		handler := docstopic.New(log, fakeRecorder(), assetSvc, bucketSvc)

		// When
		status, err := handler.Handle(ctx, testData, testData.Spec.CommonDocsTopicSpec, testData.Status.CommonDocsTopicStatus)

		// Then
		g.Expect(err).ToNot(gomega.HaveOccurred())
		g.Expect(status).ToNot(gomega.BeNil())
		g.Expect(status.Phase).To(gomega.Equal(v1alpha1.DocsTopicPending))
		g.Expect(status.Reason).To(gomega.Equal(pretty.WaitingForAssets.String()))
	})

	t.Run("BucketCreationError", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		ctx := context.TODO()
		sources := map[string]v1alpha1.Source{"swag": testSource("https://dummy.url", v1alpha1.DocsTopicSingle)}
		testData := testData("halo", sources)

		assetSvc := new(automock.AssetService)
		defer assetSvc.AssertExpectations(t)
		bucketSvc := new(automock.BucketService)
		defer bucketSvc.AssertExpectations(t)

		bucketSvc.On("List", ctx, testData.Namespace, map[string]string{"access.cms.kyma-project.io": "public"}).Return(nil, nil).Once()
		bucketSvc.On("Create", ctx, mock.Anything, false, map[string]string{"access.cms.kyma-project.io": "public"}).Return(errors.New("test-error")).Once()

		handler := docstopic.New(log, fakeRecorder(), assetSvc, bucketSvc)

		// When
		status, err := handler.Handle(ctx, testData, testData.Spec.CommonDocsTopicSpec, testData.Status.CommonDocsTopicStatus)

		// Then
		g.Expect(err).To(gomega.HaveOccurred())
		g.Expect(status).ToNot(gomega.BeNil())
		g.Expect(status.Phase).To(gomega.Equal(v1alpha1.DocsTopicFailed))
		g.Expect(status.Reason).To(gomega.Equal(pretty.BucketError.String()))
	})

	t.Run("BucketListingError", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		ctx := context.TODO()
		sources := map[string]v1alpha1.Source{"swag": testSource("https://dummy.url", v1alpha1.DocsTopicSingle)}
		testData := testData("halo", sources)

		assetSvc := new(automock.AssetService)
		defer assetSvc.AssertExpectations(t)
		bucketSvc := new(automock.BucketService)
		defer bucketSvc.AssertExpectations(t)

		bucketSvc.On("List", ctx, testData.Namespace, map[string]string{"access.cms.kyma-project.io": "public"}).Return(nil, errors.New("test-error")).Once()

		handler := docstopic.New(log, fakeRecorder(), assetSvc, bucketSvc)

		// When
		status, err := handler.Handle(ctx, testData, testData.Spec.CommonDocsTopicSpec, testData.Status.CommonDocsTopicStatus)

		// Then
		g.Expect(err).To(gomega.HaveOccurred())
		g.Expect(status).ToNot(gomega.BeNil())
		g.Expect(status.Phase).To(gomega.Equal(v1alpha1.DocsTopicFailed))
		g.Expect(status.Reason).To(gomega.Equal(pretty.BucketError.String()))
	})

	t.Run("AssetsListingError", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		ctx := context.TODO()
		sources := map[string]v1alpha1.Source{"swag": testSource("https://dummy.url", v1alpha1.DocsTopicSingle)}
		testData := testData("halo", sources)

		assetSvc := new(automock.AssetService)
		defer assetSvc.AssertExpectations(t)
		bucketSvc := new(automock.BucketService)
		defer bucketSvc.AssertExpectations(t)

		bucketSvc.On("List", ctx, testData.Namespace, map[string]string{"access.cms.kyma-project.io": "public"}).Return([]string{"test-bucket"}, nil).Once()
		assetSvc.On("List", ctx, testData.Namespace, map[string]string{"docstopic.cms.kyma-project.io": testData.Name}).Return(nil, errors.New("test-error")).Once()

		handler := docstopic.New(log, fakeRecorder(), assetSvc, bucketSvc)

		// When
		status, err := handler.Handle(ctx, testData, testData.Spec.CommonDocsTopicSpec, testData.Status.CommonDocsTopicStatus)

		// Then
		g.Expect(err).To(gomega.HaveOccurred())
		g.Expect(status).ToNot(gomega.BeNil())
		g.Expect(status.Phase).To(gomega.Equal(v1alpha1.DocsTopicFailed))
		g.Expect(status.Reason).To(gomega.Equal(pretty.AssetsListingFailed.String()))
	})
}

func TestDocstopicHandler_Handle_Status(t *testing.T) {
	t.Run("NotChanged", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		ctx := context.TODO()
		bucketName := "test-bucket"
		sourceName := "swag"
		sources := map[string]v1alpha1.Source{"swag": testSource("https://dummy.url", v1alpha1.DocsTopicSingle)}
		testData := testData("halo", sources)
		testData.Status.Phase = v1alpha1.DocsTopicPending
		existingAsset := commonAsset(sourceName, testData.Name, bucketName, sources[sourceName], v1alpha2.AssetPending)
		existingAssets := []docstopic.CommonAsset{existingAsset}

		assetSvc := new(automock.AssetService)
		defer assetSvc.AssertExpectations(t)
		bucketSvc := new(automock.BucketService)
		defer bucketSvc.AssertExpectations(t)

		bucketSvc.On("List", ctx, testData.Namespace, map[string]string{"access.cms.kyma-project.io": "public"}).Return([]string{bucketName}, nil).Once()
		assetSvc.On("List", ctx, testData.Namespace, map[string]string{"docstopic.cms.kyma-project.io": testData.Name}).Return(existingAssets, nil).Once()

		handler := docstopic.New(log, fakeRecorder(), assetSvc, bucketSvc)

		// When
		status, err := handler.Handle(ctx, testData, testData.Spec.CommonDocsTopicSpec, testData.Status.CommonDocsTopicStatus)

		// Then
		g.Expect(err).ToNot(gomega.HaveOccurred())
		g.Expect(status).To(gomega.BeNil())
	})

	t.Run("Changed", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		ctx := context.TODO()
		bucketName := "test-bucket"
		sourceName := "swag"
		sources := map[string]v1alpha1.Source{"swag": testSource("https://dummy.url", v1alpha1.DocsTopicSingle)}
		testData := testData("halo", sources)
		testData.Status.Phase = v1alpha1.DocsTopicPending
		existingAsset := commonAsset(sourceName, testData.Name, bucketName, sources[sourceName], v1alpha2.AssetReady)
		existingAssets := []docstopic.CommonAsset{existingAsset}

		assetSvc := new(automock.AssetService)
		defer assetSvc.AssertExpectations(t)
		bucketSvc := new(automock.BucketService)
		defer bucketSvc.AssertExpectations(t)

		bucketSvc.On("List", ctx, testData.Namespace, map[string]string{"access.cms.kyma-project.io": "public"}).Return([]string{bucketName}, nil).Once()
		assetSvc.On("List", ctx, testData.Namespace, map[string]string{"docstopic.cms.kyma-project.io": testData.Name}).Return(existingAssets, nil).Once()

		handler := docstopic.New(log, fakeRecorder(), assetSvc, bucketSvc)

		// When
		status, err := handler.Handle(ctx, testData, testData.Spec.CommonDocsTopicSpec, testData.Status.CommonDocsTopicStatus)

		// Then
		g.Expect(err).ToNot(gomega.HaveOccurred())
		g.Expect(status).ToNot(gomega.BeNil())
		g.Expect(status.Phase).To(gomega.Equal(v1alpha1.DocsTopicReady))
		g.Expect(status.Reason).To(gomega.Equal(pretty.AssetsReady.String()))
	})

	t.Run("AssetError", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		ctx := context.TODO()
		bucketName := "test-bucket"
		sourceName := "swag"
		sources := map[string]v1alpha1.Source{"swag": testSource("https://dummy.url", v1alpha1.DocsTopicSingle)}
		testData := testData("halo", sources)
		testData.Status.Phase = v1alpha1.DocsTopicReady
		existingAsset := commonAsset(sourceName, testData.Name, bucketName, sources[sourceName], v1alpha2.AssetFailed)
		existingAssets := []docstopic.CommonAsset{existingAsset}

		assetSvc := new(automock.AssetService)
		defer assetSvc.AssertExpectations(t)
		bucketSvc := new(automock.BucketService)
		defer bucketSvc.AssertExpectations(t)

		bucketSvc.On("List", ctx, testData.Namespace, map[string]string{"access.cms.kyma-project.io": "public"}).Return([]string{bucketName}, nil).Once()
		assetSvc.On("List", ctx, testData.Namespace, map[string]string{"docstopic.cms.kyma-project.io": testData.Name}).Return(existingAssets, nil).Once()

		handler := docstopic.New(log, fakeRecorder(), assetSvc, bucketSvc)

		// When
		status, err := handler.Handle(ctx, testData, testData.Spec.CommonDocsTopicSpec, testData.Status.CommonDocsTopicStatus)

		// Then
		g.Expect(err).ToNot(gomega.HaveOccurred())
		g.Expect(status).ToNot(gomega.BeNil())
		g.Expect(status.Phase).To(gomega.Equal(v1alpha1.DocsTopicPending))
		g.Expect(status.Reason).To(gomega.Equal(pretty.WaitingForAssets.String()))
	})
}

func fakeRecorder() record.EventRecorder {
	return record.NewFakeRecorder(20)
}

func testSource(url string, mode v1alpha1.DocsTopicMode) v1alpha1.Source {
	return v1alpha1.Source{
		URL:  url,
		Mode: mode,
	}
}

func testData(name string, sources map[string]v1alpha1.Source) *v1alpha1.DocsTopic {
	return &v1alpha1.DocsTopic{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: "test",
		},
		Spec: v1alpha1.DocsTopicSpec{
			CommonDocsTopicSpec: v1alpha1.CommonDocsTopicSpec{
				DisplayName: fmt.Sprintf("%s Display", name),
				Description: fmt.Sprintf("%s Description", name),
				Sources:     sources,
			},
		},
	}
}

func commonAsset(name, docsName, bucketName string, source v1alpha1.Source, phase v1alpha2.AssetPhase) docstopic.CommonAsset {
	return docstopic.CommonAsset{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: "test",
			Labels: map[string]string{
				"docstopic.cms.kyma-project.io": docsName,
				"type.cms.kyma-project.io":      name,
			},
		},
		Spec: v1alpha2.CommonAssetSpec{
			Source: v1alpha2.AssetSource{
				Url:    source.URL,
				Mode:   v1alpha2.AssetMode(source.Mode),
				Filter: source.Filter,
			},
			BucketRef: v1alpha2.AssetBucketRef{
				Name: bucketName,
			},
		},
		Status: v1alpha2.CommonAssetStatus{
			Phase: phase,
		},
	}
}
