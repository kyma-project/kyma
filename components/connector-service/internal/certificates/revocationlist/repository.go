package revocationlist

type RevocationListRepository interface {
	Insert(hash string) error
	Contains(hash string) (bool, error)
}

type revocationListRepository map[string]string

func NewRepository() RevocationListRepository {
	return &revocationListRepository{}
}

func (r revocationListRepository) Insert(hash string) error {
	r[hash] = hash
	return nil
}

func (r revocationListRepository) Contains(hash string) (bool, error) {
	_, ok := r[hash]
	return ok, nil
}