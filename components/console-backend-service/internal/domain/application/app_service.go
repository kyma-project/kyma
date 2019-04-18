package application

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"

	mappingTypes "github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	mappingCli "github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned/typed/applicationconnector/v1alpha1"
	mappingLister "github.com/kyma-project/kyma/components/application-broker/pkg/client/listers/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	appCli "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned/typed/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/application/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/pager"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/iosafety"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/resource"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
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
	aCli        appCli.ApplicationconnectorV1alpha1Interface
	appInformer cache.SharedIndexInformer

	mCli            mappingCli.ApplicationconnectorV1alpha1Interface
	mappingLister   mappingLister.ApplicationMappingLister
	mappingInformer cache.SharedIndexInformer

	connectorSvcURL string
	httpClient      *http.Client
	appNameRegex    *regexp.Regexp
	notifier        notifier
}

func newApplicationService(cfg Config, aCli appCli.ApplicationconnectorV1alpha1Interface, mCli mappingCli.ApplicationconnectorV1alpha1Interface, mInformer cache.SharedIndexInformer, mLister mappingLister.ApplicationMappingLister, appInformer cache.SharedIndexInformer) (*applicationService, error) {
	err := mInformer.AddIndexers(cache.Indexers{
		appMappingNameIndex: func(obj interface{}) ([]string, error) {
			mapping, ok := obj.(*mappingTypes.ApplicationMapping)
			if !ok {
				return nil, errors.New("cannot convert item")
			}

			return []string{mapping.Name}, nil
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
		mappingLister:   mLister,
		mappingInformer: mInformer,

		aCli:        aCli,
		appInformer: appInformer,

		connectorSvcURL: cfg.Connector.URL,
		httpClient: &http.Client{
			Timeout: cfg.Connector.HTTPCallTimeout,
		},
		notifier:     notifier,
		appNameRegex: regex,
	}, nil
}

func (svc *applicationService) Create(name string, description string, labels gqlschema.Labels) (*v1alpha1.Application, error) {
	return svc.aCli.Applications().Create(&v1alpha1.Application{
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

		updated, err := svc.aCli.Applications().Update(app)
		switch {
		case err == nil:
			return updated, nil
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
	return svc.aCli.Applications().Delete(name, &metav1.DeleteOptions{})
}

func (svc *applicationService) ListNamespacesFor(appName string) ([]string, error) {
	mappingObjs, err := svc.mappingInformer.GetIndexer().ByIndex(appMappingNameIndex, appName)
	if err != nil {
		return nil, errors.Wrapf(err, "while listing %s by index %q with key %q", pretty.ApplicationMapping, appMappingNameIndex, appName)
	}

	nsList := make([]string, 0, len(mappingObjs))
	for _, item := range mappingObjs {
		appMapping, ok := item.(*mappingTypes.ApplicationMapping)
		if !ok {
			return nil, fmt.Errorf("incorrect item type: %T, should be: 'ApplicationMapping' in version 'v1alpha1'", item)
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

	app, ok := item.(*v1alpha1.Application)
	if !ok {
		return nil, fmt.Errorf("incorrect item type: %T, should be: 'Application' in version 'v1alpha1'", item)
	}

	return app, nil
}

func (svc *applicationService) List(params pager.PagingParams) ([]*v1alpha1.Application, error) {
	items, err := pager.From(svc.appInformer.GetStore()).Limit(params)
	if err != nil {
		return nil, errors.Wrapf(err, "while listing %s with paging params [first: %v] [offset: %v]", pretty.Application, params.First, params.Offset)
	}

	res := make([]*v1alpha1.Application, 0, len(items))
	for _, item := range items {
		re, ok := item.(*v1alpha1.Application)
		if !ok {
			return nil, fmt.Errorf("incorrect item type: %T, should be: 'Application' in version 'v1alpha1'", item)
		}

		res = append(res, re)
	}

	return res, nil
}

func (svc *applicationService) ListInNamespace(namespace string) ([]*v1alpha1.Application, error) {
	mappings, err := svc.mappingLister.ApplicationMappings(namespace).List(labels.Everything())
	if err != nil {
		return nil, errors.Wrapf(err, "while listing %s", pretty.ApplicationMapping)
	}

	res := make([]*v1alpha1.Application, 0)
	for _, mapping := range mappings {
		// Application CR is cluster wide so the key is only the name
		item, exists, err := svc.appInformer.GetIndexer().GetByKey(mapping.Name)
		if err != nil {
			return nil, errors.Wrapf(err, "while getting %s with key %s", pretty.Application, mapping.Name)
		}

		if !exists {
			glog.Warningf("Found %s %q in namespaces %q but %s with name %q does not exists", pretty.ApplicationMapping, pretty.Application, mapping.Name, mapping.Namespace, mapping.Name)
			continue
		}

		app, ok := item.(*v1alpha1.Application)
		if !ok {
			return nil, fmt.Errorf("incorrect item type: %T, should be: 'Application' in version 'v1alpha1'", item)
		}

		//TODO: Write test to make sure that this is a real deep copy
		deepCopy := app.DeepCopy()
		res = append(res, deepCopy)
	}
	return res, nil
}

// Enable enables Application in given namespace by creating ApplicationMapping
func (svc *applicationService) Enable(namespace, name string) (*mappingTypes.ApplicationMapping, error) {
	em, err := svc.mCli.ApplicationMappings(namespace).Create(&mappingTypes.ApplicationMapping{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ApplicationMapping",
			APIVersion: "applicationconnector.kyma-project.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	})
	return em, err
}

// Disable disables Application in given namespace by removing ApplicationMapping
func (svc *applicationService) Disable(namespace, name string) error {
	return svc.mCli.ApplicationMappings(namespace).Delete(name, &metav1.DeleteOptions{})
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
