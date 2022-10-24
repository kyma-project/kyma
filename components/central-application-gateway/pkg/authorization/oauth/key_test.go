package oauth

//
//import (
//	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/authorization/testconsts"
//	"github.com/stretchr/testify/require"
//	"testing"
//)
//
//func TestGenerateKey(t *testing.T) {
//	t.Run("should", func(t *testing.T) {
//		//given
//		clientID := "clientID"
//		certificate := []byte(testconsts.Certificate)
//		privateKey := []byte(testconsts.PrivateKey)
//		authURL := "www.example.com"
//		certSha := "764a894fc802acd8edfa2771e9e424c8868d5891a58a345f04e898a5cec06a21"
//		keySha := "840a2bfed372a0b2f01b0be877978bb5a56d9b83dee199fe11d55819d20ead18"
//		expectedKey := clientID + "-" + certSha + "-" + keySha + "-" + authURL
//
//		//when
//		key, err := generateKey(clientID, authURL, certificate, privateKey)
//
//		//then
//		require.NoError(t, err)
//		require.Equal(t, expectedKey, key)
//	})
//}
