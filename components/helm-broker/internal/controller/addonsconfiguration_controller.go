package controller

import (
	"context"

	"time"

	"github.com/Masterminds/semver"
	"github.com/kyma-project/kyma/components/helm-broker/internal"
	"github.com/kyma-project/kyma/components/helm-broker/internal/bundle"
	"github.com/kyma-project/kyma/components/helm-broker/internal/controller/addons"
	"github.com/kyma-project/kyma/components/helm-broker/internal/storage"
	addonsv1alpha1 "github.com/kyma-project/kyma/components/helm-broker/pkg/apis/addons/v1alpha1"
	exerr "github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// AddonsConfigurationController holds a controller logic
type AddonsConfigurationController struct {
	reconciler reconcile.Reconciler
}

// NewAddonsConfigurationController creates a controller with a given reconciler
func NewAddonsConfigurationController(reconciler reconcile.Reconciler) *AddonsConfigurationController {
	return &AddonsConfigurationController{reconciler: reconciler}
}

// Start starts a controller
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

	chartStorage  chartStorage
	bundleStorage bundleStorage

	docsProvider docsProvider
	brokerFacade brokerFacade
	brokerSyncer brokerSyncer

	bundleProvider bundleProvider
	protection     protection

	// syncBroker informs ServiceBroker should be resync, it should be true if
	// operation insert/delete was made on storage
	syncBroker  bool
	developMode bool
}

// NewReconcileAddonsConfiguration returns a new reconcile.Reconciler
func NewReconcileAddonsConfiguration(mgr manager.Manager, bp bundleProvider, chartStorage chartStorage, bundleStorage bundleStorage, brokerFacade brokerFacade, docsProvider docsProvider, brokerSyncer brokerSyncer, developMode bool) reconcile.Reconciler {
	return &ReconcileAddonsConfiguration{
		log:    logrus.WithField("controller", "addons-configuration"),
		Client: mgr.GetClient(),
		scheme: mgr.GetScheme(),

		chartStorage:  chartStorage,
		bundleStorage: bundleStorage,

		bundleProvider: bp,
		protection:     protection{},

		brokerSyncer: brokerSyncer,
		brokerFacade: brokerFacade,
		docsProvider: docsProvider,

		developMode: developMode,
		syncBroker:  false,
	}
}

// Reconcile reads that state of the cluster for a AddonsConfiguration object and makes changes based on the state read
// and what is in the AddonsConfiguration.Spec
func (r *ReconcileAddonsConfiguration) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	addon := &addonsv1alpha1.AddonsConfiguration{}
	err := r.Get(context.TODO(), request.NamespacedName, addon)
	if err != nil {
		return reconcile.Result{}, err
	}
	r.syncBroker = false

	if addon.DeletionTimestamp != nil {
		if err := r.deleteAddonsProcess(addon); err != nil {
			r.log.Errorf("while deleting AddonsConfiguration process: %v", err)
			return reconcile.Result{RequeueAfter: time.Second * 15}, exerr.Wrapf(err, "while deleting AddonConfiguration %q", request.NamespacedName)
		}
		return reconcile.Result{}, nil
	}

	if addon.Status.ObservedGeneration == 0 {
		r.log.Infof("Start add AddonsConfiguration %s/%s process", addon.Name, addon.Namespace)

		preAddon, err := r.prepareForProcessing(addon)
		if err != nil {
			r.log.Errorf("while preparing for processing: %v", err)
			return reconcile.Result{Requeue: true}, exerr.Wrapf(err, "while adding a finalizer to AddonsConfiguration %q", request.NamespacedName)
		}
		err = r.addAddonsProcess(preAddon, preAddon.Status)
		if err != nil {
			r.log.Errorf("while adding AddonsConfiguration process: %v", err)
			return reconcile.Result{}, exerr.Wrapf(err, "while creating ClusterAddonsConfiguration %q", request.NamespacedName)
		}
		r.log.Info("Add AddonsConfiguration process completed")

	} else if addon.Generation > addon.Status.ObservedGeneration {
		r.log.Infof("Start update AddonsConfiguration %s/%s process", addon.Name, addon.Namespace)

		lastAddon := addon.DeepCopy()
		addon.Status = addonsv1alpha1.AddonsConfigurationStatus{}
		err = r.addAddonsProcess(addon, lastAddon.Status)
		if err != nil {
			r.log.Errorf("while updating AddonsConfiguration process: %v", err)
			return reconcile.Result{}, exerr.Wrapf(err, "while updating AddonsConfiguration %q", request.NamespacedName)
		}
		r.log.Info("Update AddonsConfiguration process completed")
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileAddonsConfiguration) addAddonsProcess(addon *addonsv1alpha1.AddonsConfiguration, lastStatus addonsv1alpha1.AddonsConfigurationStatus) error {
	r.log.Infof("- load bundles and charts for each addon")
	repositories := r.loadAddons(addon)

	r.log.Info("- check duplicate ID addons alongside repositories")
	repositories.ReviseBundleDuplicationInRepository()

	r.log.Info("- check duplicates ID addons in existing AddonsConfiguration")
	list, err := r.existingAddonsConfigurations(addon)
	if err != nil {
		return exerr.Wrap(err, "while fetching AddonsConfiguration list")
	}
	repositories.ReviseBundleDuplicationInStorage(list)

	if repositories.IsRepositoriesFailed() {
		addon.Status.Phase = addonsv1alpha1.AddonsConfigurationFailed
	} else {
		addon.Status.Phase = addonsv1alpha1.AddonsConfigurationReady
	}
	r.log.Infof("- status: %s", addon.Status.Phase)

	var deletedBundles []string

	switch addon.Status.Phase {
	case addonsv1alpha1.AddonsConfigurationFailed:
		if _, err = r.updateAddonStatus(r.statusSnapshot(addon, repositories)); err != nil {
			return exerr.Wrap(err, "while updating AddonsConfiguration status")
		}
		if lastStatus.Phase == addonsv1alpha1.AddonsConfigurationReady {
			deletedBundles, err = r.deleteBundlesFromRepository(addon.Namespace, lastStatus.Repositories)
			if err != nil {
				return exerr.Wrap(err, "while deleting bundles from repository")
			}
		}
	case addonsv1alpha1.AddonsConfigurationReady:
		r.log.Info("- save ready bundles and charts in storage")
		if err := r.saveBundle(internal.Namespace(addon.Namespace), repositories); err != nil {
			return exerr.Wrap(err, "while saving ready bundles and charts in storage")
		}
		if _, err = r.updateAddonStatus(r.statusSnapshot(addon, repositories)); err != nil {
			return exerr.Wrap(err, "while updating AddonsConfiguration status")
		}
		if lastStatus.Phase == addonsv1alpha1.AddonsConfigurationReady {
			deletedBundles, err = r.deleteOrphanBundles(addon.Namespace, addon.Status.Repositories, lastStatus.Repositories)
			if err != nil {
				return exerr.Wrap(err, "while deleting orphan bundles from storage")
			}
		}

	}

	if r.syncBroker {
		r.log.Info("- ensure ServiceBroker")
		if err = r.ensureBroker(addon); err != nil {
			return exerr.Wrap(err, "while ensuring ServiceBroker")
		}
	}

	if len(deletedBundles) > 0 {
		r.log.Info("- reprocessing conflicting addons configurations")
		for _, key := range deletedBundles {
			// reprocess ClusterAddonsConfiguration again if it contains a conflicting addons
			if err := r.reprocessConflictingAddonsConfiguration(key, list); err != nil {
				return exerr.Wrap(err, "while requesting processing of conflicting AddonsConfigurations")
			}
		}
	}
	return nil
}

func (r *ReconcileAddonsConfiguration) deleteAddonsProcess(addon *addonsv1alpha1.AddonsConfiguration) error {
	r.log.Infof("Start delete AddonsConfiguration %s/%s process", addon.Name, addon.Namespace)

	if addon.Status.Phase == addonsv1alpha1.AddonsConfigurationReady {
		adds, err := r.existingAddonsConfigurations(addon)
		if err != nil {
			return exerr.Wrapf(err, "while listing AddonsConfigurations in namespace %s", addon.Namespace)
		}

		deleteBroker := true
		for _, addon := range adds.Items {
			if addon.Status.Phase != addonsv1alpha1.AddonsConfigurationReady {
				// reprocess AddonConfig again if it was failed
				if err := r.reprocessAddonsConfiguration(&addon); err != nil {
					return exerr.Wrapf(err, "while requesting reprocess for AddonsConfiguration %s", addon.Name)

				}
			} else {
				deleteBroker = false
			}
		}
		if deleteBroker {
			r.log.Info("- delete ServiceBroker from namespace %s", addon.Namespace)
			if err := r.brokerFacade.Delete(addon.Namespace); err != nil {
				return exerr.Wrapf(err, "while deleting ServiceBroker from namespace %s", addon.Namespace)
			}
		}

		for _, repo := range addon.Status.Repositories {
			for _, add := range repo.Addons {
				id, err := r.removeBundle(add, internal.Namespace(addon.Namespace))
				if err != nil && !storage.IsNotFoundError(err) {
					return exerr.Wrapf(err, "while deleting bundle with charts for addon %s", add.Name)
				}

				if id != nil {
					r.log.Infof("- delete DocsTopic for bundle %s", add)
					if err := r.docsProvider.EnsureDocsTopicRemoved(string(*id), addon.Namespace); err != nil {
						return exerr.Wrapf(err, "while ensuring DocsTopic for bundle %s is removed", *id)
					}
				}
			}
		}
		if !deleteBroker && r.syncBroker {
			if err := r.brokerSyncer.SyncServiceBroker(addon.Namespace); err != nil {
				return exerr.Wrapf(err, "while syncing ClusterServiceBroker for addon %s", addon.Name)
			}
		}
	}
	if err := r.deleteFinalizer(addon); err != nil {
		return exerr.Wrapf(err, "while deleting finalizer for AddonConfiguration %s/%s", addon.Name, addon.Namespace)
	}

	r.log.Info("Delete AddonsConfiguration process completed")
	return nil
}

func (r *ReconcileAddonsConfiguration) loadAddons(addon *addonsv1alpha1.AddonsConfiguration) *addons.RepositoryCollection {
	repositories := addons.NewRepositoryCollection()
	for _, specRepository := range addon.Spec.Repositories {
		r.log.Infof("- create addons for %q repository", specRepository.URL)
		repo := addons.NewAddonsRepository(specRepository.URL)

		if err := specRepository.VerifyURL(r.developMode); err != nil {
			repo.FetchingError(err)
			repositories.AddRepository(repo)

			r.log.Errorf("url %q address is not valid: %s", specRepository.URL, err)
			continue
		}

		adds, err := r.createAddons(specRepository.URL)
		if err != nil {
			repo.FetchingError(err)
			repositories.AddRepository(repo)

			r.log.Errorf("while creating addons for repository from %q: %s", specRepository.URL, err)
			continue
		}

		repo.Addons = adds
		repositories.AddRepository(repo)
	}
	return repositories
}

func (r *ReconcileAddonsConfiguration) ensureBroker(addon *addonsv1alpha1.AddonsConfiguration) error {
	exist, err := r.brokerFacade.Exist(addon.Namespace)
	if err != nil {
		return exerr.Wrapf(err, "while checking if ServiceBroker exist in namespace %s", addon.Namespace)
	}
	if !exist {
		r.log.Infof("- creating ServiceBroker in namespace %s", addon.Namespace)
		if err := r.brokerFacade.Create(addon.Namespace); err != nil {
			return exerr.Wrapf(err, "while creating ServiceBroker for AddonConfiguration %s/%s", addon.Name, addon.Namespace)
		}
	} else {
		if err := r.brokerSyncer.SyncServiceBroker(addon.Namespace); err != nil {
			return exerr.Wrapf(err, "while syncing ServiceBroker for AddonConfiguration %s/%s", addon.Name, addon.Namespace)
		}
	}
	return nil
}

func (r *ReconcileAddonsConfiguration) createAddons(URL string) ([]*addons.AddonController, error) {
	adds := []*addons.AddonController{}

	// fetch repository index
	index, err := r.bundleProvider.GetIndex(URL)
	if err != nil {
		return adds, exerr.Wrap(err, "while reading repository index")
	}

	// for each repository entry create addon
	for _, entries := range index.Entries {
		for _, entry := range entries {
			addon := addons.NewAddon(string(entry.Name), string(entry.Version), URL)

			completeBundle, err := r.bundleProvider.LoadCompleteBundle(entry)
			if bundle.IsFetchingError(err) {
				addon.FetchingError(err)
				adds = append(adds, addon)
				r.log.Errorf("while fetching addon: %s", err)
				continue
			}
			if bundle.IsLoadingError(err) {
				addon.LoadingError(err)
				adds = append(adds, addon)
				r.log.Errorf("while loading addon: %s", err)
				continue
			}

			addon.ID = string(completeBundle.Bundle.ID)
			addon.Bundle = completeBundle.Bundle
			addon.Charts = completeBundle.Charts

			adds = append(adds, addon)
		}
	}

	return adds, nil
}

func (r *ReconcileAddonsConfiguration) existingAddonsConfigurations(addon *addonsv1alpha1.AddonsConfiguration) (*addonsv1alpha1.AddonsConfigurationList, error) {
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
	addonsConfigurationList := &addonsv1alpha1.AddonsConfigurationList{}

	err := r.Client.List(context.TODO(), &client.ListOptions{Namespace: namespace}, addonsConfigurationList)
	if err != nil {
		return addonsConfigurationList, exerr.Wrap(err, "during fetching AddonConfiguration list by client")
	}

	return addonsConfigurationList, nil
}

func (r *ReconcileAddonsConfiguration) deleteOrphanBundles(namespace string, repos []addonsv1alpha1.StatusRepository, lastRepos []addonsv1alpha1.StatusRepository) ([]string, error) {
	addonsToStay := map[string]addonsv1alpha1.Addon{}
	for _, repo := range repos {
		for _, ad := range repo.Addons {
			addonsToStay[ad.Key()] = ad
		}
	}
	var deletedBundlesIDs []string
	for _, repo := range lastRepos {
		for _, ad := range repo.Addons {
			if _, exist := addonsToStay[ad.Key()]; !exist {
				if _, err := r.removeBundle(ad, internal.Namespace(namespace)); err != nil && !storage.IsNotFoundError(err) {
					return nil, exerr.Wrapf(err, "while deleting bundles and charts for addon %s", ad.Name)
				}
				deletedBundlesIDs = append(deletedBundlesIDs, ad.Key())
			}
		}
	}
	return deletedBundlesIDs, nil
}

func (r *ReconcileAddonsConfiguration) deleteBundlesFromRepository(namespace string, repos []addonsv1alpha1.StatusRepository) ([]string, error) {
	var deletedBundlesKeys []string
	for _, repo := range repos {
		for _, ad := range repo.Addons {
			if _, err := r.removeBundle(ad, internal.Namespace(namespace)); err != nil && !storage.IsNotFoundError(err) {
				return nil, exerr.Wrapf(err, "while deleting bundles and charts for addon %s", ad.Name)
			}
			deletedBundlesKeys = append(deletedBundlesKeys, ad.Key())
		}
	}
	return deletedBundlesKeys, nil
}

func (r *ReconcileAddonsConfiguration) removeBundle(ad addonsv1alpha1.Addon, namespace internal.Namespace) (*internal.BundleID, error) {
	r.log.Infof("- delete bundle %s from storage", ad.Name)
	b, err := r.bundleStorage.Get(namespace, internal.BundleName(ad.Name), *semver.MustParse(ad.Version))
	if err != nil {
		return nil, err
	}

	err = r.bundleStorage.Remove(namespace, internal.BundleName(ad.Name), *semver.MustParse(ad.Version))
	if err != nil {
		return nil, err
	}

	for _, plan := range b.Plans {
		err = r.chartStorage.Remove(namespace, plan.ChartRef.Name, plan.ChartRef.Version)
		if err != nil {
			return nil, err
		}
	}
	r.syncBroker = true
	return &b.ID, nil
}

func (r *ReconcileAddonsConfiguration) reprocessConflictingAddonsConfiguration(key string, list *addonsv1alpha1.AddonsConfigurationList) error {
	for _, addonsCfg := range list.Items {
		if addonsCfg.Status.Phase != addonsv1alpha1.AddonsConfigurationReady {
			for _, repo := range addonsCfg.Status.Repositories {
				if repo.Status != addonsv1alpha1.RepositoryStatusReady {
					for _, add := range repo.Addons {
						if add.Key() == key {
							return r.reprocessAddonsConfiguration(&addonsCfg)
						}
					}
				}
			}
		}
	}
	return nil
}

func (r *ReconcileAddonsConfiguration) reprocessAddonsConfiguration(addon *addonsv1alpha1.AddonsConfiguration) error {
	ad := &addonsv1alpha1.AddonsConfiguration{}
	if err := r.Client.Get(context.Background(), types.NamespacedName{Name: addon.Name, Namespace: addon.Namespace}, ad); err != nil {
		return exerr.Wrapf(err, "while getting ClusterAddonsConfiguration %s", addon.Name)
	}
	ad.Spec.ReprocessRequest++
	if err := r.Client.Update(context.Background(), ad); err != nil {
		return exerr.Wrapf(err, "while incrementing a reprocess requests for ClusterAddonsConfiguration %s", addon.Name)
	}
	return nil
}

func (r *ReconcileAddonsConfiguration) saveBundle(namespace internal.Namespace, repositories *addons.RepositoryCollection) error {
	for _, addon := range repositories.ReadyAddons() {
		if len(addon.Bundle.Docs) == 1 {
			r.log.Infof("- ensure DocsTopic for bundle %s in namespace %s", addon.Bundle.ID, namespace)
			if err := r.docsProvider.EnsureDocsTopic(addon.Bundle, string(namespace)); err != nil {
				return exerr.Wrapf(err, "While ensuring DocsTopic for bundle %s/%s: %v", addon.Bundle.ID, namespace, err)
			}
		}
		exist, err := r.bundleStorage.Upsert(namespace, addon.Bundle)
		if err != nil {
			addon.RegisteringError(err)
			r.log.Errorf("cannot upsert bundle %v:%v into storage", addon.Bundle.Name, addon.Bundle.Version)
			continue
		}
		if exist {
			r.log.Infof("bundle %v:%v already existed in storage, bundle was replaced", addon.Bundle.Name, addon.Bundle.Version)
		}
		err = r.saveCharts(namespace, addon.Charts)
		if err != nil {
			addon.RegisteringError(err)
			r.log.Errorf("cannot upsert charts of %v:%v bunlde", addon.Bundle.Name, addon.Bundle.Version)
			continue
		}

		r.syncBroker = true
	}
	return nil
}

func (r *ReconcileAddonsConfiguration) saveCharts(namespace internal.Namespace, charts []*chart.Chart) error {
	for _, bundleChart := range charts {
		exist, err := r.chartStorage.Upsert(namespace, bundleChart)
		if err != nil {
			return err
		}
		if exist {
			r.log.Infof("chart %s already existed in storage, chart was replaced", bundleChart.Metadata.Name)
		}
	}
	return nil
}

func (r *ReconcileAddonsConfiguration) statusSnapshot(addon *addonsv1alpha1.AddonsConfiguration, repositories *addons.RepositoryCollection) *addonsv1alpha1.AddonsConfiguration {
	addon.Status.Repositories = nil

	for _, repo := range repositories.Repositories {
		addonsRepository := repo.Repository
		addonsRepository.Addons = []addonsv1alpha1.Addon{}
		for _, addon := range repo.Addons {
			addonsRepository.Addons = append(addonsRepository.Addons, addon.Addon)
		}
		addon.Status.Repositories = append(addon.Status.Repositories, addonsRepository)
	}

	if repositories.IsRepositoriesFailed() {
		addon.Status.Phase = addonsv1alpha1.AddonsConfigurationFailed
	} else {
		addon.Status.Phase = addonsv1alpha1.AddonsConfigurationReady
	}

	return addon
}

func (r *ReconcileAddonsConfiguration) updateAddonStatus(addon *addonsv1alpha1.AddonsConfiguration) (*addonsv1alpha1.AddonsConfiguration, error) {
	addon.Status.ObservedGeneration = addon.Generation
	addon.Status.LastProcessedTime = &v1.Time{Time: time.Now()}

	r.log.Infof("- update AddonsConfiguration %s/%s status", addon.Name, addon.Namespace)
	err := r.Status().Update(context.TODO(), addon)
	if err != nil {
		return nil, exerr.Wrap(err, "while update AddonsConfiguration status")
	}
	return addon, nil
}

func (r *ReconcileAddonsConfiguration) prepareForProcessing(addon *addonsv1alpha1.AddonsConfiguration) (*addonsv1alpha1.AddonsConfiguration, error) {
	obj := addon.DeepCopy()
	obj.Status.Phase = addonsv1alpha1.AddonsConfigurationPending

	pendingInstance, err := r.updateAddonStatus(obj)
	if err != nil {
		return nil, err
	}
	if r.protection.hasFinalizer(pendingInstance.Finalizers) {
		return pendingInstance, nil
	}
	r.log.Info("- add a finalizer")
	pendingInstance.Finalizers = r.protection.addFinalizer(pendingInstance.Finalizers)

	err = r.Client.Update(context.Background(), pendingInstance)
	if err != nil {
		return nil, err
	}
	return pendingInstance, nil
}

func (r *ReconcileAddonsConfiguration) deleteFinalizer(addon *addonsv1alpha1.AddonsConfiguration) error {
	obj := addon.DeepCopy()
	if !r.protection.hasFinalizer(obj.Finalizers) {
		return nil
	}
	r.log.Info("- delete a finalizer")
	obj.Finalizers = r.protection.removeFinalizer(obj.Finalizers)

	return r.Client.Update(context.Background(), obj)
}
