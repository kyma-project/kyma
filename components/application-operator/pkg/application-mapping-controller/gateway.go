package application_mapping_controller

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

var gateways []string

func (g *gatewayDeployerStub) DeployGateway(namespace string) error {
	appendGateway(namespace)
	return nil
}

func (g *gatewayDeployerStub) RemoveGateway(namespace string) error {
	remove(namespace)
	return nil
}

func (g *gatewayDeployerStub) GatewayExists(namespace string) bool {
	return exists(namespace)
}

func remove(namespace string) {
	for i, v := range gateways {
		if v == namespace {
			gateways = append(gateways[:i], gateways[i+1:]...)
			break
		}
	}
}

func exists(namespace string) bool {
	for _, v := range gateways {
		if v == namespace {
			return true
		}
	}
	return false
}

func appendGateway(namespace string) {
	if exists(namespace) {
		return
	}
	gateways = append(gateways, namespace)
}
