package helpers

import (
	"errors"
	"net/http"
)

func GetLoginChallenge(req *http.Request) (string, error) {
	challenges, ok := req.URL.Query()["login_challenge"]
	if !ok || len(challenges[0]) < 1 {
		return "", errors.New("login_challenge not found")
	}

	return challenges[0], nil
}
