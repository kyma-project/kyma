package application

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	mappingTypes "github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"

	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/application/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/pager"
	res "github.com/kyma-project/kyma/components/console-backend-service/internal/resource"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/iosafety"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/resource"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/application/extractor"
	"github.com/pkg/errors"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/cache"
)

const (
	appMappingNameIndex = "mapping-name"
	// This regex comes from the k8s resource name validation and has been checked against traversal attack
	// https://github.com/kubernetes/kubernetes/blob/v1.10.1/staging/src/k8s.io/apimachinery/pkg/util/validation/validation.go#L126
	appNameRegex      = `^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])$`
	maxUpdateRetries  = 5
	applicationHeader = "Application"
)

type notifier interface {
	AddListener(observer resource.Listener)
	DeleteListener(observer resource.Listener)
}

// applicationService provides listing namespaces along with Applications.
// It provides also Applications enabling/disabling in given namespace.
type applicationService struct {
	aCli        dynamic.NamespaceableResourceInterface
	appInformer cache.SharedIndexInformer

	mCli            dynamic.NamespaceableResourceInterface
	mappingInformer cache.SharedIndexInformer

	connectorSvcURL  string
	httpClient       *http.Client
	appNameRegex     *regexp.Regexp
	notifier         notifier
	extractor        extractor.ApplicationUnstructuredExtractor
	mappingExtractor extractor.ApplicationMappingUnstructuredExtractor

	appMappingConverter applicationMappingConverter
}

func newApplicationService(cfg Config, aCli dynamic.NamespaceableResourceInterface, mCli dynamic.NamespaceableResourceInterface, mInformer cache.SharedIndexInformer, appInformer cache.SharedIndexInformer) (*applicationService, error) {
	err := mInformer.AddIndexers(cache.Indexers{
		appMappingNameIndex: func(obj interface{}) ([]string, error) {
			m := &mappingTypes.ApplicationMapping{}
			u, err := res.ToUnstructured(obj)
			if err != nil {
				return nil, errors.Wrapf(err, "while converting applicationMapping obj to unstructured")
			}
			err = res.FromUnstructured(u, m)
			if err != nil {
				return nil, errors.Wrapf(err, "while converting unstructured to applicationMapping")
			}
			return []string{m.Name}, nil
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "while adding indexers")
	}

	notifier := resource.NewNotifier()
	appInformer.AddEventHandler(notifier)

	regex, err := regexp.Compile(appNameRegex)
	if err != nil {
		return nil, errors.Wrapf(err, "while compiling %s regex", appNameRegex)
	}
	return &applicationService{
		mCli:            mCli,
		mappingInformer: mInformer,

		aCli:        aCli,
		appInformer: appInformer,

		connectorSvcURL: cfg.Connector.URL,
		httpClient: &http.Client{
			Timeout: cfg.Connector.HTTPCallTimeout,
		},
		notifier:     notifier,
		appNameRegex: regex,
		extractor:    extractor.ApplicationUnstructuredExtractor{},

		appMappingConverter: applicationMappingConverter{},
	}, nil
}

func (svc *applicationService) Create(name string, description string, labels gqlschema.Labels) (*v1alpha1.Application, error) {
	u, err := svc.extractor.ToUnstructured(&v1alpha1.Application{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Application",
			APIVersion: "applicationconnector.kyma-project.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v1alpha1.ApplicationSpec{
			Labels:      labels,
			Description: description,
			Services:    []v1alpha1.Service{},
		},
	})
	if err != nil {
		return &v1alpha1.Application{}, err
	}

	created, err := svc.aCli.Create(u, metav1.CreateOptions{})

	if err != nil {
		return &v1alpha1.Application{}, err
	}
	return svc.extractor.FromUnstructured(created)
}

func (svc *applicationService) Update(name string, description string, labels gqlschema.Labels) (*v1alpha1.Application, error) {
	var lastErr error
	for i := 0; i < maxUpdateRetries; i++ {
		app, err := svc.Find(name)
		if err != nil {
			return nil, errors.Wrapf(err, "while getting %s [%s]", pretty.Application, name)
		}
		if app == nil {
			return nil, apiErrors.NewNotFound(schema.GroupResource{
				Group:    "applicationconnector.kyma-project.io",
				Resource: "applications",
			}, name)
		}
		app.Spec.Description = description
		app.Spec.Labels = labels

		unstructuredApp, err := svc.extractor.ToUnstructured(app)
		if err != nil {
			return &v1alpha1.Application{}, err
		}

		updated, err := svc.aCli.Update(unstructuredApp, metav1.UpdateOptions{})
		switch {
		case err == nil:
			return svc.extractor.FromUnstructured(updated)
		case apiErrors.IsConflict(err):
			lastErr = err
			continue
		default:
			return nil, errors.Wrapf(err, "while updating %s [%s]", pretty.Application, name)
		}
	}
	return nil, errors.Wrapf(lastErr, "couldn't update %s [%s], after %d retries", pretty.Application, name, maxUpdateRetries)
}

func (svc *applicationService) Delete(name string) error {
	return svc.aCli.Delete(name, &metav1.DeleteOptions{})
}

func (svc *applicationService) ListNamespacesFor(appName string) ([]string, error) {
	mappingObjs, err := svc.mappingInformer.GetIndexer().ByIndex(appMappingNameIndex, appName)
	if err != nil {
		return nil, errors.Wrapf(err, "while listing %s by index %q with key %q", pretty.ApplicationMapping, appMappingNameIndex, appName)
	}

	nsList := make([]string, 0, len(mappingObjs))
	for _, item := range mappingObjs {
		unstructured, err := res.ToUnstructured(item)
		if err != nil {
			return nil, errors.Wrapf(err, "while converting mappingObj to unstructured")
		}
		appMapping := &mappingTypes.ApplicationMapping{}
		err = res.FromUnstructured(unstructured, appMapping)
		if err != nil {
			return nil, errors.Wrapf(err, "while converting ApplicationMapping from unstructured")
		}
		nsList = append(nsList, appMapping.Namespace)
	}

	return nsList, nil
}

func (svc *applicationService) Find(name string) (*v1alpha1.Application, error) {
	item, exists, err := svc.appInformer.GetStore().GetByKey(name)

	if err != nil || !exists {
		return nil, err
	}

	app, ok := item.(*unstructured.Unstructured)
	if !ok {
		return nil, fmt.Errorf("incorrect item type: %T, should be: 'Application' in version 'v1alpha1'", item)
	}

	return svc.extractor.FromUnstructured(app)
}

func (svc *applicationService) List(params pager.PagingParams) ([]*v1alpha1.Application, error) {
	items, err := pager.From(svc.appInformer.GetStore()).Limit(params)
	if err != nil {
		return nil, errors.Wrapf(err, "while listing %s with paging params [first: %v] [offset: %v]", pretty.Application, params.First, params.Offset)
	}

	res := make([]*v1alpha1.Application, 0, len(items))
	for _, item := range items {
		re, err := svc.extractor.Do(item)
		if err != nil {
			return nil, fmt.Errorf("cannot convert item to 'v1alpha1.Application': %v", item)
		}

		res = append(res, re)
	}

	return res, nil
}

func (svc *applicationService) ListInNamespace(namespace string) ([]*v1alpha1.Application, error) {
	mappings, err := svc.mappingInformer.GetIndexer().ByIndex("namespace", namespace)
	if err != nil {
		return nil, errors.Wrapf(err, "while listing %s", pretty.ApplicationMapping)
	}

	res := make([]*v1alpha1.Application, 0)
	for _, item := range mappings {

		mapping, err := svc.mappingExtractor.Do(item)
		if err != nil {
			return nil, fmt.Errorf("cannot convert item to 'applicationMapping': %v", item)
		}

		// Application CR is cluster wide so the key is only the name
		item, exists, err := svc.appInformer.GetIndexer().GetByKey(mapping.Name)
		if err != nil {
			return nil, errors.Wrapf(err, "while getting %s with key %s", pretty.Application, mapping.Name)
		}

		if !exists {
			glog.Warningf("Found %s %q in namespaces %q but %s with name %q does not exists", pretty.ApplicationMapping, mapping.Name, mapping.Namespace, pretty.Application, mapping.Name)
			continue
		}

		app, err := svc.extractor.Do(item)
		if err != nil {
			return nil, fmt.Errorf("cannot convert item to 'v1alpha1.Application': %v", item)
		}

		//TODO: Write test to make sure that this is a real deep copy
		deepCopy := app.DeepCopy()
		res = append(res, deepCopy)
	}
	return res, nil
}

// Enable enables Application in given namespace by creating ApplicationMapping
func (svc *applicationService) Enable(namespace, name string, services []*gqlschema.ApplicationMappingService) (*mappingTypes.ApplicationMapping, error) {
	mappingServices := svc.appMappingConverter.transformApplicationMappingServiceFromGQL(services)
	m, err := svc.mappingExtractor.ToUnstructured(&mappingTypes.ApplicationMapping{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ApplicationMapping",
			APIVersion: "applicationconnector.kyma-project.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: mappingTypes.ApplicationMappingSpec{
			Services: mappingServices,
		},
	})

	if err != nil {
		return &mappingTypes.ApplicationMapping{}, err
	}

	created, err := svc.mCli.Namespace(namespace).Create(m, metav1.CreateOptions{})

	if err != nil {
		return &mappingTypes.ApplicationMapping{}, err
	}
	return svc.mappingExtractor.FromUnstructured(created)
}

// UpdateApplicationMapping updates ApplicationMapping based on its name and namespace
func (svc *applicationService) UpdateApplicationMapping(namespace, name string, services []*gqlschema.ApplicationMappingService) (*mappingTypes.ApplicationMapping, error) {
	key := fmt.Sprintf("%s/%s", namespace, name)
	item, exists, err := svc.mappingInformer.GetStore().GetByKey(key)
	if err != nil {
		return nil, errors.Wrapf(err, "while fetching %s", pretty.ApplicationMapping)
	}
	if !exists {
		return nil, errors.Wrapf(err, "mapping %s not found", pretty.ApplicationMapping)
	}

	em, err := svc.mappingExtractor.Do(item)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting %s", pretty.ApplicationMapping)
	}

	emUpdate := em.DeepCopy()
	emUpdate.Spec.Services = svc.appMappingConverter.transformApplicationMappingServiceFromGQL(services)

	u, err := svc.mappingExtractor.ToUnstructured(emUpdate)
	if err != nil {
		return nil, err
	}

	updated, err := svc.mCli.Namespace(namespace).Update(u, metav1.UpdateOptions{})

	if err != nil {
		return nil, errors.Wrapf(err, "while updating %s [%s]", pretty.ApplicationMapping, name)
	}

	return svc.mappingExtractor.FromUnstructured(updated)
}

// ListApplicationMapping return list of ApplicationMapping from all namespaces base on name
func (svc *applicationService) ListApplicationMapping(name string) ([]*mappingTypes.ApplicationMapping, error) {

	mappings := svc.mappingInformer.GetStore().List()

	result := make([]*mappingTypes.ApplicationMapping, 0)
	for _, item := range mappings {
		mapping, err := svc.mappingExtractor.Do(item)
		if err != nil {
			return nil, fmt.Errorf("cannot convert item to 'applicationMapping': %v", item)
		}
		if mapping.Name == name {
			result = append(result, mapping)
		}
	}

	return result, nil
}

// Disable disables Application in given namespace by removing ApplicationMapping
func (svc *applicationService) Disable(namespace, name string) error {
	return svc.mCli.Namespace(namespace).Delete(name, nil)
}

func (svc *applicationService) GetConnectionURL(appName string) (string, error) {
	if ok := svc.appNameRegex.MatchString(appName); !ok {
		return "", fmt.Errorf("%s name %q does not match regex: %s", pretty.Application, appName, appNameRegex)
	}
	reqURL := fmt.Sprintf("%s/v1/applications/tokens", svc.connectorSvcURL)

	req, err := http.NewRequest(http.MethodPost, reqURL, nil)
	if err != nil {
		return "", errors.Wrap(err, "while creating HTTP request")
	}

	req.Header.Set(applicationHeader, appName)

	resp, err := svc.httpClient.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "while making HTTP call")
	}
	defer svc.drainAndCloseBody(resp.Body)

	if resp.StatusCode != http.StatusCreated {
		cause := svc.extractErrorCause(resp.Body)
		return "", errors.Wrapf(cause, "while requesting connection URL obtained unexpected status code %d", resp.StatusCode)
	}

	connectorURL, err := svc.extractConnectionURL(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "while extracting connection URL from body")
	}

	return connectorURL, nil
}

func (svc *applicationService) extractConnectionURL(body io.ReadCloser) (string, error) {
	var dto struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(body).Decode(&dto); err != nil {
		return "", errors.Wrap(err, "while decoding json")
	}

	return dto.URL, nil
}

func (svc *applicationService) extractErrorCause(body io.ReadCloser) error {
	var dto struct {
		Error string `json:"error"`
	}
	if err := json.NewDecoder(body).Decode(&dto); err != nil {
		return errors.Wrap(err, "while decoding json to get error msg from body")
	}

	return errors.New(dto.Error)
}

func (svc *applicationService) drainAndCloseBody(body io.ReadCloser) {
	err := iosafety.DrainReader(body)
	if err != nil {
		glog.Errorf("Unable to drain body reader. Cause: %v", err)
	}
	err = body.Close()
	if err != nil {
		glog.Errorf("Unable to close body reader. Cause: %v", err)
	}
}

func (svc *applicationService) Subscribe(listener resource.Listener) {
	svc.notifier.AddListener(listener)
}

func (svc *applicationService) Unsubscribe(listener resource.Listener) {
	svc.notifier.DeleteListener(listener)
}
