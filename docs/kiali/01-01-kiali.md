---
title: Overview
---

Kyma uses [Kiali](https://www.kiali.io) to enable validation, observe the Istio Service Mesh, and provide details on microservices included in the Service Mesh and connections between them.
Kiali offers a set of dashboards and graphs that allow to have the full service mesh at a glance and quickly identify problems and configuration issues.
Some of its features are:
- Observability Features
    - Graphs
        - Health
        - Drill-Down
        - Side-Panel
        - Traffic Animation
        - Graph Types
    - Detail Views
        - Metrics
        - Services
        - Workloads
        - Custom Dashboards
    - Distributed Tracing
- Configuration and Validation Features
    - Istio Configuration
    - Validations
    - Istio Wizards
        - Weighted Routing Wizard
        - Matching Routing Wizard
        - Suspend Traffic Wizard
        - Advanced Options
        - More Wizard examples

For a more detailed feature documentation, please refer to the [official kiali documentation](https://kiali.io/documentation/features/).


>**NOTE:** Kiali is disabled by default. Read [this](/root/kyma/#configuration-custom-component-installation) document for instructions on how to enable it.