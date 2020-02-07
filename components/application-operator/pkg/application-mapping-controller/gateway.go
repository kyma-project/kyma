package application_mapping_controller

import log "github.com/sirupsen/logrus"

//go:generate mockery -name GatewayDeployer
type GatewayDeployer interface {
	DeployGateway(namespace string) error
	RemoveGateway(namespace string) error
	GatewayExists(namespace string) bool
}

func NewGatewayDeployerStub() GatewayDeployer {
	return &gatewayDeployerStub{}
}

type gatewayDeployerStub struct{}

func (g *gatewayDeployerStub) DeployGateway(namespace string) error {
	log.Infof("Deploying Gateway for namespace %s", namespace)
	return nil
}

func (g *gatewayDeployerStub) RemoveGateway(namespace string) error {
	log.Infof("Removing Gateway from namespace %s", namespace)
	return nil
}

func (g *gatewayDeployerStub) GatewayExists(namespace string) bool {
	log.Infof("Checking if Gateway for namespace %s exists", namespace)
	return false
}
