package helpers

import (
	"errors"
	"net/url"
)

func GetLoginChallenge(query url.Values) (string, error) {
	challenges, ok := query["login_challenge"]
	if !ok || len(challenges[0]) < 1 {
		return "", errors.New("login_challenge not found")
	}

	return challenges[0], nil
}
