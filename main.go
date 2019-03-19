// Copyright 2018 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/minyk/prometheus-sd-dcosl4lb/adapter"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/discovery/targetgroup"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	a             = kingpin.New("sd adapter usage", "Tool to generate file_sd target files for DC/OS L4LB SD mechanisms.")
	l4lbPrefix    = a.Flag("l4lb.prefix", "Prefix of DC/OS L4LB domain name for prometheus exporters. e.g. <l4lb.prefix>.test.marathon.l4lb.thisdcos.directory.").Default("prometheus").String()
	frameworkName = a.Flag("l4lb.framework", "Framework name part of DC/OS L4LB domain name. e.g. prometheus.test.<l4lb.framework>.l4lb.thisdcos.directory. To capture any frameworks, use \"*\" for name.").Default("*").String()
	outputFile    = a.Flag("output.file", "Output file for file_sd compatible file.").Default("custom_sd.json").String()
	listenAddress = a.Flag("listen.address", "The address the DC/OS L4LB HTTP API is listening on for requests.").Default("localhost:62080").String()
	logger        log.Logger

	// tagsLabel is the name of the label containing the tags assigned to the target.
	marathonLabel = "framework_name"
	serviceLabel = "service_id"
)

// this struct represents the response from a DC/OS L4LB request.
type L4LBService struct {
	Backend     []struct {
		Ip      string
		Port    int
	}
	Protocol    string
	Vip         string
}

// Note: create a config struct for your custom SD type here.
type sdConfig struct {
	Prefix          string
	Framework       string
	Address         string
	TagSeparator    string
	RefreshInterval int
}

// Note: This is the struct with your implementation of the Discoverer interface (see Run function).
// Discovery retrieves target information from a Consul server and updates them via watches.
type discovery struct {
	prefix          string
	framework       string
	address         string
	refreshInterval int
	tagSeparator    string
	logger          log.Logger
	oldSourceList   map[string]bool
	dcosName        string
	l4lbName        string
}

func (d *discovery) parseServiceNodes(service *L4LBService) (*targetgroup.Group, error) {
	tgroup := targetgroup.Group{
		Source: "dcos-l4lb",
		Labels: make(model.LabelSet),
	}

	var frameworkName string
	var serviceName string
	domainNames := strings.Split(service.Vip, ".")

	// first, test tempNames has "thisdcos" and "l4lb" keyword.
	if domainNames[len(domainNames)-2] == d.dcosName && domainNames[len(domainNames)-3] == d.l4lbName {
		// then test framework name and prefix.
		if (domainNames[len(domainNames)-4] == d.framework || d.framework == "*") && strings.HasPrefix(domainNames[0], d.prefix) {
			//join domainNames for service name
			frameworkName = domainNames[len(domainNames)-4]
			serviceName = strings.Join(domainNames[0:len(domainNames)-4], ".")
		} else {
			level.Debug(d.logger).Log("msg", "vip is ignored", "vip", service.Vip)
			return nil, nil
		}
	} else {
		level.Debug(d.logger).Log("msg", "vip is wrong", "vip", service.Vip)
		return nil, nil
	}

	for _, node := range service.Backend {
		var addr string
		addr = net.JoinHostPort(node.Ip, fmt.Sprintf("%d", node.Port))

		target := model.LabelSet{model.AddressLabel: model.LabelValue(addr)}
		labels := model.LabelSet{
			model.AddressLabel:                   model.LabelValue(service.Vip),
			model.LabelName(marathonLabel):       model.LabelValue(frameworkName),
			model.LabelName(serviceLabel):           model.LabelValue(serviceName),
		}
		tgroup.Labels = labels
		tgroup.Targets = append(tgroup.Targets, target)
	}
	return &tgroup, nil
}

// Note: you must implement this function for your discovery implementation as part of the
// Discoverer interface. Here you should query your SD for it's list of known targets, determine
// which of those targets you care about (for example, which of Consuls known services do you want
// to scrape for metrics), and then send those targets as a target.TargetGroup to the ch channel.
func (d *discovery) Run(ctx context.Context, ch chan<- []*targetgroup.Group) {
	for c := time.Tick(time.Duration(d.refreshInterval) * time.Second); ; {
		resp, err := http.Get(fmt.Sprintf("http://%s/v1/vips", d.address))

		if err != nil {
			level.Error(d.logger).Log("msg", "Error getting services list", "err", err)
			time.Sleep(time.Duration(d.refreshInterval) * time.Second)
			continue
		}

		var nodes []*L4LBService
		dec := json.NewDecoder(resp.Body)
		err = dec.Decode(&nodes)
		resp.Body.Close()

		if err != nil {
			level.Error(d.logger).Log("msg", "Error reading services list", "err", err)
			time.Sleep(time.Duration(d.refreshInterval) * time.Second)
			continue
		}

		var tgs []*targetgroup.Group
		// Note that we treat errors when querying specific consul services as fatal for this
		// iteration of the time.Tick loop. It's better to have some stale targets than an incomplete
		// list of targets simply because there may have been a timeout. If the service is actually
		// gone as far as consul is concerned, that will be picked up during the next iteration of
		// the outer loop.

		level.Debug(d.logger).Log("msg", "length of nodes", "value", len(nodes))

		for _, service := range nodes {
			if service.Protocol == "tcp" && strings.HasPrefix(service.Vip, d.prefix) {
				tg, err := d.parseServiceNodes(service)
				if err != nil {
					level.Error(d.logger).Log("msg", "Error parsing services nodes", "service", "err", err)
					break
				}
				if tg != nil {
					tgs = append(tgs, tg)
				}
			}
		}

		level.Debug(d.logger).Log("msg", "length of tgs", "value", len(tgs))

		if err == nil {
			// We're returning all Consul services as a single targetgroup.
			ch <- tgs
		}
		// Wait for ticker or exit when ctx is closed.
		select {
		case <-c:
			continue
		case <-ctx.Done():
			return
		}
	}
}

func newDiscovery(conf sdConfig) (*discovery, error) {
	cd := &discovery{
		prefix:          conf.Prefix,
		framework:       conf.Framework,
		address:         conf.Address,
		refreshInterval: conf.RefreshInterval,
		tagSeparator:    conf.TagSeparator,
		logger:          logger,
		oldSourceList:   make(map[string]bool),
		dcosName:        "thisdcos",
		l4lbName:        "l4lb",
	}
	return cd, nil
}

func main() {
	a.HelpFlag.Short('h')

	_, err := a.Parse(os.Args[1:])
	if err != nil {
		fmt.Println("err: ", err)
		return
	}
	logger = log.NewSyncLogger(log.NewLogfmtLogger(os.Stdout))
	logger = log.With(logger, "ts", log.DefaultTimestampUTC, "caller", log.DefaultCaller)

	ctx := context.Background()

	// NOTE: create an instance of your new SD implementation here.
	cfg := sdConfig{
		TagSeparator:    ",",
		Address:         *listenAddress,
		RefreshInterval: 30,
		Prefix:          *l4lbPrefix,
		Framework:       *frameworkName,
	}

	disc, err := newDiscovery(cfg)
	if err != nil {
		fmt.Println("err: ", err)
	}
	sdAdapter := adapter.NewAdapter(ctx, *outputFile, "dcosl4lbSD", disc, logger)
	sdAdapter.Run()

	<-ctx.Done()
}
