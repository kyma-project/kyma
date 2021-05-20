package gateway

import (
	"context"
	"fmt"
	"time"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/application/pretty"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	apiv1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/watch"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
)

const (
	// NOTE: if you are not familiar with the concept then please read contract described in comment below
	applicationLabelName = "application"
	externalAPIPortName  = "http-api-port"
)

/*
The contract, which describes, how to find service for given Application:

1. K8s service is labelled with key `applicationLabelName`, value is the name of the Application
2. K8s Service contains one port with name “http-api-port” and this port is used for status check.
3. K8s Service is in the kyma-integration namespace (this can be changed in console-backend-service chart configuration)
4. The endpoint is /v1/health, and we are expecting HTTP 200, any other status code means service is not healthy.
*/
type provider struct {
	informer cache.SharedIndexInformer
}

type ServiceData struct {
	// Host is a host, which is used to do HTTP call for status check, for example ec-default.production.svc.cluster.local:8080
	Host string

	// ApplicationName is the name of Application, taken from the `applicationLabelName` label value
	ApplicationName string
}

func newProvider(corev1Interface corev1.CoreV1Interface, integrationNamespace string, informerResyncPeriod time.Duration) *provider {
	svcInterface := corev1Interface.Services(integrationNamespace)

	svcInformer := cache.NewSharedIndexInformer(&cache.ListWatch{
		ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
			return svcInterface.List(context.Background(), options)
		},
		WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
			return svcInterface.Watch(context.Background(), options)
		},
	}, &apiv1.Service{},
		informerResyncPeriod,
		cache.Indexers{},
	)

	return &provider{
		informer: svcInformer,
	}
}

func (p *provider) ListGatewayServices() []ServiceData {
	objects := p.informer.GetStore().List()

	result := make([]ServiceData, 0)
	for _, obj := range objects {
		svc, ok := obj.(*apiv1.Service)
		if !ok {
			continue
		}

		appName, found := p.extractApplicationName(svc)
		if found {
			h, err := p.host(svc)
			if err != nil {
				glog.Errorf("Could not find correct port in %s service %s", pretty.Application, svc.Name)
			}
			result = append(result, ServiceData{
				Host:            h,
				ApplicationName: appName,
			})
		}
	}
	return result
}

func (p *provider) WaitForCacheSync(stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()

	go p.informer.Run(stopCh)
	if !cache.WaitForCacheSync(stopCh, p.informer.HasSynced) {
		glog.Error("Timeout occurred on waiting for gateway api service caches to sync.")
		return
	}
}

func (p *provider) extractApplicationName(obj *apiv1.Service) (string, bool) {
	for k, v := range obj.Labels {
		if k == applicationLabelName && v != "" {
			return v, true
		}
	}
	return "", false
}

func (p *provider) servicePort(obj *apiv1.Service) (int32, error) {
	for _, port := range obj.Spec.Ports {
		if port.Name == externalAPIPortName {
			return port.Port, nil
		}
	}
	return 0, fmt.Errorf("Could not find port with name %s", externalAPIPortName)
}

func (p *provider) host(obj *apiv1.Service) (string, error) {
	port, err := p.servicePort(obj)
	if err != nil {
		return "", errors.Wrap(err, "while creating host")
	}
	return fmt.Sprintf("%s.%s.svc.cluster.local:%d", obj.Name, obj.Namespace, port), nil
}
