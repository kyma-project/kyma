package internal

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerateCA(t *testing.T) {
	ca, err := GenerateCACert()
	require.NoError(t, err)
	require.NotNil(t, ca)
}

func TestGenerateServerCert(t *testing.T) {
	ca, _ := GenerateCACert()

	cert, err := GenerateServerCertAndKey(ca, "some-service", "some-namespace")
	require.NoError(t, err)
	require.NotNil(t, cert)
}
