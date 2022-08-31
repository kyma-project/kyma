package testing

import (
	"fmt"
	"net"

	"github.com/nats-io/nats-server/v2/server"

	eventingtesting "github.com/kyma-project/kyma/components/eventing-controller/testing"
)

func GetFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

func StartNATSServer(serverOpts ...eventingtesting.NatsServerOpt) (*server.Server, int, error) {
	natsPort, err := GetFreePort()
	if err != nil {
		return nil, 0, err

	}
	serverOpts = append(serverOpts, eventingtesting.WithPort(natsPort))
	natsServer := eventingtesting.RunNatsServerOnPort(serverOpts...)
	return natsServer, natsPort, nil
}

func NewNatsMessagePayload(data, id, source, eventTime, eventType string) string {
	jsonCE := fmt.Sprintf("{\"data\":\"%s\",\"datacontenttype\":\"application/json\",\"id\":\"%s\",\"source\":\"%s\",\"specversion\":\"1.0\",\"time\":\"%s\",\"type\":\"%s\"}", data, id, source, eventTime, eventType)
	return jsonCE
}
