---
title: Send events from outside the Kyma cluster
---

This guide shows how to send events from outside the Kyma cluster. Additionally, this tutorial includes commands for sending events of different [types](../../05-technical-reference/evnt-01-event-names.md).

## Prerequisites

- Open [Kyma Dashboard](../../02-get-started/01-quick-install.md#open-kyma-dashboard).
- Install the [one-click integration script](https://github.com/janmedrek/one-click-integration-script) to retrieve certificates.

## Steps

### Connect an external Application to Kyma

1. In the Kyma Dashboard, go to **Integration > Application**.
2. Click on **Create Application** and enter `externalapp` as the Application name. (For the purposes of this tutorial, `externalapp` represents an external solution connected to Kyma.)
3. Click on the created Application, and choose **Connect Application**.
4. Copy the token by clicking on **Copy to clipboard**. You will use the value of the token in the next steps.

### Create certificates

Paste the token obtained in the previous step and execute this script:

```bash
token='{PASTE_TOKEN_HERE}'
### e.g. token='https://connector-service.foo.test.kyma-project.dev/v1/applications/signingRequests/info?token=sampletoken'
one-click-integration.sh -u "${token}"
```

### Use the certificates generated to send an event to Kyma

1. Find out the address of the Kyma Gateway:

    ```bash
    host=$(kubectl get vs -n kyma-integration connector-service-mtls -ojsonpath='{ .spec.hosts[0] }')
    ```

2. Send the request to the gateway to publish:

    > **NOTE:** In the following commands, `@-` instructs `curl` to read the data for the body of the request from STDIN. Alternatively, you could read the content of the body from a file. For example, if the JSON payload for the request body is stored in `/tmp/msg.json`, you could use `--data @/tmp/msg.json`.


<div tabs name="Use the generated certificates to send an event">
  <details>
  <summary label="Cloud Event (structured mode)">
  Cloud Event (structured mode)
  </summary>

Use the generated certificates to send a Cloud Event in structured mode:

```bash
curl -v --cert generated.crt --key generated.key -X POST "https://${host}/externalapp/events" \
 -H "Content-Type: application/cloudevents+json" \
 --data @- << EOF
{
    "specversion": "1.0",
    "source": "/sourcename",
    "type": "sap.kyma.custom.externalapp.order.created.v1",
    "eventtypeversion": "v1",
    "id": "A234-1234-1234",
    "data" : "{\"foo1\":\"bar1\"}",
    "datacontenttype":"application/json"
}
EOF
```

The Target URL for publishing Cloud Events can be `https://${host}/{APPLICATION_NAME}/events` or `https://${host}/{APPLICATION_NAME}/v2/events`.

  </details>
  <details>
  <summary label="Cloud Event (binary mode)">
  Cloud Event (binary mode)
  </summary>

Use the generated certificates to send a Cloud Event in binary mode:

```bash
curl -v --cert generated.crt --key generated.key -X POST "https://${host}/externalapp/events" \
-H "Content-Type: application/json" \
-H "ce-specversion: 1.0" \
-H "ce-source: /sourcename" \
-H "ce-type: sap.kyma.custom.externalapp.order.created.v1" \
-H "ce-eventtypeversion: v1" \
-H "ce-id: A234-1234-1234" \
--data @- << EOF
"{\"foo2\":\"bar2\"}"
EOF
```

The target URL for publishing Cloud Events can be `https://${host}/{APPLICATION_NAME}/events` or `https://${host}/{APPLICATION_NAME}/v2/events`.

  </details>
  <details>
  <summary label="Legacy event">
  Legacy event
  </summary>

Use the generated certificates to send a legacy event:

```bash
curl -v --cert generated.crt --key generated.key -X POST "https://${host}/externalapp/v1/events" \
-H "Content-Type: application/json" \
--data @- << EOF
{
"event-type": "order.created",
"event-type-version": "v1",
"event-time": "$(date -u +'%Y-%m-%dT%H:%M:%SZ')",
"data" : "{\"foo3\":\"bar3\"}"
}
EOF
```

The target URL for publishing legacy events must be `https://${host}/{APPLICATION_NAME}/v1/events`.

  </details>
</div>

