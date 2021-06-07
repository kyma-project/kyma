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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/retry"
)

type clusterAddonsConfigurationService struct {
	addonsNotifier    notifier
	addonsCfgClient   dynamic.ResourceInterface
	addonsCfgInformer cache.SharedIndexInformer

	extractor extractor.ClusterAddonsUnstructuredExtractor
}

func newClusterAddonsConfigurationService(addonsCfgInformer cache.SharedIndexInformer, addonsCfgClient dynamic.ResourceInterface) *clusterAddonsConfigurationService {
	addonsNotifier := resource.NewNotifier()
	addonsCfgInformer.AddEventHandler(addonsNotifier)

	return &clusterAddonsConfigurationService{
		addonsCfgClient:   addonsCfgClient,
		addonsCfgInformer: addonsCfgInformer,
		addonsNotifier:    addonsNotifier,
		extractor:         extractor.ClusterAddonsUnstructuredExtractor{},
	}
}

func (s *clusterAddonsConfigurationService) List(pagingParams pager.PagingParams) ([]*v1alpha1.ClusterAddonsConfiguration, error) {
	items, err := pager.From(s.addonsCfgInformer.GetStore()).Limit(pagingParams)
	if err != nil {
		return nil, err
	}

	var addons []*v1alpha1.ClusterAddonsConfiguration
	for _, item := range items {
		u, ok := item.(*unstructured.Unstructured)
		if !ok {
			return nil, fmt.Errorf("incorrect item type: %T, should be: *v1alpha1.ClusterAddonsConfiguration", item)
		}

		addon, err := s.extractor.FromUnstructured(u)
		if err != nil {
			return nil, err
		}
		addons = append(addons, addon)
	}

	return addons, nil
}

func (s *clusterAddonsConfigurationService) AddRepos(name string, repository []*gqlschema.AddonsConfigurationRepositoryInput) (*v1alpha1.ClusterAddonsConfiguration, error) {
	var addon *v1alpha1.ClusterAddonsConfiguration
	if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		obj, err := s.addonsCfgClient.Get(context.Background(), name, metav1.GetOptions{})
		if err != nil {
			return err
		}

		addon, err = s.extractor.FromUnstructured(obj)
		if err != nil {
			return err
		}
		addon.Spec.Repositories = append(addon.Spec.Repositories, toSpecRepositories(repository)...)

		obj, err = s.extractor.ToUnstructured(addon)
		if err != nil {
			return err
		}

		_, err = s.addonsCfgClient.Update(context.Background(), obj, metav1.UpdateOptions{})
		return err
	}); err != nil {
		return nil, errors.Wrapf(err, "while updating %s %s", pretty.AddonsConfiguration, name)
	}

	return addon, nil
}

func (s *clusterAddonsConfigurationService) RemoveRepos(name string, reposToRemove []string) (*v1alpha1.ClusterAddonsConfiguration, error) {
	var addon *v1alpha1.ClusterAddonsConfiguration
	if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		obj, err := s.addonsCfgClient.Get(context.Background(), name, metav1.GetOptions{})
		if err != nil {
			return err
		}

		addon, err = s.extractor.FromUnstructured(obj)
		if err != nil {
			return err
		}
		resultRepos := filterOutRepositories(addon.Spec.Repositories, reposToRemove)
		addon.Spec.Repositories = resultRepos

		obj, err = s.extractor.ToUnstructured(addon)
		if err != nil {
			return err
		}

		_, err = s.addonsCfgClient.Update(context.Background(), obj, metav1.UpdateOptions{})
		return err
	}); err != nil {
		return nil, errors.Wrapf(err, "while updating %s %s", pretty.AddonsConfiguration, name)
	}

	return addon, nil
}

func (s *clusterAddonsConfigurationService) Create(name string, repository []*gqlschema.AddonsConfigurationRepositoryInput, labels gqlschema.Labels) (*v1alpha1.ClusterAddonsConfiguration, error) {
	addon := &v1alpha1.ClusterAddonsConfiguration{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterAddonsConfiguration",
			APIVersion: "addons.kyma-project.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: toMapLabels(labels),
		},
		Spec: v1alpha1.ClusterAddonsConfigurationSpec{
			CommonAddonsConfigurationSpec: v1alpha1.CommonAddonsConfigurationSpec{
				Repositories: toSpecRepositories(repository),
			},
		},
	}

	obj, err := s.extractor.ToUnstructured(addon)
	if err != nil {
		return nil, err
	}
	_, err = s.addonsCfgClient.Create(context.Background(), obj, metav1.CreateOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "while creating %s %s", pretty.ClusterAddonsConfiguration, addon.Name)
	}

	return addon, nil
}

func (s *clusterAddonsConfigurationService) Update(name string, repository []*gqlschema.AddonsConfigurationRepositoryInput, labels gqlschema.Labels) (*v1alpha1.ClusterAddonsConfiguration, error) {
	addon, err := s.getClusterAddonsConfiguration(name)
	if err != nil {
		return nil, err
	}
	addon.Spec.Repositories = toSpecRepositories(repository)
	addon.Labels = toMapLabels(labels)

	obj, err := s.extractor.ToUnstructured(addon)
	if err != nil {
		return nil, err
	}
	_, err = s.addonsCfgClient.Update(context.Background(), obj, metav1.UpdateOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "while updating %s %s", pretty.ClusterAddonsConfiguration, addon.Name)
	}

	return addon, nil
}

func (s *clusterAddonsConfigurationService) Delete(name string) (*v1alpha1.ClusterAddonsConfiguration, error) {
	addon, err := s.getClusterAddonsConfiguration(name)
	if err != nil {
		return nil, err
	}

	if err := s.addonsCfgClient.Delete(context.Background(), name, metav1.DeleteOptions{}); err != nil {
		return nil, errors.Wrapf(err, "while deleting %s %s", pretty.ClusterAddonsConfiguration, addon.Name)
	}

	return addon, nil
}

func (s *clusterAddonsConfigurationService) Resync(name string) (*v1alpha1.ClusterAddonsConfiguration, error) {
	var addon *v1alpha1.ClusterAddonsConfiguration
	var err error
	if err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		addon, err = s.getClusterAddonsConfiguration(name)
		if err != nil {
			return err
		}
		addon.Spec.ReprocessRequest++

		obj, err := s.extractor.ToUnstructured(addon)
		if err != nil {
			return err
		}
		_, err = s.addonsCfgClient.Update(context.Background(), obj, metav1.UpdateOptions{})
		return err
	}); err != nil {
		return nil, errors.Wrapf(err, "cannot update ClusterAddonsConfiguration %s", name)
	}

	return addon, nil
}

func (s *clusterAddonsConfigurationService) getClusterAddonsConfiguration(name string) (*v1alpha1.ClusterAddonsConfiguration, error) {
	item, exists, err := s.addonsCfgInformer.GetStore().GetByKey(name)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting %s %s", pretty.ClusterAddonsConfiguration, name)
	}

	if !exists {
		return nil, errors.Errorf("%s doesn't exists", name)
	}

	u, ok := item.(*unstructured.Unstructured)
	if !ok {
		return nil, fmt.Errorf("incorrect item type: %T, should be: *unstructured.Unstructured", item)
	}

	addons, err := s.extractor.FromUnstructured(u)
	if !ok {
		return nil, err
	}

	return addons, nil
}

func (svc *clusterAddonsConfigurationService) Subscribe(listener resource.Listener) {
	svc.addonsNotifier.AddListener(listener)
}

func (svc *clusterAddonsConfigurationService) Unsubscribe(listener resource.Listener) {
	svc.addonsNotifier.DeleteListener(listener)
}
