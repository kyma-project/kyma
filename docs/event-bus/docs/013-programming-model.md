---
title: Service Programming Model
type: Details
---
# Service Programming Model

## Details

To write an HTTP service in Kyma to handle the event bus published events, The Kyma system will forward the published cloud event to your service in terms of an HTTP request, you can use the request body to get the event payload and the HTTP headers to extract the cloud event and optional tracing information.

### The Cloud Event Object

```yaml
Cloud Event:
  source:                   #Event Source metadata Object
    source_namespace:       #String
    source_type:            #String
    source_environment:     #String
  event_type:               #String
  event_type_version:       #String
  event_time:               #String
  data:                     #Event Payload
extensions:                 #Optional map of Tracing Information
```

### Advanced Response Handling

In the example, a custom cloud event is published on the kyma system:

```yaml
Cloud Event:
  source:
    source_namespace:       "local.kyma.commerce"
    source_type:            "commerce"
    source_environment:     "devel"
  event_type:               "register"
  event_type_version:       "v1"
  event_time:               "1534203066"
  data:                     '{"customer":
                                {"customerID": "1234",
                                 "uid": "rick.sanchez@mail.com"
                                 }
                            }'
extensions:                 #Optional map of Tracing Information
```

```JavaScript
app.post('/v1/events/register', (req, res) => {
  console.log('Request Body Access')
  console.log(req.body.event.customer.customerID)
  console.log(req.body.event.customer.uid)
  
  console.log('Request Headers Access')
  console.log(req.headers['host'])
  console.log(req.headers['user-agent'])
  console.log(req.headers['content-length'])
  console.log(req.headers['content-type'])

  console.log(req.headers['kyma-event-id'])
  console.log(req.headers['kyma-event-time'])
  console.log(req.headers['kyma-event-type'])
  console.log(req.headers['kyma-event-type-version'])

  console.log(req.headers['kyma-source-environment'])
  console.log(req.headers['kyma-source-namespace'])
  console.log(req.headers['kyma-source-type'])

  console.log(req.headers['kyma-subscription'])
  console.log(req.headers['kyma-topic'])

  console.log(req.headers['x-b3-flags'])
  console.log(req.headers['x-b3-parentspanid'])
  console.log(req.headers['x-b3-sampled'])
  console.log(req.headers['x-b3-spanid'])
  console.log(req.headers['x-b3-traceid'])
  console.log(req.headers['x-forwarded-proto'])
  console.log(req.headers['x-request-id'])
  console.log(req.headers['x-envoy-decorator-operation'])
  console.log(req.headers['x-envoy-expected-rq-timeout-ms'])
  console.log(req.headers['x-istio-attributes'])
});
```

The example code logs the original request body and headers. The response is an HTTP 200.

The kyma `kyma-*` headers define the Cloud Event [properties](#the-cloud-event-object)

The tracing B3 `x-b3-*` headers are documented in the *envoy* project docs [here](https://www.envoyproxy.io/docs/envoy/latest/configuration/http_conn_man/headers.html?highlight=headers#http-header-manipulation) and the `x-envoy-*` headers are documented [here](https://www.envoyproxy.io/docs/envoy/latest/configuration/http_filters/router_filter#http-headers-consumed)
