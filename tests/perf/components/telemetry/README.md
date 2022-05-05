# Telemetry Fluent Bit Perf Tests

## Overview

Small program to deploy a bunch of log pipelines to a Kubernetes cluster. The following pipelines are deployed:
1. Simple log pipeline that logs to Loki
2. Multiple log pipelines that log to an HTTP host of choice. The upstream host, port, as well as the number of log pipelines can be set via flags. 

## Usage

Example:
```
go run ./... -count=4 -unhealthy-ratio=0.5 -host=mockserver.mockserver -port=1080
```
