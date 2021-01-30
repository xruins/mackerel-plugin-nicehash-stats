package nicehash

import (
	"context"
	"net/http"
	"os"
	"testing"
)

func TestClient_generateHMACSignature(t *testing.T) {
	type testcase struct {
		apiBaseURL     string
		method         string
		spath          string
		query          string
		organizationID string
		apiKey         string
		apiSecret      string
		nonce          string
		time           string
		body           []byte
		want           string
	}

	tcs := []testcase{
		testcase{
			apiBaseURL:     APIBaseURL,
			method:         http.MethodGet,
			spath:          "/hashpower/orderBook",
			query:          "?algorithm=X16R&page=0&size=100",
			organizationID: "da41b3bc-3d0b-4226-b7ea-aee73f94a518",
			apiKey:         "4ebd366d-76f4-4400-a3b6-e51515d054d6",
			apiSecret:      "fd8a1652-728b-42fe-82b8-f623e56da8850750f5bf-ce66-4ca7-8b84-93651abc723b",
			nonce:          "9675d0f8-1325-484b-9594-c9d6d3268890",
			time:           "1543597115712",
			want:           "21e6a16f6eb34ac476d59f969f548b47fffe3fea318d9c99e77fc710d2fed798",
		},
	}

	for _, tc := range tcs {
		client, err := NewClient(
			tc.apiBaseURL,
			tc.organizationID,
			tc.apiKey,
			tc.apiSecret,
		)
		if err != nil {
			t.Fatalf("failed to initialize client. err: %s", err)
		}

		req, err := client.newRequest(
			context.Background(),
			tc.method,
			tc.spath,
			tc.query,
			tc.body,
		)
		if err != nil {
			t.Fatalf("failed to generate request. err: %s", err)
		}

		req.Header.Set("X-Time", tc.time)
		req.Header.Set("X-Nonce", tc.nonce)

		got, err := client.generateHMACSignature(req)
		if err != nil {
			t.Fatalf("failed to exec generateHMACSignature. err: %s", err)
		}

		if got != tc.want {
			t.Errorf("unmatched HMAC. got: %s, want: %s", got, tc.want)
		}
	}
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
