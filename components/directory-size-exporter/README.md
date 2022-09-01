# Directory Size Exporter

## Overview

The directory size exporter is a typical metrics exporter in the Prometheus format. It is meant to run as a sidecar to an application to watch a specific application file storage and export metrics about that storage. At the moment, it is used only as a sidecar for the `telemetry-fluent-bit` instances to watch the file buffer sizes.

The following configuration options are available:
- **log-format** - log format parameter for the logger (`json` or `text`)
- **log-level** - log level parameter for the logger (`debug`, `info`, `warn`, `error`, `fatal`)
- **storage-path** - the path to the root directory to be observed by the exporter (default `/data/log/flb-storage/`)
- **metric-name** - the metric name to use for the metric exposure (default `telemetry_fsbuffer_usage_bytes`)
- **port** - the port under which the metrics will be exposed in the Prometheus format (default `2021`)
- **interval** - interval in seconds at which the exporter should check the storage (default `30`)


## Development

### Available Commands

For development, use the following commands:

- Run all tests and validation:

```bash
make
```

- Run the exporter locally:

```bash
make run-local
```
