package controller

import (
	"context"
	"testing"
	"time"

	"github.com/kyma-project/kyma/common/logging/logger"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/client/clientset/versioned/fake"
	applicationconnectorv1alpha1 "github.com/kyma-project/kyma/components/central-application-gateway/pkg/client/clientset/versioned/typed/applicationconnector/v1alpha1"
	"github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestCacheSync(t *testing.T) {
	const name = "my-app"
	type setup func(t *testing.T, applicationName string, fc *fakeClient, cache *cache.Cache)
	type check func(t *testing.T, applicationName string, cache *cache.Cache)

	notFoundInCache := func(t *testing.T, applicationName string, appCache *cache.Cache) {
		_, found := appCache.Get(applicationName)
		require.False(t, found)
	}

	emptyAppData := CachedAppData{}

	appDataNoClients := CachedAppData{
		ClientIDs:           []string{},
		AppPathPrefixV1:     "/my-app/v1/events",
		AppPathPrefixV2:     "/my-app/v2/events",
		AppPathPrefixEvents: "/my-app/events",
	}

	appData2Clients := CachedAppData{
		ClientIDs:           []string{"client-1", "client-2"},
		AppPathPrefixV1:     "/my-app/v1/events",
		AppPathPrefixV2:     "/my-app/v2/events",
		AppPathPrefixEvents: "/my-app/events",
	}

	appData1Client := CachedAppData{
		ClientIDs:           []string{"client-1"},
		AppPathPrefixV1:     "/my-app/v1/events",
		AppPathPrefixV2:     "/my-app/v2/events",
		AppPathPrefixEvents: "/my-app/events",
	}

	tests := []struct {
		name  string
		setup setup
		check check
	}{
		{
			name:  "Application already removed from cache",
			check: notFoundInCache,
		},
		{
			name: "Remove application from cache",
			setup: func(t *testing.T, applicationName string, fc *fakeClient, appCache *cache.Cache) {
				appCache.Set(applicationName, emptyAppData, cache.DefaultExpiration)
			},
			check: notFoundInCache,
		},
		{
			name: "Add new application to cache without compass metadata and generate endpoints",
			setup: func(t *testing.T, applicationName string, fc *fakeClient, appCache *cache.Cache) {
				require.NoError(t, fc.Create(&v1alpha1.Application{
					ObjectMeta: v1.ObjectMeta{
						Name: applicationName,
					},
				}))
			},
			check: func(t *testing.T, applicationName string, appCache *cache.Cache) {
				v, found := appCache.Get(applicationName)
				require.True(t, found)
				require.Equal(t, appDataNoClients, v)
			},
		},
		{
			name: "Overwrite authentication clients in cache",
			setup: func(t *testing.T, applicationName string, fc *fakeClient, appCache *cache.Cache) {
				appCache.Set(applicationName, appData1Client, cache.DefaultExpiration)
				require.NoError(t, fc.Create(&v1alpha1.Application{
					ObjectMeta: v1.ObjectMeta{
						Name: applicationName,
					},
				}))
			},
			check: func(t *testing.T, applicationName string, appCache *cache.Cache) {
				v, found := appCache.Get(applicationName)
				require.True(t, found)
				require.Equal(t, appDataNoClients, v)
			},
		},
		{
			name: "Add new application to cache with authentication clients and generate endpoints",
			setup: func(t *testing.T, applicationName string, fc *fakeClient, appCache *cache.Cache) {
				require.NoError(t, fc.Create(&v1alpha1.Application{
					ObjectMeta: v1.ObjectMeta{
						Name: applicationName,
					},
					Spec: v1alpha1.ApplicationSpec{
						CompassMetadata: &v1alpha1.CompassMetadata{
							Authentication: v1alpha1.Authentication{
								ClientIds: []string{"client-1", "client-2"},
							},
						},
					},
				}))
			},
			check: func(t *testing.T, applicationName string, appCache *cache.Cache) {
				v, found := appCache.Get(applicationName)
				require.True(t, found)
				require.Equal(t, appData2Clients, v)
			},
		},
		{
			name: "Delete application from cache",
			setup: func(t *testing.T, applicationName string, fc *fakeClient, appCache *cache.Cache) {
				appCache.Set(applicationName, emptyAppData, cache.DefaultExpiration)

				now := v1.NewTime(time.Now())
				require.NoError(t, fc.Create(&v1alpha1.Application{
					ObjectMeta: v1.ObjectMeta{
						Name:              applicationName,
						DeletionTimestamp: &now,
					},
				}))
			},
			check: notFoundInCache,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			applicationName := name

			log, err := logger.New(logger.TEXT, logger.DEBUG)
			require.NoError(t, err)
			appCache := cache.New(60*time.Second, 60*time.Second)
			fc := NewFakeClient()

			if tc.setup != nil {
				tc.setup(t, applicationName, fc, appCache)
			}

			cacheSync := NewCacheSync(log, fc, appCache, "test-controller", "%%APP_NAME%%", "/%%APP_NAME%%/v1/events", "/%%APP_NAME%%/v2/events", "/%%APP_NAME%%/events")
			err = cacheSync.Sync(context.Background(), applicationName)
			require.NoError(t, err)

			tc.check(t, applicationName, appCache)
		})
	}
}

func TestCacheInit(t *testing.T) {
	const name = "my-app"
	type setup func(t *testing.T, applicationName string, fc *fakeClient, cache *cache.Cache)
	type check func(t *testing.T, applicationName string, cache *cache.Cache)

	notFoundInCache := func(t *testing.T, applicationName string, appCache *cache.Cache) {
		_, found := appCache.Get(applicationName)
		require.False(t, found)
	}

	appDataNoClients := CachedAppData{
		ClientIDs:           []string{},
		AppPathPrefixV1:     "/my-app/v1/events",
		AppPathPrefixV2:     "/my-app/v2/events",
		AppPathPrefixEvents: "/my-app/events",
	}

	appData2Clients := CachedAppData{
		ClientIDs:           []string{"client-1", "client-2"},
		AppPathPrefixV1:     "/my-app/v1/events",
		AppPathPrefixV2:     "/my-app/v2/events",
		AppPathPrefixEvents: "/my-app/events",
	}

	tests := []struct {
		name  string
		setup setup
		check check
	}{
		{
			name:  "Application will not be added to cache",
			check: notFoundInCache,
		},
		{
			name: "Add application to cache without compass metadata and generate endpoints",
			setup: func(t *testing.T, applicationName string, fc *fakeClient, appCache *cache.Cache) {
				require.NoError(t, fc.Create(&v1alpha1.Application{
					ObjectMeta: v1.ObjectMeta{
						Name: applicationName,
					},
				}))
			},
			check: func(t *testing.T, applicationName string, appCache *cache.Cache) {
				v, found := appCache.Get(applicationName)
				require.True(t, found)
				require.Equal(t, appDataNoClients, v)
			},
		},
		{
			name: "Add new application to cache with authentication clients and generate endpoints",
			setup: func(t *testing.T, applicationName string, fc *fakeClient, appCache *cache.Cache) {
				require.NoError(t, fc.Create(&v1alpha1.Application{
					ObjectMeta: v1.ObjectMeta{
						Name: applicationName,
					},
					Spec: v1alpha1.ApplicationSpec{
						CompassMetadata: &v1alpha1.CompassMetadata{
							Authentication: v1alpha1.Authentication{
								ClientIds: []string{"client-1", "client-2"},
							},
						},
					},
				}))
			},
			check: func(t *testing.T, applicationName string, appCache *cache.Cache) {
				v, found := appCache.Get(applicationName)
				require.True(t, found)
				require.Equal(t, appData2Clients, v)
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			applicationName := name

			log, err := logger.New(logger.TEXT, logger.DEBUG)
			require.NoError(t, err)
			appCache := cache.New(60*time.Second, 60*time.Second)
			fc := NewFakeClient()

			if tc.setup != nil {
				tc.setup(t, applicationName, fc, appCache)
			}

			cacheSync := NewCacheSync(log, fc, appCache, "test-controller", "%%APP_NAME%%", "/%%APP_NAME%%/v1/events", "/%%APP_NAME%%/v2/events", "/%%APP_NAME%%/events")
			cacheSync.Init(context.Background())

			tc.check(t, applicationName, appCache)
		})
	}
}

type fakeClient struct {
	client.Reader
	intf applicationconnectorv1alpha1.ApplicationInterface
}

func NewFakeClient() *fakeClient {
	return &fakeClient{
		intf: fake.NewSimpleClientset(&v1alpha1.Application{}).ApplicationconnectorV1alpha1().Applications(),
	}
}

func (c fakeClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	target := obj.(*v1alpha1.Application)
	app, err := c.intf.Get(ctx, key.Name, v1.GetOptions{})
	if err != nil {
		return err
	}
	*target = *app
	return nil
}

func (c fakeClient) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	target := list.(*v1alpha1.ApplicationList)
	appList, err := c.intf.List(ctx, v1.ListOptions{})
	if err != nil {
		return err
	}
	*target = *appList
	return nil
}

func (c fakeClient) Create(application *v1alpha1.Application) error {
	_, err := c.intf.Create(context.Background(), application, v1.CreateOptions{})
	return err
}
