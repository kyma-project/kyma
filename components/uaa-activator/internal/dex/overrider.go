package dex

import (
	"context"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	dexOverrideConfigMapName      = "dex-override-uaa"
	dexOverrideConfigMapNamespace = "kyma-installer"
)

// Overrider provides functionality for creating Dex override ConfigMap used during the Kyma install/upgrade action
type Overrider struct {
	cli               client.Client
	uaaConfigProvider uaaConfigProvider
}

// NewOverrider returns a new instance of Overrider
func NewOverrider(cli client.Client, uaaConfigProvider uaaConfigProvider) *Overrider {
	return &Overrider{
		cli:               cli,
		uaaConfigProvider: uaaConfigProvider,
	}
}

// EnsureDexConfigMapOverride ensures that Dex config map exists and is up to date
func (o *Overrider) EnsureDexConfigMapOverride(ctx context.Context) error {
	uaaConnectorConfig, err := o.uaaConfigProvider.RenderUAAConnectorConfig(ctx)
	if err != nil {
		return errors.Wrap(err, "while rendering uaa connector config")
	}
	cm := v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dexOverrideConfigMapName,
			Namespace: dexOverrideConfigMapNamespace,
			Labels: map[string]string{
				"installer":                    "overrides",
				"component":                    "dex",
				"kyma-project.io/installation": "",
			},
		},
		Data: map[string]string{
			"dex.useStaticConnector": "false",
			"connectors":             uaaConnectorConfig,
		},
	}

	err = o.cli.Create(ctx, &cm)
	switch {
	case err == nil:
	case apiErrors.IsAlreadyExists(err):
		// if already created then ensure is updated
		err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			old := v1.ConfigMap{}
			key, err := client.ObjectKeyFromObject(&cm)
			if err != nil {
				return err
			}

			if err := o.cli.Get(ctx, key, &old); err != nil {
				return err
			}

			toUpdate := old.DeepCopy()
			toUpdate.Data = cm.Data
			toUpdate.Labels = cm.Labels
			return o.cli.Update(ctx, toUpdate)
		})
		if err != nil {
			return errors.Wrap(err, "while updating the ConfigMap with Dex overrides")
		}
	default:
		return errors.Wrap(err, "while creating the ConfigMap with Dex overrides")
	}

	return nil
}
