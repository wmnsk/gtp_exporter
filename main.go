// Copyright 2020 gtp_exporter authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

// +build linux

// Command gtp_exporter implements a Prometheus exporter for Linux kernel GTP.
package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"
)

// Variables for metadata.
var (
	// Version to be recognized in GoReleaser.
	Version = "unset"
	// Revision(short commit hash) to be recognized in GoReleaser.
	Revision = "unset"
)

type gtplogger struct {
	logger log.Logger
}

func (l gtplogger) Println(v ...interface{}) {
	_ = level.Error(l.logger).Log("msg", fmt.Sprint(v...))
}

func init() {
	prometheus.MustRegister(version.NewCollector("gtp_exporter"))
}

func main() {
	var (
		listenAddress = kingpin.Flag(
			"web.listen-address",
			"Address on which to expose metrics and web interface.",
		).Default(":9721").String()
		metricsPath = kingpin.Flag(
			"web.telemetry-path",
			"Path under which to expose metrics.",
		).Default("/metrics").String()
	)

	promlogConfig := &promlog.Config{}
	flag.AddFlags(kingpin.CommandLine, promlogConfig)
	kingpin.Version(version.Print("gtp_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()
	logger := promlog.New(promlogConfig)

	_ = level.Info(logger).Log("msg", "Starting gtp_exporter", "version", version.Info())
	_ = level.Info(logger).Log("msg", "Build context", "build_context", version.BuildContext())

	exporter, err := NewExporter(logger)
	if err != nil {
		_ = level.Error(logger).Log("msg", "Error creating the exporter", "err", err)
		os.Exit(1)
	}
	prometheus.MustRegister(exporter)

	http.Handle(
		*metricsPath,
		promhttp.InstrumentMetricHandler(
			prometheus.DefaultRegisterer,
			promhttp.HandlerFor(
				prometheus.DefaultGatherer,
				promhttp.HandlerOpts{ErrorLog: &gtplogger{logger: logger}},
			),
		),
	)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write(
			[]byte(
				`<html>
					<head><title>GTP Exporter</title></head>
					<body>
						<h1>GTP Exporter</h1>
						<p><a href="` + *metricsPath + `">Metrics</a></p>
					</body>
				</html>`,
			),
		)
		if err != nil {
			_ = level.Error(logger).Log(
				"msg:", "failed to handle http writer",
				"err:", err,
			)
		}
	})

	_ = level.Info(logger).Log("msg", "Listening on address", "address", *listenAddress)
	if err := http.ListenAndServe(*listenAddress, nil); err != nil {
		_ = level.Error(logger).Log("msg", "Error starting HTTP server", "err", err)
		os.Exit(1)
	}
}
