// Copyright 2020-2023 gtp_exporter authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

//go:build linux
// +build linux

package main

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/vishvananda/netlink"
)

// Namespace defines the common namespace to be used by all metrics.
const namespace = "gtp"

var (
	up = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "up"),
		"Reports whether the last query is successful.",
		nil, nil,
	)
	tunnels = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "tunnels"),
		"The number of existing tunnels.",
		[]string{"version", "peer"}, nil,
	)
	devices = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "devices"),
		"The number of existing GTP devices.",
		[]string{"name", "role"}, nil,
	)
	info = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "info"),
		"The information of GTP kernel module.",
		[]string{"filename", "description", "srcversion"}, nil,
	)
)

// Option is an option function to set optional configurations to Exporter.
type Option func(*Exporter)

// SetLogger sets given log.Logger as a logger of Exporter.
func SetLogger(l log.Logger) Option {
	return func(exp *Exporter) {
		exp.logger = l
	}
}

// Exporter implements the prometheus.Exporter interface.
type Exporter struct {
	logger log.Logger
}

// NewExporter creates a new Exporter.
//
// Put arbitrary number of Option to update Exporter at creation.
func NewExporter(logger log.Logger, opts ...Option) *Exporter {
	e := &Exporter{
		logger: logger,
	}

	for _, opt := range opts {
		opt(e)
	}

	return e
}

// Describe sends the super-set of all possible descriptors of metrics collected by
// this Exporter to the provided channel and returns once the last descriptor has
// been sent.
//
// See prometheus.Exporter.Describe for details.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- up
	ch <- tunnels
	ch <- devices
	ch <- info
}

// Collect is called by the Prometheus registry when collecting metrics.
//
// See prometheus.Exporter.Collect for details.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	ok := true
	defer func() {
		if !ok {
			ch <- prometheus.MustNewConstMetric(
				up, prometheus.GaugeValue, 0.0,
			)
		}
	}()

	if err := e.collectTunnels(ch); err != nil {
		_ = e.logger.Log("failed to collect gtp_tunnels: %v", err)
		ok = false
	}
	if err := e.collectDevices(ch); err != nil {
		_ = e.logger.Log("failed to collect gtp_devicess: %v", err)
		ok = false
	}
	if err := e.collectModuleInfo(ch); err != nil {
		_ = e.logger.Log("failed to collect gtp_module_info: %v", err)
		ok = false
	}

	ch <- prometheus.MustNewConstMetric(
		up, prometheus.GaugeValue, 1.0,
	)
}

func (e *Exporter) collectTunnels(ch chan<- prometheus.Metric) error {
	pdps, err := netlink.GTPPDPList()
	if err != nil {
		return err
	}

	for _, pdp := range pdps {
		ch <- prometheus.MustNewConstMetric(
			tunnels, prometheus.GaugeValue, 1, strconv.Itoa(int(pdp.Version)), pdp.PeerAddress.String(),
		)
	}
	return nil
}

func (e *Exporter) collectDevices(ch chan<- prometheus.Metric) error {
	links, err := netlink.LinkList()
	if err != nil {
		return err
	}

	for _, link := range links {
		g, ok := link.(*netlink.GTP)
		if !ok {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			devices, prometheus.GaugeValue, 1, g.Name, roleToString(g.Role),
		)
	}
	return nil
}

func (e *Exporter) collectModuleInfo(ch chan<- prometheus.Metric) error {
	file, err := exec.Command("modinfo", "-F", "filename", "gtp").Output()
	if err != nil {
		return err
	}
	f := strings.TrimSuffix(string(file), "\n")

	desc, err := exec.Command("modinfo", "-F", "description", "gtp").Output()
	if err != nil {
		return err
	}
	d := strings.TrimSuffix(string(desc), "\n")

	srcver, err := exec.Command("modinfo", "-F", "srcversion", "gtp").Output()
	if err != nil {
		return err
	}
	s := strings.TrimSuffix(string(srcver), "\n")

	ch <- prometheus.MustNewConstMetric(
		info, prometheus.GaugeValue, 1, f, d, s,
	)
	return nil
}

func roleToString(role int) string {
	switch role {
	case 0:
		return "GGSN"
	case 1:
		return "SGSN"
	default:
		return fmt.Sprintf("invalid role: %d", role)
	}
}
