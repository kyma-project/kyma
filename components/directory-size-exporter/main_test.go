package main

import (
	"bufio"
	"flag"
	"net/http"
	"os"
	"testing"
	"time"

	"directory-size-exporter/utils"

	"github.com/stretchr/testify/require"
)

func TestMainMetric(t *testing.T) {
	flag.Set("test.timeout", "2m0s")
	dirPath, err := utils.PrepareMockDirectories(t.TempDir())
	require.NoError(t, err)

	os.Setenv("STORAGE_PATH", dirPath)
	os.Setenv("DIRECTORIES_SIZE_METRIC", "telemetry_fsbuffer_usage_bytes")
	go main()
	time.Sleep(35 * time.Second)
	res, err := http.Get("http://localhost:2021/metrics")
	require.NoError(t, err)
	defer res.Body.Close()
	scanner := bufio.NewScanner(res.Body)
	for scanner.Scan() {
		line := scanner.Text()
		println(line)
	}
	require.True(t, true)
}

func TestReadEnvironmentVariable(t *testing.T) {
	os.Setenv("TEST_VARIABLE", "1")
	val, err := readEnvironmentVariable("TEST_VARIABLE")
	require.NoError(t, err)
	require.Equal(t, val, "1")

	os.Setenv("TEST_VARIABLE", "")
	_, err = readEnvironmentVariable("TEST_VARIABLE")
	require.Error(t, err)
}
