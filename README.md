# ufiber-exporter
`ufiber-exporter` is an Prometheus exporter for Ubiquiti UFiber OLTs.  
The goal is to export all available metrics from the OLT and attached ONUs.

## WIP
This tool is work in progress! Metrics for ONUs are missing.

## Command line flags
`ufiber-exporter` requires a path to a YAML config file supplied in the command line flag `--config.file=config.yml`.

`--debug` can be used to raise the log level.

## Configuration file
```yaml
listen: <string> | default = :9100
probe_path: <string> | default = /probe
metrics_path: <string> | default = /metrics
timeout: <int> | default = 60

devices:
  # map key is the name of the target (?target=<string>)
  <string>: <device>
```

### `<device>`
```yaml
address: <string>
username: <string>
password: <string>
```

## Building
You can either build with `go build ./cmd/ufiber-exporter/` or to build with `promu` use `make build`