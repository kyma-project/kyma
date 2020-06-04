package prom

// To parse JSON response from http://monitoring-prometheus.kyma-system:9090/api/v1/rules
type AlertResponse struct {
	Status    string    `json:"status"`
	Data      AlertData `json:"data"`
	ErrorType string    `json:"errorType"`
	Error     string    `json:"error"`
}

type AlertData struct {
	Groups []AlertDataGroup `json:"groups"`
}

type AlertDataGroup struct {
	Rules []Rule `json:"rules"`
}

type Rule struct {
	Name   string `json:"name"`
	Health string `json:"health"`
}
