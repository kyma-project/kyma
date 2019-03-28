package externalapi

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-project/kyma/components/connector-service/internal/revocation/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRevocationHandler(t *testing.T) {

	urlRevocation := "/v1/applications/certificates/revocations"
	hashedTestCert := "daf8a9b0927fc221bdc3319336e4236ab59a7c0e3986d4cf69898601cb8ee604"

	testCert := "-----BEGIN%20CERTIFICATE-----%0A" +
		"MIIFfjCCA2agAwIBAgIBAjANBgkqhkiG9w0BAQsFADBqMQswCQYDVQQGEwJQTDEK%0A" +
		"MAgGA1UECAwBTjEQMA4GA1UEBwwHR0xJV0lDRTETMBEGA1UECgwKU0FQIEh5YnJp%0A" +
		"czENMAsGA1UECwwES3ltYTEZMBcGA1UEAwwQd29ybWhvbGUua3ltYS5jeDAeFw0x%0A" +
		"OTAxMjQwOTQ1MTJaFw0yMDAxMjQwOTQ1MTJaMHIxCzAJBgNVBAYTAkRFMRAwDgYD%0A" +
		"VQQIEwdXYWxkb3JmMRAwDgYDVQQHEwdXYWxkb3JmMRUwEwYDVQQKEwxPcmdhbml6%0A" +
		"YXRpb24xEDAOBgNVBAsTB09yZ1VuaXQxFjAUBgNVBAMTDXVwcy10ZXN0LWFwcDEw%0A" +
		"ggIiMA0GCSqGSIb3DQEBAQUAA4ICDwAwggIKAoICAQC7eOxYnQy3CcNKYAYb2b48%0A" +
		"FA92GW1LWjfCBEhu+uP5hLlH572OODpdhyNmtFInpl8gmQ1454x%2FrG5bOjUxOkOi%0A" +
		"U97xM8e7O1MW4fLIC2LDkY0TapfaMn%2FoYwUHz3QFlx+BKM4yxgpJ+YY5OJVoGDEH%0A" +
		"TdaSY5LzftnfSBjUMrQIec6wv3+Vy0F8jUmxV3FifODhWxBUBEqZOSm9ruOwOSan%0A" +
		"J2sY1ovsj4QZZ9HnLJT6mdMt8lAwyDg0o68aOsjPtDEaqOIGWiYls49GIRi%2FcXft%0A" +
		"1S2L1q5WOiFlK18rkWouFJrx7YppPXEpLzwqY8DjaZ5qbfaBZzTaYHYBgTqClCbp%0A" +
		"AqwzDBsXuhPKywY+HzGhkGMHJKVxADuycgGObyQsnFDg0284zVJpCU41grVaigjf%0A" +
		"PsD8NU%2FhO11M+45xVl7WCLecHmVKZpwR3A6g6WpucitFKdHvbdD6jhbt3aUuJN+u%0A" +
		"U7GdnPk0A+y%2Fs7Y6FzxdoVtfWFIZqnFHIcL1kUT8WeEONnzmmMvOIIh8rz1U9iVO%0A" +
		"m%2FFZ1u3JloRRdGpaKewTytRsOKBE5Oug24GZSKA5dGG6ex7BDoXRgUgf3WxuUBQ2%0A" +
		"uidkXunTrcWUFI3IwbAdS9DWnlVhjvYBjOw8OsF3WtUsyIkDKFs11mfQvOKg6Y0m%0A" +
		"FrF9eKkyolUlmrHrGdG3wwIDAQABoycwJTAOBgNVHQ8BAf8EBAMCB4AwEwYDVR0l%0A" +
		"BAwwCgYIKwYBBQUHAwIwDQYJKoZIhvcNAQELBQADggIBAGpTE%2FtPB8ZJqR+mqZDB%0A" +
		"HzTE1fyoYYJvXfGwZQEWrR3GqOPe0Ni3jWQKs9NYoYSGoN9cYFvDJUQy3d5BJ1m8%0A" +
		"Y0uySIEDewYpKZR1zL7lnUYZ4msjOpNnlcIGNd0xz1EbHDPK3fMljVu+oJ4d+Z7a%0A" +
		"VdfJ+4fXUPgWxA9ou2E9GoiWtoReGc7ok2sx1NH3%2FZIaSy7zHKfve8ULrwkvVTEs%0A" +
		"W+jzuGIZvIQXMyY8c2M3sNb4sn14xRTZHA88rDJhI8Yjmhp9ZvaqAv+isaeBcGhY%0A" +
		"Cpspwgl1qqGqNsgEkoScW6WzN8HP0DvQ6UhsJKKhtTxl2Kd81PKGQIz%2F9vdkhOew%0A" +
		"irbwUBl1Wx1eoJlcTXFEMSGVVjHUUDDyftfq+hfJ4U90Ny0tjPjr+t9ejHOWfszf%0A" +
		"3Na0qT+92au76SFBddBVe3J0jaGhcwyFMHmQp8CU9ZKHXnSeumwplo33+25U3M%2Fe%0A" +
		"ALnDeejOZfFu0zDXTPZsvUFDJBke6PsILtQOAyrGATbDzPCqCZPmIl%2Fp4aMwgPJW%0A" +
		"zn6TyG%2F0YbJXH7Mysm+8k9qSWTpfT8YRDSrOUIXPFxg9jmMqF8vnksNq8cwto57Z%0A" +
		"gPzl0TAYNKqS%2FYDkeNOgrwlnJ%2Ff2rLTD0Zki16i%2FJ3s6%2FIy37XXj2No+G8bvGBDZ%0A" +
		"ffu3gnzaWei6HTmquQp55Kun%0A" +
		"-----END%20CERTIFICATE-----"

	t.Run("should revoke certificate and return http code 201", func(t *testing.T) {
		//given
		revocationListRepository := &mocks.RevocationListRepository{}
		revocationListRepository.On("Insert", hashedTestCert).Return(nil)

		handler := NewRevocationHandler(revocationListRepository)

		rr := httptest.NewRecorder()

		req := httptest.NewRequest(http.MethodPost, urlRevocation, nil)
		req.Header.Set(CertificateHeader, testCert)

		//when
		handler.Revoke(rr, req)

		//then
		assert.Equal(t, http.StatusCreated, rr.Code)
		revocationListRepository.AssertExpectations(t)
	})

	t.Run("should return http code 201 when certificate already revoked", func(t *testing.T) {
		//given
		revocationListRepository := &mocks.RevocationListRepository{}
		revocationListRepository.On("Insert", hashedTestCert).Return(nil)

		handler := NewRevocationHandler(revocationListRepository)

		rr := httptest.NewRecorder()

		req := httptest.NewRequest(http.MethodPost, urlRevocation, nil)
		req.Header.Set(CertificateHeader, testCert)

		//when
		handler.Revoke(rr, req)

		//then
		assert.Equal(t, http.StatusCreated, rr.Code)
		revocationListRepository.AssertExpectations(t)

		//when
		handler.Revoke(rr, req)

		//then
		assert.Equal(t, http.StatusCreated, rr.Code)
		revocationListRepository.AssertExpectations(t)
	})

	t.Run("should return http code 500 when certificate not passed", func(t *testing.T) {
		//given
		testCert := ""

		revocationListRepository := &mocks.RevocationListRepository{}

		handler := NewRevocationHandler(revocationListRepository)

		rr := httptest.NewRecorder()

		req := httptest.NewRequest(http.MethodPost, urlRevocation, nil)
		req.Header.Set(CertificateHeader, testCert)

		//when
		handler.Revoke(rr, req)

		//then
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		revocationListRepository.AssertNotCalled(t, "Insert", mock.AnythingOfType("string"))
	})

	t.Run("should return http code 500 when certificate revocation not persisted", func(t *testing.T) {
		//given
		revocationListRepository := &mocks.RevocationListRepository{}
		revocationListRepository.On("Insert", hashedTestCert).Return(errors.New("Error"))

		handler := NewRevocationHandler(revocationListRepository)

		rr := httptest.NewRecorder()

		req := httptest.NewRequest(http.MethodPost, urlRevocation, nil)
		req.Header.Set(CertificateHeader, testCert)

		//when
		handler.Revoke(rr, req)

		//then
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		revocationListRepository.AssertExpectations(t)
	})

	t.Run("should return http code 500 when error occurred during hash calculation", func(t *testing.T) {
		//given
		testCertIncorrect := "-----BEGIN%20CERTIFICATE-----%0" +
			"MIIFfjCCA2agAwIBAgIBAjANBgkqhkiG9w0BAQsFADBqMQswCQYDVQQGEwJQTDEK%0" +
			"MAgGA1UECAwBTjEQMA4GA1UEBwwHR0xJV0lDRTETMBEGA1UECgwKU0FQIEh5YnJp%0" +
			"czENMAsGA1UECwwES3ltYTEZMBcGA1UEAwwQd29ybWhvbGUua3ltYS5jeDAeFw0x%0" +
			"OTAxMjQwOTQ1MTJaFw0yMDAxMjQwOTQ1MTJaMHIxCzAJBgNVBAYTAkRFMRAwDgYD%0" +
			"VQQIEwdXYWxkb3JmMRAwDgYDVQQHEwdXYWxkb3JmMRUwEwYDVQQKEwxPcmdhbml6%0" +
			"YXRpb24xEDAOBgNVBAsTB09yZ1VuaXQxFjAUBgNVBAMTDXVwcy10ZXN0LWFwcDEw%0" +
			"ggIiMA0GCSqGSIb3DQEBAQUAA4ICDwAwggIKAoICAQC7eOxYnQy3CcNKYAYb2b48%0" +
			"FA92GW1LWjfCBEhu+uP5hLlH572OODpdhyNmtFInpl8gmQ1454x%2FrG5bOjUxOkOi%0" +
			"U97xM8e7O1MW4fLIC2LDkY0TapfaMn%2FoYwUHz3QFlx+BKM4yxgpJ+YY5OJVoGDEH%0" +
			"TdaSY5LzftnfSBjUMrQIec6wv3+Vy0F8jUmxV3FifODhWxBUBEqZOSm9ruOwOSan%0" +
			"J2sY1ovsj4QZZ9HnLJT6mdMt8lAwyDg0o68aOsjPtDEaqOIGWiYls49GIRi%2FcXft%0" +
			"1S2L1q5WOiFlK18rkWouFJrx7YppPXEpLzwqY8DjaZ5qbfaBZzTaYHYBgTqClCbp%0" +
			"AqwzDBsXuhPKywY+HzGhkGMHJKVxADuycgGObyQsnFDg0284zVJpCU41grVaigjf%0" +
			"PsD8NU%2FhO11M+45xVl7WCLecHmVKZpwR3A6g6WpucitFKdHvbdD6jhbt3aUuJN+u%0" +
			"U7GdnPk0A+y%2Fs7Y6FzxdoVtfWFIZqnFHIcL1kUT8WeEONnzmmMvOIIh8rz1U9iVO%0" +
			"m%2FFZ1u3JloRRdGpaKewTytRsOKBE5Oug24GZSKA5dGG6ex7BDoXRgUgf3WxuUBQ2%0" +
			"uidkXunTrcWUFI3IwbAdS9DWnlVhjvYBjOw8OsF3WtUsyIkDKFs11mfQvOKg6Y0m%0" +
			"FrF9eKkyolUlmrHrGdG3wwIDAQABoycwJTAOBgNVHQ8BAf8EBAMCB4AwEwYDVR0l%0" +
			"BAwwCgYIKwYBBQUHAwIwDQYJKoZIhvcNAQELBQADggIBAGpTE%2FtPB8ZJqR+mqZDB%0" +
			"HzTE1fyoYYJvXfGwZQEWrR3GqOPe0Ni3jWQKs9NYoYSGoN9cYFvDJUQy3d5BJ1m8%0" +
			"Y0uySIEDewYpKZR1zL7lnUYZ4msjOpNnlcIGNd0xz1EbHDPK3fMljVu+oJ4d+Z7a%0" +
			"VdfJ+4fXUPgWxA9ou2E9GoiWtoReGc7ok2sx1NH3%2FZIaSy7zHKfve8ULrwkvVTEs%0" +
			"W+jzuGIZvIQXMyY8c2M3sNb4sn14xRTZHA88rDJhI8Yjmhp9ZvaqAv+isaeBcGhY%0" +
			"Cpspwgl1qqGqNsgEkoScW6WzN8HP0DvQ6UhsJKKhtTxl2Kd81PKGQIz%2F9vdkhOew%0" +
			"irbwUBl1Wx1eoJlcTXFEMSGVVjHUUDDyftfq+hfJ4U90Ny0tjPjr+t9ejHOWfszf%0" +
			"3Na0qT+92au76SFBddBVe3J0jaGhcwyFMHmQp8CU9ZKHXnSeumwplo33+25U3M%2Fe%0" +
			"ALnDeejOZfFu0zDXTPZsvUFDJBke6PsILtQOAyrGATbDzPCqCZPmIl%2Fp4aMwgPJW%0" +
			"zn6TyG%2F0YbJXH7Mysm+8k9qSWTpfT8YRDSrOUIXPFxg9jmMqF8vnksNq8cwto57Z%0" +
			"gPzl0TAYNKqS%2FYDkeNOgrwlnJ%2Ff2rLTD0Zki16i%2FJ3s6%2FIy37XXj2No+G8bvGBDZ%0" +
			"ffu3gnzaWei6HTmquQp55Kun%0" +
			"-----END%20CERTIFICATE-----"

		revocationListRepository := &mocks.RevocationListRepository{}

		handler := NewRevocationHandler(revocationListRepository)

		rr := httptest.NewRecorder()

		req := httptest.NewRequest(http.MethodPost, urlRevocation, nil)
		req.Header.Set(CertificateHeader, testCertIncorrect)

		//when
		handler.Revoke(rr, req)

		//then
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}
