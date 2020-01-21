# knative-kafka

This chart is taken from <https://github.com/kyma-incubator/knative-kafka>

## Update chart

```bash
cp -r $GOPATH/src/github.com/kyma-incubator/knative-kafka/resources/knative-kafka/ $GOPATH/src/github.com/kyma-project/kyma/resources/knative-kafka
```

## Install chart

```bash
cd resources
helm install -n knative-kafka ./knative-kafka
```