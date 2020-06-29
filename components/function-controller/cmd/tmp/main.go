package main

import (
	"fmt"
	"github.com/kyma-project/kyma/components/function-controller/internal/controllers/serverless/gitops"
	"github.com/pkg/errors"
)

func main() {
	mgr, err := gitops.NewManager(gitops.NewOperator())
	if err != nil {
		panic(errors.Wrap(err, "during creating gitops manager"))
	}

	commit, err := mgr.GetLastCommit("https://github.com/go-git/go-git", "pr-1152", nil)
	if err != nil {
		panic(errors.Wrap(err, "during getting latest commit from branch"))
	}

	fmt.Printf("Commit: %s", commit)
}
