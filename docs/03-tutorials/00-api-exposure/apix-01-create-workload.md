---
title: Create a workload
---

The tutorial comes with a sample HttpBin service deployment and a sample Function.

## Steps

Follow these steps to deploy an instance of the HttpBin service or a sample Function.

1. Create a Namespace and export its value as an environment variable. Skip the step if you already have a Namespace. Run:

   ```bash
   export NAMESPACE={NAMESPACE_NAME}
   kubectl create ns $NAMESPACE
   kubectl label namespace $NAMESPACE istio-injection=enabled --overwrite
   ```

<div tabs>

  <details>
  <summary>
  HttpBin
  </summary>

2. Deploy an instance of the HttpBin service in your Namespace:

   ```bash
   kubectl -n $NAMESPACE create -f https://raw.githubusercontent.com/istio/istio/master/samples/httpbin/httpbin.yaml
   ```

  </details>

  <details>
  <summary>
 Function
  </summary>

2. Create a Function in your Namespace using the [supplied code](./assets/function.yaml):

   ```shell
   kubectl -n $NAMESPACE apply -f https://raw.githubusercontent.com/kyma-project/kyma/main/docs/03-tutorials/assets/function.yaml
   ```

  </details>
</div>


## Next steps

Once you have your workload deployed, you can continue by choosing one of the following tutorials:

- [Set up a custom domain for a workload](./apix-01-setup-custom-domain-for-workload.md)
- [Expose a workload](./apix-02-expose-workload-apigateway.md)
- [Expose multiple workloads on the same host](./apix-03-expose-multiple-workloads.md)
- [Expose and secure a workload with OAuth2](./apix-04-expose-and-secure-workload-oauth2.md)
- [Expose and secure a workload with Istio](./apix-05-expose-and-secure-workload-istio.md)
- [Expose and secure a workload with JWT](./apix-05-expose-and-secure-workload-jwt.md)
