// Code generated by mockery v1.0.0. DO NOT EDIT.

package automock

import context "context"
import minio "github.com/minio/minio-go"
import mock "github.com/stretchr/testify/mock"

// MinioClient is an autogenerated mock type for the MinioClient type
type MinioClient struct {
	mock.Mock
}

// FPutObjectWithContext provides a mock function with given fields: ctx, bucketName, objectName, filePath, opts
func (_m *MinioClient) FPutObjectWithContext(ctx context.Context, bucketName string, objectName string, filePath string, opts minio.PutObjectOptions) (int64, error) {
	ret := _m.Called(ctx, bucketName, objectName, filePath, opts)

	var r0 int64
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string, minio.PutObjectOptions) int64); ok {
		r0 = rf(ctx, bucketName, objectName, filePath, opts)
	} else {
		r0 = ret.Get(0).(int64)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string, string, minio.PutObjectOptions) error); ok {
		r1 = rf(ctx, bucketName, objectName, filePath, opts)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListObjects provides a mock function with given fields: bucketName, objectPrefix, recursive, doneCh
func (_m *MinioClient) ListObjects(bucketName string, objectPrefix string, recursive bool, doneCh <-chan struct{}) <-chan minio.ObjectInfo {
	ret := _m.Called(bucketName, objectPrefix, recursive, doneCh)

	var r0 <-chan minio.ObjectInfo
	if rf, ok := ret.Get(0).(func(string, string, bool, <-chan struct{}) <-chan minio.ObjectInfo); ok {
		r0 = rf(bucketName, objectPrefix, recursive, doneCh)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(<-chan minio.ObjectInfo)
		}
	}

	return r0
}
