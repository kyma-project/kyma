package git

import (
	"archive/tar"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	git2go "github.com/libgit2/git2go/v34"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

const (
	tarGitRepoPath = "./testdata"
	tarGitName     = "test-repo.tar"
	branchName     = "branch2"
	branchCommit   = "728c47705dabc65c12583ff5feb2e5300983afc3"
	tagName        = "tag1"
	trickyTagName  = "tricky1"
	tagCommit      = "6eff122e8afb57a6f270285dc3bfcc9a4ef4b8ad"
	secondCommitID = "8b27a9d6f148533773ae0666dc27c5b359b46553"

	azureRepo   = "https://kyma-wookiee@dev.azure.com/kyma-wookiee/kyma-function/_git/kyma-function"
	azureTag    = "python-tag"
	azureCommit = "6dac23dd3b697970cf351101ff5c3e9733c2bdfc"
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
			name:             "Success, tricky tag name from bitbucket",
			refName:          trickyTagName,
			expectedCommitID: branchCommit,
		},
		{
			name:             "Return error when not found",
			refName:          "testcase",
			expectedCommitID: "11111705dabc65c12583ff5feb2e5300983afc3",
			expectedError:    errors.New("while lookup tag: no reference found for shorthand 'testcase'"),
		},
	}

	for _, testcase := range testCases {
		t.Run(testcase.name, func(t *testing.T) {
			repoPath := prepareRepo(t)
			defer deleteTmpRepo(t, repoPath)
			fetcher := &git2goFetcherMock{repoPath: repoPath}

			opts := Options{Reference: testcase.refName, URL: repoPath}
			client := git2GoClient{fetcher: fetcher}
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

func TestNewGit2Go_LastCommitWithAzureTag(t *testing.T) {
	//GIVEN
	client := NewGit2Go(zap.L().Sugar())

	//WHEN
	commitID, err := client.LastCommit(Options{
		URL:       azureRepo,
		Reference: azureTag,
	})
	//THEN
	require.NoError(t, err)
	assert.Equal(t, azureCommit, commitID)
}

func TestGo2GitClient_Clone(t *testing.T) {
	//GIVEN
	repoPath := prepareRepo(t)
	defer deleteTmpRepo(t, repoPath)
	cloner := &git2goClonerMock{repoPath: repoPath}
	assertHeadCommitNotEqual(t, repoPath, secondCommitID)

	client := git2GoClient{cloner: cloner}
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

func (g *git2goClonerMock) git2goClone(url, outputPath string, remoteCallbacks git2go.RemoteCallbacks) (*git2go.Repository, error) {
	return git2go.OpenRepository(g.repoPath)
}

type git2goFetcherMock struct {
	repoPath string
}

func (g *git2goFetcherMock) git2goFetch(url, outputPath string, remoteCallbacks git2go.RemoteCallbacks) (*git2go.Repository, error) {
	return git2go.OpenRepository(g.repoPath)
}

func prepareRepo(t *testing.T) string {
	f, err := os.Open(filepath.Join(tarGitRepoPath, tarGitName))
	require.NoError(t, err)
	defer closeAssert(t, f.Close)
	r := tar.NewReader(f)
	for {
		h, err := r.Next()
		if err == io.EOF {
			break
		} else {
			require.NoError(t, err)
		}
		//nolint:gosec
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
		//nolint:gosec
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
	repo, err := cloner.git2goClone("", "", git2go.RemoteCallbacks{})
	require.NoError(t, err)
	head, err := repo.Head()
	require.NoError(t, err)
	assert.NotEqual(t, head.Target().String(), commit)
}
