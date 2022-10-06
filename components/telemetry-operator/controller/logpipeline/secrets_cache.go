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

func (s *secretsCache) addOrUpdate(secretName types.NamespacedName, pipelineName string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if _, found := s.cache[secretName]; !found {
		s.cache[secretName] = make(map[string]struct{})
	}
	s.cache[secretName][pipelineName] = struct{}{}
}

func (s *secretsCache) get(secretName types.NamespacedName) []string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	if pipelines, found := s.cache[secretName]; found {
		return toSlice(pipelines)
	}
	return nil
}

func toSlice(from map[string]struct{}) []string {
	var result []string
	for k := range from {
		result = append(result, k)
	}
	return result
}

func (s *secretsCache) delete(secretName types.NamespacedName, pipelineName string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if pipelines, found := s.cache[secretName]; found {
		delete(pipelines, pipelineName)
	}
}
