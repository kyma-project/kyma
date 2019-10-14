package main

import (
	"encoding/json"
	"github.com/kyma-project/kyma/tests/service-catalog/test"
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {
	http.HandleFunc("/envs", envChecker)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func envChecker(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	envVariables := os.Environ()
	out := make([]test.EnvVariable, len(envVariables))
	for i, line := range envVariables {
		splitted := strings.SplitN(line, "=", 2)
		out[i] = test.EnvVariable{Name: splitted[0], Value: splitted[1]}
	}

	err := json.NewEncoder(w).Encode(out)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
