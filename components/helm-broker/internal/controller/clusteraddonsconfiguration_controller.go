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

	chartStorage  chartStorage
	bundleStorage bundleStorage

	docsProvider clusterDocsProvider
	brokerFacade clusterBrokerFacade
	brokerSyncer clusterBrokerSyncer

	bundleProvider bundleProvider
	protection     protection

	// syncBroker informs ServiceBroker should be resync, it should be true if
	// operation insert/delete was made on storage
	syncBroker  bool
	developMode bool
}

// NewReconcileClusterAddonsConfiguration returns a new reconcile.Reconciler
func NewReconcileClusterAddonsConfiguration(mgr manager.Manager, bundleProvider bundleProvider, chartStorage chartStorage, bundleStorage bundleStorage, brokerFacade clusterBrokerFacade, docsProvider clusterDocsProvider, brokerSyncer clusterBrokerSyncer, developMode bool) reconcile.Reconciler {
	return &ReconcileClusterAddonsConfiguration{
		log:    logrus.WithField("controller", "cluster-addons-configuration"),
		Client: mgr.GetClient(),
		scheme: mgr.GetScheme(),

		bundleStorage: bundleStorage,
		chartStorage:  chartStorage,

		brokerFacade:   brokerFacade,
		docsProvider:   docsProvider,
		brokerSyncer:   brokerSyncer,
		bundleProvider: bundleProvider,

		protection: protection{},

		developMode: developMode,
		syncBroker:  false,
	}
}

// Reconcile reads that state of the cluster for a ClusterAddonsConfiguration object and makes changes based on the state read
// and what is in the ClusterAddonsConfiguration.Spec
func (r *ReconcileClusterAddonsConfiguration) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	addon := &addonsv1alpha1.ClusterAddonsConfiguration{}
	err := r.Get(context.TODO(), request.NamespacedName, addon)
	if err != nil {
		return reconcile.Result{}, err
	}
	r.syncBroker = false

	if addon.DeletionTimestamp != nil {
		if err := r.deleteAddonsProcess(addon); err != nil {
			r.log.Errorf("while deleting ClusterAddonsConfiguration process: %v", err)
			return reconcile.Result{RequeueAfter: time.Second * 15}, exerr.Wrapf(err, "while deleting ClusterAddonConfiguration %q", request.NamespacedName)
		}
		return reconcile.Result{}, nil
	}

	if addon.Status.ObservedGeneration == 0 {
		r.log.Infof("Start add ClusterAddonsConfiguration %s process", addon.Name)

		preAddon, err := r.prepareForProcessing(addon)
		if err != nil {
			r.log.Errorf("while preparing for processing: %v", err)
			return reconcile.Result{Requeue: true}, exerr.Wrapf(err, "while adding a finalizer to AddonsConfiguration %q", request.NamespacedName)
		}
		err = r.addAddonsProcess(preAddon, preAddon.Status)
		if err != nil {
			r.log.Errorf("while adding ClusterAddonsConfiguration process: %v", err)
			return reconcile.Result{}, exerr.Wrapf(err, "while creating ClusterAddonsConfiguration %q", request.NamespacedName)
		}
		r.log.Infof("Add ClusterAddonsConfiguration process completed")

	} else if addon.Generation > addon.Status.ObservedGeneration {
		r.log.Infof("Start update ClusterAddonsConfiguration %s process", addon.Name)

		lastAddon := addon.DeepCopy()
		addon.Status = addonsv1alpha1.ClusterAddonsConfigurationStatus{}
		err = r.addAddonsProcess(addon, lastAddon.Status)
		if err != nil {
			r.log.Errorf("while updating ClusterAddonsConfiguration process: %v", err)
			return reconcile.Result{}, exerr.Wrapf(err, "while updating ClusterAddonsConfiguration %q", request.NamespacedName)
		}
		r.log.Infof("Update ClusterAddonsConfiguration %s process completed", addon.Name)
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileClusterAddonsConfiguration) addAddonsProcess(addon *addonsv1alpha1.ClusterAddonsConfiguration, lastStatus addonsv1alpha1.ClusterAddonsConfigurationStatus) error {
	r.log.Infof("- load bundles and charts for each addon")
	repositories := r.loadAddons(addon)

	r.log.Info("- check duplicate ID addons alongside repositories")
	repositories.ReviseBundleDuplicationInRepository()

	r.log.Info("- check duplicates ID addons in existing ClusterAddonsConfigurations")
	list, err := r.existingAddonsConfigurations(addon.Name)
	if err != nil {
		return exerr.Wrap(err, "while fetching ClusterAddonsConfigurations list")
	}
	repositories.ReviseBundleDuplicationInClusterStorage(list)

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
			return exerr.Wrap(err, "while update ClusterAddonsConfiguration status")
		}
		if lastStatus.Phase == addonsv1alpha1.AddonsConfigurationReady {
			deletedBundles, err = r.deleteBundlesFromRepository(lastStatus.Repositories)
			if err != nil {
				return exerr.Wrap(err, "while deleting bundles from repository")
			}

		}
	case addonsv1alpha1.AddonsConfigurationReady:
		r.log.Info("- save ready bundles and charts in storage")
		if err := r.saveBundle(repositories); err != nil {
			return exerr.Wrap(err, "while saving ready bundles and charts in storage")
		}
		if _, err = r.updateAddonStatus(r.statusSnapshot(addon, repositories)); err != nil {
			return exerr.Wrap(err, "while update ClusterAddonsConfiguration status")
		}
		if lastStatus.Phase == addonsv1alpha1.AddonsConfigurationReady {
			deletedBundles, err = r.deleteOrphanBundles(addon.Status.Repositories, lastStatus.Repositories)
			if err != nil {
				return exerr.Wrap(err, "while deleting orphan bundles from storage")
			}
		}
	}

	if r.syncBroker {
		r.log.Info("- ensure ClusterServiceBroker")
		if err = r.ensureBroker(addon); err != nil {
			return exerr.Wrap(err, "while ensuring ClusterServiceBroker")
		}
	}

	if len(deletedBundles) > 0 {
		r.log.Info("- reprocessing conflicting addons configurations")
		for _, key := range deletedBundles {
			// reprocess ClusterAddonsConfiguration again if it contains a conflicting addons
			if err := r.reprocessConflictingAddonsConfiguration(key, list); err != nil {
				return exerr.Wrap(err, "while requesting processing of conflicting ClusterAddonsConfigurations")
			}
		}
	}

	return nil
}

func (r *ReconcileClusterAddonsConfiguration) deleteAddonsProcess(addon *addonsv1alpha1.ClusterAddonsConfiguration) error {
	r.log.Infof("Start delete ClusterAddonsConfiguration %s", addon.Name)

	if addon.Status.Phase == addonsv1alpha1.AddonsConfigurationReady {
		adds, err := r.existingAddonsConfigurations(addon.Name)
		if err != nil {
			return exerr.Wrap(err, "while listing ClusterAddonsConfigurations")
		}

		deleteBroker := true
		for _, addon := range adds.Items {
			if addon.Status.Phase != addonsv1alpha1.AddonsConfigurationReady {
				// reprocess ClusterAddonsConfiguration again if was failed
				if err := r.reprocessAddonsConfiguration(&addon); err != nil {
					return exerr.Wrapf(err, "while requesting reprocess for ClusterAddonsConfiguration %s", addon.Name)
				}
			} else {
				deleteBroker = false
			}
		}
		if deleteBroker {
			r.log.Info("- delete ClusterServiceBroker")
			if err := r.brokerFacade.Delete(); err != nil {
				return exerr.Wrap(err, "while deleting ClusterServiceBroker")
			}
		}

		for _, repo := range addon.Status.Repositories {
			for _, add := range repo.Addons {
				id, err := r.removeBundle(add)
				if err != nil && !storage.IsNotFoundError(err) {
					return exerr.Wrapf(err, "while deleting bundle with charts for addon %s", add.Name)
				}
				if id != nil {
					r.log.Infof("- delete ClusterDocsTopic for bundle %s", add.Name)
					if err := r.docsProvider.EnsureClusterDocsTopicRemoved(string(*id)); err != nil {
						return exerr.Wrapf(err, "while ensuring ClusterDocsTopic for bundle %s is removed", *id)
					}
				}
			}
		}
		if !deleteBroker && r.syncBroker {
			if err := r.brokerSyncer.Sync(); err != nil {
				return exerr.Wrapf(err, "while syncing ClusterServiceBroker for addon %s", addon.Name)
			}
		}
	}
	if err := r.deleteFinalizer(addon); err != nil {
		return exerr.Wrapf(err, "while deleting finalizer from ClusterAddonsConfiguration %s", addon.Name)
	}

	r.log.Info("Delete ClusterAddonsConfiguration process completed")
	return nil
}

func (r *ReconcileClusterAddonsConfiguration) ensureBroker(addon *addonsv1alpha1.ClusterAddonsConfiguration) error {
	exist, err := r.brokerFacade.Exist()
	if err != nil {
		return exerr.Wrap(err, "while checking if ClusterServiceBroker exists")
	}
	if !exist {
		r.log.Info("- creating ClusterServiceBroker")
		if err := r.brokerFacade.Create(); err != nil {
			return exerr.Wrapf(err, "while creating ClusterServiceBroker for addon %s", addon.Name)
		}
	} else if r.syncBroker {
		if err := r.brokerSyncer.Sync(); err != nil {
			return exerr.Wrapf(err, "while syncing ClusterServiceBroker for addon %s", addon.Name)
		}
	}
	return nil
}

func (r *ReconcileClusterAddonsConfiguration) loadAddons(addon *addonsv1alpha1.ClusterAddonsConfiguration) *addons.RepositoryCollection {
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

func (r *ReconcileClusterAddonsConfiguration) createAddons(URL string) ([]*addons.AddonController, error) {
	var adds []*addons.AddonController

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

func (r *ReconcileClusterAddonsConfiguration) existingAddonsConfigurations(addonName string) (*addonsv1alpha1.ClusterAddonsConfigurationList, error) {
	addonsList := &addonsv1alpha1.ClusterAddonsConfigurationList{}
	addonsConfigurationList, err := r.addonsConfigurationList()
	if err != nil {
		return nil, exerr.Wrap(err, "while listing ClusterAddonsConfigurations")
	}

	for _, existAddon := range addonsConfigurationList.Items {
		if existAddon.Name != addonName {
			addonsList.Items = append(addonsList.Items, existAddon)
		}
	}

	return addonsList, nil
}

func (r *ReconcileClusterAddonsConfiguration) addonsConfigurationList() (*addonsv1alpha1.ClusterAddonsConfigurationList, error) {
	addonsConfigurationList := &addonsv1alpha1.ClusterAddonsConfigurationList{}

	err := r.Client.List(context.TODO(), &client.ListOptions{}, addonsConfigurationList)
	if err != nil {
		return addonsConfigurationList, exerr.Wrap(err, "during fetching ClusterAddonConfiguration list by client")
	}

	return addonsConfigurationList, nil
}

func (r *ReconcileClusterAddonsConfiguration) deleteOrphanBundles(repos []addonsv1alpha1.StatusRepository, lastRepos []addonsv1alpha1.StatusRepository) ([]string, error) {
	addonsToStay := map[string]addonsv1alpha1.Addon{}
	for _, repo := range repos {
		for _, ad := range repo.Addons {
			addonsToStay[ad.Key()] = ad
		}
	}

	var deletedBundlesKeys []string
	for _, repo := range lastRepos {
		for _, ad := range repo.Addons {
			if _, exist := addonsToStay[ad.Key()]; !exist {
				if _, err := r.removeBundle(ad); err != nil && !storage.IsNotFoundError(err) {
					return nil, exerr.Wrapf(err, "while deleting bundles and charts for addon %s", ad.Name)
				}
				deletedBundlesKeys = append(deletedBundlesKeys, ad.Key())
			}
		}
	}
	return deletedBundlesKeys, nil
}

func (r *ReconcileClusterAddonsConfiguration) deleteBundlesFromRepository(repos []addonsv1alpha1.StatusRepository) ([]string, error) {
	var deletedBundlesKeys []string
	for _, repo := range repos {
		for _, ad := range repo.Addons {
			if _, err := r.removeBundle(ad); err != nil && !storage.IsNotFoundError(err) {
				return nil, exerr.Wrapf(err, "while deleting bundles and charts for addon %s", ad.Name)
			}
			deletedBundlesKeys = append(deletedBundlesKeys, ad.Key())
		}
	}
	return deletedBundlesKeys, nil
}

func (r *ReconcileClusterAddonsConfiguration) removeBundle(ad addonsv1alpha1.Addon) (*internal.BundleID, error) {
	r.log.Infof("- delete bundle %s from storage", ad.Name)
	b, err := r.bundleStorage.Get(internal.ClusterWide, internal.BundleName(ad.Name), *semver.MustParse(ad.Version))
	if err != nil {
		return nil, err
	}
	err = r.bundleStorage.Remove(internal.ClusterWide, internal.BundleName(ad.Name), *semver.MustParse(ad.Version))
	if err != nil {
		return nil, err
	}

	for _, plan := range b.Plans {
		err = r.chartStorage.Remove(internal.ClusterWide, plan.ChartRef.Name, plan.ChartRef.Version)
		if err != nil {
			return nil, err
		}
	}
	r.syncBroker = true
	return &b.ID, nil
}

func (r *ReconcileClusterAddonsConfiguration) reprocessConflictingAddonsConfiguration(key string, list *addonsv1alpha1.ClusterAddonsConfigurationList) error {
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

func (r *ReconcileClusterAddonsConfiguration) reprocessAddonsConfiguration(addon *addonsv1alpha1.ClusterAddonsConfiguration) error {
	ad := &addonsv1alpha1.ClusterAddonsConfiguration{}
	if err := r.Client.Get(context.Background(), types.NamespacedName{Name: addon.Name}, ad); err != nil {
		return exerr.Wrapf(err, "while getting ClusterAddonsConfiguration %s", addon.Name)
	}
	ad.Spec.ReprocessRequest++
	if err := r.Client.Update(context.Background(), ad); err != nil {
		return exerr.Wrapf(err, "while incrementing a reprocess requests for ClusterAddonsConfiguration %s", addon.Name)
	}
	return nil
}

func (r *ReconcileClusterAddonsConfiguration) saveBundle(repositories *addons.RepositoryCollection) error {
	for _, addon := range repositories.ReadyAddons() {
		if len(addon.Bundle.Docs) == 1 {
			r.log.Infof("- ensure ClusterDocsTopic for bundle %s", addon.Bundle.ID)
			if err := r.docsProvider.EnsureClusterDocsTopic(addon.Bundle); err != nil {
				return exerr.Wrapf(err, "While ensuring ClusterDocsTopic for bundle %s: %v", addon.Bundle.ID, err)
			}
		}
		exist, err := r.bundleStorage.Upsert(internal.ClusterWide, addon.Bundle)
		if err != nil {
			addon.RegisteringError(err)
			r.log.Errorf("cannot upsert bundle %v:%v into storage", addon.Bundle.Name, addon.Bundle.Version)
			continue
		}
		if exist {
			r.log.Infof("bundle %v:%v already existed in storage, bundle was replaced", addon.Bundle.Name, addon.Bundle.Version)
		}
		err = r.saveCharts(addon.Charts)
		if err != nil {
			addon.RegisteringError(err)
			r.log.Errorf("cannot upsert charts of %v:%v bundle", addon.Bundle.Name, addon.Bundle.Version)
			continue
		}

		r.syncBroker = true
	}
	return nil
}

func (r *ReconcileClusterAddonsConfiguration) saveCharts(charts []*chart.Chart) error {
	for _, bundleChart := range charts {
		exist, err := r.chartStorage.Upsert(internal.ClusterWide, bundleChart)
		if err != nil {
			return err
		}
		if exist {
			r.log.Infof("chart %s already existed in storage, chart was replaced", bundleChart.Metadata.Name)
		}
	}
	return nil
}

func (r *ReconcileClusterAddonsConfiguration) statusSnapshot(addon *addonsv1alpha1.ClusterAddonsConfiguration, repositories *addons.RepositoryCollection) *addonsv1alpha1.ClusterAddonsConfiguration {
	addon.Status.Repositories = nil

	for _, repo := range repositories.Repositories {
		addonsRepository := repo.Repository
		addonsRepository.Addons = []addonsv1alpha1.Addon{}
		for _, addon := range repo.Addons {
			addonsRepository.Addons = append(addonsRepository.Addons, addon.Addon)
		}
		//addonsRepository.Status
		addon.Status.Repositories = append(addon.Status.Repositories, addonsRepository)
	}

	if repositories.IsRepositoriesFailed() {
		addon.Status.Phase = addonsv1alpha1.AddonsConfigurationFailed
	} else {
		addon.Status.Phase = addonsv1alpha1.AddonsConfigurationReady
	}

	return addon
}

func (r *ReconcileClusterAddonsConfiguration) updateAddonStatus(addon *addonsv1alpha1.ClusterAddonsConfiguration) (*addonsv1alpha1.ClusterAddonsConfiguration, error) {
	addon.Status.ObservedGeneration = addon.Generation
	addon.Status.LastProcessedTime = &v1.Time{Time: time.Now()}

	r.log.Infof("- update ClusterAddonsConfiguration %s status", addon.Name)
	err := r.Status().Update(context.TODO(), addon)
	if err != nil {
		return nil, exerr.Wrap(err, "while update ClusterAddonsConfiguration")
	}
	return addon, nil
}

func (r *ReconcileClusterAddonsConfiguration) prepareForProcessing(addon *addonsv1alpha1.ClusterAddonsConfiguration) (*addonsv1alpha1.ClusterAddonsConfiguration, error) {
	obj := addon.DeepCopy()
	obj.Status.Phase = addonsv1alpha1.AddonsConfigurationPending

	pendingInstance, err := r.updateAddonStatus(obj)
	if err != nil {
		return nil, exerr.Wrap(err, "while updating addons status")
	}

	if r.protection.hasFinalizer(pendingInstance.Finalizers) {
		return pendingInstance, nil
	}
	r.log.Info("- add a finalizer")
	pendingInstance.Finalizers = r.protection.addFinalizer(pendingInstance.Finalizers)

	err = r.Client.Update(context.Background(), pendingInstance)
	if err != nil {
		return nil, exerr.Wrap(err, "while updating addons status")
	}
	return pendingInstance, nil
}

func (r *ReconcileClusterAddonsConfiguration) deleteFinalizer(addon *addonsv1alpha1.ClusterAddonsConfiguration) error {
	obj := addon.DeepCopy()
	if !r.protection.hasFinalizer(obj.Finalizers) {
		return nil
	}
	r.log.Info("- delete a finalizer")
	obj.Finalizers = r.protection.removeFinalizer(obj.Finalizers)

	return r.Client.Update(context.Background(), obj)
}
