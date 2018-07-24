package suite

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/kyma/components/binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	"github.com/kyma-project/kyma/components/binding-usage-controller/pkg/client/clientset/versioned"
	clientset "github.com/kyma-project/kyma/components/binding-usage-controller/pkg/client/clientset/versioned/typed/servicecatalog/v1alpha1"
	"github.com/stretchr/testify/require"
)

func (ts *TestSuite) CreateTesterBindingUsage() {
	bi := ts.bindingUsageClient()
	bu, err := bi.Create(&v1alpha1.ServiceBindingUsage{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceBindingUsage",
			APIVersion: "servicecatalog.kyma.cx/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "binding-usage-tester",
		},
		Spec: v1alpha1.ServiceBindingUsageSpec{
			ServiceBindingRef: v1alpha1.LocalReferenceByName{
				Name: ts.bindingName,
			},
			UsedBy: v1alpha1.LocalReferenceByKindAndName{
				Kind: "Deployment",
				Name: ts.gwClientSvcDeploymentName,
			},
		},
	})
	require.NoError(ts.t, err)
	ts.testerBindingUsage = bu
}

func (ts *TestSuite) DeleteTesterBindingUsage() {
	err := ts.bindingUsageClient().Delete(ts.testerBindingUsage.Name, &metav1.DeleteOptions{})
	require.NoError(ts.t, err)
}

func (ts *TestSuite) bindingUsageClient() clientset.ServiceBindingUsageInterface {
	client, err := versioned.NewForConfig(ts.config)
	require.NoError(ts.t, err)
	return client.ServicecatalogV1alpha1().ServiceBindingUsages(ts.namespace)
}
