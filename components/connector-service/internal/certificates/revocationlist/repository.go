package revocationlist

type RevocationListRepository interface {
	Insert(hash string) error
	Contains(hash string) (bool, error)
}

type revocationListRepository struct {

}

func NewRepository() RevocationListRepository {
	return &revocationListRepository{}
}

func (r revocationListRepository) Insert(hash string) error {
	return nil
}

func (r revocationListRepository) Contains(hash string) (bool, error) {
	return false, nil
}