package main

import (
	"fmt"

	"github.com/kyma-project/kyma/components/function-controller/internal/gitops"
	"github.com/pkg/errors"
)

func main() {
	opr := gitops.NewOperator()

	config := gitops.Config{
		RepoUrl:      "https://github.com/kyma-project/kyma",
		Branch:       "master",
		ActualCommit: "",
		BaseDir:      "",
		Secret:       nil,
	}
	commit, _, err := opr.CheckBranchChanges(config)
	if err != nil {
		panic(errors.Wrap(err, "during getting latest commit from branch"))
	}

	fmt.Printf("Commit: %s", commit)
}
