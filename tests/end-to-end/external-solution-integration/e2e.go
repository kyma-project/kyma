package main

import (
	"fmt"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/testsuite"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/rest"
	"time"
)

//TODO: This is example use of testing suite, this still needs to be finished and cleaned up
func main() {
	config, err := rest.InClusterConfig()
	if err != nil {
		fmt.Println(err)
	}

	ts, err := testsuite.NewTestSuite(config, logrus.New())
	if err != nil {
		fmt.Println(err)
	}

	err = ts.CreateResources()
	if err != nil {
		fmt.Println(err)
	}

	//cert, err := ts.FetchCertificate()

	id, err := ts.RegisterService()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("ID:", id)

	_, err = ts.CreateInstance(id)
	if err != nil {
		fmt.Println(err)
	}

	time.Sleep(30 * time.Second)

	err = ts.CreateServiceBinding()
	if err != nil {
		fmt.Println(err)
	}

	time.Sleep(30 * time.Second)

	err = ts.CreateServiceBindingUsage()
	if err != nil {
		fmt.Println(err)
	}
}