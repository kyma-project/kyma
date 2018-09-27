---
title: Event Service
type: Architecture
---

## Overview

Event Service is responsible for publishing events in Kyma.

##API

Event Service exposes API for publishing events to the [Even Bus](https://github.com/kyma-project/kyma/tree/master/docs/event-bus/docs).
It verifies event payload and proxies the event to the Event Bus.

For a complete information on sending events, please see [Sending Events to Kyma Guide](TODO).