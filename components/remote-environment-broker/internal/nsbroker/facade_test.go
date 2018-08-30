package nsbroker_test

import (
	"fmt"
	"testing"

	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/fake"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sFake "k8s.io/client-go/kubernetes/fake"
	"github.com/kyma-project/kyma/components/helm-broker/platform/logger/spy"
	"github.com/kyma-project/kyma/components/remote-environment-broker/internal/nsbroker"
	"github.com/stretchr/testify/assert"
)

func TestNsBrokerCreate(t *testing.T) {

	t.Run("happy path", func(t *testing.T) {
		// GIVEN
		//	client := fake.NewSimpleClientset(&fixEM)
		scFake := fake.NewSimpleClientset()
		k8sClientSet := k8sFake.NewSimpleClientset()

		sut := nsbroker.NewFacade(scFake.ServicecatalogV1beta1(), k8sClientSet.CoreV1(), fixRebSelectorKey(), fixRebSelectorValue(), fixTargetPort(), spy.NewLogDummy())
		sut.Create(fixDestNs(), fixSystemNs())
		// WHEN
		err := sut.Create(fixDestNs(), fixSystemNs())
		require.NoError(t, err)
		actualBroker, err := scFake.Servicecatalog().ServiceBrokers(fixDestNs()).Get("remote-env-broker",v1.GetOptions{})
	require.NoError(t,err)
		actualLabelVal := actualBroker.Labels["namespaced-remote-env-broker"]
		assert.Equal(t,"true", actualLabelVal)
		assert.Equal(t,"http://TODO",actualBroker.Spec.URL)
	})


}

func TestNsBrokerDelete(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {

	})
}

func TestNsBrokerExist(t *testing.T) {
	t.Run("when exist", func (t *testing.T) {

	})

	t.Run("when does not exist", func(t *testing.T) {

	})

}

func fixRebSelectorKey() string {
	return "app"
}

func fixRebSelectorValue() string {
	return "reb"
}

func fixTargetPort() int32 {
	return int32(8080)
}

func fixDestNs() string {
	return "stage"
}

func fixSystemNs() string {
	return "kyma-system"
}
