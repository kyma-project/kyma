package testkit

import (
	"fmt"

	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	restclient "k8s.io/client-go/rest"
)

type K8sResourcesClient interface {
	CreateDummyApplication(namePrefix string, spec v1alpha1.ApplicationSpec) (*v1alpha1.Application, error)
	DeleteApplication(name string, options *v1.DeleteOptions) error
}
type k8sResourcesClient struct {
	applicationsClient *versioned.Clientset
}

func NewK8sResourcesClient() (K8sResourcesClient, error) {
	k8sConfig, err := restclient.InClusterConfig()
	if err != nil {
		return nil, err
	}
	return initClient(k8sConfig)
}

func initClient(k8sConfig *restclient.Config) (K8sResourcesClient, error) {
	applicationsClientSet, err := versioned.NewForConfig(k8sConfig)
	if err != nil {
		return nil, err
	}
	return &k8sResourcesClient{
		applicationsClient: applicationsClientSet,
	}, nil
}

func (c *k8sResourcesClient) CreateDummyApplication(namePrefix string, spec v1alpha1.ApplicationSpec) (*v1alpha1.Application, error) {
	dummyAppName := addRandomPostfix(namePrefix)

	dummyApp := &v1alpha1.Application{
		TypeMeta:   v1.TypeMeta{Kind: "Application", APIVersion: v1alpha1.SchemeGroupVersion.String()},
		ObjectMeta: v1.ObjectMeta{Name: dummyAppName},
		Spec:       spec,
	}

	return c.applicationsClient.ApplicationconnectorV1alpha1().Applications().Create(dummyApp)
}

func addRandomPostfix(s string) string {
	return fmt.Sprintf(s+"-%s", rand.String(5))
}

func (c *k8sResourcesClient) DeleteApplication(name string, options *v1.DeleteOptions) error {
	return c.applicationsClient.ApplicationconnectorV1alpha1().Applications().Delete(name, options)
}
