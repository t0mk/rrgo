package rrgo

import (
	"bytes"

	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net/http/httputil"
	"os"
	"strconv"
	"time"

	"github.com/google/go-querystring/query"

	hchttp "github.com/hashicorp/go-retryablehttp"
)

const (
	endpoint  = "https://api.radarrelay.com/0x/v0"
	mediaType = "application/json"
	userAgent = "github.com/t0mk/rrgo"

	debugEnvVar    = "RRGO_DEBUG"
	endpointEnvVar = "RRGO_URL"

	headerRateLimit     = "X-RateLimit-Limit"
	headerRateRemaining = "X-RateLimit-Remaining"
	headerRateReset     = "X-RateLimit-Reset"
)

type Client struct {
	client  *hchttp.Client
	debug   bool
	baseUrl string
}

func (r *Response) populateRate() {
	if limit := r.Header.Get(headerRateLimit); limit != "" {
		r.Rate.RequestLimit, _ = strconv.Atoi(limit)
	}
	if remaining := r.Header.Get(headerRateRemaining); remaining != "" {
		r.Rate.RequestsRemaining, _ = strconv.Atoi(remaining)
	}
	if reset := r.Header.Get(headerRateReset); reset != "" {
		if v, _ := strconv.ParseInt(reset, 10, 64); v != 0 {
			r.Rate.Reset = Timestamp{time.Unix(v, 0)}
		}
	}
}

func NewClient() *Client {
	envD := os.Getenv(debugEnvVar)
	envU := os.Getenv(endpointEnvVar)
	if envU == "" {
		envU = endpoint
	}

	c := &Client{
		client:  hchttp.NewClient(),
		debug:   (envD != "") && (envD != "0"),
		baseUrl: envU,
	}
	c.client.RetryMax = 5
	c.client.Logger = log.New(ioutil.Discard, "", log.LstdFlags)
	return c
}

func (c *Client) Do(method string, path string, body, out interface{}) (*Response, error) {

	url := c.baseUrl + path

	r := bytes.NewReader([]byte{})
	if body != nil {

		bs, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		r = bytes.NewReader(bs)
	}
	req, err := hchttp.NewRequest(method, url, r)
	if err != nil {
		return nil, err
	}

	req.Close = true

	req.Header.Add("Content-Type", mediaType)
	req.Header.Add("Accept", mediaType)
	req.Header.Add("User-Agent", userAgent)

	if c.debug {
		o, _ := httputil.DumpRequestOut(req.Request, true)
		log.Printf("%s\n", string(o))
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	response := Response{Response: resp}
	response.populateRate()
	if c.debug {
		o, _ := httputil.DumpResponse(response.Response, true)
		log.Printf("%s\n", string(o))
	}

	if out != nil {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return &response, err
		}
		err = json.Unmarshal(body, out)
		if err != nil {
			return &response, err
		}
	}

	return &response, nil
}

func (c *Client) Pairs(po PairsOpts) ([]Pair, *Response, error) {

	pairs := []Pair{}
	v, err := query.Values(po)
	if err != nil {
		return nil, nil, err
	}
	murl := "/token_pairs"
	if len(v) > 0 {
		murl = murl + "?" + v.Encode()
	}

	resp, err := c.Do("GET", murl, nil, &pairs)
	if err != nil {
		return nil, resp, err
	}
	return pairs, resp, nil

}

func (t *Token) String() string {
	return fmt.Sprintf("{Adr: %s, Min: %s, Max: %s, Pre: %d}",
		t.Address,
		new(big.Int).SetBytes(t.MinAmount[:]).String(),
		new(big.Int).SetBytes(t.MaxAmount[:]).String(),
		t.Precision)

}

func (c *Client) Orders(oo OrdersOpts) ([]APIOrder, *Response, error) {
	orders := []APIOrder{}
	v, err := query.Values(oo)
	if err != nil {
		return nil, nil, err
	}
	murl := "/orders"
	if len(v) > 0 {
		murl = murl + "?" + v.Encode()
	}

	resp, err := c.Do("GET", murl, nil, &orders)
	if err != nil {
		return nil, resp, err
	}

	for i := range orders {
		err = orders[i].Process()
		if err != nil {
			return nil, resp, err
		}
	}

	return orders, resp, nil

}

func (c *Client) Orderbook(oo OrderbookOpts) (*Orderbook, *Response, error) {
	ob := Orderbook{}
	if oo.BaseTokenAddress == "" {
		return nil, nil, fmt.Errorf("missing baseTokenAddres in %s", oo)
	}
	if oo.QuoteTokenAddress == "" {
		return nil, nil, fmt.Errorf("missing quoteTokenAddres in %s", oo)
	}
	v, err := query.Values(oo)

	if err != nil {
		return nil, nil, err
	}
	murl := "/orderbook?" + v.Encode()

	resp, err := c.Do("GET", murl, nil, &ob)
	if err != nil {
		return nil, resp, err
	}
	for i := range ob.Asks {
		err = ob.Asks[i].Process()
		if err != nil {
			return nil, resp, err
		}
	}
	for i := range ob.Bids {
		err = ob.Bids[i].Process()
		if err != nil {
			return nil, resp, err
		}
	}

	return &ob, resp, nil
}
