# Kyma Binding pipelines

Kyma Binding pipelines are based on Github Actions. This decision is due to the fact that scripts to run k3s are already [developed](https://github.com/kyma-incubator/local-kyma/blob/main/create-cluster-k3s.sh) and k3s is objectively the [best](https://github.com/kyma-incubator/local-kyma#i-see-k3s-k3d-kind-and-minikube---what-should-i-use) to run on CI. Main problem for this approach was that GitHub doesn't propagate the secrets created on the main repository, to pull requests opened from forked repositories. As the workflow in Kyma is based on using forks we had to create a workaround for that. The idea for that was found [here](https://github.com/imjohnbo/ok-to-test). For this purpose, we developed the following jobs:

- [`Is trusted dev`](https://github.com/kyma-project/kyma/blob/master/.github/workflows/trusted-dev.yaml)
- [`Ok to test`](https://github.com/kyma-project/kyma/blob/master/.github/workflows/ok-to-test.yaml)
- [`Pre-master binding k3s`](https://github.com/kyma-project/kyma/blob/master/.github/workflows/pre-master-binding-k3s.yml)
- [`Post-master binding k3s`](https://github.com/kyma-project/kyma/blob/master/.github/workflows/post-master-binding-k3s.yml)

## `Is trusted dev` job

This job is executed on every PR opened in the `components/binding` directory. It checks if the author of the PR has the write access to the `kyma` repository. If the author has the write access, the job sends the `repository-dispatch` event which triggers the main pipeline.

## `Ok to test` job

This job allows you to run the pipeline on PRs made by external contributors. First, you must do a review to check whether the PR doesn't contain any malicious code. If the PR looks good, comment it with the `/ok-to-test` command and then put the commit hash of the HEAD commit of the PR, for example:

```
/ok-to-test sha=44fb1091e18f025a365495c2d268bce944b4239f
```

Such a comment triggers the main pipeline.

## `Pre-master binding k3s` and `Post-master binding k3s` jobs

These are the main integration jobs for the Kyma Binding component. Their responsibility is to:

- Run `go test`, `go vet`, and `go fmt` on the code
- Build a Docker image and push it to GCR
- Start k3s and install the Kyma Binding chart with the newly built image
- Test if the component works as expected
