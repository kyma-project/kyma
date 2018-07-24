package remoteenvironment

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/remote-environment-broker/pkg/apis/remoteenvironment/v1alpha1"
	remoteenvironmentv1alpha1 "github.com/kyma-project/kyma/components/remote-environment-broker/pkg/client/clientset/versioned/typed/remoteenvironment/v1alpha1"
	reMappinglister "github.com/kyma-project/kyma/components/remote-environment-broker/pkg/client/listers/remoteenvironment/v1alpha1"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/pager"
	"github.com/kyma-project/kyma/components/ui-api-layer/pkg/iosafety"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

const (
	remoteMappingNameIndex = "mapping-name"
	// This regex comes from the k8s resource name validation and has been checked against traversal attack
	// https://github.com/kubernetes/kubernetes/blob/v1.10.1/staging/src/k8s.io/apimachinery/pkg/util/validation/validation.go#L126
	reNameRegex = `^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])$`
)

// remoteEnvironmentService provides listing environments along with remote environments.
// It provides also remote environment enabling/disabling in given namespace.
type remoteEnvironmentService struct {
	client          remoteenvironmentv1alpha1.RemoteenvironmentV1alpha1Interface
	mappingLister   reMappinglister.EnvironmentMappingLister
	mappingInformer cache.SharedIndexInformer
	reInformer      cache.SharedIndexInformer
	connectorSvcURL string
	httpClient      *http.Client
	reNameRegex     *regexp.Regexp
}

func newRemoteEnvironmentService(client remoteenvironmentv1alpha1.RemoteenvironmentV1alpha1Interface, cfg Config, mappingInformer cache.SharedIndexInformer, mappingLister reMappinglister.EnvironmentMappingLister, reInformer cache.SharedIndexInformer) (*remoteEnvironmentService, error) {
	mappingInformer.AddIndexers(cache.Indexers{
		remoteMappingNameIndex: func(obj interface{}) ([]string, error) {
			mapping, ok := obj.(*v1alpha1.EnvironmentMapping)
			if !ok {
				return nil, errors.New("cannot convert item")
			}

			return []string{mapping.Name}, nil
		},
	})

	regex, err := regexp.Compile(reNameRegex)
	if err != nil {
		return nil, errors.Wrapf(err, "while compiling %s regex", reNameRegex)
	}
	return &remoteEnvironmentService{
		mappingLister:   mappingLister,
		mappingInformer: mappingInformer,
		reInformer:      reInformer,
		client:          client,
		connectorSvcURL: cfg.Connector.URL,
		httpClient: &http.Client{
			Timeout: cfg.Connector.HTTPCallTimeout,
		},
		reNameRegex: regex,
	}, nil
}

func (svc *remoteEnvironmentService) ListNamespacesFor(reName string) ([]string, error) {
	mappingObjs, err := svc.mappingInformer.GetIndexer().ByIndex(remoteMappingNameIndex, reName)
	if err != nil {
		return nil, errors.Wrapf(err, "while listing environment mappings by index %q with key %q", remoteMappingNameIndex, reName)
	}

	nsList := make([]string, 0, len(mappingObjs))
	for _, item := range mappingObjs {
		reMapping, ok := item.(*v1alpha1.EnvironmentMapping)
		if !ok {
			return nil, fmt.Errorf("incorrect item type: %T, should be: 'EnvironmentMapping' in version 'v1alpha1'", item)
		}
		nsList = append(nsList, reMapping.Namespace)
	}

	return nsList, nil
}

func (svc *remoteEnvironmentService) Find(name string) (*v1alpha1.RemoteEnvironment, error) {
	remoteEnvironment, exists, err := svc.reInformer.GetStore().GetByKey(name)

	if err != nil || !exists {
		return nil, err
	}

	re, ok := remoteEnvironment.(*v1alpha1.RemoteEnvironment)
	if !ok {

		return nil, fmt.Errorf("Cannot process RemoteEnvironment")
	}

	return re, nil
}

func (svc *remoteEnvironmentService) List(params pager.PagingParams) ([]*v1alpha1.RemoteEnvironment, error) {
	reObjs, err := pager.From(svc.reInformer.GetStore()).Limit(params)
	if err != nil {
		return nil, errors.Wrapf(err, "while listing remote environments with paging params [first: %v] [offset: %v]: %v", params.First, params.Offset)
	}

	res := make([]*v1alpha1.RemoteEnvironment, 0, len(reObjs))
	for _, item := range reObjs {
		re, ok := item.(*v1alpha1.RemoteEnvironment)
		if !ok {
			return nil, fmt.Errorf("incorrect item type: %T, should be: 'RemoteEnvironment' in version 'v1alpha1'", item)
		}

		res = append(res, re)
	}

	return res, nil
}

func (svc *remoteEnvironmentService) ListInEnvironment(environment string) ([]*v1alpha1.RemoteEnvironment, error) {
	mappings, err := svc.mappingLister.EnvironmentMappings(environment).List(labels.Everything())
	if err != nil {
		return nil, errors.Wrap(err, "while listing environment mappings")
	}

	res := make([]*v1alpha1.RemoteEnvironment, 0)
	for _, mapping := range mappings {
		// Remote Environment CR is cluster wide so the key is only the name
		reObj, exists, err := svc.reInformer.GetIndexer().GetByKey(mapping.Name)
		if err != nil {
			return nil, errors.Wrapf(err, "while getting remote environment with key %s", mapping.Name)
		}

		if !exists {
			glog.Warningf("Found environment mapping %q in namespaces %q but remote environment with name %q does not exists", mapping.Name, mapping.Namespace, mapping.Name)
			continue
		}

		reCR, ok := reObj.(*v1alpha1.RemoteEnvironment)
		if !ok {
			return nil, fmt.Errorf("incorrect item type: %T, should be: 'RemoteEnvironment' in version 'v1alpha1'", reObj)
		}

		//TODO: Write test to make sure that this is a real deep copy
		deepCopy := reCR.DeepCopy()
		res = append(res, deepCopy)
	}
	return res, nil
}

// Enable enables remote environment in given namespace by creating EnvironmentMappinggo
func (svc *remoteEnvironmentService) Enable(namespace, name string) (*v1alpha1.EnvironmentMapping, error) {
	em, err := svc.client.EnvironmentMappings(namespace).Create(&v1alpha1.EnvironmentMapping{
		TypeMeta: metav1.TypeMeta{
			Kind:       "EnvironmentMapping",
			APIVersion: "remoteenvironment.kyma.cx/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	})
	return em, err
}

// Disable disables remote environment in given namespace by removing EnvironmentMapping
func (svc *remoteEnvironmentService) Disable(namespace, name string) error {
	return svc.client.EnvironmentMappings(namespace).Delete(name, &metav1.DeleteOptions{})
}

func (svc *remoteEnvironmentService) GetConnectionUrl(remoteEnvironment string) (string, error) {
	if ok := svc.reNameRegex.MatchString(remoteEnvironment); !ok {
		return "", fmt.Errorf("Remote evironment name %q does not match regex: %s", remoteEnvironment, reNameRegex)
	}
	reqURL := fmt.Sprintf("%s/v1/remoteenvironments/%s/tokens", svc.connectorSvcURL, remoteEnvironment)

	req, err := http.NewRequest(http.MethodPost, reqURL, nil)
	if err != nil {
		return "", errors.Wrap(err, "while creating HTTP request")
	}

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

func (svc *remoteEnvironmentService) extractConnectionURL(body io.ReadCloser) (string, error) {
	var dto struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(body).Decode(&dto); err != nil {
		return "", errors.Wrap(err, "while decoding json")
	}

	return dto.URL, nil
}

func (svc *remoteEnvironmentService) extractErrorCause(body io.ReadCloser) error {
	var dto struct {
		Error string `json:"error"`
	}
	if err := json.NewDecoder(body).Decode(&dto); err != nil {
		return errors.Wrap(err, "while decoding json to get error msg from body")
	}

	return errors.New(dto.Error)
}

func (svc *remoteEnvironmentService) drainAndCloseBody(body io.ReadCloser) {
	_ = iosafety.DrainReader(body)
	body.Close()
}
