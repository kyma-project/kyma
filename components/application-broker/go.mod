module github.com/kyma-project/kyma/components/application-broker

go 1.13

require (
	cloud.google.com/go v0.53.0 // indirect
	contrib.go.opencensus.io/exporter/ocagent v0.6.0 // indirect
	contrib.go.opencensus.io/exporter/prometheus v0.1.0 // indirect
	contrib.go.opencensus.io/exporter/stackdriver v0.12.9-0.20191108183826-59d068f8d8ff // indirect
	github.com/99designs/gqlgen v0.10.2 // indirect
	github.com/asaskevich/govalidator v0.0.0-20200108200545-475eaeb16496
	github.com/aws/aws-sdk-go v1.29.12 // indirect
	github.com/docker/distribution v2.6.0-rc.1.0.20170726174610-edc3ab29cdff+incompatible // indirect
	github.com/emicklei/go-restful v2.11.2+incompatible // indirect
	github.com/go-openapi/spec v0.19.6 // indirect
	github.com/go-openapi/swag v0.19.7 // indirect
	github.com/google/go-cmp v0.4.0
	github.com/googleapis/gnostic v0.4.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.13.0 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/kubernetes-sigs/service-catalog v0.2.2
	github.com/kyma-incubator/compass/components/connectivity-adapter v0.0.0-20200526095328-17a095d4cc56
	github.com/kyma-incubator/compass/components/director v0.0.0-20200526095328-17a095d4cc56
	github.com/kyma-project/kyma v0.5.1-0.20191031094708-0559e17e7979
	github.com/kyma-project/kyma/components/application-operator v0.0.0-20191031094708-0559e17e7979 // indirect
	github.com/machinebox/graphql v0.2.3-0.20181106130121-3a9253180225
	github.com/mailru/easyjson v0.7.1 // indirect
	github.com/mattbaird/jsonpatch v0.0.0-20171005235357-81af80346b1a // indirect
	github.com/mcuadros/go-defaults v1.1.0 // indirect
	github.com/meatballhat/negroni-logrus v1.1.1-0.20191208165538-6f72cade44a3 // indirect
	github.com/opencontainers/go-digest v1.0.0-rc1 // indirect
	github.com/openzipkin/zipkin-go v0.2.2 // indirect
	github.com/pkg/errors v0.9.1
	github.com/pmorie/go-open-service-broker-client v0.0.0-20180928143052-79b374a2302f // indirect
	github.com/robfig/cron v1.2.0 // indirect
	github.com/stretchr/testify v1.5.1
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	go.uber.org/multierr v1.5.0 // indirect
	go.uber.org/zap v1.14.0 // indirect
	golang.org/x/crypto v0.0.0-20200221231518-2aa609cf4a9d // indirect
	golang.org/x/net v0.0.0-20200226121028-0de0cce0169b // indirect
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0 // indirect
	golang.org/x/tools v0.0.0-20200227222343-706bc42d1f0d // indirect
	google.golang.org/api v0.19.0 // indirect
	google.golang.org/genproto v0.0.0-20200227132054-3f1135a288c9 // indirect
	gopkg.in/square/go-jose.v2 v2.4.1 // indirect
	istio.io/api v0.0.0-20200107183329-ed4b507c54e1
	istio.io/client-go v0.0.0-20200107185429-9053b0f86b03
	istio.io/gogo-genproto v0.0.0-20191029161641-f7d19ec0141d // indirect
	k8s.io/api v0.17.4
	k8s.io/apiextensions-apiserver v0.17.4 // indirect
	k8s.io/apimachinery v0.17.4
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/kubernetes v0.16.0 // indirect
	k8s.io/utils v0.0.0-20200124190032-861946025e34 // indirect
	knative.dev/eventing v0.12.0
	knative.dev/pkg v0.0.0-20200113182502-b8dc5fbc6d2f
	sigs.k8s.io/testing_frameworks v0.1.1 // indirect
	sigs.k8s.io/yaml v1.2.0 // indirect
)

replace github.com/kyma-project/kyma/components/application-operator v0.0.0-20191031094708-0559e17e7979 => ../application-operator
