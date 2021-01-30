package mackerel

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	mp "github.com/mackerelio/go-mackerel-plugin"
	"golang.org/x/xerrors"

	"github.com/xruins/mackerel-plugin-nicehash-stats/pkg/coindesk"
	"github.com/xruins/mackerel-plugin-nicehash-stats/pkg/nicehash"
)

var graphdef = map[string]mp.Graphs{
	"#": {
		Label: "NiceHash Profitabilities",
		Unit:  "float",
		Metrics: []mp.Metrics{
			{Name: "profitability", Label: "Profitability", Diff: false},
		},
	},
}

type Plugin struct {
	// plugin-specific fields
	NiceHashAPIKey         string
	NiceHashAPISecret      string
	NiceHashOrganizationID string
	CurrencyCode           string
	CurrencyUpdateInterval time.Duration

	// mackerel-plugin common fields
	Tempfile string

	MetricPrefix      string
	ServiceMetricMode bool
}

func (p *Plugin) FetchMetrics() (map[string]float64, error) {
	ctx := context.Background()

	nhcl, err := nicehash.NewClient(
		nicehash.APIBaseURL,
		p.NiceHashOrganizationID,
		p.NiceHashAPIKey,
		p.NiceHashAPISecret,
	)
	if err != nil {
		return nil, xerrors.Errorf("failed to create Nicehash client: %w", err)
	}

	// get rigs information
	res, err := nhcl.GetRigs2(ctx)
	if err != nil {
		return nil, xerrors.Errorf("failed to get rigs information: %w", err)
	}

	// get BTC price
	btcPrice, err := p.fetchCurrency(ctx)
	if err != nil {
		return nil, xerrors.Errorf("failed to get BTC price: %w", err)
	}

	priceMap := make(map[string]float64, len(res.MiningRigs))
	// convert profitability to specified currency
	for _, rig := range res.MiningRigs {
		name := p.metricName(rig.Name)
		priceMap[name] = rig.Profitability * btcPrice
	}

	// calculate total profitability of rigs
	var sum float64
	for _, prof := range priceMap {
		sum += prof
	}
	priceMap[p.metricName("Total")] = sum

	return priceMap, nil
}

func (p *Plugin) metricName(name string) string {
	if p.ServiceMetricMode {
		return fmt.Sprintf("nicehash.profitability.%s", name)
	}
	return fmt.Sprintf("%s.profitability", name)
}

func (p *Plugin) GraphDefinition() map[string]mp.Graphs {
	return graphdef
}

func (p *Plugin) MetricKeyPrefix() string {
	if p.MetricPrefix == "" {
		return "nicehash"
	}
	return p.MetricPrefix
}

type cacheCurrency struct {
	CurrencyCode string
	Price        float64
	LastFetch    time.Time
}

func (p *Plugin) fetchCurrency(ctx context.Context) (float64, error) {
	if p.Tempfile != "" {
		price, err := p.loadCurrencyCache()
		if err != nil {
			return 0, xerrors.Errorf("failed to load currencyCache: %w", err)
		}
		if price != 0 {
			return price, nil
		}
	}

	cdcl, err := coindesk.NewClient(p.CurrencyCode)
	if err != nil {
		return 0, xerrors.Errorf("failed to create Coindesk client: %w", err)
	}

	// get BTC to currency price
	btcPrice, err := cdcl.GetPrice(ctx)
	if err != nil {
		return 0, xerrors.Errorf("failed to get price of BTC: %w", err)
	}

	if p.Tempfile != "" {
		err = p.saveCurrencyCache(btcPrice)
		if err != nil {
			return 0, xerrors.Errorf("failed to save currencyCache: %w", err)
		}
	}

	return btcPrice, nil
}

func (p *Plugin) loadCurrencyCache() (float64, error) {
	f, err := os.Open(p.Tempfile)
	if os.IsNotExist(err) {
		return 0, nil
	} else if err != nil {
		return 0, xerrors.Errorf("failed to open cache file: %w", err)
	}
	defer f.Close()

	decoder := json.NewDecoder(f)
	cache := &cacheCurrency{}
	decoder.Decode(cache)

	cacheExpiryTime := cache.LastFetch.Add(p.CurrencyUpdateInterval)
	if p.CurrencyCode == cache.CurrencyCode && time.Now().Before(cacheExpiryTime) {
		return cache.Price, nil
	}

	return 0, nil
}

func (p *Plugin) saveCurrencyCache(price float64) error {
	f, err := os.Create(p.Tempfile)
	if err != nil {
		return xerrors.Errorf("failed to open cache file: %w", err)
	}
	defer f.Close()

	cache := &cacheCurrency{
		CurrencyCode: p.CurrencyCode,
		Price:        price,
		LastFetch:    time.Now(),
	}

	decoder := json.NewEncoder(f)
	decoder.Encode(cache)

	return nil
}
