package nicehash

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/xerrors"
)

type Client struct {
	URL            string
	HTTPClient     *http.Client
	OrganizationID string
	APIKey         string
	APISecret      string
}

// APIBaseURL is a base URL og Nicehash API
const APIBaseURL = `https://api2.nicehash.com/main/api/v2`

func NewClient(baseUrl, organizationID, apiKey, apiSecret string) (*Client, error) {
	tr := &http.Transport{}
	hcl := &http.Client{Transport: tr}

	client := &Client{
		URL:            APIBaseURL,
		HTTPClient:     hcl,
		OrganizationID: organizationID,
		APIKey:         apiKey,
		APISecret:      apiSecret,
	}

	return client, nil
}

func (c *Client) newRequest(ctx context.Context, method, spath, query string, body []byte) (*http.Request, error) {
	u, err := url.Parse(c.URL + spath + query)

	closer := ioutil.NopCloser(bytes.NewReader(body))

	req, err := http.NewRequest(method, u.String(), closer)
	if err != nil {
		return nil, xerrors.Errorf("failed to create new request: %w", err)
	}

	req = req.WithContext(ctx)

	// set headers
	now := time.Now().UnixNano() / int64(time.Millisecond)
	req.Header.Set("X-Time", fmt.Sprint(now))

	nonce := uuid.New().String()
	req.Header.Set("X-Nonce", nonce)

	requestID := uuid.New().String()
	req.Header.Set("X-Request-Id", requestID)

	req.Header.Set("X-Organization-Id", c.OrganizationID)

	c.signRequest(req)
	return req, nil
}

func (c *Client) signRequest(req *http.Request) error {
	hmac, err := c.generateHMACSignature(req)
	if err != nil {
		return xerrors.Errorf("failed to sign request: %w", err)
	}

	req.Header.Set("X-Auth", strings.Join([]string{c.APIKey, hmac}, ":"))

	return nil
}

func (c *Client) generateHMACSignature(req *http.Request) (string, error) {
	h := req.Header

	input := fmt.Sprintf(
		"%s\x00%s\x00%s\x00\x00%s\x00\x00%s\x00%s\x00%s",
		c.APIKey,
		h.Get("X-Time"),
		h.Get("X-Nonce"),
		h.Get("X-Organization-Id"),
		req.Method,
		req.URL.Path,
		req.URL.RawQuery,
	)

	if req.Method != http.MethodGet {
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return "", err
		}

		input += "\x00"
		input += string(body)

		req.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	}

	mac := hmac.New(sha256.New, []byte(c.APISecret))
	mac.Write([]byte(input))
	return hex.EncodeToString(mac.Sum(nil)), nil
}

func decodeBody(resp *http.Response, out interface{}) error {
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)

	return decoder.Decode(out)
}

const getRigs2spath = "/mining/rigs2"

func (c *Client) GetRigs2(ctx context.Context) (*GetRigs2Response, error) {
	req, err := c.newRequest(ctx, http.MethodGet, getRigs2spath, "", []byte{})
	if err != nil {
		return nil, err
	}

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, xerrors.Errorf("failed to read response body: %w", err)
		}

		return nil, fmt.Errorf("api returns non-OK state. status: %s, body: %s", res.Status, body)
	}

	rig2Resp := &GetRigs2Response{}
	err = decodeBody(res, rig2Resp)
	if err != nil {
		return nil, err
	}

	return rig2Resp, nil
}
