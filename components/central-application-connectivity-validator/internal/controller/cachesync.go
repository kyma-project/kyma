package controller

import (
	"context"
	"github.com/kyma-project/kyma/common/logging/logger"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/apis/applicationconnector/v1alpha1"
	gocache "github.com/patrickmn/go-cache"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

type CacheSync interface {
	Sync(ctx context.Context, applicationName string) error
	Init(ctx context.Context)
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

func NewCacheSync(
	log *logger.Logger,
	client client.Reader,
	appCache *gocache.Cache,
	controllerName,
	appNamePlaceholder,
	eventingPathPrefixV1,
	eventingPathPrefixV2,
	eventingPathPrefixEvents string) CacheSync {
	return &cacheSync{
		client:                   client,
		appCache:                 appCache,
		log:                      log,
		controllerName:           controllerName,
		appNamePlaceholder:       appNamePlaceholder,
		eventingPathPrefixV1:     eventingPathPrefixV1,
		eventingPathPrefixV2:     eventingPathPrefixV2,
		eventingPathPrefixEvents: eventingPathPrefixEvents,
	}
}

func (c *cacheSync) Init(ctx context.Context) {

	c.log.WithContext().With("controller", c.controllerName).Infof("Cache initialisation")

	var applicationList v1alpha1.ApplicationList
	err := c.client.List(ctx, &applicationList)

	if apierrors.IsNotFound(err) {
		c.log.WithContext().Infof("No application are present on the cluster")
	}

	if err != nil {
		c.log.WithContext().Warnf("Unable to read applications")
	}

	for _, app := range applicationList.Items {
		c.syncApplication(&app)
	}
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
	c.syncApplication(&application)
	return nil
}

func (c *cacheSync) syncApplication(application *v1alpha1.Application) {
	key := application.Name
	if !application.DeletionTimestamp.IsZero() {
		c.appCache.Delete(key)
		c.log.WithContext().
			With("controller", c.controllerName).
			With("name", application.Name).
			Infof("Deleted the application from the cache on graceful deletion.")
		return
	}

	applicationInfo := c.getAppDataFromResource(application)
	c.appCache.Set(key, applicationInfo, gocache.DefaultExpiration)
	c.log.WithContext().
		With("controller", c.controllerName).
		With("name", application.Name).
		Infof("Added/Updated the application in the cache with values %v.", applicationInfo)
}

func (c *cacheSync) getAppDataFromResource(application *v1alpha1.Application) CachedAppData {

	appData := CachedAppData{ClientIDs: []string{}}

	appData.AppPathPrefixV1 = c.getApplicationPrefix(c.eventingPathPrefixV1, application.Name)
	appData.AppPathPrefixV2 = c.getApplicationPrefix(c.eventingPathPrefixV2, application.Name)
	appData.AppPathPrefixEvents = c.getApplicationPrefix(c.eventingPathPrefixEvents, application.Name)

	if application.Spec.CompassMetadata != nil {
		appData.ClientIDs = append(appData.ClientIDs, application.Spec.CompassMetadata.Authentication.ClientIds...)
	}

	return appData
}

func (c *cacheSync) getApplicationPrefix(path string, applicationName string) string {
	if c.appNamePlaceholder != "" {
		return strings.ReplaceAll(path, c.appNamePlaceholder, applicationName)
	}
	return path
}
