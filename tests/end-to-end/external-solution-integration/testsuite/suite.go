package testsuite

type TestSuite interface {
	DeployResources()
	FetchCertificate()
	RegisterService()
	StartTestServer()
	SendEvent()
}

type testSuite struct {
}
