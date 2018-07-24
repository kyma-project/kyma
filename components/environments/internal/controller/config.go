package controller

import (
	"k8s.io/apimachinery/pkg/api/resource"
)

type EnvironmentsConfig struct {
	Namespace        string
	LimitRangeMemory LimitRangeConfig
	ResourceQuota    ResourceQuotaConfig
}

type FormattedQuantity int64

func (fq *FormattedQuantity) Unmarshal(value string) error {
	val, err := resource.ParseQuantity(value)
	*fq = FormattedQuantity(val.Value())
	return err
}

func (fq *FormattedQuantity) AsQuantity() *resource.Quantity {
	return resource.NewQuantity(int64(*fq), resource.BinarySI)
}

type ResourceQuotaConfig struct {
	LimitsMemory   FormattedQuantity
	RequestsMemory FormattedQuantity
}

type LimitRangeConfig struct {
	DefaultRequest FormattedQuantity
	Default        FormattedQuantity
	Max            FormattedQuantity
}
