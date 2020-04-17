package knative

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/kyma-project/kyma/components/function-controller/internal/resource"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	serviceLabelKey = "serving.knative.dev/service"
)

type ServiceConfig struct {
	RequeueDuration time.Duration `envconfig:"default=1m"`
}

type ServiceReconciler struct {
	client.Client
	Log logr.Logger

	config         ServiceConfig
	resourceClient resource.Resource
	recorder       record.EventRecorder
	scheme         *runtime.Scheme
}

func NewServiceReconciler(client client.Client, log logr.Logger, cfg ServiceConfig, scheme *runtime.Scheme, recorder record.EventRecorder) *ServiceReconciler {
	resourceClient := resource.New(client, scheme)

	return &ServiceReconciler{
		Client:         client,
		Log:            log,
		config:         cfg,
		resourceClient: resourceClient,
		scheme:         scheme,
		recorder:       recorder,
	}
}

func (r *ServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&servingv1.Service{}).
		Owns(&servingv1.Revision{}).
		Complete(r)
}

// Reconcile reads that state of the cluster for a Function object and makes changes based on the state read and what is in the Function.Spec
// +kubebuilder:rbac:groups="serving.knative.dev",resources=services;revisions,verbs=get;list;watch;create;update;patch;delete;deletecollection
// +kubebuilder:rbac:groups="serving.knative.dev",resources=services/status,verbs=get
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch

func (r *ServiceReconciler) Reconcile(request ctrl.Request) (ctrl.Result, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	instance := &servingv1.Service{}
	err := r.Get(ctx, request.NamespacedName, instance)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if !instance.Status.IsReady() {
		return ctrl.Result{RequeueAfter: r.config.RequeueDuration}, nil
	}

	log := r.Log.WithValues("kind", instance.GetObjectKind().GroupVersionKind().Kind, "name", instance.GetName(), "namespace", instance.GetNamespace(), "version", instance.GetGeneration())

	log.Info("Listing Revisions")
	var revisions servingv1.RevisionList
	if err := r.resourceClient.ListByLabel(ctx, instance.GetNamespace(), r.serviceLabel(*instance), &revisions); err != nil {
		log.Error(err, "Cannot list Revisions")
		return ctrl.Result{}, err
	}

	rev, _ := json.Marshal(revisions)

	fmt.Printf("%s", string(rev))

	return ctrl.Result{}, nil
}

func (r *ServiceReconciler) serviceLabel(s servingv1.Service) map[string]string {
	return map[string]string{
		serviceLabelKey: s.Name,
	}
}
