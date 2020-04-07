package servicecatalog

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"

	sbu "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	bu "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/client/clientset/versioned"
	"github.com/sirupsen/logrus"
)

const (
	usageKindName = "usagekind-for-testing"
)

func NewServiceCatalogAddonsTest() (ServiceCatalogAddonsTest, error) {
	config, err := clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
	if err != nil {
		return ServiceCatalogAddonsTest{}, err
	}

	buSc, err := bu.NewForConfig(config)
	if err != nil {
		return ServiceCatalogAddonsTest{}, err
	}

	return ServiceCatalogAddonsTest{
		buSc,
		serviceCatalogAddonsFlow{
			buInterface: buSc,
			log:         logrus.New(),
		},
	}, nil
}

type ServiceCatalogAddonsTest struct {
	buInterface bu.Interface

	serviceCatalogAddonsFlow serviceCatalogAddonsFlow
}

type serviceCatalogAddonsFlow struct {
	buInterface bu.Interface

	log logrus.FieldLogger
}

func (scat ServiceCatalogAddonsTest) CreateResources(t *testing.T, namespace string) {
	scat.serviceCatalogAddonsFlow.createResources(t, namespace)
}

func (scat ServiceCatalogAddonsTest) TestResources(t *testing.T, namespace string) {
	scat.serviceCatalogAddonsFlow.testResources(t, namespace)
}

func (f *serviceCatalogAddonsFlow) createResources(t *testing.T, namespace string) {
	err := f.createUsageKind()
	require.NoError(t, err)
}

func (f *serviceCatalogAddonsFlow) testResources(t *testing.T, namespace string) {
	err := f.verifyUsageKind()
	require.NoError(t, err)
}

func (f *serviceCatalogAddonsFlow) createUsageKind() error {
	f.log.Infof("Creating UsageKind %s", usageKindName)
	_, err := f.buInterface.ServicecatalogV1alpha1().UsageKinds().Create(&sbu.UsageKind{
		TypeMeta: metav1.TypeMeta{
			Kind:       "UsageKind",
			APIVersion: "servicecatalog.kyma-project.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: usageKindName,
		},
		Spec: sbu.UsageKindSpec{
			DisplayName: "test",
			LabelsPath:  "test",
			Resource: &sbu.ResourceReference{
				Group:   "test",
				Version: "test",
				Kind:    "test",
			},
		},
	})
	return err
}

func (f *serviceCatalogAddonsFlow) verifyUsageKind() error {
	f.log.Infof("Checking if UsageKind %s exists", usageKindName)
	_, err := f.buInterface.ServicecatalogV1alpha1().UsageKinds().Get(usageKindName, metav1.GetOptions{})
	return err
}
