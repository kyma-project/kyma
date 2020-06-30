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
		expectedIsOk   gomega.OmegaMatcher
		expectedErr    gomega.OmegaMatcher
	}{
		"should be ok": {
			config: Config{
				RepoUrl:      "https://github.com/kyma-project/kyma",
				Branch:       "master",
				ActualCommit: "1234",
				BaseDir:      "",
				Secret:       nil,
			},
			commitsCount:   1,
			expectedCommit: gomega.HaveLen(40),
			expectedIsOk:   gomega.BeTrue(),
			expectedErr:    gomega.BeNil(),
		},
		"ok with many commits in repo": {
			config: Config{
				RepoUrl:      "https://github.com/kyma-project/kyma",
				Branch:       "master",
				ActualCommit: "1234",
				BaseDir:      "",
				Secret:       nil,
			},
			commitsCount:   10,
			expectedCommit: gomega.HaveLen(40),
			expectedIsOk:   gomega.BeTrue(),
			expectedErr:    gomega.BeNil(),
		},
		"error on empty auth map": {
			config: Config{
				RepoUrl:      "https://github.com/kyma-project/kyma",
				Branch:       "master",
				ActualCommit: "1234",
				BaseDir:      "",
				Secret:       map[string]interface{}{},
			},
			expectedCommit: gomega.HaveLen(0),
			expectedIsOk:   gomega.BeFalse(),
			expectedErr:    gomega.HaveOccurred(),
		},
		"error on incomplete auth map": {
			config: Config{
				RepoUrl:      "https://github.com/kyma-project/kyma",
				Branch:       "master",
				ActualCommit: "1234",
				BaseDir:      "",
				Secret:       map[string]interface{}{"login": "test", "test": "test", "test-2": "test"},
			},
			expectedCommit: gomega.HaveLen(0),
			expectedIsOk:   gomega.BeFalse(),
			expectedErr:    gomega.HaveOccurred(),
		},
		"error on cloning repository": {
			config: Config{
				RepoUrl:      "https://github.com/kyma-project/kyma",
				Branch:       "master",
				ActualCommit: "1234",
				BaseDir:      "",
				Secret:       nil,
			},
			mockErr:        errors.New("test error"),
			withoutRepo:    true,
			expectedCommit: gomega.HaveLen(0),
			expectedIsOk:   gomega.BeFalse(),
			expectedErr:    gomega.HaveOccurred(),
		},
		"error on getting HEAD": {
			config: Config{
				RepoUrl:      "https://github.com/kyma-project/kyma",
				Branch:       "master",
				ActualCommit: "1234",
				BaseDir:      "",
				Secret:       nil,
			},
			expectedCommit: gomega.HaveLen(0),
			expectedIsOk:   gomega.BeFalse(),
			expectedErr:    gomega.HaveOccurred(),
		},
	} {
		t.Run(testName, func(t *testing.T) {
			// given
			tmpDir, _ := ioutil.TempDir(os.TempDir(), TmpPrefix)
			defer os.RemoveAll(tmpDir)

			git := &automock.GitInterface{}
			operator := Operator{gitOperator: git}
			lastCommit := ""
			g := gomega.NewWithT(t)
			call := git.On("Clone", memory.NewStorage(), nil,
				fixCloneOptions(testData.config.RepoUrl, testData.config.Branch, testData.config.Secret))

			if testData.withoutRepo {
				call.Return(nil, testData.mockErr).Once()
			} else {
				repo, commit := fixTmpRepository(g, tmpDir, testData.commitsCount)
				lastCommit = commit
				call.Return(repo, testData.mockErr).Once()
			}

			// when
			hash, isOk, err := operator.CheckBranchChanges(testData.config)

			//then
			g.Expect(hash).To(testData.expectedCommit)
			if hash != "" {
				g.Expect(hash).To(gomega.Equal(lastCommit))
			}
			g.Expect(isOk).To(testData.expectedIsOk)
			g.Expect(err).To(testData.expectedErr)
		})
	}
}

func fixCloneOptions(url, branch string, auth map[string]interface{}) *git.CloneOptions {
	basicAuth := &http.BasicAuth{}
	if auth != nil {
		username := ""
		password := ""
		if key, ok := auth[usernameKey]; ok {
			username = key.(string)
		}
		if key, ok := auth[passwordKey]; ok {
			password = key.(string)
		}
		basicAuth = &http.BasicAuth{
			Username: username,
			Password: password,
		}
	}
	return &git.CloneOptions{
		URL:           url,
		ReferenceName: plumbing.NewBranchReferenceName(branch),
		Auth:          basicAuth,
		SingleBranch:  true,
		NoCheckout:    true,
		Tags:          git.NoTags,
	}
}

func fixTmpRepository(g *gomega.WithT, dirPath string, commitCount int) (*git.Repository, string) {
	repo, initErr := git.PlainInit(dirPath, false)
	g.Expect(initErr).To(gomega.BeNil())
	tree, treeErr := repo.Worktree()
	g.Expect(treeErr).To(gomega.BeNil())

	author := &object.Signature{
		Name:  "test",
		Email: "test@test.test",
		When:  time.Now(),
	}

	var commit string
	for i := 0; i < commitCount; i++ {
		hash, commitErr := tree.Commit("test message", &git.CommitOptions{Author: author})
		commit = hash.String()
		g.Expect(commitErr).To(gomega.BeNil())
	}

	return repo, commit
}
