package certificates

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCalculateHash(t *testing.T) {

	t.Run("should calculate correct fingerprint", func(t *testing.T) {
		// given
		rawPemCertificate := []byte(`-----BEGIN CERTIFICATE-----
MIIFfjCCA2agAwIBAgIBAjANBgkqhkiG9w0BAQsFADBqMQswCQYDVQQGEwJQTDEK
MAgGA1UECAwBTjEQMA4GA1UEBwwHR0xJV0lDRTETMBEGA1UECgwKU0FQIEh5YnJp
czENMAsGA1UECwwES3ltYTEZMBcGA1UEAwwQd29ybWhvbGUua3ltYS5jeDAeFw0x
OTAxMjQwOTQ1MTJaFw0yMDAxMjQwOTQ1MTJaMHIxCzAJBgNVBAYTAkRFMRAwDgYD
VQQIEwdXYWxkb3JmMRAwDgYDVQQHEwdXYWxkb3JmMRUwEwYDVQQKEwxPcmdhbml6
YXRpb24xEDAOBgNVBAsTB09yZ1VuaXQxFjAUBgNVBAMTDXVwcy10ZXN0LWFwcDEw
ggIiMA0GCSqGSIb3DQEBAQUAA4ICDwAwggIKAoICAQC7eOxYnQy3CcNKYAYb2b48
FA92GW1LWjfCBEhu+uP5hLlH572OODpdhyNmtFInpl8gmQ1454x/rG5bOjUxOkOi
U97xM8e7O1MW4fLIC2LDkY0TapfaMn/oYwUHz3QFlx+BKM4yxgpJ+YY5OJVoGDEH
TdaSY5LzftnfSBjUMrQIec6wv3+Vy0F8jUmxV3FifODhWxBUBEqZOSm9ruOwOSan
J2sY1ovsj4QZZ9HnLJT6mdMt8lAwyDg0o68aOsjPtDEaqOIGWiYls49GIRi/cXft
1S2L1q5WOiFlK18rkWouFJrx7YppPXEpLzwqY8DjaZ5qbfaBZzTaYHYBgTqClCbp
AqwzDBsXuhPKywY+HzGhkGMHJKVxADuycgGObyQsnFDg0284zVJpCU41grVaigjf
PsD8NU/hO11M+45xVl7WCLecHmVKZpwR3A6g6WpucitFKdHvbdD6jhbt3aUuJN+u
U7GdnPk0A+y/s7Y6FzxdoVtfWFIZqnFHIcL1kUT8WeEONnzmmMvOIIh8rz1U9iVO
m/FZ1u3JloRRdGpaKewTytRsOKBE5Oug24GZSKA5dGG6ex7BDoXRgUgf3WxuUBQ2
uidkXunTrcWUFI3IwbAdS9DWnlVhjvYBjOw8OsF3WtUsyIkDKFs11mfQvOKg6Y0m
FrF9eKkyolUlmrHrGdG3wwIDAQABoycwJTAOBgNVHQ8BAf8EBAMCB4AwEwYDVR0l
BAwwCgYIKwYBBQUHAwIwDQYJKoZIhvcNAQELBQADggIBAGpTE/tPB8ZJqR+mqZDB
HzTE1fyoYYJvXfGwZQEWrR3GqOPe0Ni3jWQKs9NYoYSGoN9cYFvDJUQy3d5BJ1m8
Y0uySIEDewYpKZR1zL7lnUYZ4msjOpNnlcIGNd0xz1EbHDPK3fMljVu+oJ4d+Z7a
VdfJ+4fXUPgWxA9ou2E9GoiWtoReGc7ok2sx1NH3/ZIaSy7zHKfve8ULrwkvVTEs
W+jzuGIZvIQXMyY8c2M3sNb4sn14xRTZHA88rDJhI8Yjmhp9ZvaqAv+isaeBcGhY
Cpspwgl1qqGqNsgEkoScW6WzN8HP0DvQ6UhsJKKhtTxl2Kd81PKGQIz/9vdkhOew
irbwUBl1Wx1eoJlcTXFEMSGVVjHUUDDyftfq+hfJ4U90Ny0tjPjr+t9ejHOWfszf
3Na0qT+92au76SFBddBVe3J0jaGhcwyFMHmQp8CU9ZKHXnSeumwplo33+25U3M/e
ALnDeejOZfFu0zDXTPZsvUFDJBke6PsILtQOAyrGATbDzPCqCZPmIl/p4aMwgPJW
zn6TyG/0YbJXH7Mysm+8k9qSWTpfT8YRDSrOUIXPFxg9jmMqF8vnksNq8cwto57Z
gPzl0TAYNKqS/YDkeNOgrwlnJ/f2rLTD0Zki16i/J3s6/Iy37XXj2No+G8bvGBDZ
ffu3gnzaWei6HTmquQp55Kun
-----END CERTIFICATE-----`)
		expectedFingerprint := "daf8a9b0927fc221bdc3319336e4236ab59a7c0e3986d4cf69898601cb8ee604"

		// when
		calculatedHash, err := FingerprintSHA256(rawPemCertificate)
		require.NoError(t, err)

		// then
		assert.Equal(t, expectedFingerprint, calculatedHash)
	})

	t.Run("should return error, when unable to decode cert", func(t *testing.T) {
		// given
		invalidPemCertificate := []byte("invalid-input")

		// when
		hash, err := FingerprintSHA256(invalidPemCertificate)

		// then
		assert.Equal(t, "", hash)
		assert.Error(t, err)
	})
}
