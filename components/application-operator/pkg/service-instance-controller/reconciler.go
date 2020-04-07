package service_instance_controller

import (
	"context"

	"k8s.io/helm/pkg/proto/hapi/release"

	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/application-operator/pkg/kymahelm/gateway"
	"github.com/kyma-project/kyma/components/application-operator/pkg/utils"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/sirupsen/logrus"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

//go:generate mockery -name ServiceInstanceReconciler
type ServiceInstanceReconciler interface {
	Reconcile(request reconcile.Request) (reconcile.Result, error)
}

//go:generate mockery -name ServiceInstanceManagerClient
type ServiceInstanceManagerClient interface {
	Get(ctx context.Context, key client.ObjectKey, obj runtime.Object) error
	List(ctx context.Context, list runtime.Object, opts ...client.ListOption) error
}

type serviceInstanceReconciler struct {
	serviceInstanceClient ServiceInstanceManagerClient
	gatewayDeployer       gateway.GatewayManager
	logger                *logrus.Entry
}

func NewReconciler(appConnClient ServiceInstanceManagerClient, gatewayDeployer gateway.GatewayManager, logger *logrus.Entry) ServiceInstanceReconciler {
	return &serviceInstanceReconciler{
		serviceInstanceClient: appConnClient,
		gatewayDeployer:       gatewayDeployer,
		logger:                logger,
	}
}

func (r *serviceInstanceReconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	log := r.logger.WithField("service-instance", request.NamespacedName)
	// Gateway should not be deployed in system namespaces
	if utils.IsSystemNamespace(request.Namespace) {
		return reconcile.Result{}, nil
	}

	log.Infof("Processing %s Service Instance...", request.Name)

	instance := &v1beta1.ServiceInstance{}

	err := r.serviceInstanceClient.Get(context.Background(), request.NamespacedName, instance)

	if err != nil {
		return r.handleErrorWhileGettingInstance(err, request.NamespacedName, log)
	}

	status, exists, err := r.getGatewayStatus(request.Namespace, log)
	if err != nil {
		return reconcile.Result{}, logAndError(err, "Error checking if Gateway should be created", log)
	}

	if !exists {
		return reconcile.Result{}, r.createGateway(request.Namespace, log)
	}

	// If Gateway release is not properly deployed - delete release and requeue request
	if status == release.Status_FAILED {
		err := r.deleteGateway(request.Namespace, log)
		if err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true}, nil
	}

	log.Infof("Gateway for %s Service Instance already deployed", request.Name)

	return reconcile.Result{}, nil
}

func (r *serviceInstanceReconciler) handleErrorWhileGettingInstance(err error, namespacedName types.NamespacedName, log *logrus.Entry) (reconcile.Result, error) {
	if k8sErrors.IsNotFound(err) {
		log.Infof("Service Instance %s deleted", namespacedName.Name)

		shouldBeDeleted, err := r.gatewayShouldBeDeleted(namespacedName.Namespace, log)

		if err != nil {
			return reconcile.Result{}, logAndError(err, "Error checking if Gateway should be deleted", log)
		}

		if shouldBeDeleted {
			return reconcile.Result{}, r.deleteGateway(namespacedName.Namespace, log)
		}
		return reconcile.Result{}, nil
	}
	return reconcile.Result{}, logAndError(err, "Error getting Service Instance: %s", log)
}

func (r *serviceInstanceReconciler) deleteGateway(namespace string, log *logrus.Entry) error {
	log.Info("Deleting Gateway")
	if err := r.gatewayDeployer.DeleteGateway(namespace); err != nil {
		return logAndError(err, "Error deleting gateway", log)
	}
	log.Info("Successfully deleted Gateway")
	return nil
}

func (r *serviceInstanceReconciler) createGateway(namespace string, log *logrus.Entry) error {
	log.Info("Deploying Gateway")
	err := r.gatewayDeployer.InstallGateway(namespace)

	if err != nil {
		return logAndError(err, "Error deploying Gateway", log)
	}

	log.Info("Successfully deployed Gateway")
	return nil
}

func (r *serviceInstanceReconciler) gatewayShouldBeDeleted(namespace string, log *logrus.Entry) (bool, error) {
	log.Info("Checking if Gateway should be deleted")
	list, err := r.getServiceInstanceList(namespace)

	if err != nil {
		return false, err
	}

	return len(list.Items) == 0, nil
}

func (r *serviceInstanceReconciler) getGatewayStatus(namespace string, log *logrus.Entry) (release.Status_Code, bool, error) {
	log.Info("Checking if Gateway exists and is in correct state")

	exists, status, err := r.gatewayDeployer.GatewayExists(namespace)
	if err != nil {
		return release.Status_UNKNOWN, false, err
	}

	log.Infof("Gateway status: Exists: %v, Status: %s", exists, release.Status_Code_name[int32(status)])

	return status, exists, nil
}

func (r *serviceInstanceReconciler) getServiceInstanceList(namespace string) (*v1beta1.ServiceInstanceList, error) {
	list := &v1beta1.ServiceInstanceList{}

	err := r.serviceInstanceClient.List(context.Background(), list, &client.ListOptions{Namespace: namespace})

	if err != nil {
		return &v1beta1.ServiceInstanceList{}, err
	}

	return list, err
}

func logAndError(err error, msg string, log *logrus.Entry) error {
	log.Errorf("%s: %s", msg, err.Error())
	return err
}
