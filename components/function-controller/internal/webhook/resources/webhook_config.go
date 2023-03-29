package resources

import (
	"bytes"
	"context"
	"github.com/pkg/errors"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	"k8s.io/apimachinery/pkg/types"
	ctlrclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type WebhookConfig struct {
	CABundle         []byte
	ServiceName      string
	ServiceNamespace string
}
type WebHookType string

const (
	MutatingWebhook   WebHookType = "Mutating"
	ValidatingWebHook WebHookType = "Validating"

	DefaultingWebhookName = "defaulting.webhook.serverless.kyma-project.io"
	ValidationWebhookName = "validation.webhook.serverless.kyma-project.io"

	FunctionDefaultingWebhookPath       = "/defaulting/functions"
	RegistryConfigDefaultingWebhookPath = "/defaulting/registry-config-secrets"
	FunctionValidationWebhookPath       = "/validation/function"
)

func InjectCABundleIntoWebhooks(ctx context.Context, client ctlrclient.Client, config WebhookConfig, wt WebHookType) error {
	switch wt {
	case MutatingWebhook:
		return injectCAIntoMutatingWebhook(ctx, client, config.CABundle)
	case ValidatingWebHook:
		return injectCAIntoValidationWebhook(ctx, client, config.CABundle)
	default:
		return errors.Errorf("Unknow webhook type: %s", wt)
	}
}

func injectCAIntoMutatingWebhook(ctx context.Context, client ctlrclient.Client, CABundle []byte) error {
	mwhc := &admissionregistrationv1.MutatingWebhookConfiguration{}
	if err := client.Get(ctx, types.NamespacedName{Name: DefaultingWebhookName}, mwhc); err != nil {
		return errors.Wrapf(err, "failed to get defaulting MutatingWebhookConfiguration: %s", DefaultingWebhookName)
	}
	var updatedWebhooks []admissionregistrationv1.MutatingWebhook
	shouldBeUpdated := false
	for _, webhook := range mwhc.Webhooks {
		if !bytes.Equal(webhook.ClientConfig.CABundle, CABundle) {
			shouldBeUpdated = true
			webhook.ClientConfig.CABundle = CABundle
			updatedWebhooks = append(updatedWebhooks, webhook)
		}
	}

	if shouldBeUpdated {
		mwhc.Webhooks = updatedWebhooks
		return errors.Wrap(client.Update(ctx, mwhc), "while  injecting CA Bundle into mutation webhook configuration")
	}
	return nil
}

func injectCAIntoValidationWebhook(ctx context.Context, client ctlrclient.Client, CABundle []byte) error {
	vwhc := &admissionregistrationv1.ValidatingWebhookConfiguration{}
	if err := client.Get(ctx, types.NamespacedName{Name: ValidationWebhookName}, vwhc); err != nil {
		return errors.Wrapf(err, "failed to get validation ValidatingWebhookConfiguration: %s", ValidationWebhookName)
	}

	var updatedWebhooks []admissionregistrationv1.ValidatingWebhook
	shouldBeUpdated := false
	for _, webhook := range vwhc.Webhooks {
		if !bytes.Equal(webhook.ClientConfig.CABundle, CABundle) {
			shouldBeUpdated = true
			webhook.ClientConfig.CABundle = CABundle
			updatedWebhooks = append(updatedWebhooks, webhook)
		}

	}

	if shouldBeUpdated {
		vwhc.Webhooks = updatedWebhooks
		return errors.Wrap(client.Update(ctx, vwhc), "while injecting CA Bundle into validation webhook configuration")
	}
	return nil
}
