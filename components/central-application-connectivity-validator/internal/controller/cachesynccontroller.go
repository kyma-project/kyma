package controller

import (
	"context"

	"github.com/kyma-project/kyma/common/logging/logger"
	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	gocache "github.com/patrickmn/go-cache"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const controllerName = "cache_sync_controller"

type Controller interface {
	Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error)
	SetupWithManager(mgr ctrl.Manager) error
}

type controller struct {
	client   client.Client
	appCache *gocache.Cache
	log      *logger.Logger
}

func NewController(log *logger.Logger, client client.Client, appCache *gocache.Cache) Controller {
	return &controller{
		client:   client,
		appCache: appCache,
		log:      log,
	}
}

func (c *controller) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	key := request.Name
	var application v1alpha1.Application
	if err := c.client.Get(ctx, request.NamespacedName, &application); err != nil {
		err = client.IgnoreNotFound(err)
		if err != nil {
			c.log.WithContext().
				With("controller", controllerName).
				With("name", application.Name).
				Error("Unable to fetch application: %s", err.Error())
		} else {
			c.appCache.Delete(key)
			c.log.WithContext().
				With("controller", controllerName).
				With("name", application.Name).
				Infof("Deleted the application from the cache.")
		}
		return ctrl.Result{}, err
	}
	err := c.syncHandler(&application)
	return reconcile.Result{}, err
}

func (c *controller) syncHandler(application *v1alpha1.Application) error {
	key := application.Name
	if !application.ObjectMeta.DeletionTimestamp.IsZero() {
		c.appCache.Delete(key)
		c.log.WithContext().
			With("controller", controllerName).
			With("name", application.Name).
			Infof("Deleted the application from the cache on graceful deletion.")
		return nil
	}

	applicationClientIDs := c.getClientIDsFromResource(application)
	c.appCache.Set(key, applicationClientIDs, gocache.DefaultExpiration)
	c.log.WithContext().
		With("controller", controllerName).
		With("name", application.Name).
		Infof("Added/Updated the application in the cache with values %v.", applicationClientIDs)
	return nil
}

func (c *controller) getClientIDsFromResource(application *v1alpha1.Application) []string {
	if application.Spec.CompassMetadata == nil {
		return []string{}
	}
	return application.Spec.CompassMetadata.Authentication.ClientIds
}

func (c *controller) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Application{}).
		Complete(c)
}
