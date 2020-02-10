package application_mapping_controller

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1alpha12 "github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	log "github.com/sirupsen/logrus"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	kymaSystemNamespace      = "kyma-system"
	kymaIntegrationNamespace = "kyma-integration"
)

//go:generate mockery -name AppMappingReconciler
type AppMappingReconciler interface {
	Reconcile(request reconcile.Request) (reconcile.Result, error)
}

//go:generate mockery -name ApplicationMappingManagerClient
type ApplicationMappingManagerClient interface {
	Get(ctx context.Context, key client.ObjectKey, obj runtime.Object) error
	List(ctx context.Context, list runtime.Object, opts ...client.ListOption) error
}

type appMappingReconciler struct {
	appConnClient   ApplicationMappingManagerClient
	gatewayDeployer GatewayDeployer
}

func NewReconciler(appConnClient ApplicationMappingManagerClient, gatewayDeployer GatewayDeployer) AppMappingReconciler {
	return &appMappingReconciler{
		appConnClient:   appConnClient,
		gatewayDeployer: gatewayDeployer,
	}
}

func (r *appMappingReconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	if request.Namespace == kymaIntegrationNamespace || request.Namespace == kymaSystemNamespace {
		return reconcile.Result{}, nil
	}

	log.Infof("Processing %s Application Mapping...", request.Name)

	instance := &v1alpha12.ApplicationMapping{}

	err := r.appConnClient.Get(context.Background(), request.NamespacedName, instance)

	if err != nil {
		return r.handleErrorWhileGettingInstance(err, request.NamespacedName)
	}

	if r.gatewayShouldBeCreated(request.Namespace) {
		return reconcile.Result{}, r.createGateway(request.Namespace)
	}

	return reconcile.Result{}, nil
}

func (r *appMappingReconciler) handleErrorWhileGettingInstance(err error, namespacedName types.NamespacedName) (reconcile.Result, error) {
	if k8sErrors.IsNotFound(err) {
		log.Infof("Application Mapping %s deleted", namespacedName.Name)

		if r.gatewayShouldBeDeleted(namespacedName.Namespace) {
			if err := r.gatewayDeployer.RemoveGateway(namespacedName.Namespace); err != nil {
				return reconcile.Result{}, logAndError(err, "Error deleting gateway from namespace %s", namespacedName.Namespace)
			}
			log.Infof("Successfully deleted Gateway from namespace %s", namespacedName.Namespace)
			return reconcile.Result{}, nil
		}
	}
	return reconcile.Result{}, logAndError(err, "Error getting ApplicationMapping: %s", namespacedName.Name)
}

func (r *appMappingReconciler) createGateway(namespace string) error {
	err := r.gatewayDeployer.DeployGateway(namespace)

	if err != nil {
		return logAndError(err, "Error deploying Gateway for namespace %s", namespace)
	}

	log.Infof("Successfully deployed Gateway for namespace %s", namespace)
	return nil
}

func (r *appMappingReconciler) gatewayShouldBeDeleted(namespace string) bool {
	list, err := r.getAppMappingsList(namespace)

	if err != nil {
		return false
	}

	return len(list.Items) == 0
}

func (r *appMappingReconciler) gatewayShouldBeCreated(namespace string) bool {
	list, err := r.getAppMappingsList(namespace)

	if err != nil {
		return false
	}

	return len(list.Items) > 0 && !r.gatewayDeployer.GatewayExists(namespace)
}

func (r *appMappingReconciler) getAppMappingsList(namespace string) (*v1alpha12.ApplicationMappingList, error) {
	list := &v1alpha12.ApplicationMappingList{}

	err := r.appConnClient.List(context.Background(), list, &client.ListOptions{Namespace: namespace})

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
