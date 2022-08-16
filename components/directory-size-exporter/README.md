# Directory Size Exporter

## Overview

The directory size exporter is monitoring operators storage by collecting information about the size of top level directories as metrics. Meanwhile, Prometheus will pull the metrics. 

The following configuration options are available:
* `log-format` - Log format parameter for logger (`json` or `text`)
* `log-level` - Log level parameter for logger (`debug`, `info`, `warn`, `error`, `fatal`)
* `storage-path` - the path to the root directory to be observed by the exporter (default `/data/log/flb-storage/`)
* `metric-name` - the metric name to use for the metric exposure (default `telemetry_fsbuffer_usage_bytes`)
* `port` - the port under which the metrics will be exposed in the prometheus format (default `2021`)
* `interval` - how frequently we record to prometheus in seconds (default `30`)


## Development

### Available Commands

For development, you can use the following commands:

- Run all tests and validation

```bash
make
```

- Run the exporter locally

```bash
make run-local
```
