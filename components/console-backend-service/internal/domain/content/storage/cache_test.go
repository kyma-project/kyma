package storage_test

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"testing"
	"time"

	"github.com/allegro/bigcache"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/content/storage"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/content/storage/automock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const synchronizationTimeout = 1 * time.Second

func TestCache_Initialize(t *testing.T) {
	t.Run("Initialize twice", func(t *testing.T) {
		storeGetter := storage.NewStoreGetter()
		cacheClient := new(automock.Cache)

		cache := storage.NewCache(storeGetter, cacheClient)
		stop := make(chan struct{})
		defer close(stop)
		notifications := storage.NewNotificationChan()
		defer close(notifications)

		cacheClient.On("Reset").Return(nil).Once()
		storeGetter.On("NotificationChannel", mock.Anything).Return(storage.GetDirectNotificationChan(notifications)).Once()
		defer cacheClient.AssertExpectations(t)
		defer storeGetter.AssertExpectations(t)

		cache.Initialize(stop)
		cache.Initialize(stop)

		err := waitAtMost(cache.IsSynced, synchronizationTimeout)
		require.NoError(t, err)
	})

	t.Run("Initialize once", func(t *testing.T) {
		storeGetter := storage.NewStoreGetter()
		cacheClient := new(automock.Cache)

		cache := storage.NewCache(storeGetter, cacheClient)
		stop := make(chan struct{})
		defer close(stop)
		notifications := storage.NewNotificationChan()
		defer close(notifications)

		cacheClient.On("Reset").Return(nil).Once()
		storeGetter.On("NotificationChannel", mock.Anything).Return(storage.GetDirectNotificationChan(notifications)).Once()
		defer cacheClient.AssertExpectations(t)
		defer storeGetter.AssertExpectations(t)

		cache.Initialize(stop)

		err := waitAtMost(cache.IsSynced, synchronizationTimeout)
		require.NoError(t, err)
	})
}

func TestCache_ApiSpec_Initialized(t *testing.T) {
	filename := "apiSpec.json"
	fieldName := "apiSpec"
	function := "ApiSpec"

	id := "some-object"
	cacheId := fmt.Sprintf("%s/%s/%s", id, filename, fieldName)

	expected := new(storage.ApiSpec)
	expectedBytes, err := convertToCache(&expected)
	if err != nil {
		t.Error(err)
	}

	t.Run("Not existing object", func(t *testing.T) {
		storeGetter := storage.NewStoreGetter()
		cacheClient := new(automock.Cache)

		cache := storage.NewCache(storeGetter, cacheClient)
		stop := make(chan struct{})
		defer close(stop)
		notifications := storage.NewNotificationChan()
		defer close(notifications)

		storeGetter.On("NotificationChannel", mock.Anything).Return(storage.GetDirectNotificationChan(notifications)).Once()
		storeGetter.On(function, id).Return(nil, false, nil).Once()
		cacheClient.On("Reset").Return(nil).Once()
		cacheClient.On("Get", cacheId).Return(nil, &bigcache.EntryNotFoundError{}).Twice()
		cacheClient.On("Delete", cacheId).Return(nil).Once()
		defer cacheClient.AssertExpectations(t)
		defer storeGetter.AssertExpectations(t)

		cache.Initialize(stop)
		err := waitAtMost(cache.IsSynced, synchronizationTimeout)
		require.NoError(t, err)

		_, exists, err := cache.ApiSpec(id)

		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("Existing not cached object", func(t *testing.T) {
		storeGetter := storage.NewStoreGetter()
		cacheClient := new(automock.Cache)

		cache := storage.NewCache(storeGetter, cacheClient)
		stop := make(chan struct{})
		defer close(stop)
		notifications := storage.NewNotificationChan()
		defer close(notifications)

		storeGetter.On("NotificationChannel", mock.Anything).Return(storage.GetDirectNotificationChan(notifications)).Once()
		storeGetter.On(function, id).Return(expected, true, nil).Once()
		cacheClient.On("Reset").Return(nil).Once()
		cacheClient.On("Get", cacheId).Return(nil, &bigcache.EntryNotFoundError{}).Once()
		cacheClient.On("Set", cacheId, expectedBytes).Return(nil).Once()
		cacheClient.On("Get", cacheId).Return(expectedBytes, nil).Once()
		defer cacheClient.AssertExpectations(t)
		defer storeGetter.AssertExpectations(t)

		cache.Initialize(stop)
		err := waitAtMost(cache.IsSynced, synchronizationTimeout)
		require.NoError(t, err)

		apiSpec, exists, err := cache.ApiSpec(id)

		require.NoError(t, err)
		assert.True(t, exists)
		assert.Equal(t, expected, apiSpec)
	})

	t.Run("Existing cached object", func(t *testing.T) {
		storeGetter := storage.NewStoreGetter()
		cacheClient := new(automock.Cache)

		cache := storage.NewCache(storeGetter, cacheClient)
		stop := make(chan struct{})
		defer close(stop)
		notifications := storage.NewNotificationChan()
		defer close(notifications)

		storeGetter.On("NotificationChannel", mock.Anything).Return(storage.GetDirectNotificationChan(notifications)).Once()
		cacheClient.On("Reset").Return(nil).Once()
		cacheClient.On("Get", cacheId).Return(expectedBytes, nil).Once()
		defer cacheClient.AssertExpectations(t)
		defer storeGetter.AssertExpectations(t)

		cache.Initialize(stop)
		err := waitAtMost(cache.IsSynced, synchronizationTimeout)
		require.NoError(t, err)

		apiSpec, exists, err := cache.ApiSpec(id)

		require.NoError(t, err)
		assert.True(t, exists)
		assert.Equal(t, expected, apiSpec)
	})

	t.Run("store error", func(t *testing.T) {
		storeGetter := storage.NewStoreGetter()
		cacheClient := new(automock.Cache)

		cache := storage.NewCache(storeGetter, cacheClient)
		stop := make(chan struct{})
		defer close(stop)
		notifications := storage.NewNotificationChan()
		defer close(notifications)

		storeGetter.On("NotificationChannel", mock.Anything).Return(storage.GetDirectNotificationChan(notifications)).Once()
		storeGetter.On(function, id).Return(nil, false, errors.New(id))
		cacheClient.On("Reset").Return(nil).Once()
		cacheClient.On("Get", cacheId).Return(nil, &bigcache.EntryNotFoundError{}).Once()
		defer cacheClient.AssertExpectations(t)
		defer storeGetter.AssertExpectations(t)

		cache.Initialize(stop)
		err := waitAtMost(cache.IsSynced, synchronizationTimeout)
		require.NoError(t, err)

		_, exists, err := cache.ApiSpec(id)

		require.Error(t, err)
		assert.False(t, exists)
	})

	t.Run("Cache error", func(t *testing.T) {
		storeGetter := storage.NewStoreGetter()
		cacheClient := new(automock.Cache)

		cache := storage.NewCache(storeGetter, cacheClient)
		stop := make(chan struct{})
		defer close(stop)
		notifications := storage.NewNotificationChan()
		defer close(notifications)

		storeGetter.On("NotificationChannel", mock.Anything).Return(storage.GetDirectNotificationChan(notifications)).Once()
		cacheClient.On("Reset").Return(nil).Once()
		cacheClient.On("Get", cacheId).Return(nil, errors.New(id)).Once()
		defer cacheClient.AssertExpectations(t)
		defer storeGetter.AssertExpectations(t)

		cache.Initialize(stop)
		err := waitAtMost(cache.IsSynced, synchronizationTimeout)
		require.NoError(t, err)

		_, exists, err := cache.ApiSpec(id)

		require.Error(t, err)
		assert.False(t, exists)
	})

	t.Run("Cache error after update", func(t *testing.T) {
		storeGetter := storage.NewStoreGetter()
		cacheClient := new(automock.Cache)

		cache := storage.NewCache(storeGetter, cacheClient)
		stop := make(chan struct{})
		defer close(stop)
		notifications := storage.NewNotificationChan()
		defer close(notifications)

		storeGetter.On("NotificationChannel", mock.Anything).Return(storage.GetDirectNotificationChan(notifications)).Once()
		storeGetter.On(function, id).Return(expected, true, nil).Once()
		cacheClient.On("Reset").Return(nil).Once()
		cacheClient.On("Get", cacheId).Return(nil, &bigcache.EntryNotFoundError{}).Once()
		cacheClient.On("Set", cacheId, expectedBytes).Return(nil).Once()
		cacheClient.On("Get", cacheId).Return(nil, errors.New(id)).Once()
		defer cacheClient.AssertExpectations(t)
		defer storeGetter.AssertExpectations(t)

		cache.Initialize(stop)
		err := waitAtMost(cache.IsSynced, synchronizationTimeout)
		require.NoError(t, err)

		_, exists, err := cache.ApiSpec(id)

		require.Error(t, err)
		assert.False(t, exists)
	})

	t.Run("Error while storing in cache", func(t *testing.T) {
		storeGetter := storage.NewStoreGetter()
		cacheClient := new(automock.Cache)

		cache := storage.NewCache(storeGetter, cacheClient)
		stop := make(chan struct{})
		defer close(stop)
		notifications := storage.NewNotificationChan()
		defer close(notifications)

		storeGetter.On("NotificationChannel", mock.Anything).Return(storage.GetDirectNotificationChan(notifications)).Once()
		storeGetter.On(function, id).Return(expected, true, nil).Once()
		cacheClient.On("Reset").Return(nil).Once()
		cacheClient.On("Get", cacheId).Return(nil, &bigcache.EntryNotFoundError{}).Once()
		cacheClient.On("Set", cacheId, expectedBytes).Return(errors.New(id)).Once()
		defer cacheClient.AssertExpectations(t)
		defer storeGetter.AssertExpectations(t)

		cache.Initialize(stop)
		err := waitAtMost(cache.IsSynced, synchronizationTimeout)
		require.NoError(t, err)

		_, exists, err := cache.ApiSpec(id)

		require.Error(t, err)
		assert.False(t, exists)
	})
}

func TestCache_ApiSpec_NotInitialized(t *testing.T) {
	filename := "apiSpec.json"
	fieldName := "apiSpec"
	function := "ApiSpec"

	id := "some-object"
	cacheId := fmt.Sprintf("%s/%s/%s", id, filename, fieldName)

	expected := new(storage.ApiSpec)
	expectedBytes, err := convertToCache(&expected)
	if err != nil {
		t.Error(err)
	}

	t.Run("Not existing object", func(t *testing.T) {
		storeGetter := storage.NewStoreGetter()
		cacheClient := new(automock.Cache)

		cache := storage.NewCache(storeGetter, cacheClient)
		stop := make(chan struct{})
		defer close(stop)
		notifications := storage.NewNotificationChan()
		defer close(notifications)

		storeGetter.On(function, id).Return(nil, false, nil).Once()
		cacheClient.On("Get", cacheId).Return(nil, &bigcache.EntryNotFoundError{}).Twice()
		cacheClient.On("Delete", cacheId).Return(nil).Once()
		defer cacheClient.AssertExpectations(t)
		defer storeGetter.AssertExpectations(t)

		_, exists, err := cache.ApiSpec(id)

		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("Existing not cached object", func(t *testing.T) {
		storeGetter := storage.NewStoreGetter()
		cacheClient := new(automock.Cache)

		cache := storage.NewCache(storeGetter, cacheClient)
		stop := make(chan struct{})
		defer close(stop)
		notifications := storage.NewNotificationChan()
		defer close(notifications)

		storeGetter.On(function, id).Return(expected, true, nil).Once()
		cacheClient.On("Get", cacheId).Return(nil, &bigcache.EntryNotFoundError{}).Once()
		cacheClient.On("Set", cacheId, expectedBytes).Return(nil).Once()
		cacheClient.On("Get", cacheId).Return(expectedBytes, nil).Once()
		defer cacheClient.AssertExpectations(t)
		defer storeGetter.AssertExpectations(t)

		apiSpec, exists, err := cache.ApiSpec(id)

		require.NoError(t, err)
		assert.True(t, exists)
		assert.Equal(t, expected, apiSpec)
	})

	t.Run("Existing cached object", func(t *testing.T) {
		notExpected := storage.ApiSpec{
			Raw: map[string]interface{}{
				"test": nil,
			},
		}
		notExpectedBytes, err := convertToCache(notExpected)
		if err != nil {
			t.Error(err)
		}

		storeGetter := storage.NewStoreGetter()
		cacheClient := new(automock.Cache)

		cache := storage.NewCache(storeGetter, cacheClient)
		stop := make(chan struct{})
		defer close(stop)
		notifications := storage.NewNotificationChan()
		defer close(notifications)

		storeGetter.On(function, id).Return(expected, true, nil).Once()
		cacheClient.On("Get", cacheId).Return(notExpectedBytes, nil).Once()
		cacheClient.On("Set", cacheId, expectedBytes).Return(nil).Once()
		cacheClient.On("Get", cacheId).Return(expectedBytes, nil).Once()
		defer cacheClient.AssertExpectations(t)
		defer storeGetter.AssertExpectations(t)

		apiSpec, exists, err := cache.ApiSpec(id)

		require.NoError(t, err)
		assert.True(t, exists)
		assert.Equal(t, expected, apiSpec)
	})
}

func TestCache_OpenApiSpec_Initialized(t *testing.T) {
	filename := "apiSpec.json"
	fieldName := "openApiSpec"
	function := "OpenApiSpec"

	id := "some-object"
	cacheId := fmt.Sprintf("%s/%s/%s", id, filename, fieldName)

	expected := new(storage.OpenApiSpec)
	expectedBytes, err := convertToCache(&expected)
	if err != nil {
		t.Error(err)
	}

	t.Run("Not existing object", func(t *testing.T) {
		storeGetter := storage.NewStoreGetter()
		cacheClient := new(automock.Cache)

		cache := storage.NewCache(storeGetter, cacheClient)
		stop := make(chan struct{})
		defer close(stop)
		notifications := storage.NewNotificationChan()
		defer close(notifications)

		storeGetter.On("NotificationChannel", mock.Anything).Return(storage.GetDirectNotificationChan(notifications)).Once()
		storeGetter.On(function, id).Return(nil, false, nil).Once()
		cacheClient.On("Reset").Return(nil).Once()
		cacheClient.On("Get", cacheId).Return(nil, &bigcache.EntryNotFoundError{}).Twice()
		cacheClient.On("Delete", cacheId).Return(nil).Once()
		defer cacheClient.AssertExpectations(t)
		defer storeGetter.AssertExpectations(t)

		cache.Initialize(stop)
		err := waitAtMost(cache.IsSynced, synchronizationTimeout)
		require.NoError(t, err)

		_, exists, err := cache.OpenApiSpec(id)

		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("Existing not cached object", func(t *testing.T) {
		storeGetter := storage.NewStoreGetter()
		cacheClient := new(automock.Cache)

		cache := storage.NewCache(storeGetter, cacheClient)
		stop := make(chan struct{})
		defer close(stop)
		notifications := storage.NewNotificationChan()
		defer close(notifications)

		storeGetter.On("NotificationChannel", mock.Anything).Return(storage.GetDirectNotificationChan(notifications)).Once()
		storeGetter.On(function, id).Return(expected, true, nil).Once()
		cacheClient.On("Reset").Return(nil).Once()
		cacheClient.On("Get", cacheId).Return(nil, &bigcache.EntryNotFoundError{}).Once()
		cacheClient.On("Set", cacheId, expectedBytes).Return(nil).Once()
		cacheClient.On("Get", cacheId).Return(expectedBytes, nil).Once()
		defer cacheClient.AssertExpectations(t)
		defer storeGetter.AssertExpectations(t)

		cache.Initialize(stop)
		err := waitAtMost(cache.IsSynced, synchronizationTimeout)
		require.NoError(t, err)

		openApiSpec, exists, err := cache.OpenApiSpec(id)

		require.NoError(t, err)
		assert.True(t, exists)
		assert.Equal(t, expected, openApiSpec)
	})

	t.Run("Existing cached object", func(t *testing.T) {
		storeGetter := storage.NewStoreGetter()
		cacheClient := new(automock.Cache)

		cache := storage.NewCache(storeGetter, cacheClient)
		stop := make(chan struct{})
		defer close(stop)
		notifications := storage.NewNotificationChan()
		defer close(notifications)

		storeGetter.On("NotificationChannel", mock.Anything).Return(storage.GetDirectNotificationChan(notifications)).Once()
		cacheClient.On("Reset").Return(nil).Once()
		cacheClient.On("Get", cacheId).Return(expectedBytes, nil).Once()
		defer cacheClient.AssertExpectations(t)
		defer storeGetter.AssertExpectations(t)

		cache.Initialize(stop)
		err := waitAtMost(cache.IsSynced, synchronizationTimeout)
		require.NoError(t, err)

		openApiSpec, exists, err := cache.OpenApiSpec(id)

		require.NoError(t, err)
		assert.True(t, exists)
		assert.Equal(t, expected, openApiSpec)
	})

	t.Run("store error", func(t *testing.T) {
		storeGetter := storage.NewStoreGetter()
		cacheClient := new(automock.Cache)

		cache := storage.NewCache(storeGetter, cacheClient)
		stop := make(chan struct{})
		defer close(stop)
		notifications := storage.NewNotificationChan()
		defer close(notifications)

		storeGetter.On("NotificationChannel", mock.Anything).Return(storage.GetDirectNotificationChan(notifications)).Once()
		storeGetter.On(function, id).Return(nil, false, errors.New(id))
		cacheClient.On("Reset").Return(nil).Once()
		cacheClient.On("Get", cacheId).Return(nil, &bigcache.EntryNotFoundError{}).Once()
		defer cacheClient.AssertExpectations(t)
		defer storeGetter.AssertExpectations(t)

		cache.Initialize(stop)
		err := waitAtMost(cache.IsSynced, synchronizationTimeout)
		require.NoError(t, err)

		_, exists, err := cache.OpenApiSpec(id)

		require.Error(t, err)
		assert.False(t, exists)
	})

	t.Run("Cache error", func(t *testing.T) {
		storeGetter := storage.NewStoreGetter()
		cacheClient := new(automock.Cache)

		cache := storage.NewCache(storeGetter, cacheClient)
		stop := make(chan struct{})
		defer close(stop)
		notifications := storage.NewNotificationChan()
		defer close(notifications)

		storeGetter.On("NotificationChannel", mock.Anything).Return(storage.GetDirectNotificationChan(notifications)).Once()
		cacheClient.On("Reset").Return(nil).Once()
		cacheClient.On("Get", cacheId).Return(nil, errors.New(id)).Once()
		defer cacheClient.AssertExpectations(t)
		defer storeGetter.AssertExpectations(t)

		cache.Initialize(stop)
		err := waitAtMost(cache.IsSynced, synchronizationTimeout)
		require.NoError(t, err)

		_, exists, err := cache.OpenApiSpec(id)

		require.Error(t, err)
		assert.False(t, exists)
	})

	t.Run("Cache error after update", func(t *testing.T) {
		storeGetter := storage.NewStoreGetter()
		cacheClient := new(automock.Cache)

		cache := storage.NewCache(storeGetter, cacheClient)
		stop := make(chan struct{})
		defer close(stop)
		notifications := storage.NewNotificationChan()
		defer close(notifications)

		storeGetter.On("NotificationChannel", mock.Anything).Return(storage.GetDirectNotificationChan(notifications)).Once()
		storeGetter.On(function, id).Return(expected, true, nil).Once()
		cacheClient.On("Reset").Return(nil).Once()
		cacheClient.On("Get", cacheId).Return(nil, &bigcache.EntryNotFoundError{}).Once()
		cacheClient.On("Set", cacheId, expectedBytes).Return(nil).Once()
		cacheClient.On("Get", cacheId).Return(nil, errors.New(id)).Once()
		defer cacheClient.AssertExpectations(t)
		defer storeGetter.AssertExpectations(t)

		cache.Initialize(stop)
		err := waitAtMost(cache.IsSynced, synchronizationTimeout)
		require.NoError(t, err)

		_, exists, err := cache.OpenApiSpec(id)

		require.Error(t, err)
		assert.False(t, exists)
	})

	t.Run("Error while storing in cache", func(t *testing.T) {
		storeGetter := storage.NewStoreGetter()
		cacheClient := new(automock.Cache)

		cache := storage.NewCache(storeGetter, cacheClient)
		stop := make(chan struct{})
		defer close(stop)
		notifications := storage.NewNotificationChan()
		defer close(notifications)

		storeGetter.On("NotificationChannel", mock.Anything).Return(storage.GetDirectNotificationChan(notifications)).Once()
		storeGetter.On(function, id).Return(expected, true, nil).Once()
		cacheClient.On("Reset").Return(nil).Once()
		cacheClient.On("Get", cacheId).Return(nil, &bigcache.EntryNotFoundError{}).Once()
		cacheClient.On("Set", cacheId, expectedBytes).Return(errors.New(id)).Once()
		defer cacheClient.AssertExpectations(t)
		defer storeGetter.AssertExpectations(t)

		cache.Initialize(stop)
		err := waitAtMost(cache.IsSynced, synchronizationTimeout)
		require.NoError(t, err)

		_, exists, err := cache.OpenApiSpec(id)

		require.Error(t, err)
		assert.False(t, exists)
	})
}

func TestCache_OpenApiSpec_NotInitialized(t *testing.T) {
	filename := "apiSpec.json"
	fieldName := "openApiSpec"
	function := "OpenApiSpec"

	id := "some-object"
	cacheId := fmt.Sprintf("%s/%s/%s", id, filename, fieldName)

	expected := new(storage.OpenApiSpec)
	expectedBytes, err := convertToCache(&expected)
	if err != nil {
		t.Error(err)
	}

	t.Run("Not existing object", func(t *testing.T) {
		storeGetter := storage.NewStoreGetter()
		cacheClient := new(automock.Cache)

		cache := storage.NewCache(storeGetter, cacheClient)
		stop := make(chan struct{})
		defer close(stop)
		notifications := storage.NewNotificationChan()
		defer close(notifications)

		storeGetter.On(function, id).Return(nil, false, nil).Once()
		cacheClient.On("Get", cacheId).Return(nil, &bigcache.EntryNotFoundError{}).Twice()
		cacheClient.On("Delete", cacheId).Return(nil).Once()
		defer cacheClient.AssertExpectations(t)
		defer storeGetter.AssertExpectations(t)

		_, exists, err := cache.OpenApiSpec(id)

		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("Existing not cached object", func(t *testing.T) {
		storeGetter := storage.NewStoreGetter()
		cacheClient := new(automock.Cache)

		cache := storage.NewCache(storeGetter, cacheClient)
		stop := make(chan struct{})
		defer close(stop)
		notifications := storage.NewNotificationChan()
		defer close(notifications)

		storeGetter.On(function, id).Return(expected, true, nil).Once()
		cacheClient.On("Get", cacheId).Return(nil, &bigcache.EntryNotFoundError{}).Once()
		cacheClient.On("Set", cacheId, expectedBytes).Return(nil).Once()
		cacheClient.On("Get", cacheId).Return(expectedBytes, nil).Once()
		defer cacheClient.AssertExpectations(t)
		defer storeGetter.AssertExpectations(t)

		openApiSpec, exists, err := cache.OpenApiSpec(id)

		require.NoError(t, err)
		assert.True(t, exists)
		assert.Equal(t, expected, openApiSpec)
	})

	t.Run("Existing cached object", func(t *testing.T) {
		notExpected := storage.OpenApiSpec{
			Raw: map[string]interface{}{
				"test": nil,
			},
		}
		notExpectedBytes, err := convertToCache(notExpected)
		if err != nil {
			t.Error(err)
		}

		storeGetter := storage.NewStoreGetter()
		cacheClient := new(automock.Cache)

		cache := storage.NewCache(storeGetter, cacheClient)
		stop := make(chan struct{})
		defer close(stop)
		notifications := storage.NewNotificationChan()
		defer close(notifications)

		storeGetter.On(function, id).Return(expected, true, nil).Once()
		cacheClient.On("Get", cacheId).Return(notExpectedBytes, nil).Once()
		cacheClient.On("Set", cacheId, expectedBytes).Return(nil).Once()
		cacheClient.On("Get", cacheId).Return(expectedBytes, nil).Once()
		defer cacheClient.AssertExpectations(t)
		defer storeGetter.AssertExpectations(t)

		openApiSpec, exists, err := cache.OpenApiSpec(id)

		require.NoError(t, err)
		assert.True(t, exists)
		assert.Equal(t, expected, openApiSpec)
	})
}

func TestCache_ODataSpec_Initialized(t *testing.T) {
	filename := "apiSpec.json"
	fieldName := "odataSpec"
	function := "ODataSpec"

	id := "some-object"
	cacheId := fmt.Sprintf("%s/%s/%s", id, filename, fieldName)

	expected := new(storage.ODataSpec)
	expectedBytes, err := convertToCache(&expected)
	if err != nil {
		t.Error(err)
	}

	t.Run("Not existing object", func(t *testing.T) {
		storeGetter := storage.NewStoreGetter()
		cacheClient := new(automock.Cache)

		cache := storage.NewCache(storeGetter, cacheClient)
		stop := make(chan struct{})
		defer close(stop)
		notifications := storage.NewNotificationChan()
		defer close(notifications)

		storeGetter.On("NotificationChannel", mock.Anything).Return(storage.GetDirectNotificationChan(notifications)).Once()
		storeGetter.On(function, id).Return(nil, false, nil).Once()
		cacheClient.On("Reset").Return(nil).Once()
		cacheClient.On("Get", cacheId).Return(nil, &bigcache.EntryNotFoundError{}).Twice()
		cacheClient.On("Delete", cacheId).Return(nil).Once()
		defer cacheClient.AssertExpectations(t)
		defer storeGetter.AssertExpectations(t)

		cache.Initialize(stop)
		err := waitAtMost(cache.IsSynced, synchronizationTimeout)
		require.NoError(t, err)

		_, exists, err := cache.ODataSpec(id)

		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("Existing not cached object", func(t *testing.T) {
		storeGetter := storage.NewStoreGetter()
		cacheClient := new(automock.Cache)

		cache := storage.NewCache(storeGetter, cacheClient)
		stop := make(chan struct{})
		defer close(stop)
		notifications := storage.NewNotificationChan()
		defer close(notifications)

		storeGetter.On("NotificationChannel", mock.Anything).Return(storage.GetDirectNotificationChan(notifications)).Once()
		storeGetter.On(function, id).Return(expected, true, nil).Once()
		cacheClient.On("Reset").Return(nil).Once()
		cacheClient.On("Get", cacheId).Return(nil, &bigcache.EntryNotFoundError{}).Once()
		cacheClient.On("Set", cacheId, expectedBytes).Return(nil).Once()
		cacheClient.On("Get", cacheId).Return(expectedBytes, nil).Once()
		defer cacheClient.AssertExpectations(t)
		defer storeGetter.AssertExpectations(t)

		cache.Initialize(stop)
		err := waitAtMost(cache.IsSynced, synchronizationTimeout)
		require.NoError(t, err)

		odataSpec, exists, err := cache.ODataSpec(id)

		require.NoError(t, err)
		assert.True(t, exists)
		assert.Equal(t, expected, odataSpec)
	})

	t.Run("Existing cached object", func(t *testing.T) {
		storeGetter := storage.NewStoreGetter()
		cacheClient := new(automock.Cache)

		cache := storage.NewCache(storeGetter, cacheClient)
		stop := make(chan struct{})
		defer close(stop)
		notifications := storage.NewNotificationChan()
		defer close(notifications)

		storeGetter.On("NotificationChannel", mock.Anything).Return(storage.GetDirectNotificationChan(notifications)).Once()
		cacheClient.On("Reset").Return(nil).Once()
		cacheClient.On("Get", cacheId).Return(expectedBytes, nil).Once()
		defer cacheClient.AssertExpectations(t)
		defer storeGetter.AssertExpectations(t)

		cache.Initialize(stop)
		err := waitAtMost(cache.IsSynced, synchronizationTimeout)
		require.NoError(t, err)

		odataSpec, exists, err := cache.ODataSpec(id)

		require.NoError(t, err)
		assert.True(t, exists)
		assert.Equal(t, expected, odataSpec)
	})

	t.Run("store error", func(t *testing.T) {
		storeGetter := storage.NewStoreGetter()
		cacheClient := new(automock.Cache)

		cache := storage.NewCache(storeGetter, cacheClient)
		stop := make(chan struct{})
		defer close(stop)
		notifications := storage.NewNotificationChan()
		defer close(notifications)

		storeGetter.On("NotificationChannel", mock.Anything).Return(storage.GetDirectNotificationChan(notifications)).Once()
		storeGetter.On(function, id).Return(nil, false, errors.New(id))
		cacheClient.On("Reset").Return(nil).Once()
		cacheClient.On("Get", cacheId).Return(nil, &bigcache.EntryNotFoundError{}).Once()
		defer cacheClient.AssertExpectations(t)
		defer storeGetter.AssertExpectations(t)

		cache.Initialize(stop)
		err := waitAtMost(cache.IsSynced, synchronizationTimeout)
		require.NoError(t, err)

		_, exists, err := cache.ODataSpec(id)

		require.Error(t, err)
		assert.False(t, exists)
	})

	t.Run("Cache error", func(t *testing.T) {
		storeGetter := storage.NewStoreGetter()
		cacheClient := new(automock.Cache)

		cache := storage.NewCache(storeGetter, cacheClient)
		stop := make(chan struct{})
		defer close(stop)
		notifications := storage.NewNotificationChan()
		defer close(notifications)

		storeGetter.On("NotificationChannel", mock.Anything).Return(storage.GetDirectNotificationChan(notifications)).Once()
		cacheClient.On("Reset").Return(nil).Once()
		cacheClient.On("Get", cacheId).Return(nil, errors.New(id)).Once()
		defer cacheClient.AssertExpectations(t)
		defer storeGetter.AssertExpectations(t)

		cache.Initialize(stop)
		err := waitAtMost(cache.IsSynced, synchronizationTimeout)
		require.NoError(t, err)

		_, exists, err := cache.ODataSpec(id)

		require.Error(t, err)
		assert.False(t, exists)
	})

	t.Run("Cache error after update", func(t *testing.T) {
		storeGetter := storage.NewStoreGetter()
		cacheClient := new(automock.Cache)

		cache := storage.NewCache(storeGetter, cacheClient)
		stop := make(chan struct{})
		defer close(stop)
		notifications := storage.NewNotificationChan()
		defer close(notifications)

		storeGetter.On("NotificationChannel", mock.Anything).Return(storage.GetDirectNotificationChan(notifications)).Once()
		storeGetter.On(function, id).Return(expected, true, nil).Once()
		cacheClient.On("Reset").Return(nil).Once()
		cacheClient.On("Get", cacheId).Return(nil, &bigcache.EntryNotFoundError{}).Once()
		cacheClient.On("Set", cacheId, expectedBytes).Return(nil).Once()
		cacheClient.On("Get", cacheId).Return(nil, errors.New(id)).Once()
		defer cacheClient.AssertExpectations(t)
		defer storeGetter.AssertExpectations(t)

		cache.Initialize(stop)
		err := waitAtMost(cache.IsSynced, synchronizationTimeout)
		require.NoError(t, err)

		_, exists, err := cache.ODataSpec(id)

		require.Error(t, err)
		assert.False(t, exists)
	})

	t.Run("Error while storing in cache", func(t *testing.T) {
		storeGetter := storage.NewStoreGetter()
		cacheClient := new(automock.Cache)

		cache := storage.NewCache(storeGetter, cacheClient)
		stop := make(chan struct{})
		defer close(stop)
		notifications := storage.NewNotificationChan()
		defer close(notifications)

		storeGetter.On("NotificationChannel", mock.Anything).Return(storage.GetDirectNotificationChan(notifications)).Once()
		storeGetter.On(function, id).Return(expected, true, nil).Once()
		cacheClient.On("Reset").Return(nil).Once()
		cacheClient.On("Get", cacheId).Return(nil, &bigcache.EntryNotFoundError{}).Once()
		cacheClient.On("Set", cacheId, expectedBytes).Return(errors.New(id)).Once()
		defer cacheClient.AssertExpectations(t)
		defer storeGetter.AssertExpectations(t)

		cache.Initialize(stop)
		err := waitAtMost(cache.IsSynced, synchronizationTimeout)
		require.NoError(t, err)

		_, exists, err := cache.ODataSpec(id)

		require.Error(t, err)
		assert.False(t, exists)
	})
}

func TestCache_ODataSpec_NotInitialized(t *testing.T) {
	filename := "apiSpec.json"
	fieldName := "odataSpec"
	function := "ODataSpec"

	id := "some-object"
	cacheId := fmt.Sprintf("%s/%s/%s", id, filename, fieldName)

	expected := new(storage.ODataSpec)
	expectedBytes, err := convertToCache(&expected)
	if err != nil {
		t.Error(err)
	}

	t.Run("Not existing object", func(t *testing.T) {
		storeGetter := storage.NewStoreGetter()
		cacheClient := new(automock.Cache)

		cache := storage.NewCache(storeGetter, cacheClient)
		stop := make(chan struct{})
		defer close(stop)
		notifications := storage.NewNotificationChan()
		defer close(notifications)

		storeGetter.On(function, id).Return(nil, false, nil).Once()
		cacheClient.On("Get", cacheId).Return(nil, &bigcache.EntryNotFoundError{}).Twice()
		cacheClient.On("Delete", cacheId).Return(nil).Once()
		defer cacheClient.AssertExpectations(t)
		defer storeGetter.AssertExpectations(t)

		_, exists, err := cache.ODataSpec(id)

		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("Existing not cached object", func(t *testing.T) {
		storeGetter := storage.NewStoreGetter()
		cacheClient := new(automock.Cache)

		cache := storage.NewCache(storeGetter, cacheClient)
		stop := make(chan struct{})
		defer close(stop)
		notifications := storage.NewNotificationChan()
		defer close(notifications)

		storeGetter.On(function, id).Return(expected, true, nil).Once()
		cacheClient.On("Get", cacheId).Return(nil, &bigcache.EntryNotFoundError{}).Once()
		cacheClient.On("Set", cacheId, expectedBytes).Return(nil).Once()
		cacheClient.On("Get", cacheId).Return(expectedBytes, nil).Once()
		defer cacheClient.AssertExpectations(t)
		defer storeGetter.AssertExpectations(t)

		odataSpec, exists, err := cache.ODataSpec(id)

		require.NoError(t, err)
		assert.True(t, exists)
		assert.Equal(t, expected, odataSpec)
	})

	t.Run("Existing cached object", func(t *testing.T) {
		notExpected := storage.ODataSpec{
			Raw: "example",
		}
		notExpectedBytes, err := convertToCache(notExpected)
		if err != nil {
			t.Error(err)
		}

		storeGetter := storage.NewStoreGetter()
		cacheClient := new(automock.Cache)

		cache := storage.NewCache(storeGetter, cacheClient)
		stop := make(chan struct{})
		defer close(stop)
		notifications := storage.NewNotificationChan()
		defer close(notifications)

		storeGetter.On(function, id).Return(expected, true, nil).Once()
		cacheClient.On("Get", cacheId).Return(notExpectedBytes, nil).Once()
		cacheClient.On("Set", cacheId, expectedBytes).Return(nil).Once()
		cacheClient.On("Get", cacheId).Return(expectedBytes, nil).Once()
		defer cacheClient.AssertExpectations(t)
		defer storeGetter.AssertExpectations(t)

		odataSpec, exists, err := cache.ODataSpec(id)

		require.NoError(t, err)
		assert.True(t, exists)
		assert.Equal(t, expected, odataSpec)
	})
}

func TestCache_AsyncApiSpec_Initialized(t *testing.T) {
	filename := "asyncApiSpec.json"
	fieldName := "asyncApiSpec"
	function := "AsyncApiSpec"

	id := "some-object"
	cacheId := fmt.Sprintf("%s/%s/%s", id, filename, fieldName)

	expected := new(storage.AsyncApiSpec)
	expectedBytes, err := convertToCache(&expected)
	if err != nil {
		t.Error(err)
	}

	t.Run("Not existing object", func(t *testing.T) {
		storeGetter := storage.NewStoreGetter()
		cacheClient := new(automock.Cache)

		cache := storage.NewCache(storeGetter, cacheClient)
		stop := make(chan struct{})
		defer close(stop)
		notifications := storage.NewNotificationChan()
		defer close(notifications)

		storeGetter.On("NotificationChannel", mock.Anything).Return(storage.GetDirectNotificationChan(notifications)).Once()
		storeGetter.On(function, id).Return(nil, false, nil).Once()
		cacheClient.On("Reset").Return(nil).Once()
		cacheClient.On("Get", cacheId).Return(nil, &bigcache.EntryNotFoundError{}).Twice()
		cacheClient.On("Delete", cacheId).Return(nil).Once()
		defer cacheClient.AssertExpectations(t)
		defer storeGetter.AssertExpectations(t)

		cache.Initialize(stop)
		err := waitAtMost(cache.IsSynced, synchronizationTimeout)
		require.NoError(t, err)

		_, exists, err := cache.AsyncApiSpec(id)

		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("Existing not cached object", func(t *testing.T) {
		storeGetter := storage.NewStoreGetter()
		cacheClient := new(automock.Cache)

		cache := storage.NewCache(storeGetter, cacheClient)
		stop := make(chan struct{})
		defer close(stop)
		notifications := storage.NewNotificationChan()
		defer close(notifications)

		storeGetter.On("NotificationChannel", mock.Anything).Return(storage.GetDirectNotificationChan(notifications)).Once()
		storeGetter.On(function, id).Return(expected, true, nil).Once()
		cacheClient.On("Reset").Return(nil).Once()
		cacheClient.On("Get", cacheId).Return(nil, &bigcache.EntryNotFoundError{}).Once()
		cacheClient.On("Set", cacheId, expectedBytes).Return(nil).Once()
		cacheClient.On("Get", cacheId).Return(expectedBytes, nil).Once()
		defer cacheClient.AssertExpectations(t)
		defer storeGetter.AssertExpectations(t)

		cache.Initialize(stop)
		err := waitAtMost(cache.IsSynced, synchronizationTimeout)
		require.NoError(t, err)

		apiSpec, exists, err := cache.AsyncApiSpec(id)

		require.NoError(t, err)
		assert.True(t, exists)
		assert.Equal(t, expected, apiSpec)
	})

	t.Run("Existing cached object", func(t *testing.T) {
		storeGetter := storage.NewStoreGetter()
		cacheClient := new(automock.Cache)

		cache := storage.NewCache(storeGetter, cacheClient)
		stop := make(chan struct{})
		defer close(stop)
		notifications := storage.NewNotificationChan()
		defer close(notifications)

		storeGetter.On("NotificationChannel", mock.Anything).Return(storage.GetDirectNotificationChan(notifications)).Once()
		cacheClient.On("Reset").Return(nil).Once()
		cacheClient.On("Get", cacheId).Return(expectedBytes, nil).Once()
		defer cacheClient.AssertExpectations(t)
		defer storeGetter.AssertExpectations(t)

		cache.Initialize(stop)
		err := waitAtMost(cache.IsSynced, synchronizationTimeout)
		require.NoError(t, err)

		apiSpec, exists, err := cache.AsyncApiSpec(id)

		require.NoError(t, err)
		assert.True(t, exists)
		assert.Equal(t, expected, apiSpec)
	})

	t.Run("store error", func(t *testing.T) {
		storeGetter := storage.NewStoreGetter()
		cacheClient := new(automock.Cache)

		cache := storage.NewCache(storeGetter, cacheClient)
		stop := make(chan struct{})
		defer close(stop)
		notifications := storage.NewNotificationChan()
		defer close(notifications)

		storeGetter.On("NotificationChannel", mock.Anything).Return(storage.GetDirectNotificationChan(notifications)).Once()
		storeGetter.On(function, id).Return(nil, false, errors.New(id))
		cacheClient.On("Reset").Return(nil).Once()
		cacheClient.On("Get", cacheId).Return(nil, &bigcache.EntryNotFoundError{}).Once()
		defer cacheClient.AssertExpectations(t)
		defer storeGetter.AssertExpectations(t)

		cache.Initialize(stop)
		err := waitAtMost(cache.IsSynced, synchronizationTimeout)
		require.NoError(t, err)

		_, exists, err := cache.AsyncApiSpec(id)

		require.Error(t, err)
		assert.False(t, exists)
	})

	t.Run("Cache error", func(t *testing.T) {
		storeGetter := storage.NewStoreGetter()
		cacheClient := new(automock.Cache)

		cache := storage.NewCache(storeGetter, cacheClient)
		stop := make(chan struct{})
		defer close(stop)
		notifications := storage.NewNotificationChan()
		defer close(notifications)

		storeGetter.On("NotificationChannel", mock.Anything).Return(storage.GetDirectNotificationChan(notifications)).Once()
		cacheClient.On("Reset").Return(nil).Once()
		cacheClient.On("Get", cacheId).Return(nil, errors.New(id)).Once()
		defer cacheClient.AssertExpectations(t)
		defer storeGetter.AssertExpectations(t)

		cache.Initialize(stop)
		err := waitAtMost(cache.IsSynced, synchronizationTimeout)
		require.NoError(t, err)

		_, exists, err := cache.AsyncApiSpec(id)

		require.Error(t, err)
		assert.False(t, exists)
	})

	t.Run("Cache error after update", func(t *testing.T) {
		storeGetter := storage.NewStoreGetter()
		cacheClient := new(automock.Cache)

		cache := storage.NewCache(storeGetter, cacheClient)
		stop := make(chan struct{})
		defer close(stop)
		notifications := storage.NewNotificationChan()
		defer close(notifications)

		storeGetter.On("NotificationChannel", mock.Anything).Return(storage.GetDirectNotificationChan(notifications)).Once()
		storeGetter.On(function, id).Return(expected, true, nil).Once()
		cacheClient.On("Reset").Return(nil).Once()
		cacheClient.On("Get", cacheId).Return(nil, &bigcache.EntryNotFoundError{}).Once()
		cacheClient.On("Set", cacheId, expectedBytes).Return(nil).Once()
		cacheClient.On("Get", cacheId).Return(nil, errors.New(id)).Once()
		defer cacheClient.AssertExpectations(t)
		defer storeGetter.AssertExpectations(t)

		cache.Initialize(stop)
		err := waitAtMost(cache.IsSynced, synchronizationTimeout)
		require.NoError(t, err)

		_, exists, err := cache.AsyncApiSpec(id)

		require.Error(t, err)
		assert.False(t, exists)
	})

	t.Run("Error while storing in cache", func(t *testing.T) {
		storeGetter := storage.NewStoreGetter()
		cacheClient := new(automock.Cache)

		cache := storage.NewCache(storeGetter, cacheClient)
		stop := make(chan struct{})
		defer close(stop)
		notifications := storage.NewNotificationChan()
		defer close(notifications)

		storeGetter.On("NotificationChannel", mock.Anything).Return(storage.GetDirectNotificationChan(notifications)).Once()
		storeGetter.On(function, id).Return(expected, true, nil).Once()
		cacheClient.On("Reset").Return(nil).Once()
		cacheClient.On("Get", cacheId).Return(nil, &bigcache.EntryNotFoundError{}).Once()
		cacheClient.On("Set", cacheId, expectedBytes).Return(errors.New(id)).Once()
		defer cacheClient.AssertExpectations(t)
		defer storeGetter.AssertExpectations(t)

		cache.Initialize(stop)
		err := waitAtMost(cache.IsSynced, synchronizationTimeout)
		require.NoError(t, err)

		_, exists, err := cache.AsyncApiSpec(id)

		require.Error(t, err)
		assert.False(t, exists)
	})
}

func TestCache_AsyncApiSpec_NotInitialized(t *testing.T) {
	filename := "asyncApiSpec.json"
	fieldName := "asyncApiSpec"
	function := "AsyncApiSpec"

	id := "some-object"
	cacheId := fmt.Sprintf("%s/%s/%s", id, filename, fieldName)

	expected := new(storage.AsyncApiSpec)
	expectedBytes, err := convertToCache(&expected)
	if err != nil {
		t.Error(err)
	}

	t.Run("Not existing object", func(t *testing.T) {
		storeGetter := storage.NewStoreGetter()
		cacheClient := new(automock.Cache)

		cache := storage.NewCache(storeGetter, cacheClient)
		stop := make(chan struct{})
		defer close(stop)
		notifications := storage.NewNotificationChan()
		defer close(notifications)

		storeGetter.On(function, id).Return(nil, false, nil).Once()
		cacheClient.On("Get", cacheId).Return(nil, &bigcache.EntryNotFoundError{}).Twice()
		cacheClient.On("Delete", cacheId).Return(nil).Once()
		defer cacheClient.AssertExpectations(t)
		defer storeGetter.AssertExpectations(t)

		_, exists, err := cache.AsyncApiSpec(id)

		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("Existing not cached object", func(t *testing.T) {
		storeGetter := storage.NewStoreGetter()
		cacheClient := new(automock.Cache)

		cache := storage.NewCache(storeGetter, cacheClient)
		stop := make(chan struct{})
		defer close(stop)
		notifications := storage.NewNotificationChan()
		defer close(notifications)

		storeGetter.On(function, id).Return(expected, true, nil).Once()
		cacheClient.On("Get", cacheId).Return(nil, &bigcache.EntryNotFoundError{}).Once()
		cacheClient.On("Set", cacheId, expectedBytes).Return(nil).Once()
		cacheClient.On("Get", cacheId).Return(expectedBytes, nil).Once()
		defer cacheClient.AssertExpectations(t)
		defer storeGetter.AssertExpectations(t)

		apiSpec, exists, err := cache.AsyncApiSpec(id)

		require.NoError(t, err)
		assert.True(t, exists)
		assert.Equal(t, expected, apiSpec)
	})

	t.Run("Existing cached object", func(t *testing.T) {
		notExpected := storage.AsyncApiSpec{
			Raw: map[string]interface{}{
				"test": nil,
			},
		}
		notExpectedBytes, err := convertToCache(notExpected)
		if err != nil {
			t.Error(err)
		}

		storeGetter := storage.NewStoreGetter()
		cacheClient := new(automock.Cache)

		cache := storage.NewCache(storeGetter, cacheClient)
		stop := make(chan struct{})
		defer close(stop)
		notifications := storage.NewNotificationChan()
		defer close(notifications)

		storeGetter.On(function, id).Return(expected, true, nil).Once()
		cacheClient.On("Get", cacheId).Return(notExpectedBytes, nil).Once()
		cacheClient.On("Set", cacheId, expectedBytes).Return(nil).Once()
		cacheClient.On("Get", cacheId).Return(expectedBytes, nil).Once()
		defer cacheClient.AssertExpectations(t)
		defer storeGetter.AssertExpectations(t)

		apiSpec, exists, err := cache.AsyncApiSpec(id)

		require.NoError(t, err)
		assert.True(t, exists)
		assert.Equal(t, expected, apiSpec)
	})
}

func TestCache_Content_Initialized(t *testing.T) {
	filename := "content.json"
	fieldName := "content"
	function := "Content"

	id := "some-object"
	cacheId := fmt.Sprintf("%s/%s/%s", id, filename, fieldName)

	expected := new(storage.Content)
	expectedBytes, err := convertToCache(&expected)
	if err != nil {
		t.Error(err)
	}

	t.Run("Not existing object", func(t *testing.T) {
		storeGetter := storage.NewStoreGetter()
		cacheClient := new(automock.Cache)

		cache := storage.NewCache(storeGetter, cacheClient)
		stop := make(chan struct{})
		defer close(stop)
		notifications := storage.NewNotificationChan()
		defer close(notifications)

		storeGetter.On("NotificationChannel", mock.Anything).Return(storage.GetDirectNotificationChan(notifications)).Once()
		storeGetter.On(function, id).Return(nil, false, nil).Once()
		cacheClient.On("Reset").Return(nil).Once()
		cacheClient.On("Get", cacheId).Return(nil, &bigcache.EntryNotFoundError{}).Twice()
		cacheClient.On("Delete", cacheId).Return(nil).Once()
		defer cacheClient.AssertExpectations(t)
		defer storeGetter.AssertExpectations(t)

		cache.Initialize(stop)
		err := waitAtMost(cache.IsSynced, synchronizationTimeout)
		require.NoError(t, err)

		_, exists, err := cache.Content(id)

		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("Existing not cached object", func(t *testing.T) {
		storeGetter := storage.NewStoreGetter()
		cacheClient := new(automock.Cache)

		cache := storage.NewCache(storeGetter, cacheClient)
		stop := make(chan struct{})
		defer close(stop)
		notifications := storage.NewNotificationChan()
		defer close(notifications)

		storeGetter.On("NotificationChannel", mock.Anything).Return(storage.GetDirectNotificationChan(notifications)).Once()
		storeGetter.On(function, id).Return(expected, true, nil).Once()
		cacheClient.On("Reset").Return(nil).Once()
		cacheClient.On("Get", cacheId).Return(nil, &bigcache.EntryNotFoundError{}).Once()
		cacheClient.On("Set", cacheId, expectedBytes).Return(nil).Once()
		cacheClient.On("Get", cacheId).Return(expectedBytes, nil).Once()
		defer cacheClient.AssertExpectations(t)
		defer storeGetter.AssertExpectations(t)

		cache.Initialize(stop)
		err := waitAtMost(cache.IsSynced, synchronizationTimeout)
		require.NoError(t, err)

		apiSpec, exists, err := cache.Content(id)

		require.NoError(t, err)
		assert.True(t, exists)
		assert.Equal(t, expected, apiSpec)
	})

	t.Run("Existing cached object", func(t *testing.T) {
		storeGetter := storage.NewStoreGetter()
		cacheClient := new(automock.Cache)

		cache := storage.NewCache(storeGetter, cacheClient)
		stop := make(chan struct{})
		defer close(stop)
		notifications := storage.NewNotificationChan()
		defer close(notifications)

		storeGetter.On("NotificationChannel", mock.Anything).Return(storage.GetDirectNotificationChan(notifications)).Once()
		cacheClient.On("Reset").Return(nil).Once()
		cacheClient.On("Get", cacheId).Return(expectedBytes, nil).Once()
		defer cacheClient.AssertExpectations(t)
		defer storeGetter.AssertExpectations(t)

		cache.Initialize(stop)
		err := waitAtMost(cache.IsSynced, synchronizationTimeout)
		require.NoError(t, err)

		apiSpec, exists, err := cache.Content(id)

		require.NoError(t, err)
		assert.True(t, exists)
		assert.Equal(t, expected, apiSpec)
	})

	t.Run("store error", func(t *testing.T) {
		storeGetter := storage.NewStoreGetter()
		cacheClient := new(automock.Cache)

		cache := storage.NewCache(storeGetter, cacheClient)
		stop := make(chan struct{})
		defer close(stop)
		notifications := storage.NewNotificationChan()
		defer close(notifications)

		storeGetter.On("NotificationChannel", mock.Anything).Return(storage.GetDirectNotificationChan(notifications)).Once()
		storeGetter.On(function, id).Return(nil, false, errors.New(id))
		cacheClient.On("Reset").Return(nil).Once()
		cacheClient.On("Get", cacheId).Return(nil, &bigcache.EntryNotFoundError{}).Once()
		defer cacheClient.AssertExpectations(t)
		defer storeGetter.AssertExpectations(t)

		cache.Initialize(stop)
		err := waitAtMost(cache.IsSynced, synchronizationTimeout)
		require.NoError(t, err)

		_, exists, err := cache.Content(id)

		require.Error(t, err)
		assert.False(t, exists)
	})

	t.Run("Cache error", func(t *testing.T) {
		storeGetter := storage.NewStoreGetter()
		cacheClient := new(automock.Cache)

		cache := storage.NewCache(storeGetter, cacheClient)
		stop := make(chan struct{})
		defer close(stop)
		notifications := storage.NewNotificationChan()
		defer close(notifications)

		storeGetter.On("NotificationChannel", mock.Anything).Return(storage.GetDirectNotificationChan(notifications)).Once()
		cacheClient.On("Reset").Return(nil).Once()
		cacheClient.On("Get", cacheId).Return(nil, errors.New(id)).Once()
		defer cacheClient.AssertExpectations(t)
		defer storeGetter.AssertExpectations(t)

		cache.Initialize(stop)
		err := waitAtMost(cache.IsSynced, synchronizationTimeout)
		require.NoError(t, err)

		_, exists, err := cache.Content(id)

		require.Error(t, err)
		assert.False(t, exists)
	})

	t.Run("Cache error after update", func(t *testing.T) {
		storeGetter := storage.NewStoreGetter()
		cacheClient := new(automock.Cache)

		cache := storage.NewCache(storeGetter, cacheClient)
		stop := make(chan struct{})
		defer close(stop)
		notifications := storage.NewNotificationChan()
		defer close(notifications)

		storeGetter.On("NotificationChannel", mock.Anything).Return(storage.GetDirectNotificationChan(notifications)).Once()
		storeGetter.On(function, id).Return(expected, true, nil).Once()
		cacheClient.On("Reset").Return(nil).Once()
		cacheClient.On("Get", cacheId).Return(nil, &bigcache.EntryNotFoundError{}).Once()
		cacheClient.On("Set", cacheId, expectedBytes).Return(nil).Once()
		cacheClient.On("Get", cacheId).Return(nil, errors.New(id)).Once()
		defer cacheClient.AssertExpectations(t)
		defer storeGetter.AssertExpectations(t)

		cache.Initialize(stop)
		err := waitAtMost(cache.IsSynced, synchronizationTimeout)
		require.NoError(t, err)

		_, exists, err := cache.Content(id)

		require.Error(t, err)
		assert.False(t, exists)
	})

	t.Run("Error while storing in cache", func(t *testing.T) {
		storeGetter := storage.NewStoreGetter()
		cacheClient := new(automock.Cache)

		cache := storage.NewCache(storeGetter, cacheClient)
		stop := make(chan struct{})
		defer close(stop)
		notifications := storage.NewNotificationChan()
		defer close(notifications)

		storeGetter.On("NotificationChannel", mock.Anything).Return(storage.GetDirectNotificationChan(notifications)).Once()
		storeGetter.On(function, id).Return(expected, true, nil).Once()
		cacheClient.On("Reset").Return(nil).Once()
		cacheClient.On("Get", cacheId).Return(nil, &bigcache.EntryNotFoundError{}).Once()
		cacheClient.On("Set", cacheId, expectedBytes).Return(errors.New(id)).Once()
		defer cacheClient.AssertExpectations(t)
		defer storeGetter.AssertExpectations(t)

		cache.Initialize(stop)
		err := waitAtMost(cache.IsSynced, synchronizationTimeout)
		require.NoError(t, err)

		_, exists, err := cache.Content(id)

		require.Error(t, err)
		assert.False(t, exists)
	})
}

func TestCache_Content_NotInitialized(t *testing.T) {
	filename := "content.json"
	fieldName := "content"
	function := "Content"

	id := "some-object"
	cacheId := fmt.Sprintf("%s/%s/%s", id, filename, fieldName)

	expected := new(storage.Content)
	expectedBytes, err := convertToCache(&expected)
	if err != nil {
		t.Error(err)
	}

	t.Run("Not existing object", func(t *testing.T) {
		storeGetter := storage.NewStoreGetter()
		cacheClient := new(automock.Cache)

		cache := storage.NewCache(storeGetter, cacheClient)
		stop := make(chan struct{})
		defer close(stop)
		notifications := storage.NewNotificationChan()
		defer close(notifications)

		storeGetter.On(function, id).Return(nil, false, nil).Once()
		cacheClient.On("Get", cacheId).Return(nil, &bigcache.EntryNotFoundError{}).Twice()
		cacheClient.On("Delete", cacheId).Return(nil).Once()
		defer cacheClient.AssertExpectations(t)
		defer storeGetter.AssertExpectations(t)

		_, exists, err := cache.Content(id)

		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("Existing not cached object", func(t *testing.T) {
		storeGetter := storage.NewStoreGetter()
		cacheClient := new(automock.Cache)

		cache := storage.NewCache(storeGetter, cacheClient)
		stop := make(chan struct{})
		defer close(stop)
		notifications := storage.NewNotificationChan()
		defer close(notifications)

		storeGetter.On(function, id).Return(expected, true, nil).Once()
		cacheClient.On("Get", cacheId).Return(nil, &bigcache.EntryNotFoundError{}).Once()
		cacheClient.On("Set", cacheId, expectedBytes).Return(nil).Once()
		cacheClient.On("Get", cacheId).Return(expectedBytes, nil).Once()
		defer cacheClient.AssertExpectations(t)
		defer storeGetter.AssertExpectations(t)

		apiSpec, exists, err := cache.Content(id)

		require.NoError(t, err)
		assert.True(t, exists)
		assert.Equal(t, expected, apiSpec)
	})

	t.Run("Existing cached object", func(t *testing.T) {
		notExpected := storage.Content{
			Raw: map[string]interface{}{
				"test": nil,
			},
		}
		notExpectedBytes, err := convertToCache(notExpected)
		if err != nil {
			t.Error(err)
		}

		storeGetter := storage.NewStoreGetter()
		cacheClient := new(automock.Cache)

		cache := storage.NewCache(storeGetter, cacheClient)
		stop := make(chan struct{})
		defer close(stop)
		notifications := storage.NewNotificationChan()
		defer close(notifications)

		storeGetter.On(function, id).Return(expected, true, nil).Once()
		cacheClient.On("Get", cacheId).Return(notExpectedBytes, nil).Once()
		cacheClient.On("Set", cacheId, expectedBytes).Return(nil).Once()
		cacheClient.On("Get", cacheId).Return(expectedBytes, nil).Once()
		defer cacheClient.AssertExpectations(t)
		defer storeGetter.AssertExpectations(t)

		apiSpec, exists, err := cache.Content(id)

		require.NoError(t, err)
		assert.True(t, exists)
		assert.Equal(t, expected, apiSpec)
	})
}

func convertToCache(object interface{}) ([]byte, error) {
	buffer := new(bytes.Buffer)
	err := gob.NewEncoder(buffer).Encode(object)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func waitAtMost(fn func() bool, duration time.Duration) error {
	timeout := time.After(duration)
	tick := time.Tick(1 * time.Millisecond)

	for {
		select {
		case <-timeout:
			return errors.New(fmt.Sprintf("Waiting for resource failed in given timeout %f second(s)", duration.Seconds()))
		case <-tick:
			if fn() {
				return nil
			}
		}
	}
}
