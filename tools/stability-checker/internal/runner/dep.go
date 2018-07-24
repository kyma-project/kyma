package runner

import (
	"k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type configMapClient interface {
	Create(*v1.ConfigMap) (*v1.ConfigMap, error)
	Update(*v1.ConfigMap) (*v1.ConfigMap, error)
	Get(name string, options metaV1.GetOptions) (*v1.ConfigMap, error)
}
