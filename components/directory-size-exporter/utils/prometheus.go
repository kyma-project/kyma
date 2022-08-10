package utils

import (
	"bufio"
	"fmt"
	"net/http"
	"strings"
)

func GetMetrics(port int) (map[string]string, error) {
	metrics := map[string]string{}
	res, err := http.Get("http://localhost:" + fmt.Sprint(port) + "/metrics")
	if err != nil {
		return metrics, err
	}
	defer res.Body.Close()
	scanner := bufio.NewScanner(res.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}

		lineMetrics := strings.Split(line, " ")
		if len(lineMetrics) < 2 {
			continue
		}

		metrics[lineMetrics[0]] = lineMetrics[1]
	}
	return metrics, err
}
