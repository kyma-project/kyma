---
title: Enable Kyma with JetStream
---

This guide shows how to enable JetStream and how it ensures `at least once` delivery.

### Enable JetStream

Install Kyma and enable the JetStream flag by running:

```
kyma deploy --value global.jetstream.enabled=true --value global.jetstream.storage=file
```

> **NOTE:** The storage flag set to `file` enables persistence of messages and streams if the NATS server restarts.