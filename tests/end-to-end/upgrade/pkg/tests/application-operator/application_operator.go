package applicationoperator

import "github.com/sirupsen/logrus"

// UpgradeTest checks if Application Operator upgrades image versions of Application Proxy and Event Service of the created Application.
type UpgradeTest struct {
}

// NewApplicationOperatorUpgradeTest returns new instance of the UpgradeTest.
func NewApplicationOperatorUpgradeTest() *UpgradeTest {
	return &UpgradeTest{}
}

// CreateResources creates resources needed for e2e upgrade test
func (ut *UpgradeTest) CreateResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	log.Println("ApplicationOperator UpgradeTest creating resources...")
	//TODO:

	log.Println("ApplicationOperator UpgradeTest resources created!")
	return nil
}

// TestResources tests resources after backup phase
func (ut *UpgradeTest) TestResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	log.Println("ApplicationOperator UpgradeTest testing resources...")
	//TODO:

	log.Println("ApplicationOperator UpgradeTest test passed!")
	return nil
}
