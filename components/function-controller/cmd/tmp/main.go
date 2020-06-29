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
	config := gitops.Config{
		RepoUrl: "https://github.com/kyma-project/kyma",
		Branch: "master",
		ActualCommit: "",
		BaseDir: "",
		Secret: nil,
	}
	commit, _, err := mgr.CheckBranchChanges(config)
	if err != nil {
		panic(errors.Wrap(err, "during getting latest commit from branch"))
	}

	fmt.Printf("Commit: %s", commit)
}
