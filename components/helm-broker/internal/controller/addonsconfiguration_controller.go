package controller

import (
	"context"

	"time"

	"github.com/Masterminds/semver"
	"github.com/kyma-project/kyma/components/helm-broker/internal"
	"github.com/kyma-project/kyma/components/helm-broker/internal/bundle"
	"github.com/kyma-project/kyma/components/helm-broker/internal/controller/addons"
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
	log    logrus.FieldLogger
	client.Client
	scheme *runtime.Scheme

	chartStorage  chartStorage
	bundleStorage bundleStorage

	brokerFacade      brokerFacade
	brokerSyncer      brokerSyncer
	docsTopicProvider docsProvider

	protection     protection
	bundleProvider bundleProvider

	// syncBroker informs ServiceBroker should be resync, it should be true if
	// operation insert/delete was made on storage
	syncBroker  bool
	developMode bool
}

// NewReconcileAddonsConfiguration returns a new reconcile.Reconciler
func NewReconcileAddonsConfiguration(mgr manager.Manager, bp bundleProvider, brokerFacade brokerFacade, chartStorage chartStorage, bundleStorage bundleStorage, developMode bool, docsTopicProvider docsProvider, brokerSyncer brokerSyncer) reconcile.Reconciler {
	return &ReconcileAddonsConfiguration{
		log:    logrus.WithField("controller", "addons-configuration"),
		Client: mgr.GetClient(),
		scheme: mgr.GetScheme(),

		chartStorage:  chartStorage,
		bundleStorage: bundleStorage,

		bundleProvider: bp,
		protection:     protection{},

		brokerSyncer:      brokerSyncer,
		brokerFacade:      brokerFacade,
		docsTopicProvider: docsTopicProvider,

		developMode: developMode,
		syncBroker:  false,
	}
}

// Reconcile reads that state of the cluster for a AddonsConfiguration object and makes changes based on the state read
// and what is in the AddonsConfiguration.Spec
func (r *ReconcileAddonsConfiguration) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	instance := &addonsv1alpha1.AddonsConfiguration{}
	err := r.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		return reconcile.Result{}, err
	}

	if instance.DeletionTimestamp != nil {
		if err := r.deleteAddonsProcess(instance); err != nil {
			return reconcile.Result{RequeueAfter: time.Second * 15}, exerr.Wrapf(err, "while deleting AddonConfiguration %q", request.NamespacedName)
		}
		return reconcile.Result{}, nil
	}

	if instance.Status.ObservedGeneration == 0 {
		r.log.Infof("Start add AddonsConfiguration %s/%s process", instance.Name, instance.Namespace)

		updatedInstance, err := r.addFinalizer(instance)
		if err != nil {
			return reconcile.Result{Requeue: true}, exerr.Wrapf(err, "while adding a finalizer to AddonsConfiguration %q", request.NamespacedName)
		}
		err = r.addAddonsProcess(updatedInstance)
		if err != nil {
			return reconcile.Result{}, exerr.Wrapf(err, "while creating AddonsConfiguration %q", request.NamespacedName)
		}
		r.log.Info("Add AddonsConfiguration process completed")

	} else if instance.Generation > instance.Status.ObservedGeneration {
		r.log.Infof("Start update AddonsConfiguration %s/%s process", instance.Name, instance.Namespace)

		instance.Status = addonsv1alpha1.AddonsConfigurationStatus{}
		err = r.addAddonsProcess(instance)
		if err != nil {
			return reconcile.Result{}, exerr.Wrapf(err, "while updating AddonsConfiguration %q", request.NamespacedName)
		}
		r.log.Info("Update AddonsConfiguration process completed")
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileAddonsConfiguration) addAddonsProcess(addon *addonsv1alpha1.AddonsConfiguration) error {
	repositories := addons.NewRepositoryCollection()

	r.log.Infof("- load bundles and charts for each addon")
	for _, specRepository := range addon.Spec.Repositories {
		if err := specRepository.VerifyURL(r.developMode); err != nil {
			r.log.Errorf("url %q address is not valid: %s", specRepository.URL, err)
			continue
		}
		r.log.Infof("create addons for %q repository", specRepository.URL)
		repo := addons.NewAddonsRepository(specRepository.URL)

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
	if err = r.updateAddonStatus(addon); err != nil {
		r.log.Errorf("cannot update AddonsConfiguration %s: %v", addon.Name, err)
		return exerr.Wrap(err, "while update AddonsConfiguration status")
	}

	r.log.Info("- ensuring ServiceBroker")
	if err := r.ensureBroker(addon); err != nil {
		return exerr.Wrap(err, "while ensuring ServiceBroker")
	}

	return nil
}

func (r *ReconcileAddonsConfiguration) deleteAddonsProcess(addon *addonsv1alpha1.AddonsConfiguration) error {
	r.log.Infof("Start delete AddonsConfiguration %s/%s process", addon.Name, addon.Namespace)

	addonsCfgs, err := r.existingAddonsConfigurationList(addon)
	if err != nil {
		return exerr.Wrapf(err, "while listing AddonsConfigurations in namespace %s", addon.Namespace)
	}

	deleteBroker := true
	for _, addon := range addonsCfgs.Items {
		if addon.Status.Phase != addonsv1alpha1.AddonsConfigurationReady {
			// reprocess AddonConfig again if it was failed
			addon.Spec.ReprocessRequest++
			if err := r.Client.Update(context.Background(), &addon); err != nil {
				return exerr.Wrapf(err, "while incrementing a reprocess requests for AddonConfiguration %s/%s", addon.Name, addon.Namespace)
			}
		} else {
			deleteBroker = false
		}
	}
	if deleteBroker {
		if err := r.brokerFacade.Delete(addon.Namespace); err != nil {
			return exerr.Wrapf(err, "while deleting ServiceBroker from namespace %s", addon.Namespace)
		}
	}

	for _, repo := range addon.Status.Repositories {
		if repo.Status == addonsv1alpha1.RepositoryStatusReady {
			for _, add := range repo.Addons {
				if add.Status == addonsv1alpha1.AddonStatusReady {

					r.log.Infof("- deleting DocsTopic for bundle %s", add)
					b, err := r.bundleStorage.Get(internal.Namespace(addon.Namespace), internal.BundleName(add.Name), *semver.MustParse(add.Version))
					if err != nil {
						return exerr.Wrapf(err, "while getting bundle %s from namespace %s", add.Name, addon.Namespace)
					}
					if err := r.docsTopicProvider.EnsureDocsTopicRemoved(string(b.ID), addon.Namespace); err != nil {
						return exerr.Wrapf(err, "while ensuring ClusterDocsTopic for bundle %s is removed", b.ID)
					}

					r.log.Infof("- deleting bundle %s from namespace %s", add, addon.Namespace)
					if err := r.bundleStorage.Remove(internal.Namespace(addon.Namespace), internal.BundleName(add.Name), *semver.MustParse(add.Version)); err != nil {
						return exerr.Wrapf(err, "while deleting bundle %s from storage", add.Name)
					}
				}
			}
		}
	}

	if err := r.deleteFinalizer(addon); err != nil {
		return exerr.Wrapf(err, "while deleting finalizer for AddonConfiguration %s/%s", addon.Name, addon.Namespace)
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
	addonsConfigurationList := &addonsv1alpha1.AddonsConfigurationList{}

	err := r.Client.List(context.TODO(), &client.ListOptions{Namespace: namespace}, addonsConfigurationList)
	if err != nil {
		return addonsConfigurationList, exerr.Wrap(err, "during fetching AddonConfiguration list by client")
	}

	return addonsConfigurationList, nil
}

func (r *ReconcileAddonsConfiguration) ensureBroker(addon *addonsv1alpha1.AddonsConfiguration) error {
	exist, err := r.brokerFacade.Exist(addon.Namespace)
	if err != nil {
		return exerr.Wrapf(err, "while checking if ServiceBroker exist in namespace %s", addon.Namespace)
	}
	if !exist {
		// status
		if err := r.brokerFacade.Create(addon.Namespace); err != nil {
			return exerr.Wrapf(err, "while creating ServiceBroker for AddonConfiguration %s/%s", addon.Name, addon.Namespace)
		}
	} else if r.syncBroker {
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

func (r *ReconcileAddonsConfiguration) saveBundle(namespace internal.Namespace, repositories *addons.RepositoryCollection) error {
	for _, addon := range repositories.ReadyAddons() {
		if len(addon.Bundle.Docs) == 1 {
			r.log.Infof("- creating DocsTopic for bundle %s", addon.Bundle.ID)
			if err := r.docsTopicProvider.EnsureDocsTopic(addon.Bundle, string(namespace)); err != nil {
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
			r.log.Errorf("cannot upsert %s chart: %s", bundleChart.Metadata.Name, err)
			return err
		}
		if exist {
			r.log.Infof("chart %s already existed in storage, chart was replaced", bundleChart.Metadata.Name)
		}
	}
	return nil
}

func (r *ReconcileAddonsConfiguration) statusSnapshot(addon *addonsv1alpha1.AddonsConfiguration, repositories *addons.RepositoryCollection) {
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

func (r *ReconcileAddonsConfiguration) updateAddonStatus(addon *addonsv1alpha1.AddonsConfiguration) error {
	instance := &addonsv1alpha1.AddonsConfiguration{}
	err := r.Get(context.TODO(), types.NamespacedName{Name: addon.Name, Namespace: addon.Namespace}, instance)
	if err != nil {
		return exerr.Wrap(err, "while getting AddonsConfiguration")
	}

	instance.Status.ObservedGeneration = instance.Generation
	instance.Status.LastProcessedTime = &v1.Time{Time: time.Now()}

	err = r.Status().Update(context.TODO(), instance)
	if err != nil {
		return exerr.Wrap(err, "while update AddonsConfiguration")
	}
	return nil
}

func (r *ReconcileAddonsConfiguration) addFinalizer(addon *addonsv1alpha1.AddonsConfiguration) (*addonsv1alpha1.AddonsConfiguration, error) {
	obj := addon.DeepCopy()
	if r.protection.hasFinalizer(obj.Finalizers) {
		return obj, nil
	}
	r.log.Info("- adding a finalizer")
	obj.Finalizers = r.protection.addFinalizer(obj.Finalizers)

	err := r.Client.Update(context.Background(), obj)
	if err != nil {
		return nil, err
	}
	return obj, nil
}

func (r *ReconcileAddonsConfiguration) deleteFinalizer(addon *addonsv1alpha1.AddonsConfiguration) error {
	obj := addon.DeepCopy()
	if !r.protection.hasFinalizer(obj.Finalizers) {
		return nil
	}
	r.log.Info("- deleting a finalizer")
	obj.Finalizers = r.protection.removeFinalizer(obj.Finalizers)

	return r.Client.Update(context.Background(), obj)
}
