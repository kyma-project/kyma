# Directory Size Exporter

## Overview

The directory size exporter is monitoring operators storage by collecting information about the size of top level directories as metrics. Meanwhile, Prometheus will pull the metrics. 

One can path specific values for such things like path to the directory one want to monitor or prometheus metrics as parameters. 
The list of parameters goes as follows:
* `log-format` - Log format parameter for logger (`json` or `text`)
* `log-level` - Log level parameter for logger (`debug`, `info`, `warn`, `error`, `fatal`)
* `storage-path` - the path to the root directory which data we want to observe (default `/data/log/flb-storage/`)
* `metric-name` - which name we use for our metrics in prometheus (default `telemetry_fsbuffer_usage_bytes`)
* `port` - which port application listens (default `2021`)
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
