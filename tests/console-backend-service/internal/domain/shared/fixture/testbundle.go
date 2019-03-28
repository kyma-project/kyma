package fixture

const TestingBrokerName = "helm-broker"

const TestingBundleClassName = "faebbe18-0a84-11e9-ab14-d663bd873d94"
const TestingBundleClassExternalName = "testing"

const TestingBundleFullPlanName = "a6078799-70a1-4674-af91-aba44dd6a56"
const TestingBundleFullPlanExternalName = "full"

const TestingBundleMinimalPlanName = "631dae68-98e1-4e45-b79f-1036ca5b29cb"
const TestingBundleMinimalPlanExternalName = "minimal"

var TestingBundleFullPlanSpec = map[string]interface{}{
	"planName":       "test",
	"additionalData": "foo",
}
