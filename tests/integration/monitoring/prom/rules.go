package prom

// To parse JSON response from http://monitoring-prometheus.kyma-system:9090/api/v1/rules
type RulesResponse struct {
	Status    string    `json:"status"`
	Data      RulesData `json:"data"`
	ErrorType string    `json:"errorType"`
	Error     string    `json:"error"`
}

type RulesData struct {
	Groups []RulesGroup `json:"groups"`
}

type RulesGroup struct {
	Rules []Rule `json:"rules"`
}

type Rule struct {
	Name   string `json:"name"`
	Health string `json:"health"`
}
