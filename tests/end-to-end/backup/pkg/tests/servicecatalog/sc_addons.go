package servicecatalog

import (
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"

	sbu "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	bu "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/client/clientset/versioned"
	"github.com/sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
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

func (t ServiceCatalogAddonsTest) CreateResources(namespace string) {
	t.serviceCatalogAddonsFlow.createResources(namespace)
}

func (t ServiceCatalogAddonsTest) TestResources(namespace string) {
	t.serviceCatalogAddonsFlow.testResources(namespace)
}

func (f *serviceCatalogAddonsFlow) createResources(namespace string) {
	err := f.createUsageKind()
	So(err, ShouldBeNil)
}

func (f *serviceCatalogAddonsFlow) testResources(namespace string) {
	err := f.verifyUsageKind()
	So(err, ShouldBeNil)
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
