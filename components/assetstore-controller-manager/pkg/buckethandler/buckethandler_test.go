package buckethandler_test

import (
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/buckethandler"
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/buckethandler/automock"
	"github.com/onsi/gomega"
	"github.com/pkg/errors"
	"testing"
)

func TestBucketHandler_Create(t *testing.T) {
	t.Run("SuccessRegion", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		minioCli := &automock.MinioClient{}
		handler := buckethandler.New(minioCli, nil)

		bucketName := "bucket"
		region := "region"

		minioCli.On("BucketExists", bucketName).Return(false, nil).Once()
		minioCli.On("MakeBucket", bucketName, region).Return(nil).Once()
		defer minioCli.AssertExpectations(t)

		// When
		created, err := handler.CreateIfDoesntExist(bucketName, region)

		// Then
		g.Expect(created).To(gomega.BeTrue())
		g.Expect(err).NotTo(gomega.HaveOccurred())
	})

	t.Run("Exists", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		minioCli := &automock.MinioClient{}
		handler := buckethandler.New(minioCli, nil)

		bucketName := "bucket"
		region := "region"

		minioCli.On("BucketExists", bucketName).Return(true, nil).Once()
		defer minioCli.AssertExpectations(t)

		// When
		created, err := handler.CreateIfDoesntExist(bucketName, region)

		// Then
		g.Expect(created).To(gomega.BeFalse())
		g.Expect(err).NotTo(gomega.HaveOccurred())
	})

	t.Run("Error", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		minioCli := &automock.MinioClient{}
		handler := buckethandler.New(minioCli, nil)

		bucketName := "bucket"
		region := "region"
		testErr := errors.New("test error")

		minioCli.On("BucketExists", bucketName).Return(false, nil).Once()
		minioCli.On("MakeBucket", bucketName, region).Return(testErr).Once()
		defer minioCli.AssertExpectations(t)

		// When
		_, err := handler.CreateIfDoesntExist(bucketName, region)

		// Then
		g.Expect(err.Error()).To(gomega.ContainSubstring(testErr.Error()))
	})
}

func TestBucketHandler_CheckIfExists(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		minioCli := &automock.MinioClient{}
		handler := buckethandler.New(minioCli, nil)

		bucketName := "bucket"

		minioCli.On("BucketExists", bucketName).Return(true, nil).Once()
		defer minioCli.AssertExpectations(t)

		// When
		exists, err := handler.CheckIfExists(bucketName)

		// Then
		g.Expect(exists).To(gomega.BeTrue())
		g.Expect(err).NotTo(gomega.HaveOccurred())
	})

	t.Run("Error", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		minioCli := &automock.MinioClient{}
		handler := buckethandler.New(minioCli, nil)

		bucketName := "bucket"
		testErr := errors.New("test error")

		minioCli.On("BucketExists", bucketName).Return(false, testErr).Once()
		defer minioCli.AssertExpectations(t)

		// When
		_, err := handler.CheckIfExists(bucketName)

		// Then
		g.Expect(err.Error()).To(gomega.ContainSubstring(testErr.Error()))
	})
}

func TestBucketHandler_Delete(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		minioCli := &automock.MinioClient{}
		handler := buckethandler.New(minioCli, nil)

		bucketName := "bucket"

		minioCli.On("BucketExists", bucketName).Return(true, nil).Once()
		minioCli.On("RemoveBucket", bucketName).Return(nil).Once()
		defer minioCli.AssertExpectations(t)

		// When
		err := handler.Delete(bucketName)

		// Then
		g.Expect(err).NotTo(gomega.HaveOccurred())
	})

	t.Run("NotExists", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		minioCli := &automock.MinioClient{}
		handler := buckethandler.New(minioCli, nil)

		bucketName := "bucket"

		minioCli.On("BucketExists", bucketName).Return(false, nil).Once()
		defer minioCli.AssertExpectations(t)

		// When
		err := handler.Delete(bucketName)

		// Then
		g.Expect(err).NotTo(gomega.HaveOccurred())
	})

	t.Run("Error", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		minioCli := &automock.MinioClient{}
		handler := buckethandler.New(minioCli, nil)

		bucketName := "bucket"
		testErr := errors.New("test error")

		minioCli.On("BucketExists", bucketName).Return(true, nil).Once()
		minioCli.On("RemoveBucket", bucketName).Return(testErr).Once()
		defer minioCli.AssertExpectations(t)

		// When
		err := handler.Delete(bucketName)

		// Then
		g.Expect(err.Error()).To(gomega.ContainSubstring(testErr.Error()))
	})
}

func TestBucketHandler_SetPolicy(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		minioCli := &automock.MinioClient{}
		handler := buckethandler.New(minioCli, nil)

		bucketName := "bucket"
		policy := "readonly"

		minioCli.On("GetBucketPolicy", bucketName).Return("none", nil).Once()
		minioCli.On("SetBucketPolicy", bucketName, policy).Return(nil).Once()
		defer minioCli.AssertExpectations(t)

		// When
		updated, err := handler.SetPolicyIfNotEqual(bucketName, policy)

		// Then
		g.Expect(updated).To(gomega.BeTrue())
		g.Expect(err).NotTo(gomega.HaveOccurred())
	})

	t.Run("AlreadySet", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		minioCli := &automock.MinioClient{}
		handler := buckethandler.New(minioCli, nil)

		bucketName := "bucket"
		policy := "readonly"

		minioCli.On("GetBucketPolicy", bucketName).Return(policy, nil).Once()
		defer minioCli.AssertExpectations(t)

		// When
		updated, err := handler.SetPolicyIfNotEqual(bucketName, policy)

		// Then
		g.Expect(updated).To(gomega.BeFalse())
		g.Expect(err).NotTo(gomega.HaveOccurred())
	})

	t.Run("Error", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		minioCli := &automock.MinioClient{}
		handler := buckethandler.New(minioCli, nil)

		bucketName := "bucket"
		policy := "readonly"
		testErr := errors.New("test error")

		minioCli.On("GetBucketPolicy", bucketName).Return("none", nil).Once()
		minioCli.On("SetBucketPolicy", bucketName, policy).Return(testErr).Once()
		defer minioCli.AssertExpectations(t)

		// When
		_, err := handler.SetPolicyIfNotEqual(bucketName, policy)

		// Then
		g.Expect(err.Error()).To(gomega.ContainSubstring(testErr.Error()))
	})
}

func TestBucketHandler_GetPolicy(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		minioCli := &automock.MinioClient{}
		handler := buckethandler.New(minioCli, nil)

		bucketName := "bucket"
		expectedPolicy := "readonly"

		minioCli.On("GetBucketPolicy", bucketName).Return(expectedPolicy, nil).Once()
		defer minioCli.AssertExpectations(t)

		// When
		policy, err := handler.GetPolicy(bucketName)

		// Then
		g.Expect(policy).To(gomega.Equal(expectedPolicy))
		g.Expect(err).NotTo(gomega.HaveOccurred())
	})

	t.Run("Error", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		minioCli := &automock.MinioClient{}
		handler := buckethandler.New(minioCli, nil)

		bucketName := "bucket"
		testErr := errors.New("test error")

		minioCli.On("GetBucketPolicy", bucketName).Return("", testErr).Once()
		defer minioCli.AssertExpectations(t)

		// When
		_, err := handler.GetPolicy(bucketName)

		// Then
		g.Expect(err.Error()).To(gomega.ContainSubstring(testErr.Error()))
	})
}

func TestBucketHandler_ComparePolicy(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		minioCli := &automock.MinioClient{}
		handler := buckethandler.New(minioCli, nil)

		bucketName := "bucket"
		expectedPolicy := "readonly"

		minioCli.On("GetBucketPolicy", bucketName).Return(expectedPolicy, nil).Once()
		defer minioCli.AssertExpectations(t)

		// When
		policy, err := handler.GetPolicy(bucketName)

		// Then
		g.Expect(policy).To(gomega.Equal(expectedPolicy))
		g.Expect(err).NotTo(gomega.HaveOccurred())
	})

	t.Run("Error", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		minioCli := &automock.MinioClient{}
		handler := buckethandler.New(minioCli, nil)

		bucketName := "bucket"
		testErr := errors.New("test error")

		minioCli.On("GetBucketPolicy", bucketName).Return("", testErr).Once()
		defer minioCli.AssertExpectations(t)

		// When
		_, err := handler.GetPolicy(bucketName)

		// Then
		g.Expect(err.Error()).To(gomega.ContainSubstring(testErr.Error()))
	})
}
