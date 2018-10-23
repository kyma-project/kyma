package brokerinstaller

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests"
	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
)

const (
	installTimeout    int64 = 240
	tillerConnTimeout int64 = 10
)

//type BrokerInstaller struct {
//	k8sCli      *v1.CoreV1Client
//	helmCli     *helm.Client
//	chartPath   string
//	releaseName string
//	namespace   string
//}

//func New(releaseName, namespace string) (*BrokerInstaller, error) {
//	tillerHost := os.Getenv("TILLER_HOST")
//	if tillerHost == "" {
//		tillerHost = fmt.Sprintf("127.0.0.1:44134")
//	}
//
//	cli := helm.NewClient(helm.Host(tillerHost), helm.ConnectTimeout(tillerConnTimeout))
//	return &BrokerInstaller{
//		helmCli:     cli,
//		releaseName: releaseName,
//		chartPath:   chartPath,
//		namespace:   namespace,
//	}, nil
//}

type BrokerInstaller struct {
	releaseName string
	namespace   string
	typeOf 		string
}

func New(releaseName, namespace, typeOf string) *BrokerInstaller {
	return &BrokerInstaller{
		releaseName: releaseName,
		namespace:   namespace,
		typeOf: 	 typeOf,
	}
}

func (t *BrokerInstaller) Install(svcatCli *clientset.Clientset) error {
	url := "http://" + t.releaseName + "." + t.namespace + ".svc.cluster.local"

	var err error
	if t.typeOf == tester.ClusterServiceBroker {
		_, err = svcatCli.ServicecatalogV1beta1().ClusterServiceBrokers().Create(newUpsClusterServiceBroker(t.releaseName, url))
	} else {
		_, err = svcatCli.ServicecatalogV1beta1().ServiceBrokers(t.namespace).Create(newUpsServiceBroker(t.releaseName, url))
	}
	return err
}

func (t *BrokerInstaller) Uninstall(svcatCli *clientset.Clientset) error {
	var err error
	if t.typeOf == tester.ClusterServiceBroker {
		err = svcatCli.ServicecatalogV1beta1().ClusterServiceBrokers().Delete(t.releaseName, nil)
	} else {
		err = svcatCli.ServicecatalogV1beta1().ServiceBrokers(t.namespace).Delete(t.releaseName, nil)
	}
	return err
}

func (t *BrokerInstaller) ReleaseName() string {
	return t.releaseName
}

func (t *BrokerInstaller) TypeOf() string {
	return t.typeOf
}

func newUpsClusterServiceBroker(name, url string) *v1beta1.ClusterServiceBroker {
	return &v1beta1.ClusterServiceBroker{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v1beta1.ClusterServiceBrokerSpec{
			CommonServiceBrokerSpec: v1beta1.CommonServiceBrokerSpec{
				URL: url,
			},
		},
	}
}

func newUpsServiceBroker(name, url string) *v1beta1.ServiceBroker {
	return &v1beta1.ServiceBroker{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v1beta1.ServiceBrokerSpec{
			CommonServiceBrokerSpec: v1beta1.CommonServiceBrokerSpec{
				URL: url,
			},
		},
	}
}
