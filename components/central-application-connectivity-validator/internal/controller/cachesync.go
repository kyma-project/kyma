package controller

import (
	"context"
	"github.com/kyma-project/kyma/common/logging/logger"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/apis/applicationconnector/v1alpha1"
	gocache "github.com/patrickmn/go-cache"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

type CacheSync interface {
	Sync(ctx context.Context, applicationName string) error
	Init(ctx context.Context) error
}

type cacheSync struct {
	client                   client.Reader
	appCache                 *gocache.Cache
	log                      *logger.Logger
	controllerName           string
	eventingPathPrefixV1     string
	eventingPathPrefixV2     string
	eventingPathPrefixEvents string
	appNamePlaceholder       string
}

type CachedAppData struct {
	ClientIDs           []string
	AppPathPrefixV1     string
	AppPathPrefixV2     string
	AppPathPrefixEvents string
}

func NewCacheSync(log *logger.Logger, client client.Reader, appCache *gocache.Cache, controllerName, eventingPathPrefixV1, eventingPathPrefixV2, eventingPathPrefixEvents, appNamePlaceholder string) CacheSync {
	return &cacheSync{
		client:                   client,
		appCache:                 appCache,
		log:                      log,
		controllerName:           controllerName,
		eventingPathPrefixV1:     eventingPathPrefixV1,
		eventingPathPrefixV2:     eventingPathPrefixV2,
		eventingPathPrefixEvents: eventingPathPrefixEvents,
		appNamePlaceholder:       appNamePlaceholder,
	}
}

func (c *cacheSync) Init(ctx context.Context) error {

	c.log.WithContext().With("controller", c.controllerName).Infof("Cache initialisation")

	var applicationList v1alpha1.ApplicationList

	if err := c.client.List(ctx, &applicationList); err != nil {
		err = client.IgnoreNotFound(err)
		if err != nil {
			c.log.WithContext().Infof("Unable to read applications")
		} else {
			c.log.WithContext().Infof("No application are present on the cluster")
		}
		return err
	} else {
		for _, app := range applicationList.Items {
			c.syncApplication(&app)
		}
	}
	return nil
}

func (c *cacheSync) Sync(ctx context.Context, applicationName string) error {
	var application v1alpha1.Application
	if err := c.client.Get(ctx, types.NamespacedName{Name: applicationName}, &application); err != nil {
		err = client.IgnoreNotFound(err)
		if err != nil {
			c.log.WithContext().
				With("controller", c.controllerName).
				With("name", applicationName).
				Error("Unable to fetch application: %s", err.Error())
		} else {
			c.appCache.Delete(applicationName)
			c.log.WithContext().
				With("controller", c.controllerName).
				With("name", applicationName).
				Infof("Application not found, deleting from the cache.")
		}
		return err
	}
	return c.syncApplication(&application)
}

func (c *cacheSync) syncApplication(application *v1alpha1.Application) error {
	key := application.Name
	if !application.ObjectMeta.DeletionTimestamp.IsZero() {
		c.appCache.Delete(key)
		c.log.WithContext().
			With("controller", c.controllerName).
			With("name", application.Name).
			Infof("Deleted the application from the cache on graceful deletion.")
		return nil
	}

	applicationInfo := c.getAppDataFromResource(application)
	c.appCache.Set(key, applicationInfo, gocache.DefaultExpiration)
	c.log.WithContext().
		With("controller", c.controllerName).
		With("name", application.Name).
		Infof("Added/Updated the application in the cache with values %v.", applicationInfo)
	return nil
}

func (c *cacheSync) getAppDataFromResource(application *v1alpha1.Application) CachedAppData {

	var appData CachedAppData

	appData.AppPathPrefixV1 = c.getApplicationPrefix(c.eventingPathPrefixV1, application.Name)
	appData.AppPathPrefixV2 = c.getApplicationPrefix(c.eventingPathPrefixV2, application.Name)
	appData.AppPathPrefixEvents = c.getApplicationPrefix(c.eventingPathPrefixEvents, application.Name)

	if application.Spec.CompassMetadata == nil {
		appData.ClientIDs = []string{}
	} else {
		appData.ClientIDs = make([]string, len(application.Spec.CompassMetadata.Authentication.ClientIds))
		copy(appData.ClientIDs, application.Spec.CompassMetadata.Authentication.ClientIds)
	}

	return appData
}

func (c *cacheSync) getApplicationPrefix(path string, applicationName string) string {
	if c.appNamePlaceholder != "" {
		return strings.ReplaceAll(path, c.appNamePlaceholder, applicationName)
	}
	return path
}
