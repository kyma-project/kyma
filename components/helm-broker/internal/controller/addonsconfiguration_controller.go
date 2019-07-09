package controller

import (
	"context"
	"time"

	"github.com/kyma-project/kyma/components/helm-broker/internal/bundle"
	"github.com/kyma-project/kyma/components/helm-broker/internal/controller/repository"
	"github.com/kyma-project/kyma/components/helm-broker/internal/storage"
	addonsv1alpha1 "github.com/kyma-project/kyma/components/helm-broker/pkg/apis/addons/v1alpha1"
	exerr "github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"github.com/kyma-project/kyma/components/helm-broker/internal"
)

type brokerFacade interface {
	Create(ns string) error
	Exist(ns string) (bool, error)
	Delete(ns string) error
}

type bundleProvider interface {
	GetIndex(string) (*bundle.IndexDTO, error)
	LoadCompleteBundle(bundle.EntryDTO) (bundle.CompleteBundle, error)
}

type docsProvider interface {
	EnsureDocsTopic(bundle *internal.Bundle) error
	EnsureDocsTopicRemoved(id string) error
}

type brokerSyncer interface {
	SyncServiceBroker(namespace string) error
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
	log               logrus.FieldLogger
	client.Client
	scheme            *runtime.Scheme
	provider          bundleProvider
	strg              storage.Factory
	brokerFacade      brokerFacade
	brokerSyncer      brokerSyncer
	docsTopicProvider docsProvider
	developMode       bool

	// syncBroker informs ServiceBroker should be resync, it should be true if
	// operation insert/delete was made on storage
	syncBroker bool
}

// newReconciler returns a new reconcile.Reconciler
func NewReconcileAddonsConfiguration(mgr manager.Manager, bp bundleProvider, brokerFacade brokerFacade, s storage.Factory, dev bool, docsTopicProvider docsProvider, brokerSyncer brokerSyncer) reconcile.Reconciler {
	return &ReconcileAddonsConfiguration{
		log:      logrus.WithField("controller", "addons-configuration"),
		Client:   mgr.GetClient(),
		scheme:   mgr.GetScheme(),
		strg:     s,
		provider: bp,

		brokerSyncer:      brokerSyncer,
		brokerFacade:      brokerFacade,
		developMode:       dev,
		docsTopicProvider: docsTopicProvider,
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
			if err := r.deleteAddonsProcess(request.Namespace); err != nil {
				return reconcile.Result{}, exerr.Wrapf(err, "while deleting Addon Configuration from namespace %s", request.Namespace)
			}
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
		return reconcile.Result{}, exerr.Wrapf(err, "while checking if ServiceBroker exist in namespace %s", instance.Namespace)
	}
	if !exist {
		// status
		if err := r.brokerFacade.Create(instance.Namespace); err != nil {
			return reconcile.Result{}, exerr.Wrapf(err, "while creating ServiceBroker for addon %s in namespace %s", instance.Name, instance.Namespace)
		}
	} else if r.syncBroker {
		if err := r.brokerSyncer.SyncServiceBroker(instance.Namespace); err != nil {
			return reconcile.Result{}, exerr.Wrapf(err, "while syncing ServiceBroker for addon %s in namespace %s", instance.Name, instance.Namespace)
		}
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileAddonsConfiguration) addAddonsProcess(addon *addonsv1alpha1.AddonsConfiguration) error {
	r.log.Info("Start add AddonsConfiguration process")
	repositories := repository.NewRepositoryCollection()

	r.log.Infof("- load bundles and charts for each addon")
	for _, specRepository := range addon.Spec.Repositories {
		if err := specRepository.VerifyURL(r.developMode); err != nil {
			r.log.Errorf("url %q address is not valid: %s", specRepository.URL, err)
			continue
		}
		r.log.Infof("create addons for %q repository", specRepository.URL)
		repo := repository.NewAddonsRepository(specRepository.URL)

		addons, err := r.createAddons(specRepository.URL)
		if err != nil {
			repo.Failed()
			repo.Repository.Reason = addonsv1alpha1.RepositoryURLFetchingError
			repo.Repository.Message = err.Error()
			repositories.AddRepository(repo)

			r.log.Errorf("while creating addons for repository from %q: %s", specRepository.URL, err)
			continue
		}

		repo.Addons = addons
		repositories.AddRepository(repo)
	}

	r.log.Info("- check duplicate ID addons alongside repositories")
	repositories.ReviseBundleDuplicationInRepository()

	r.log.Info("- check duplicates ID addons in existing addons configuration")
	list, err := r.existingAddonsConfigurationList(addon)
	if err != nil {
		r.log.Errorf("cannot fetch AddonsConfiguration list: %s", err)
		return exerr.Wrap(err, "while fetching addons configuration list")
	}
	repositories.ReviseBundleDuplicationInStorage(list)

	r.log.Info("- save ready bundles and charts in storage")
	r.saveBundle(internal.Namespace(addon.Namespace), repositories)

	r.log.Info("- update AddonsConfiguration status")
	r.statusSnapshot(addon, repositories)
	err = r.updateAddonStatus(addon)
	if err != nil {
		r.log.Errorf("cannot update AddonsConfiguration %q: %s", addon.Name, err)
		return exerr.Wrap(err, "while update AddonsConfiguration status")
	}

	r.log.Info("Add AddonsConfiguration process completed")
	return nil
}

func (r *ReconcileAddonsConfiguration) deleteAddonsProcess(namespace string) error {
	r.log.Infof("Start delete AddonsConfiguration from namespace %s", namespace)

	addonsCfgs, err := r.addonsConfigurationList(namespace)
	if err != nil {
		return exerr.Wrapf(err, "while listing AddonsConfigurations in namespace %s", namespace)
	}

	deleteBroker := true
	for _, addon := range addonsCfgs.Items {
		if addon.Status.Phase != addonsv1alpha1.AddonsConfigurationReady {
			// reprocess AddonConfig again if it was failed
			addon.Spec.ReprocessRequest += 1
			if err := r.Client.Update(context.Background(), &addon); err != nil {
				return exerr.Wrapf(err, "while incrementing a reprocess requests for AddonConfiguration %s/%s", addon.Name, addon.Namespace)
			}
		} else {
			deleteBroker = false
		}
	}
	if deleteBroker {
		if err := r.brokerFacade.Delete(namespace); err != nil {
			return exerr.Wrapf(err, "while deleting ServiceBroker from namespace %s", namespace)
		}
	}

	r.log.Info("Delete AddonsConfiguration process completed")
	return nil
}

func (r *ReconcileAddonsConfiguration) existingAddonsConfigurationList(addon *addonsv1alpha1.AddonsConfiguration) (*addonsv1alpha1.AddonsConfigurationList, error) {
	addonsList := &addonsv1alpha1.AddonsConfigurationList{}
	addonsConfigurationList, err := r.addonsConfigurationList(addon.Namespace)
	if err != nil {
		return nil, exerr.Wrapf(err, "while listing AddonsConfigurations from namespace %s", addon.Namespace)
	}

	for _, existAddon := range addonsConfigurationList.Items {
		if existAddon.Name != addon.Name {
			addonsList.Items = append(addonsList.Items, existAddon)
		}
	}

	return addonsList, nil
}

func (r *ReconcileAddonsConfiguration) addonsConfigurationList(namespace string) (*addonsv1alpha1.AddonsConfigurationList, error) {
	addonsList := &addonsv1alpha1.AddonsConfigurationList{}
	addonsConfigurationList := &addonsv1alpha1.AddonsConfigurationList{}

	err := r.Client.List(context.TODO(), &client.ListOptions{Namespace: namespace}, addonsConfigurationList)
	if err != nil {
		return addonsList, exerr.Wrap(err, "during fetching AddonConfiguration list by client")
	}

	return addonsList, nil
}

func (r *ReconcileAddonsConfiguration) createAddons(URL string) ([]*repository.AddonController, error) {
	addons := []*repository.AddonController{}

	// fetch repository index
	index, err := r.provider.GetIndex(URL)
	if err != nil {
		return addons, exerr.Wrap(err, "while reading repository index")
	}

	// for each repository entry create addon
	for _, entries := range index.Entries {
		for _, entry := range entries {
			addon := repository.NewAddon(string(entry.Name), string(entry.Version), URL)

			completeBundle, err := r.provider.LoadCompleteBundle(entry)
			if bundle.IsFetchingError(err) {
				addon.FetchingError(err)
				addons = append(addons, addon)
				logrus.Errorf("while fetching addon: %s", err)
				continue
			}
			if bundle.IsLoadingError(err) {
				addon.LoadingError(err)
				addons = append(addons, addon)
				logrus.Errorf("while loading addon: %s", err)
				continue
			}

			addon.ID = string(completeBundle.Bundle.ID)
			addon.Bundle = completeBundle.Bundle
			addon.Charts = completeBundle.Charts

			addons = append(addons, addon)
		}
	}

	return addons, nil
}

func (r *ReconcileAddonsConfiguration) saveBundle(namespace internal.Namespace, repositories *repository.RepositoryCollection) {
	for _, addon := range repositories.ReadyAddons() {
		exist, err := r.strg.Bundle().Upsert(namespace, addon.Bundle)
		if err != nil {
			addon.RegisteringError(err)
			r.log.Errorf("cannot upsert bundle %s:%s into storage", addon.Bundle.Name, addon.Bundle.Version)
			continue
		}
		if exist {
			r.log.Infof("bundle %s:%s already existed in storage, bundle was replaced", addon.Bundle.Name, addon.Bundle.Version)
		}
		err = r.saveCharts(namespace, addon.Charts)
		if err != nil {
			addon.RegisteringError(err)
			r.log.Errorf("cannot upsert charts of %s:%s bunlde", addon.Bundle.Name, addon.Bundle.Version)
			continue
		}

		r.syncBroker = true
	}
}

func (r *ReconcileAddonsConfiguration) saveCharts(namespace internal.Namespace, charts []*chart.Chart) error {
	for _, bundleChart := range charts {
		exist, err := r.strg.Chart().Upsert(namespace, bundleChart)
		if err != nil {
			r.log.Errorf("cannot upsert %s chart: %s", bundleChart.Metadata.Name, err)
			return err
		}
		if exist {
			r.log.Infof("chart %s already existed in storage, chart was replaced", bundleChart.Metadata.Name)
		}
	}

	return nil
}

func (r *ReconcileAddonsConfiguration) statusSnapshot(addon *addonsv1alpha1.AddonsConfiguration, repositories *repository.RepositoryCollection) {
	addon.Status.Repositories = nil

	for _, repo := range repositories.Repositories {
		addonsRepository := repo.Repository
		addonsRepository.Addons = []addonsv1alpha1.Addon{}
		for _, addon := range repo.Addons {
			addonsRepository.Addons = append(addonsRepository.Addons, addon.Addon)
		}
		addon.Status.Repositories = append(addon.Status.Repositories, addonsRepository)
	}

	if repositories.IsRepositoriesIdConflict() {
		addon.Status.Phase = addonsv1alpha1.AddonsConfigurationFailed
	} else {
		addon.Status.Phase = addonsv1alpha1.AddonsConfigurationReady
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
