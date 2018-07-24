package broker

func NewBindService(i instanceBindDataGetter) *bindService {
	return &bindService{
		instanceBindDataGetter: i,
	}
}
