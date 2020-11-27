package prom

// To parse JSON response from http://monitoring-prometheus.kyma-system:9090/api/v1/alerts
type AlertsResponse struct {
	Status    string     `json:"status"`
	Data      AlertsData `json:"data"`
	ErrorType string     `json:"errorType"`
	Error     string     `json:"error"`
}

type AlertsData struct {
	Alerts []Alert `json:"alerts"`
}

type Alert struct {
	Labels AlertLabels `json:"labels"`
	State  string      `json:"state"`
}

type AlertLabels struct {
	AlertName string `json:"alertname"`
	Severity  string `json:"severity"`
}
