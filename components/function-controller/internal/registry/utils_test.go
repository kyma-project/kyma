package registry

import (
	"reflect"
	"testing"

	"github.com/docker/distribution/reference"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	testSecretName = "test-secret"
	testNS         = "test-ns"
)

func Test_listFunctionDeployments(t *testing.T) {
	tests := []struct {
		name      string
		k8sClient client.Client
		want      *appsv1.DeploymentList
		wantErr   bool
	}{
		{
			name:      "no deployments",
			k8sClient: fake.NewClientBuilder().Build(),
			want:      &appsv1.DeploymentList{Items: []appsv1.Deployment{}},
			wantErr:   false,
		},
		{
			name: "multiple deployments without function labels",
			k8sClient: fake.NewClientBuilder().WithObjects(
				&appsv1.Deployment{
					ObjectMeta: v1.ObjectMeta{Name: "d1", Labels: map[string]string{"label1": "value1"}},
				}, &appsv1.Deployment{
					ObjectMeta: v1.ObjectMeta{Name: "d2", Labels: map[string]string{"label2": "value2"}},
				},
			).Build(),
			want:    &appsv1.DeploymentList{Items: []appsv1.Deployment{}},
			wantErr: false,
		},
		{
			name: "multiple deployments with one with function labels",
			k8sClient: fake.NewClientBuilder().WithObjects(
				&appsv1.Deployment{
					ObjectMeta: v1.ObjectMeta{Name: "d1", Labels: map[string]string{"label1": "value1"}},
				}, &appsv1.Deployment{
					ObjectMeta: v1.ObjectMeta{Name: "d2", Labels: map[string]string{"label2": "value2"}},
				}, &appsv1.Deployment{
					ObjectMeta: v1.ObjectMeta{Name: "d3", Labels: functionRuntimeLabels},
				},
			).Build(),
			want: &appsv1.DeploymentList{Items: []appsv1.Deployment{{
				ObjectMeta: v1.ObjectMeta{Name: "d3", Labels: functionRuntimeLabels},
			}},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := listFunctionDeployments(tt.k8sClient)
			if (err != nil) != tt.wantErr {
				t.Errorf("listFunctionDeployments() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got.Items) != len(tt.want.Items) {
				t.Errorf("listFunctionDeployments() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestGetFunctionImage(t *testing.T) {
	tests := []struct {
		name    string
		d       appsv1.Deployment
		want    reference.NamedTagged
		wantErr bool
	}{
		{
			name:    "deployment without function container",
			d:       deploymentWithContainerImages(containerImageMap{"container1": "test1", "container2": "test2"}),
			want:    nil,
			wantErr: true,
		},
		{
			name:    "deployment with function container",
			d:       deploymentWithContainerImages(containerImageMap{"function": "registry.local/function/image:tagged", "container2": "test2"}),
			want:    taggedImage("registry.local/function/image:tagged"),
			wantErr: false,
		},
		{
			name:    "deployment with function container and uncanonical image name",
			d:       deploymentWithContainerImages(containerImageMap{"function": "function/image:tagged", "container2": "test2"}),
			want:    nil,
			wantErr: true,
		},
		{
			name:    "deployment with function container and untagged image",
			d:       deploymentWithContainerImages(containerImageMap{"function": "registry.local/function/image", "container2": "test2"}),
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetFunctionImage(tt.d)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetFunctionImage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetFunctionImage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getConfigSecretData(t *testing.T) {
	type args struct {
		k8sClient  client.Client
		secretName string
		namespace  string
	}
	tests := []struct {
		name    string
		args    args
		want    *RegistryClientOptions
		wantErr bool
	}{
		{
			name: "correct secret",
			args: args{
				k8sClient:  fake.NewClientBuilder().WithObjects(secretWithData("username", "password", "registry.local:5000")).Build(),
				secretName: testSecretName,
				namespace:  testNS,
			},
			want: &RegistryClientOptions{
				Username: "username",
				Password: "password",
				URL:      "registry.local:5000",
			},
		},
		{
			name: "secret with missing key",
			args: args{
				k8sClient:  fake.NewClientBuilder().WithObjects(secretWithData("username", "", "registry.local:5000")).Build(),
				secretName: testSecretName,
				namespace:  testNS,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "no secret",
			args: args{
				k8sClient:  fake.NewClientBuilder().Build(),
				secretName: testSecretName,
				namespace:  testNS,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getConfigSecretData(tt.args.k8sClient, tt.args.secretName, tt.args.namespace)
			if (err != nil) != tt.wantErr {
				t.Errorf("getConfigSecretData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getConfigSecretData() = %v, want %v", got, tt.want)
			}
		})
	}
}

type containerImageMap map[string]string

func deploymentWithContainerImages(m containerImageMap) appsv1.Deployment {

	d := appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "",
							Image: "image",
						},
					},
				},
			},
		},
	}
	for c, i := range m {
		d.Spec.Template.Spec.Containers = append(d.Spec.Template.Spec.Containers, corev1.Container{Name: c, Image: i})
	}
	return d
}

func taggedImage(i string) reference.NamedTagged {
	r, _ := reference.ParseNamed(i)
	return r.(reference.NamedTagged)
}

func secretWithData(username, password, url string) *corev1.Secret {

	s := &corev1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name:      testSecretName,
			Namespace: testNS,
		},
		Data: map[string][]byte{},
	}
	if username != "" {
		s.Data[UsernameSecretKeyName] = []byte(username)
	}
	if password != "" {
		s.Data[PasswordSecretKeyName] = []byte(password)
	}
	if url != "" {

		s.Data[URLSecretKeyName] = []byte(url)
	}
	return s
}
