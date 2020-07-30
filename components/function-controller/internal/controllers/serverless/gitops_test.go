package serverless

import (
	"testing"

	"github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"github.com/onsi/gomega"
)

func Test_isOnSourceChange(t *testing.T) {
	testCases := []struct {
		desc           string
		fn             v1alpha1.Function
		revision       string
		expectedResult bool
	}{
		{
			desc: "new function",
			fn: v1alpha1.Function{
				Spec: v1alpha1.FunctionSpec{
					SourceType: v1alpha1.SourceTypeGit,
				},
			},
			expectedResult: true,
		},
		{
			desc: "new function fixed on commit",
			fn: v1alpha1.Function{
				Spec: v1alpha1.FunctionSpec{
					SourceType: v1alpha1.SourceTypeGit,
					Repository: v1alpha1.Repository{
						Commit: "1",
					},
				},
			},
			expectedResult: true,
		},
		{
			desc: "new function follow head",
			fn: v1alpha1.Function{
				Spec: v1alpha1.FunctionSpec{
					SourceType: v1alpha1.SourceTypeGit,
					Repository: v1alpha1.Repository{
						Commit: "1",
					},
				},
			},
			expectedResult: true,
		},
		{
			desc: "function did not change",
			fn: v1alpha1.Function{
				Spec: v1alpha1.FunctionSpec{
					SourceType: v1alpha1.SourceTypeGit,
					Repository: v1alpha1.Repository{
						Commit: "1",
					},
				},
				Status: v1alpha1.FunctionStatus{
					Repository: v1alpha1.Repository{
						Commit: "1",
					},
				},
			},
			revision:       "1",
			expectedResult: false,
		},
		{
			desc: "function change fixed revision",
			fn: v1alpha1.Function{
				Spec: v1alpha1.FunctionSpec{
					SourceType: v1alpha1.SourceTypeGit,
					Repository: v1alpha1.Repository{
						Commit: "2",
					},
				},
				Status: v1alpha1.FunctionStatus{
					Repository: v1alpha1.Repository{
						Commit: "1",
					},
				},
			},
			expectedResult: true,
		},
		{
			desc: "function change",
			fn: v1alpha1.Function{
				Status: v1alpha1.FunctionStatus{
					Repository: v1alpha1.Repository{
						Commit: "1",
					},
				},
			},
			revision:       "2",
			expectedResult: true,
		},
		{
			desc: "function change source",
			fn: v1alpha1.Function{
				Spec: v1alpha1.FunctionSpec{
					SourceType: v1alpha1.SourceTypeGit,
					Repository: v1alpha1.Repository{
						Commit: "1",
					},
					Source: "new_src",
				},
				Status: v1alpha1.FunctionStatus{
					Repository: v1alpha1.Repository{
						Commit: "1",
					},
				},
			},
			expectedResult: true,
		},
		{
			desc: "function change base dir",
			fn: v1alpha1.Function{
				Spec: v1alpha1.FunctionSpec{
					SourceType: v1alpha1.SourceTypeGit,
					Repository: v1alpha1.Repository{
						Commit:  "2",
						BaseDir: "base_dir",
					},
				},
				Status: v1alpha1.FunctionStatus{
					Repository: v1alpha1.Repository{
						Commit: "2",
					},
				},
			},
			expectedResult: true,
		},
		{
			desc: "function change branch",
			fn: v1alpha1.Function{
				Spec: v1alpha1.FunctionSpec{
					SourceType: v1alpha1.SourceTypeGit,
					Repository: v1alpha1.Repository{
						Commit: "2",
						Branch: "branch",
					},
				},
				Status: v1alpha1.FunctionStatus{
					Repository: v1alpha1.Repository{
						Commit: "2",
					},
				},
			},
			expectedResult: true,
		},
		{
			desc: "function change dockerfile",
			fn: v1alpha1.Function{
				Spec: v1alpha1.FunctionSpec{
					SourceType: v1alpha1.SourceTypeGit,
					Repository: v1alpha1.Repository{
						Commit:  "2",
						Runtime: v1alpha1.RuntimeNodeJS12,
					},
				},
				Status: v1alpha1.FunctionStatus{
					Repository: v1alpha1.Repository{
						Commit: "2",
					},
				},
			},
			expectedResult: true,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			g := gomega.NewGomegaWithT(t)
			actual := isOnSourceChange(&tC.fn, tC.revision)
			g.Expect(actual).To(gomega.Equal(tC.expectedResult))
		})
	}
}
