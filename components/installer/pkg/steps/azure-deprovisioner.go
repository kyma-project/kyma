package steps

import (
	"log"
	"time"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	installationConfig "github.com/kyma-project/kyma/components/installer/pkg/config"
	internalerrors "github.com/kyma-project/kyma/components/installer/pkg/errors"
	serviceCatalog "github.com/kyma-project/kyma/components/installer/pkg/servicecatalog"
)

// For now we're only interested in these
var azureBrokerServiceClasses = map[string]string{
	"fb9bc99e-0aa9-11e6-8a8a-000d3a002ed5": "azure-sqldb",
	"997b8372-8dac-40ac-ae65-758b4a5075a5": "azure-mysqldb",
	"0346088a-d4b2-4478-aa32-f18e295ec1d9": "azure-rediscache",
}

// DeprovisionConfig is used to parametrize deprovisioning of Azure Resources
type DeprovisionConfig struct {
	BindingDeleteMaxReps    int
	BindingDeleteSleepTime  int
	InstanceDeleteMaxReps   int
	InstanceDeleteSleepTime int
}

// DefaultDeprovisionConfig returns default config for deprovisioning Azure Resources
func DefaultDeprovisionConfig() *DeprovisionConfig {
	res := DeprovisionConfig{
		BindingDeleteMaxReps:    3,  //times
		BindingDeleteSleepTime:  3,  //seconds
		InstanceDeleteMaxReps:   20, //times
		InstanceDeleteSleepTime: 15, //seconds
	}

	return &res
}

// DeprovisionAzureResources performs automatic removal of all resources created with Azure Broker.
func (steps *InstallationSteps) DeprovisionAzureResources(config *DeprovisionConfig, installation installationConfig.InstallationContext) error {

	const stepName string = "Deprovisioning Azure Broker resources"
	steps.PrintInstallationStep(stepName)

	steps.statusManager.InProgress(stepName)

	if config == nil {
		config = DefaultDeprovisionConfig()
	}

	d := deprovisioner{
		config:         config,
		serviceCatalog: steps.serviceCatalog,
		errorHandlers:  steps.errorHandlers,
	}

	if err := d.deprovision(); err != nil {
		return err
	}

	log.Println(stepName + "...DONE")
	return nil
}

type deprovisioner struct {
	config         *DeprovisionConfig
	serviceCatalog serviceCatalog.ClientInterface
	errorHandlers  internalerrors.ErrorHandlersInterface
}

func (d deprovisioner) deprovision() error {

	const allNamespaces = "" //empty string == all namespaces
	instancesList, err := d.serviceCatalog.GetServiceInstances(allNamespaces)
	if d.errorHandlers.CheckError("Error while getting Service Instances: ", err) {
		return err
	}
	azureInstances := filterAzureBrokerInstances(instancesList.Items)

	bindingsList, err := d.serviceCatalog.GetServiceBindings(allNamespaces)

	if d.errorHandlers.CheckError("Error while getting Service Bindings: ", err) {
		return err
	}

	azureBindings := filterAzureBrokerBindings(azureInstances, bindingsList.Items)

	//First, delete all bindings
	if len(azureBindings) > 0 {
		err = d.deleteBindings(azureBindings)
		d.errorHandlers.LogError("An error occurred during deleting Service Bindings: ", err)
	} else {
		log.Println("--> No Service Bindings found!")
	}

	//Then delete all instances
	if len(azureInstances) > 0 {
		err = d.deleteInstances(azureInstances)
		if d.errorHandlers.CheckError("An error occurred during deleting Service Instances: ", err) {
			return err
		}
	} else {
		log.Println("--> No Service Instances found!")
	}

	return err
}

//deleteBindings tries to delete provided objects and waits for completion
func (d deprovisioner) deleteBindings(bindings []v1beta1.ServiceBinding) error {
	log.Println("--> Deleting Service Bindings...")
	deletedBindings := []v1beta1.ServiceBinding{}

	for _, binding := range bindings {
		log.Printf("----> Deleting Service Binding: [%s/%s], Service Instance: %s\n", binding.Namespace, binding.Name, binding.Spec.ServiceInstanceRef.Name)
		err := d.serviceCatalog.DeleteBinding(binding.Namespace, binding.Name)
		if !d.errorHandlers.CheckError("----> An error occurred during deleting Service Binding: ", err) {
			deletedBindings = append(deletedBindings, binding)
		}
	}

	if len(deletedBindings) > 0 {
		log.Println("----> Waiting until all Service Bindings are deleted...")

		//Bindings are quite fast to delete
		time.Sleep(time.Second * time.Duration(d.config.BindingDeleteSleepTime))

		//Wait for bindings to disappear
		bindingsTest := func() (bool, error) {
			return d.bindingsExist(deletedBindings)
		}

		waitUntilExists(d.config.BindingDeleteMaxReps, d.config.BindingDeleteSleepTime, bindingsTest, "Service Binding")
	} else {
		log.Println("--> Warning! No Service Bindings were deleted...")
	}

	return nil
}

//deleteInstances tries to delete provided objects and waits for completion
func (d deprovisioner) deleteInstances(instances []v1beta1.ServiceInstance) error {
	log.Println("--> Deleting Service Instances...")
	deletedInstances := []v1beta1.ServiceInstance{}

	for _, instance := range instances {
		log.Printf("----> Deleting Service Instance: [%s/%s]\n", instance.Namespace, instance.Name)
		err := d.serviceCatalog.DeleteInstance(instance.Namespace, instance.Name)
		if !d.errorHandlers.CheckError("----> An error occurred during deleting Service Instance: ", err) {
			deletedInstances = append(deletedInstances, instance)
		}
	}

	if len(deletedInstances) > 0 {
		log.Println("----> Waiting until all Service Instances are deleted...")

		//Instances take long time to delete
		time.Sleep(time.Second * time.Duration(d.config.InstanceDeleteSleepTime))

		//Wait for instances to disappear
		instancesTest := func() (bool, error) {
			return d.instancesExist(deletedInstances)
		}

		waitUntilExists(d.config.InstanceDeleteMaxReps, d.config.InstanceDeleteSleepTime, instancesTest, "Service Instance")
	} else {
		log.Println("----> Warning! No Service Instances were deleted...")
	}

	return nil
}

// maxReps: maximum number of loop repetitions
// waitTime: time to wait between repetitions (seconds)
// existsFunc: function that returns true if the resource still exists
func waitUntilExists(maxReps, waitTime int, existsFunc func() (bool, error), typeName string) {
	for i := 0; i < maxReps; i++ {
		exists, err := existsFunc()

		if err != nil {
			log.Println("--> An error occured while checking if "+typeName+" exists: ", err)
		} else {
			if !exists {
				log.Printf("----> No more %s(s) exist!\n", typeName)
				return
			}
		}

		if i < maxReps {
			log.Printf("----> Some %s(s) still exist, keep waiting... (done: %v/%v)\n", typeName, i+1, maxReps)
			time.Sleep(time.Second * time.Duration(waitTime))
		}
	}

	log.Printf("----> Warning! Some %s(s) still exist, manual cleanup may be required!\n", typeName)
}

func (d deprovisioner) instancesExist(deletedInstances []v1beta1.ServiceInstance) (bool, error) {
	stillExists := func(existing *v1beta1.ServiceInstance) bool {
		for _, deletedInstance := range deletedInstances {
			if deletedInstance.UID == existing.UID {
				return true
			}
		}

		return false
	}

	findRes, err := d.serviceCatalog.GetServiceInstances("")

	if err != nil {
		return false, err
	}

	result := false

	if len(findRes.Items) > 0 {
		for _, existing := range findRes.Items {
			if stillExists(&existing) {
				log.Printf("------> Service Instance: [%s/%s] still exists!\n", existing.Namespace, existing.Name)
				result = true
			}
		}
	}

	return result, nil
}

func (d deprovisioner) bindingsExist(deletedBindings []v1beta1.ServiceBinding) (bool, error) {

	stillExists := func(existing *v1beta1.ServiceBinding) bool {
		for _, deletedBinding := range deletedBindings {
			if deletedBinding.UID == existing.UID {
				return true
			}
		}

		return false
	}

	findRes, err := d.serviceCatalog.GetServiceBindings("")

	if err != nil {
		return false, err
	}

	result := false

	if len(findRes.Items) > 0 {
		for _, existing := range findRes.Items {
			if stillExists(&existing) {
				log.Printf("------> Service Binding: [%s/%s] still exists!\n", existing.Namespace, existing.Name)
				result = true
			}
		}
	}

	return result, nil
}

//Filters given ServiceInstance slice, returning only those managed by Azure Broker
func filterAzureBrokerInstances(items []v1beta1.ServiceInstance) []v1beta1.ServiceInstance {
	res := []v1beta1.ServiceInstance{}

	for _, item := range items {
		if azureBrokerServiceClasses[item.Spec.ClusterServiceClassRef.Name] != "" {
			res = append(res, item)
		}
	}

	return res
}

//Filters given ServiceBinding slice, returning only those managed by Azure Broker
//Since ServiceBinding objects don't have any metadata related to the Broker, we have to find corresponding ServiceInstance object.
func filterAzureBrokerBindings(azureBrokerInstances []v1beta1.ServiceInstance, bindings []v1beta1.ServiceBinding) []v1beta1.ServiceBinding {
	res := []v1beta1.ServiceBinding{}

	isAzureBrokerBinding := func(binding v1beta1.ServiceBinding) bool {
		for _, instance := range azureBrokerInstances {
			//ServiceInstanceRef is LocalObjectReference (same namespace)
			if instance.Name == binding.Spec.ServiceInstanceRef.Name {
				return true
			}
		}
		return false
	}

	for _, binding := range bindings {
		if isAzureBrokerBinding(binding) {
			res = append(res, binding)
		}
	}

	return res
}
