package certificates

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseCertHeader(t *testing.T) {

	t.Run("Should return valid CertInfo for standalone Connector", func(t *testing.T) {
		//given
		r, _ := http.NewRequest("GET", "", nil)
		r.Header.Set(ClientCertHeader, "Hash=f4cf22fb633d4df500e371daf703d4b4d14a0ea9d69cd631f95f9e6ba840f8ad;Subject=\"CN=test-application,OU=OrgUnit,O=organization,L=Waldorf,ST=Waldorf,C=DE\";URI=spiffe://cluster.local/ns/kyma-integration/sa/default;"+
			"Hash=6d1f9f3a6ac94ff925841aeb9c15bb3323014e3da2c224ea7697698acf413226;Subject=\"\";URI=spiffe://cluster.local/ns/istio-system/sa/istio-ingressgateway-service-account")

		hp := NewHeaderParser("DE", "Waldorf", "Waldorf", "organization", "OrgUnit", false)

		//when
		certInfo, e := hp.ParseCertificateHeader(*r)

		require.NoError(t, e)

		//then
		assert.Equal(t, "f4cf22fb633d4df500e371daf703d4b4d14a0ea9d69cd631f95f9e6ba840f8ad", certInfo.Hash)
		assert.Equal(t, "CN=test-application,OU=OrgUnit,O=organization,L=Waldorf,ST=Waldorf,C=DE", certInfo.Subject)
	})

	t.Run("Should return error when unable to find matching CertInfo", func(t *testing.T) {
		//given
		r, _ := http.NewRequest("GET", "", nil)
		r.Header.Set(ClientCertHeader, "Hash=f4cf22fb633d4df500e371daf703d4b4d14a0ea9d69cd631f95f9e6ba840f8ad;Subject=\"\";URI=spiffe://cluster.local/ns/kyma-integration/sa/default;"+
			"Hash=6d1f9f3a6ac94ff925841aeb9c15bb3323014e3da2c224ea7697698acf413226;Subject=\"\";URI=spiffe://cluster.local/ns/istio-system/sa/istio-ingressgateway-service-account")

		hp := NewHeaderParser("DE", "Waldorf", "Waldorf", "organization", "OrgUnit", false)

		//when
		_, e := hp.ParseCertificateHeader(*r)

		require.Error(t, e)
	})

	t.Run("Should return valid CertInfo for central Connector", func(t *testing.T) {
		//given
		r, _ := http.NewRequest("GET", "", nil)
		r.Header.Set(ClientCertHeader, "Hash=f4cf22fb633d4df500e371daf703d4b4d14a0ea9d69cd631f95f9e6ba840f8ad;Subject=\"CN=test-application,OU=OrgUnit,O=organization,L=Waldorf,ST=Waldorf,C=DE\";URI=spiffe://cluster.local/ns/kyma-integration/sa/default;"+
			"Hash=6d1f9f3a6ac94ff925841aeb9c15bb3323014e3da2c224ea7697698acf413226;Subject=\"\";URI=spiffe://cluster.local/ns/istio-system/sa/istio-ingressgateway-service-account")

		hp := NewHeaderParser("DE", "Waldorf", "Waldorf", "organization", "OrgUnit", true)

		//when
		certInfo, e := hp.ParseCertificateHeader(*r)

		require.NoError(t, e)

		//then
		assert.Equal(t, "f4cf22fb633d4df500e371daf703d4b4d14a0ea9d69cd631f95f9e6ba840f8ad", certInfo.Hash)
		assert.Equal(t, "CN=test-application,OU=OrgUnit,O=organization,L=Waldorf,ST=Waldorf,C=DE", certInfo.Subject)
	})

	t.Run("Should return error when Certificate header is empty", func(t *testing.T) {
		//given
		r, _ := http.NewRequest("GET", "", nil)

		hp := NewHeaderParser("DE", "Waldorf", "Waldorf", "organization", "OrgUnit", false)

		//when
		_, e := hp.ParseCertificateHeader(*r)

		require.Error(t, e)
	})
}
