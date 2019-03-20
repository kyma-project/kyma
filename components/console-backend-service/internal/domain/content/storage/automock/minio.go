// Code generated by mockery v1.0.0. DO NOT EDIT.

package automock

import minio "github.com/minio/minio-go"
import mock "github.com/stretchr/testify/mock"

// Minio is an autogenerated mock type for the Minio type
type Minio struct {
	mock.Mock
}

// GetObject provides a mock function with given fields: bucketName, objectName, opts
func (_m *Minio) GetObject(bucketName string, objectName string, opts minio.GetObjectOptions) (*minio.Object, error) {
	ret := _m.Called(bucketName, objectName, opts)

	var r0 *minio.Object
	if rf, ok := ret.Get(0).(func(string, string, minio.GetObjectOptions) *minio.Object); ok {
		r0 = rf(bucketName, objectName, opts)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*minio.Object)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, minio.GetObjectOptions) error); ok {
		r1 = rf(bucketName, objectName, opts)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListenBucketNotification provides a mock function with given fields: bucketName, prefix, suffix, events, doneCh
func (_m *Minio) ListenBucketNotification(bucketName string, prefix string, suffix string, events []string, doneCh <-chan struct{}) <-chan minio.NotificationInfo {
	ret := _m.Called(bucketName, prefix, suffix, events, doneCh)

	var r0 <-chan minio.NotificationInfo
	if rf, ok := ret.Get(0).(func(string, string, string, []string, <-chan struct{}) <-chan minio.NotificationInfo); ok {
		r0 = rf(bucketName, prefix, suffix, events, doneCh)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(<-chan minio.NotificationInfo)
		}
	}

	return r0
}
