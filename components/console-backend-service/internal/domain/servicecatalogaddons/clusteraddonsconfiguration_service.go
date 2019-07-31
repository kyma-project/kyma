package servicecatalogaddons

import (
	"fmt"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalogaddons/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/pager"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/resource"
	"github.com/kyma-project/kyma/components/helm-broker/pkg/apis/addons/v1alpha1"
	addonsClientset "github.com/kyma-project/kyma/components/helm-broker/pkg/client/clientset/versioned/typed/addons/v1alpha1"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
)

type clusterAddonsConfigurationService struct {
	addonsNotifier    notifier
	addonsCfgClient   addonsClientset.AddonsV1alpha1Interface
	addonsCfgInformer cache.SharedIndexInformer
}

func newClusterAddonsConfigurationService(addonsCfgInformer cache.SharedIndexInformer, addonsCfgClient addonsClientset.AddonsV1alpha1Interface) *clusterAddonsConfigurationService {
	addonsNotifier := resource.NewNotifier()
	addonsCfgInformer.AddEventHandler(addonsNotifier)

	return &clusterAddonsConfigurationService{
		addonsCfgClient:   addonsCfgClient,
		addonsCfgInformer: addonsCfgInformer,
		addonsNotifier:    addonsNotifier,
	}
}

func (s *clusterAddonsConfigurationService) List(pagingParams pager.PagingParams) ([]*v1alpha1.ClusterAddonsConfiguration, error) {
	items, err := pager.From(s.addonsCfgInformer.GetStore()).Limit(pagingParams)
	if err != nil {
		return nil, err
	}

	var addons []*v1alpha1.ClusterAddonsConfiguration
	for _, item := range items {
		ac, ok := item.(*v1alpha1.ClusterAddonsConfiguration)
		if !ok {
			return nil, fmt.Errorf("incorrect item type: %T, should be: *v1alpha1.ClusterAddonsConfiguration", item)
		}

		addons = append(addons, ac)
	}

	return addons, nil
}

func (s *clusterAddonsConfigurationService) AddRepos(name string, urls []string) (*v1alpha1.ClusterAddonsConfiguration, error) {
	addon, err := s.getClusterAddonsConfiguration(name)
	if err != nil {
		return nil, err
	}

	addonCpy := addon.DeepCopy()
	for _, u := range urls {
		addonCpy.Spec.Repositories = append(addonCpy.Spec.Repositories, v1alpha1.SpecRepository{URL: u})
	}

	result, err := s.addonsCfgClient.ClusterAddonsConfigurations().Update(addonCpy)
	if err != nil {
		return nil, errors.Wrapf(err, "while updating %s %s", pretty.ClusterAddonsConfiguration, addonCpy.Name)
	}

	return result, nil
}

func (s *clusterAddonsConfigurationService) RemoveRepos(name string, urls []string) (*v1alpha1.ClusterAddonsConfiguration, error) {
	addon, err := s.getClusterAddonsConfiguration(name)
	if err != nil {
		return nil, err
	}

	resultRepos := filterOutRepositories(addon.Spec.Repositories, urls)

	addonCpy := addon.DeepCopy()
	addonCpy.Spec.Repositories = resultRepos

	updatedAddon, err := s.addonsCfgClient.ClusterAddonsConfigurations().Update(addonCpy)
	if err != nil {
		return nil, errors.Wrapf(err, "while updating %s %s", pretty.ClusterAddonsConfiguration, addonCpy.Name)
	}

	return updatedAddon, nil
}

func (s *clusterAddonsConfigurationService) Create(name string, urls []string, labels *gqlschema.Labels) (*v1alpha1.ClusterAddonsConfiguration, error) {
	addon := &v1alpha1.ClusterAddonsConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: toMapLabels(labels),
		},
		Spec: v1alpha1.ClusterAddonsConfigurationSpec{
			CommonAddonsConfigurationSpec: v1alpha1.CommonAddonsConfigurationSpec{
				Repositories: toSpecRepositories(urls),
			},
		},
	}

	result, err := s.addonsCfgClient.ClusterAddonsConfigurations().Create(addon)
	if err != nil {
		return nil, errors.Wrapf(err, "while creating %s %s", pretty.ClusterAddonsConfiguration, addon.Name)
	}

	return result, nil
}

func (s *clusterAddonsConfigurationService) Update(name string, urls []string, labels *gqlschema.Labels) (*v1alpha1.ClusterAddonsConfiguration, error) {
	addon, err := s.getClusterAddonsConfiguration(name)
	if err != nil {
		return nil, err
	}

	addonCpy := addon.DeepCopy()
	addonCpy.Spec.Repositories = toSpecRepositories(urls)
	addonCpy.Labels = toMapLabels(labels)

	result, err := s.addonsCfgClient.ClusterAddonsConfigurations().Update(addonCpy)
	if err != nil {
		return nil, errors.Wrapf(err, "while updating %s %s", pretty.ClusterAddonsConfiguration, addon.Name)
	}

	return result, nil
}

func (s *clusterAddonsConfigurationService) Delete(name string) (*v1alpha1.ClusterAddonsConfiguration, error) {
	addon, err := s.getClusterAddonsConfiguration(name)
	if err != nil {
		return nil, err
	}

	if err := s.addonsCfgClient.ClusterAddonsConfigurations().Delete(name, &metav1.DeleteOptions{}); err != nil {
		return nil, errors.Wrapf(err, "while deleting %s %s", pretty.ClusterAddonsConfiguration, addon.Name)
	}

	return addon, nil
}

func (s *clusterAddonsConfigurationService) Resync(name string) (*v1alpha1.ClusterAddonsConfiguration, error) {
	addon, err := s.getClusterAddonsConfiguration(name)
	if err != nil {
		return nil, err
	}
	addonCpy := addon.DeepCopy()
	addonCpy.Spec.ReprocessRequest++

	result, err := s.addonsCfgClient.ClusterAddonsConfigurations().Update(addonCpy)
	if err != nil {
		return nil, errors.Wrapf(err, "while updating %s %s", pretty.ClusterAddonsConfiguration, addon.Name)
	}

	return result, nil
}

func (s *clusterAddonsConfigurationService) getClusterAddonsConfiguration(name string) (*v1alpha1.ClusterAddonsConfiguration, error) {
	item, exists, err := s.addonsCfgInformer.GetStore().GetByKey(name)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting %s %s", pretty.ClusterAddonsConfiguration, name)
	}

	if !exists {
		return nil, errors.Errorf("%s doesn't exists", name)
	}

	addons, ok := item.(*v1alpha1.ClusterAddonsConfiguration)
	if !ok {
		return nil, fmt.Errorf("incorrect item type: %T, should be: *v1alpha1.ClusterAddonsConfiguration", item)
	}

	return addons, nil
}

func (svc *clusterAddonsConfigurationService) Subscribe(listener resource.Listener) {
	svc.addonsNotifier.AddListener(listener)
}

func (svc *clusterAddonsConfigurationService) Unsubscribe(listener resource.Listener) {
	svc.addonsNotifier.DeleteListener(listener)
}
