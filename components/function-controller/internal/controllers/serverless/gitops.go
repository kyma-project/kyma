package serverless

import (
	"context"
	"fmt"

	"github.com/kyma-project/kyma/components/function-controller/internal/git"
	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *FunctionReconciler) syncRevision(instance *serverlessv1alpha1.Function, options git.Options) (string, error) {
	if instance.Spec.Type == serverlessv1alpha1.SourceTypeGit {
		return r.gitOperator.LastCommit(options)
	}
	return "", nil
}

func (r *FunctionReconciler) readGITOptions(ctx context.Context, instance *serverlessv1alpha1.Function) (git.Options, error) {
	if instance.Spec.Type != serverlessv1alpha1.SourceTypeGit {
		return git.Options{}, nil
	}

	var gitRepository serverlessv1alpha1.GitRepository
	if err := r.client.Get(ctx, client.ObjectKey{Namespace: instance.Namespace, Name: instance.Spec.Source}, &gitRepository); err != nil {
		return git.Options{}, err
	}

	var auth *git.AuthOptions
	if gitRepository.Spec.Auth != nil {
		var secret corev1.Secret
		if err := r.client.Get(ctx, client.ObjectKey{Namespace: instance.Namespace, Name: gitRepository.Spec.Auth.SecretName}, &secret); err != nil {
			return git.Options{}, err
		}
		auth = &git.AuthOptions{
			Type:        git.RepositoryAuthType(gitRepository.Spec.Auth.Type),
			Credentials: r.readSecretData(secret.Data),
			SecretName:  gitRepository.Spec.Auth.SecretName,
		}
	}

	if instance.Spec.Reference == "" {
		return git.Options{}, fmt.Errorf("reference has to specified")
	}

	return git.Options{
		URL:       gitRepository.Spec.URL,
		Reference: instance.Spec.Reference,
		Auth:      auth,
	}, nil
}
