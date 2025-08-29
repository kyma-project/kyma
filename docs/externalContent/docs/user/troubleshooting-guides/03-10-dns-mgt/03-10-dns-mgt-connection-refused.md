# Connection Refused or Timeout

## Symptom

After you have finished all the steps required to [set up your custom domain](../../tutorials/01-10-setup-custom-domain-for-workload.md), you receive the `connection refused` or `connection timeout` error when you try to expose a Service. It shows up when you call the Service endpoint by sending a GET request. The error looks as follows:

```txt
curl: (7) Failed to connect to httpbin.mydomain.com port 443: Connection refused
```

## Cause

DNS resolves to an incorrect IP address.

## Solution

Check if the IP address provided as the value of the **spec.targets** parameter of the DNSEntry custom resource is the IP address of the Ingress Gateway you are using. To check the Ingress Gateway IP, run:

```bash
kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.status.loadBalancer.ingress[0].ip}'`
```

In addition, ensure that your OS resolves the target host name to the same Ingress Gateway IP address.
Run:

```bash
host {YOUR_SUBDOMAIN}
```
> [!NOTE]
> `YOUR_SUBDOMAIN` specifies the name of your subdomain, for example, `httpbin.mydomain.com`.