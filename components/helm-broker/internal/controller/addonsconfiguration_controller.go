package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/kyma-project/kyma/components/helm-broker/internal/controller/bundle"
	"github.com/kyma-project/kyma/components/helm-broker/internal/controller/repository"
	"github.com/kyma-project/kyma/components/helm-broker/internal/storage"
	addonsv1alpha1 "github.com/kyma-project/kyma/components/helm-broker/pkg/apis/addons/v1alpha1"
	exerr "github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"github.com/sirupsen/logrus"
)

type brokerFacade interface {
	Create(ns string) error
	Exist(ns string) (bool, error)
	Delete(ns string) error
}

//
type AddonsConfigurationController struct {
	reconciler reconcile.Reconciler
}

//
func NewAddonsConfigurationController(reconciler reconcile.Reconciler) *AddonsConfigurationController {
	return &AddonsConfigurationController{reconciler: reconciler}
}

//
func (acc *AddonsConfigurationController) Start(mgr manager.Manager) error {
	// Create a new controller
	c, err := controller.New("addonsconfiguration-controller", mgr, controller.Options{Reconciler: acc.reconciler})
	if err != nil {
		return err
	}

	// Watch for changes to AddonsConfiguration
	err = c.Watch(&source.Kind{Type: &addonsv1alpha1.AddonsConfiguration{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileAddonsConfiguration{}

// ReconcileAddonsConfiguration reconciles a AddonsConfiguration object
type ReconcileAddonsConfiguration struct {
	log logrus.FieldLogger
	client.Client
	scheme *runtime.Scheme
	strg   storage.Factory
	brokerFacade brokerFacade
}

// newReconciler returns a new reconcile.Reconciler
func NewReconcileAddonsConfiguration(mgr manager.Manager, brokerFacade brokerFacade, s storage.Factory) reconcile.Reconciler {
	return &ReconcileAddonsConfiguration{
		log: logrus.WithField("controller", "addons-configuration"),
		Client: mgr.GetClient(),
		scheme: mgr.GetScheme(),
		strg:   s,

		brokerFacade: brokerFacade,
	}
}

// Reconcile reads that state of the cluster for a AddonsConfiguration object and makes changes based on the state read
// and what is in the AddonsConfiguration.Spec
func (r *ReconcileAddonsConfiguration) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Fetch the AddonsConfiguration instance
	instance := &addonsv1alpha1.AddonsConfiguration{}
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

	foundAddon := &addonsv1alpha1.AddonsConfiguration{}
	err = r.Get(context.TODO(), types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, foundAddon)
	if err != nil {
		return reconcile.Result{}, err
	}

	if foundAddon.Status.ObservedGeneration == 0 {
		err = r.addAddonsProcess(instance)
		if err != nil {
			return reconcile.Result{}, exerr.Wrapf(err, "while creating AddonsConfiguration %q", request.NamespacedName)
		}
	}

	exist, err := r.brokerFacade.Exist(instance.Namespace)
	if err != nil {
		return reconcile.Result{}, exerr.Wrap(err, "while checking if ServiceBroker exists")
	}
	if !exist {
		// status
		if err := r.brokerFacade.Create(instance.Namespace); err != nil {
			return reconcile.Result{}, exerr.Wrapf(err, "while creating ServiceBroker for addon %s in namespace %s", instance.Name, instance.Namespace)
		}
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileAddonsConfiguration) addAddonsProcess(addon *addonsv1alpha1.AddonsConfiguration) error {
	r.log.Info("Start add addons process")
	repositories := repository.NewRepositoryCollection()

	for _, specRepository := range addon.Spec.Repositories {
		// TODO: read from config if it is develop mode and inject value to VerifyURL method
		if err := specRepository.VerifyURL(false); err != nil {
			r.log.Error(err, "")
			continue
		}
		r.log.Info(fmt.Sprintf("create addons for %q repository", specRepository.URL))
		repo := repository.NewAddonsRepository(specRepository.URL)
		repo.Ready()
		addons, err := r.createAddons(specRepository.URL)
		if err != nil {
			repo.Failed()
			repo.Repository.Reason = addonsv1alpha1.RepositoryURLFetchingError
			repo.Repository.Message = err.Error()

			r.log.Error(err, fmt.Sprintf("while creating addons for repository from %q", specRepository.URL))
			continue
		}

		repo.Addons = addons
		repositories.AddRepository(repo)
	}

	r.log.Info("check duplicate ID addons alongside repositories")
	repositories.ReviseBundleDuplicationInRepository()

	r.log.Info("check duplicates ID addons in existing addons configuration")
	addonsList := &addonsv1alpha1.AddonsConfigurationList{}
	err := r.Client.List(context.TODO(), &client.ListOptions{Namespace: addon.Namespace}, addonsList)
	if err != nil {
		return exerr.Wrap(err, "while fetching addons configuration list")
	}
	repositories.ReviseBundleDuplicationInStorage(r.filterAddonsConfigurationList(addon.Name, addonsList))

	r.statusSnapshot(addon, repositories)
	err = r.updateAddonStatus(addon)
	if err != nil {
		return exerr.Wrap(err, "while update AddonsConfiguration status")
	}

	return nil
}

func (r *ReconcileAddonsConfiguration) filterAddonsConfigurationList(addonName string, list *addonsv1alpha1.AddonsConfigurationList) *addonsv1alpha1.AddonsConfigurationList {
	addonsList := &addonsv1alpha1.AddonsConfigurationList{}

	for _, existAddon := range list.Items {
		if existAddon.Name != addonName {
			addonsList.Items = append(addonsList.Items, existAddon)
		}
	}

	return addonsList
}

func (r *ReconcileAddonsConfiguration) createAddons(URL string) ([]*repository.AddonController, error) {
	addons := []*repository.AddonController{}
	provider := bundle.NewBundleProvider(bundle.NewHTTPClient(URL), bundle.NewLoader("/tmp"))

	// fetch repository index
	index, err := provider.GetIndex()
	if err != nil {
		return addons, exerr.Wrap(err, "while reading repository index")
	}

	// for each repository entry create addon
	for _, entries := range index.Entries {
		for _, entry := range entries {
			addon := repository.NewAddon(entry.Name, entry.Version)
			addon.Ready()

			completeBundle, err := provider.ProvideBundle(entry)
			if bundle.IsFetchingError(err) {
				addon.Failed()
				addon.SetAddonFailedInfo(addonsv1alpha1.AddonFetchingError, err.Error())
				continue
			}
			if bundle.IsValidationError(err) {
				addon.Failed()
				addon.SetAddonFailedInfo(addonsv1alpha1.AddonValidationError, err.Error())
				continue
			}

			addon.SetID(string(completeBundle.Bundle.ID))
			addon.AddBundle(completeBundle.Bundle)
			addon.AddCharts(completeBundle.Charts)

			addons = append(addons, addon)
		}
	}

	return addons, nil
}

func (r *ReconcileAddonsConfiguration) statusSnapshot(addon *addonsv1alpha1.AddonsConfiguration, repositories *repository.RepositoryCollection) {
	addon.Status.Repositories = nil

	for _, repo := range repositories.Collection() {
		addonsRepository := repo.Repository
		for _, addon := range repo.Addons {
			addonsRepository.Addons = append(addonsRepository.Addons, addon.Addon)
		}
		addon.Status.Repositories = append(addon.Status.Repositories, addonsRepository)
	}

	if repositories.IsReady() {
		addon.Status.Phase = addonsv1alpha1.AddonsConfigurationReady
	} else {
		addon.Status.Phase = addonsv1alpha1.AddonsConfigurationFailed
	}
}

func (r *ReconcileAddonsConfiguration) updateAddonStatus(addon *addonsv1alpha1.AddonsConfiguration) error {
	addon.Status.ObservedGeneration = addon.Status.ObservedGeneration + 1
	addon.Status.LastProcessedTime = &v1.Time{time.Now()}

	err := r.Update(context.TODO(), addon)
	if err != nil {
		return exerr.Wrap(err, "while update AddonsConfiguration")
	}

	return nil
}
