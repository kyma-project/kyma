package gitrepo

import (
	"context"

	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	serverlessv1alpha2 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha2"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type GitRepoReconciler struct {
	Log    *zap.SugaredLogger
	client client.Client
}

func NewGitRepoUpdateController(client client.Client, log *zap.SugaredLogger) *GitRepoReconciler {
	return &GitRepoReconciler{
		client: client,
		Log:    log,
	}
}

// should only match updates ?
func (r *GitRepoReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("gitrepo-update-controller").
		For(&serverlessv1alpha1.GitRepository{}).
		Complete(r)
}

func (r *GitRepoReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	// if deleted do nothing
	repo := &serverlessv1alpha1.GitRepository{}
	if err := r.client.Get(ctx, request.NamespacedName, repo); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	if !repo.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	functionList := &serverlessv1alpha1.FunctionList{}
	if err := r.client.List(ctx, functionList, &client.ListOptions{Namespace: repo.Namespace}); err != nil || len(functionList.Items) == 0 {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if err := r.updateV1Alpha2FunctionsWithRepo(ctx, functionList.Items, repo); err != nil {
		return ctrl.Result{}, errors.Wrap(err, "while updating v1alpha2 functions")
	}
	return ctrl.Result{}, nil
}

func (r *GitRepoReconciler) updateV1Alpha2FunctionsWithRepo(ctx context.Context, v1alpha1List []serverlessv1alpha1.Function, repo *serverlessv1alpha1.GitRepository) error {
	for _, f := range v1alpha1List {
		if f.Spec.Type != serverlessv1alpha1.SourceTypeGit {
			continue
		}
		if f.Spec.Source != repo.Name {
			continue
		}
		if err := r.updateV1Alpha2FunctionWithRepo(ctx, &f, repo); err != nil {
			return errors.Wrap(err, "while updating v1alpha2 function")
		}
	}
	return nil
}

func (r *GitRepoReconciler) updateV1Alpha2FunctionWithRepo(ctx context.Context, f *serverlessv1alpha1.Function, repo *serverlessv1alpha1.GitRepository) error {
	r.Log.With(
		"APIVersion", f.APIVersion,
		"namespace", f.Namespace,
		"function", f.Name,
		"git repo", repo.Name).Info("applying Git Repository update to function")

	v1alpha2Function := &serverlessv1alpha2.Function{}

	if err := r.client.Get(ctx, types.NamespacedName{Name: f.Name, Namespace: f.Namespace}, v1alpha2Function); err != nil {
		return errors.Wrap(err, "while getting v1alpha2 function")
	}
	UpdatedRepo := v1alpha2Function.Spec.Source.GitRepository.DeepCopy()
	UpdatedRepo.URL = repo.Spec.URL
	if repo.Spec.Auth != nil {
		UpdatedRepo.Auth = &serverlessv1alpha2.RepositoryAuth{
			Type:       serverlessv1alpha2.RepositoryAuthType(repo.Spec.Auth.Type),
			SecretName: repo.Spec.Auth.SecretName,
		}
	} else { // auth config is removed
		UpdatedRepo.Auth = nil
	}

	v1alpha2Function.Spec.Source.GitRepository = UpdatedRepo

	if err := r.client.Update(ctx, v1alpha2Function); err != nil {
		return errors.Wrap(err, "while updating v1alpha2 function")
	}
	return nil
}
