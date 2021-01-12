package client

import (
	"net/http"
	"context"
	"golang.org/x/time/rate"
)

// RateClient ...
type RateClient struct {
	Client *http.Client
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