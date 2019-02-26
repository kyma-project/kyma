package controller

import (
	"encoding/json"
	"fmt"

	"github.com/kyma-project/kyma/components/binding-usage-controller/internal/controller/pretty"
	sbuTypes "github.com/kyma-project/kyma/components/binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	"github.com/pkg/errors"
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type configMapClient interface {
	Get(name string, options metaV1.GetOptions) (*coreV1.ConfigMap, error)
	Update(*coreV1.ConfigMap) (*coreV1.ConfigMap, error)
}

// BindingUsageSpecStorage provides functionality to get/delete/save ServiceBindingUsage.Spec
type BindingUsageSpecStorage struct {
	cfgMapClient configMapClient
	cfgMapName   string
}

// NewBindingUsageSpecStorage returns new instance of BindingUsageSpecStorage
func NewBindingUsageSpecStorage(cfgMapClient configMapClient, cfgMapName string) *BindingUsageSpecStorage {
	return &BindingUsageSpecStorage{
		cfgMapClient: cfgMapClient,
		cfgMapName:   cfgMapName,
	}
}

// Get returns stored spec for given ServiceBindingUsage
func (c *BindingUsageSpecStorage) Get(usageNS, usageName string) (*UsageSpec, bool, error) {
	cfg, err := c.cfgMapClient.Get(c.cfgMapName, metaV1.GetOptions{})
	if err != nil {
		return nil, false, errors.Wrapf(err, "while getting config map with stored Spec for %s/%s", usageNS, usageName)
	}

	rawSpec, exists := cfg.Data[c.specUsedByKey(usageNS, usageName)]
	if !exists {
		return nil, false, nil
	}

	var spec UsageSpec
	if err := json.Unmarshal([]byte(rawSpec), &spec); err != nil {
		return nil, false, errors.Wrap(err, "while unmarshalling spec")
	}

	return &spec, true, nil
}

// Delete deletes stored spec for given ServiceBindingUsage
func (c *BindingUsageSpecStorage) Delete(namespace, name string) error {
	cfg, err := c.cfgMapClient.Get(c.cfgMapName, metaV1.GetOptions{})
	if err != nil {
		return errors.Wrapf(err, "while getting config map with stored Spec for %s/%s", namespace, name)
	}

	cfgCopy := cfg.DeepCopy()
	cfgKey := c.specUsedByKey(namespace, name)

	delete(cfgCopy.Data, cfgKey)

	if _, err := c.cfgMapClient.Update(cfgCopy); err != nil {
		return errors.Wrapf(err, "while updating config map %q", c.cfgMapName)
	}

	return nil
}

// Upsert upserts spec for given ServiceBindingUsage
func (c *BindingUsageSpecStorage) Upsert(bUsage *sbuTypes.ServiceBindingUsage, applied bool) error {
	cfg, err := c.cfgMapClient.Get(c.cfgMapName, metaV1.GetOptions{})
	if err != nil {
		return errors.Wrapf(err, "while getting config map with stored Spec for %s", pretty.ServiceBindingUsageName(bUsage))
	}

	storedSpec := UsageSpec{
		UsedBy:  bUsage.Spec.UsedBy,
		Applied: applied,
	}
	marshaledSpec, err := json.Marshal(storedSpec)
	if err != nil {
		return errors.Wrapf(err, "while marshaling Spec.UsedBy from %s", pretty.ServiceBindingUsageName(bUsage))
	}

	cfgCopy := cfg.DeepCopy()
	cfgKey := c.specUsedByKey(bUsage.Namespace, bUsage.Name)

	cfgCopy.Data = EnsureMapIsInitiated(cfgCopy.Data)
	cfgCopy.Data[cfgKey] = string(marshaledSpec)

	if _, err := c.cfgMapClient.Update(cfgCopy); err != nil {
		return errors.Wrapf(err, "while updating config map %q", c.cfgMapName)
	}

	return nil
}

func (c *BindingUsageSpecStorage) specUsedByKey(namespace, name string) string {
	return fmt.Sprintf("%s.%s.spec.usedBy", namespace, name)
}

// UsageSpec represents DTO which is used to store information about applied sbu in config map
type UsageSpec struct {
	UsedBy  sbuTypes.LocalReferenceByKindAndName
	Applied bool
}
