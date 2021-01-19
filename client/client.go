package client

import (
	"context"
	"crypto/tls"
	"net/http"
	"net/url"
	"time"

	"github.com/whatust/transmission-rss/logger"
	"golang.org/x/time/rate"
)

// RateClient ...
type RateClient struct {
	Client      *http.Client
	RateLimiter *rate.Limiter
}

// Client ...
type Client interface {
	Do(req *http.Request) (*http.Response, error)
}

// Do function from http client with rate limit
func (c RateClient) Do(req *http.Request) (*http.Response, error) {

	ctx := context.Background()
	err := c.RateLimiter.Wait(ctx)

	if err != nil {
		return nil, err
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// NewRateClient ...
func NewRateClient(proxyAddr string, validateCert bool, timeout int, rateTime int) *RateClient {

	var proxy func(*http.Request) (*url.URL, error)

	if len(proxyAddr) != 0 {

		proxyURL, err := url.Parse(proxyAddr)
		if err != nil {
			logger.Warn("Could not parse proxy address: %v", err)
		} else {
			proxy = http.ProxyURL(proxyURL)
		}
	}

	client := &RateClient{
		Client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: !validateCert,
				},
				Proxy: proxy,
			},
			Timeout: time.Duration(timeout) * time.Second,
		},
		RateLimiter: rate.NewLimiter(rate.Every(time.Duration(rateTime)*time.Millisecond), 1),
	}

	return client
}
