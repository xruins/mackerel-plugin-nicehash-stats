package main

import (
	"encoding/json"
	"log"
	"os"
	"time"

	flags "github.com/jessevdk/go-flags"
	mp "github.com/mackerelio/go-mackerel-plugin"

	"github.com/xruins/mackerel-plugin-nicehash-stats/pkg/mackerel"
)

type options struct {
	Tempfile         string `short:"t" long:"tempfile" description:"Path to temporary file"`
	MetricsKeyPrefix string `short:"m" long:"metrics-key-prefix" description:"Prefix for Mackerel metrics key"`

	CurrencyCode           string        `short:"c" long:"currency-code" default:"USD" description:"Code of currency (such as USD and JPY)"`
	CurrencyUpdateInterval time.Duration `short:"i" long:"currency-update-interval" default:"1h" description:"Interval to update currency price"`

	NiceHashAPIKey         string `short:"k" long:"api-key" description:"API key for NiceHash API"`
	NiceHashAPISecret      string `short:"s" long:"api-secret" description:"API secret for NiceHash API"`
	NiceHashOrganizationID string `short:"o" long:"organization-id" description:"OrganizationID for NiceHash API"`

	ServiceMetricMode bool `short:"e" long:"service-metric-mode" description:"Flag to output metric for ServiceMetric"`
}

type ServiceMetric struct {
	Name  string  `json:"name"`
	Time  int64   `json:"time"`
	Value float64 `json:"value"`
}

func main() {
	var opts options

	_, err := flags.Parse(&opts)
	if err != nil {
		panic(err)
	}

	plugin := mackerel.Plugin{
		NiceHashAPIKey:         opts.NiceHashAPIKey,
		NiceHashAPISecret:      opts.NiceHashAPISecret,
		NiceHashOrganizationID: opts.NiceHashOrganizationID,
		CurrencyCode:           opts.CurrencyCode,
		CurrencyUpdateInterval: opts.CurrencyUpdateInterval,
		Tempfile:               opts.Tempfile,
		ServiceMetricMode:      opts.ServiceMetricMode,
	}

	helper := mp.NewMackerelPlugin(&plugin)
	helper.Tempfile = plugin.Tempfile

	// output as service metrics
	if opts.ServiceMetricMode {
		metrics, err := helper.FetchMetrics()
		if err != nil {
			log.Fatalf("failed to get metrics. err: %s", err)
		}

		now := time.Now().Unix()
		out := make([]*ServiceMetric, 0, len(metrics))

		for key, value := range metrics {
			sm := &ServiceMetric{
				Name:  key,
				Value: value,
				Time:  now,
			}
			out = append(out, sm)
		}

		enc := json.NewEncoder(os.Stdout)
		err = enc.Encode(out)
		if err != nil {
			log.Fatalf("failed to encode JSON. err: %s", err)
		}

		os.Exit(0)
	}

	// output as host metrics
	helper.Run()
}
