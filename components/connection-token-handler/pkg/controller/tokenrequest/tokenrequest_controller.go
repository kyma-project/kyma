package tokenrequest

import (
	"context"
	"log"
	"time"

	applicationconnectorv1alpha1 "github.com/kyma-project/kyma/components/connection-token-handler/pkg/apis/applicationconnector/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// Add creates a new TokenRequest Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager, opts Options) error {
	return add(mgr, newReconciler(mgr, opts))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager, opts Options) reconcile.Reconciler {
	connectorServiceClient := NewConnectorServiceClient(opts.ConnectorServiceURL)

	return &ReconcileTokenRequest{
		Client:                 mgr.GetClient(),
		scheme:                 mgr.GetScheme(),
		connectorServiceClient: connectorServiceClient,
		options:                opts,
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	c, err := controller.New("tokenrequest-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &applicationconnectorv1alpha1.TokenRequest{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileTokenRequest{}

// ReconcileTokenRequest reconciles a TokenRequest object
type ReconcileTokenRequest struct {
	client.Client
	scheme                 *runtime.Scheme
	connectorServiceClient ConnectorServiceClient
	options                Options
}

// Reconcile reads that state of the cluster for a TokenRequest object and makes changes based on the state read
// and what is in the TokenRequest.Spec
func (r *ReconcileTokenRequest) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	log.Printf("Processing TokenRequest %s", request.NamespacedName)

	instance := &applicationconnectorv1alpha1.TokenRequest{}
	err := r.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Printf("TokenRequest %s not found", request.NamespacedName)
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	if instance.ShouldExpire() {
		log.Printf("TokenRequest %s expired, deleting", request.NamespacedName)

		err = r.deleteTokenRequest(instance)

		return reconcile.Result{}, err
	} else if instance.ShouldFetch() {
		log.Printf("Fetching token for TokenRequest %s", request.NamespacedName)

		tokenDto, err := r.connectorServiceClient.FetchToken(instance.ObjectMeta.Name, instance.Context.Tenant, instance.Context.Group)
		if err != nil {
			log.Printf("Fetching token for TokenRequest %s failed", request.NamespacedName)

			instance.Status.State = applicationconnectorv1alpha1.TokenRequestStateERR
		} else {
			log.Printf("Token for TokenRequest %s fetched", request.NamespacedName)

			instance.Status.State = applicationconnectorv1alpha1.TokenRequestStateOK
			instance.Status.Token = tokenDto.Token
			instance.Status.URL = tokenDto.URL
			instance.Status.ExpireAfter = metav1.NewTime(metav1.Now().Add(time.Second * time.Duration(r.options.TokenTTL)))
		}

		instance.Status.Application = instance.ObjectMeta.Name

		log.Printf("Updating CR for TokenRequest %s", request.NamespacedName)

		err = r.updateTokenRequest(instance)

		return reconcile.Result{}, err
	}

	log.Printf("TokenRequest %s looks fine, nothing to do", request.NamespacedName)

	return reconcile.Result{}, nil
}

func (r *ReconcileTokenRequest) updateTokenRequest(tokenRequest *applicationconnectorv1alpha1.TokenRequest) error {
	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		return r.Update(context.Background(), tokenRequest)

	})
}

func (r *ReconcileTokenRequest) deleteTokenRequest(tokenRequest *applicationconnectorv1alpha1.TokenRequest) error {
	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		return r.Delete(context.TODO(), tokenRequest)
	})
}
