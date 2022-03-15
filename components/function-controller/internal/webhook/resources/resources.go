package resources

import (
	"context"
	"fmt"

	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	ctlrclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type WebhookConfig struct {
	Prefix           string
	Type             WebHookType
	CABundel         []byte
	ServiceName      string
	ServiceNamespace string
	Path             string
	Port             int32
	Resources        []string
}
type WebHookType string

const (
	MutatingWebhook   WebHookType = "Mutating"
	ValidatingWebHook WebHookType = "Validating"

	serverlessAPIGroup   = "serverless.kyma-project.io"
	serverlessAPIVersion = "v1alpha1"
)

func EnsureWebhookConfigurationFor(ctx context.Context, client ctlrclient.Client, config WebhookConfig) error {
	if config.Type == MutatingWebhook {
		mwh := createMutatingWebhookConfiguration(config)
		return client.Create(ctx, mwh)
	}

	mwh := createMutatingWebhookConfiguration(config)
	return client.Create(ctx, mwh)
}

func createMutatingWebhookConfiguration(config WebhookConfig) *admissionregistrationv1.MutatingWebhookConfiguration {
	failurePolicy := admissionregistrationv1.Fail
	matchPolicy := admissionregistrationv1.Exact
	reinvocationPolicy := admissionregistrationv1.NeverReinvocationPolicy
	scope := admissionregistrationv1.AllScopes
	sideEffects := admissionregistrationv1.SideEffectClassNone
	name := fmt.Sprintf("%s-defaulting.webhook.serverless.kyma-project.io", config.Prefix)

	return &admissionregistrationv1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Webhooks: []admissionregistrationv1.MutatingWebhook{
			{
				Name: name,
				AdmissionReviewVersions: []string{
					"v1beta1",
					"v1",
				},
				ClientConfig: admissionregistrationv1.WebhookClientConfig{
					CABundle: config.CABundel,
					Service: &admissionregistrationv1.ServiceReference{
						Namespace: config.ServiceNamespace,
						Name:      config.ServiceName,
						Path:      &config.Path,
						Port:      &config.Port,
					},
				},
				FailurePolicy:      &failurePolicy,
				MatchPolicy:        &matchPolicy,
				ReinvocationPolicy: &reinvocationPolicy,
				Rules: []admissionregistrationv1.RuleWithOperations{
					{
						Rule: admissionregistrationv1.Rule{
							APIGroups: []string{
								serverlessAPIGroup,
							},
							APIVersions: []string{
								serverlessAPIVersion,
							},
							Resources: config.Resources,
							Scope:     &scope,
						},
						Operations: []admissionregistrationv1.OperationType{
							admissionregistrationv1.Create,
							admissionregistrationv1.Update,
						},
					},
				},
				SideEffects:    &sideEffects,
				TimeoutSeconds: pointer.Int32(30),
			},
		},
	}
}

func createValidatingWebhookConfiguration(config WebhookConfig) *admissionregistrationv1.ValidatingWebhookConfiguration {
	failurePolicy := admissionregistrationv1.Fail
	matchPolicy := admissionregistrationv1.Exact
	reinvocationPolicy := admissionregistrationv1.NeverReinvocationPolicy
	scope := admissionregistrationv1.AllScopes
	sideEffects := admissionregistrationv1.SideEffectClassNone
	name := fmt.Sprintf("%s-defaulting.webhook.serverless.kyma-project.io", config.Prefix)

	return &admissionregistrationv1.ValidatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Webhooks: []admissionregistrationv1.ValidatingWebhook{
			{
				Name: name,
				AdmissionReviewVersions: []string{
					"v1beta1",
					"v1",
				},
				ClientConfig: admissionregistrationv1.WebhookClientConfig{
					CABundle: config.CABundel,
					Service: &admissionregistrationv1.ServiceReference{
						Namespace: config.ServiceNamespace,
						Name:      config.ServiceName,
						Path:      &config.Path,
						Port:      &config.Port,
					},
				},
				FailurePolicy:      &failurePolicy,
				MatchPolicy:        &matchPolicy,
				ReinvocationPolicy: &reinvocationPolicy,
				Rules: []admissionregistrationv1.RuleWithOperations{
					{
						Rule: admissionregistrationv1.Rule{
							APIGroups: []string{
								serverlessAPIGroup,
							},
							APIVersions: []string{
								serverlessAPIVersion,
							},
							Resources: config.Resources,
							Scope:     &scope,
						},
						Operations: []admissionregistrationv1.OperationType{
							admissionregistrationv1.Create,
							admissionregistrationv1.Update,
						},
					},
				},
				SideEffects:    &sideEffects,
				TimeoutSeconds: pointer.Int32(30),
			},
		},
	}
}
