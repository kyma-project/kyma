package dex

import (
	"context"
	"fmt"

	"github.com/kyma-project/kyma/components/uaa-activator/internal/waiter"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	"k8s.io/kubernetes/pkg/controller/deployment/util"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	dexConfigMapConfigKey                = "config.yaml"
	dexDeploymentStaticUserInitContainer = "dex-users-configurator"
)

// Config holds configuration for configurator
type Config struct {
	ConfigMap  client.ObjectKey
	Deployment client.ObjectKey
}

// Configurator provides functionality for updating Dex ConfigMap and Deployment
// to work with provided UAA Connector
type Configurator struct {
	cli               client.Client
	config            Config
	uaaConfigProvider uaaConfigProvider
}

// NewConfigurator returns a new instance of the Configurator
func NewConfigurator(config Config, cli client.Client, uaaConfigProvider uaaConfigProvider) *Configurator {
	return &Configurator{
		cli:               cli,
		config:            config,
		uaaConfigProvider: uaaConfigProvider,
	}
}

// ConfigureUAAInDex update the Dex ConfigMap.
//
// BE AWARE:
// - The `connectors` entry is overridden
// - The `enablePasswordDB` field is set to `false`
func (c *Configurator) ConfigureUAAInDex(ctx context.Context) error {
	// get configuration resources
	dexConfigYAML, err := c.dexConfigYAML(ctx)
	if err != nil {
		return errors.Wrap(err, "while fetching Dex configuration")
	}

	uaaConnectorYAML, err := c.uaaConnectorConfigYAML(ctx)
	if err != nil {
		return errors.Wrap(err, "while fetching UAA configuration")
	}

	// prepare dex config map suited for uaa connector
	dexConfigYAML["enablePasswordDB"] = "false"
	dexConfigYAML["connectors"] = uaaConnectorYAML

	marshaledDexConfigYAML, err := yaml.Marshal(dexConfigYAML)
	if err != nil {
		return errors.Wrap(err, "while marshaling Dex `config.yaml` entry")
	}

	// update dex config map in k8s cluster
	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		old := v1.ConfigMap{}
		if err := c.cli.Get(ctx, c.config.ConfigMap, &old); err != nil {
			return err
		}

		toUpdate := old.DeepCopy()
		toUpdate.Data[dexConfigMapConfigKey] = string(marshaledDexConfigYAML)
		// we need to return err itself here (not wrapped), so it can be identify correctly
		return c.cli.Update(ctx, toUpdate)
	})
	if err != nil {
		return errors.Wrap(err, "while updating Dex ConfigMap")
	}

	return nil
}

// ConfigureDexDeployment mutate dex deployment to use UAA connector
// and waits until Pods are restarted.
func (c *Configurator) ConfigureDexDeployment(ctx context.Context) error {
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		deploy := appsv1.Deployment{}
		err := c.cli.Get(ctx, c.config.Deployment, &deploy)
		if err != nil {
			return errors.Wrapf(err, "while fetching Dex deployment %s", c.config.Deployment)
		}

		var filterOutInitCont []v1.Container
		for _, container := range deploy.Spec.Template.Spec.InitContainers {
			if container.Name != dexDeploymentStaticUserInitContainer {
				filterOutInitCont = append(filterOutInitCont, container)
			}
		}
		deploy.Spec.Template.Spec.InitContainers = filterOutInitCont
		deploy.Spec.Template.Spec.Volumes = []v1.Volume{
			{
				Name: "config",
				VolumeSource: v1.VolumeSource{
					ConfigMap: &v1.ConfigMapVolumeSource{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "dex-config",
						},
						Items: []v1.KeyToPath{
							{Key: "config.yaml", Path: "config.yaml"},
						},
					},
				},
			},
		}

		// We need to return err itself here (not wrapped inside another error)
		// so it can be identify correctly.
		return c.cli.Update(ctx, &deploy)
	})
	if err != nil {
		return errors.Wrapf(err, "while updating Dex %q deployment", c.config.Deployment.String())
	}

	err = waiter.WaitForSuccess(ctx, c.dexDeploymentIsReady(ctx))
	if err != nil {
		return errors.Wrapf(err, "while waiting for ready Dex %q deployment", c.config.Deployment.String())
	}

	return nil
}

func (c *Configurator) dexConfigYAML(ctx context.Context) (map[string]interface{}, error) {
	dexConfigMap := v1.ConfigMap{}
	err := c.cli.Get(ctx, c.config.ConfigMap, &dexConfigMap)
	if err != nil {
		return nil, errors.Wrap(err, "while fetching Dex ConfigMap")
	}

	rawConfigYAML := dexConfigMap.Data[dexConfigMapConfigKey]

	dexConfigYAML := map[string]interface{}{}
	if err := yaml.Unmarshal([]byte(rawConfigYAML), &dexConfigYAML); err != nil {
		return nil, errors.Wrapf(err, "while unmarshaling %q entry from Dex ConfigMap", dexConfigMapConfigKey)
	}

	return dexConfigYAML, nil
}

func (c *Configurator) uaaConnectorConfigYAML(ctx context.Context) ([]interface{}, error) {
	uaaConnectorConfigString, err := c.uaaConfigProvider.RenderUAAConnectorConfig(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "while rendering UAA connector config")
	}

	var uaaConnectorConfigYAML []interface{}
	err = yaml.Unmarshal([]byte(uaaConnectorConfigString), &uaaConnectorConfigYAML)
	if err != nil {
		return nil, errors.Wrap(err, "while unmarshaling UAA connector config")
	}

	return uaaConnectorConfigYAML, nil
}

func (c *Configurator) dexDeploymentIsReady(ctx context.Context) func() error {
	return func() error {
		dexDeploy := &appsv1.Deployment{}
		if err := c.cli.Get(ctx, c.config.Deployment, dexDeploy); err != nil {
			return errors.Wrap(err, "while fetching dex deployment")
		}

		// we need to fetch all Deployment replica sets to find the newest one
		rsList, err := c.listReplicaSets(ctx, dexDeploy)
		if err != nil {
			return errors.Wrap(err, "while listing Dex replica sets")
		}

		// find the newest replica set
		newReplicaSet := util.FindNewReplicaSet(dexDeploy, rsList)
		if newReplicaSet == nil {
			return errors.New("the new replica set doesn't exist yet")
		}

		// wait until the newest replica set is deployed
		err = c.isDeploymentReady(newReplicaSet, dexDeploy)
		if err != nil {
			return errors.Wrap(err, "while checking if new Dex replica set is deployed")
		}

		return nil
	}
}

func (c *Configurator) listReplicaSets(ctx context.Context, deployment *appsv1.Deployment) ([]*appsv1.ReplicaSet, error) {
	selector, err := metav1.LabelSelectorAsSelector(deployment.Spec.Selector)
	if err != nil {
		return nil, errors.Wrap(err, "while creating label selector")
	}

	all := &appsv1.ReplicaSetList{}
	err = c.cli.List(ctx, all, client.MatchingLabelsSelector{Selector: selector}, client.InNamespace(deployment.Namespace))
	if err != nil {
		return nil, errors.Wrap(err, "while fetching all replica set based on label selector")
	}

	// Only include those whose ControllerRef matches the Deployment.
	owned := make([]*appsv1.ReplicaSet, 0, len(all.Items))
	for _, rs := range all.Items {
		if metav1.IsControlledBy(&rs, deployment) {
			owned = append(owned, &rs)
		}
	}
	return owned, nil
}

func (c *Configurator) isDeploymentReady(rs *appsv1.ReplicaSet, dep *appsv1.Deployment) error {
	expectedReady := *dep.Spec.Replicas - util.MaxUnavailable(*dep)
	if !(rs.Status.ReadyReplicas >= expectedReady) {
		return fmt.Errorf("deployment is not ready: %s/%s. %d out of %d expected pods are ready", dep.Namespace, dep.Name, rs.Status.ReadyReplicas, expectedReady)
	}
	return nil
}
