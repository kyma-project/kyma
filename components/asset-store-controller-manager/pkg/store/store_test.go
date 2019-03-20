package store_test

import (
	"context"
	"fmt"
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/apis/assetstore/v1alpha2"
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/store"
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/store/automock"
	"github.com/minio/minio-go"
	"github.com/onsi/gomega"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"path/filepath"
	"testing"
)

func TestStore_BucketExists(t *testing.T) {
	t.Run("Exists", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		name := "test-bucket"

		minio := new(automock.MinioClient)
		minio.On("BucketExists", name).Return(true, nil).Once()
		defer minio.AssertExpectations(t)

		store := store.New(minio)

		// When
		exists, err := store.BucketExists(name)

		// Then
		g.Expect(err).NotTo(gomega.HaveOccurred())
		g.Expect(exists).To(gomega.Equal(true))
	})

	t.Run("NotExists", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		name := "test-bucket"

		minio := new(automock.MinioClient)
		minio.On("BucketExists", name).Return(false, nil).Once()
		defer minio.AssertExpectations(t)

		store := store.New(minio)

		// When
		exists, err := store.BucketExists(name)

		// Then
		g.Expect(err).NotTo(gomega.HaveOccurred())
		g.Expect(exists).To(gomega.Equal(false))
	})

	t.Run("Error", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		name := "test-bucket"

		minio := new(automock.MinioClient)
		minio.On("BucketExists", name).Return(false, fmt.Errorf("test error")).Once()
		defer minio.AssertExpectations(t)

		store := store.New(minio)

		// When
		exists, err := store.BucketExists(name)

		// Then
		g.Expect(err).To(gomega.HaveOccurred())
		g.Expect(exists).To(gomega.Equal(false))
	})
}

func TestStore_CompareBucketPolicy(t *testing.T) {
	t.Run("SuccessNone", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		bucketName := "test-bucket"
		expectedPolicy := v1alpha2.BucketPolicyNone
		remotePolicy := "{\"Version\":\"2012-10-17\",\"Statement\":[]}"

		minio := new(automock.MinioClient)
		minio.On("GetBucketPolicy", bucketName).Return(remotePolicy, nil).Once()
		defer minio.AssertExpectations(t)

		store := store.New(minio)

		// When
		equal, err := store.CompareBucketPolicy(bucketName, expectedPolicy)

		// Then
		g.Expect(err).NotTo(gomega.HaveOccurred())
		g.Expect(equal).To(gomega.Equal(true))
	})

	t.Run("SuccessReadOnly", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		bucketName := "test-bucket"
		expectedPolicy := v1alpha2.BucketPolicyReadOnly
		remotePolicy := "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Action\":[\"s3:GetBucketLocation\",\"s3:ListBucket\"],\"Effect\":\"Allow\",\"Principal\":{\"AWS\":[\"*\"]},\"Resource\":[\"arn:aws:s3:::test-bucket\"],\"Sid\":\"\"},{\"Action\":[\"s3:GetObject\"],\"Effect\":\"Allow\",\"Principal\":{\"AWS\":[\"*\"]},\"Resource\":[\"arn:aws:s3:::test-bucket/*\"],\"Sid\":\"\"}]}"

		minio := new(automock.MinioClient)
		minio.On("GetBucketPolicy", bucketName).Return(remotePolicy, nil).Once()
		defer minio.AssertExpectations(t)

		store := store.New(minio)

		// When
		equal, err := store.CompareBucketPolicy(bucketName, expectedPolicy)

		// Then
		g.Expect(err).NotTo(gomega.HaveOccurred())
		g.Expect(equal).To(gomega.Equal(true))
	})

	t.Run("SuccessWriteOnly", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		bucketName := "test-bucket"
		expectedPolicy := v1alpha2.BucketPolicyWriteOnly
		remotePolicy := "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Action\":[\"s3:GetBucketLocation\",\"s3:ListBucketMultipartUploads\"],\"Effect\":\"Allow\",\"Principal\":{\"AWS\":[\"*\"]},\"Resource\":[\"arn:aws:s3:::test-bucket\"],\"Sid\":\"\"},{\"Action\":[\"s3:AbortMultipartUpload\",\"s3:DeleteObject\",\"s3:ListMultipartUploadParts\",\"s3:PutObject\"],\"Effect\":\"Allow\",\"Principal\":{\"AWS\":[\"*\"]},\"Resource\":[\"arn:aws:s3:::test-bucket/*\"],\"Sid\":\"\"}]}"

		minio := new(automock.MinioClient)
		minio.On("GetBucketPolicy", bucketName).Return(remotePolicy, nil).Once()
		defer minio.AssertExpectations(t)

		store := store.New(minio)

		// When
		equal, err := store.CompareBucketPolicy(bucketName, expectedPolicy)

		// Then
		g.Expect(err).NotTo(gomega.HaveOccurred())
		g.Expect(equal).To(gomega.Equal(true))
	})

	t.Run("SuccessReadWrite", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		bucketName := "test-bucket"
		expectedPolicy := v1alpha2.BucketPolicyReadWrite
		remotePolicy := "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Action\":[\"s3:GetBucketLocation\",\"s3:ListBucket\",\"s3:ListBucketMultipartUploads\"],\"Effect\":\"Allow\",\"Principal\":{\"AWS\":[\"*\"]},\"Resource\":[\"arn:aws:s3:::test-bucket\"],\"Sid\":\"\"},{\"Action\":[\"s3:AbortMultipartUpload\",\"s3:DeleteObject\",\"s3:GetObject\",\"s3:ListMultipartUploadParts\",\"s3:PutObject\"],\"Effect\":\"Allow\",\"Principal\":{\"AWS\":[\"*\"]},\"Resource\":[\"arn:aws:s3:::test-bucket/*\"],\"Sid\":\"\"}]}"

		minio := new(automock.MinioClient)
		minio.On("GetBucketPolicy", bucketName).Return(remotePolicy, nil).Once()
		defer minio.AssertExpectations(t)

		store := store.New(minio)

		// When
		equal, err := store.CompareBucketPolicy(bucketName, expectedPolicy)

		// Then
		g.Expect(err).NotTo(gomega.HaveOccurred())
		g.Expect(equal).To(gomega.Equal(true))
	})

	t.Run("EmptyRemotePolicy", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		bucketName := "test-bucket"
		expectedPolicy := v1alpha2.BucketPolicyNone

		minio := new(automock.MinioClient)
		minio.On("GetBucketPolicy", bucketName).Return("", nil).Once()
		defer minio.AssertExpectations(t)

		store := store.New(minio)

		// When
		equal, err := store.CompareBucketPolicy(bucketName, expectedPolicy)

		// Then
		g.Expect(err).NotTo(gomega.HaveOccurred())
		g.Expect(equal).To(gomega.Equal(false))
	})

	t.Run("Error", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		bucketName := "test-bucket"
		expectedPolicy := v1alpha2.BucketPolicyNone

		minio := new(automock.MinioClient)
		minio.On("GetBucketPolicy", bucketName).Return("", errors.New("test-error")).Once()
		defer minio.AssertExpectations(t)

		store := store.New(minio)

		// When
		_, err := store.CompareBucketPolicy(bucketName, expectedPolicy)

		// Then
		g.Expect(err).To(gomega.HaveOccurred())
	})
}

func TestStore_ContainsAllObjects(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		bucketName := "test-bucket"
		assetName := "test-asset"
		files := []string{"test/a.txt", "test/b/c/d.txt"}
		objCh := fixObjectsChannel(minio.ObjectInfo{Key: "/test-asset/test/a.txt"}, minio.ObjectInfo{Key: "/test-asset/test/b/c/d.txt"})
		ctx := context.TODO()

		minio := new(automock.MinioClient)
		minio.On("ListObjects", bucketName, fmt.Sprintf("/%s", assetName), true, ctx.Done()).Return(objCh).Once()
		defer minio.AssertExpectations(t)

		store := store.New(minio)

		// When
		contains, err := store.ContainsAllObjects(ctx, bucketName, assetName, files)

		// Then
		g.Expect(err).NotTo(gomega.HaveOccurred())
		g.Expect(contains).To(gomega.Equal(true))
	})

	t.Run("MissingFiles", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		bucketName := "test-bucket"
		assetName := "test-asset"
		files := []string{"test/a.txt", "test/b/c/d.txt"}
		objCh := fixObjectsChannel(minio.ObjectInfo{Key: "/test-asset/test/a.txt"})
		ctx := context.TODO()

		minio := new(automock.MinioClient)
		minio.On("ListObjects", bucketName, fmt.Sprintf("/%s", assetName), true, ctx.Done()).Return(objCh).Once()
		defer minio.AssertExpectations(t)

		store := store.New(minio)

		// When
		contains, err := store.ContainsAllObjects(ctx, bucketName, assetName, files)

		// Then
		g.Expect(err).NotTo(gomega.HaveOccurred())
		g.Expect(contains).To(gomega.Equal(false))
	})

	t.Run("EmptyBucket", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		bucketName := "test-bucket"
		assetName := "test-asset"
		files := []string{"test/a.txt", "test/b/c/d.txt"}
		objCh := fixObjectsChannel()
		ctx := context.TODO()

		minio := new(automock.MinioClient)
		minio.On("ListObjects", bucketName, fmt.Sprintf("/%s", assetName), true, ctx.Done()).Return(objCh).Once()
		defer minio.AssertExpectations(t)

		store := store.New(minio)

		// When
		contains, err := store.ContainsAllObjects(ctx, bucketName, assetName, files)

		// Then
		g.Expect(err).NotTo(gomega.HaveOccurred())
		g.Expect(contains).To(gomega.Equal(false))
	})

	t.Run("EmptyAsset", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		bucketName := "test-bucket"
		assetName := "test-asset"
		files := make([]string, 0)
		objCh := fixObjectsChannel(minio.ObjectInfo{Key: "/test-asset/test/a.txt"})
		ctx := context.TODO()

		minio := new(automock.MinioClient)
		minio.On("ListObjects", bucketName, fmt.Sprintf("/%s", assetName), true, ctx.Done()).Return(objCh).Once()
		defer minio.AssertExpectations(t)

		store := store.New(minio)

		// When
		contains, err := store.ContainsAllObjects(ctx, bucketName, assetName, files)

		// Then
		g.Expect(err).NotTo(gomega.HaveOccurred())
		g.Expect(contains).To(gomega.Equal(true))
	})

	t.Run("ListObjectsError", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		bucketName := "test-bucket"
		assetName := "test-asset"
		files := make([]string, 0)
		objCh := fixObjectsChannel(minio.ObjectInfo{Key: "/test-asset/test/a.txt", Err: errors.New("test-err")})
		ctx := context.TODO()

		minio := new(automock.MinioClient)
		minio.On("ListObjects", bucketName, fmt.Sprintf("/%s", assetName), true, ctx.Done()).Return(objCh).Once()
		defer minio.AssertExpectations(t)

		store := store.New(minio)

		// When
		_, err := store.ContainsAllObjects(ctx, bucketName, assetName, files)

		// Then
		g.Expect(err).To(gomega.HaveOccurred())
	})
}

func TestStore_CreateBucket(t *testing.T) {
	t.Run("SuccessNamespaced", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		namespace := "space"
		crName := "test-bucket"
		region := "asia"

		minio := new(automock.MinioClient)
		minio.On("BucketExists", mock.AnythingOfType("string")).Return(false, nil).Once()
		minio.On("MakeBucket", mock.AnythingOfType("string"), region).Return(nil).Once()
		defer minio.AssertExpectations(t)

		store := store.New(minio)

		// When
		name, err := store.CreateBucket(namespace, crName, region)

		// Then
		g.Expect(err).NotTo(gomega.HaveOccurred())
		g.Expect(name).To(gomega.HavePrefix(crName))
		g.Expect(name).ToNot(gomega.HaveSuffix(crName))
	})

	t.Run("SuccessClusterWide", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		crName := "test-bucket"
		region := "asia"

		minio := new(automock.MinioClient)
		minio.On("BucketExists", mock.AnythingOfType("string")).Return(false, nil).Once()
		minio.On("MakeBucket", mock.AnythingOfType("string"), region).Return(nil).Once()
		defer minio.AssertExpectations(t)

		store := store.New(minio)

		// When
		name, err := store.CreateBucket("", crName, region)

		// Then
		g.Expect(err).NotTo(gomega.HaveOccurred())
		g.Expect(name).To(gomega.HavePrefix(crName))
		g.Expect(name).ToNot(gomega.HaveSuffix(crName))
	})

	t.Run("BucketExists", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		namespace := "space"
		crName := "test-bucket"
		region := "asia"

		minio := new(automock.MinioClient)
		minio.On("BucketExists", mock.AnythingOfType("string")).Return(true, nil).Once()
		minio.On("BucketExists", mock.AnythingOfType("string")).Return(false, nil).Once()
		minio.On("MakeBucket", mock.AnythingOfType("string"), region).Return(nil).Once()
		defer minio.AssertExpectations(t)

		store := store.New(minio)

		// When
		name, err := store.CreateBucket(namespace, crName, region)

		// Then
		g.Expect(err).NotTo(gomega.HaveOccurred())
		g.Expect(name).To(gomega.HavePrefix(crName))
		g.Expect(name).ToNot(gomega.HaveSuffix(crName))
	})

	t.Run("CannotFindBucketName", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		namespace := "space"
		crName := "test-bucket"
		region := "asia"

		minio := new(automock.MinioClient)
		minio.On("BucketExists", mock.AnythingOfType("string")).Return(true, nil).Times(10)
		defer minio.AssertExpectations(t)

		store := store.New(minio)

		// When
		_, err := store.CreateBucket(namespace, crName, region)

		// Then
		g.Expect(err).To(gomega.HaveOccurred())
	})

	t.Run("BucketExistsError", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		namespace := "space"
		crName := "test-bucket"
		region := "asia"

		minio := new(automock.MinioClient)
		minio.On("BucketExists", mock.AnythingOfType("string")).Return(false, errors.New("test-err")).Once()
		defer minio.AssertExpectations(t)

		store := store.New(minio)

		// When
		_, err := store.CreateBucket(namespace, crName, region)

		// Then
		g.Expect(err).To(gomega.HaveOccurred())
	})

	t.Run("MakeBucketError", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		namespace := "space"
		crName := "test-bucket"
		region := "asia"

		minio := new(automock.MinioClient)
		minio.On("BucketExists", mock.AnythingOfType("string")).Return(false, nil).Once()
		minio.On("MakeBucket", mock.AnythingOfType("string"), region).Return(errors.New("test-error")).Once()

		defer minio.AssertExpectations(t)

		store := store.New(minio)

		// When
		_, err := store.CreateBucket(namespace, crName, region)

		// Then
		g.Expect(err).To(gomega.HaveOccurred())
	})
}

func TestStore_DeleteBucket(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		name := "test-bucket"
		ctx := context.TODO()
		objCh := fixObjectsChannel(minio.ObjectInfo{Key: "obj1"}, minio.ObjectInfo{Key: "obj2"}, minio.ObjectInfo{Key: "obj3"})
		errCh := fixRemoveObjectErrorChannel()

		minio := new(automock.MinioClient)
		minio.On("BucketExists", name).Return(true, nil).Once()
		minio.On("ListObjects", name, "", true, ctx.Done()).Return(objCh).Once()
		minio.On("RemoveObjectsWithContext", ctx, name, mock.Anything).Return(errCh).Once()
		minio.On("RemoveBucket", name).Return(nil).Once()
		defer minio.AssertExpectations(t)

		store := store.New(minio)

		// When
		err := store.DeleteBucket(ctx, name)

		// Then
		g.Expect(err).NotTo(gomega.HaveOccurred())
	})

	t.Run("EmptyBucket", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		name := "test-bucket"
		ctx := context.TODO()
		objCh := fixObjectsChannel()

		minio := new(automock.MinioClient)
		minio.On("BucketExists", name).Return(true, nil).Once()
		minio.On("ListObjects", name, "", true, ctx.Done()).Return(objCh).Once()
		minio.On("RemoveBucket", name).Return(nil).Once()
		defer minio.AssertExpectations(t)

		store := store.New(minio)

		// When
		err := store.DeleteBucket(ctx, name)

		// Then
		g.Expect(err).NotTo(gomega.HaveOccurred())
	})

	t.Run("BucketNotExists", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		name := "test-bucket"
		ctx := context.TODO()

		minio := new(automock.MinioClient)
		minio.On("BucketExists", name).Return(false, nil).Once()
		defer minio.AssertExpectations(t)

		store := store.New(minio)

		// When
		err := store.DeleteBucket(ctx, name)

		// Then
		g.Expect(err).NotTo(gomega.HaveOccurred())
	})

	t.Run("BucketExistsError", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		name := "test-bucket"
		ctx := context.TODO()

		minio := new(automock.MinioClient)
		minio.On("BucketExists", name).Return(false, fmt.Errorf("test error")).Once()
		defer minio.AssertExpectations(t)

		store := store.New(minio)

		// When
		err := store.DeleteBucket(ctx, name)

		// Then
		g.Expect(err).To(gomega.HaveOccurred())
	})

	t.Run("ListObjectsError", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		name := "test-bucket"
		ctx := context.TODO()
		objCh := fixObjectsChannel(minio.ObjectInfo{Key: "obj1"}, minio.ObjectInfo{Key: "obj2", Err: fmt.Errorf("test error")})

		minio := new(automock.MinioClient)
		minio.On("BucketExists", name).Return(true, nil).Once()
		minio.On("ListObjects", name, "", true, ctx.Done()).Return(objCh).Once()
		defer minio.AssertExpectations(t)

		store := store.New(minio)

		// When
		err := store.DeleteBucket(ctx, name)

		// Then
		g.Expect(err).To(gomega.HaveOccurred())
	})

	t.Run("RemoveBucketError", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		name := "test-bucket"
		ctx := context.TODO()
		objCh := fixObjectsChannel(minio.ObjectInfo{Key: "obj1"}, minio.ObjectInfo{Key: "obj2"}, minio.ObjectInfo{Key: "obj3"})
		errCh := fixRemoveObjectErrorChannel()

		minio := new(automock.MinioClient)
		minio.On("BucketExists", name).Return(true, nil).Once()
		minio.On("ListObjects", name, "", true, ctx.Done()).Return(objCh).Once()
		minio.On("RemoveObjectsWithContext", ctx, name, mock.Anything).Return(errCh).Once()
		minio.On("RemoveBucket", name).Return(errors.New("test-error")).Once()
		defer minio.AssertExpectations(t)

		store := store.New(minio)

		// When
		err := store.DeleteBucket(ctx, name)

		// Then
		g.Expect(err).To(gomega.HaveOccurred())
	})
}

func TestStore_DeleteObjects(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		name := "test-bucket"
		ctx := context.TODO()
		objCh := fixObjectsChannel(minio.ObjectInfo{Key: "obj1"}, minio.ObjectInfo{Key: "obj2"}, minio.ObjectInfo{Key: "obj3"})
		errCh := fixRemoveObjectErrorChannel()

		minio := new(automock.MinioClient)
		minio.On("ListObjects", name, "", true, ctx.Done()).Return(objCh).Once()
		minio.On("RemoveObjectsWithContext", ctx, name, mock.Anything).Return(errCh).Once()
		defer minio.AssertExpectations(t)

		store := store.New(minio)

		// When
		err := store.DeleteObjects(ctx, name, "")

		// Then
		g.Expect(err).NotTo(gomega.HaveOccurred())
	})

	t.Run("SuccessWithPrefix", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		name := "test-bucket"
		prefix := "/test"
		ctx := context.TODO()
		objCh := fixObjectsChannel(minio.ObjectInfo{Key: "/test/obj1"}, minio.ObjectInfo{Key: "/test/obj2"}, minio.ObjectInfo{Key: "/test/obj3"})
		errCh := fixRemoveObjectErrorChannel()

		minio := new(automock.MinioClient)
		minio.On("ListObjects", name, prefix, true, ctx.Done()).Return(objCh).Once()
		minio.On("RemoveObjectsWithContext", ctx, name, mock.Anything).Return(errCh).Once()
		defer minio.AssertExpectations(t)

		store := store.New(minio)

		// When
		err := store.DeleteObjects(ctx, name, prefix)

		// Then
		g.Expect(err).NotTo(gomega.HaveOccurred())
	})

	t.Run("NoObjects", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		name := "test-bucket"
		ctx := context.TODO()
		objCh := fixObjectsChannel()

		minio := new(automock.MinioClient)
		minio.On("ListObjects", name, "", true, ctx.Done()).Return(objCh).Once()
		defer minio.AssertExpectations(t)

		store := store.New(minio)

		// When
		err := store.DeleteObjects(ctx, name, "")

		// Then
		g.Expect(err).NotTo(gomega.HaveOccurred())
	})

	t.Run("ListObjectsError", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		name := "test-bucket"
		ctx := context.TODO()
		objCh := fixObjectsChannel(minio.ObjectInfo{Key: "obj1"}, minio.ObjectInfo{Key: "obj2", Err: fmt.Errorf("test error")})

		minio := new(automock.MinioClient)
		minio.On("ListObjects", name, "", true, ctx.Done()).Return(objCh).Once()
		defer minio.AssertExpectations(t)

		store := store.New(minio)

		// When
		err := store.DeleteObjects(ctx, name, "")

		// Then
		g.Expect(err).To(gomega.HaveOccurred())
	})

	t.Run("RemoveObjectsError", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		name := "test-bucket"
		ctx := context.TODO()
		objCh := fixObjectsChannel(minio.ObjectInfo{Key: "obj1"}, minio.ObjectInfo{Key: "obj2"}, minio.ObjectInfo{Key: "obj3"})
		errCh := fixRemoveObjectErrorChannel(errors.New("test-error"))

		minio := new(automock.MinioClient)
		minio.On("ListObjects", name, "", true, ctx.Done()).Return(objCh).Once()
		minio.On("RemoveObjectsWithContext", ctx, name, mock.Anything).Return(errCh).Once()
		defer minio.AssertExpectations(t)

		store := store.New(minio)

		// When
		err := store.DeleteObjects(ctx, name, "")

		// Then
		g.Expect(err).To(gomega.HaveOccurred())
	})
}

func TestStore_PutObjects(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		bucketName := "test-bucket"
		assetName := "test-asset"
		sourceBasePath := "/tmp"
		files := []string{"test/a.txt", "test/b.txt"}
		ctx := context.TODO()

		minio := new(automock.MinioClient)
		minio.On("FPutObjectWithContext", ctx, bucketName, filepath.Join(assetName, files[0]), filepath.Join(sourceBasePath, files[0]), mock.Anything).Return(int64(1), nil).Once()
		minio.On("FPutObjectWithContext", ctx, bucketName, filepath.Join(assetName, files[1]), filepath.Join(sourceBasePath, files[1]), mock.Anything).Return(int64(1), nil).Once()
		defer minio.AssertExpectations(t)

		store := store.New(minio)

		// When
		err := store.PutObjects(ctx, bucketName, assetName, sourceBasePath, files)

		// Then
		g.Expect(err).NotTo(gomega.HaveOccurred())
	})

	t.Run("Error", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		bucketName := "test-bucket"
		assetName := "test-asset"
		sourceBasePath := "/tmp"
		files := []string{"test/a.txt", "test/b.txt"}
		ctx := context.TODO()

		minio := new(automock.MinioClient)
		minio.On("FPutObjectWithContext", ctx, bucketName, filepath.Join(assetName, files[0]), filepath.Join(sourceBasePath, files[0]), mock.Anything).Return(int64(1), nil).Once()
		minio.On("FPutObjectWithContext", ctx, bucketName, filepath.Join(assetName, files[1]), filepath.Join(sourceBasePath, files[1]), mock.Anything).Return(int64(1), errors.New("test-error")).Once()
		defer minio.AssertExpectations(t)

		store := store.New(minio)

		// When
		err := store.PutObjects(ctx, bucketName, assetName, sourceBasePath, files)

		// Then
		g.Expect(err).To(gomega.HaveOccurred())
	})

	t.Run("NoFiles", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		bucketName := "test-bucket"
		assetName := "test-asset"
		sourceBasePath := "/tmp"
		files := make([]string, 0)
		ctx := context.TODO()

		minio := new(automock.MinioClient)
		defer minio.AssertExpectations(t)

		store := store.New(minio)

		// When
		err := store.PutObjects(ctx, bucketName, assetName, sourceBasePath, files)

		// Then
		g.Expect(err).ToNot(gomega.HaveOccurred())
	})
}

func TestStore_SetBucketPolicy(t *testing.T) {
	t.Run("SuccessNone", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		bucketName := "test-bucket"
		policy := v1alpha2.BucketPolicyNone
		marshaledPolicy := "{\"Version\":\"2012-10-17\",\"Statement\":[]}"

		minio := new(automock.MinioClient)
		minio.On("SetBucketPolicy", bucketName, marshaledPolicy).Return(nil).Once()
		defer minio.AssertExpectations(t)

		store := store.New(minio)

		// When
		err := store.SetBucketPolicy(bucketName, policy)

		// Then
		g.Expect(err).NotTo(gomega.HaveOccurred())
	})

	t.Run("SuccessReadOnly", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		bucketName := "test-bucket"
		expectedPolicy := v1alpha2.BucketPolicyReadOnly
		marshaledPolicy := "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Action\":[\"s3:GetBucketLocation\",\"s3:ListBucket\"],\"Effect\":\"Allow\",\"Principal\":{\"AWS\":[\"*\"]},\"Resource\":[\"arn:aws:s3:::test-bucket\"],\"Sid\":\"\"},{\"Action\":[\"s3:GetObject\"],\"Effect\":\"Allow\",\"Principal\":{\"AWS\":[\"*\"]},\"Resource\":[\"arn:aws:s3:::test-bucket/*\"],\"Sid\":\"\"}]}"

		minio := new(automock.MinioClient)
		minio.On("SetBucketPolicy", bucketName, marshaledPolicy).Return(nil).Once()
		defer minio.AssertExpectations(t)

		store := store.New(minio)

		// When
		err := store.SetBucketPolicy(bucketName, expectedPolicy)

		// Then
		g.Expect(err).NotTo(gomega.HaveOccurred())
	})

	t.Run("SuccessWriteOnly", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		bucketName := "test-bucket"
		expectedPolicy := v1alpha2.BucketPolicyWriteOnly
		marshaledPolicy := "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Action\":[\"s3:GetBucketLocation\",\"s3:ListBucketMultipartUploads\"],\"Effect\":\"Allow\",\"Principal\":{\"AWS\":[\"*\"]},\"Resource\":[\"arn:aws:s3:::test-bucket\"],\"Sid\":\"\"},{\"Action\":[\"s3:AbortMultipartUpload\",\"s3:DeleteObject\",\"s3:ListMultipartUploadParts\",\"s3:PutObject\"],\"Effect\":\"Allow\",\"Principal\":{\"AWS\":[\"*\"]},\"Resource\":[\"arn:aws:s3:::test-bucket/*\"],\"Sid\":\"\"}]}"

		minio := new(automock.MinioClient)
		minio.On("SetBucketPolicy", bucketName, marshaledPolicy).Return(nil).Once()
		defer minio.AssertExpectations(t)

		store := store.New(minio)

		// When
		err := store.SetBucketPolicy(bucketName, expectedPolicy)

		// Then
		g.Expect(err).NotTo(gomega.HaveOccurred())
	})

	t.Run("SuccessReadWrite", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		bucketName := "test-bucket"
		expectedPolicy := v1alpha2.BucketPolicyReadWrite
		marshaledPolicy := "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Action\":[\"s3:GetBucketLocation\",\"s3:ListBucket\",\"s3:ListBucketMultipartUploads\"],\"Effect\":\"Allow\",\"Principal\":{\"AWS\":[\"*\"]},\"Resource\":[\"arn:aws:s3:::test-bucket\"],\"Sid\":\"\"},{\"Action\":[\"s3:AbortMultipartUpload\",\"s3:DeleteObject\",\"s3:GetObject\",\"s3:ListMultipartUploadParts\",\"s3:PutObject\"],\"Effect\":\"Allow\",\"Principal\":{\"AWS\":[\"*\"]},\"Resource\":[\"arn:aws:s3:::test-bucket/*\"],\"Sid\":\"\"}]}"

		minio := new(automock.MinioClient)
		minio.On("SetBucketPolicy", bucketName, marshaledPolicy).Return(nil).Once()
		defer minio.AssertExpectations(t)

		store := store.New(minio)

		// When
		err := store.SetBucketPolicy(bucketName, expectedPolicy)

		// Then
		g.Expect(err).NotTo(gomega.HaveOccurred())
	})

	t.Run("Error", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		bucketName := "test-bucket"
		expectedPolicy := v1alpha2.BucketPolicyNone
		marshaledPolicy := "{\"Version\":\"2012-10-17\",\"Statement\":[]}"

		minio := new(automock.MinioClient)
		minio.On("SetBucketPolicy", bucketName, marshaledPolicy).Return(errors.New("test-error")).Once()
		defer minio.AssertExpectations(t)

		store := store.New(minio)

		// When
		err := store.SetBucketPolicy(bucketName, expectedPolicy)

		// Then
		g.Expect(err).To(gomega.HaveOccurred())
	})
}

func fixObjectsChannel(objects ...minio.ObjectInfo) <-chan minio.ObjectInfo {
	objCh := make(chan minio.ObjectInfo, len(objects)+1)
	defer close(objCh)
	for _, object := range objects {
		objCh <- object
	}

	return objCh
}

func fixRemoveObjectErrorChannel(errs ...error) <-chan minio.RemoveObjectError {
	objCh := make(chan minio.RemoveObjectError, len(errs)+1)
	defer close(objCh)
	for _, err := range errs {
		info := minio.RemoveObjectError{
			Err: err,
		}

		objCh <- info
	}

	return objCh
}
