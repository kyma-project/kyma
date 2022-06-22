package resources

import (
	"context"
	"reflect"

	"github.com/pkg/errors"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
	ctlrclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type WebhookConfig struct {
	CABundel         []byte
	ServiceName      string
	ServiceNamespace string
}
type WebHookType string

const (
	MutatingWebhook   WebHookType = "Mutating"
	ValidatingWebHook WebHookType = "Validating"

	serverlessAPIGroup   = "serverless.kyma-project.io"
	serverlessAPIVersion = "v1alpha1"

	DefaultingWebhookName     = "defaulting.webhook.serverless.kyma-project.io"
	SecretMutationWebhookName = "mutating.secret.webhook.serverless.kyma-project.io"
	ValidationWebhookName     = "validation.webhook.serverless.kyma-project.io"

	FunctionDefaultingWebhookPath       = "/defaulting/functions"
	RegistryConfigDefaultingWebhookPath = "/defaulting/registry-config-secrets"
	FunctionValidationWebhookPath       = "/validation/function"

	RemoteRegistryLabelKey = "serverless.kyma-project.io/remote-registry"
)

func EnsureWebhookConfigurationFor(ctx context.Context, client ctlrclient.Client, config WebhookConfig, wt WebHookType) error {
	if wt == MutatingWebhook {
		return ensureMutatingWebhookConfigFor(ctx, client, config)
	}
	return ensureValidatingWebhookConfigFor(ctx, client, config)
}

func ensureMutatingWebhookConfigFor(ctx context.Context, client ctlrclient.Client, config WebhookConfig) error {
	mwhc := &admissionregistrationv1.MutatingWebhookConfiguration{}
	if err := client.Get(ctx, types.NamespacedName{Name: DefaultingWebhookName}, mwhc); err != nil {
		if apiErrors.IsNotFound(err) {
			return client.Create(ctx, createMutatingWebhookConfiguration(config))
		}
		return errors.Wrapf(err, "failed to get defaulting MutatingWebhookConfiguration: %s", DefaultingWebhookName)
	}
	ensuredMwhc := createMutatingWebhookConfiguration(config)
	if !reflect.DeepEqual(ensuredMwhc.Webhooks, mwhc.Webhooks) {
		ensuredMwhc.ObjectMeta = *mwhc.ObjectMeta.DeepCopy()
		return client.Update(ctx, ensuredMwhc)
	}
	return nil
}

func ensureValidatingWebhookConfigFor(ctx context.Context, client ctlrclient.Client, config WebhookConfig) error {
	vwhc := &admissionregistrationv1.ValidatingWebhookConfiguration{}
	if err := client.Get(ctx, types.NamespacedName{Name: ValidationWebhookName}, vwhc); err != nil {
		if apiErrors.IsNotFound(err) {
			return client.Create(ctx, createValidatingWebhookConfiguration(config))
		}
		return errors.Wrapf(err, "failed to get validation ValidatingWebhookConfiguration: %s", ValidationWebhookName)
	}
	ensuredVwhc := createValidatingWebhookConfiguration(config)
	if !reflect.DeepEqual(ensuredVwhc.Webhooks, vwhc.Webhooks) {
		ensuredVwhc.ObjectMeta = *vwhc.ObjectMeta.DeepCopy()
		return client.Update(ctx, ensuredVwhc)
	}
	return nil
}

func createMutatingWebhookConfiguration(config WebhookConfig) *admissionregistrationv1.MutatingWebhookConfiguration {
	return &admissionregistrationv1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: DefaultingWebhookName,
		},
		Webhooks: []admissionregistrationv1.MutatingWebhook{
			getFunctionMutatingWebhookCfg(config),
			getRegistryConfigSecretMutatingWebhook(config),
		},
	}
}

func getFunctionMutatingWebhookCfg(config WebhookConfig) admissionregistrationv1.MutatingWebhook {
	failurePolicy := admissionregistrationv1.Fail
	matchPolicy := admissionregistrationv1.Exact
	reinvocationPolicy := admissionregistrationv1.NeverReinvocationPolicy
	scope := admissionregistrationv1.AllScopes
	sideEffects := admissionregistrationv1.SideEffectClassNone

	return admissionregistrationv1.MutatingWebhook{
		Name: DefaultingWebhookName,
		AdmissionReviewVersions: []string{
			"v1beta1",
			"v1",
		},
		ClientConfig: admissionregistrationv1.WebhookClientConfig{
			CABundle: config.CABundel,
			Service: &admissionregistrationv1.ServiceReference{
				Namespace: config.ServiceNamespace,
				Name:      config.ServiceName,
				Path:      pointer.String(FunctionDefaultingWebhookPath),
				Port:      pointer.Int32(443),
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
					Resources: []string{"functions", "functions/status"},
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
	}
}

func getRegistryConfigSecretMutatingWebhook(config WebhookConfig) admissionregistrationv1.MutatingWebhook {
	failurePolicy := admissionregistrationv1.Fail
	matchPolicy := admissionregistrationv1.Exact
	sideEffects := admissionregistrationv1.SideEffectClassNone
	secretSelector := map[string]string{
		RemoteRegistryLabelKey: "config",
	}

	return admissionregistrationv1.MutatingWebhook{
		Name: SecretMutationWebhookName,
		ClientConfig: admissionregistrationv1.WebhookClientConfig{
			CABundle: config.CABundel,
			Service: &admissionregistrationv1.ServiceReference{
				Namespace: config.ServiceNamespace,
				Name:      config.ServiceName,
				Path:      pointer.String(RegistryConfigDefaultingWebhookPath),
				Port:      pointer.Int32(443),
			},
		},
		FailurePolicy:           &failurePolicy,
		MatchPolicy:             &matchPolicy,
		TimeoutSeconds:          pointer.Int32(30),
		SideEffects:             &sideEffects,
		AdmissionReviewVersions: []string{"v1beta1", "v1"},
		ObjectSelector: &metav1.LabelSelector{
			MatchLabels: secretSelector,
		},
		Rules: []admissionregistrationv1.RuleWithOperations{
			{
				Rule: admissionregistrationv1.Rule{
					APIGroups:   []string{""},
					APIVersions: []string{"v1"},
					Resources:   []string{"secrets"},
				},
				Operations: []admissionregistrationv1.OperationType{
					admissionregistrationv1.Create,
					admissionregistrationv1.Update,
				},
			},
		},
	}
}

func createValidatingWebhookConfiguration(config WebhookConfig) *admissionregistrationv1.ValidatingWebhookConfiguration {
	failurePolicy := admissionregistrationv1.Fail
	matchPolicy := admissionregistrationv1.Exact
	scope := admissionregistrationv1.AllScopes
	sideEffects := admissionregistrationv1.SideEffectClassNone

	return &admissionregistrationv1.ValidatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: ValidationWebhookName,
		},
		Webhooks: []admissionregistrationv1.ValidatingWebhook{
			{
				Name: ValidationWebhookName,
				AdmissionReviewVersions: []string{
					"v1beta1",
					"v1",
				},
				ClientConfig: admissionregistrationv1.WebhookClientConfig{
					CABundle: config.CABundel,
					Service: &admissionregistrationv1.ServiceReference{
						Namespace: config.ServiceNamespace,
						Name:      config.ServiceName,
						Path:      pointer.String(FunctionValidationWebhookPath),
						Port:      pointer.Int32(443),
					},
				},
				FailurePolicy: &failurePolicy,
				MatchPolicy:   &matchPolicy,
				Rules: []admissionregistrationv1.RuleWithOperations{
					{
						Rule: admissionregistrationv1.Rule{
							APIGroups: []string{
								serverlessAPIGroup,
							},
							APIVersions: []string{
								serverlessAPIVersion,
							},
							Resources: []string{"functions", "functions/status"},
							Scope:     &scope,
						},
						Operations: []admissionregistrationv1.OperationType{
							admissionregistrationv1.Create,
							admissionregistrationv1.Update,
						},
					},
					{
						Rule: admissionregistrationv1.Rule{
							APIGroups: []string{
								serverlessAPIGroup,
							},
							APIVersions: []string{
								serverlessAPIVersion,
							},
							Resources: []string{"gitrepositories", "gitrepositories/status"},
							Scope:     &scope,
						},
						Operations: []admissionregistrationv1.OperationType{
							admissionregistrationv1.Create,
							admissionregistrationv1.Update,
							admissionregistrationv1.Delete,
						},
					},
				},
				SideEffects:    &sideEffects,
				TimeoutSeconds: pointer.Int32(30),
			},
		},
	}
}
