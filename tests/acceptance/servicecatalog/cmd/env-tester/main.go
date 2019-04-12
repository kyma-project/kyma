package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/kyma-project/kyma/tests/acceptance/servicecatalog"
)

func main() {
	http.HandleFunc("/envs", envChecker)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func envChecker(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	envVariables := os.Environ()
	out := make([]servicecatalog.Variable, len(envVariables))
	for i, line := range envVariables {
		splitted := strings.SplitN(line, "=", 2)
		out[i] = servicecatalog.Variable{Name: splitted[0], Value: splitted[1]}
	}

	err := json.NewEncoder(w).Encode(out)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func isEnvVariableSetInOS(envName, expEnvValue string) bool {
	osEnvValue, found := os.LookupEnv(envName)
	if !found {
		return false
	}

	if osEnvValue != expEnvValue {
		return false
	}

	return true
}
