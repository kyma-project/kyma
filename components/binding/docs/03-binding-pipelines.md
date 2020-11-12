# Kyma Binding Pipelines

During the development of the Kyma Binding component we decided to change the pipelines from the regular jobs run on Prow. Main focus was to make them faster. The idea was to run the job on k3s, since it's objectively the [best](https://github.com/kyma-incubator/local-kyma#i-see-k3s-k3d-kind-and-minikube---what-should-i-use), and install only the Binding component, without the whole Kyma. The decision to use Github Actions was made since the k3s scripts were already [developed](https://www.facebook.com/groups/1345683025593684/permalink/1636584193170231/). Main problem for this approach was that GitHub doesn't propagate the secrets created on the main repository, to pull requests opened from forked repositories. As the workflow in Kyma is based on using forks we had to create a workaround for that. The idea for that was found [here](https://github.com/imjohnbo/ok-to-test). The jobs that we developed:

## [Is trusted dev](https://github.com/kyma-project/kyma/blob/master/.github/workflows/trusted-dev.yaml)

This job is opened on every PR to the `components/binding` directory. It check if the author of the PR has write access to the Kyma repo. If it has it sends a `repository-dispatch` event which triggers the main pipeline

## [Ok to test](https://github.com/kyma-project/kyma/blob/master/.github/workflows/ok-to-test.yaml)

This job allows to run the pipeline on PRs made by external contributors. A reviewer has to check whether the PR doesn't contain any malicious code and then put a comment with `/ok-to-test` command and then put the commit hash of the HEAD commit of the PR, eg.

```
/ok-to-test sha=44fb1091e18f025a365495c2d268bce944b4239f
```

Such comment will trigger the main pipeline

## [Pre-master binding k3s](https://github.com/kyma-project/kyma/blob/master/.github/workflows/pre-master-binding-k3s.yml) and [Post-master binding k3s](https://github.com/kyma-project/kyma/blob/master/.github/workflows/post-master-binding-k3s.yml)

Those are the main integration jobs for the Kyma Binding component. Their responsibility is to

- go test, vet, and fmt the code
- build a docker image and push it to gcr
- start k3s and install the Kyma Binding chart with the newly built image
- test if the component works as expected



