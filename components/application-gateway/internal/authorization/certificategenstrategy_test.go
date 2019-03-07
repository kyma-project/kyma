package authorization

import (
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"net/http/httputil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	proxyStub = &httputil.ReverseProxy{}

	certificate = []byte(`-----BEGIN CERTIFICATE-----
MIICwDCCAaigAwIBAgIBAjANBgkqhkiG9w0BAQsFADAPMQ0wCwYDVQQDEwR0ZXN0
MB4XDTE5MDExNzExMDg0M1oXDTIwMDExNzExMDg0M1owDzENMAsGA1UEAxMEdGVz
dDCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBALG4LTb4U5AfzzZ+5eGv
37yiuBoG7NHl0JWrwv7gGNOUMgN6KzR4SbIhDxEerg0NXF33MaLHsBH7XpXdfu1K
gZzbos+jcMhq9obxByIpWZjVzDqGvFhtJ13GsXHky4Iz01wJkd7Lerbpe06eJfGA
iHgY9XOOl6Ckx0OXyiGwti1ab+Z17W28UuX4rloq15HWzgzWWGLv8dIeG79mKBLX
JjwPpBLQGFDUR29soI0tlcldNlJTDB4I5O04mdBKiJBlNs/k6UU5hPgARFP3vsy6
xbmECYsiV46RaIvh4pm4tkSvQ2WjaIL5V4oNc00STUPMgM36yxbCcSpOfvbJzwZD
TTUCAwEAAaMnMCUwDgYDVR0PAQH/BAQDAgeAMBMGA1UdJQQMMAoGCCsGAQUFBwMC
MA0GCSqGSIb3DQEBCwUAA4IBAQAXk05jHOpwvBtrKAapzy3zPSIL20KzUwFbE4Ey
FjEB1L5yiJ98DgkeBe2ayJtOI5qADq9ZeZ+Bq2QQZ+f2Y8lwm5TLtc3IAZUzI/sv
4NUj5NaiTeu/6kX165mWUtwZftetUHgJKqyQv1ClRNN6Ayfuv76Qq0DIEfHAvelD
rHWXo6ju4eD8IDRl+xv0wSbID/OUi9n5vBir+CJSIxmwW0jDKo7RZ//7gAm25sue
2oqaDiVVH2bsYbY41SRS/RrJqLYQQCqjNVAxhuGFV1uigz9LxIXpOfD2GrAVHBGJ
WHGZtRZO7LQJ2Yxy4mKYo0ndJOZIPAXzVveu+FY+CxSnWe4y
-----END CERTIFICATE-----`)
	privateKey = []byte(`-----BEGIN RSA PRIVATE KEY-----
	MIIEowIBAAKCAQEAsbgtNvhTkB/PNn7l4a/fvKK4Ggbs0eXQlavC/uAY05QyA3or
NHhJsiEPER6uDQ1cXfcxosewEfteld1+7UqBnNuiz6NwyGr2hvEHIilZmNXMOoa8
WG0nXcaxceTLgjPTXAmR3st6tul7Tp4l8YCIeBj1c46XoKTHQ5fKIbC2LVpv5nXt
bbxS5fiuWirXkdbODNZYYu/x0h4bv2YoEtcmPA+kEtAYUNRHb2ygjS2VyV02UlMM
Hgjk7TiZ0EqIkGU2z+TpRTmE+ABEU/e+zLrFuYQJiyJXjpFoi+Himbi2RK9DZaNo
gvlXig1zTRJNQ8yAzfrLFsJxKk5+9snPBkNNNQIDAQABAoIBAQCvDGI8ysxEIZDQ
dJ2pdrpB62S6/ic1d8/EHLKsqb7KaCX7FtKHlBPCyJ02l6mIYaihRWI2m8npbFb1
9n2W7NRN1GjBbJMiVXkN4UrNcz01NSE8ZIoP7zPfQl51eI6baMC+3w43DOWKVN+H
yG5HOfsxkCPz9neTW7qJ8XC/TneiotRD7lc2szXVrmpnGsebKiHL4zIKw2JvDox8
NU7tiGpab24Yn7M5CRyrZTWtzGADuheiCBi77DW0W87gpQSQJjOAZr7n92F1HnNR
6knBEBlG0ZkGOd5WSIkLw98Wvh4+4wY+yFlXjp7Q4xmJhEnZ/dfSXXsDYLB4peVB
1s2/KQ4BAoGBAOX288zW/imek38Odp0Gqffei+p7pb1B8kJe4s6fK8Vgzf9cRwGM
rVAeYmWM4pj/SAcY9V/60qbBI+gnLStBKMFSgDjOdtq60OK/i80z6+lCxt7LiLRj
hfk0yczWT+7v+uCJYAFa7qLVN+8O+V5tbcx2JkorTh1aowIVEaVr0VSxAoGBAMXX
AVZokkQFStn+TYkbNLeCYKVehxltbVF7aPDOmYxv60g7AHXAhKnCbCpdVpOFWufa
6ug+OR+eEVFy2ODbtaOgH/DnvqltkXITdqcXElV48skADn/K6xOiTDtZQ5NX30ql
YdwZo2EsPpWse2Xx6N2WJr+ry/C5d8g7naQTO3HFAoGAWeYUoPtbGMIZPw5UaEZ2
o6OoZt43iKkDH9cgK04mOl8BqNZWG9D2399A8BoHa3BApCWppv/S4cWXV+YYzlQG
rqyl2486/38QsdPXvzyQ+PtV6zr+Eibl9OoiCaWuUeYW2ThbA6ycpaNc3mOoMLXu
uoNlrJEJVIheOS4rW9OuXcECgYA7RGLJMQCIUhGPZqhxp23Of8dWIxBT5L04CMFy
SmIjeS/B7rL/k5HqjSz8MAQMo4mNJb7znhhcyWykusQP8KHkh8ap07MBbKqCwyPr
gHTkmBwbbOHrFK4BrsApk180F8HzycGcPy37oVaKXiaFbsf1AdNP3jZ1QgqJOJrM
GVrYhQKBgCNxoVJ+yi1PWWzd+erdSr/rqOwWEKIykDxAOtJUOjAUoH3t8qDzHDPx
Kw3UoadKDOno5X1xjDMGOe/s48bg0o3wsklC/C6QnYnIVCZUR7hvw865T1fQA5kB
eQL/kF0yUbR1b5deQ8Rq7x1UUQV1BcBFwfTaiAutq1sPRTNSHWtP
-----END RSA PRIVATE KEY-----`)
)

func TestCertificateGenStrategy(t *testing.T) {

	t.Run("should add certificates to proxy", func(t *testing.T) {
		// given
		proxy := &httputil.ReverseProxy{}

		expectedProxy := &httputil.ReverseProxy{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					Certificates: []tls.Certificate{
						{
							Certificate: [][]byte{cert()},
							PrivateKey:  key(),
						},
					},
				},
			},
		}

		certGenStrategy := newCertificateGenStrategy(certificate, privateKey)

		request, err := http.NewRequest("GET", "www.example.com", nil)
		require.NoError(t, err)

		// when
		err = certGenStrategy.AddAuthorization(request, proxy)

		// then
		require.NoError(t, err)
		assert.Equal(t, expectedProxy, proxy)
	})

	t.Run("should return error when key is invalid", func(t *testing.T) {
		// given
		proxy := &httputil.ReverseProxy{}

		certGenStrategy := newCertificateGenStrategy(certificate, []byte("invalid key"))

		request, err := http.NewRequest("GET", "www.example.com", nil)
		require.NoError(t, err)

		// when
		err = certGenStrategy.AddAuthorization(request, proxy)

		// then
		require.Error(t, err)
	})

	t.Run("should return error when certificate is invalid", func(t *testing.T) {
		// given
		proxy := &httputil.ReverseProxy{}

		certGenStrategy := newCertificateGenStrategy([]byte("invalid cert"), privateKey)

		request, err := http.NewRequest("GET", "www.example.com", nil)
		require.NoError(t, err)

		// when
		err = certGenStrategy.AddAuthorization(request, proxy)

		// then
		require.Error(t, err)
	})
}

func key() *rsa.PrivateKey {
	pemBlock, _ := pem.Decode(privateKey)
	key, _ := x509.ParsePKCS1PrivateKey(pemBlock.Bytes)
	return key
}

func cert() []byte {
	pemBlock, _ := pem.Decode(certificate)
	return pemBlock.Bytes
}
