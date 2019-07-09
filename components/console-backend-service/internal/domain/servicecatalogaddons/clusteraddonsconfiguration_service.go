package servicecatalogaddons

import (
	"fmt"
	"strings"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalogaddons/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/pager"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/resource"
	"github.com/kyma-project/kyma/components/helm-broker/pkg/apis/addons/v1alpha1"
	addonsClientset "github.com/kyma-project/kyma/components/helm-broker/pkg/client/clientset/versioned/typed/addons/v1alpha1"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
)

const (
	addonsCfgKey        = "URLs"
	addonsCfgLabelValue = "true"
	addonsCfgLabelKey   = "helm-broker-repo"

	systemNs = "kyma-system"
)

type clusterAddonsConfigurationService struct {
	cfgMapinformer    cache.SharedIndexInformer
	cfgMapClient      corev1.ConfigMapInterface
	cmNotifier        notifier
	addonsNotifier    notifier
	addonsCfgClient   addonsClientset.AddonsV1alpha1Interface
	addonsCfgInformer cache.SharedIndexInformer
}

func newClusterAddonsConfigurationService(cmInformer cache.SharedIndexInformer, addonsCfgInformer cache.SharedIndexInformer, cfgMapClient corev1.ConfigMapInterface, addonsCfgClient addonsClientset.AddonsV1alpha1Interface) *clusterAddonsConfigurationService {
	cmNotifier := resource.NewNotifier()
	addonsNotifier := resource.NewNotifier()

	if cmInformer != nil {
		cmInformer.AddEventHandler(cmNotifier)
	}

	if addonsCfgInformer != nil {
		addonsCfgInformer.AddEventHandler(addonsNotifier)
	}

	return &clusterAddonsConfigurationService{
		cfgMapinformer:    cmInformer,
		cfgMapClient:      cfgMapClient,
		addonsCfgClient:   addonsCfgClient,
		addonsCfgInformer: addonsCfgInformer,
		cmNotifier:        cmNotifier,
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
			return nil, fmt.Errorf("incorrect item type: %T, should be: *ConfigMap", item)
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
		return nil, errors.Wrapf(err, "while updating %s %s", pretty.AddonsConfiguration, addonCpy.Name)
	}

	return result, nil
}

func (s *clusterAddonsConfigurationService) RemoveRepos(name string, urls []string) (*v1alpha1.ClusterAddonsConfiguration, error) {
	addon, err := s.getClusterAddonsConfiguration(name)
	if err != nil {
		return nil, err
	}

	resultRepos := s.filterOutRepositories(addon.Spec.Repositories, urls)

	addonCpy := addon.DeepCopy()
	addonCpy.Spec.Repositories = resultRepos

	updatedAddon, err := s.addonsCfgClient.ClusterAddonsConfigurations().Update(addonCpy)
	if err != nil {
		return nil, errors.Wrapf(err, "while updating %s %s", pretty.AddonsConfiguration, addonCpy.Name)
	}

	return updatedAddon, nil
}

func (s *clusterAddonsConfigurationService) Create(name string, urls []string, labels *gqlschema.Labels) (*v1alpha1.ClusterAddonsConfiguration, error) {
	addon := &v1alpha1.ClusterAddonsConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: s.toMapLabels(labels),
		},
		Spec: v1alpha1.ClusterAddonsConfigurationSpec{
			CommonAddonsConfigurationSpec: v1alpha1.CommonAddonsConfigurationSpec{
				Repositories: s.toSpecRepositories(urls),
			},
		},
	}

	result, err := s.addonsCfgClient.ClusterAddonsConfigurations().Create(addon)
	if err != nil {
		return nil, errors.Wrapf(err, "while creating %s %s", pretty.AddonsConfiguration, addon.Name)
	}

	return result, nil
}

func (s *clusterAddonsConfigurationService) Update(name string, urls []string, labels *gqlschema.Labels) (*v1alpha1.ClusterAddonsConfiguration, error) {
	addon, err := s.getClusterAddonsConfiguration(name)
	if err != nil {
		return nil, err
	}

	addonCpy := addon.DeepCopy()
	addonCpy.Spec.Repositories = s.toSpecRepositories(urls)
	addonCpy.Labels = s.toMapLabels(labels)

	result, err := s.addonsCfgClient.ClusterAddonsConfigurations().Update(addonCpy)
	if err != nil {
		return nil, errors.Wrapf(err, "while updating %s %s", pretty.AddonsConfiguration, addon.Name)
	}

	return result, nil
}

func (s *clusterAddonsConfigurationService) Delete(name string) (*v1alpha1.ClusterAddonsConfiguration, error) {
	addon, err := s.getClusterAddonsConfiguration(name)
	if err != nil {
		return nil, err
	}

	if err := s.addonsCfgClient.ClusterAddonsConfigurations().Delete(name, &metav1.DeleteOptions{}); err != nil {
		return nil, errors.Wrapf(err, "while deleting %s %s", pretty.AddonsConfiguration, addon.Name)
	}

	return addon, nil
}

func (s *clusterAddonsConfigurationService) filterOutRepositories(repos []v1alpha1.SpecRepository, urls []string) []v1alpha1.SpecRepository {
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

func (s *clusterAddonsConfigurationService) toMapLabels(givenLabels *gqlschema.Labels) map[string]string {
	if givenLabels == nil {
		return nil
	}

	labels := map[string]string{}
	for k, v := range *givenLabels {
		labels[k] = v
	}
	return labels
}

func (s *clusterAddonsConfigurationService) toSpecRepositories(urls []string) []v1alpha1.SpecRepository {
	var repos []v1alpha1.SpecRepository
	for _, u := range urls {
		repos = append(repos, v1alpha1.SpecRepository{URL: u})
	}

	return repos
}

func (s *clusterAddonsConfigurationService) getClusterAddonsConfiguration(name string) (*v1alpha1.ClusterAddonsConfiguration, error) {
	item, exists, err := s.addonsCfgInformer.GetStore().GetByKey(name)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting %s %s", pretty.AddonsConfiguration, name)
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

// Deprecated, the ClusterAddonsConfiguration should be used instead
func (s *clusterAddonsConfigurationService) ListConfigMaps(pagingParams pager.PagingParams) ([]*v1.ConfigMap, error) {
	items, err := pager.From(s.cfgMapinformer.GetStore()).Limit(pagingParams)
	if err != nil {
		return nil, err
	}

	var configMaps []*v1.ConfigMap
	for _, item := range items {
		configMap, ok := item.(*v1.ConfigMap)
		if !ok {
			return nil, fmt.Errorf("incorrect item type: %T, should be: *ConfigMap", item)
		}

		configMaps = append(configMaps, configMap)
	}

	return configMaps, nil
}

// Deprecated, the ClusterAddonsConfiguration should be used instead
func (s *clusterAddonsConfigurationService) AddReposToConfigMap(name string, urls []string) (*v1.ConfigMap, error) {
	configMap, err := s.getConfigMapAddonsConfiguration(name)
	if err != nil {
		return nil, err
	}

	configMap.Data[addonsCfgKey] = s.joinURLs(configMap.Data[addonsCfgKey], urls)

	result, err := s.cfgMapClient.Update(configMap)
	if err != nil {
		return nil, errors.Wrapf(err, "while updating %s %s", pretty.AddonsConfiguration, configMap.Name)
	}

	return result, nil
}

func (s *clusterAddonsConfigurationService) joinURLs(base string, newUrls []string) string {
	builder := strings.Builder{}
	builder.WriteString(base)

	for _, s := range newUrls {
		if strings.HasSuffix(builder.String(), "\n") {
			builder.WriteString(fmt.Sprintf("%s", s))
		} else {
			builder.WriteString(fmt.Sprintf("\n%s", s))
		}
	}
	return builder.String()
}

// Deprecated, the ClusterAddonsConfiguration should be used instead
func (s *clusterAddonsConfigurationService) RemoveReposFromConfigMap(name string, urls []string) (*v1.ConfigMap, error) {
	configMap, err := s.getConfigMapAddonsConfiguration(name)
	if err != nil {
		return nil, err
	}

	for _, url := range urls {
		configMap.Data[addonsCfgKey] = s.removeURL(strings.Split(configMap.Data[addonsCfgKey], "\n"), url)
	}

	result, err := s.cfgMapClient.Update(configMap)
	if err != nil {
		return nil, errors.Wrapf(err, "while updating %s %s", pretty.AddonsConfiguration, configMap.Name)
	}

	return result, nil
}

func (svc *clusterAddonsConfigurationService) removeURL(s []string, r string) string {
	for i, v := range s {
		if v == r {
			return strings.Join(append(s[:i], s[i+1:]...), "\n")
		}
	}
	return strings.Join(s, "\n")
}

// Deprecated, the ClusterAddonsConfiguration should be used instead
func (s *clusterAddonsConfigurationService) CreateConfigMap(name string, urls []string, labels *gqlschema.Labels) (*v1.ConfigMap, error) {
	configMap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: systemNs,
		},
		Data: map[string]string{
			"URLs": strings.Join(urls, "\n"),
		},
	}
	configMap.Labels = s.setLabels(labels)

	result, err := s.cfgMapClient.Create(configMap)
	if err != nil {
		return nil, errors.Wrap(err, "while creating addons configuration")
	}

	return result, nil
}

// Deprecated, the ClusterAddonsConfiguration should be used instead
func (s *clusterAddonsConfigurationService) UpdateConfigMap(name string, urls []string, labels *gqlschema.Labels) (*v1.ConfigMap, error) {
	configMap, err := s.getConfigMapAddonsConfiguration(name)
	if err != nil {
		return nil, err
	}
	if len(urls) > 0 {
		configMap.Data[addonsCfgKey] = strings.Join(urls, "\n")
	}
	configMap.Labels = s.setLabels(labels)

	result, err := s.cfgMapClient.Update(configMap)
	if err != nil {
		return nil, errors.Wrapf(err, "while updating %s %s", pretty.AddonsConfiguration, configMap.Name)
	}

	return result, nil
}

func (s *clusterAddonsConfigurationService) setLabels(givenLabels *gqlschema.Labels) map[string]string {
	labels := map[string]string{
		addonsCfgLabelKey: addonsCfgLabelValue,
	}
	if givenLabels != nil {
		for k, v := range *givenLabels {
			if k == addonsCfgLabelKey {
				continue
			}
			labels[k] = v
		}
	}
	return labels
}

// Deprecated, the ClusterAddonsConfiguration should be used instead
func (s *clusterAddonsConfigurationService) DeleteConfigMap(name string) (*v1.ConfigMap, error) {
	cfg, err := s.getConfigMapAddonsConfiguration(name)
	if err != nil {
		return nil, err
	}

	if err := s.cfgMapClient.Delete(name, &metav1.DeleteOptions{}); err != nil {
		return nil, errors.Wrapf(err, "while deleting %s", pretty.AddonsConfiguration)
	}

	return cfg, nil
}

// Deprecated, the ClusterAddonsConfiguration should be used instead
func (s *clusterAddonsConfigurationService) getConfigMapAddonsConfiguration(name string) (*v1.ConfigMap, error) {
	key := fmt.Sprintf("%s/%s", systemNs, name)
	item, exists, err := s.cfgMapinformer.GetIndexer().GetByKey(key)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting %s %s", pretty.AddonsConfiguration, name)
	}
	if !exists {
		return nil, errors.Errorf("%s doesn't exists", key)
	}
	configMap, ok := item.(*v1.ConfigMap)
	if !ok {
		return nil, fmt.Errorf("incorrect item type: %T, should be: *ConfigMap", item)
	}

	return configMap, nil
}

// Deprecated, the ClusterAddonsConfiguration should be used instead
func (svc *clusterAddonsConfigurationService) ConfigMapSubscribe(listener resource.Listener) {
	svc.cmNotifier.AddListener(listener)
}

// Deprecated, the ClusterAddonsConfiguration should be used instead
func (svc *clusterAddonsConfigurationService) ConfigMapUnsubscribe(listener resource.Listener) {
	svc.cmNotifier.DeleteListener(listener)
}
