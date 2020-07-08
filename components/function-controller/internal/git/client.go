package git

import (
	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/storage/memory"
)

type client struct{}

func newClient() *client {
	return &client{}
}

func (o *client) ListRefs(repoUrl string, auth transport.AuthMethod) ([]*plumbing.Reference, error) {
	r := gogit.NewRemote(memory.NewStorage(), &config.RemoteConfig{
		URLs: []string{repoUrl},
	})

	return r.List(&gogit.ListOptions{
		Auth: auth,
	})
}

func (o *client) PlainClone(path string, isBare bool, options *gogit.CloneOptions) (*gogit.Repository, error) {
	return gogit.PlainClone(path, isBare, options)
}
