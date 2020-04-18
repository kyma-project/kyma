package knative

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/go-logr/logr"
	"github.com/kyma-project/kyma/components/function-controller/internal/controllers/serverless"
	"github.com/kyma-project/kyma/components/function-controller/internal/resource"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	apilabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/tools/record"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
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

func (r *ServiceReconciler) getPredicate() predicate.Predicate {
	var log = r.Log.WithName("predicates")
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			o := e.Object.(*unstructured.Unstructured)
			log.Info("Skipping reconciliation for dependent resource creation", "name", o.GetName(), "namespace", o.GetNamespace(), "apiVersion", o.GroupVersionKind().GroupVersion(), "kind", o.GroupVersionKind().Kind)
			return false
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			newObj := e.ObjectNew.(*unstructured.Unstructured).DeepCopy()
			log.Info("Reconciling due to dependent resource update", "name", newObj.GetName(), "namespace", newObj.GetNamespace(), "apiVersion", newObj.GroupVersionKind().GroupVersion(), "kind", newObj.GroupVersionKind().Kind)
			return true
		},
		GenericFunc: func(e event.GenericEvent) bool {
			o := e.Object.(*unstructured.Unstructured)
			log.Info("Reconcile due to generic event", "name", o.GetName(), "namespace", o.GetNamespace(), "apiVersion", o.GroupVersionKind().GroupVersion(), "kind", o.GroupVersionKind().Kind)
			return true
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			o := e.Object.(*unstructured.Unstructured)
			log.Info("Skipping reconciliation for dependent resource deletion", "name", o.GetName(), "namespace", o.GetNamespace(), "apiVersion", o.GroupVersionKind().GroupVersion(), "kind", o.GroupVersionKind().Kind)
			return false
		},
	}
}

func (r *ServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&servingv1.Service{}).
		Owns(&servingv1.Revision{}).
		WithEventFilter(r.getPredicate()).
		Complete(r)
}

// Reconcile reads that state of the cluster for a Function object and makes changes based on the state read and what is in the Function.Spec
// +kubebuilder:rbac:groups="serving.knative.dev",resources=services;revisions,verbs=get;list;watch;create;update;patch;delete;deletecollection
// +kubebuilder:rbac:groups="serving.knative.dev",resources=services/status,verbs=get
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch

func reconcileMiddleware(result ctrl.Result, err error) (ctrl.Result, error) {
	if err != nil {
		return ctrl.Result{RequeueAfter: 5 * time.Minute}, err
	}

	// if we set requeue manually, leave it be
	if result.RequeueAfter != 0*time.Second {
		return result, nil
	}

	return ctrl.Result{RequeueAfter: 30 * time.Hour}, nil
}

func (r *ServiceReconciler) Reconcile(request ctrl.Request) (ctrl.Result, error) {
	return reconcileMiddleware(r.rawReconcile(request))
}

func (r *ServiceReconciler) rawReconcile(request ctrl.Request) (ctrl.Result, error) {
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

func (r *ServiceReconciler) getOldRevisionSelector(parentService string, revisions []servingv1.Revision) (apilabels.Selector, error) {
	maxGen, err := getNewestGeneration(revisions)
	if err != nil {
		return nil, err
	}

	selector := apilabels.NewSelector()
	uuidReq, err := apilabels.NewRequirement(serviceLabelKey, selection.Equals, []string{parentService})
	if err != nil {
		return nil, err
	}
	generationReq, err := apilabels.NewRequirement(serverless.CfgGenerationLabel, selection.NotEquals, []string{strconv.Itoa(maxGen)})
	if err != nil {
		return nil, err
	}

	return selector.Add(*uuidReq, *generationReq), nil
}

func getNewestGeneration(revisions []servingv1.Revision) (int, error) {
	maxGeneration := -1
	for _, revision := range revisions {
		generationString, ok := revision.Labels[serverless.CfgGenerationLabel]
		if !ok {
			// todo extract to var
			return -1, errors.New(fmt.Sprintf("Revision %s in namespace %s doesn't have %s label", revision.Name, revision.Namespace, serverless.CfgGenerationLabel))
		}
		generation, err := strconv.Atoi(generationString)
		if err != nil {
			// todo extract to var
			return -1, errors.New(fmt.Sprintf("Couldn't convert label key %s to number, revision %s in namespace %s", generationString, revision.Name, revision.Namespace))
		}
		if generation > maxGeneration {
			maxGeneration = generation
		}
	}
	return maxGeneration, nil
}

func (r *ServiceReconciler) deleteRevisions(ctx context.Context, log logr.Logger, service *servingv1.Service, revisions []servingv1.Revision) (ctrl.Result, error) {
	log.Info("Deleting all old revisions")
	selector, err := r.getOldRevisionSelector(service.Name, revisions)
	if err != nil {
		log.Error(err, "Cannot create proper selector for old revisions")
		return ctrl.Result{}, err
	}

	if err := r.resourceClient.DeleteAllBySelector(ctx, &servingv1.Revision{}, service.GetNamespace(), selector); err != nil {
		log.Error(err, "Cannot delete old Revisions")
		return ctrl.Result{}, err
	}
	log.Info("Old Revisions deleted")
	return ctrl.Result{}, nil
}
