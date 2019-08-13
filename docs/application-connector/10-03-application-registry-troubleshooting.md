---
title: Application Registry Troubleshooting
type: Troubleshooting
---

<wstep>

## App Registry - No certificate
```
curl -X POST --data @test.json https://gateway.jmedrek-test.cluster.stage.faros.kyma.cx/test-app/v1/metadata/services
```

```
error:1401E410:SSL routines:CONNECT_CR_FINISHED:sslv3 alert handshake failure
```

## App Registry - Expired certificate
```
curl -X POST --data @test.json --cert ./expired/generated.pem https://gateway.jmedrek-test.cluster.stage.faros.kyma.cx/test-app/v1/metadata/services
```

```
error:1401E415:SSL routines:CONNECT_CR_FINISHED:sslv3 alert certificate expired
```

## App Registry - Invalid subject
```
curl -X POST --data @test.json --cert ./expired/generated.pem https://gateway.jmedrek-test.cluster.stage.faros.kyma.cx/another-test-app/v1/metadata/services
```

```
{"code":403,"error":"No valid subject found"}
```

## One Click Integration - Invalid Common Name

https://github.wdf.sap.corp/framefrog/on-call-guides/blob/master/on-call-guide/client-side-issues/wrong-usage/invalid-csr.md

## On-Call guides

https://github.wdf.sap.corp/framefrog/on-call-guides/tree/master/on-call-guide/client-side-issues