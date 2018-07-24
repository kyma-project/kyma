package steps

import (
	"testing"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	internalerrors "github.com/kyma-project/kyma/components/installer/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"
)

func TestAzureDeprovisioner(t *testing.T) {
	Convey("filterAzureBrokerInstances function should", t, func() {

		Convey("only return instances corresponding to Azure Classes", func() {

			const azureClassID = "fb9bc99e-0aa9-11e6-8a8a-000d3a002ed5"
			//given
			i1 := stubServiceInstance("i1", azureClassID)
			i2 := stubServiceInstance("i2", "shouldBeFilteredOut1")
			i3 := stubServiceInstance("i3", "shouldBeFilteredOut2")

			//when
			filtered := filterAzureBrokerInstances([]v1beta1.ServiceInstance{*i1, *i2, *i3})

			So(len(filtered), ShouldEqual, 1)
			So(filtered[0].Name, ShouldEqual, "i1")
		})
	})

	Convey("filterAzureBrokerBindings function should", t, func() {

		Convey("only return bindings corresponding to Azure Classes", func() {

			const azureClassID = "fb9bc99e-0aa9-11e6-8a8a-000d3a002ed5"
			//given
			i1 := stubServiceInstance("i1", azureClassID)
			b1 := stubServiceBinding("b1", "i1", "")
			i2 := stubServiceInstance("i2", azureClassID)
			b2 := stubServiceBinding("b2", "i2", "")
			b3 := stubServiceBinding("b3", "shouldBeFiltered", "")
			//when
			filtered := filterAzureBrokerBindings([]v1beta1.ServiceInstance{*i2, *i1}, []v1beta1.ServiceBinding{*b1, *b2, *b3})

			So(len(filtered), ShouldEqual, 2)
			So(filtered[0].Name, ShouldEqual, "b1")
			So(filtered[1].Name, ShouldEqual, "b2")
		})
	})

	Convey("deleteBindings function should", t, func() {

		Convey("retry for configured number of times", func() {

			const azureClassID = "fb9bc99e-0aa9-11e6-8a8a-000d3a002ed5"
			//given
			b1 := stubServiceBinding("b1", "i1", "11")
			b1.Namespace = "abc"
			b2 := stubServiceBinding("b2", "i2", "22")
			b2.Namespace = "abc"
			b3 := stubServiceBinding("b3", "i3", "33")
			b3.Namespace = "abc"
			bindings := []v1beta1.ServiceBinding{*b1, *b2, *b3}

			stubClient := deleteBindingsMockClient{
				maxGetBindingsCount: 2,
				items:               bindings,
			}

			config := DeprovisionConfig{
				BindingDeleteMaxReps:   5,
				BindingDeleteSleepTime: 0,
			}

			d := deprovisioner{
				config:         &config,
				serviceCatalog: &stubClient,
				errorHandlers:  &internalerrors.ErrorHandlers{},
			}

			//when
			err := d.deleteBindings(bindings)

			//then
			So(err, ShouldBeNil)
			//Ensure deleteBinding was called properly
			So(len(stubClient.deletedBindings), ShouldEqual, 3)
			So(stubClient.deletedBindings[0], ShouldEqual, "abc/b1")
			So(stubClient.deletedBindings[1], ShouldEqual, "abc/b2")
			So(stubClient.deletedBindings[2], ShouldEqual, "abc/b3")
			//Ensure getBindings was only called twice
			So(stubClient.actualGetBindingsCount, ShouldEqual, 2)
		})
	})

}

func stubServiceInstance(name, clusterServiceClassName string) *v1beta1.ServiceInstance {
	ref := v1beta1.ClusterObjectReference{}
	ref.Name = clusterServiceClassName

	res := v1beta1.ServiceInstance{}
	res.Name = name
	res.Spec.ClusterServiceClassRef = &ref

	return &res
}

func stubServiceBinding(name, serviceInstanceName, UID string) *v1beta1.ServiceBinding {
	instanceRef := v1beta1.LocalObjectReference{
		Name: serviceInstanceName,
	}

	res := v1beta1.ServiceBinding{}
	res.Name = name
	res.Spec.ServiceInstanceRef = instanceRef

	return &res
}

func mockCatalogInterfaceDeleteBindings() {
	return
}

//ServiceCatalog interface mock for deleteBindings function
type deleteBindingsMockClient struct {
	maxGetBindingsCount    int
	actualGetBindingsCount int
	items                  []v1beta1.ServiceBinding
	deletedBindings        []string //registers calls to "DeleteBinding"
}

func (c *deleteBindingsMockClient) GetServiceBindings(ns string) (*v1beta1.ServiceBindingList, error) {
	res := v1beta1.ServiceBindingList{}
	c.maxGetBindingsCount--
	c.actualGetBindingsCount++
	if c.maxGetBindingsCount > 0 {
		res := v1beta1.ServiceBindingList{}
		res.Items = c.items
		return &res, nil
	}
	return &res, nil
}

func (c *deleteBindingsMockClient) GetServiceInstances(ns string) (*v1beta1.ServiceInstanceList, error) {
	return nil, nil
}

func (c *deleteBindingsMockClient) DeleteBinding(namespace, name string) error {
	c.deletedBindings = append(c.deletedBindings, namespace+"/"+name)
	return nil
}
func (c *deleteBindingsMockClient) DeleteInstance(namespace, name string) error { return nil }
