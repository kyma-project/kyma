module kyma-project.io/compass-runtime-agent

go 1.12

require (
	cloud.google.com/go v0.44.3 // indirect
	github.com/asaskevich/govalidator v0.0.0-20200108200545-475eaeb16496 // indirect
	github.com/kyma-incubator/compass v0.0.0-20200302114843-fb6306fe65c8
	github.com/kyma-incubator/compass/components/director v0.0.0-20200302114843-fb6306fe65c8
	github.com/kyma-project/kyma v0.5.1-0.20200317154738-0bb20217c2cb
	github.com/kyma-project/rafter v0.0.0-20200129064709-d30581e6e574
	github.com/machinebox/graphql v0.2.3-0.20181106130121-3a9253180225
	github.com/pkg/errors v0.8.1
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.4.0
	github.com/vrischmann/envconfig v1.2.0
	k8s.io/api v0.17.2
	k8s.io/apimachinery v0.17.2
	k8s.io/client-go v0.17.2
	k8s.io/metrics v0.15.9
	sigs.k8s.io/controller-runtime v0.5.0
)
