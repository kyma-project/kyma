package hydra

import hydraAPI "github.com/ory/hydra-client-go/models"

func GetLoginRequest(challenge string) hydraAPI.LoginRequest {
	f := false
	return hydraAPI.LoginRequest{
		Challenge: &challenge,
		Skip:      &f,
	}
}

func AcceptLoginRequest(challenge string, request hydraAPI.AcceptLoginRequest) hydraAPI.CompletedRequest {
	addr := "url.com"
	return hydraAPI.CompletedRequest{
		RedirectTo: &addr,
	}
}
