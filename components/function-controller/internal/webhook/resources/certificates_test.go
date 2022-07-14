package resources

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"reflect"
	"testing"
	"time"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	testSecretName    = "test-secret"
	testNamespaceName = "test-namespace"
	testServiceName   = "test-service"
)

func Test_serviceAltNames(t *testing.T) {
	type args struct {
		serviceName string
		namespace   string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "service AltNames are generated correctly",
			args: args{serviceName: "test-service", namespace: "test-namespace"},
			// not using consts here to make it as readable as possible.
			want: []string{
				"test-service.test-namespace.svc",
				"test-service",
				"test-service.test-namespace",
				"test-service.test-namespace.svc.cluster.local",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := serviceAltNames(tt.args.serviceName, tt.args.namespace)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("invalid serviec altNames: serviceAltNames() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEnsureWebhookSecret(t *testing.T) {
	ctx := context.Background()
	cert, key, err := generateWebhookCertificates(testServiceName, testNamespaceName)
	require.NoError(t, err)

	t.Run("can ensure the secret if it doesn't exist", func(t *testing.T) {
		client := fake.NewClientBuilder().Build()

		err := EnsureWebhookSecret(ctx, client, testSecretName, testNamespaceName, testServiceName)
		require.NoError(t, err)

		secret := &corev1.Secret{}
		err = client.Get(ctx, types.NamespacedName{Name: testSecretName, Namespace: testNamespaceName}, secret)

		require.NoError(t, err)
		require.NotNil(t, secret)
		require.Equal(t, testSecretName, secret.Name)
		require.Equal(t, testNamespaceName, secret.Namespace)
		require.Contains(t, secret.Data, KeyFile)
		require.Contains(t, secret.Data, CertFile)
	})

	t.Run("can ensure the secret is updated if it exists", func(t *testing.T) {
		client := fake.NewClientBuilder().Build()
		secret := &corev1.Secret{
			ObjectMeta: v1.ObjectMeta{
				Name:      testSecretName,
				Namespace: testNamespaceName,
				Labels: map[string]string{
					"dont-remove-me": "true",
				},
			},
		}
		err := client.Create(ctx, secret)
		require.NoError(t, err)

		err = EnsureWebhookSecret(ctx, client, testSecretName, testNamespaceName, testServiceName)
		require.NoError(t, err)

		updatedSecret := &corev1.Secret{}
		err = client.Get(ctx, types.NamespacedName{Name: testSecretName, Namespace: testNamespaceName}, updatedSecret)

		require.NoError(t, err)
		require.NotNil(t, secret)
		require.Equal(t, testSecretName, updatedSecret.Name)
		require.Equal(t, testNamespaceName, updatedSecret.Namespace)
		require.Contains(t, updatedSecret.Data, KeyFile)
		require.Contains(t, updatedSecret.Data, CertFile)
		require.Contains(t, updatedSecret.Labels, "dont-remove-me")
	})

	t.Run("can ensure the secret is updated if it's missing a value", func(t *testing.T) {
		client := fake.NewClientBuilder().Build()
		secret := &corev1.Secret{
			ObjectMeta: v1.ObjectMeta{
				Name:      testSecretName,
				Namespace: testNamespaceName,
				Labels: map[string]string{
					"dont-remove-me": "true",
				},
			},
			Data: map[string][]byte{
				KeyFile: key,
			},
		}
		err := client.Create(ctx, secret)
		require.NoError(t, err)

		err = EnsureWebhookSecret(ctx, client, testSecretName, testNamespaceName, testServiceName)
		require.NoError(t, err)

		updatedSecret := &corev1.Secret{}
		err = client.Get(ctx, types.NamespacedName{Name: testSecretName, Namespace: testNamespaceName}, updatedSecret)

		require.NoError(t, err)
		require.NotNil(t, secret)
		require.Equal(t, testSecretName, updatedSecret.Name)
		require.Equal(t, testNamespaceName, updatedSecret.Namespace)
		// make sure the test is updated
		require.NotEqual(t, secret.ResourceVersion, updatedSecret.ResourceVersion)
		require.Contains(t, updatedSecret.Data, KeyFile)
		require.Contains(t, updatedSecret.Data, CertFile)
		require.Contains(t, updatedSecret.Labels, "dont-remove-me")
	})

	t.Run("doesn't update the secret if it's ok", func(t *testing.T) {
		client := fake.NewClientBuilder().Build()
		secret := &corev1.Secret{
			ObjectMeta: v1.ObjectMeta{
				Name:      testSecretName,
				Namespace: testNamespaceName,
				Labels: map[string]string{
					"dont-remove-me": "true",
				},
			},
			Data: map[string][]byte{
				KeyFile:  key,
				CertFile: cert,
			},
		}
		err := client.Create(ctx, secret)
		require.NoError(t, err)

		err = EnsureWebhookSecret(ctx, client, testSecretName, testNamespaceName, testServiceName)
		require.NoError(t, err)

		updatedSecret := &corev1.Secret{}
		err = client.Get(ctx, types.NamespacedName{Name: testSecretName, Namespace: testNamespaceName}, updatedSecret)

		require.NoError(t, err)
		require.NotNil(t, secret)
		require.Equal(t, testSecretName, updatedSecret.Name)
		require.Equal(t, testNamespaceName, updatedSecret.Namespace)
		// make sure it's not updated
		require.Equal(t, secret.ResourceVersion, updatedSecret.ResourceVersion)
		require.Contains(t, updatedSecret.Data, KeyFile)
		require.Contains(t, updatedSecret.Data, CertFile)
		require.Equal(t, key, updatedSecret.Data[KeyFile])
		require.Equal(t, cert, updatedSecret.Data[CertFile])
		require.Contains(t, updatedSecret.Labels, "dont-remove-me")
	})

	t.Run("should update if the cert will expire in 10 days", func(t *testing.T) {
		client := fake.NewClientBuilder().Build()

		tenDaysCert, err := generateShortLivedCertWithKey(key, testServiceName, 10*24*time.Hour)
		require.NoError(t, err)

		secret := &corev1.Secret{
			ObjectMeta: v1.ObjectMeta{
				Name:      testSecretName,
				Namespace: testNamespaceName,
				Labels: map[string]string{
					"dont-remove-me": "true",
				},
			},
			Data: map[string][]byte{
				KeyFile:  key,
				CertFile: tenDaysCert,
			},
		}
		err = client.Create(ctx, secret)
		require.NoError(t, err)

		err = EnsureWebhookSecret(ctx, client, testSecretName, testNamespaceName, testServiceName)
		require.NoError(t, err)

		updatedSecret := &corev1.Secret{}
		err = client.Get(ctx, types.NamespacedName{Name: testSecretName, Namespace: testNamespaceName}, updatedSecret)

		require.NoError(t, err)
		require.NotNil(t, secret)
		require.Equal(t, testSecretName, updatedSecret.Name)
		require.Equal(t, testNamespaceName, updatedSecret.Namespace)
		require.Contains(t, updatedSecret.Data, KeyFile)
		require.Contains(t, updatedSecret.Data, CertFile)
		// make sure it's updated, not overridden.
		require.NotEqual(t, secret.ResourceVersion, updatedSecret.ResourceVersion)
		require.NotEqual(t, key, updatedSecret.Data[KeyFile])
		require.NotEqual(t, cert, updatedSecret.Data[CertFile])
		require.Contains(t, updatedSecret.Labels, "dont-remove-me")
	})

	t.Run("should not update if the cert will expire in more than 10 days", func(t *testing.T) {
		client := fake.NewClientBuilder().Build()

		elevenDaysCert, err := generateShortLivedCertWithKey(key, testServiceName, 11*24*time.Hour)
		require.NoError(t, err)

		secret := &corev1.Secret{
			ObjectMeta: v1.ObjectMeta{
				Name:      testSecretName,
				Namespace: testNamespaceName,
				Labels: map[string]string{
					"dont-remove-me": "true",
				},
			},
			Data: map[string][]byte{
				KeyFile:  key,
				CertFile: elevenDaysCert,
			},
		}
		err = client.Create(ctx, secret)
		require.NoError(t, err)

		err = EnsureWebhookSecret(ctx, client, testSecretName, testNamespaceName, testServiceName)
		require.NoError(t, err)

		updatedSecret := &corev1.Secret{}
		err = client.Get(ctx, types.NamespacedName{Name: testSecretName, Namespace: testNamespaceName}, updatedSecret)

		require.NoError(t, err)
		require.NotNil(t, secret)
		require.Equal(t, testSecretName, updatedSecret.Name)
		require.Equal(t, testNamespaceName, updatedSecret.Namespace)
		require.Contains(t, updatedSecret.Data, KeyFile)
		require.Contains(t, updatedSecret.Data, CertFile)
		// make sure it's NOT updated, not overridden.
		require.Equal(t, secret.ResourceVersion, updatedSecret.ResourceVersion)
		require.Equal(t, key, updatedSecret.Data[KeyFile])
		require.Equal(t, elevenDaysCert, updatedSecret.Data[CertFile])
		require.Contains(t, updatedSecret.Labels, "dont-remove-me")
	})
}

func TestUpdateCRD(t *testing.T) {

	testCaBundle := []byte("test-ca-bundle")
	testCases := map[string]struct {
		input     *apiextensionsv1.CustomResourceDefinition
		returnErr bool
	}{
		"Correct crd": {
			input: &apiextensionsv1.CustomResourceDefinition{
				ObjectMeta: v1.ObjectMeta{
					Name: FunctionCRDName,
				},
				Spec: apiextensionsv1.CustomResourceDefinitionSpec{
					Conversion: &apiextensionsv1.CustomResourceConversion{
						Webhook: &apiextensionsv1.WebhookConversion{ClientConfig: &apiextensionsv1.WebhookClientConfig{}},
					},
				},
			},
			returnErr: false,
		},
		"CRD without conversion part": {
			input: &apiextensionsv1.CustomResourceDefinition{
				ObjectMeta: v1.ObjectMeta{
					Name: FunctionCRDName,
				},
				Spec: apiextensionsv1.CustomResourceDefinitionSpec{},
			},
			returnErr: true,
		},
		"CRD without conversion webhook part": {
			input: &apiextensionsv1.CustomResourceDefinition{
				ObjectMeta: v1.ObjectMeta{
					Name: FunctionCRDName,
				},
				Spec: apiextensionsv1.CustomResourceDefinitionSpec{
					Conversion: &apiextensionsv1.CustomResourceConversion{},
				},
			},
			returnErr: true,
		},
		"CRD without conversion webhook client config part": {
			input: &apiextensionsv1.CustomResourceDefinition{
				ObjectMeta: v1.ObjectMeta{
					Name: FunctionCRDName,
				},
				Spec: apiextensionsv1.CustomResourceDefinitionSpec{
					Conversion: &apiextensionsv1.CustomResourceConversion{
						Webhook: &apiextensionsv1.WebhookConversion{},
					},
				},
			},
			returnErr: true,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			ctx := context.TODO()
			client := fake.NewClientBuilder().Build()
			require.NoError(t, apiextensionsv1.AddToScheme(client.Scheme()))
			require.NoError(t, client.Create(ctx, testCase.input))

			//WHEN
			err := AddCertToConversionWebhook(context.TODO(), client, testCaBundle)

			//THEN
			if testCase.returnErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				updatedCRD := &apiextensionsv1.CustomResourceDefinition{}
				require.NoError(t, client.Get(ctx, types.NamespacedName{Name: FunctionCRDName}, updatedCRD))
				require.Equal(t, testCaBundle, updatedCRD.Spec.Conversion.Webhook.ClientConfig.CABundle)
			}
		})
	}
}

func generateShortLivedCertWithKey(keyBytes []byte, host string, age time.Duration) ([]byte, error) {
	pemKey, _ := pem.Decode(keyBytes)
	key, err := x509.ParsePKCS1PrivateKey(pemKey.Bytes)
	if err != nil {
		return nil, err
	}
	t := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: host,
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(age),
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, &t, &t, &key.PublicKey, key)
	if err != nil {
		return nil, err
	}
	buf := bytes.Buffer{}
	if err := pem.Encode(&buf, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes}); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
