package broker

func NewInstanceStateService(ocg operationCollectionGetter) *instanceStateService {
	return &instanceStateService{
		operationCollectionGetter: ocg,
	}
}
