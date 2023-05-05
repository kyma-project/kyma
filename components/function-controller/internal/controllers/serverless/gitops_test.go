package serverless

import (
	"errors"
	"strings"
	"testing"

	git2go "github.com/libgit2/git2go/v34"
	"github.com/stretchr/testify/assert"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha2"
	"github.com/onsi/gomega"
)

func Test_isOnSourceChange(t *testing.T) {
	testCases := []struct {
		desc           string
		fn             v1alpha2.Function
		revision       string
		expectedResult bool
	}{
		{
			desc: "new function",
			fn: v1alpha2.Function{
				Spec: v1alpha2.FunctionSpec{
					Source: v1alpha2.Source{
						GitRepository: &v1alpha2.GitRepositorySource{},
					},
					Runtime: v1alpha2.NodeJs18,
				},
			},
			expectedResult: true,
		},
		{
			desc: "new function fixed on commit",
			fn: v1alpha2.Function{
				Spec: v1alpha2.FunctionSpec{
					Source: v1alpha2.Source{
						GitRepository: &v1alpha2.GitRepositorySource{
							Repository: v1alpha2.Repository{
								Reference: "1",
							},
						},
					},
					Runtime: v1alpha2.NodeJs18,
				},
			},
			expectedResult: true,
		},
		{
			desc: "new function follow head",
			fn: v1alpha2.Function{
				Spec: v1alpha2.FunctionSpec{
					Source: v1alpha2.Source{
						GitRepository: &v1alpha2.GitRepositorySource{
							Repository: v1alpha2.Repository{
								Reference: "1",
							},
						},
					},
					Runtime: v1alpha2.NodeJs18,
				},
			},
			expectedResult: true,
		},
		{
			desc: "function did not change",
			fn: v1alpha2.Function{
				Spec: v1alpha2.FunctionSpec{
					Source: v1alpha2.Source{
						GitRepository: &v1alpha2.GitRepositorySource{
							Repository: v1alpha2.Repository{
								Reference: "1",
							},
						},
					},
					Runtime: v1alpha2.NodeJs18,
				},
				Status: v1alpha2.FunctionStatus{
					Repository: v1alpha2.Repository{
						Reference: "1",
					},
					Commit:  "1",
					Runtime: v1alpha2.NodeJs18,
				},
			},
			revision:       "1",
			expectedResult: false,
		},
		{
			desc: "function change fixed revision",
			fn: v1alpha2.Function{
				Spec: v1alpha2.FunctionSpec{
					Source: v1alpha2.Source{
						GitRepository: &v1alpha2.GitRepositorySource{
							Repository: v1alpha2.Repository{
								Reference: "2",
							},
						},
					},
					Runtime: v1alpha2.NodeJs18,
				},
				Status: v1alpha2.FunctionStatus{
					Repository: v1alpha2.Repository{
						Reference: "1",
					},
				},
			},
			expectedResult: true,
		},
		{
			desc: "function change",
			fn: v1alpha2.Function{
				Status: v1alpha2.FunctionStatus{
					Repository: v1alpha2.Repository{
						Reference: "1",
					},
				},
			},
			revision:       "2",
			expectedResult: true,
		},
		{
			desc: "function change base dir",
			fn: v1alpha2.Function{
				Spec: v1alpha2.FunctionSpec{
					Source: v1alpha2.Source{
						GitRepository: &v1alpha2.GitRepositorySource{
							Repository: v1alpha2.Repository{
								Reference: "2",
								BaseDir:   "base_dir",
							},
						},
					},
				},
				Status: v1alpha2.FunctionStatus{
					Repository: v1alpha2.Repository{
						Reference: "2",
					},
				},
			},
			expectedResult: true,
		},
		{
			desc: "function change branch",
			fn: v1alpha2.Function{
				Spec: v1alpha2.FunctionSpec{
					Source: v1alpha2.Source{
						GitRepository: &v1alpha2.GitRepositorySource{
							Repository: v1alpha2.Repository{
								Reference: "branch",
							},
						},
					},
				},
				Status: v1alpha2.FunctionStatus{
					Repository: v1alpha2.Repository{
						Reference: "2",
					},
				},
			},
			expectedResult: true,
		},
		{
			desc: "function change dockerfile",
			fn: v1alpha2.Function{
				Spec: v1alpha2.FunctionSpec{
					Source: v1alpha2.Source{
						GitRepository: &v1alpha2.GitRepositorySource{
							Repository: v1alpha2.Repository{
								Reference: "2",
							},
						},
					},
					Runtime: v1alpha2.NodeJs18,
				},
				Status: v1alpha2.FunctionStatus{
					Repository: v1alpha2.Repository{
						Reference: "2",
					},
				},
			},
			expectedResult: true,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			g := gomega.NewGomegaWithT(t)
			s := systemState{
				instance: tC.fn,
			}
			actual := s.gitFnSrcChanged(tC.revision)
			g.Expect(actual).To(gomega.Equal(tC.expectedResult))
		})
	}
}

func TestNextRequeue(t *testing.T) {
	//GIVEN
	testCases := []struct {
		name           string
		inputErr       error
		expectedErrMsg string
		expectedResult ctrl.Result
	}{
		{
			name:           "Git unrecoverable error",
			inputErr:       git2go.MakeGitError2(int(git2go.ErrorCodeNotFound)),
			expectedErrMsg: "Stop reconciliation, reason:",
			expectedResult: ctrl.Result{Requeue: false},
		},
		{
			name:           "Git authorization error",
			inputErr:       errors.New("unexpected http status code: 403"),
			expectedErrMsg: "Authorization to git server failed",
			expectedResult: ctrl.Result{Requeue: true},
		}, {
			name:           "Git generic error",
			inputErr:       git2go.MakeGitError2(int(git2go.ErrAmbiguous)),
			expectedErrMsg: "Sources update failed, reason:",
			expectedResult: ctrl.Result{Requeue: true},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			//WHEN
			result, errMsg := NextRequeue(testCase.inputErr)

			//THEN
			assert.Equal(t, testCase.expectedResult, result)
			assert.NotEmpty(t, testCase.expectedErrMsg)
			assert.True(t, strings.HasPrefix(errMsg, testCase.expectedErrMsg), "errMsg: %s, doesn't start with: %s", errMsg, testCase.expectedErrMsg)
		})
	}
}
