package v1alpha1

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestTokenRequest_ShouldExpire(t *testing.T) {
	t.Run("should return false on default value for ExpireAfter", func(t *testing.T) {
		tokenRequest := TokenRequest{}

		assert.Equal(t, tokenRequest.ShouldExpire(), false)
	})

	t.Run("should return false when ExpireAfter after current time", func(t *testing.T) {
		tokenRequest := TokenRequest{
			Status: TokenRequestStatus{
				ExpireAfter: metav1.NewTime(metav1.Now().Add(time.Second * time.Duration(10))),
			},
		}

		assert.Equal(t, tokenRequest.ShouldExpire(), false)
	})

	t.Run("should return true when ExpireAfter before current time", func(t *testing.T) {
		tokenRequest := TokenRequest{
			Status: TokenRequestStatus{
				ExpireAfter: metav1.NewTime(metav1.Now().Add(time.Second * time.Duration(-10))),
			},
		}

		assert.Equal(t, tokenRequest.ShouldExpire(), true)
	})
}
