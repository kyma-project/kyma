package compassconnection

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/pkg/correlation"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/config"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/pkg/apis/compass/v1alpha1"
)

const (
	controllerName = "compass-connection-controller"
)

type Client interface {
	Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error
	Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error
	Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error
}

// Reconciler reconciles a CompassConnection object
type Reconciler struct {
	client     Client
	supervisor Supervisor

	minimalConfigSyncTime time.Duration

	log            *logrus.Entry
	runtimeID      string
	configProvider config.Provider
}

func InitCompassConnectionController(
	mgr manager.Manager,
	supervisior Supervisor,
	minimalConfigSyncTime time.Duration,
	configProvider config.Provider) error {

	reconciler := newReconciler(mgr.GetClient(), supervisior, minimalConfigSyncTime, configProvider)

	return startController(mgr, reconciler)
}

func startController(mgr manager.Manager, reconciler reconcile.Reconciler) error {
	c, err := controller.New(controllerName, mgr, controller.Options{Reconciler: reconciler})
	if err != nil {
		return err
	}

	return c.Watch(&source.Kind{Type: &v1alpha1.CompassConnection{}}, &handler.EnqueueRequestForObject{})
}

func newReconciler(client Client, supervisior Supervisor, minimalConfigSyncTime time.Duration, configProvider config.Provider) reconcile.Reconciler {
	return &Reconciler{
		client:                client,
		supervisor:            supervisior,
		minimalConfigSyncTime: minimalConfigSyncTime,
		log:                   logrus.WithField("Controller", "CompassConnection"),
		configProvider:        configProvider,
	}
}

func (r *Reconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {

	if r.runtimeID == "" {
		runtimeConfig, err := r.configProvider.GetRuntimeConfig()
		if err == nil {
			r.runtimeID = runtimeConfig.RuntimeId
		}
	}

	correlationID := r.runtimeID + "_" + uuid.New().String()
	logFields := logrus.Fields{
		"CompassConnection":            request.Name,
		correlation.RequestIDHeaderKey: correlationID,
	}
	log := r.log.WithFields(logFields)
	ctx = correlation.SaveCorrelationIDHeaderToContext(ctx, str.Ptr(correlation.RequestIDHeaderKey), str.Ptr(correlationID))

	connection, err := r.getConnection(ctx, log, request)
	if err != nil {
		return reconcile.Result{}, err
	}

	if connection == nil {
		_, err := r.initConnection(ctx, log)
		return reconcile.Result{}, err
	}

	// Make sure the minimal time passed since last Compass Connection synchronization.
	// This allows to rate limit Compass calls
	if skipConnectionSync(connection, log, r.minimalConfigSyncTime) {
		return reconcile.Result{}, nil
	}

	if connection.Failed() {
		_, err := r.initConnection(ctx, log)
		return reconcile.Result{}, err
	}

	if err := r.ensureCertificateIsValid(ctx, connection, log); err != nil {
		return reconcile.Result{}, err
	}

	log.Info("Trying to connect to Compass and apply Runtime configuration...")

	// Make sure the minimal time passed since last Application synchronization.
	// This allows to rate limit Compass calls
	if skipApplicationSync(connection, log, r.minimalConfigSyncTime) {
		return reconcile.Result{}, nil
	}

	return reconcile.Result{}, r.synchronizeApplications(ctx, connection, log)
}

func (r *Reconciler) getConnection(ctx context.Context, log *logrus.Entry, request reconcile.Request) (*v1alpha1.CompassConnection, error) {
	instance := &v1alpha1.CompassConnection{}
	err := r.client.Get(ctx, request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Compass Connection deleted")
			return nil, nil
		}

		log.Info("Failed to read Compass Connection.")
		return nil, err
	}

	return instance, nil
}

func (r *Reconciler) initConnection(ctx context.Context, log *logrus.Entry) (*v1alpha1.CompassConnection, error) {
	log.Info("Trying to initialize new connection...")

	instance, err := r.supervisor.InitializeCompassConnection(ctx)
	if err != nil {
		log.Errorf("Failed to initialize Compass Connection: %s", err.Error())
		return nil, err
	}

	log.Infof("Attempt to initialize Compass Connection ended with status: %s", instance.Status)
	return instance, nil
}

func (r *Reconciler) synchronizeApplications(ctx context.Context, connection *v1alpha1.CompassConnection, log *logrus.Entry) error {
	synchronized, err := r.supervisor.SynchronizeWithCompass(ctx, connection)
	if err != nil {
		log.Errorf("Failed to synchronize with Compass: %s", err.Error())
		return err
	}

	log.Infof("Synchronization finished. Compass Connection status: %s", synchronized.Status)
	return nil
}

func (r *Reconciler) ensureCertificateIsValid(ctx context.Context, connection *v1alpha1.CompassConnection, log *logrus.Entry) error {
	log.Infof("Attempting to maintain connection with Compass...")
	err := r.supervisor.MaintainCompassConnection(ctx, connection)

	if err != nil {
		log.Errorf("Failed to maintain connection with Compass: %s", err.Error())
		return err
	}

	return nil
}

func skipConnectionSync(connection *v1alpha1.CompassConnection, log *logrus.Entry, minimalConfigSyncTime time.Duration) bool {
	if connection.Spec.ResyncNow || connection.Status.ConnectionStatus == nil {
		return false
	}
	timeSinceLastConnAttempt := time.Now().Unix() - connection.Status.ConnectionStatus.LastSync.Unix()

	if timeSinceLastConnAttempt < int64(minimalConfigSyncTime.Seconds()) {
		log.Infof("Skipping connection initialization/maintenance. Minimal resync time not passed. Last attempt: %v", connection.Status.ConnectionStatus.LastSync)
		return true
	}
	return false
}

func skipApplicationSync(connection *v1alpha1.CompassConnection, log *logrus.Entry, minimalConfigSyncTime time.Duration) bool {
	if connection.Spec.ResyncNow || connection.Status.SynchronizationStatus == nil {
		return false
	}
	timeSinceLastSyncAttempt := time.Now().Unix() - connection.Status.SynchronizationStatus.LastAttempt.Unix()

	if timeSinceLastSyncAttempt < int64(minimalConfigSyncTime.Seconds()) {
		log.Infof("Skipping application synchronization. Minimal resync time not passed. Last attempt: %v", connection.Status.SynchronizationStatus.LastAttempt)
		return true
	}
	return false
}
