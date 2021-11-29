package git

import (
	"github.com/stretchr/testify/require"
	"testing"
)
// wrong branch = ErrNotFound
func Test_testa(t *testing.T) {
	//GIVEN
	gitClient := NewGit2Go()
	opts := Options{
		URL:       "https://gdsadasdsa.com/kyma-project/cliiii.git",
		Reference: "main1",
		Auth:      nil,
	}

	//WHEN
	_, err := gitClient.LastCommit(opts)

	//THEN
	require.NoError(t, err)

}
