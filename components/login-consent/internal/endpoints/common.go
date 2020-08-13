package endpoints

import (
	hydraAPI "github.com/ory/hydra-client-go/models"
	"math/rand"
	"time"
)

var challenge string
var state = generateRandomString(16)

func (cfg *Config) rejectLoginRequest(err error, statusCode int64) (string, error) {
	body := &hydraAPI.RejectRequest{
		Error:            err.Error(),
		ErrorDebug:       "",
		ErrorDescription: "",
		ErrorHint:        "",
		StatusCode:       statusCode,
	}
	res, e := cfg.client.RejectLoginRequest(challenge, body)
	if e != nil {
		return "", e
	}
	return *res.RedirectTo, nil
}

func generateRandomString(length int) string {
	rand.Seed(time.Now().UnixNano())

	letterRunes := []rune("abcdefghijklmnopqrstuvwxyz")

	b := make([]rune, length)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
