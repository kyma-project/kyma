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
	caBundle = []byte("LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUVEakNDQW5ZQ0NRRE5vZFh6bERqdk9qQU5CZ2txaGtpRzl3MEJBUXNGQURCSk1SVXdFd1lEVlFRS0RBeGoKWlhKMExXMWhibUZuWlhJeEZUQVRCZ05WQkFzTURHTmxjblF0YldGdVlXZGxjakVaTUJjR0ExVUVBd3dRYTNsdApZUzFsZUdGdGNHeGxMbU52YlRBZUZ3MHlNVEE1TWpreE5ERXhOVGRhRncweU1qQTVNamt4TkRFeE5UZGFNRWt4CkZUQVRCZ05WQkFvTURHTmxjblF0YldGdVlXZGxjakVWTUJNR0ExVUVDd3dNWTJWeWRDMXRZVzVoWjJWeU1Sa3cKRndZRFZRUUREQkJyZVcxaExXVjRZVzF3YkdVdVkyOXRNSUlCb2pBTkJna3Foa2lHOXcwQkFRRUZBQU9DQVk4QQpNSUlCaWdLQ0FZRUFvMHNzdEtYZ01sR0FvV0F6WmxCZVNHMTVJZ0VBRkZwbWFTcnArQ0Y1bGdjTXVSZFNBamlXCk1tS0F5OWlzdy9meTN3YmFWUWNoSWFLZXJvZzVFcnYyR1hlMmZaK0Q2anJpbFp4SG01ZURxZmtxdmQrcXh6K3IKN25Dd1JOS2t0Q093YmJqOWtMSWduQkdUVGZQL1NiVUMxd2R5R3BuYnpwVXo1b3RvbVJRMWlsWi9OZFl2T2UwYworK21zZ2xSUVpNWmNtZVI0RWJhZlNwRTdWT1M2L0srbTZpMzRZMzdmSTBhSUlUV2lVR1RuMzRVMEZhTG9DSTF5CjVNaFhSYUhyQTAyWVBsNmhUMlJFVVlRUTBuNUMxVnpQYm02VHkvUjdoOXI2RnoxQWxPQ0I5NnZtalNPemR2dmsKaHdMcVRqdHArbFNNcW1iRExEV2ZVY2JpVFIwbzliQ3Rxam1aUjQwcXhqa1hieW5hYU42TzFnbm9Fem50a1J3Sgo1cytZaVVtTHpvVFIyV2V1bHc5VVZaZXdiQm85Zk1FSFVRa1RJOEd0aWVjUDhNMGF4R1RSSVFCWUc0bTBYcEF1CmFrd2ZPNlRlZzR2T1FTQmpVanE0Rm1nYkZkVVZsUkZrVm95NXkra3JrV2FpT21hSFlqWHViMkduVWQ1WFBiamgKeG56a1R4UlE3UFh6QWdNQkFBRXdEUVlKS29aSWh2Y05BUUVMQlFBRGdnR0JBRCs5QlExWERXY0pkZ2Jma2ZJeApNa1ZLc3pXRWdvUC9GWWNmcVVEc1drcTNkR3M0Q1Z4ODRzTjNRd0hSY2JkV2trVTd1WXh3VmE0NVZWbVZrZjhmClZGN3ZiblhWWUoydHp5K2lTb0JDcVFjMUNHV1ZwQmdIOGlFVWdYb2hwdTczWjhtSkJhNUhQT1lXTHM5dEVlVTUKdy9VOG9HOUJZRHRsclBGSzZLWkJpYU8veERMQXlFOFRxUEc3U3oxUXJFdC9HbE1wS1RHakFUYXNNQ2hzT0IrMgpuc2xLdExuNXZoS3hOdkRIYjJ4Y1pGZHlOdGQ5Vk9mcFQvNXZ2VE4wb3ozcERJd2RhUkNyTTMwU0tCZVl1OFdGClJPb1NsNUE5dDl3S08vZ0xwcFJ1WW9IdzU4Z1VudWpCOUoyT1oxaUp2YUFLTXZMNGFFbTE4VjVCYkZaUVdPUFkKcWRYZGRnYk80R0VsUFRXdUxFb01HZTM5UnBITmVtMmcrNFdUcjl1K2FkSVN3MnBQbjFRYXhqbzJmb0ZiUUpUcgpXeXR5MnVLU2diU3RzNzEvOUZ4bUIvM3Z6Y21oOUhLdzJWb0FQSzY4ckdhUkhvTjdqZ0dCR3JtdDdHZ01UeGlqCmRuOEt1OHBhRERWSTFldUx6RlJwUlBvdVVwOEVDdmRHTkxoRTJrc0dsdEk3WEE9PQotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg==")
)

func TestInjectCABundleIntoMutatingWebhooks(t *testing.T) {
	testCases := []struct {
		name     string
		webhooks []admissionregistrationv1.MutatingWebhook
	}{
		{
			name:     "ensure a Mutating Webhook CABundles are updated if empty",
			webhooks: []admissionregistrationv1.MutatingWebhook{{}, {}},
		},
		{
			name: "ca bundle is already present in mutating webhook's configs",
			webhooks: []admissionregistrationv1.MutatingWebhook{
				{ClientConfig: admissionregistrationv1.WebhookClientConfig{CABundle: caBundle}},
				{ClientConfig: admissionregistrationv1.WebhookClientConfig{CABundle: caBundle}},
			},
		},
		{
			name: "ca bundle is different in mutating webhook's configs",
			webhooks: []admissionregistrationv1.MutatingWebhook{
				{ClientConfig: admissionregistrationv1.WebhookClientConfig{CABundle: []byte("aaaa")}},
				{ClientConfig: admissionregistrationv1.WebhookClientConfig{CABundle: []byte("bbbb")}},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ctx := context.TODO()
			client := fake.NewClientBuilder().Build()
			mwh := &admissionregistrationv1.MutatingWebhookConfiguration{
				ObjectMeta: metav1.ObjectMeta{
					Name: DefaultingWebhookName,
					Labels: map[string]string{
						"dont-remove-me": "true",
					},
				},
				Webhooks: testCase.webhooks,
			}
			err := client.Create(ctx, mwh)
			require.NoError(t, err)

			wc := WebhookConfig{
				CABundle:         caBundle,
				ServiceName:      testServiceName,
				ServiceNamespace: testNamespaceName,
			}
			err = InjectCABundleIntoWebhooks(ctx, client, wc, MutatingWebhook)
			require.NoError(t, err)

			updatedWh := &admissionregistrationv1.MutatingWebhookConfiguration{}
			err = client.Get(ctx, types.NamespacedName{Name: DefaultingWebhookName}, updatedWh)

			require.NoError(t, err)
			require.NotNil(t, updatedWh)
			require.Equal(t, updatedWh.Name, DefaultingWebhookName)
			require.NotNil(t, updatedWh.Webhooks)
			require.Contains(t, updatedWh.Labels, "dont-remove-me")
			require.Len(t, updatedWh.Webhooks, 2)
			require.Equal(t, wc.CABundle, updatedWh.Webhooks[0].ClientConfig.CABundle)
			require.Equal(t, wc.CABundle, updatedWh.Webhooks[1].ClientConfig.CABundle)
		})
	}
}
func TestInjectCABundleIntoValidationWebhooks(t *testing.T) {
	testCases := []struct {
		name     string
		webhooks []admissionregistrationv1.ValidatingWebhook
	}{
		{
			name:     "ensure a Validating Webhook CABundles are updated if empty",
			webhooks: []admissionregistrationv1.ValidatingWebhook{{}, {}},
		},
		{
			name: "ca bundle is already present in validation webhook's configs",
			webhooks: []admissionregistrationv1.ValidatingWebhook{{
				ClientConfig: admissionregistrationv1.WebhookClientConfig{CABundle: caBundle}},
				{ClientConfig: admissionregistrationv1.WebhookClientConfig{CABundle: caBundle}},
			},
		},
		{
			name: "ca bundle is different in validation webhook's configs",
			webhooks: []admissionregistrationv1.ValidatingWebhook{{
				ClientConfig: admissionregistrationv1.WebhookClientConfig{CABundle: []byte("cccc")}},
				{ClientConfig: admissionregistrationv1.WebhookClientConfig{CABundle: []byte("dddd")}},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ctx := context.TODO()
			client := fake.NewClientBuilder().Build()
			vwh := &admissionregistrationv1.ValidatingWebhookConfiguration{
				ObjectMeta: metav1.ObjectMeta{
					Name: ValidationWebhookName,
					Labels: map[string]string{
						"dont-remove-me": "true",
					},
				},
				Webhooks: testCase.webhooks,
			}
			err := client.Create(ctx, vwh)
			require.NoError(t, err)

			wc := WebhookConfig{
				CABundle:         caBundle,
				ServiceName:      testServiceName,
				ServiceNamespace: testNamespaceName,
			}
			err = InjectCABundleIntoWebhooks(ctx, client, wc, ValidatingWebHook)
			require.NoError(t, err)

			updatedWh := &admissionregistrationv1.ValidatingWebhookConfiguration{}
			err = client.Get(ctx, types.NamespacedName{Name: ValidationWebhookName}, updatedWh)

			require.NoError(t, err)
			require.Len(t, updatedWh.Webhooks, 2)
			require.Equal(t, wc.CABundle, updatedWh.Webhooks[0].ClientConfig.CABundle)
			require.Equal(t, wc.CABundle, updatedWh.Webhooks[1].ClientConfig.CABundle)
		})
	}
}
