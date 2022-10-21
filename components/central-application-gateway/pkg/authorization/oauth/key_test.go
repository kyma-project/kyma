package oauth

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGenerateKey(t *testing.T) {
	t.Run("should", func(t *testing.T) {
		//given
		clientID := "clientID"
		privateKey := "-----BEGIN PRIVATE KEY-----\nMIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQDFyjWCE2FiVs5H\n1+KTs6DAaxCmytVFSykHyJYmvYBSw1TI8+Ho1WWKOY8q/EtYVpYdaq0ILeGFhA6z\nkl32VYK8oJER+cyG65ivpIoFCGS52VWyjinrjcFxsEf4S4vgl/QPbaSZz32tHH1h\n56wMnhRR8fLGsY+k2VVwpTduag9EUAQsvlO3r0aEr53/g2yTmPgwx5GGHiDMamB6\n9g+6ULN42RxGnWhgUnDLQfXY0yN/pwUKC/ukrTMy/+ImVDuskJvQnPLkV70FoPwq\nJ3gYUA3QfQ/lRLpt9N87NFVdVsJcVruNaR7B7Ta8Ghlu5eXndy+VRNU+rS8KdL4j\n1jExeMrHAgMBAAECggEAVpsOp/jFfRJme8XXg/Y1Dtwyq94H2bIp8qsNuEPlAxhd\nsSo9Ar8iGY7PljJn6XPsgk/6GSlB5T0oVM/jzd+ugdrK+vSG7pMNxecFumNs+4Xj\nRO6EA40MJbRbJykpQ/w1VWYcm27j6F+ftTWEu/eiDSmktQT90WCKzWrCpVnSeoXL\npssENiEzhU3lsgK+M8bj7GmCiSqz1Ki0qVQzxH1DJeOJ/XI7TYjhYdU+Lchm+PMK\nTUBEtgqEh0GD6XzQ70zjEIzGtpToMTsmIPRhC5t+vNwZp/2ZyONFkj74fP84XZbs\nDW5Ji9JfAyDudPSGSe8+EzA9TBGhd4ik1LmQcvvedQKBgQDV/YS/a+2bQlIrhkdH\n2tWfwuJwAjW0D61t/JftAY7GR3Nys963/4T+rd2TsDs9MF+zACEg7cKeTaC5IgHP\nQ8+TGyHeVaxn+ZnAfF5lrHjXBeMFWZIus/rBDogRy+0mh1oR58Fa0gvFmRBnvaxq\n/q+g1B0/kuJG5k3PrePA99EOBQKBgQDsnoH+uv8/uwRihzgVH47u3XBPwoflSYqi\nezhHWpeZkQ8qfEsX89KUnDc2TGoqBDxHTwcBNu4OQ1cTi0xgMgRFvNiwnyQGlykp\nJixu+MTdcegZbxcS0ippawerm97YNxH6X8LZflguZjVqN/nCk3tJs9iiZyjxNBRW\nRBxvyH/DWwKBgQCpCDEr4900nxa5OsBjigDkydSEFbrGGPwtvTFlDa3yAc639E0h\nmr07T6uPVc31b5iolJmWoTjyQu+KTcqQJkh5Mx11uscM+qTw30zRk4OAli3VtAM8\n0P5qMUhahnM11ATZz+90Bic2VsoWqETh33xr1iGkbio/Rvx/6CPX8ek44QKBgDMx\nXAijpoPAT4ONo8mWKVNun1TyTnqB/beHlzaA2BnGc5SKjaih/OZgIeXihHmQrwXy\niB5wJvL5CMbWtXB+gcQgxnT4CVBPtf0MIELmGZmbgk62ZTSSOdDS8jbjo0P+LiqQ\nO1TY6/Ul8dqIP8YkKGFawrzoOshsrxW26LwakeHPAoGBAKOs3CswKEVU23SY/vsL\nUMMciKXOclS77P+et2aQpodyqd8zDf8Zo4AzXDP7P1hndR3DFN7DK7FfzfGjbzoI\neOrbYDKM/8g/7G2BD53isaqRxXe0mbCsnGGS9qW0LnqbZHroIHzkSfaE07RZSuy3\ncGULnAlIuR23/9VjSUP7wAO2\n-----END PRIVATE KEY-----\n"
		certificate := "-----BEGIN CERTIFICATE-----\nMIIDXDCCAkSgAwIBAgIUBX/p1ZN7UFuCsygfIOHmEDpVHtswDQYJKoZIhvcNAQEL\nBQAwOzELMAkGA1UEBhMCUEwxCjAIBgNVBAgMAUExDDAKBgNVBAoMA1NBUDESMBAG\nA1UEAwwJbG9jYWxob3N0MB4XDTIyMTAwMzEwMzYxMloXDTIzMTAwMzEwMzYxMlow\nOzELMAkGA1UEBhMCUEwxCjAIBgNVBAgMAUExDDAKBgNVBAoMA1NBUDESMBAGA1UE\nAwwJbG9jYWxob3N0MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAxco1\nghNhYlbOR9fik7OgwGsQpsrVRUspB8iWJr2AUsNUyPPh6NVlijmPKvxLWFaWHWqt\nCC3hhYQOs5Jd9lWCvKCREfnMhuuYr6SKBQhkudlVso4p643BcbBH+EuL4Jf0D22k\nmc99rRx9YeesDJ4UUfHyxrGPpNlVcKU3bmoPRFAELL5Tt69GhK+d/4Nsk5j4MMeR\nhh4gzGpgevYPulCzeNkcRp1oYFJwy0H12NMjf6cFCgv7pK0zMv/iJlQ7rJCb0Jzy\n5Fe9BaD8Kid4GFAN0H0P5US6bfTfOzRVXVbCXFa7jWkewe02vBoZbuXl53cvlUTV\nPq0vCnS+I9YxMXjKxwIDAQABo1gwVjAUBgNVHREEDTALgglsb2NhbGhvc3QwHQYD\nVR0OBBYEFP/P8Vy9b+Kvx9t6i5TVOjiD5OT1MB8GA1UdIwQYMBaAFI62bpw2BVd6\n5l3PN3wR83xxhk0VMA0GCSqGSIb3DQEBCwUAA4IBAQCWNO04okw24eoQVdapxkZP\n+YiCRwV9AWUvssr9qccrXZCVpERBVTFu1rx20KDenU8u8weGTu9Esx7uzkn6zaqV\n83mNYJi4FjrVMRz75YdvMjIG8E0/+9P3/Zw+3ui5HFD5e2pPgN03EgXivM/BswGz\nxctkAC04lu2bvkGHeyzURSMB65Wtv+YvaGC7WigdO+PQavStGGOuv4koIbs3ZNyg\nh2LJ7Uc6TiRSEHTnics+tsBbvy23v4At9hSw5xdicCe/TODcTcmZutelnHp0NjH1\nHiRJdUhfEnQm3VhdJGLhrO19QU4cD9TKp5csixZgY2DUqnsZAerwOqccJN1bfAvT\n-----END CERTIFICATE-----\n"
		authURL := "www.example.com"
		certSha := "6e268674edb6685600ffcb61552c900c6ea9d42d391c63e188fc7ccff967f86a"
		keySha := "0ebb97467eb55862b26c5c10ec25a57114bc7e25a99530a52b7f5fdb5ff0f377"
		expectedKey := clientID + "-" + certSha + "-" + keySha + "-" + authURL

		//when
		//tlsCert, err := tls.X509KeyPair([]byte(certificate), []byte(privateKey))
		key, err := generateKey(clientID, certificate, privateKey, authURL)

		//then
		require.NoError(t, err)
		require.Equal(t, expectedKey, key)
	})
}

func BenchmarkBase64decode(b *testing.B) {
	//given
	clientID := "clientID"
	privateKey := "-----BEGIN PRIVATE KEY-----\nMIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQDFyjWCE2FiVs5H\n1+KTs6DAaxCmytVFSykHyJYmvYBSw1TI8+Ho1WWKOY8q/EtYVpYdaq0ILeGFhA6z\nkl32VYK8oJER+cyG65ivpIoFCGS52VWyjinrjcFxsEf4S4vgl/QPbaSZz32tHH1h\n56wMnhRR8fLGsY+k2VVwpTduag9EUAQsvlO3r0aEr53/g2yTmPgwx5GGHiDMamB6\n9g+6ULN42RxGnWhgUnDLQfXY0yN/pwUKC/ukrTMy/+ImVDuskJvQnPLkV70FoPwq\nJ3gYUA3QfQ/lRLpt9N87NFVdVsJcVruNaR7B7Ta8Ghlu5eXndy+VRNU+rS8KdL4j\n1jExeMrHAgMBAAECggEAVpsOp/jFfRJme8XXg/Y1Dtwyq94H2bIp8qsNuEPlAxhd\nsSo9Ar8iGY7PljJn6XPsgk/6GSlB5T0oVM/jzd+ugdrK+vSG7pMNxecFumNs+4Xj\nRO6EA40MJbRbJykpQ/w1VWYcm27j6F+ftTWEu/eiDSmktQT90WCKzWrCpVnSeoXL\npssENiEzhU3lsgK+M8bj7GmCiSqz1Ki0qVQzxH1DJeOJ/XI7TYjhYdU+Lchm+PMK\nTUBEtgqEh0GD6XzQ70zjEIzGtpToMTsmIPRhC5t+vNwZp/2ZyONFkj74fP84XZbs\nDW5Ji9JfAyDudPSGSe8+EzA9TBGhd4ik1LmQcvvedQKBgQDV/YS/a+2bQlIrhkdH\n2tWfwuJwAjW0D61t/JftAY7GR3Nys963/4T+rd2TsDs9MF+zACEg7cKeTaC5IgHP\nQ8+TGyHeVaxn+ZnAfF5lrHjXBeMFWZIus/rBDogRy+0mh1oR58Fa0gvFmRBnvaxq\n/q+g1B0/kuJG5k3PrePA99EOBQKBgQDsnoH+uv8/uwRihzgVH47u3XBPwoflSYqi\nezhHWpeZkQ8qfEsX89KUnDc2TGoqBDxHTwcBNu4OQ1cTi0xgMgRFvNiwnyQGlykp\nJixu+MTdcegZbxcS0ippawerm97YNxH6X8LZflguZjVqN/nCk3tJs9iiZyjxNBRW\nRBxvyH/DWwKBgQCpCDEr4900nxa5OsBjigDkydSEFbrGGPwtvTFlDa3yAc639E0h\nmr07T6uPVc31b5iolJmWoTjyQu+KTcqQJkh5Mx11uscM+qTw30zRk4OAli3VtAM8\n0P5qMUhahnM11ATZz+90Bic2VsoWqETh33xr1iGkbio/Rvx/6CPX8ek44QKBgDMx\nXAijpoPAT4ONo8mWKVNun1TyTnqB/beHlzaA2BnGc5SKjaih/OZgIeXihHmQrwXy\niB5wJvL5CMbWtXB+gcQgxnT4CVBPtf0MIELmGZmbgk62ZTSSOdDS8jbjo0P+LiqQ\nO1TY6/Ul8dqIP8YkKGFawrzoOshsrxW26LwakeHPAoGBAKOs3CswKEVU23SY/vsL\nUMMciKXOclS77P+et2aQpodyqd8zDf8Zo4AzXDP7P1hndR3DFN7DK7FfzfGjbzoI\neOrbYDKM/8g/7G2BD53isaqRxXe0mbCsnGGS9qW0LnqbZHroIHzkSfaE07RZSuy3\ncGULnAlIuR23/9VjSUP7wAO2\n-----END PRIVATE KEY-----\n"
	certificate := "-----BEGIN CERTIFICATE-----\nMIIDXDCCAkSgAwIBAgIUBX/p1ZN7UFuCsygfIOHmEDpVHtswDQYJKoZIhvcNAQEL\nBQAwOzELMAkGA1UEBhMCUEwxCjAIBgNVBAgMAUExDDAKBgNVBAoMA1NBUDESMBAG\nA1UEAwwJbG9jYWxob3N0MB4XDTIyMTAwMzEwMzYxMloXDTIzMTAwMzEwMzYxMlow\nOzELMAkGA1UEBhMCUEwxCjAIBgNVBAgMAUExDDAKBgNVBAoMA1NBUDESMBAGA1UE\nAwwJbG9jYWxob3N0MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAxco1\nghNhYlbOR9fik7OgwGsQpsrVRUspB8iWJr2AUsNUyPPh6NVlijmPKvxLWFaWHWqt\nCC3hhYQOs5Jd9lWCvKCREfnMhuuYr6SKBQhkudlVso4p643BcbBH+EuL4Jf0D22k\nmc99rRx9YeesDJ4UUfHyxrGPpNlVcKU3bmoPRFAELL5Tt69GhK+d/4Nsk5j4MMeR\nhh4gzGpgevYPulCzeNkcRp1oYFJwy0H12NMjf6cFCgv7pK0zMv/iJlQ7rJCb0Jzy\n5Fe9BaD8Kid4GFAN0H0P5US6bfTfOzRVXVbCXFa7jWkewe02vBoZbuXl53cvlUTV\nPq0vCnS+I9YxMXjKxwIDAQABo1gwVjAUBgNVHREEDTALgglsb2NhbGhvc3QwHQYD\nVR0OBBYEFP/P8Vy9b+Kvx9t6i5TVOjiD5OT1MB8GA1UdIwQYMBaAFI62bpw2BVd6\n5l3PN3wR83xxhk0VMA0GCSqGSIb3DQEBCwUAA4IBAQCWNO04okw24eoQVdapxkZP\n+YiCRwV9AWUvssr9qccrXZCVpERBVTFu1rx20KDenU8u8weGTu9Esx7uzkn6zaqV\n83mNYJi4FjrVMRz75YdvMjIG8E0/+9P3/Zw+3ui5HFD5e2pPgN03EgXivM/BswGz\nxctkAC04lu2bvkGHeyzURSMB65Wtv+YvaGC7WigdO+PQavStGGOuv4koIbs3ZNyg\nh2LJ7Uc6TiRSEHTnics+tsBbvy23v4At9hSw5xdicCe/TODcTcmZutelnHp0NjH1\nHiRJdUhfEnQm3VhdJGLhrO19QU4cD9TKp5csixZgY2DUqnsZAerwOqccJN1bfAvT\n-----END CERTIFICATE-----\n"
	authURL := "www.example.com"
	certSha := "6e268674edb6685600ffcb61552c900c6ea9d42d391c63e188fc7ccff967f86a"
	keySha := "0ebb97467eb55862b26c5c10ec25a57114bc7e25a99530a52b7f5fdb5ff0f377"
	expectedKey := clientID + "-" + certSha + "-" + keySha + "-" + authURL

	//when
	//tlsCert, err := tls.X509KeyPair([]byte(certificate), []byte(privateKey))
	key, err := generateKey(clientID, certificate, privateKey, authURL)

	//then
	require.NoError(b, err)
	require.Equal(b, expectedKey, key)
}
