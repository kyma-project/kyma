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
	"k8s.io/client-go/util/retry"
)

type addonsConfigurationService struct {
	addonsNotifier    notifier
	addonsCfgClient   addonsClientset.AddonsV1alpha1Interface
	addonsCfgInformer cache.SharedIndexInformer
}

func newAddonsConfigurationService(addonsCfgInformer cache.SharedIndexInformer, addonsCfgClient addonsClientset.AddonsV1alpha1Interface) *addonsConfigurationService {
	addonsNotifier := resource.NewNotifier()
	addonsCfgInformer.AddEventHandler(addonsNotifier)

	return &addonsConfigurationService{
		addonsCfgClient:   addonsCfgClient,
		addonsCfgInformer: addonsCfgInformer,
		addonsNotifier:    addonsNotifier,
	}
}

func (s *addonsConfigurationService) List(namespace string, pagingParams pager.PagingParams) ([]*v1alpha1.AddonsConfiguration, error) {
	items, err := s.addonsCfgInformer.GetIndexer().ByIndex("namespace", namespace)
	if err != nil {
		return nil, err
	}

	var addons []*v1alpha1.AddonsConfiguration
	for _, item := range items {
		ac, ok := item.(*v1alpha1.AddonsConfiguration)
		if !ok {
			return nil, fmt.Errorf("incorrect item type: %T, should be: *v1alpha1.AddonsConfiguration", item)
		}

		addons = append(addons, ac)
	}

	return addons, nil
}

func (s *addonsConfigurationService) AddRepos(name, namespace string, urls []string) (*v1alpha1.AddonsConfiguration, error) {
	var addon *v1alpha1.AddonsConfiguration
	var err error
	if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		addon, err = s.addonsCfgClient.AddonsConfigurations(namespace).Get(name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		for _, u := range urls {
			addon.Spec.Repositories = append(addon.Spec.Repositories, v1alpha1.SpecRepository{URL: u})
		}

		_, err = s.addonsCfgClient.AddonsConfigurations(namespace).Update(addon)
		return err
	}); err != nil {
		return nil, errors.Wrapf(err, "while updating %s %s", pretty.AddonsConfiguration, name)
	}

	return addon, nil
}

func (s *addonsConfigurationService) RemoveRepos(name, namespace string, urls []string) (*v1alpha1.AddonsConfiguration, error) {
	var addon *v1alpha1.AddonsConfiguration
	var err error
	if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		addon, err = s.addonsCfgClient.AddonsConfigurations(namespace).Get(name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		resultRepos := filterOutRepositories(addon.Spec.Repositories, urls)
		addon.Spec.Repositories = resultRepos

		_, err = s.addonsCfgClient.AddonsConfigurations(namespace).Update(addon)
		return err
	}); err != nil {
		return nil, errors.Wrapf(err, "while updating %s %s", pretty.AddonsConfiguration, name)
	}

	return addon, nil
}

func (s *addonsConfigurationService) Create(name, namespace string, urls []string, labels *gqlschema.Labels) (*v1alpha1.AddonsConfiguration, error) {
	addon := &v1alpha1.AddonsConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: toMapLabels(labels),
		},
		Spec: v1alpha1.AddonsConfigurationSpec{
			CommonAddonsConfigurationSpec: v1alpha1.CommonAddonsConfigurationSpec{
				Repositories: toSpecRepositories(urls),
			},
		},
	}

	result, err := s.addonsCfgClient.AddonsConfigurations(namespace).Create(addon)
	if err != nil {
		return nil, errors.Wrapf(err, "while creating %s %s", pretty.AddonsConfiguration, addon.Name)
	}

	return result, nil
}

func (s *addonsConfigurationService) Update(name, namespace string, urls []string, labels *gqlschema.Labels) (*v1alpha1.AddonsConfiguration, error) {
	addon, err := s.getAddonsConfiguration(name, namespace)
	if err != nil {
		return nil, err
	}

	addonCpy := addon.DeepCopy()
	addonCpy.Spec.Repositories = toSpecRepositories(urls)
	addonCpy.Labels = toMapLabels(labels)

	result, err := s.addonsCfgClient.AddonsConfigurations(namespace).Update(addonCpy)
	if err != nil {
		return nil, errors.Wrapf(err, "while updating %s %s", pretty.AddonsConfiguration, addon.Name)
	}

	return result, nil
}

func (s *addonsConfigurationService) Delete(name, namespace string) (*v1alpha1.AddonsConfiguration, error) {
	addon, err := s.getAddonsConfiguration(name, namespace)
	if err != nil {
		return nil, err
	}

	if err := s.addonsCfgClient.AddonsConfigurations(namespace).Delete(name, &metav1.DeleteOptions{}); err != nil {
		return nil, errors.Wrapf(err, "while deleting %s %s", pretty.AddonsConfiguration, addon.Name)
	}

	return addon, nil
}

func (s *addonsConfigurationService) Resync(name, namespace string) (*v1alpha1.AddonsConfiguration, error) {
	var result *v1alpha1.AddonsConfiguration
	if err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		addon, err := s.getAddonsConfiguration(name, namespace)
		if err != nil {
			return err
		}
		addonCpy := addon.DeepCopy()
		addonCpy.Spec.ReprocessRequest++

		result, err = s.addonsCfgClient.AddonsConfigurations(namespace).Update(addonCpy)
		return err
	}); err != nil {
		return nil, errors.Wrapf(err, "cannot update AddonsConfiguration %s/%s", name, namespace)
	}

	return result, nil
}

func (s *addonsConfigurationService) getAddonsConfiguration(name, namespace string) (*v1alpha1.AddonsConfiguration, error) {
	key := fmt.Sprintf("%s/%s", namespace, name)
	item, exists, err := s.addonsCfgInformer.GetStore().GetByKey(key)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting %s %s", pretty.AddonsConfiguration, name)
	}

	if !exists {
		return nil, errors.Errorf("%s doesn't exists", name)
	}

	addons, ok := item.(*v1alpha1.AddonsConfiguration)
	if !ok {
		return nil, fmt.Errorf("incorrect item type: %T, should be: *v1alpha1.AddonsConfiguration", item)
	}

	return addons, nil
}

func filterOutRepositories(repos []v1alpha1.SpecRepository, urls []string) []v1alpha1.SpecRepository {
	idxURLs := map[string]struct{}{}
	for _, u := range urls {
		idxURLs[u] = struct{}{}
	}

	var result []v1alpha1.SpecRepository
	for _, r := range repos {
		if _, found := idxURLs[r.URL]; !found {
			result = append(result, r)
		}
	}
	return result
}

func toMapLabels(givenLabels *gqlschema.Labels) map[string]string {
	if givenLabels == nil {
		return nil
	}

	labels := map[string]string{}
	for k, v := range *givenLabels {
		labels[k] = v
	}
	return labels
}

func toSpecRepositories(urls []string) []v1alpha1.SpecRepository {
	var repos []v1alpha1.SpecRepository
	for _, u := range urls {
		repos = append(repos, v1alpha1.SpecRepository{URL: u})
	}

	return repos
}

func (svc *addonsConfigurationService) Subscribe(listener resource.Listener) {
	svc.addonsNotifier.AddListener(listener)
}

func (svc *addonsConfigurationService) Unsubscribe(listener resource.Listener) {
	svc.addonsNotifier.DeleteListener(listener)
}
