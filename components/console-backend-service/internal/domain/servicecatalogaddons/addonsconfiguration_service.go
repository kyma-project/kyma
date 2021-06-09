package servicecatalogaddons

import (
	"context"
	"fmt"

	"github.com/kyma-project/helm-broker/pkg/apis/addons/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalogaddons/extractor"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalogaddons/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/pager"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/resource"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/retry"
)

type addonsConfigurationService struct {
	addonsNotifier    notifier
	addonsCfgClient   dynamic.NamespaceableResourceInterface
	addonsCfgInformer cache.SharedIndexInformer

	extractor extractor.AddonsUnstructuredExtractor
}

func newAddonsConfigurationService(addonsCfgInformer cache.SharedIndexInformer, addonsCfgClient dynamic.NamespaceableResourceInterface) *addonsConfigurationService {
	addonsNotifier := resource.NewNotifier()
	addonsCfgInformer.AddEventHandler(addonsNotifier)

	return &addonsConfigurationService{
		addonsCfgClient:   addonsCfgClient,
		addonsCfgInformer: addonsCfgInformer,
		addonsNotifier:    addonsNotifier,
		extractor:         extractor.AddonsUnstructuredExtractor{},
	}
}

func (s *addonsConfigurationService) List(namespace string, pagingParams pager.PagingParams) ([]*v1alpha1.AddonsConfiguration, error) {
	items, err := s.addonsCfgInformer.GetIndexer().ByIndex("namespace", namespace)
	if err != nil {
		return nil, err
	}

	var addons []*v1alpha1.AddonsConfiguration
	for _, item := range items {
		u, ok := item.(*unstructured.Unstructured)
		if !ok {
			return nil, fmt.Errorf("incorrect item type: %T, should be: *v1alpha1.AddonsConfiguration", item)
		}

		addon, err := s.extractor.FromUnstructured(u)
		if err != nil {
			return nil, err
		}
		addons = append(addons, addon)
	}

	return addons, nil
}

func (s *addonsConfigurationService) AddRepos(name, namespace string, repositories []*gqlschema.AddonsConfigurationRepositoryInput) (*v1alpha1.AddonsConfiguration, error) {
	var addon *v1alpha1.AddonsConfiguration
	var err error
	if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		addon, err = s.getAddonsConfiguration(name, namespace)
		if err != nil {
			return err
		}
		addon.Spec.Repositories = append(addon.Spec.Repositories, toSpecRepositories(repositories)...)

		obj, err := s.extractor.ToUnstructured(addon)
		if err != nil {
			return err
		}

		_, err = s.addonsCfgClient.Namespace(namespace).Update(context.Background(), obj, metav1.UpdateOptions{})
		return err
	}); err != nil {
		return nil, errors.Wrapf(err, "while updating %s %s", pretty.AddonsConfiguration, name)
	}

	return addon, nil
}

func (s *addonsConfigurationService) RemoveRepos(name, namespace string, reposToRemove []string) (*v1alpha1.AddonsConfiguration, error) {
	var addon *v1alpha1.AddonsConfiguration
	var err error
	if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		addon, err = s.getAddonsConfiguration(name, namespace)
		if err != nil {
			return err
		}
		resultRepos := filterOutRepositories(addon.Spec.Repositories, reposToRemove)
		addon.Spec.Repositories = resultRepos

		obj, err := s.extractor.ToUnstructured(addon)
		if err != nil {
			return err
		}

		_, err = s.addonsCfgClient.Namespace(namespace).Update(context.Background(), obj, metav1.UpdateOptions{})
		return err
	}); err != nil {
		return nil, errors.Wrapf(err, "while updating %s %s", pretty.AddonsConfiguration, name)
	}

	return addon, nil
}

func (s *addonsConfigurationService) Create(name, namespace string, repository []*gqlschema.AddonsConfigurationRepositoryInput, labels gqlschema.Labels) (*v1alpha1.AddonsConfiguration, error) {
	addon := &v1alpha1.AddonsConfiguration{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AddonsConfiguration",
			APIVersion: "addons.kyma-project.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: toMapLabels(labels),
		},
		Spec: v1alpha1.AddonsConfigurationSpec{
			CommonAddonsConfigurationSpec: v1alpha1.CommonAddonsConfigurationSpec{
				Repositories: toSpecRepositories(repository),
			},
		},
	}
	obj, err := s.extractor.ToUnstructured(addon)
	if err != nil {
		return nil, err
	}
	_, err = s.addonsCfgClient.Namespace(namespace).Create(context.Background(), obj, metav1.CreateOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "while creating %s %s", pretty.AddonsConfiguration, addon.Name)
	}

	return addon, nil
}

func (s *addonsConfigurationService) Update(name, namespace string, repository []*gqlschema.AddonsConfigurationRepositoryInput, labels gqlschema.Labels) (*v1alpha1.AddonsConfiguration, error) {
	addon, err := s.getAddonsConfiguration(name, namespace)
	if err != nil {
		return nil, err
	}
	addon.Spec.Repositories = toSpecRepositories(repository)
	addon.Labels = toMapLabels(labels)

	obj, err := s.extractor.ToUnstructured(addon)
	if err != nil {
		return nil, err
	}

	_, err = s.addonsCfgClient.Namespace(namespace).Update(context.Background(), obj, metav1.UpdateOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "while updating %s %s", pretty.AddonsConfiguration, addon.Name)
	}

	return addon, nil
}

func (s *addonsConfigurationService) Delete(name, namespace string) (*v1alpha1.AddonsConfiguration, error) {
	addon, err := s.getAddonsConfiguration(name, namespace)
	if err != nil {
		return nil, err
	}

	if err := s.addonsCfgClient.Namespace(namespace).Delete(context.Background(), name, metav1.DeleteOptions{}); err != nil {
		return nil, errors.Wrapf(err, "while deleting %s %s", pretty.AddonsConfiguration, addon.Name)
	}

	return addon, nil
}

func (s *addonsConfigurationService) Resync(name, namespace string) (*v1alpha1.AddonsConfiguration, error) {
	var addon *v1alpha1.AddonsConfiguration
	var err error
	if err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		addon, err = s.getAddonsConfiguration(name, namespace)
		if err != nil {
			return err
		}
		addon.Spec.ReprocessRequest++

		obj, err := s.extractor.ToUnstructured(addon)
		if err != nil {
			return err
		}

		_, err = s.addonsCfgClient.Namespace(namespace).Update(context.Background(), obj, metav1.UpdateOptions{})
		return err
	}); err != nil {
		return nil, errors.Wrapf(err, "cannot update AddonsConfiguration %s/%s", name, namespace)
	}

	return addon, nil
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

	u, ok := item.(*unstructured.Unstructured)
	if !ok {
		return nil, fmt.Errorf("incorrect item type: %T, should be: *unstructured.Unstructured", item)
	}
	addons, err := s.extractor.FromUnstructured(u)
	if err != nil {
		return nil, err
	}

	return addons, nil
}

func filterOutRepositories(repository []v1alpha1.SpecRepository, repos []string) []v1alpha1.SpecRepository {
	idxURLs := map[string]struct{}{}
	for _, repo := range repos {
		idxURLs[repo] = struct{}{}
	}

	result := make([]v1alpha1.SpecRepository, 0)
	for _, r := range repository {
		if _, found := idxURLs[r.URL]; !found {
			result = append(result, r)
		}
	}
	return result
}

func toMapLabels(givenLabels gqlschema.Labels) map[string]string {
	if givenLabels == nil {
		return nil
	}

	labels := map[string]string{}
	for k, v := range givenLabels {
		labels[k] = v
	}
	return labels
}

func toSpecRepository(repo *gqlschema.AddonsConfigurationRepositoryInput) v1alpha1.SpecRepository {
	secretRef := &v1.SecretReference{}
	if repo.SecretRef != nil {
		secretRef.Name = repo.SecretRef.Name
		secretRef.Namespace = repo.SecretRef.Namespace
	} else {
		secretRef = nil
	}
	return v1alpha1.SpecRepository{URL: repo.URL, SecretRef: secretRef}
}

func toSpecRepositories(repositories []*gqlschema.AddonsConfigurationRepositoryInput) []v1alpha1.SpecRepository {
	var result []v1alpha1.SpecRepository

	for _, repo := range repositories {
		result = append(result, toSpecRepository(repo))
	}
	return result
}

func (svc *addonsConfigurationService) Subscribe(listener resource.Listener) {
	svc.addonsNotifier.AddListener(listener)
}

func (svc *addonsConfigurationService) Unsubscribe(listener resource.Listener) {
	svc.addonsNotifier.DeleteListener(listener)
}
