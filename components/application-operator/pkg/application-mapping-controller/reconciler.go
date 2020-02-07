package application_mapping_controller

import (
	"fmt"

	v1alpha12 "github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned/typed/applicationconnector/v1alpha1"
	log "github.com/sirupsen/logrus"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

//go:generate mockery -name AppMappingReconciler
type AppMappingReconciler interface {
	Reconcile(request reconcile.Request) (reconcile.Result, error)
}

type appMappingReconciler struct {
	appConnClient   v1alpha1.ApplicationconnectorV1alpha1Interface
	gatewayDeployer GatewayDeployer
}

func NewReconciler(appConnClient v1alpha1.ApplicationconnectorV1alpha1Interface, gatewayDeployer GatewayDeployer) AppMappingReconciler {
	return &appMappingReconciler{
		appConnClient:   appConnClient,
		gatewayDeployer: gatewayDeployer,
	}
}

func (r *appMappingReconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	log.Infof("Processing %s Application Mapping...", request.Name)

	mappings := r.appConnClient.ApplicationMappings(request.Namespace)

	_, err := mappings.Get(request.Name, v1.GetOptions{})

	if err != nil {
		return r.handleErrorWhileGettingInstance(err, request, mappings)
	}

	if r.gatewayShouldBeCreated(mappings, request.Namespace) {
		return reconcile.Result{}, r.createGateway(request.Namespace)
	}

	return reconcile.Result{}, nil
}

func (r *appMappingReconciler) handleErrorWhileGettingInstance(err error, request reconcile.Request, mappings v1alpha1.ApplicationMappingInterface) (reconcile.Result, error) {
	if k8sErrors.IsNotFound(err) {
		log.Infof("Application Mapping %s deleted", request.Name)

		if r.gatewayShouldBeDeleted(mappings, request.Namespace) {
			if err := r.gatewayDeployer.RemoveGateway(request.Namespace); err != nil {
				return reconcile.Result{}, logAndError(err, "Error deleting gateway from namespace %s", request.Namespace)
			}
			log.Infof("Successfully deleted Gateway from namespace %s", request.Namespace)
			return reconcile.Result{}, nil
		}
	}
	return reconcile.Result{}, logAndError(err, "Error getting ApplicationMapping: %s", request.Name)
}

func (r *appMappingReconciler) createGateway(namespace string) error {
	log.Infof("Deploying Gateway for namespace %s", namespace)
	err := r.gatewayDeployer.DeployGateway(namespace)

	if err != nil {
		return logAndError(err, "Error deploying Gateway for namespace %s", namespace)
	}

	log.Infof("Successfully deployed Gateway for namespace %s", namespace)
	return nil
}

func (r *appMappingReconciler) gatewayShouldBeDeleted(mappings v1alpha1.ApplicationMappingInterface, namespace string) bool {
	list, err := getAppMappingsList(mappings, namespace)

	if err != nil {
		return false
	}

	return len(list.Items) == 0
}

func (r *appMappingReconciler) gatewayShouldBeCreated(mappings v1alpha1.ApplicationMappingInterface, namespace string) bool {
	list, err := getAppMappingsList(mappings, namespace)

	if err != nil {
		return false
	}

	return len(list.Items) > 0 && r.gatewayDeployer.CheckIfGatewayExists(namespace)
}

func getAppMappingsList(mappings v1alpha1.ApplicationMappingInterface, namespace string) (*v1alpha12.ApplicationMappingList, error) {
	list, err := mappings.List(v1.ListOptions{})

	if err != nil {
		log.Errorf("Error getting list of ApplicationMappings for namespace %s", namespace)
		return &v1alpha12.ApplicationMappingList{}, err
	}

	return list, err
}

func logAndError(err error, format string, args ...interface{}) error {
	msg := fmt.Sprintf(format, args...)
	log.Errorf("%s: %s", msg, err.Error())
	return err
}
