// +build integration

package integration_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	urlhelper "github.com/hashicorp/go-getter/helper/url"
)

const (
	addonSource = "testdata/addons.tgz"

	repositoryDirName   = "fake-repo"
	repositoryDirPrefix = "fake-"
)

func newGitRepository(t *testing.T, source string) (*gitRepo, error) {
	repo, err := testGitRepo(t, repositoryDirName)
	if err != nil {
		return nil, err
	}

	err = repo.fillRepo(source)
	if err != nil {
		return nil, err
	}

	return repo, nil
}

type gitRepo struct {
	t   *testing.T
	url *url.URL
	dir string
}

func testGitRepo(t *testing.T, name string) (*gitRepo, error) {
	dir, err := ioutil.TempDir("", repositoryDirPrefix)
	if err != nil {
		return nil, err
	}
	dir = filepath.Join(dir, name)
	if err := os.Mkdir(dir, 0700); err != nil {
		return nil, err
	}

	r := &gitRepo{
		t:   t,
		dir: dir,
	}

	gitURL, err := urlhelper.Parse("file://" + r.dir)
	if err != nil {
		return nil, err
	}
	r.url = gitURL

	t.Logf("initializing git repo in %s", dir)
	r.git("init")
	r.git("config", "user.name", "fake-client")
	r.git("config", "user.email", "fake@kyma-project.io")

	return r, nil
}

func (r *gitRepo) fillRepo(name string) error {
	f, err := os.Open(name)
	if err != nil {
		return err
	}
	defer f.Close()

	uncompressedStream, err := gzip.NewReader(f)
	if err != nil {
		return err
	}

	tarReader := tar.NewReader(uncompressedStream)
	for true {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		switch header.Typeflag {
		case tar.TypeDir:
			path := filepath.Join(r.dir, header.Name)
			if err := os.Mkdir(path, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			path := filepath.Join(r.dir, header.Name)
			outFile, err := os.Create(path)
			if err != nil {
				return err
			}
			defer outFile.Close()
			if _, err := io.Copy(outFile, tarReader); err != nil {
				return err
			}
			r.git("add", path)
			r.git("commit", "-m", "Adding "+header.Name)
			r.t.Logf("File %s added to repository", header.Name)
		default:
			r.t.Logf("ExtractTarGz: uknown type: %v in %s", header.Typeflag, header.Name)
		}
	}

	return nil
}

func (r *gitRepo) git(args ...string) {
	cmd := exec.Command("git", args...)
	cmd.Dir = r.dir
	bfr := bytes.NewBuffer(nil)
	cmd.Stderr = bfr
	if err := cmd.Run(); err != nil {
		r.t.Fatal(err, bfr.String())
	}
}

func (r *gitRepo) path(index string) string {
	return fmt.Sprintf("%s//addons/%s", r.url.String(), index)
}

func (r *gitRepo) removeTmpDir() {
	elements := strings.Split(r.dir, "/")
	err := os.RemoveAll(strings.Join(elements[:len(elements)-1], "/"))
	if err != nil {
		r.t.Logf("failed in defer, cannot remove temporary git reposiotory: %s", err)
	}
	r.t.Logf("Temp dir with repository %q was removed", r.dir)
}
