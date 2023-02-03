package envvar

import (
	"fmt"
	"strings"

	"github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
)

func GenerateName(prefix string, secretKeyRef v1alpha1.SecretKeyRef) string {
	result := fmt.Sprintf("%s_%s_%s_%s", prefix, secretKeyRef.Namespace, secretKeyRef.Name, secretKeyRef.Key)
	return MakeEnvVarCompliant(result)
}

func MakeEnvVarCompliant(input string) string {
	result := input
	result = strings.ToUpper(result)
	result = strings.Replace(result, ".", "_", -1)
	result = strings.Replace(result, "-", "_", -1)
	return result
}
