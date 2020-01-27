package tsm

import (
	"testing"

	"github.com/kyma-project/kyma/tests/end-to-end/backup/pkg/tests/ory"
	. "github.com/smartystreets/goconvey/convey"
)

func TestOry(t *testing.T) {

	Convey("Test ORY hydra-maester\n", t, func() {
		hydraClientTest, err := ory.NewHydraClientTest()
		if err != nil {
			panic(err)
		}
		hydraClientTest.CreateResources("padu")
	})
}
