package coindesk

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-playground/validator"
	"golang.org/x/xerrors"
)

type Client struct {
	httpClient   *http.Client
	url          string
	CurrencyCode string
}

var validate = validator.New()

func NewClient(currencyCode string) (*Client, error) {
	client := &Client{
		httpClient:   http.DefaultClient,
		CurrencyCode: currencyCode,
		url:          coindeskAPIURL,
	}

	return client, nil
}

const (
	coindeskAPIURL          = `https://api.coindesk.com`
	coindeskAPICurrentPrice = `/v1/bpi/currentprice/%s.json`
)

var ErrCurrencyCodeNotFound = errors.New("currency does not found on coindesk API result")

func (c *Client) GetPrice(ctx context.Context) (float64, error) {
	u := c.url + fmt.Sprintf(coindeskAPICurrentPrice, c.CurrencyCode)

	req, err := http.NewRequest(
		http.MethodGet,
		u,
		nil,
	)
	if err != nil {
		xerrors.Errorf("failed to create request: %w", err)
	}

	req = req.WithContext(ctx)

	res, err := c.httpClient.Do(req)
	if err != nil {
		xerrors.Errorf("failed to invoke API: %w", err)
	}
	defer res.Body.Close()

	cp := &CurrentPrice{}
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(cp)
	if err != nil {
		return 0, xerrors.Errorf("failed to decode JSON: %w", err)
	}

	for currencyCode, bpi := range cp.Bpi {
		if currencyCode == c.CurrencyCode {
			return bpi.RateFloat, nil
		}
	}

	return 0, ErrCurrencyCodeNotFound
}
