package application_mapping_controller

import (
	"context"
	"github.com/kyma-project/kyma/components/application-operator/pkg/utils"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1alpha12 "github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	"github.com/sirupsen/logrus"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
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
	logger          *logrus.Entry
}

func NewReconciler(appConnClient ApplicationMappingManagerClient, gatewayDeployer GatewayDeployer, logger *logrus.Entry) AppMappingReconciler {
	return &appMappingReconciler{
		appConnClient:   appConnClient,
		gatewayDeployer: gatewayDeployer,
		logger:          logger,
	}
}

func (r *appMappingReconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	log := r.logger.WithField("application-mapping", request.NamespacedName)
	// Gateway should not be deployed in system namespaces
	if utils.IsSystemNamespace(request.Namespace) {
		return reconcile.Result{}, nil
	}

	log.Infof("Processing %s Application Mapping...", request.Name)

	instance := &v1alpha12.ApplicationMapping{}

	err := r.appConnClient.Get(context.Background(), request.NamespacedName, instance)

	if err != nil {
		return r.handleErrorWhileGettingInstance(err, request.NamespacedName, log)
	}

	shouldBeCreated, err := r.gatewayShouldBeCreated(request.Namespace, log)

	if err != nil {
		return reconcile.Result{}, logAndError(err, "Error checking if Gateway should be created", log)
	}

	if shouldBeCreated {
		return reconcile.Result{}, r.createGateway(request.Namespace, log)
	}

	return reconcile.Result{}, nil
}

func (r *appMappingReconciler) handleErrorWhileGettingInstance(err error, namespacedName types.NamespacedName, log *logrus.Entry) (reconcile.Result, error) {
	if k8sErrors.IsNotFound(err) {
		log.Infof("Application Mapping %s deleted", namespacedName.Name)

		shouldBeDeleted, err := r.gatewayShouldBeDeleted(namespacedName.Namespace, log)

		if err != nil {
			return reconcile.Result{}, logAndError(err, "Error checking if Gateway should be deleted", log)
		}

		if shouldBeDeleted {
			log.Info("Deleting Gateway")
			if err := r.gatewayDeployer.RemoveGateway(namespacedName.Namespace); err != nil {
				return reconcile.Result{}, logAndError(err, "Error deleting gateway", log)
			}
			log.Info("Successfully deleted Gateway")
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, nil
	}
	return reconcile.Result{}, logAndError(err, "Error getting ApplicationMapping: %s", log)
}

func (r *appMappingReconciler) createGateway(namespace string, log *logrus.Entry) error {
	log.Info("Deploying Gateway")
	err := r.gatewayDeployer.DeployGateway(namespace)

	if err != nil {
		return logAndError(err, "Error deploying Gateway", log)
	}

	log.Info("Successfully deployed Gateway")
	return nil
}

func (r *appMappingReconciler) gatewayShouldBeDeleted(namespace string, log *logrus.Entry) (bool, error) {
	log.Info("Checking if Gateway should be deleted")
	list, err := r.getAppMappingsList(namespace)

	if err != nil {
		return false, err
	}

	return len(list.Items) == 0, nil
}

func (r *appMappingReconciler) gatewayShouldBeCreated(namespace string, log *logrus.Entry) (bool, error) {
	log.Info("Checking if Gateway should be created")
	list, err := r.getAppMappingsList(namespace)

	if err != nil {
		return false, err
	}

	return len(list.Items) > 0 && !r.gatewayDeployer.GatewayExists(namespace), nil
}

func (r *appMappingReconciler) getAppMappingsList(namespace string) (*v1alpha12.ApplicationMappingList, error) {
	list := &v1alpha12.ApplicationMappingList{}

	err := r.appConnClient.List(context.Background(), list, &client.ListOptions{Namespace: namespace})

	if err != nil {
		return &v1alpha12.ApplicationMappingList{}, err
	}

	return list, err
}

func logAndError(err error, msg string, log *logrus.Entry) error {
	log.Errorf("%s: %s", msg, err.Error())
	return err
}
