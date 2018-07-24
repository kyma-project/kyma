## Overview
Contains the following components that are used for `Event Bus` End to End tests.

### E2E Tester
The `e2e-tester` directory contains the source code and `Dockerfile` of the test `event-bus-e2e-tester` application. It is used to test `Event Bus` functionality by running a simple orchestrated scenario of spinning up the `event-bus-e2e-subscriber` app, creating a subscription, an event activation, publishing a message and verifying that it reached the `event-bus-e2e-subscriber` app. 

This app is used as part of `kyma core` helm chart tests.

## E2E Subscriber
This is a sample subscriber app that simulates an event subscriber via exposing a set of APIs which can be used to receive and list  events. It is used by the `event-bus-e2e-tester` app as described above.