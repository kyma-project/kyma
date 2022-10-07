package testing

import (
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"time"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
)

// binary cloudevent headers
const (
	CeIDHeader          = "ce-id"
	CeTypeHeader        = "ce-type"
	CeSourceHeader      = "ce-source"
	CeSpecVersionHeader = "ce-specversion"
)

type SubscriptionOpt func(*eventingv1alpha1.Subscription)

// GeneratePortOrDie generates a random 5 digit port or fail
func GeneratePortOrDie() int {
	tick := time.NewTicker(time.Second / 2)
	defer tick.Stop()

	timeout := time.NewTimer(time.Minute)
	defer timeout.Stop()

	for {
		select {
		case <-tick.C:
			{
				port, err := generatePort()
				if err != nil {
					break
				}

				if !isPortAvailable(port) {
					break
				}

				return port
			}
		case <-timeout.C:
			{
				log.Fatal("Failed to generate port")
			}
		}
	}
}

func generatePort() (int, error) {
	max := 4
	// Add 4 as prefix to make it 5 digits but less than 65535
	add4AsPrefix := "4"
	b := make([]byte, max)
	n, err := io.ReadAtLeast(rand.Reader, b, max)
	if n != max {
		return 0, err
	}
	if err != nil {
		return 0, err
	}
	for i := 0; i < len(b); i++ {
		b[i] = table[int(b[i])%len(table)]
	}

	num, err := strconv.Atoi(fmt.Sprintf("%s%s", add4AsPrefix, string(b)))
	if err != nil {
		return 0, err
	}

	return num, nil
}

var table = [...]byte{'1', '2', '3', '4', '5', '6', '7', '8', '9'}

// isPortAvailable returns true if the port is available for use, otherwise returns false
func isPortAvailable(port int) bool {
	address := fmt.Sprintf("localhost:%d", port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return false
	}

	if err := listener.Close(); err != nil {
		return false
	}

	return true
}
