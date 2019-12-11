/*
Copyright 2019 The Kyma Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package object

import (
	"knative.dev/eventing/pkg/apis/messaging"
	"knative.dev/pkg/apis"
	"knative.dev/serving/pkg/apis/serving"
)

// List of annotations set on Knative Messaging objects by the Knative Eventing
// admission webhook.
var knativeMessagingAnnotations = []string{
	messaging.GroupName + apis.CreatorAnnotationSuffix,
	messaging.GroupName + apis.UpdaterAnnotationSuffix,
}

// List of annotations set on Knative Serving objects by the Knative Serving
// admission webhook.
var knativeServingAnnotations = []string{
	serving.GroupName + apis.CreatorAnnotationSuffix,
	serving.GroupName + apis.UpdaterAnnotationSuffix,
}
