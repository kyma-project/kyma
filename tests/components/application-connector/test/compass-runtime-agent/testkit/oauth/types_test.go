package oauth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestToken_EmptyOrExpired(t *testing.T) {
	t.Run("Should return true when token is empty", func(t *testing.T) {
		//given
		token := Token{}

		//when
		empty := token.EmptyOrExpired()

		//then
		assert.True(t, empty)
	})

	t.Run("Should return true when expired", func(t *testing.T) {
		//given
		time2000 := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC).Unix()

		token := Token{
			AccessToken: "token",
			Expiration:  time2000,
		}

		//when
		expired := token.EmptyOrExpired()

		//then
		assert.True(t, expired)
	})

	t.Run("Should return false when not empty or expired", func(t *testing.T) {
		//given
		time3000 := time.Date(3000, 1, 1, 0, 0, 0, 0, time.UTC).Unix()

		token := Token{
			AccessToken: "token",
			Expiration:  time3000,
		}

		//when
		notExpired := token.EmptyOrExpired()

		//then
		assert.False(t, notExpired)
	})
}
