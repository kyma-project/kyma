package git

import (
	"io/ioutil"
	"strings"
	"time"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/pkg/errors"
)

type Client struct {
	url string
}

func New(repoURL string) Client {
	return Client{
		url: repoURL,
	}
}

func (c Client) TryCloning() error {
	_, _, err := c.cloneToMemory()
	return err
}

func (c Client) PullRemote(filePath string) (string, error) {
	_, fs, err := c.cloneToMemory()
	if err != nil {
		return "", errors.Wrap(err, "while cloning repository")
	}

	return c.readFileContent(fs, filePath)
}

func (c Client) ReplaceInRemoteFile(filePath, oldValue, newValue string) error {
	r, fs, err := c.cloneToMemory()
	if err != nil {
		return err
	}

	content, err := c.readFileContent(fs, filePath)
	if err != nil {
		return err
	}
	newContent := strings.Replace(content, oldValue, newValue, -1)

	err = c.replaceFileContent(fs, filePath, newContent)
	if err != nil {
		return err
	}

	err = c.commitAndPush(r, filePath)
	if err != nil {
		return err
	}

	return nil
}

func (c Client) cloneToMemory() (*git.Repository, billy.Filesystem, error) {
	fs := memfs.New()
	storer := memory.NewStorage()

	r, err := git.Clone(storer, fs, &git.CloneOptions{
		URL: c.url,
	})
	if err != nil {
		return nil, nil, errors.Wrap(err, "while cloning repository")
	}

	return r, fs, nil
}

func (c Client) readFileContent(fs billy.Filesystem, filePath string) (string, error) {
	file, err := fs.Open(filePath)
	if err != nil {
		return "", errors.Wrap(err, "while opening file from memory")
	}
	defer file.Close()

	content, err := ioutil.ReadAll(file)
	if err != nil {
		return "", errors.Wrap(err, "while reading file")
	}

	return string(content), nil
}

func (c Client) replaceFileContent(fs billy.Filesystem, filePath, newContent string) error {
	file, err := fs.Create(filePath)
	if err != nil {
		return errors.Wrap(err, "while creating file in memory")
	}

	_, err = file.Write([]byte(newContent))
	if err != nil {
		return errors.Wrap(err, "while writing new content to file in memory")
	}

	return nil
}

func (c Client) commitAndPush(repository *git.Repository, filePath string) error {
	w, err := repository.Worktree()
	if err != nil {
		return errors.Wrap(err, "while accessing repository worktree")
	}

	_, err = w.Add(filePath)
	if err != nil {
		return errors.Wrap(err, "while adding modified file to the staging area")
	}

	_, err = w.Commit("Replace values", &git.CommitOptions{
		All: false,
		Author: &object.Signature{
			Name:  "Chewbacca",
			Email: "chewbacca@kashyyyk.sw",
			When:  time.Time{},
		},
	})
	if err != nil {
		return errors.Wrap(err, "while committing changes")
	}

	err = repository.Push(&git.PushOptions{})
	if err != nil {
		return errors.Wrap(err, "while pushing changes")
	}

	return nil
}
