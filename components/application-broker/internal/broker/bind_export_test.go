package broker

import "github.com/kyma-project/kyma/components/application-broker/internal"

func NewBindService(reFinder reFinder) *bindService {
	return &bindService{
		reSvcFinder: reFinder,
	}
}

func (svc *bindService) GetCredentials(id internal.RemoteServiceID, re *internal.RemoteEnvironment) (map[string]interface{}, error) {
	return svc.getCredentials(id, re)
}
