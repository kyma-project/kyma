package application_mapping_controller

import (
	"github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned/fake"
	"github.com/kyma-project/kyma/components/application-operator/pkg/application-mapping-controller/mocks"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"testing"
)

func TestAppMappingReconciler_Reconcile(t *testing.T) {
	amName := "testMapping"
	namespace := "testNamespace"

	t.Run("should deploy Gateway when first Application Mapping is created in namespace", func(t *testing.T) {
		//given
		fakeK8SClient := fake.NewSimpleClientset()

		gatewayDeployer := &mocks.GatewayDeployer{}
		gatewayDeployer.On("CheckIfGatewayExists", namespace).Return(false)
		gatewayDeployer.On("DeployGateway", namespace).Return(nil)

		reconciler := NewReconciler(fakeK8SClient.ApplicationconnectorV1alpha1(), gatewayDeployer)

		mapping := v1alpha1.ApplicationMapping{
			ObjectMeta: v1.ObjectMeta{
				Name:      amName,
				Namespace: namespace,
			},
			Spec: v1alpha1.ApplicationMappingSpec{},
		}

		_, err := fakeK8SClient.ApplicationconnectorV1alpha1().ApplicationMappings(namespace).Create(&mapping)
		require.NoError(t, err)

		request := reconcile.Request{
			NamespacedName: types.NamespacedName{
				Namespace: amName,
				Name:      namespace,
			}}

		//when
		_, err = reconciler.Reconcile(request)

		//then
		require.NoError(t, err)
		gatewayDeployer.AssertExpectations(t)
	})
}
