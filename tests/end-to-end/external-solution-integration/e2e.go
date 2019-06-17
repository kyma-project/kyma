package main

import (
	"time"

	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/testsuite"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/rest"
)

//TODO: This is example use of testing suite, this still needs to be finished and cleaned up
func main() {
	time.Sleep(10 * time.Second)
	log.SetReportCaller(true)
	log.SetLevel(log.TraceLevel)

	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal(err)
	}

	ts, err := testsuite.NewTestSuite(config, log.New())
	if err != nil {
		log.Fatal(err)
	}

	log.RegisterExitHandler(func() {
		log.Error("Starting cleanup")
		err := ts.CleanUp()
		if err != nil {
			log.Error(err)
		}
	})

	defer func() {
		log.Info("Starting cleanup")
		// err := ts.CleanUp()
		// if err != nil {
		// 	log.Error(err)
		// }
	}()

	log.Trace("creating resources")
	err = ts.CreateResources()
	if err != nil {
		log.Fatal(err)
	}

	err = ts.StartTestServer()
	if err != nil {
		log.Fatal(err)
	}

	url := ts.GetTestServiceURL()

	// log.Trace("creating Certificate")

	// cert, err := ts.FetchCertificate()

	log.Trace("registering Service")
	id, err := ts.RegisterService(url)
	if err != nil {
		log.Fatal(err)
	}

	log.Debug("ID:", id)

	log.Trace("Creating Instance")
	_, err = ts.CreateInstance(id)
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(120 * time.Second)

	log.Trace("Creating Service Binding")
	err = ts.CreateServiceBinding()
	if err != nil {
		log.Fatal(err)
	}

	//TODO: Get rid of constant time waits, you can check for readiness of SI / SB / SBU
	time.Sleep(30 * time.Second)

	log.Trace("Creating Service Binding Usage")
	err = ts.CreateServiceBindingUsage()
	if err != nil {
		log.Fatal(err)
	}

	log.Trace("Sending Event")
	err = ts.SendEvent()
	if err != nil {
		log.Fatal(err)
	}
	time.Sleep(30 * time.Second)
	log.Trace("Checking counter pod for the count.")
	err = ts.CheckCounterPod()
	if err != nil {
		log.Fatal(err)
	}

	log.Info("Successfully Finished the e2e test!!")
}
