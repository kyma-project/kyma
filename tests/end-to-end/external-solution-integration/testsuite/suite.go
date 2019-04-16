package testsuite

type TestSuite interface {
	DeployResources()
	FetchCertificate()
	RegisterServices()
	StartTestServer()
	SendEvent()
}

type testSuite struct {
}
