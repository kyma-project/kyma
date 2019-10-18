package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/kyma-project/kyma/tests/service-catalog/cmd/env-tester/dto"
)

func main() {
	http.HandleFunc("/envs", envChecker)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func envChecker(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	envVariables := os.Environ()
	out := make([]dto.EnvVariable, len(envVariables))
	for i, line := range envVariables {
		splitted := strings.SplitN(line, "=", 2)
		out[i] = dto.EnvVariable{Name: splitted[0], Value: splitted[1]}
	}

	err := json.NewEncoder(w).Encode(out)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
