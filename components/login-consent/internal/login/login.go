package login

import (
	"context"
	"errors"
	"fmt"
	"github.com/kyma-project/kyma/components/login-consent/internal/hydra"
	"log"
	"net/http"
	"net/url"
	"time"
)

type HydraConfig struct {
	hydraAddr string
	hydraPort string
}

func NewHydraConfig(hydraAddr string, hydraPort string) *HydraConfig {
	return &HydraConfig{
		hydraAddr: hydraAddr,
		hydraPort: hydraPort,
	}
}

func (hc *HydraConfig) Login(w http.ResponseWriter, req *http.Request) {
	challenge, err := getLoginChallenge(req.URL.Query())
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}

	ctx, cancel := context.WithCancel(context.Background())
	ch := make(chan string)
	go hydra.GetLoginRequest(ctx, ch, challenge)

	select {
	case result := <-ch:
		fmt.Fprint(w, result)
		cancel()
		return
	case <-time.After(time.Second * 10):
		fmt.Fprint(w, "Server is busy.")
	}
	cancel()
	<-ch
}

func getLoginChallenge(query url.Values) (string, error) {
	challenges, ok := query["login_challenge"]
	if !ok || len(challenges[0]) < 1 {
		log.Println("Url Param 'key' is missing")
		return "", errors.New("login_challenge not found")
	}

	return challenges[0], nil
}
