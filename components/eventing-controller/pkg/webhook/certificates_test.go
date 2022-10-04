package webhook

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/xerrors"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func TestWebhookHandler(t *testing.T) {
	ctx := context.Background()
	logger := zap.New()
	_ = apiextensionsv1.AddToScheme(scheme.Scheme)
	fakeClient := fake.NewClientBuilder().WithScheme(scheme.Scheme).Build()

	url := "test-webhook-url"
	path := "test-webhook-path"
	var port int32 = 8080
	secretName := "test-secret"
	testCases := []struct {
		description string
		crd         *apiextensionsv1.CustomResourceDefinition
		secretName  string
		wantErr     bool
	}{
		{
			"should create secret and inject caBundle successfully",
			&apiextensionsv1.CustomResourceDefinition{
				ObjectMeta: v1.ObjectMeta{
					Name: "test.kyma.com",
				},
				Spec: apiextensionsv1.CustomResourceDefinitionSpec{
					Conversion: &apiextensionsv1.CustomResourceConversion{
						Webhook: &apiextensionsv1.WebhookConversion{
							ClientConfig: &apiextensionsv1.WebhookClientConfig{
								URL: &url,
								Service: &apiextensionsv1.ServiceReference{
									Namespace: "test-crd-namespace",
									Name:      "test-crd-name",
									Path:      &path,
									Port:      &port,
								},
							},
						},
					},
				},
			},
			secretName,
			false,
		},
		{
			"should fail to get crd",
			nil,
			secretName,
			true,
		},
		{
			"should not be able to get crd",
			&apiextensionsv1.CustomResourceDefinition{
				ObjectMeta: v1.ObjectMeta{
					Name: "test.kyma.com",
				},
				Spec: apiextensionsv1.CustomResourceDefinitionSpec{},
			},
			secretName,
			true,
		},
	}

	for _, tc := range testCases {
		crdName := "test-name"
		if tc.crd != nil {
			crdName = tc.crd.Name
			require.NoError(t, fakeClient.Create(ctx, tc.crd))
		}

		webhookHandler := NewCertificates(ctx, fakeClient, &logger, crdName, secretName)
		err := webhookHandler.Setup()
		if (err != nil) != tc.wantErr {
			t.Errorf("%s: got err = %v; want (err != nil) = %v", tc.description, err, tc.wantErr)
		}

		// success case
		if err == nil {
			// verify if secret is created
			secret := &corev1.Secret{}
			err = fakeClient.Get(ctx, types.NamespacedName{Name: secretName, Namespace: tc.crd.Spec.Conversion.Webhook.ClientConfig.Service.Namespace}, secret)
			if err != nil {
				t.Errorf("%s: there must be the secret [%s] created", tc.description, secretName)
			}
			// verify if bundleCA is injected into CRD
			crdFromEtcd := &apiextensionsv1.CustomResourceDefinition{}
			err = fakeClient.Get(ctx, types.NamespacedName{Name: crdName}, crdFromEtcd)
			require.NoError(t, err)
			if crdFromEtcd.Spec.Conversion.Webhook.ClientConfig.CABundle == nil {
				t.Errorf("%s: CRD [%s] must have the caBundle value", tc.description, crdName)
			}
		}

		if tc.crd != nil {
			require.NoError(t, fakeClient.Delete(ctx, tc.crd))
		}
	}

}

func TestCertificateHandlerServiceAltNames(t *testing.T) {
	ctx := context.Background()
	logger := zap.New()
	certHandler := CertificateHandler{
		ctx:    ctx,
		logger: &logger,
	}

	testCases := []struct {
		serviceName  string
		namespace    string
		wantAltNames []string
	}{
		{
			serviceName:  "service-name1",
			namespace:    "namespace1",
			wantAltNames: []string{"service-name1", "service-name1.namespace1", "service-name1.namespace1.svc", "service-name1.namespace1.svc.cluster.local"},
		},
		{
			serviceName:  "service-name2",
			namespace:    "namespace2",
			wantAltNames: []string{"service-name2", "service-name2.namespace2", "service-name2.namespace2.svc", "service-name2.namespace2.svc.cluster.local"},
		},
	}

	for _, tc := range testCases {
		if gotAltNames := certHandler.serviceAltNames(tc.serviceName, tc.namespace); !reflect.DeepEqual(gotAltNames, tc.wantAltNames) {
			t.Errorf("get service alt names failed, want:[%v] but got:[%v]", tc.wantAltNames, gotAltNames)
		}
	}
}

func TestCertHandlerBuildCert(t *testing.T) {
	ctx := context.Background()
	logger := zap.New()
	certHandler := CertificateHandler{
		clientGoCert: &ClientGoCert{},
		ctx:          ctx,
		logger:       &logger,
	}
	mockCertHandler := CertificateHandler{
		clientGoCert: &MockClientGoCert{},
		ctx:          ctx,
		logger:       &logger,
	}

	testCases := []struct {
		description string
		namespace   string
		serviceName string
		certHandler ICertificateHandler
		wantErr     error
	}{
		{
			description: "should create certificate successfully",
			namespace:   "test-namespace1",
			serviceName: "test-service-name1",
			certHandler: &certHandler,
		},
		{
			description: "should fail to create certificate",
			certHandler: &mockCertHandler,
			wantErr:     xerrors.New("failed to generate certificates: fake self signed certificate generation failed"),
		},
	}

	for _, tc := range testCases {
		crt, key, err := tc.certHandler.buildCert(tc.namespace, tc.serviceName)
		if tc.wantErr != nil {
			if err.Error() != tc.wantErr.Error() {
				t.Errorf("%s: BuildCert() failed, want:[%v] but got:[%v]", tc.description, tc.wantErr, err)
			}
		} else {
			if err = certHandler.verifyCertificate(crt); err != nil {
				t.Errorf("%s: cert must have been valid, but got invalid certificate", tc.description)
			}
			if err = certHandler.verifyKey(key); err != nil {
				t.Errorf("%s: key must have been valid, but got invalid key", tc.description)
			}
		}
	}
}

func TestSecretHandlerEnsureSecret(t *testing.T) {
	ctx := context.Background()
	logger := zap.New()
	fakeClient := fake.NewClientBuilder().WithScheme(scheme.Scheme).Build()
	secretHandler := SecretHandler{
		certHandler: &CertificateHandler{
			clientGoCert: &ClientGoCert{},
			ctx:          ctx,
			logger:       &logger,
		},
		Client: fakeClient,
		ctx:    ctx,
		logger: &logger,
	}
	sec, _ := secretHandler.createSecret("test-secret1", "test-namespace1", "test-service-name1")

	testCases := []struct {
		description string
		secretName  string
		namespace   string
		serviceName string
		client      client.Client
		wantSecret  *corev1.Secret
		wantErr     error
	}{
		{
			description: "should create a new secret",
			secretName:  "test-secret1",
			namespace:   "test-namespace1",
			serviceName: "test-service-name1",
			client:      fakeClient,
		},
		{
			description: "should update the new secret",
			secretName:  "test-secret1",
			namespace:   "test-namespace1",
			serviceName: "test-service-name1",
			client:      fakeClient,
			wantSecret:  sec,
		},
		{
			description: "should fail to get secret",
			secretName:  "test-secret2",
			namespace:   "test-namespace2",
			serviceName: "test-service-name2",
			client:      &MockGetFailedClient{},
			wantErr:     xerrors.Errorf("failed to get webhook secret: %w", ErrFakeGet),
		},
		{
			description: "should fail to create secret",
			secretName:  "test-secret3",
			namespace:   "test-namespace3",
			serviceName: "test-service-name3",
			client:      &MockCreateFailedClient{},
			wantErr:     xerrors.Errorf("failed to create secret: %w", ErrFakeCreate),
		},
		{
			description: "should fail to update secret",
			secretName:  "test-secret4",
			namespace:   "test-namespace4",
			serviceName: "test-service-name4",
			client:      &MockUpdateFailedClient{},
			wantErr:     xerrors.Errorf("failed to update secret: %w", ErrFakeUpdate),
		},
	}

	for _, tc := range testCases {
		secretHandler = SecretHandler{
			certHandler: &CertificateHandler{
				clientGoCert: &ClientGoCert{},
				ctx:          ctx,
				logger:       &logger,
			},
			Client: tc.client,
			ctx:    ctx,
			logger: &logger,
		}

		_, err := secretHandler.ensureSecret(tc.secretName, tc.namespace, tc.serviceName)

		if tc.wantErr != nil {
			require.EqualError(t, err, tc.wantErr.Error())
		}

		if tc.wantErr == nil {
			require.NoError(t, err)
			secret := &corev1.Secret{}
			err = tc.client.Get(ctx, types.NamespacedName{Name: tc.secretName, Namespace: tc.namespace}, secret)
			require.NoError(t, err)
			if err != nil && apierrors.IsNotFound(err) {
				t.Errorf("%s: there must a secret created", tc.description)
			}

			// after update, secret must be different
			if tc.wantSecret != nil {
				require.NotEqual(t, tc.wantSecret, secret)
			}
		}
	}
}
