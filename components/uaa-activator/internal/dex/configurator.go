package dex

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/kyma-project/kyma/components/uaa-activator/internal/repeat"
	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	intstrutil "k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	dexConfigMapConfigKey                = "config.yaml"
	dexDeploymentStaticUserInitContainer = "dex-users-configurator"
	modifiedAtAnnotationKey              = "uaa-activator.kyma-project.io/modified-at"
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

// EnsureUAAConnectorInDexConfigMap update the Dex ConfigMap.
//
// BE AWARE:
// - The `connectors` entry is overridden
// - The `enablePasswordDB` field is set to `false`
func (c *Configurator) EnsureUAAConnectorInDexConfigMap(ctx context.Context) error {
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
	dexConfigYAML["enablePasswordDB"] = false
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

// EnsureConfiguredUAAInDexDeployment mutate dex deployment to use UAA connector
// and waits until Pods are restarted.
func (c *Configurator) EnsureConfiguredUAAInDexDeployment(ctx context.Context) error {
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
		// needs to be set/updated to ensure that deployment will be restarted to reload new config
		deploy.Spec.Template.Annotations[modifiedAtAnnotationKey] = currentTimestamp()

		// We need to return err itself here (not wrapped inside another error)
		// so it can be identify correctly.
		return c.cli.Update(ctx, &deploy)
	})
	if err != nil {
		return errors.Wrapf(err, "while updating Dex %q deployment", c.config.Deployment.String())
	}

	err = repeat.UntilSuccess(ctx, c.dexDeploymentIsReady(ctx))
	if err != nil {
		return errors.Wrapf(err, "while waiting for ready Dex %q deployment", c.config.Deployment.String())
	}

	return nil
}

func currentTimestamp() string {
	return fmt.Sprintf("%v", metav1.Now().Unix())
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
		dexDeploy := appsv1.Deployment{}
		if err := c.cli.Get(ctx, c.config.Deployment, &dexDeploy); err != nil {
			return errors.Wrap(err, "while fetching dex deployment")
		}

		// we need to fetch all Deployment replica sets to find the newest one
		rsList, err := c.listReplicaSets(ctx, &dexDeploy)
		if err != nil {
			return errors.Wrap(err, "while listing Dex replica sets")
		}

		// find the newest replica set
		newReplicaSet := FindNewReplicaSet(&dexDeploy, rsList)
		if newReplicaSet == nil {
			return errors.New("the new replica set doesn't exist yet")
		}

		// wait until the newest replica set is deployed
		err = c.isDeploymentReady(newReplicaSet, &dexDeploy)
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

	all := appsv1.ReplicaSetList{}
	err = c.cli.List(ctx, &all, client.MatchingLabelsSelector{Selector: selector}, client.InNamespace(deployment.Namespace))
	if err != nil {
		return nil, errors.Wrap(err, "while fetching all replica set based on label selector")
	}

	// Only include those whose ControllerRef matches the Deployment.
	owned := make([]*appsv1.ReplicaSet, 0, len(all.Items))
	for i := range all.Items {
		rs := &all.Items[i]
		if metav1.IsControlledBy(rs, deployment) {
			owned = append(owned, rs)
		}
	}
	return owned, nil
}

func (c *Configurator) isDeploymentReady(rs *appsv1.ReplicaSet, dep *appsv1.Deployment) error {
	expectedReady := *dep.Spec.Replicas - MaxUnavailable(*dep)
	if !(rs.Status.ReadyReplicas >= expectedReady) {
		return fmt.Errorf("deployment is not ready: %s/%s. %d out of %d expected pods are ready", dep.Namespace, dep.Name, rs.Status.ReadyReplicas, expectedReady)
	}
	return nil
}


//TODO: I would move it to other package calledn k8sutils or something like that with comment that this code is from kubernetess/kubernetess
func FindNewReplicaSet(deployment *appsv1.Deployment, rsList []*appsv1.ReplicaSet) *appsv1.ReplicaSet {
	sort.Sort(ReplicaSetsByCreationTimestamp(rsList))
	for i := range rsList {
		if EqualIgnoreHash(&rsList[i].Spec.Template, &deployment.Spec.Template) {
			// In rare cases, such as after cluster upgrades, Deployment may end up with
			// having more than one new ReplicaSets that have the same template as its template,
			// see https://github.com/kubernetes/kubernetes/issues/40415
			// We deterministically choose the oldest new ReplicaSet.
			return rsList[i]
		}
	}
	// new ReplicaSet does not exist.
	return nil
}

// ReplicaSetsByCreationTimestamp sorts a list of ReplicaSet by creation timestamp, using their names as a tie breaker.
type ReplicaSetsByCreationTimestamp []*appsv1.ReplicaSet

func (o ReplicaSetsByCreationTimestamp) Len() int      { return len(o) }
func (o ReplicaSetsByCreationTimestamp) Swap(i, j int) { o[i], o[j] = o[j], o[i] }
func (o ReplicaSetsByCreationTimestamp) Less(i, j int) bool {
	if o[i].CreationTimestamp.Equal(&o[j].CreationTimestamp) {
		return o[i].Name < o[j].Name
	}
	return o[i].CreationTimestamp.Before(&o[j].CreationTimestamp)
}

// EqualIgnoreHash returns true if two given podTemplateSpec are equal, ignoring the diff in value of Labels[pod-template-hash]
// We ignore pod-template-hash because:
// 1. The hash result would be different upon podTemplateSpec API changes
//    (e.g. the addition of a new field will cause the hash code to change)
// 2. The deployment template won't have hash labels
func EqualIgnoreHash(template1, template2 *v1.PodTemplateSpec) bool {
	t1Copy := template1.DeepCopy()
	t2Copy := template2.DeepCopy()
	// Remove hash labels from template.Labels before comparing
	delete(t1Copy.Labels, appsv1.DefaultDeploymentUniqueLabelKey)
	delete(t2Copy.Labels, appsv1.DefaultDeploymentUniqueLabelKey)
	return apiequality.Semantic.DeepEqual(t1Copy, t2Copy)
}

// MaxUnavailable returns the maximum unavailable pods a rolling deployment can take.
func MaxUnavailable(deployment appsv1.Deployment) int32 {
	if !IsRollingUpdate(&deployment) || *(deployment.Spec.Replicas) == 0 {
		return int32(0)
	}
	// Error caught by validation
	_, maxUnavailable, _ := ResolveFenceposts(deployment.Spec.Strategy.RollingUpdate.MaxSurge, deployment.Spec.Strategy.RollingUpdate.MaxUnavailable, *(deployment.Spec.Replicas))
	if maxUnavailable > *deployment.Spec.Replicas {
		return *deployment.Spec.Replicas
	}
	return maxUnavailable
}

// IsRollingUpdate returns true if the strategy type is a rolling update.
func IsRollingUpdate(deployment *appsv1.Deployment) bool {
	return deployment.Spec.Strategy.Type == appsv1.RollingUpdateDeploymentStrategyType
}

// ResolveFenceposts resolves both maxSurge and maxUnavailable. This needs to happen in one
// step. For example:
//
// 2 desired, max unavailable 1%, surge 0% - should scale old(-1), then new(+1), then old(-1), then new(+1)
// 1 desired, max unavailable 1%, surge 0% - should scale old(-1), then new(+1)
// 2 desired, max unavailable 25%, surge 1% - should scale new(+1), then old(-1), then new(+1), then old(-1)
// 1 desired, max unavailable 25%, surge 1% - should scale new(+1), then old(-1)
// 2 desired, max unavailable 0%, surge 1% - should scale new(+1), then old(-1), then new(+1), then old(-1)
// 1 desired, max unavailable 0%, surge 1% - should scale new(+1), then old(-1)
func ResolveFenceposts(maxSurge, maxUnavailable *intstrutil.IntOrString, desired int32) (int32, int32, error) {
	surge, err := GetScaledValueFromIntOrPercent(intstrutil.ValueOrDefault(maxSurge, intstrutil.FromInt(0)), int(desired), true)
	if err != nil {
		return 0, 0, err
	}
	unavailable, err := GetScaledValueFromIntOrPercent(intstrutil.ValueOrDefault(maxUnavailable, intstrutil.FromInt(0)), int(desired), false)
	if err != nil {
		return 0, 0, err
	}

	if surge == 0 && unavailable == 0 {
		// Validation should never allow the user to explicitly use zero values for both maxSurge
		// maxUnavailable. Due to rounding down maxUnavailable though, it may resolve to zero.
		// If both fenceposts resolve to zero, then we should set maxUnavailable to 1 on the
		// theory that surge might not work due to quota.
		unavailable = 1
	}

	return int32(surge), int32(unavailable), nil
}

// GetScaledValueFromIntOrPercent is meant to replace GetValueFromIntOrPercent.
// This method returns a scaled value from an IntOrString type. If the IntOrString
// is a percentage string value it's treated as a percentage and scaled appropriately
// in accordance to the total, if it's an int value it's treated as a a simple value and
// if it is a string value which is either non-numeric or numeric but lacking a trailing '%' it returns an error.
func GetScaledValueFromIntOrPercent(intOrPercent *intstrutil.IntOrString, total int, roundUp bool) (int, error) {
	if intOrPercent == nil {
		return 0, errors.New("nil value for IntOrString")
	}
	value, isPercent, err := getIntOrPercentValueSafely(intOrPercent)
	if err != nil {
		return 0, fmt.Errorf("invalid value for IntOrString: %v", err)
	}
	if isPercent {
		if roundUp {
			value = int(math.Ceil(float64(value) * (float64(total)) / 100))
		} else {
			value = int(math.Floor(float64(value) * (float64(total)) / 100))
		}
	}
	return value, nil
}

func getIntOrPercentValueSafely(intOrStr *intstrutil.IntOrString) (int, bool, error) {
	switch intOrStr.Type {
	case intstrutil.Int:
		return intOrStr.IntValue(), false, nil
	case intstrutil.String:
		isPercent := false
		s := intOrStr.StrVal
		if strings.HasSuffix(s, "%") {
			isPercent = true
			s = strings.TrimSuffix(intOrStr.StrVal, "%")
		} else {
			return 0, false, fmt.Errorf("invalid type: string is not a percentage")
		}
		v, err := strconv.Atoi(s)
		if err != nil {
			return 0, false, fmt.Errorf("invalid value %q: %v", intOrStr.StrVal, err)
		}
		return int(v), isPercent, nil
	}
	return 0, false, fmt.Errorf("invalid type: neither int nor percentage")
}
