package main

import (
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"fmt"
	"io/ioutil"
	"strings"

	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/typed/core/v1"
	restclient "k8s.io/client-go/rest"
)

const (
	// gatewayUrlEnvName is the env name, which contains gateway url and is injected by REB
	gatewayUrlEnvName = "GATEWAY_URL"

	// targetUrlEnvName is the env name, which contains gateway url, defined in the deployment
	targetUrlEnvName = "TARGET_URL"

	namespaceEnvName = "NAMESPACE"

	sleepTimeBetweenCalls = 3 * time.Second

	envInjectedKey   = "envInjected"
	callSucceededKey = "callSucceeded"
	callForbiddenKey = "callForbidden"
)

/**
The goal of the tester is to try call to gateway (the gateway url is injected as GATEWAY_URL env) and expect
 - 403 if the GATEWAY_URL is not injected (the call use url from TARGET_URL)
 - 200 if the GATEWAY_URL is injected

When the tester starts, updates config map key envInjected. Then tries to call gateway. If the expected status is reached,
saves the result and stops calling gateway.
*/
func main() {
	l := logrus.New()

	gatewayUrl := os.Getenv(gatewayUrlEnvName)
	targetUrl := os.Getenv(targetUrlEnvName)
	namespace := os.Getenv(namespaceEnvName)

	l.Infof("Starting tester")
	l.Infof(" %s=%s", namespaceEnvName, namespace)
	l.Infof(" %s=%s", gatewayUrlEnvName, gatewayUrl)
	l.Infof(" %s=%s", targetUrlEnvName, targetUrl)

	config, err := restclient.InClusterConfig()
	panicOnError(err)
	clientSet, err := kubernetes.NewForConfig(config)
	panicOnError(err)

	cfgMapInterface := clientSet.CoreV1().ConfigMaps(namespace)

	envInjected := false
	if gatewayUrl != "" {
		envInjected = true
	}

	results := resultSaver{
		cfgMapInterface: cfgMapInterface,
		logger:          l,
	}

	client := http.Client{
		Transport: &http.Transport{
			// do not reuse TCP connections because of Istio can return 404 for long time
			DisableKeepAlives: true,
		},
	}

	panicOnError(
		executeWithRetries(time.Minute, func() error {
			return results.save(envInjectedKey, strconv.FormatBool(envInjected))
		},
		))

	url := gatewayUrl
	if !envInjected {
		url = targetUrl
	}

	go func() {
		for {
			req, err := http.NewRequest(http.MethodGet, url, http.NoBody)
			if err != nil {
				l.Errorf(err.Error())
				return
			}
			resp, err := client.Do(req)
			if err != nil {
				l.Warnf(err.Error())
				continue
			}
			body, _ := ioutil.ReadAll(resp.Body)
			resp.Body.Close()

			l.Infof("Response from %s (from env %v): %d", url, envInjected, resp.StatusCode)
			if len(body) > 0 {
				fmt.Println(string(body))
			}

			// this is a part of testing scenario:
			// when the pod has GATEWAY_URL injected (envInjected=true) - we expect HTTP 200
			// when GATEWAY_URL is not injected - we expect HTTP 403 from the Istio denier with the denier message
			if envInjected {
				// expecting HTTP 200
				if resp.StatusCode == http.StatusOK {
					panicOnError(results.save(callSucceededKey, "true"))
					return
				}
			} else {
				// env not injected - expect 403 forbidden
				if resp.StatusCode == http.StatusForbidden && strings.Contains(string(body), "Not allowed by istio denier") {
					panicOnError(results.save(callForbiddenKey, "true"))
					return
				}
			}

			time.Sleep(sleepTimeBetweenCalls)
		}
	}()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	l.Infoln("Shutdown signal received shutting down gracefully...")
}

type resultSaver struct {
	cfgMapInterface v1.ConfigMapInterface
	logger          logrus.FieldLogger
}

func (rs *resultSaver) save(key, value string) error {
	rs.logger.Infof("Updating config map key=%s, value=%s", key, value)
	cfg, err := rs.cfgMapInterface.Get("test-output", metav1.GetOptions{})
	if err != nil {
		rs.logger.Warnf("Cannot get config map, error: %s", err.Error())
		return err
	}

	cfgCopy := cfg.DeepCopy()
	cfgCopy.Data[key] = value

	_, err = rs.cfgMapInterface.Update(cfgCopy)
	if err != nil {
		rs.logger.Warnf("Update get config map, error: %s", err.Error())
		return err
	}
	return err
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}

func executeWithRetries(timeout time.Duration, fn func() error) error {
	done := time.After(timeout)

	for range time.Tick(2 * time.Second) {
		err := fn()
		if err == nil {
			break
		}

		select {
		case <-done:
			return fmt.Errorf("timeout while executing command, last error: %s", err.Error())
		default:
		}
	}

	return nil
}
