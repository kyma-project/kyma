package director

type Client interface {
	RegisterApplication(name string) (string, error)
	UnregisterApplication(id string) error
}

type client struct {
	tenant string
}

func NewDirectorClient(tenant string) (Client, error) {
	return client{}, nil
}

func (c client) RegisterApplication(name string) (string, error) {
	return "", nil
}

func (c client) UnregisterApplication(id string) error {
	return nil
}
