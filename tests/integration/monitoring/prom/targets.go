package prom

// To parse JSON response from http://monitoring-prometheus.kyma-system:9090/api/v1/targets
type TargetsResponse struct {
	Status    string      `json:"status"`
	Data      TargetsData `json:"data"`
	ErrorType string      `json:"errorType"`
	Error     string      `json:"error"`
}

type TargetsData struct {
	ActiveTargets []ActiveTarget `json:"activeTargets"`
}

type ActiveTarget struct {
	Labels     TargetLabels `json:"labels"`
	ScrapePool string       `json:"scrapePool"`
	LastError  string       `json:"lastError"`
	Health     string       `json:"health"`
}

type TargetLabels map[string]string
