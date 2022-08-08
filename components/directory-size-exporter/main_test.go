package main

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"directory-size-exporter/utils"

	"github.com/stretchr/testify/require"
)

func TestMainMetric(t *testing.T) {
	dirPath, err := utils.PrepareMockDirectories(t.TempDir())
	require.NoError(t, err)

	os.Setenv("STORAGE_PATH", dirPath)
	os.Setenv("DIRECTORIES_SIZE_METRIC", "telemetry_fsbuffer_usage_bytes")
	go main()
	time.Sleep(35 * time.Second)
	// res, err := http.Get("http://localhost:2021/metrics")
	// require.NoError(t, err)
	// defer res.Body.Close()
	// scanner := bufio.NewScanner(res.Body)
	// for scanner.Scan() {
	// 	line := scanner.Text()
	// 	println(line)
	// }

	initialMetrics, err := utils.GetMetrics(2021)
	require.NoError(t, err)
	print(initialMetrics)

	emitters, err := ioutil.ReadDir(dirPath)
	require.NoError(t, err)

	_, err = utils.WriteMockFileToDirectory(dirPath+"/"+emitters[0].Name(), "main_test.txt", 500)
	require.NoError(t, err)

	time.Sleep(35 * time.Second)
	// res, err = http.Get("http://localhost:2021/metrics")
	// require.NoError(t, err)
	// defer res.Body.Close()
	// scanner = bufio.NewScanner(res.Body)
	// for scanner.Scan() {
	// 	line := scanner.Text()
	// 	println(line)
	// }

	metrics, err := utils.GetMetrics(2021)
	require.NoError(t, err)

	print(metrics)

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
