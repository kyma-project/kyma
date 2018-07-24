package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	http.HandleFunc("/envs", envChecker)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func envChecker(w http.ResponseWriter, r *http.Request) {
	var (
		envName     = r.URL.Query().Get("name")
		expEnvValue = r.URL.Query().Get("value")
	)

	if envName == "" {
		http.Error(w, "Missing 'name' query param", http.StatusBadRequest)
		return
	}

	if !isEnvVariableSetInOS(envName, expEnvValue) {
		w.WriteHeader(http.StatusNotFound)
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
