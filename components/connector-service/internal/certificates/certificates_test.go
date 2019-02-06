package certificates

import (
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"testing"

	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	privateKey = `-----BEGIN PRIVATE KEY-----
MIIEvwIBADANBgkqhkiG9w0BAQEFAASCBKkwggSlAgEAAoIBAQClJYXxG+pUe7oa
XBHnbxR+5oEMtD3Ft01TXhq0Ad/0+5+qgSZ1GNE6t8dO9q5syo237ZQ1kWHXIs0T
sCD2aWYsAIlr/R9f0ED3oOkiYx0DVV8849+eEcaKFhyzUBkm1zns+MjunYBWyR0o
J/uJO3mAszd8wTNRoEd5X4KcKTikIMkIttip35dcH6Nf5jDF0QIamOET3rp4T2rm
A3Vc+v0xChzEdKiTiVaT9LLaKeL1OUplJa90MoHZ4zHLLqMFiX12j/rdWzVkyi2M
0dS8ynEtpVhvxvna5vaooY3yr2SDKyv4+Zf2ZKrfenS5Dru8QVrGH8yDhOhTQeFC
kB+LA6RNAgMBAAECggEAMd3DtQs11a7Kgh0c9uIOsUbO3tQp9uKjgbHfpE0Qn/u+
uZBn2WHWA8Hsd8Z64rTC2C/v2cD9ZyXGANTlDyLCTDUZSbdT2u2aQGuhGdYNs6z6
pfs00ZkSdy24Gtjrz1Ob1RdGLO74CryNhkuUY1rHFHqJHa2E3nfkPRz+5kJ4LO6R
Mt4m1XKpNdi6Eg/WurnCKiIgq8yPckhYRuMhF3aNNevvqKGlL4QSYHMCjXAw18E4
XqhcH5Y8ZhUR8Nv87N6dgTOHGXNYHyGN1ZpNii6+jb+JgqeitbUXI4SCh9ev+S/Y
G5qE/DjoP6JtcQH/2GoFJw5muIijt/C6fUESQH/0YQKBgQDRvA8EEkamKaWVl0MM
BeY90GqDklkmALctQ+4zgSAm//BzofGZWYyNWexV30xlAgz5xKbBU1xh7f48BR7C
La21HnTPgcp+h/GLywKPxD117drjX1GfMLZQRK4QQhNF1WNvv9NTWfq/2wmQCvtD
FLdobgRoEbR5YRvcYglRPD+F1QKBgQDJk4ev6NUGC1Gg387En9QJ7aE3xWwhdsGr
R5zsIk+/dM7Qzbw6aAm5Ui4ZUPyWcHmb2qlnSsJHwv8AOi5IJJCalxfrHYrE4TJz
RjZ2PVst7Y1uvGHhSlID39/NEEW033wKQ2MWdPwYG15eBL6pYX1ThaLhBhMUMGFQ
dxL3UmYImQKBgQDC67BY7FNUomgNuuLJDcKJuGUFmsHXm9qh6vw6ScuD82GZVeyf
xKXnyKbot/rb9SfyCV2hVsQJD5K0XV3UwXcrWP7ey5VSOy216hqbWpp0O3au0iud
czw9JVdQLNiUklkzxme0k2+DVyJwCIS0N1CtcXIO9kVweVvXWhWmtgOjcQKBgQCy
5tn1OOrfi2ouIpSLk+KH0TxVmEUoyhKG5m8ScD1hCdWIIiBdofqHXLWHSIZ1Kmvz
9DSHdSVKtXjGhdyPsMwaN+FFjZmctNWm03kApeHnuD7fOhiQ7/oscCRcBoYnSnX3
Uel+g+M9rgSp4wIoqFqnpyJxHogOUgX8eUH++UWPeQKBgQC/NBa3TNFgFcn5eWZS
7qsVMFeIhvXOdKqrBhbIR32C2px+OW93TbMOKdnpQSFlbmvYBkqAQ+TsjC3r8u33
xwfuqKS+KeByw2+7ac53rRZ806IMjZNHiX+N9HakgqdQdM8XwNG2GIpPQ2VOMCu0
iEWRwvmD+/sYdSyhdEAZYoAndA==
-----END PRIVATE KEY-----`

	cert = `-----BEGIN CERTIFICATE-----
MIIDhDCCAmwCCQCCgClWcqHk4DANBgkqhkiG9w0BAQsFADCBgzELMAkGA1UEBhMC
VVMxDjAMBgNVBAgMBXN0YXRlMQ0wCwYDVQQHDARjaXR5MRAwDgYDVQQKDAdjb21w
YW55MRAwDgYDVQQLDAdzZWN0aW9uMRQwEgYDVQQDDAtob3N0LmV4LmNvbTEbMBkG
CSqGSIb3DQEJARYMZW1haWxAZXguY29tMB4XDTE4MDYxMTA4MDMyM1oXDTIxMDMz
MTA4MDMyM1owgYMxCzAJBgNVBAYTAlVTMQ4wDAYDVQQIDAVzdGF0ZTENMAsGA1UE
BwwEY2l0eTEQMA4GA1UECgwHY29tcGFueTEQMA4GA1UECwwHc2VjdGlvbjEUMBIG
A1UEAwwLaG9zdC5leC5jb20xGzAZBgkqhkiG9w0BCQEWDGVtYWlsQGV4LmNvbTCC
ASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAKUlhfEb6lR7uhpcEedvFH7m
gQy0PcW3TVNeGrQB3/T7n6qBJnUY0Tq3x072rmzKjbftlDWRYdcizROwIPZpZiwA
iWv9H1/QQPeg6SJjHQNVXzzj354RxooWHLNQGSbXOez4yO6dgFbJHSgn+4k7eYCz
N3zBM1GgR3lfgpwpOKQgyQi22Knfl1wfo1/mMMXRAhqY4RPeunhPauYDdVz6/TEK
HMR0qJOJVpP0stop4vU5SmUlr3QygdnjMcsuowWJfXaP+t1bNWTKLYzR1LzKcS2l
WG/G+drm9qihjfKvZIMrK/j5l/Zkqt96dLkOu7xBWsYfzIOE6FNB4UKQH4sDpE0C
AwEAATANBgkqhkiG9w0BAQsFAAOCAQEAYccH2RdbliHmEVhTajfId66xl0lmwTVx
rVkMRvtEJ1M8rIwABVCu/w+DSorm8sNq8n9ZCwhXflFCEySk8wevg5/lLtSz4ghn
A97O/CNEABohwLZXQYkOQqGDXz6yWmCugtt8Y5of16NDj2AzqXZ++tUvo/CvB/Q8
1iL+JpgQs15b0QEIpXRyxOAc5FdHm+I9qtx+BeA3I3tMPhYlM9mDVev8fdHtURN8
9QM4wWFHncmNvlTK51HPexFI3TF9sEqDUQ7dozcUD8GexHlsvZh95+5TmSlA0kfl
fWXUGZObOGD246zwfHLHP3AwzFKU0bfIvqckcw23I+ZUMIbdajw9eg==
-----END CERTIFICATE-----`

	clientCRT = `-----BEGIN CERTIFICATE-----
MIIDPDCCAiSgAwIBAgIBAjANBgkqhkiG9w0BAQsFADCBgzELMAkGA1UEBhMCVVMx
DjAMBgNVBAgMBXN0YXRlMQ0wCwYDVQQHDARjaXR5MRAwDgYDVQQKDAdjb21wYW55
MRAwDgYDVQQLDAdzZWN0aW9uMRQwEgYDVQQDDAtob3N0LmV4LmNvbTEbMBkGCSqG
SIb3DQEJARYMZW1haWxAZXguY29tMB4XDTE4MDkyNjA5MjQ0MFoXDTE4MTIyNTA5
MjQ0MFowFjEUMBIGA1UEAxMLaG9zdC5leC5jb20wggEiMA0GCSqGSIb3DQEBAQUA
A4IBDwAwggEKAoIBAQCxFMs7w1t30mZqV8qK8/7PkJTdIWz5AOJEIu0LiQYOibSC
xnQSj9FiUEnxU3CjWqtxP4VcShCco78mgeY7rpSFKWz3JWhwBr3VkgSFHPCQWNN0
wLUDu7WVBkVkiDUs+Nbz3/Po9oh4+syZKfrdkHVXpSepoU8Gl+Kb8Apb55Q/RW/T
0MZHJ4sJps0j3fjpyQUKH9pt9768Aw3cJ8KuS8MtgNS46+PnjtA02tebzQdfO6F6
mi7PfG9Wp7AX9cJNukGiZaMjsk93OrPei4c3CIkKEBsqVtC4yVXwk/szUD+KxSDr
Eq+NwfxHY5pemHVIN2U4YuNfK+xEtG2trVIIu8oTAgMBAAGjJzAlMA4GA1UdDwEB
/wQEAwIHgDATBgNVHSUEDDAKBggrBgEFBQcDAjANBgkqhkiG9w0BAQsFAAOCAQEA
On0/O1iBwRA+bCNguRaIaHojLqEENAVneNA7HbRYLwIN1nUwfZvII1ZsKs0xo5M+
1XfLukKDTOIWE6NvQ4q1Y5zzMHVg5/N+o5tMze+aZxvtlBKfV2dgwddnwgCK/huO
G6gfxQO88Y7JpZmLmIl4TLH4a2TFH/t1rEQNXE8e+HwNCKOYxhnYfvvt6U1pZhNz
XExXcKBJ5oiblhW+NiqoiSHxRk9JWV679Wa51nML66khttQOUCZzVkAMhIPIJc0k
JEEx2RazbgxRj23+bclb/nrPQj4X1G5d2JsvM6jcRiyrp/llfQOn3TgiqtiIUCA0
JK2K4FJavFZ2tpvqVXyQpg==
-----END CERTIFICATE-----`

	CSR = `-----BEGIN CERTIFICATE REQUEST-----
MIICWzCCAUMCAQAwFjEUMBIGA1UEAwwLaG9zdC5leC5jb20wggEiMA0GCSqGSIb3
DQEBAQUAA4IBDwAwggEKAoIBAQCxFMs7w1t30mZqV8qK8/7PkJTdIWz5AOJEIu0L
iQYOibSCxnQSj9FiUEnxU3CjWqtxP4VcShCco78mgeY7rpSFKWz3JWhwBr3VkgSF
HPCQWNN0wLUDu7WVBkVkiDUs+Nbz3/Po9oh4+syZKfrdkHVXpSepoU8Gl+Kb8Apb
55Q/RW/T0MZHJ4sJps0j3fjpyQUKH9pt9768Aw3cJ8KuS8MtgNS46+PnjtA02teb
zQdfO6F6mi7PfG9Wp7AX9cJNukGiZaMjsk93OrPei4c3CIkKEBsqVtC4yVXwk/sz
UD+KxSDrEq+NwfxHY5pemHVIN2U4YuNfK+xEtG2trVIIu8oTAgMBAAGgADANBgkq
hkiG9w0BAQsFAAOCAQEAVmUqt+hkv7urWE2eRzqFuHoi5QCfXaex77+3Avyt+cAh
Fjk80VLgDBIRJF7dBZr6wemQ+h36Fj5zD51Ijgol4VnNDQ42GWiYpocAhXGQM1gV
YBRBMUOSP4o5nMrRrqlU3HmYXxG5ZeyYnAn+r2sXXxgyLaFZcUqTJ4GTptYLl7yP
/aYw+brjj9dKL4sFsSdmzSWI82zBFaLuC+RFCi4Ra6pBmo1nogw4GzLZP/U1LBKL
8nLBy/Vrh5qH+VQmSHGSCxuJpxI4/eqmYdEn+7ORjIimUqjCDl9juthRjJf4upv2
SFQVvLnLavw1WaO4XGDPYqfPvwlnrYIwc1AGBegN8g==
-----END CERTIFICATE REQUEST-----`

	CrtChain = `-----BEGIN CERTIFICATE-----
MIIDPDCCAiSgAwIBAgIBAjANBgkqhkiG9w0BAQsFADCBgzELMAkGA1UEBhMCVVMx
DjAMBgNVBAgMBXN0YXRlMQ0wCwYDVQQHDARjaXR5MRAwDgYDVQQKDAdjb21wYW55
MRAwDgYDVQQLDAdzZWN0aW9uMRQwEgYDVQQDDAtob3N0LmV4LmNvbTEbMBkGCSqG
SIb3DQEJARYMZW1haWxAZXguY29tMB4XDTE4MDkyNjA5MjQ0MFoXDTE4MTIyNTA5
MjQ0MFowFjEUMBIGA1UEAxMLaG9zdC5leC5jb20wggEiMA0GCSqGSIb3DQEBAQUA
A4IBDwAwggEKAoIBAQCxFMs7w1t30mZqV8qK8/7PkJTdIWz5AOJEIu0LiQYOibSC
xnQSj9FiUEnxU3CjWqtxP4VcShCco78mgeY7rpSFKWz3JWhwBr3VkgSFHPCQWNN0
wLUDu7WVBkVkiDUs+Nbz3/Po9oh4+syZKfrdkHVXpSepoU8Gl+Kb8Apb55Q/RW/T
0MZHJ4sJps0j3fjpyQUKH9pt9768Aw3cJ8KuS8MtgNS46+PnjtA02tebzQdfO6F6
mi7PfG9Wp7AX9cJNukGiZaMjsk93OrPei4c3CIkKEBsqVtC4yVXwk/szUD+KxSDr
Eq+NwfxHY5pemHVIN2U4YuNfK+xEtG2trVIIu8oTAgMBAAGjJzAlMA4GA1UdDwEB
/wQEAwIHgDATBgNVHSUEDDAKBggrBgEFBQcDAjANBgkqhkiG9w0BAQsFAAOCAQEA
On0/O1iBwRA+bCNguRaIaHojLqEENAVneNA7HbRYLwIN1nUwfZvII1ZsKs0xo5M+
1XfLukKDTOIWE6NvQ4q1Y5zzMHVg5/N+o5tMze+aZxvtlBKfV2dgwddnwgCK/huO
G6gfxQO88Y7JpZmLmIl4TLH4a2TFH/t1rEQNXE8e+HwNCKOYxhnYfvvt6U1pZhNz
XExXcKBJ5oiblhW+NiqoiSHxRk9JWV679Wa51nML66khttQOUCZzVkAMhIPIJc0k
JEEx2RazbgxRj23+bclb/nrPQj4X1G5d2JsvM6jcRiyrp/llfQOn3TgiqtiIUCA0
JK2K4FJavFZ2tpvqVXyQpg==
-----END CERTIFICATE-----
-----BEGIN CERTIFICATE-----
MIIDhDCCAmwCCQCCgClWcqHk4DANBgkqhkiG9w0BAQsFADCBgzELMAkGA1UEBhMC
VVMxDjAMBgNVBAgMBXN0YXRlMQ0wCwYDVQQHDARjaXR5MRAwDgYDVQQKDAdjb21w
YW55MRAwDgYDVQQLDAdzZWN0aW9uMRQwEgYDVQQDDAtob3N0LmV4LmNvbTEbMBkG
CSqGSIb3DQEJARYMZW1haWxAZXguY29tMB4XDTE4MDYxMTA4MDMyM1oXDTIxMDMz
MTA4MDMyM1owgYMxCzAJBgNVBAYTAlVTMQ4wDAYDVQQIDAVzdGF0ZTENMAsGA1UE
BwwEY2l0eTEQMA4GA1UECgwHY29tcGFueTEQMA4GA1UECwwHc2VjdGlvbjEUMBIG
A1UEAwwLaG9zdC5leC5jb20xGzAZBgkqhkiG9w0BCQEWDGVtYWlsQGV4LmNvbTCC
ASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAKUlhfEb6lR7uhpcEedvFH7m
gQy0PcW3TVNeGrQB3/T7n6qBJnUY0Tq3x072rmzKjbftlDWRYdcizROwIPZpZiwA
iWv9H1/QQPeg6SJjHQNVXzzj354RxooWHLNQGSbXOez4yO6dgFbJHSgn+4k7eYCz
N3zBM1GgR3lfgpwpOKQgyQi22Knfl1wfo1/mMMXRAhqY4RPeunhPauYDdVz6/TEK
HMR0qJOJVpP0stop4vU5SmUlr3QygdnjMcsuowWJfXaP+t1bNWTKLYzR1LzKcS2l
WG/G+drm9qihjfKvZIMrK/j5l/Zkqt96dLkOu7xBWsYfzIOE6FNB4UKQH4sDpE0C
AwEAATANBgkqhkiG9w0BAQsFAAOCAQEAYccH2RdbliHmEVhTajfId66xl0lmwTVx
rVkMRvtEJ1M8rIwABVCu/w+DSorm8sNq8n9ZCwhXflFCEySk8wevg5/lLtSz4ghn
A97O/CNEABohwLZXQYkOQqGDXz6yWmCugtt8Y5of16NDj2AzqXZ++tUvo/CvB/Q8
1iL+JpgQs15b0QEIpXRyxOAc5FdHm+I9qtx+BeA3I3tMPhYlM9mDVev8fdHtURN8
9QM4wWFHncmNvlTK51HPexFI3TF9sEqDUQ7dozcUD8GexHlsvZh95+5TmSlA0kfl
fWXUGZObOGD246zwfHLHP3AwzFKU0bfIvqckcw23I+ZUMIbdajw9eg==
-----END CERTIFICATE-----
`

	invalidCert = `-----BEGIN CERTIFICATE-----
-----END CERTIFICATE-----`

	invalidKey = `-----BEGIN RSA PRIVATE KEY-----
-----END RSA PRIVATE KEY-----`

	invalidCSR = `-----BEGIN CERTIFICATE REQUEST-----
-----END CERTIFICATE REQUEST-----`

	countryCode        = "US"
	province           = "state"
	locality           = "city"
	organization       = "company"
	organizationalUnit = "section"
	commonName         = "host.ex.com"
)

var (
	encodedKey         = []byte(privateKey)
	encodedCert        = []byte(cert)
	encodedInvalidCert = []byte(invalidCert)
	encodedInvalidKey  = []byte(invalidKey)
)

func TestCertificateUtility_LoadCert(t *testing.T) {

	t.Run("should load cert", func(t *testing.T) {
		// given
		certificateUtility := NewCertificateUtility()

		// when
		crt, err := certificateUtility.LoadCert(encodedCert)

		// then
		require.NoError(t, err)

		assert.Equal(t, countryCode, crt.Subject.Country[0])
		assert.Equal(t, province, crt.Subject.Province[0])
		assert.Equal(t, locality, crt.Subject.Locality[0])
		assert.Equal(t, organization, crt.Subject.Organization[0])
		assert.Equal(t, organizationalUnit, crt.Subject.OrganizationalUnit[0])
		assert.Equal(t, commonName, crt.Subject.CommonName)
		assert.Equal(t, commonName, crt.Subject.CommonName)
	})

	t.Run("should fail decoding cert", func(t *testing.T) {
		// given
		certificateUtility := NewCertificateUtility()

		// when
		crt, err := certificateUtility.LoadCert([]byte("invalid data"))

		// then
		require.Error(t, err)
		assert.Equal(t, apperrors.CodeInternal, err.Code())
		assert.Nil(t, crt)
	})

	t.Run("should fail parsing cert", func(t *testing.T) {
		// given
		certificateUtility := NewCertificateUtility()

		// when
		crt, err := certificateUtility.LoadCert(encodedInvalidCert)

		// then
		require.Error(t, err)
		assert.Equal(t, apperrors.CodeInternal, err.Code())
		assert.Nil(t, crt)
	})
}

func TestCertificateUtility_LoadKey(t *testing.T) {

	t.Run("should load key", func(t *testing.T) {
		// given
		certificateUtility := NewCertificateUtility()

		// when
		key, err := certificateUtility.LoadKey(encodedKey)

		// then
		require.NoError(t, err)
		assert.NotNil(t, key)
	})

	t.Run("should fail decoding key", func(t *testing.T) {
		// given
		certificateUtility := NewCertificateUtility()

		// when
		crt, err := certificateUtility.LoadKey([]byte("invalid data"))

		// then
		require.Error(t, err)
		assert.Equal(t, apperrors.CodeInternal, err.Code())
		assert.Nil(t, crt)
	})

	t.Run("should fail parsing key", func(t *testing.T) {
		// given
		certificateUtility := NewCertificateUtility()

		// when
		crt, err := certificateUtility.LoadKey(encodedInvalidKey)

		// then
		require.Error(t, err)
		assert.Equal(t, apperrors.CodeInternal, err.Code())
		assert.Nil(t, crt)
	})
}

func TestCertificateUtility_LoadCSR(t *testing.T) {

	t.Run("should load CSR", func(t *testing.T) {
		// given
		certificateUtility := NewCertificateUtility()

		// when
		key, err := certificateUtility.LoadCSR([]byte(CSR))

		// then
		require.NoError(t, err)
		assert.NotNil(t, key)
	})

	t.Run("should fail decoding CSR", func(t *testing.T) {
		// given
		certificateUtility := NewCertificateUtility()

		// when
		crt, err := certificateUtility.LoadCSR([]byte("aW52YWxpZCBkYXRh"))

		// then
		require.Error(t, err)
		assert.Equal(t, apperrors.CodeBadRequest, err.Code())
		assert.Nil(t, crt)
	})

	t.Run("should fail parsing CSR", func(t *testing.T) {
		// given
		certificateUtility := NewCertificateUtility()

		// when
		crt, err := certificateUtility.LoadCSR([]byte(invalidCSR))

		// then
		require.Error(t, err)
		assert.Equal(t, apperrors.CodeBadRequest, err.Code())
		assert.Nil(t, crt)
	})
}

func TestCertificateUtility_CheckCSRValues(t *testing.T) {

	csr := &x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName:         "cname",
			Country:            []string{"country"},
			Organization:       []string{"organization"},
			OrganizationalUnit: []string{"organizationalUnit"},
			Locality:           []string{"locality"},
			Province:           []string{"province"},
		},
	}

	t.Run("should successfully check CSR values", func(t *testing.T) {
		// given
		csrSubject := CSRSubject{
			CommonName:         "cname",
			Country:            "country",
			Organization:       "organization",
			OrganizationalUnit: "organizationalUnit",
			Locality:           "locality",
			Province:           "province",
		}

		certificateUtility := NewCertificateUtility()

		// when
		err := certificateUtility.CheckCSRValues(csr, csrSubject)

		// then
		require.NoError(t, err)
	})

	t.Run("should fail when subject country is nil", func(t *testing.T) {
		// given
		csrSubject := CSRSubject{
			CommonName: "cname",
		}

		csr := &x509.CertificateRequest{
			Subject: pkix.Name{
				CommonName: "cname",
			},
		}

		certificateUtility := NewCertificateUtility()

		// when
		err := certificateUtility.CheckCSRValues(csr, csrSubject)

		// then
		require.Error(t, err)
		assert.Equal(t, apperrors.CodeWrongInput, err.Code())
		assert.Contains(t, err.Error(), "No country")
	})

	t.Run("should fail when CommonName differs", func(t *testing.T) {
		// given
		csrSubject := CSRSubject{
			CommonName:         "differentCname",
			Country:            "country",
			Organization:       "organization",
			OrganizationalUnit: "organizationalUnit",
			Locality:           "locality",
			Province:           "province",
		}

		certificateUtility := NewCertificateUtility()

		// when
		err := certificateUtility.CheckCSRValues(csr, csrSubject)

		// then
		require.Error(t, err)
		assert.Equal(t, apperrors.CodeWrongInput, err.Code())
		assert.Contains(t, err.Error(), "Invalid common name")
	})

	t.Run("should fail when Country differs", func(t *testing.T) {
		// given
		csrSubject := CSRSubject{
			CommonName:         "cname",
			Country:            "invalidCountry",
			Organization:       "organization",
			OrganizationalUnit: "organizationalUnit",
			Locality:           "locality",
			Province:           "province",
		}

		certificateUtility := NewCertificateUtility()

		// when
		err := certificateUtility.CheckCSRValues(csr, csrSubject)

		// then
		require.Error(t, err)
		assert.Equal(t, apperrors.CodeWrongInput, err.Code())
		assert.Contains(t, err.Error(), "Invalid country")

	})

	t.Run("should fail when Organization differs", func(t *testing.T) {
		// given
		csrSubject := CSRSubject{
			CommonName:         "cname",
			Country:            "country",
			Organization:       "invalidOrganization",
			OrganizationalUnit: "organizationalUnit",
			Locality:           "locality",
			Province:           "province",
		}

		certificateUtility := NewCertificateUtility()

		// when
		err := certificateUtility.CheckCSRValues(csr, csrSubject)

		// then
		require.Error(t, err)
		assert.Equal(t, apperrors.CodeWrongInput, err.Code())
		assert.Contains(t, err.Error(), "CSR: Invalid organization provided.")
	})

	t.Run("should fail when OrganizationalUnit differs", func(t *testing.T) {
		// given
		csrSubject := CSRSubject{
			CommonName:         "cname",
			Country:            "country",
			Organization:       "organization",
			OrganizationalUnit: "invalidOrganizationalUnit",
			Locality:           "locality",
			Province:           "province",
		}

		certificateUtility := NewCertificateUtility()

		// when
		err := certificateUtility.CheckCSRValues(csr, csrSubject)

		// then
		require.Error(t, err)
		assert.Equal(t, apperrors.CodeWrongInput, err.Code())
		assert.Contains(t, err.Error(), "CSR: Invalid organizational unit provided.")
	})

	t.Run("should fail when Locality differs", func(t *testing.T) {
		// given
		csrSubject := CSRSubject{
			CommonName:         "cname",
			Country:            "country",
			Organization:       "organization",
			OrganizationalUnit: "organizationalUnit",
			Locality:           "invalidLocality",
			Province:           "province",
		}

		certificateUtility := NewCertificateUtility()

		// when
		err := certificateUtility.CheckCSRValues(csr, csrSubject)

		// then
		require.Error(t, err)
		assert.Equal(t, apperrors.CodeWrongInput, err.Code())
		assert.Contains(t, err.Error(), "CSR: Invalid locality provided.")
	})

	t.Run("should fail when Province differs", func(t *testing.T) {
		// given
		csrSubject := CSRSubject{
			CommonName:         "cname",
			Country:            "country",
			Organization:       "organization",
			OrganizationalUnit: "organizationalUnit",
			Locality:           "locality",
			Province:           "invalidProvince",
		}

		certificateUtility := NewCertificateUtility()

		// when
		err := certificateUtility.CheckCSRValues(csr, csrSubject)

		// then
		require.Error(t, err)
		assert.Equal(t, apperrors.CodeWrongInput, err.Code())
		assert.Contains(t, err.Error(), "CSR: Invalid province provided.")
	})
}

func TestCertificateUtility_SignCSR(t *testing.T) {

	t.Run("should sign client certificate", func(t *testing.T) {
		// given
		certificateUtility := NewCertificateUtility()
		caCrt, csr, key := prepareCrtAndKey(certificateUtility)

		// when
		rawClientCRT, apperr := certificateUtility.SignCSR(caCrt, csr, key)

		//then
		require.NoError(t, apperr)
		assert.NotEmpty(t, rawClientCRT)

		decodedCrt, err := x509.ParseCertificate(rawClientCRT)
		require.NoError(t, err)

		validityTime := calculateValidityTime(decodedCrt)
		assert.Equal(t, CertificateValidityDays, validityTime)
	})

	t.Run("should return when failed to create certificate", func(t *testing.T) {
		// given
		caCrt := &x509.Certificate{}
		csr := &x509.CertificateRequest{}
		key := &rsa.PrivateKey{}

		certificateUtility := NewCertificateUtility()

		// when
		rawClientCRT, err := certificateUtility.SignCSR(caCrt, csr, key)

		// then
		require.Error(t, err)
		assert.Equal(t, apperrors.CodeInternal, err.Code())
		assert.Nil(t, rawClientCRT)
	})

}

func TestCertificateUtility_AddCertificateHeaderAndFooter(t *testing.T) {

	t.Run("should add certificate header and footer", func(t *testing.T) {
		// given
		certificateUtility := NewCertificateUtility()
		certificate, apperr := certificateUtility.LoadCert([]byte(cert))
		require.NoError(t, apperr)

		// when
		encodedCert := certificateUtility.AddCertificateHeaderAndFooter(certificate.Raw)

		// then
		require.NotNil(t, encodedCert)

		// when
		pemBlock, _ := pem.Decode(encodedCert)
		decodedCert, err := x509.ParseCertificate(pemBlock.Bytes)

		// then
		require.NoError(t, err)
		assert.Equal(t, certificate, decodedCert)
	})

}

func calculateValidityTime(certificate *x509.Certificate) int {
	expirationDate := certificate.NotAfter
	fmt.Print(expirationDate.String())
	creationDate := certificate.NotBefore
	fmt.Println(creationDate.String())
	difference := expirationDate.Sub(creationDate)

	const hoursInDay = 24

	daysFloat := difference.Hours() / hoursInDay

	return int(daysFloat)
}

func prepareCrtAndKey(certificateUtility CertificateUtility) (*x509.Certificate, *x509.CertificateRequest, *rsa.PrivateKey) {
	caCrt, _ := certificateUtility.LoadCert(encodedCert)
	csr, _ := certificateUtility.LoadCSR([]byte(CSR))
	key, _ := certificateUtility.LoadKey(encodedKey)
	return caCrt, csr, key
}

func rawCrtTox509Certificates(rawCrt []byte) ([]*x509.Certificate, error) {
	pemBlock, rest := pem.Decode(rawCrt)

	if len(rest) == 0 {
		return nil, x509.CertificateInvalidError{}
	}

	pemBlock2, _ := pem.Decode(rest)

	pemBlocks := append(pemBlock.Bytes, pemBlock2.Bytes...)

	return x509.ParseCertificates(pemBlocks)
}
