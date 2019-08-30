// Code generated by mockery v1.0.0. DO NOT EDIT.

package automock

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	v1alpha2 "github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/apis/assetstore/v1alpha2"
)

// Store is an autogenerated mock type for the Store type
type Store struct {
	mock.Mock
}

// BucketExists provides a mock function with given fields: name
func (_m *Store) BucketExists(name string) (bool, error) {
	ret := _m.Called(name)

	var r0 bool
	if rf, ok := ret.Get(0).(func(string) bool); ok {
		r0 = rf(name)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CompareBucketPolicy provides a mock function with given fields: name, expected
func (_m *Store) CompareBucketPolicy(name string, expected v1alpha2.BucketPolicy) (bool, error) {
	ret := _m.Called(name, expected)

	var r0 bool
	if rf, ok := ret.Get(0).(func(string, v1alpha2.BucketPolicy) bool); ok {
		r0 = rf(name, expected)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, v1alpha2.BucketPolicy) error); ok {
		r1 = rf(name, expected)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ContainsAllObjects provides a mock function with given fields: ctx, bucketName, assetName, files
func (_m *Store) ContainsAllObjects(ctx context.Context, bucketName string, assetName string, files []string) (bool, error) {
	ret := _m.Called(ctx, bucketName, assetName, files)

	var r0 bool
	if rf, ok := ret.Get(0).(func(context.Context, string, string, []string) bool); ok {
		r0 = rf(ctx, bucketName, assetName, files)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string, []string) error); ok {
		r1 = rf(ctx, bucketName, assetName, files)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CreateBucket provides a mock function with given fields: namespace, crName, region
func (_m *Store) CreateBucket(namespace string, crName string, region string) (string, error) {
	ret := _m.Called(namespace, crName, region)

	var r0 string
	if rf, ok := ret.Get(0).(func(string, string, string) string); ok {
		r0 = rf(namespace, crName, region)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, string) error); ok {
		r1 = rf(namespace, crName, region)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DeleteBucket provides a mock function with given fields: ctx, name
func (_m *Store) DeleteBucket(ctx context.Context, name string) error {
	ret := _m.Called(ctx, name)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, name)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteObjects provides a mock function with given fields: ctx, bucketName, prefix
func (_m *Store) DeleteObjects(ctx context.Context, bucketName string, prefix string) error {
	ret := _m.Called(ctx, bucketName, prefix)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, bucketName, prefix)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ListObjects provides a mock function with given fields: ctx, bucketName, prefix
func (_m *Store) ListObjects(ctx context.Context, bucketName string, prefix string) ([]string, error) {
	ret := _m.Called(ctx, bucketName, prefix)

	var r0 []string
	if rf, ok := ret.Get(0).(func(context.Context, string, string) []string); ok {
		r0 = rf(ctx, bucketName, prefix)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, bucketName, prefix)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// PutObjects provides a mock function with given fields: ctx, bucketName, assetName, sourceBasePath, files
func (_m *Store) PutObjects(ctx context.Context, bucketName string, assetName string, sourceBasePath string, files []string) error {
	ret := _m.Called(ctx, bucketName, assetName, sourceBasePath, files)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string, []string) error); ok {
		r0 = rf(ctx, bucketName, assetName, sourceBasePath, files)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetBucketPolicy provides a mock function with given fields: name, policy
func (_m *Store) SetBucketPolicy(name string, policy v1alpha2.BucketPolicy) error {
	ret := _m.Called(name, policy)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, v1alpha2.BucketPolicy) error); ok {
		r0 = rf(name, policy)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
