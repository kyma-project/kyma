package serverless

import (
	"context"
	"testing"

	resourceFake "github.com/kyma-project/kyma/components/console-backend-service/internal/resource/fake"
	"github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	name       = "test-git-repository"
	namespace  = "test-namespace"
	authType   = v1alpha1.RepositoryAuthBasic
	url        = "test-url"
	secretName = "test-secret-name"
)

func TestGitRepository_Queries(t *testing.T) {
	t.Run("Should list a GitRepositories", func(t *testing.T) {
		gitRepositoryAuth := fixMockRepositoryAuth(authType, secretName)
		gitRepository1 := fixMockGitRepository(name, namespace, url, gitRepositoryAuth)
		gitRepository2 := fixMockGitRepository(name, "test-namespace-2", url, gitRepositoryAuth)
		resolver := fixMockGitRepositoryResolver(t, gitRepository1, gitRepository2)

		result, err := resolver.GitRepositoriesQuery(context.Background(), "test-namespace-2")

		require.NoError(t, err)
		assert.Equal(t, len(result), 1)
		assert.Equal(t, gitRepository2, result[0])
	})
}

func TestGitRepository_Query(t *testing.T) {
	t.Run("Should find a GitRepository", func(t *testing.T) {
		gitRepositoryAuth := fixMockRepositoryAuth(authType, secretName)
		gitRepository := fixMockGitRepository(name, namespace, url, gitRepositoryAuth)
		resolver := fixMockGitRepositoryResolver(t, gitRepository)

		result, err := resolver.GitRepositoryQuery(context.Background(), namespace, name)

		require.NoError(t, err)
		assert.Equal(t, gitRepository, result)
	})

	t.Run("Should return error if not found", func(t *testing.T) {
		resolver := fixMockGitRepositoryResolver(t)
		_, err := resolver.GitRepositoryQuery(context.Background(), namespace, name)
		require.Error(t, err)
	})
}

func TestGitRepository_Create(t *testing.T) {
	t.Run("Should create a GitRepository", func(t *testing.T) {
		gitRepositoryAuth := fixMockRepositoryAuth(authType, secretName)
		gitRepository := fixMockGitRepository(name, namespace, url, gitRepositoryAuth)
		resolver := fixMockGitRepositoryResolver(t)

		result, err := resolver.CreateGitRepository(context.Background(), namespace, name, gitRepository.Spec)
		require.NoError(t, err)

		require.NoError(t, err)
		assert.Equal(t, result.Name, name)
		assert.Equal(t, result.Namespace, namespace)
		assert.Equal(t, result.Spec, gitRepository.Spec)
	})

	t.Run("Should throw an error if a GitRepository exists", func(t *testing.T) {
		gitRepositoryAuth := fixMockRepositoryAuth(authType, secretName)
		gitRepository := fixMockGitRepository(name, namespace, url, gitRepositoryAuth)
		resolver := fixMockGitRepositoryResolver(t, gitRepository)

		_, err := resolver.CreateGitRepository(context.Background(), namespace, name, gitRepository.Spec)
		require.Error(t, err)
	})
}

func TestGitRepository_Update(t *testing.T) {
	t.Run("Should update a GitRepository", func(t *testing.T) {
		gitRepositoryAuth := fixMockRepositoryAuth(authType, secretName)
		gitRepository := fixMockGitRepository(name, namespace, url, gitRepositoryAuth)
		resolver := fixMockGitRepositoryResolver(t, gitRepository)

		gitRepositoryUpdated := fixMockGitRepository(name, namespace, "new-url", gitRepositoryAuth)
		result, err := resolver.UpdateGitRepository(context.Background(), namespace, name, gitRepositoryUpdated.Spec)

		require.NoError(t, err)
		assert.Equal(t, result.Name, name)
		assert.Equal(t, result.Namespace, namespace)
		assert.Equal(t, result.Spec, gitRepositoryUpdated.Spec)
	})

	t.Run("Should throw an error if a GitRepository doesn't exist", func(t *testing.T) {
		gitRepositoryAuth := fixMockRepositoryAuth(authType, secretName)
		gitRepository := fixMockGitRepository(name, namespace, url, gitRepositoryAuth)
		resolver := fixMockGitRepositoryResolver(t)
		_, err := resolver.UpdateGitRepository(context.Background(), namespace, name, gitRepository.Spec)
		require.Error(t, err)
	})
}

func TestGitRepository_Delete(t *testing.T) {
	t.Run("Should delete a GitRepository", func(t *testing.T) {
		gitRepositoryAuth := fixMockRepositoryAuth(authType, secretName)
		gitRepository := fixMockGitRepository(name, namespace, url, gitRepositoryAuth)
		resolver := fixMockGitRepositoryResolver(t, gitRepository)

		_, err := resolver.DeleteGitRepository(context.Background(), namespace, name)
		require.NoError(t, err)
	})

	t.Run("Should throw an error if a GitRepository doesn't exist", func(t *testing.T) {
		resolver := fixMockGitRepositoryResolver(t)
		_, err := resolver.DeleteGitRepository(context.Background(), namespace, name)
		require.Error(t, err)
	})
}

func fixMockGitRepositoryResolver(t *testing.T, items ...runtime.Object) *NewResolver {
	serviceFactory, err := resourceFake.NewFakeGenericServiceFactory(v1alpha1.AddToScheme, items...)
	require.NoError(t, err)

	resolver := NewR(serviceFactory)
	err = resolver.Enable()
	require.NoError(t, err)

	serviceFactory.InformerFactory.WaitForCacheSync(make(chan struct{}))
	return resolver
}

func fixMockGitRepository(name, namespace string, url string, auth *v1alpha1.RepositoryAuth) *v1alpha1.GitRepository {
	return &v1alpha1.GitRepository{
		TypeMeta: v1.TypeMeta{
			APIVersion: gitRepositoriesGroupVersionResource.GroupVersion().String(),
			Kind:       gitRepositoryKind,
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1alpha1.GitRepositorySpec{
			URL:  url,
			Auth: auth,
		},
	}
}

func fixMockRepositoryAuth(authType v1alpha1.RepositoryAuthType, secretName string) *v1alpha1.RepositoryAuth {
	return &v1alpha1.RepositoryAuth{
		Type:       authType,
		SecretName: secretName,
	}
}
