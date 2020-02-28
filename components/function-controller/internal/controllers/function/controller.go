package function

import (
	"context"

	"github.com/kyma-project/kyma/components/function-controller/internal/controllers"

	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/go-logr/logr"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// FunctionReconciler reconciles a Function object
type FunctionReconciler struct {
	client.Client
	Log logr.Logger

	config FunctionConfig
}

type FunctionConfig struct {
	MaxConcurrentReconciles int `envconfig:"default=1"`
}

func NewController(config FunctionConfig, log logr.Logger, di *controllers.Container) *FunctionReconciler {
	return &FunctionReconciler{
		Client: di.Manager.GetClient(),
		Log:    log,
		config: config,
	}
}

// Reconcile reads that state of the cluster for a Function object and makes changes based on the state read
// and what is in the Function.Spec
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;watch;update;list
// +kubebuilder:rbac:groups="apps;extensions",resources=deployments,verbs=create;get;watch;update;delete;list;update;patch
// +kubebuilder:rbac:groups="serverless.kyma-project.io",resources=functions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="serverless.kyma-project.io",resources=functions/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="admissionregistration.k8s.io",resources=mutatingwebhookconfigurations;validatingwebhookconfigurations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="serving.knative.dev",resources=services;routes;configurations;revisions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="build.knative.dev",resources=builds;buildtemplates;clusterbuildtemplates;services,verbs=get;list;create;update;delete;patch;watch
// +kubebuilder:rbac:groups="tekton.dev",resources=tasks;taskruns,verbs=get;list;watch;create;update;patch;delete
func (r *FunctionReconciler) Reconcile(req reconcile.Request) (reconcile.Result, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	function := &serverlessv1alpha1.Function{}
	if err := r.Client.Get(ctx, req.NamespacedName, function); err != nil {
		if apiErrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *FunctionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&serverlessv1alpha1.Function{}).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: r.config.MaxConcurrentReconciles,
		}).
		Complete(r)
}
