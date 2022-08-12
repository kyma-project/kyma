package exporter

import (
	"flag"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/directory-size-exporter/utils"

	"github.com/kyma-project/kyma/common/logging/logger"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func initExporterAndRecordMetrics(path string) {
	var logFormat string
	var logLevel string

	flag.StringVar(&logFormat, "log-format", "text", "Log format (json or text)")
	flag.StringVar(&logLevel, "log-level", "info", "Log level (debug, info, warn, error, fatal)")

	exporterLogger, err := logger.New(logger.Format(logFormat), logger.Level(logLevel))
	if err != nil {
		panic(err)
	}

	exp := NewExporter(path, "telemetry_fsbuffer_usage_bytes")
	exporterLogger.WithContext().Info("Exporter is initialized")

	exp.RecordMetrics(5)
	exporterLogger.WithContext().Info("Started recording metrics")

	http.Handle("/metrics", promhttp.Handler())
	err = http.ListenAndServe(":2021", nil)
	if err != nil {
		panic(err)
	}
	exporterLogger.WithContext().Info("Listening on port '2021'")
}

func TestListDir(t *testing.T) {
	dirPath, errDirs := utils.PrepareMockDirectories(t.TempDir())
	assert.NoError(t, errDirs)

	expectedDirectories := []directory{
		{name: "emitter1", size: int64(0)},
		{name: "emitter2", size: int64(100)},
		{name: "emitter3", size: int64(200)},
	}

	directories, err := listDirs(dirPath)
	assert.NoError(t, err)

	isTrue := (len(directories) == len(expectedDirectories))
	for i, dir := range directories {
		if dir != expectedDirectories[i] {
			isTrue = false
			break
		}
	}

	require.True(t, isTrue)
}

func TestDirSize(t *testing.T) {
	dirPath, errDirs := utils.PrepareMockDirectories(t.TempDir())
	assert.NoError(t, errDirs)

	size, err := dirSize(dirPath)
	assert.NoError(t, err)

	require.Equal(t, int64(300), size)
}

func TestNewExporter(t *testing.T) {
	exporter := NewExporter("data/log", "metric_name")
	require.NotNil(t, exporter)
}

func TestRecordMetric(t *testing.T) {
	dirPath, err := utils.PrepareMockDirectories(t.TempDir())
	require.NoError(t, err)

	go initExporterAndRecordMetrics(dirPath)
	time.Sleep(10 * time.Second)

	initialMetrics, err := utils.GetMetrics(2021)
	require.NoError(t, err)

	emitters, err := os.ReadDir(dirPath)
	require.NoError(t, err)
	emitterMetricInitialValue, prs := initialMetrics["telemetry_fsbuffer_usage_bytes{directory=\""+emitters[0].Name()+"\"}"]
	require.True(t, prs)

	_, err = utils.WriteMockFileToDirectory(dirPath+"/"+emitters[0].Name(), "main_test.txt", 500)
	require.NoError(t, err)
	time.Sleep(10 * time.Second)

	metrics, err := utils.GetMetrics(2021)
	require.NoError(t, err)
	emitterMetricValue, prs := metrics["telemetry_fsbuffer_usage_bytes{directory=\""+emitters[0].Name()+"\"}"]
	require.True(t, prs)

	require.NotEqual(t, emitterMetricInitialValue, emitterMetricValue)
	require.Equal(t, "500", emitterMetricValue)
}
