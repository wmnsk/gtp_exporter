# GTP Exporter

A Prometheus exporter for [Linux kernel GTP-U](https://www.kernel.org/doc/html/latest/networking/gtp.html).

[![CircleCI](https://circleci.com/gh/wmnsk/gtp_exporter.svg?style=shield)](https://circleci.com/gh/wmnsk/gtp_exporter)
[![GoDoc](https://godoc.org/github.com/wmnsk/gtp_exporter?status.svg)](https://godoc.org/github.com/wmnsk/gtp_exporter)
[![LICENSE](https://img.shields.io/github/license/mashape/apistatus.svg)](https://github.com/wmnsk/gtp_exporter/blob/master/LICENSE)

GTP Exporter retieves data from Linux GTP kernel driver using netlink, and exports metrics with them. No other implementations of GTP-U nor other platforms are supported.

## Getting started

### Prerequisites

See `go.mod` for what packages this tool depends on, or just run `go mod tidy` to get all the required packages.

### Usage

```
$ ./gtp_exporter -h
usage: gtp_exporter [<flags>]

Flags:
  -h, --help               Show context-sensitive help (also try --help-long and --help-man).
      --web.listen-address=":9721"
                           Address on which to expose metrics and web interface.
      --web.telemetry-path="/metrics"
                           Path under which to expose metrics.
      --log.level=info     Only log messages with the given severity or above. One of: [debug, info, warn, error]
      --log.format=logfmt  Output format of log messages. One of: [logfmt, json]
      --version            Show application version.
```

## Supported features

### Metrics

| Name        | Description                                                          | Labels                            |
|-------------|----------------------------------------------------------------------|-----------------------------------|
| gtp_up      | Whether the last query is successful.                                | -                                 |
| gtp_tunnels | The number of existing tunnels.                                      | version, peer                     |
| gtp_devices | The number of existing GTP devices.                                  | name, role                        |
| gtp_info    | Some of the information of GTP kernel module retrieved by `modinfo`. | filename, description, srcversion |

Here's a example output from a sample S-GW that has GTP devices on S1-U and S5-U interfaces.

```
# HELP gtp_devices The number of existing GTP devices.
# TYPE gtp_devices gauge
gtp_devices{name="s1u",role="GGSN"} 1 // should be "SGSN". Perhaps a bug in kernel GTP or netlink package...?
gtp_devices{name="s5u",role="GGSN"} 1
# HELP gtp_exporter_build_info A metric with a constant '1' value labeled by version, revision, branch, and goversion from which gtp_exporter was built.
# TYPE gtp_exporter_build_info gauge
gtp_exporter_build_info{branch="",goversion="go1.14.1",revision="",version=""} 1
# HELP gtp_info The information of GTP kernel module.
# TYPE gtp_info gauge
gtp_info{description="Interface driver for GTP encapsulated traffic",filename="/lib/modules/5.7.0-1.el7.elrepo.x86_64/kernel/drivers/net/gtp.ko",srcversion="191407DA5399304D93D62C7"} 1
# HELP gtp_tunnels The number of existing tunnels.
# TYPE gtp_tunnels gauge
gtp_tunnels{peer="192.168.200.1",version="1"} 1
gtp_tunnels{peer="192.168.200.5",version="1"} 1
```

### Visualization

_(not available yet)_

## Author(s)

Yoshiyuki Kurauchi ([Website](https://wmnsk.com/) / [Twitter](https://twitter.com/wmnskdmms))

## LICENSE

[MIT](https://github.com/wmnsk/gtp_exporter/blob/master/LICENSE)
