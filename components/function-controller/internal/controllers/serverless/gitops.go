package serverless

import (
	"context"
	"fmt"
	fluxv1Beta "github.com/fluxcd/source-controller/api/v1beta1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/kyma-project/kyma/components/function-controller/internal/git"
	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (r *FunctionReconciler) syncRevision(instance *serverlessv1alpha1.Function, options git.Options) (string, error) {
	if instance.Spec.Type == serverlessv1alpha1.SourceTypeGit {
		return r.gitOperator.LastCommit(options)
	}
	return "", nil
}

func NextRequeue(err error) (res ctrl.Result, errMsg string) {
	if git.IsNotRecoverableError(err) {
		return ctrl.Result{Requeue: false}, fmt.Sprintf("Stop reconciliation, reason: %s", err.Error())
	}

	errMsg = fmt.Sprintf("Sources update failed, reason: %v", err)
	if git.IsAuthErr(err) {
		errMsg = "Authorization to git server failed"
	}

	// use exponential delay
	return ctrl.Result{Requeue: true}, errMsg
}

func (r *FunctionReconciler) readGITOptions(ctx context.Context, instance *serverlessv1alpha1.Function) (git.Options, error) {
	if instance.Spec.Type != serverlessv1alpha1.SourceTypeGit {
		return git.Options{}, nil
	}

	r.Log.Info("Start Reconciliation")
	gitRepository := fluxv1Beta.GitRepository{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: GitRepoName, Namespace: "flux-system"}, &gitRepository)
	if err != nil {
		return git.Options{}, err
	}

	//var auth *git.AuthOptions
	//if gitRepository.Spec.Auth != nil {
	//	var secret corev1.Secret
	//	if err := r.client.Get(ctx, client.ObjectKey{Namespace: instance.Namespace, Name: gitRepository.Spec.Auth.SecretName}, &secret); err != nil {
	//		return git.Options{}, err
	//	}
	//	auth = &git.AuthOptions{
	//		Type:        git.RepositoryAuthType(gitRepository.Spec.Auth.Type),
	//		Credentials: r.readSecretData(secret.Data),
	//		SecretName:  gitRepository.Spec.Auth.SecretName,
	//	}
	//}

	if instance.Spec.Reference == "" {
		return git.Options{}, fmt.Errorf("reference has to specified")
	}

	return git.Options{
		URL:       gitRepository.Status.Artifact.URL,
		Reference: instance.Spec.Reference,
		//Auth:      auth,
	}, nil
}
