package logpipeline

import (
	"sync"

	"k8s.io/apimachinery/pkg/types"
)

type pipelineName string

type secretsCache struct {
	contexts map[types.NamespacedName][]pipelineName
	mutex    *sync.RWMutex
}

func newSecretsCache() secretsCache {
	return secretsCache{
		contexts: make(map[types.NamespacedName][]pipelineName),
		mutex:    &sync.RWMutex{},
	}
}

func (s *secretsCache) add(key types.NamespacedName, name pipelineName) error {
	return nil
}

func (s *secretsCache) get(key types.NamespacedName) []pipelineName {
	return nil
}
