package object

import (
	"knative.dev/eventing/pkg/apis/messaging"
	"knative.dev/pkg/apis"
)

// List of annotations set on Knative Messaging objects by the Knative Eventing
// admission webhook.
var knativeMessagingAnnotations = []string{
	messaging.GroupName + apis.CreatorAnnotationSuffix,
	messaging.GroupName + apis.UpdaterAnnotationSuffix,
}

// List of annotations set on Deployment objects
var deploymentAnnotations = []string{
	"deployment.kubernetes.io/revision",
}
