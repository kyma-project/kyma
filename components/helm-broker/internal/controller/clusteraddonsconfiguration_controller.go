package controller

import (
	"context"

	"time"

	"github.com/kyma-project/kyma/components/helm-broker/internal"
	"github.com/kyma-project/kyma/components/helm-broker/internal/bundle"
	"github.com/kyma-project/kyma/components/helm-broker/internal/controller/addons"
	addonsv1alpha1 "github.com/kyma-project/kyma/components/helm-broker/pkg/apis/addons/v1alpha1"
	exerr "github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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

	clusterBrokerFacade clusterBrokerFacade
	clusterDocsProvider clusterDocsProvider
	clusterBrokerSyncer clusterBrokerSyncer

	bundleProvider bundleProvider
	protection     protection

	// syncBroker informs ServiceBroker should be resync, it should be true if
	// operation insert/delete was made on storage
	syncBroker  bool
	developMode bool
}

// NewReconcileClusterAddonsConfiguration returns a new reconcile.Reconciler
func NewReconcileClusterAddonsConfiguration(mgr manager.Manager, bundleProvider bundleProvider, chartStorage chartStorage, bundleStorage bundleStorage, clusterBrokerFacade clusterBrokerFacade, clusterDocsProvider clusterDocsProvider, clusterBrokerSyncer clusterBrokerSyncer, developMode bool) reconcile.Reconciler {
	return &ReconcileClusterAddonsConfiguration{
		log:    logrus.WithField("controller", "cluster-addons-configuration"),
		Client: mgr.GetClient(),
		scheme: mgr.GetScheme(),

		bundleStorage: bundleStorage,
		chartStorage:  chartStorage,

		clusterBrokerFacade: clusterBrokerFacade,
		clusterDocsProvider: clusterDocsProvider,
		clusterBrokerSyncer: clusterBrokerSyncer,
		bundleProvider:      bundleProvider,

		protection: protection{},

		developMode: developMode,
		syncBroker:  false,
	}
}

// Reconcile reads that state of the cluster for a ClusterAddonsConfiguration object and makes changes based on the state read
// and what is in the ClusterAddonsConfiguration.Spec
func (r *ReconcileClusterAddonsConfiguration) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Fetch the ClusterAddonsConfiguration instance
	instance := &addonsv1alpha1.ClusterAddonsConfiguration{}
	err := r.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		return reconcile.Result{}, err
	}

	if instance.DeletionTimestamp != nil {
		if err := r.deleteAddonsProcess(instance); err != nil {
			return reconcile.Result{}, exerr.Wrapf(err, "while deleting ClusterAddonConfiguration %q", request.NamespacedName)
		}
		return reconcile.Result{}, nil
	}

	if instance.Status.ObservedGeneration == 0 {
		r.log.Infof("Start add ClusterAddonsConfiguration %s process", instance.Name)

		updatedInstance, err := r.addFinalizer(instance)
		if err != nil {
			return reconcile.Result{}, exerr.Wrapf(err, "while adding a finalizer to ClusterAddonsConfiguration %q", request.NamespacedName)
		}
		err = r.addAddonsProcess(updatedInstance)
		if err != nil {
			return reconcile.Result{}, exerr.Wrapf(err, "while creating ClusterAddonsConfiguration %q", request.NamespacedName)
		}
		r.log.Infof("Add ClusterAddonsConfiguration process completed")

	} else if instance.Generation > instance.Status.ObservedGeneration {
		r.log.Infof("Start update ClusterAddonsConfiguration %s process", instance.Name)

		instance.Status = addonsv1alpha1.ClusterAddonsConfigurationStatus{}
		err = r.addAddonsProcess(instance)
		if err != nil {
			return reconcile.Result{}, exerr.Wrapf(err, "while updating ClusterAddonsConfiguration %q", request.NamespacedName)
		}
		r.log.Infof("Update ClusterAddonsConfiguration %s process completed", instance.Name)
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileClusterAddonsConfiguration) addAddonsProcess(addon *addonsv1alpha1.ClusterAddonsConfiguration) error {
	repositories := addons.NewRepositoryCollection()

	r.log.Infof("- load bundles and charts for each addon")
	for _, specRepository := range addon.Spec.Repositories {
		if err := specRepository.VerifyURL(r.developMode); err != nil {
			r.log.Errorf("url %q address is not valid: %s", specRepository.URL, err)
			continue
		}
		r.log.Infof("create adds for %q repository", specRepository.URL)
		repo := addons.NewAddonsRepository(specRepository.URL)

		adds, err := r.createAddons(specRepository.URL)
		if err != nil {
			repo.Failed()
			repo.Repository.Reason = addonsv1alpha1.RepositoryURLFetchingError
			repo.Repository.Message = err.Error()
			repositories.AddRepository(repo)

			r.log.Errorf("while creating adds for repository from %q: %s", specRepository.URL, err)
			continue
		}

		repo.Addons = adds
		repositories.AddRepository(repo)
	}

	r.log.Info("- check duplicate ID addons alongside repositories")
	repositories.ReviseBundleDuplicationInRepository()

	r.log.Info("- check duplicates ID addons in existing addons configuration")
	list, err := r.existingAddonsConfigurationList(addon.Name)
	if err != nil {
		r.log.Errorf("cannot fetch AddonsConfiguration list: %s", err)
		return exerr.Wrap(err, "while fetching addons configuration list")
	}
	repositories.ReviseBundleDuplicationInClusterStorage(list)

	existingBundles, err := r.bundleStorage.FindAll(internal.ClusterWide)
	if err != nil {
		return exerr.Wrap(err, "while getting existing cluster-wide bundles from storage")
	}

	r.log.Info("- deleting unused ClusterDocsTopics")
	if err := r.deleteUnusedDocsTopics(existingBundles, repositories.ReadyAddons()); err != nil {
		return exerr.Wrap(err, "while deleting unused DocsTopics")
	}

	r.log.Info("- save ready bundles and charts in storage")
	r.saveBundle(repositories)

	r.log.Info("- update AddonsConfiguration status")
	r.statusSnapshot(addon, repositories)
	err = r.updateAddonStatus(addon)
	if err != nil {
		r.log.Errorf("cannot update ClusterAddonsConfiguration %s: %v", addon.Name, err)
		return exerr.Wrap(err, "while update AddonsConfiguration status")
	}

	r.log.Info("- ensuring ClusterServiceBroker")
	if err := r.ensureBroker(addon); err != nil {
		return exerr.Wrap(err, "while ensuring ClusterServiceBroker")
	}

	return nil
}

func (r *ReconcileClusterAddonsConfiguration) deleteAddonsProcess(addon *addonsv1alpha1.ClusterAddonsConfiguration) error {
	r.log.Infof("Start delete ClusterAddonsConfiguration %s", addon.Name)

	addonsCfgs, err := r.existingAddonsConfigurationList(addon.Name)
	if err != nil {
		return exerr.Wrap(err, "while listing ClusterAddonsConfigurations")
	}

	deleteBroker := true
	for _, addon := range addonsCfgs.Items {
		if addon.Status.Phase != addonsv1alpha1.AddonsConfigurationReady {
			// reprocess ClusterAddonConfig again if was failed
			addon.Spec.ReprocessRequest++
			if err := r.Client.Update(context.Background(), &addon); err != nil {
				return exerr.Wrapf(err, "while incrementing a reprocess requests for ClusterAddonConfiguration %s", addon.Name)
			}
		} else {
			deleteBroker = false
		}
	}
	if deleteBroker {
		r.log.Info("- deleting ClusterServiceBroker")
		if err := r.clusterBrokerFacade.Delete(); err != nil {
			return exerr.Wrap(err, "while deleting ClusterServiceBroker")
		}
	}

	if err := r.deleteFinalizer(addon); err != nil {
		return exerr.Wrapf(err, "while deleting finalizer from ClusterAddonsConfiguration %s", addon.Name)
	}

	r.log.Info("Delete ClusterAddonsConfiguration process completed")
	return nil
}

func (r *ReconcileClusterAddonsConfiguration) ensureBroker(addon *addonsv1alpha1.ClusterAddonsConfiguration) error {
	exist, err := r.clusterBrokerFacade.Exist()
	if err != nil {
		return exerr.Wrap(err, "while checking if ClusterServiceBroker exists")
	}
	if !exist {
		r.log.Info("- creating ClusterServiceBroker")
		if err := r.clusterBrokerFacade.Create(); err != nil {
			return exerr.Wrapf(err, "while creating ClusterServiceBroker for addon %s", addon.Name)
		}
	} else if r.syncBroker {
		if err := r.clusterBrokerSyncer.Sync(); err != nil {
			return exerr.Wrapf(err, "while syncing ClusterServiceBroker for addon %s", addon.Name)
		}
	}
	return nil
}

func (r *ReconcileClusterAddonsConfiguration) createAddons(URL string) ([]*addons.AddonController, error) {
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
				logrus.Errorf("while fetching addon: %s", err)
				continue
			}
			if bundle.IsLoadingError(err) {
				addon.LoadingError(err)
				adds = append(adds, addon)
				logrus.Errorf("while loading addon: %s", err)
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

func (r *ReconcileClusterAddonsConfiguration) existingAddonsConfigurationList(addonName string) (*addonsv1alpha1.ClusterAddonsConfigurationList, error) {
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

func (r *ReconcileClusterAddonsConfiguration) saveBundle(repositories *addons.RepositoryCollection) error {
	for _, addon := range repositories.ReadyAddons() {
		if len(addon.Bundle.Docs) == 1 {
			r.log.Infof("- creating ClusterDocsTopic for bundle %s", addon.Bundle.ID)
			if err := r.clusterDocsProvider.EnsureClusterDocsTopic(addon.Bundle); err != nil {
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
			r.log.Errorf("cannot upsert %s chart: %s", bundleChart.Metadata.Name, err)
			return err
		}
		if exist {
			r.log.Infof("chart %s already existed in storage, chart was replaced", bundleChart.Metadata.Name)
		}
	}
	return nil
}

func (r *ReconcileClusterAddonsConfiguration) deleteUnusedDocsTopics(existingBundles []*internal.Bundle, newBundles []*addons.AddonController) error {
	for _, v := range existingBundles {
		deleteDocsTopic := true
		for _, b := range newBundles {
			// don't delete docs topics if bundle exists in the new collection
			if b.Bundle.ID == v.ID {
				deleteDocsTopic = false
			}
		}
		if deleteDocsTopic {
			if err := r.clusterDocsProvider.EnsureClusterDocsTopicRemoved(string(v.ID)); err != nil {
				return exerr.Wrapf(err, "while ensuring ClusterDocsTopic %s is removed", v.ID)
			}
		}
	}

	return nil
}

func (r *ReconcileClusterAddonsConfiguration) statusSnapshot(addon *addonsv1alpha1.ClusterAddonsConfiguration, repositories *addons.RepositoryCollection) {
	addon.Status.Repositories = nil

	for _, repo := range repositories.Repositories {
		addonsRepository := repo.Repository
		addonsRepository.Addons = []addonsv1alpha1.Addon{}
		for _, addon := range repo.Addons {
			addonsRepository.Addons = append(addonsRepository.Addons, addon.Addon)
		}
		addon.Status.Repositories = append(addon.Status.Repositories, addonsRepository)
	}

	if repositories.IsRepositoriesIDConflict() {
		addon.Status.Phase = addonsv1alpha1.AddonsConfigurationFailed
	} else {
		addon.Status.Phase = addonsv1alpha1.AddonsConfigurationReady
	}
}

func (r *ReconcileClusterAddonsConfiguration) updateAddonStatus(addon *addonsv1alpha1.ClusterAddonsConfiguration) error {
	addon.Status.ObservedGeneration = addon.Generation
	addon.Status.LastProcessedTime = &v1.Time{Time: time.Now()}

	err := r.Status().Update(context.TODO(), addon)
	if err != nil {
		return exerr.Wrap(err, "while update ClusterAddonsConfiguration")
	}
	return nil
}

func (r *ReconcileClusterAddonsConfiguration) addFinalizer(addon *addonsv1alpha1.ClusterAddonsConfiguration) (*addonsv1alpha1.ClusterAddonsConfiguration, error) {
	obj := addon.DeepCopy()
	if r.protection.hasFinalizer(obj.Finalizers) {
		return obj, nil
	}
	obj.Finalizers = r.protection.addFinalizer(obj.Finalizers)

	err := r.Client.Update(context.Background(), obj)
	if err != nil {
		return nil, err
	}
	return obj, nil
}

func (r *ReconcileClusterAddonsConfiguration) deleteFinalizer(addon *addonsv1alpha1.ClusterAddonsConfiguration) error {
	obj := addon.DeepCopy()
	if !r.protection.hasFinalizer(obj.Finalizers) {
		return nil
	}
	obj.Finalizers = r.protection.removeFinalizer(obj.Finalizers)

	return r.Client.Update(context.Background(), obj)
}
