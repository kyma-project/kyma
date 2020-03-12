package director

import (
	"strings"

	"github.com/pkg/errors"
)

// IsGQLNotFoundError checks if graphql response error
// represents not found error.
//
// This is so funny that I had to right such piece of code. DON'T TRY THIS AT HOME
//
// Fortunately there is a light at the end of the tunnel,
// check https://github.com/kyma-incubator/compass/issues/66
func IsGQLNotFoundError(err error) bool {
	err = errors.Cause(err)

	if err == nil {
		return false
	}

	return strings.Contains(err.Error(), "Object was not found")
}
