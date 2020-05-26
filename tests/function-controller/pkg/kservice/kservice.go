package kservice

import (
	"time"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/resource"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/shared"
)

type KService struct {
	resCli      *resource.Resource
	name        string
	namespace   string
	waitTimeout time.Duration
	log         shared.Logger
	verbose     bool
}

func New(name string, c shared.Container) *KService {
	return &KService{
		resCli:      resource.New(c.DynamicCli, servingv1.SchemeGroupVersion.WithResource("services"), c.Namespace, c.Log, c.Verbose),
		name:        name,
		namespace:   c.Namespace,
		waitTimeout: c.WaitTimeout,
		log:         c.Log,
		verbose:     c.Verbose,
	}
}

func (k *KService) Get() (*servingv1.Service, error) {
	u, err := k.resCli.Get(k.name)
	if err != nil {
		return &servingv1.Service{}, errors.Wrapf(err, "while getting Knative Service %s in namespace %s", k.name, k.namespace)
	}

	srv, err := convertFromUnstructuredToKService(*u)
	if err != nil {
		return &servingv1.Service{}, err
	}

	return &srv, nil
}

func convertFromUnstructuredToKService(u unstructured.Unstructured) (servingv1.Service, error) {
	srv := servingv1.Service{}
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &srv)
	return srv, err
}
