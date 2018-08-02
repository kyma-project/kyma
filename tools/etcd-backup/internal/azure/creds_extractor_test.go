package azure_test

import (
	"errors"
	"testing"

	"github.com/kyma-project/kyma/tools/etcd-backup/internal/azure"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	coreTypes "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	k8sTesting "k8s.io/client-go/testing"
)

func TestExtractCredsFromSecretSuccess(t *testing.T) {
	// given
	var (
		fixStorageAccountName = "myk-myk"
		fixStorageAccountKey  = "123-456-88901-ab"
	)

	fixSecretResource := fixSecret(map[string][]byte{
		"storage-account": []byte(fixStorageAccountName),
		"storage-key":     []byte(fixStorageAccountKey),
	})
	k8sFakeCli := fake.NewSimpleClientset(fixSecretResource)

	// when
	out, err := azure.ExtractCredsFromSecret(fixSecretResource.Name, k8sFakeCli.CoreV1().Secrets(fixSecretResource.Namespace))

	// then
	require.NoError(t, err)
	assert.Equal(t, fixStorageAccountName, out.AccountName)
	assert.Equal(t, fixStorageAccountKey, out.AccountKey)
}

func TestExtractCredsFromSecretFailure(t *testing.T) {
	t.Run("on getting secret from k8s", func(t *testing.T) {
		// given
		k8sFakeCli := fake.NewSimpleClientset()
		k8sFakeCli.PrependReactor(failingRector(onGetSecret()))

		// when
		out, err := azure.ExtractCredsFromSecret("fix-name", k8sFakeCli.CoreV1().Secrets("fix-ns"))

		// then
		require.EqualError(t, err, `while getting secret "fix-name": custom error`)
		assert.Nil(t, out)
	})

	t.Run("on extracting data field", func(t *testing.T) {
		tests := map[string]struct {
			givenData map[string][]byte
		}{
			"is nil":   {givenData: nil},
			"is empty": {givenData: map[string][]byte{}},
		}
		for tn, tc := range tests {
			t.Run(tn, func(t *testing.T) {
				// given
				fixSecretResource := fixSecret(tc.givenData)
				k8sFakeCli := fake.NewSimpleClientset(fixSecretResource)

				// when
				out, err := azure.ExtractCredsFromSecret(fixSecretResource.Name, k8sFakeCli.CoreV1().Secrets(fixSecretResource.Namespace))

				// then
				require.EqualError(t, err, "Secret \"fix-sec-name\" data cannot be empty")
				assert.Nil(t, out)
			})
		}

	})
}

func onGetSecret() (string, string) {
	return "get", "secrets"
}

func failingRector(verb, resource string) (string, string, k8sTesting.ReactionFunc) {
	failingFn := func(action k8sTesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, nil, errors.New("custom error")
	}
	return verb, resource, failingFn
}

func fixSecret(data map[string][]byte) *coreTypes.Secret {
	return &coreTypes.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "fix-sec-name",
			Namespace: "fix-ns-name",
		},
		Data: data,
	}
}
