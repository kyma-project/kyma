package gitops

import (
	"errors"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/kyma-project/kyma/components/function-controller/internal/gitops/automock"
	"github.com/onsi/gomega"
)

const (
	TmpPrefix = "manager-test-"
)

func TestGetLastCommit(t *testing.T) {
	for testName, testData := range map[string]struct {
		config       Config
		mockErr      error
		withoutRepo  bool
		commitsCount int

		expectedCommit gomega.OmegaMatcher
		expectedErr    gomega.OmegaMatcher
	}{
		"should be ok": {
			config: Config{
				RepoUrl: "https://github.com/kyma-project/kyma",
				Branch:  "master",
				BaseDir: "",
				Secret:  nil,
			},
			commitsCount:   1,
			expectedCommit: gomega.HaveLen(40),
			expectedErr:    gomega.BeNil(),
		},
		"should be ok with auth": {
			config: Config{
				RepoUrl: "https://github.com/kyma-project/kyma",
				Branch:  "master",
				BaseDir: "",
				Secret:  map[string]string{usernameKey: "test", passwordKey: "test"},
			},
			commitsCount:   1,
			expectedCommit: gomega.HaveLen(40),
			expectedErr:    gomega.BeNil(),
		},
		"ok with many commits in repo": {
			config: Config{
				RepoUrl: "https://github.com/kyma-project/kyma",
				Branch:  "master",
				BaseDir: "",
				Secret:  nil,
			},
			commitsCount:   10,
			expectedCommit: gomega.HaveLen(40),
			expectedErr:    gomega.BeNil(),
		},
		"error on empty auth map": {
			config: Config{
				RepoUrl: "https://github.com/kyma-project/kyma",
				Branch:  "master",
				BaseDir: "",
				Secret:  map[string]string{},
			},
			expectedCommit: gomega.HaveLen(0),
			expectedErr:    gomega.HaveOccurred(),
		},
		"error when incomplete auth map": {
			config: Config{
				RepoUrl: "https://github.com/kyma-project/kyma",
				Branch:  "master",
				BaseDir: "",
				Secret:  map[string]string{usernameKey: "test", "test": "test", "test-2": "test"},
			},
			expectedCommit: gomega.HaveLen(0),
			expectedErr:    gomega.HaveOccurred(),
		},
		"error when cloning repository": {
			config: Config{
				RepoUrl: "https://github.com/kyma-project/kyma",
				Branch:  "master",
				BaseDir: "",
				Secret:  nil,
			},
			mockErr:        errors.New("test error"),
			withoutRepo:    true,
			expectedCommit: gomega.HaveLen(0),
			expectedErr:    gomega.HaveOccurred(),
		},
		"error when getting HEAD": {
			config: Config{
				RepoUrl: "https://github.com/kyma-project/kyma",
				Branch:  "master",
				BaseDir: "",
				Secret:  nil,
			},
			expectedCommit: gomega.HaveLen(0),
			expectedErr:    gomega.HaveOccurred(),
		},
	} {
		t.Run(testName, func(t *testing.T) {
			// given
			g := gomega.NewWithT(t)

			tmpDir, _ := ioutil.TempDir(os.TempDir(), TmpPrefix)
			defer os.RemoveAll(tmpDir)

			git := new(automock.GitInterface)
			operator := Operator{gitInterface: git}
			lastCommit := ""
			call := git.On("Clone", memory.NewStorage(), nil,
				fixCloneOptions(testData.config.RepoUrl, testData.config.Branch, testData.config.Secret))

			if testData.withoutRepo {
				call.Return(nil, testData.mockErr).Once()
			} else {
				repo, commit, _ := fixTmpRepository(g, tmpDir, testData.commitsCount, false)
				lastCommit = commit
				call.Return(repo, testData.mockErr).Once()
			}

			// when
			hash, err := operator.GetLastCommit(testData.config)

			//then
			g.Expect(hash).To(testData.expectedCommit)
			if hash != "" {
				g.Expect(hash).To(gomega.Equal(lastCommit))
			}
			g.Expect(err).To(testData.expectedErr)
		})
	}
}

func TestCloneRepoFromCommit(t *testing.T) {
	for testName, testData := range map[string]struct {
		repoUrl            string
		repoAuth           map[string]string
		repoCommitOverride string
		mockErr            error
		withoutRepo        bool
		bareRepo           bool
		commitsCount       int

		expectedCommit gomega.OmegaMatcher
		expectedErr    gomega.OmegaMatcher
	}{
		"should be ok": {
			repoUrl:      "https://github.com/kyma-project/kyma",
			repoAuth:     nil,
			commitsCount: 5,

			expectedErr:    gomega.BeNil(),
			expectedCommit: gomega.HaveLen(40),
		},
		"should be ok with auth": {
			repoUrl: "https://github.com/kyma-project/kyma",
			repoAuth: map[string]string{
				usernameKey: "test",
				passwordKey: "test",
			},
			commitsCount: 5,

			expectedErr:    gomega.BeNil(),
			expectedCommit: gomega.HaveLen(40),
		},
		"error on auth parsing": {
			repoUrl: "https://github.com/kyma-project/kyma",
			repoAuth: map[string]string{
				usernameKey: "test",
			},
			commitsCount: 5,

			expectedErr:    gomega.HaveOccurred(),
			expectedCommit: gomega.HaveLen(0),
		},
		"error when repo don't exist": {
			repoUrl:     "https://github.com/kyma-project/kyma",
			repoAuth:    nil,
			withoutRepo: true,
			mockErr:     errors.New("test"),

			expectedErr:    gomega.HaveOccurred(),
			expectedCommit: gomega.HaveLen(0),
		},
		"error when worktree don't exist": {
			repoUrl:  "https://github.com/kyma-project/kyma",
			repoAuth: nil,
			bareRepo: true,

			expectedErr:    gomega.HaveOccurred(),
			expectedCommit: gomega.HaveLen(0),
		},
		"error when checkout to wrong commit": {
			repoUrl:            "https://github.com/kyma-project/kyma",
			repoAuth:           nil,
			repoCommitOverride: "123",
			commitsCount:       5,

			expectedErr:    gomega.HaveOccurred(),
			expectedCommit: gomega.HaveLen(0),
		},
	} {
		t.Run(testName, func(t *testing.T) {
			// given
			g := gomega.NewWithT(t)

			tmpDir, _ := ioutil.TempDir(os.TempDir(), TmpPrefix)
			defer os.RemoveAll(tmpDir)

			firstCommit := ""
			var repository *git.Repository
			gitMock := new(automock.GitInterface)
			if !testData.withoutRepo {
				repository, _, firstCommit = fixTmpRepository(g, tmpDir, testData.commitsCount, testData.bareRepo)
			}
			if testData.repoCommitOverride != "" {
				firstCommit = testData.repoCommitOverride
			}
			gitMock.On("PlainClone", tmpDir, false, fixPlainCloneOptions(testData.repoUrl, testData.repoAuth)).
				Return(repository, testData.mockErr)
			operator := Operator{gitInterface: gitMock}

			// when
			commit, err := operator.CloneRepoFromCommit(
				tmpDir,
				testData.repoUrl,
				firstCommit,
				testData.repoAuth,
			)

			// then
			g.Expect(commit).To(testData.expectedCommit)
			if commit != "" {
				g.Expect(commit).To(gomega.Equal(firstCommit))
			}
			g.Expect(err).To(testData.expectedErr)
		})
	}
}

func TestConvertToMap(t *testing.T) {
	for testName, testData := range map[string]struct {
		username string
		password string

		expectedMap gomega.OmegaMatcher
	}{
		"should be ok": {
			username: "test",
			password: "test",
			expectedMap: gomega.And(
				gomega.HaveLen(2),
				gomega.HaveKeyWithValue(usernameKey, "test"),
				gomega.HaveKeyWithValue(passwordKey, "test"),
			),
		},
		"should return nil map when empty strings are given": {
			expectedMap: gomega.BeNil(),
		},
		"should return nil map when only username is given": {
			username:    "test",
			expectedMap: gomega.BeNil(),
		},
		"should return nil map when only password is given": {
			password:    "test",
			expectedMap: gomega.BeNil(),
		},
	} {
		t.Run(testName, func(t *testing.T) {
			// given
			g := gomega.NewWithT(t)
			operator := NewOperator()

			// when
			auth := operator.ConvertToMap(testData.username, testData.password)

			// then
			g.Expect(auth).To(testData.expectedMap)
		})
	}
}

func fixPlainCloneOptions(url string, auth map[string]string) *git.CloneOptions {
	basicAuth := fixBasicAuth(auth)
	return &git.CloneOptions{
		URL:  url,
		Auth: basicAuth,
	}
}

func fixCloneOptions(url, branch string, auth map[string]string) *git.CloneOptions {
	basicAuth := fixBasicAuth(auth)
	return &git.CloneOptions{
		URL:           url,
		ReferenceName: plumbing.NewBranchReferenceName(branch),
		Auth:          basicAuth,
		SingleBranch:  true,
		NoCheckout:    true,
		Tags:          git.NoTags,
	}
}

func fixBasicAuth(auth map[string]string) *http.BasicAuth {
	basicAuth := &http.BasicAuth{}
	if auth != nil {
		username := ""
		password := ""
		if key, ok := auth[usernameKey]; ok {
			username = key
		}
		if key, ok := auth[passwordKey]; ok {
			password = key
		}
		basicAuth = &http.BasicAuth{
			Username: username,
			Password: password,
		}
	}
	return basicAuth
}

func fixTmpRepository(g *gomega.WithT, dirPath string, commitsCount int, isBare bool) (*git.Repository, string, string) {
	repo, initErr := git.PlainInit(dirPath, isBare)
	g.Expect(initErr).To(gomega.BeNil())

	author := &object.Signature{
		Name:  "test",
		Email: "test@test.test",
		When:  time.Now(),
	}

	var commit string
	var firstCommit string
	for i := 0; i < commitsCount; i++ {
		tree, treeErr := repo.Worktree()
		g.Expect(treeErr).To(gomega.BeNil())

		hash, commitErr := tree.Commit("test message", &git.CommitOptions{Author: author})
		commit = hash.String()
		g.Expect(commitErr).To(gomega.BeNil())
		if i == 0 {
			firstCommit = commit
		}
	}

	return repo, commit, firstCommit
}
