package envvar

import (
	"fmt"
	"strings"

	"github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
)

func GenerateName(logpipelineName string, secretKeyRef v1alpha1.SecretKeyRef) string {
	result := fmt.Sprintf("%s_%s_%s_%s", logpipelineName, secretKeyRef.Namespace, secretKeyRef.Name, secretKeyRef.Key)
	result = strings.ToUpper(result)
	result = strings.Replace(result, ".", "_", -1)
	result = strings.Replace(result, "-", "_", -1)
	return result
}
