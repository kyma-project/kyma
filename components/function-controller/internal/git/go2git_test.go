package git

import (
	"archive/tar"
	"errors"
	git2go "github.com/libgit2/git2go/v31"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"os"
	"path/filepath"
	"testing"
)

const (
	tarGitRepoPath = "./testdata"
	tarGitName     = "test-repo.tar"
	repoName       = "test-repo"
	branchName     = "branch2"
	branchCommit   = "728c47705dabc65c12583ff5feb2e5300983afc3"
	tagName        = "tag1"
	tagCommit      = "6eff122e8afb57a6f270285dc3bfcc9a4ef4b8ad"
	secondCommitID = "8b27a9d6f148533773ae0666dc27c5b359b46553"
)

func TestNewGit2Go_LastCommit(t *testing.T) {
	//GIVEN
	testCases := []struct {
		name             string
		refName          string
		expectedCommitID string
		expectedError    error
	}{
		{
			name:             "Success branch name",
			refName:          branchName,
			expectedCommitID: branchCommit,
		},
		{
			name:             "Success tag name",
			refName:          tagName,
			expectedCommitID: tagCommit,
		},
		{
			name:             "Success commit",
			refName:          tagCommit,
			expectedCommitID: tagCommit,
		},
		{
			name:             "Return error when not found",
			refName:          "testcase",
			expectedCommitID: "11111705dabc65c12583ff5feb2e5300983afc3",
			expectedError:    errors.New("Could find commit,branch or tag with given ref: testcase"),
		},
	}

	for _, testcase := range testCases {
		t.Run(testcase.name, func(t *testing.T) {
			repoPath := prepareRepo(t)
			defer deleteTmpRepo(t, repoPath)
			cloner := &git2goClonerMock{repoPath: repoPath}

			opts := Options{Reference: testcase.refName}
			client := Git2GoClient{cloner}
			//WHEN
			commitID, err := client.LastCommit(opts)

			//THEN
			if testcase.expectedError == nil {
				require.NoError(t, err)
				assert.Equal(t, testcase.expectedCommitID, commitID)
			} else {
				require.Error(t, err)
				assert.EqualError(t, testcase.expectedError, err.Error())
			}
		})
	}
}

func TestGo2GitClient_Clone(t *testing.T) {
	//GIVEN
	repoPath := prepareRepo(t)
	defer deleteTmpRepo(t, repoPath)
	cloner := &git2goClonerMock{repoPath: repoPath}
	assertHeadCommitNotEqual(t, repoPath, secondCommitID)

	client := Git2GoClient{cloner}
	opts := Options{Reference: secondCommitID}
	//WHEN
	commitID, err := client.Clone("", opts)

	// THEN
	require.NoError(t, err)
	assert.Equal(t, secondCommitID, commitID)
}

type git2goClonerMock struct {
	repoPath string
}

func (g *git2goClonerMock) cloneRepo(options Options, outputPath string) (*git2go.Repository, error) {
	return git2go.OpenRepository(g.repoPath)

}

func prepareRepo(t *testing.T) string {
	f, err := os.Open(filepath.Join(tarGitRepoPath, tarGitName))
	require.NoError(t, err)
	defer closeAssert(t, f.Close)
	r := tar.NewReader(f)
	for ; ; {
		h, err := r.Next()
		if err == io.EOF {
			break
		} else {
			require.NoError(t, err)
		}
		path := filepath.Join(tarGitRepoPath, h.Name)
		info := h.FileInfo()
		if info.IsDir() {
			err = os.Mkdir(path, info.Mode())
			require.NoError(t, err)
			continue
		}
		f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, info.Mode())
		require.NoError(t, err)
		defer closeAssert(t, f.Close)
		_, err = io.Copy(f, r)
		require.NoError(t, err)
	}

	return filepath.Join(tarGitRepoPath, "test-repo")
}

func closeAssert(t *testing.T, fn func() error) {
	require.NoError(t, fn())
}

func deleteTmpRepo(t *testing.T, tmpPath string) {
	err := os.RemoveAll(tmpPath)
	require.NoError(t, err)
}

func assertHeadCommitNotEqual(t *testing.T, repoPath, commit string) {
	cloner := &git2goClonerMock{repoPath: repoPath}
	repo, err := cloner.cloneRepo(Options{}, "")
	require.NoError(t, err)
	head, err := repo.Head()
	require.NoError(t, err)
	assert.NotEqual(t, head.Target().String(), commit)
}
