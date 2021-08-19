package main

import (
	natsv1alpha2 "github.com/nats-io/nats-operator/pkg/apis/nats/v1alpha2"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.Info("Started")

	_ = &natsv1alpha2.NatsCluster{}
}
