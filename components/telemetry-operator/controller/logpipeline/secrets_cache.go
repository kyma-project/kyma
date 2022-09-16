package logpipeline

import (
	"sync"

	"k8s.io/apimachinery/pkg/types"
)

type secretsCache struct {
	cache map[types.NamespacedName]map[string]struct{}
	mutex sync.RWMutex
}

func newSecretsCache() secretsCache {
	return secretsCache{
		cache: make(map[types.NamespacedName]map[string]struct{}),
		mutex: sync.RWMutex{},
	}
}

func (s *secretsCache) set(key types.NamespacedName, name string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if _, found := s.cache[key]; !found {
		s.cache[key] = make(map[string]struct{})
	}
	s.cache[key][name] = struct{}{}
}

func (s *secretsCache) get(key types.NamespacedName) map[string]struct{} {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	if pipelines, found := s.cache[key]; found {
		return pipelines
	}
	return nil
}

func (s *secretsCache) delete(key types.NamespacedName, name string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if pipelines, found := s.cache[key]; found {
		delete(pipelines, name)
	}
}
