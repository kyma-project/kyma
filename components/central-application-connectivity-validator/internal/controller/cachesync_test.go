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
				appCache.Set(applicationName, []string{}, cache.DefaultExpiration)
			},
			check: notFoundInCache,
		},
		{
			name: "Add new application to cache without compass metadata",
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
				require.Equal(t, []string{}, v)
			},
		},
		{
			name: "Overwrite authentication clients in cache",
			setup: func(t *testing.T, applicationName string, fc *fakeClient, appCache *cache.Cache) {
				appCache.Set(applicationName, []string{"client-1"}, cache.DefaultExpiration)
				require.NoError(t, fc.Create(&v1alpha1.Application{
					ObjectMeta: v1.ObjectMeta{
						Name: applicationName,
					},
				}))
			},
			check: func(t *testing.T, applicationName string, appCache *cache.Cache) {
				v, found := appCache.Get(applicationName)
				require.True(t, found)
				require.Equal(t, []string{}, v)
			},
		},
		{
			name: "Add new application to cache with not authentication clients",
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
				require.Equal(t, []string{"client-1", "client-2"}, v)
			},
		},
		{
			name: "Delete application from cache",
			setup: func(t *testing.T, applicationName string, fc *fakeClient, appCache *cache.Cache) {
				appCache.Set(applicationName, []string{}, cache.DefaultExpiration)

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

			cacheSync := NewCacheSync(log, fc, appCache, "test-controller")
			err = cacheSync.Sync(context.Background(), applicationName)
			require.NoError(t, err)

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

func (c fakeClient) Create(application *v1alpha1.Application) error {
	_, err := c.intf.Create(context.Background(), application, v1.CreateOptions{})
	return err
}
