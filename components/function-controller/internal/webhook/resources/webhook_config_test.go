package resources

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var (
	caBundel = []byte("LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUVEakNDQW5ZQ0NRRE5vZFh6bERqdk9qQU5CZ2txaGtpRzl3MEJBUXNGQURCSk1SVXdFd1lEVlFRS0RBeGoKWlhKMExXMWhibUZuWlhJeEZUQVRCZ05WQkFzTURHTmxjblF0YldGdVlXZGxjakVaTUJjR0ExVUVBd3dRYTNsdApZUzFsZUdGdGNHeGxMbU52YlRBZUZ3MHlNVEE1TWpreE5ERXhOVGRhRncweU1qQTVNamt4TkRFeE5UZGFNRWt4CkZUQVRCZ05WQkFvTURHTmxjblF0YldGdVlXZGxjakVWTUJNR0ExVUVDd3dNWTJWeWRDMXRZVzVoWjJWeU1Sa3cKRndZRFZRUUREQkJyZVcxaExXVjRZVzF3YkdVdVkyOXRNSUlCb2pBTkJna3Foa2lHOXcwQkFRRUZBQU9DQVk4QQpNSUlCaWdLQ0FZRUFvMHNzdEtYZ01sR0FvV0F6WmxCZVNHMTVJZ0VBRkZwbWFTcnArQ0Y1bGdjTXVSZFNBamlXCk1tS0F5OWlzdy9meTN3YmFWUWNoSWFLZXJvZzVFcnYyR1hlMmZaK0Q2anJpbFp4SG01ZURxZmtxdmQrcXh6K3IKN25Dd1JOS2t0Q093YmJqOWtMSWduQkdUVGZQL1NiVUMxd2R5R3BuYnpwVXo1b3RvbVJRMWlsWi9OZFl2T2UwYworK21zZ2xSUVpNWmNtZVI0RWJhZlNwRTdWT1M2L0srbTZpMzRZMzdmSTBhSUlUV2lVR1RuMzRVMEZhTG9DSTF5CjVNaFhSYUhyQTAyWVBsNmhUMlJFVVlRUTBuNUMxVnpQYm02VHkvUjdoOXI2RnoxQWxPQ0I5NnZtalNPemR2dmsKaHdMcVRqdHArbFNNcW1iRExEV2ZVY2JpVFIwbzliQ3Rxam1aUjQwcXhqa1hieW5hYU42TzFnbm9Fem50a1J3Sgo1cytZaVVtTHpvVFIyV2V1bHc5VVZaZXdiQm85Zk1FSFVRa1RJOEd0aWVjUDhNMGF4R1RSSVFCWUc0bTBYcEF1CmFrd2ZPNlRlZzR2T1FTQmpVanE0Rm1nYkZkVVZsUkZrVm95NXkra3JrV2FpT21hSFlqWHViMkduVWQ1WFBiamgKeG56a1R4UlE3UFh6QWdNQkFBRXdEUVlKS29aSWh2Y05BUUVMQlFBRGdnR0JBRCs5QlExWERXY0pkZ2Jma2ZJeApNa1ZLc3pXRWdvUC9GWWNmcVVEc1drcTNkR3M0Q1Z4ODRzTjNRd0hSY2JkV2trVTd1WXh3VmE0NVZWbVZrZjhmClZGN3ZiblhWWUoydHp5K2lTb0JDcVFjMUNHV1ZwQmdIOGlFVWdYb2hwdTczWjhtSkJhNUhQT1lXTHM5dEVlVTUKdy9VOG9HOUJZRHRsclBGSzZLWkJpYU8veERMQXlFOFRxUEc3U3oxUXJFdC9HbE1wS1RHakFUYXNNQ2hzT0IrMgpuc2xLdExuNXZoS3hOdkRIYjJ4Y1pGZHlOdGQ5Vk9mcFQvNXZ2VE4wb3ozcERJd2RhUkNyTTMwU0tCZVl1OFdGClJPb1NsNUE5dDl3S08vZ0xwcFJ1WW9IdzU4Z1VudWpCOUoyT1oxaUp2YUFLTXZMNGFFbTE4VjVCYkZaUVdPUFkKcWRYZGRnYk80R0VsUFRXdUxFb01HZTM5UnBITmVtMmcrNFdUcjl1K2FkSVN3MnBQbjFRYXhqbzJmb0ZiUUpUcgpXeXR5MnVLU2diU3RzNzEvOUZ4bUIvM3Z6Y21oOUhLdzJWb0FQSzY4ckdhUkhvTjdqZ0dCR3JtdDdHZ01UeGlqCmRuOEt1OHBhRERWSTFldUx6RlJwUlBvdVVwOEVDdmRHTkxoRTJrc0dsdEk3WEE9PQotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg==")
)

func TestEnsureWebhookConfigurationFor(t *testing.T) {
	ctx := context.Background()

	t.Run("ensure a Mutating Webhook is created if it doesn't exist", func(t *testing.T) {
		client := fake.NewClientBuilder().Build()
		wc := WebhookConfig{
			CABundel:         caBundel,
			ServiceName:      testServiceName,
			ServiceNamespace: testNamespaceName,
		}
		err := EnsureWebhookConfigurationFor(ctx, client, wc, MutatingWebhook)
		require.NoError(t, err)

		mwh := &admissionregistrationv1.MutatingWebhookConfiguration{}
		err = client.Get(ctx, types.NamespacedName{Name: DefaultingWebhookName}, mwh)

		require.NoError(t, err)
		require.NotNil(t, mwh)
		require.Equal(t, mwh.Name, DefaultingWebhookName)
		validateMutatingWebhook(t, wc, mwh.Webhooks)
	})

	t.Run("ensure a Validating Webhook is created if it doesn't exist", func(t *testing.T) {
		client := fake.NewClientBuilder().Build()
		wc := WebhookConfig{
			CABundel:         caBundel,
			ServiceName:      testServiceName,
			ServiceNamespace: testNamespaceName,
		}
		err := EnsureWebhookConfigurationFor(ctx, client, wc, ValidatingWebHook)
		require.NoError(t, err)

		vwh := &admissionregistrationv1.ValidatingWebhookConfiguration{}
		err = client.Get(ctx, types.NamespacedName{Name: ValidationWebhookName}, vwh)

		require.NoError(t, err)
		require.NotNil(t, vwh)
		require.Equal(t, vwh.Name, ValidationWebhookName)
		require.NotNil(t, vwh.Webhooks)

		require.Equal(t, int32(443), *vwh.Webhooks[0].ClientConfig.Service.Port)
		require.Equal(t, FunctionValidationWebhookPath, *vwh.Webhooks[0].ClientConfig.Service.Path)
		require.Equal(t, wc.ServiceName, vwh.Webhooks[0].ClientConfig.Service.Name)
		require.Equal(t, wc.ServiceNamespace, vwh.Webhooks[0].ClientConfig.Service.Namespace)
		require.Contains(t, vwh.Webhooks[0].Rules[0].Resources, "functions")
		require.Contains(t, vwh.Webhooks[0].Rules[0].Resources, "functions/status")
		require.Contains(t, vwh.Webhooks[0].Rules[1].Resources, "gitrepositories")
		require.Contains(t, vwh.Webhooks[0].Rules[1].Resources, "gitrepositories/status")

	})

	t.Run("ensure a Mutating Webhook is updated if it already exist", func(t *testing.T) {
		client := fake.NewClientBuilder().Build()
		mwh := &admissionregistrationv1.MutatingWebhookConfiguration{
			ObjectMeta: metav1.ObjectMeta{
				Name: DefaultingWebhookName,
				Labels: map[string]string{
					"dont-remove-me": "true",
				},
			},
		}
		err := client.Create(ctx, mwh)
		require.NoError(t, err)

		wc := WebhookConfig{
			CABundel:         caBundel,
			ServiceName:      testServiceName,
			ServiceNamespace: testNamespaceName,
		}
		err = EnsureWebhookConfigurationFor(ctx, client, wc, MutatingWebhook)
		require.NoError(t, err)

		UpdateMwh := &admissionregistrationv1.MutatingWebhookConfiguration{}
		err = client.Get(ctx, types.NamespacedName{Name: DefaultingWebhookName}, UpdateMwh)

		require.NoError(t, err)
		require.NotNil(t, UpdateMwh)
		require.Equal(t, UpdateMwh.Name, DefaultingWebhookName)
		require.NotNil(t, UpdateMwh.Webhooks)
		require.Contains(t, UpdateMwh.Labels, "dont-remove-me")
		validateMutatingWebhook(t, wc, UpdateMwh.Webhooks)
	})

	t.Run("ensure a Validating Webhook is updated if it already exist", func(t *testing.T) {
		client := fake.NewClientBuilder().Build()
		vwh := &admissionregistrationv1.ValidatingWebhookConfiguration{
			ObjectMeta: metav1.ObjectMeta{
				Name: ValidationWebhookName,
				Labels: map[string]string{
					"dont-remove-me": "true",
				},
			},
		}
		err := client.Create(ctx, vwh)
		require.NoError(t, err)

		wc := WebhookConfig{
			CABundel:         caBundel,
			ServiceName:      testServiceName,
			ServiceNamespace: testNamespaceName,
		}
		err = EnsureWebhookConfigurationFor(ctx, client, wc, ValidatingWebHook)
		require.NoError(t, err)

		UpdateVwh := &admissionregistrationv1.ValidatingWebhookConfiguration{}
		err = client.Get(ctx, types.NamespacedName{Name: ValidationWebhookName}, UpdateVwh)

		require.NoError(t, err)
		validateValidationWebhook(t, wc, UpdateVwh)
	})
}

func validateMutatingWebhook(t *testing.T, wc WebhookConfig, webhooks []admissionregistrationv1.MutatingWebhook) {
	require.Len(t, webhooks, 2)
	functionWebhook := webhooks[0]

	require.NotNil(t, functionWebhook.ClientConfig.Service.Port)
	require.Equal(t, int32(443), *functionWebhook.ClientConfig.Service.Port)
	require.NotNil(t, functionWebhook.ClientConfig.Service.Path)
	require.Equal(t, FunctionDefaultingWebhookPath, *functionWebhook.ClientConfig.Service.Path)
	require.Equal(t, wc.ServiceName, functionWebhook.ClientConfig.Service.Name)
	require.Equal(t, wc.ServiceNamespace, functionWebhook.ClientConfig.Service.Namespace)
	require.Len(t, functionWebhook.Rules, 1)
	require.Contains(t, functionWebhook.Rules[0].Resources, "functions")
	require.Contains(t, functionWebhook.Rules[0].Resources, "functions/status")
	require.Contains(t, functionWebhook.Rules[0].APIVersions, DeprecatedServerlessAPIVersion)
	require.Contains(t, functionWebhook.Rules[0].APIVersions, ServerlessCurrentAPIVersion)

	secretWebhook := webhooks[1]
	require.NotNil(t, secretWebhook.ClientConfig.Service.Port)
	require.Equal(t, int32(443), *secretWebhook.ClientConfig.Service.Port)
	require.NotNil(t, secretWebhook.ClientConfig.Service.Path)
	require.Equal(t, RegistryConfigDefaultingWebhookPath, *secretWebhook.ClientConfig.Service.Path)
	require.Equal(t, wc.ServiceName, secretWebhook.ClientConfig.Service.Name)
	require.Equal(t, wc.ServiceNamespace, secretWebhook.ClientConfig.Service.Namespace)
	require.Len(t, secretWebhook.Rules, 1)
	require.Contains(t, secretWebhook.Rules[0].Resources, "secrets")

}

func validateValidationWebhook(t *testing.T, wc WebhookConfig, webhook *admissionregistrationv1.ValidatingWebhookConfiguration) {
	require.Contains(t, webhook.Labels, "dont-remove-me")
	require.Equal(t, webhook.Name, ValidationWebhookName)
	require.Len(t, webhook.Webhooks, 1)
	functionWebhook := webhook.Webhooks[0]

	require.Equal(t, int32(443), *functionWebhook.ClientConfig.Service.Port)
	require.Equal(t, FunctionValidationWebhookPath, *functionWebhook.ClientConfig.Service.Path)
	require.Equal(t, wc.ServiceName, functionWebhook.ClientConfig.Service.Name)
	require.Equal(t, wc.ServiceNamespace, functionWebhook.ClientConfig.Service.Namespace)
	require.Len(t, functionWebhook.Rules, 2)
	require.Contains(t, functionWebhook.Rules[0].Resources, "functions")
	require.Contains(t, functionWebhook.Rules[0].Resources, "functions/status")
	require.Contains(t, functionWebhook.Rules[0].APIVersions, DeprecatedServerlessAPIVersion)
	require.Contains(t, functionWebhook.Rules[0].APIVersions, ServerlessCurrentAPIVersion)
	require.Contains(t, functionWebhook.Rules[1].Resources, "gitrepositories")
	require.Contains(t, functionWebhook.Rules[1].Resources, "gitrepositories/status")
}
