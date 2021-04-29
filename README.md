# ufiber-exporter
`ufiber-exporter` is an Prometheus exporter for Ubiquiti UFiber OLTs.  
The goal is to export all available metrics from the OLT and attached ONUs.

## Probing
Devices can be probed (when using the defaults) by requesting:  
http://localhost:9777/probe?target=...  
Configured [options](#options) can be overwritten by using query parameters e.g.  
http://localhost:9777/probe?target=...&export_olt=1&export_onus=0

`target` can be either an address or hostname that is scraped using the globally configured credentials, or a key from the `devices:` section of the configuration.

## Docker image

Docker image is available on Docker Hub, Quay.io and GitHub

`docker pull swoga/ufiber-exporter`  
`docker pull quay.io/swoga/ufiber-exporter`  
`docker pull ghcr.io/swoga/ufiber-exporter`

You just need to map your config file into the container at `/etc/ufiber-exporter/config.yml`  
`docker run -v config.yml:/etc/ufiber-exporter/config.yml swoga/ufiber-exporter`

## Command line flags
`ufiber-exporter` requires a path to a YAML config file supplied in the command line flag `--config.file=config.yml`.

`--debug` can be used to raise the log level.

## Configuration file
```yaml
listen: <string> | default = :9777
probe_path: <string> | default = /probe
metrics_path: <string> | default = /metrics
timeout: <int> | default = 60
global: <global>

devices:
  # map key is the name of the target (?target=<string>)
  <string>: <device>
```

### `<global>`
```yaml
username: <string>
password: <string>
options: <options>
```

### `<options>`
```yaml
export_olt: <bool> | default = true
export_onus: <bool> | default = true
export_mac_table: <bool> | default = false
```

### `<device>`
```yaml
address: <string>
username: <string> | default = global.username
password: <string> | default = global.password
options: <options> | default = global.options
```

## Building
You can either build with `go build ./cmd/ufiber-exporter/` or to build with `promu` use `make build`