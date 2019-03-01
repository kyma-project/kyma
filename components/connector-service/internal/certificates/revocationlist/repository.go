package revocationlist

import "sync"

type RevocationListRepository interface {
	Insert(hash string) error
	Contains(hash string) (bool, error)
}

type revocationListRepository struct {
	revocationList map[string]string
	mutex sync.RWMutex
}

func NewRepository() RevocationListRepository {
	return &revocationListRepository{
		make(map[string]string),
		sync.RWMutex{},
	}
}

func (r *revocationListRepository) Insert(hash string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.revocationList[hash] = hash
	return nil
}

func (r *revocationListRepository) Contains(hash string) (bool, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	_, ok := r.revocationList[hash]
	return ok, nil
}