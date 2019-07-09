package controller

import (
	"context"

	"github.com/kyma-project/kyma/components/helm-broker/internal"
	addonsv1alpha1 "github.com/kyma-project/kyma/components/helm-broker/pkg/apis/addons/v1alpha1"
	exerr "github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type clusterBrokerFacade interface {
	Create() error
	Exist() (bool, error)
	Delete() error
}

type clusterDocsProvider interface {
	EnsureClusterDocsTopic(bundle *internal.Bundle) error
	EnsureClusterDocsTopicRemoved(id string) error
}

type clusterBrokerSyncer interface {
	Sync() error
}

// ClusterAddonsConfigurationController holds controller logic
type ClusterAddonsConfigurationController struct {
	reconciler reconcile.Reconciler
}

// NewClusterAddonsConfigurationController creates new controller with a given reconciler
func NewClusterAddonsConfigurationController(reconciler reconcile.Reconciler) *ClusterAddonsConfigurationController {
	return &ClusterAddonsConfigurationController{reconciler: reconciler}
}

// Start starts a controller
func (cacc *ClusterAddonsConfigurationController) Start(mgr manager.Manager) error {
	// Create a new controller
	c, err := controller.New("clusteraddonsconfiguration-controller", mgr, controller.Options{Reconciler: cacc.reconciler})
	if err != nil {
		return err
	}

	// Watch for changes to ClusterAddonsConfiguration
	err = c.Watch(&source.Kind{Type: &addonsv1alpha1.ClusterAddonsConfiguration{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileClusterAddonsConfiguration{}

// ReconcileClusterAddonsConfiguration reconciles a ClusterAddonsConfiguration object
type ReconcileClusterAddonsConfiguration struct {
	log logrus.FieldLogger
	client.Client
	scheme *runtime.Scheme

	clusterBrokerFacade clusterBrokerFacade
	clusterDocsProvider clusterDocsProvider
	clusterBrokerSyncer clusterBrokerSyncer

	// syncBroker informs ServiceBroker should be resync, it should be true if
	// operation insert/delete was made on storage
	syncBroker bool
}

// NewReconcileClusterAddonsConfiguration returns a new reconcile.Reconciler
func NewReconcileClusterAddonsConfiguration(mgr manager.Manager, clusterBrokerFacade clusterBrokerFacade, clusterDocsProvider clusterDocsProvider, clusterBrokerSyncer clusterBrokerSyncer) reconcile.Reconciler {
	return &ReconcileClusterAddonsConfiguration{
		log:    logrus.WithField("controller", "cluster-addons-configuration"),
		Client: mgr.GetClient(),
		scheme: mgr.GetScheme(),

		clusterBrokerFacade: clusterBrokerFacade,
		clusterDocsProvider: clusterDocsProvider,
		clusterBrokerSyncer: clusterBrokerSyncer,

		syncBroker: false,
	}
}

// Reconcile reads that state of the cluster for a ClusterAddonsConfiguration object and makes changes based on the state read
// and what is in the ClusterAddonsConfiguration.Spec
func (r *ReconcileClusterAddonsConfiguration) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Fetch the ClusterAddonsConfiguration instance
	instance := &addonsv1alpha1.ClusterAddonsConfiguration{}
	err := r.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	exist, err := r.clusterBrokerFacade.Exist()
	if err != nil {
		return reconcile.Result{}, exerr.Wrap(err, "while checking if ClusterServiceBroker exists")
	}
	if !exist {
		if err := r.clusterBrokerFacade.Create(); err != nil {
			return reconcile.Result{}, exerr.Wrapf(err, "while creating ClusterServiceBroker for addon %s in namespace %s", instance.Name, instance.Namespace)
		}
	} else if r.syncBroker {
		if err := r.clusterBrokerSyncer.Sync(); err != nil {
			return reconcile.Result{}, exerr.Wrapf(err, "while syncing ClusterServiceBroker for addon %s in namespace %s", instance.Name, instance.Namespace)
		}
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileClusterAddonsConfiguration) deleteAddonsProcess(namespace string) error {
	r.log.Infof("Start delete AddonsConfiguration from namespace %s", namespace)

	addonsCfgs, err := r.addonsConfigurationList(namespace)
	if err != nil {
		return exerr.Wrapf(err, "while listing AddonsConfigurations in namespace %s", namespace)
	}

	deleteBroker := true
	for _, addon := range addonsCfgs.Items {
		if addon.Status.Phase != addonsv1alpha1.AddonsConfigurationReady {
			// reprocess ClusterAddonConfig again if it was failed
			addon.Spec.ReprocessRequest++
			if err := r.Client.Update(context.Background(), &addon); err != nil {
				return exerr.Wrapf(err, "while incrementing a reprocess requests for AddonConfiguration %s/%s", addon.Name, addon.Namespace)
			}
		} else {
			deleteBroker = false
		}
	}
	if deleteBroker {
		if err := r.clusterBrokerFacade.Delete(); err != nil {
			return exerr.Wrapf(err, "while deleting ServiceBroker from namespace %s", namespace)
		}
	}

	r.log.Info("Delete AddonsConfiguration process completed")
	return nil
}

func (r *ReconcileClusterAddonsConfiguration) addonsConfigurationList(namespace string) (*addonsv1alpha1.AddonsConfigurationList, error) {
	addonsList := &addonsv1alpha1.AddonsConfigurationList{}
	addonsConfigurationList := &addonsv1alpha1.ClusterAddonsConfigurationList{}

	err := r.Client.List(context.TODO(), &client.ListOptions{}, addonsConfigurationList)
	if err != nil {
		return addonsList, exerr.Wrap(err, "during fetching AddonConfiguration list by client")
	}

	return addonsList, nil
}
