package k8s_test

import (
	"errors"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s/automock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/api/core/v1"
	k8sTesting "k8s.io/client-go/testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"testing"
	"time"
)

func failingReactor(action k8sTesting.Action) (handled bool, ret runtime.Object, err error) {
	return true, nil, errors.New("custom error")
}

func TestSecretResolver(t *testing.T) {
	// GIVEN
	t1 := time.Unix(1, 0)
	expected := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "my-secret",
			Namespace:         "production",
			CreationTimestamp: metav1.NewTime(t1),
			Annotations:       map[string]string{"second-annot": "content"},
		},
	}
	resource := &v1.Secret{}
	resourceGetter := automock.NewSecretSvc()
	name := "name"
	namespace := "namespace"
	resourceGetter.On("Find", name, namespace).Return(resource, nil).Once()
	defer resourceGetter.AssertExpectations(t)

	converter := automock.NewGQLSecretConverter()
	converter.On("ToGQL", resource).Return(expected, nil).Once()
	defer converter.AssertExpectations(t)

	resolver := k8s.NewSecretResolver(resourceGetter)
	resolver.SetSecretConverter(converter)

	result, err := resolver.SecretQuery(nil, name, namespace)

	require.NoError(t, err)
	assert.Equal(t, expected, result)
	//resolver := k8s.NewSecretResolver(fakeClientSet.CoreV1())
	// WHEN
	//actualSecret, err := resolver.SecretQuery(context.Background(), "my-secret", "production")
	//// THEN
	//require.NoError(t, err)
	//assert.Equal(t, "my-secret", actualSecret.Name)
	//assert.Equal(t, "production", actualSecret.Namespace)
	//assert.Equal(t, t1, actualSecret.CreationTime)
	//assert.Equal(t, gqlschema.JSON{"second-annot": "content"}, actualSecret.Annotations)
}

//func TestSecretResolverOnNotFound(t *testing.T) {
//	// GIVEN
//	fakeClientSet := fake.NewSimpleClientset()
//	resolver := k8s.NewSecretResolver(fakeClientSet.CoreV1())
//	// WHEN
//	secret, err := resolver.SecretQuery(context.Background(), "my-secret", "production")
//	// THEN
//	assert.NoError(t, err)
//	assert.Nil(t, secret)
//}
//
//func TestSecretResolverOnError(t *testing.T) {
//	// GIVEN
//	fakeClientSet := fake.NewSimpleClientset()
//	fakeClientSet.PrependReactor("get", "secrets", failingReactor)
//	resolver := k8s.NewSecretResolver(fakeClientSet.CoreV1())
//	// WHEN
//	_, err := resolver.SecretQuery(context.Background(), "my-secret", "production")
//	// THEN
//	require.Error(t, err)
//	assert.True(t, gqlerror.IsInternal(err))
//}
//
//func failingReactor(action k8sTesting.Action) (handled bool, ret runtime.Object, err error) {
//	return true, nil, errors.New("custom error")
//}
//
//func TestSecretsResolver(t *testing.T) {
//	// GIVEN
//	t1 := time.Unix(1, 0)
//
//	fakeClientSet := fake.NewSimpleClientset(
//		&v1.SecretList{
//			Items: []v1.Secret{
//				v1.Secret{
//					ObjectMeta: metav1.ObjectMeta{
//						Name:              "my-secret",
//						Namespace:         "production",
//						CreationTimestamp: metav1.NewTime(t1),
//						Annotations:       map[string]string{"second-annot": "content"},
//					},
//				},
//			},
//		})
//
//	resolver := k8s.NewSecretResolver(fakeClientSet.CoreV1())
//	// WHEN
//	secretList, err := resolver.SecretsQuery(context.Background(), "production")
//	// THEN
//	require.NoError(t, err)
//	assert.Equal(t, "my-secret", secretList[0].Name)
//	assert.Equal(t, "production", secretList[0].Namespace)
//	assert.Equal(t, t1, secretList[0].CreationTime)
//	assert.Equal(t, gqlschema.JSON{"second-annot": "content"}, secretList[0].Annotations)
//	assert.Equal(t, len(secretList), 1)
//}
//func TestSecretsResolverOnNotFound(t *testing.T) {
//	// GIVEN
//	fakeClientSet := fake.NewSimpleClientset()
//	resolver := k8s.NewSecretResolver(fakeClientSet.CoreV1())
//	// WHEN
//	secretList, err := resolver.SecretsQuery(context.Background(), "production")
//	// THEN
//	assert.NoError(t, err)
//	assert.Nil(t, secretList)
//}
//func TestSecretsResolverOnError(t *testing.T) {
//	// GIVEN
//	fakeClientSet := fake.NewSimpleClientset()
//	fakeClientSet.PrependReactor("list", "secrets", failingReactor)
//	resolver := k8s.NewSecretResolver(fakeClientSet.CoreV1())
//	// WHEN
//	list, err := resolver.SecretsQuery(context.Background(), "production")
//	// THEN
//
//	require.Error(t, err)
//	assert.Nil(t, list)
//	assert.True(t, gqlerror.IsInternal(err))
//}
