# ufiber-exporter
`ufiber-exporter` is an Prometheus exporter for Ubiquiti UFiber OLTs.  
The goal is to export all available metrics from the OLT and attached ONUs.

## Probing
Devices can be probed by requesting:  
<pre>http://localhost:9777/probe?<b>target=xxx</b></pre>
Configured [options](#options) can be overwritten by using query parameters e.g.  
<pre>http://localhost:9777/probe?target=xxx&<b>export_olt=1&export_onus=0</b></pre>

`target` can be either an address or hostname that is scraped using the globally configured credentials, or the name of a device in the configuration.

For troubleshooting there are also two log levels available:
<pre>http://localhost:9777/probe?target=xxx&<b>debug=1</b></pre>
<pre>http://localhost:9777/probe?target=xxx&<b>trace=1</b></pre>

## Docker image

Docker image is available on Docker Hub, Quay.io and GitHub

<pre>
docker pull swoga/ufiber-exporter
docker pull quay.io/swoga/ufiber-exporter
docker pull ghcr.io/swoga/ufiber-exporter
</pre>

You just need to map your config file into the container at `/etc/ufiber-exporter/config.yml`  
<pre>
docker run -v config.yml:/etc/ufiber-exporter/config.yml swoga/ufiber-exporter
</pre>

## Command line flags
<pre>
--config.file=config.yml
--debug
</pre>

## Configuration file
```yaml
listen: <string> | default = :9777
probe_path: <string> | default = /probe
metrics_path: <string> | default = /metrics
timeout: <int> | default = 60
global: <global>

devices:
  - <device>
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
name: <string>
address: <string>
username: <string> | default = global.username
password: <string> | default = global.password
options: <options> | default = global.options
```
