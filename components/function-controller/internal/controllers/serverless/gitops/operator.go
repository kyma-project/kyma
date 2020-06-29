package gitops

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/storage"
)

type operator struct {}

func NewOperator() *operator {
	return &operator{}
}

func (o *operator) Clone(storage storage.Storer, worktree billy.Filesystem, options *git.CloneOptions) (*git.Repository, error){
	return git.Clone(storage, worktree, options)
}
