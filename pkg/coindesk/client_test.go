package coindesk

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"testing"
)

func mockGetPrice(w http.ResponseWriter, r *http.Request, currency string) {
	fname := fmt.Sprintf("testing/%s.json", currency)
	f, err := os.Open(fname)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	b, err := ioutil.ReadAll(f)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(b)
}

// currentpriceRegexp is a regexp to match string as `/v1/bpi/currentprice/JPY.json`.
// it provides curency code as first submatch.
var currentpriceRegexp = regexp.MustCompile(`\/v1\/bpi\/currentprice\/([A-Z]{3})\.json`)

func mockCoindeskHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// mock v1/bpi/currentprice/{currency_code}.json
	if submatches := currentpriceRegexp.FindAllStringSubmatch(path, -1); len(submatches) > 0 {
		mockGetPrice(w, r, submatches[0][1])
		return
	}

	return
}

func TestClient_GetPrice(t *testing.T) {
	testServer := httptest.NewServer(
		http.HandlerFunc(
			mockCoindeskHandler,
		),
	)
	defer testServer.Close()

	type pattern struct {
		currencyCode string
		want         float64
	}

	patterns := []*pattern{
		&pattern{
			currencyCode: "JPY",
			want:         3489524.2986,
		},
		&pattern{
			currencyCode: "USD",
			want:         33330.3817,
		},
	}

	for _, pattern := range patterns {
		client := &Client{
			httpClient:   http.DefaultClient,
			url:          testServer.URL,
			CurrencyCode: pattern.currencyCode,
		}

		got, err := client.GetPrice(context.Background())
		if err != nil {
			t.Fatalf("failed to get price. err: %s", err)
		}

		if got != pattern.want {
			t.Errorf("unmatched price. got: %f, want: %f", got, pattern.want)
		}
	}
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
