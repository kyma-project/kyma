package crd

import (
	"k8s.io/api/autoscaling/v2beta1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Function custom type is needed, because we add "topic" field to FunctionSpec ourselves in cbs/ui when you select "http" as trigger...

type Function struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              FunctionSpec `json:"spec"`
}

// FunctionSpec contains func specification
type FunctionSpec struct {
	// topic is added ourselves!
	Topic string `json:"topic"`
	// topic is added ourselves!
	Handler                 string                          `json:"handler"`               // Function handler: "file.function"
	Function                string                          `json:"function"`              // Function file content or URL of the function
	FunctionContentType     string                          `json:"function-content-type"` // Function file content type (plain text, base64 or zip)
	Checksum                string                          `json:"checksum"`              // Checksum of the file
	Runtime                 string                          `json:"runtime"`               // Function runtime to use
	Timeout                 string                          `json:"timeout"`               // Maximum timeout for the function to complete its execution
	Deps                    string                          `json:"deps"`                  // Function dependencies
	Deployment              v1beta1.Deployment              `json:"deployment" protobuf:"bytes,3,opt,name=template"`
	ServiceSpec             v1.ServiceSpec                  `json:"service"`
	HorizontalPodAutoscaler v2beta1.HorizontalPodAutoscaler `json:"horizontalPodAutoscaler" protobuf:"bytes,3,opt,name=horizontalPodAutoscaler"`
}
