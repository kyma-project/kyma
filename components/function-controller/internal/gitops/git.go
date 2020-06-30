package gitops

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/storage"
)

type Git struct{}

func NewGit() *Git {
	return &Git{}
}

func (o *Git) Clone(storage storage.Storer, worktree billy.Filesystem, options *git.CloneOptions) (*git.Repository, error) {
	return git.Clone(storage, worktree, options)
}

func (o *Git) PlainClone(path string, isBare bool, options *git.CloneOptions) (*git.Repository, error) {
	return git.PlainClone(path, isBare, options)
}
