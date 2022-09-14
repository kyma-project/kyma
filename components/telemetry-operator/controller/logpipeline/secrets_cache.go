package logpipeline

import (
	"sync"

	"k8s.io/apimachinery/pkg/types"
)

type pipelineName string

type secretsCache struct {
	cache map[types.NamespacedName]map[pipelineName]struct{}
	mutex *sync.RWMutex
}

func newSecretsCache() secretsCache {
	return secretsCache{
		cache: make(map[types.NamespacedName]map[pipelineName]struct{}),
		mutex: &sync.RWMutex{},
	}
}

func (s *secretsCache) add(key types.NamespacedName, name pipelineName) {
	s.mutex.Lock()
	if _, found := s.cache[key]; !found {
		s.cache[key] = make(map[pipelineName]struct{})
	}
	s.cache[key][name] = struct{}{}
	s.mutex.Unlock()
}

func (s *secretsCache) get(key types.NamespacedName) map[pipelineName]struct{} {
	s.mutex.RLock()
	if pipelines, found := s.cache[key]; found {
		return pipelines
	}
	s.mutex.RUnlock()
	return nil
}

func (s *secretsCache) delete(key types.NamespacedName) {
	s.mutex.Lock()
	delete(s.cache, key)
	s.mutex.Unlock()
}
