package git

import (
	"errors"

	git2go "github.com/libgit2/git2go/v34"
)

// Unrecoverable means that something is wrong with configuration of git function CR
// and cannot be fixed without changing the function cr. For example, branch doesn't exist.
var notRecoverableErrors = []git2go.ErrorCode{
	git2go.ErrorCodeNotFound, git2go.ErrorCodeInvalidSpec,
}

func IsNotRecoverableError(err error) bool {
	gitErr := getGitErr(err)
	if gitErr == nil {
		return false
	}

	for _, errCode := range notRecoverableErrors {
		if gitErr.Code == errCode {
			return true
		}
	}
	return false
}

func getGitErr(err error) *git2go.GitError {
	gitErr, ok := err.(*git2go.GitError)
	if ok {
		return gitErr
	}
	unwrappedErr := errors.Unwrap(err)
	if unwrappedErr != nil {
		return getGitErr(unwrappedErr)
	}
	return nil
}
