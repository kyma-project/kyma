package store_test

import (
	"github.com/kyma-project/kyma/tests/asset-store/internal/store"
	"github.com/kyma-project/kyma/tests/asset-store/internal/store/automock"
	"testing"
	"github.com/onsi/gomega"
	"github.com/pkg/errors"
)

func TestBucketHandler_CheckIfExists(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		minioCli := &automock.MinioClient{}
		s := store.New(minioCli)

		bucketName := "bucket"

		minioCli.On("BucketExists", bucketName).Return(true, nil).Once()
		defer minioCli.AssertExpectations(t)

		// When
		exists, err := s.Exists(bucketName)

		// Then
		g.Expect(exists).To(gomega.BeTrue())
		g.Expect(err).NotTo(gomega.HaveOccurred())
	})

	t.Run("Error", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		minioCli := &automock.MinioClient{}
		s := store.New(minioCli)

		bucketName := "bucket"
		testErr := errors.New("test error")

		minioCli.On("BucketExists", bucketName).Return(false, testErr).Once()
		defer minioCli.AssertExpectations(t)

		// When
		_, err := s.Exists(bucketName)

		// Then
		g.Expect(err.Error()).To(gomega.ContainSubstring(testErr.Error()))
	})
}

func TestBucketHandler_GetPolicy(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		minioCli := &automock.MinioClient{}
		s := store.New(minioCli)

		bucketName := "bucket"
		expectedPolicy := `{"foo":"bar"}`

		minioCli.On("GetBucketPolicy", bucketName).Return(expectedPolicy, nil).Once()
		defer minioCli.AssertExpectations(t)

		// When
		policy, err := s.GetPolicy(bucketName)

		// Then
		g.Expect(policy).To(gomega.Equal(expectedPolicy))
		g.Expect(err).NotTo(gomega.HaveOccurred())
	})

	t.Run("Error", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		minioCli := &automock.MinioClient{}
		s := store.New(minioCli)

		bucketName := "bucket"
		testErr := errors.New("test error")

		minioCli.On("GetBucketPolicy", bucketName).Return("", testErr).Once()
		defer minioCli.AssertExpectations(t)

		// When
		_, err := s.GetPolicy(bucketName)

		// Then
		g.Expect(err.Error()).To(gomega.ContainSubstring(testErr.Error()))
	})
}