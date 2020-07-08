package git

import (
	"errors"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/kyma-project/kyma/components/function-controller/internal/git/automock"
	"github.com/onsi/gomega"
)

const (
	TmpPrefix = "manager-test-"
)

func TestLastCommit(t *testing.T) {
	exampleHash := plumbing.NewHash("649ed3e95dd9478785120a7572a71bdec2b0d660")

	for testName, testData := range map[string]struct {
		repoUrl    string
		repoBranch string
		repoAuth   transport.AuthMethod
		mockErr    error
		mockRefs   []*plumbing.Reference

		expectedCommit gomega.OmegaMatcher
		expectedErr    gomega.OmegaMatcher
	}{
		"should be ok": {
			repoUrl:    "https://github.com/kyma-project/kyma",
			repoBranch: "master",
			repoAuth:   nil,
			mockErr:    nil,
			mockRefs: []*plumbing.Reference{
				plumbing.NewHashReference("refs/heads/master", exampleHash),
			},

			expectedCommit: gomega.Equal(exampleHash.String()),
			expectedErr:    gomega.BeNil(),
		},
		"should be ok with auth": {
			repoUrl:    "https://github.com/kyma-project/kyma",
			repoBranch: "master",
			repoAuth: &http.BasicAuth{
				Username: "test",
				Password: "test",
			},
			mockErr: nil,
			mockRefs: []*plumbing.Reference{
				plumbing.NewHashReference("refs/heads/master", exampleHash),
			},

			expectedCommit: gomega.Equal(exampleHash.String()),
			expectedErr:    gomega.BeNil(),
		},
		"ok when on empty auth": {
			repoUrl:    "https://github.com/kyma-project/kyma",
			repoBranch: "master",
			repoAuth:   &http.BasicAuth{},
			mockErr:    nil,
			mockRefs: []*plumbing.Reference{
				plumbing.NewHashReference("refs/heads/master", exampleHash),
			},

			expectedCommit: gomega.Equal(exampleHash.String()),
			expectedErr:    gomega.BeNil(),
		},
		"ok with many refs in repo": {
			repoUrl:    "https://github.com/kyma-project/kyma",
			repoBranch: "master",
			repoAuth:   nil,
			mockErr:    nil,
			mockRefs: []*plumbing.Reference{
				plumbing.NewHashReference("refs/heads/test1", plumbing.NewHash("")),
				plumbing.NewHashReference("refs/heads/test2", plumbing.NewHash("")),
				plumbing.NewHashReference("refs/heads/master", exampleHash),
				plumbing.NewHashReference("refs/heads/test3", plumbing.NewHash("")),
			},

			expectedCommit: gomega.Equal(exampleHash.String()),
			expectedErr:    gomega.BeNil(),
		},
		"ok when ref don't provide commit hash": {
			repoUrl:    "https://github.com/kyma-project/kyma",
			repoBranch: "master",
			repoAuth:   nil,
			mockErr:    nil,
			mockRefs: []*plumbing.Reference{
				plumbing.NewHashReference("refs/heads/master", plumbing.NewHash("")),
			},

			expectedCommit: gomega.Equal(plumbing.NewHash("").String()),
			expectedErr:    gomega.BeNil(),
		},
		"error on no permissions to repo": {
			repoUrl:    "https://github.com/kyma-project/kyma",
			repoBranch: "master",
			repoAuth:   nil,
			mockErr:    errors.New("test error"),
			mockRefs:   nil,

			expectedCommit: gomega.HaveLen(0),
			expectedErr:    gomega.HaveOccurred(),
		},
		"error on no refs in repository": {
			repoUrl:    "https://github.com/kyma-project/kyma",
			repoBranch: "master",
			repoAuth:   nil,
			mockErr:    nil,
			mockRefs:   nil,

			expectedCommit: gomega.HaveLen(0),
			expectedErr:    gomega.HaveOccurred(),
		},
		"error on no expected ref in repository": {
			repoUrl:    "https://github.com/kyma-project/kyma",
			repoBranch: "master",
			repoAuth:   nil,
			mockErr:    nil,
			mockRefs: []*plumbing.Reference{
				plumbing.NewHashReference("refs/heads/test1", plumbing.NewHash("")),
				plumbing.NewHashReference("refs/heads/test2", plumbing.NewHash("")),
				plumbing.NewHashReference("refs/heads/test3", plumbing.NewHash("")),
			},

			expectedCommit: gomega.HaveLen(0),
			expectedErr:    gomega.HaveOccurred(),
		},
	} {
		t.Run(testName, func(t *testing.T) {
			// given
			g := gomega.NewWithT(t)

			client := new(automock.Client)
			git := Git{client: client}
			client.On("ListRefs", testData.repoUrl, testData.repoAuth).
				Return(testData.mockRefs, testData.mockErr).Once()

			// when
			hash, err := git.LastCommit(testData.repoUrl, testData.repoBranch, testData.repoAuth)

			//then
			g.Expect(hash).To(testData.expectedCommit)
			g.Expect(err).To(testData.expectedErr)
		})
	}
}

func TestClone(t *testing.T) {
	for testName, testData := range map[string]struct {
		repoUrl            string
		repoCommitOverride string
		repoAuth           transport.AuthMethod
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
			repoAuth: &http.BasicAuth{
				Username: "test",
				Password: "test",
			},
			commitsCount: 5,

			expectedErr:    gomega.BeNil(),
			expectedCommit: gomega.HaveLen(40),
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
			gitMock := new(automock.Client)
			if !testData.withoutRepo {
				repository, _, firstCommit = fixTmpRepository(g, tmpDir, testData.commitsCount, testData.bareRepo)
			}
			if testData.repoCommitOverride != "" {
				firstCommit = testData.repoCommitOverride
			}
			gitMock.On("PlainClone", tmpDir, false, &git.CloneOptions{
				URL:  testData.repoUrl,
				Auth: testData.repoAuth,
			}).
				Return(repository, testData.mockErr)
			operator := Git{client: gitMock}

			// when
			commit, err := operator.Clone(
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
