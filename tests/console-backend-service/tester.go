package tester

import "time"

const (
	DefaultSubscriptionTimeout = 5 * time.Second
	DefaultReadyTimeout        = time.Minute * 3
	TestLabelKey               = "testName"
	TestLabelValue             = "console-backend-service-test"
)
