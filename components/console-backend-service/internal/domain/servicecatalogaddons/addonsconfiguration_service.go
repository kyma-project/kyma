package servicecatalogaddons

import (
	"fmt"
	"strings"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalogaddons/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/pager"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/resource"
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

type addonsConfigurationService struct {
	informer     cache.SharedIndexInformer
	cfgMapClient corev1.ConfigMapInterface
	notifier     notifier
}

func newAddonsConfigurationService(informer cache.SharedIndexInformer, cfgMapClient corev1.ConfigMapInterface) *addonsConfigurationService {
	notifier := resource.NewNotifier()
	informer.AddEventHandler(notifier)
	return &addonsConfigurationService{
		informer:     informer,
		cfgMapClient: cfgMapClient,
		notifier:     notifier,
	}
}

func (s *addonsConfigurationService) List(pagingParams pager.PagingParams) ([]*v1.ConfigMap, error) {
	items, err := pager.From(s.informer.GetStore()).Limit(pagingParams)
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

func (s *addonsConfigurationService) AddRepos(name string, urls []string) (*v1.ConfigMap, error) {
	configMap, err := s.getAddonsConfiguration(name)
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

func (s *addonsConfigurationService) joinURLs(base string, newUrls []string) string {
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

func (s *addonsConfigurationService) RemoveRepos(name string, urls []string) (*v1.ConfigMap, error) {
	configMap, err := s.getAddonsConfiguration(name)
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

func (svc *addonsConfigurationService) removeURL(s []string, r string) string {
	for i, v := range s {
		if v == r {
			return strings.Join(append(s[:i], s[i+1:]...), "\n")
		}
	}
	return strings.Join(s, "\n")
}

func (s *addonsConfigurationService) Create(name string, urls []string, labels *gqlschema.Labels) (*v1.ConfigMap, error) {
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

func (s *addonsConfigurationService) Update(name string, urls []string, labels *gqlschema.Labels) (*v1.ConfigMap, error) {
	configMap, err := s.getAddonsConfiguration(name)
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

func (s *addonsConfigurationService) setLabels(givenLabels *gqlschema.Labels) map[string]string {
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

func (s *addonsConfigurationService) Delete(name string) (*v1.ConfigMap, error) {
	cfg, err := s.getAddonsConfiguration(name)
	if err != nil {
		return nil, err
	}

	if err := s.cfgMapClient.Delete(name, &metav1.DeleteOptions{}); err != nil {
		return nil, errors.Wrapf(err, "while deleting %s", pretty.AddonsConfiguration)
	}

	return cfg, nil
}

func (s *addonsConfigurationService) getAddonsConfiguration(name string) (*v1.ConfigMap, error) {
	key := fmt.Sprintf("%s/%s", systemNs, name)
	item, exists, err := s.informer.GetIndexer().GetByKey(key)
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

func (svc *addonsConfigurationService) Subscribe(listener resource.Listener) {
	svc.notifier.AddListener(listener)
}

func (svc *addonsConfigurationService) Unsubscribe(listener resource.Listener) {
	svc.notifier.DeleteListener(listener)
}
