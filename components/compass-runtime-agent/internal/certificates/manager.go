package certificates

type Manager interface {
	GetCredentials() (Credentials, error)
	PreserveCredentials(credentials Credentials) error
}

func NewCerdentialsManager() *credentialsManager {
	return &credentialsManager{}
}

type credentialsManager struct{}

func (*credentialsManager) GetCredentials() (Credentials, error) {
	// TODO: provide implementation after the server side is ready

	return Credentials{}, nil
}

func (*credentialsManager) PreserveCredentials(credentials Credentials) error {
	// TODO: provide implementation after the server side is ready

	return nil
}
