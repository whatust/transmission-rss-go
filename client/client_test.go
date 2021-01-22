package client

import (
	"net/http"
	"testing"
	"time"
)

func TestNewRateClient(t *testing.T) {

	var tests = []struct {
		proxy    string
		validate bool
		timeout  int
		rateTime int
	}{
		{"", false, 0, 0},
		{"http://localhost:8080", true, 10, 600},
	}

	for idx, test := range tests {

		client := NewRateClient(test.proxy, test.validate, test.timeout, test.rateTime)

		if client.Client.Timeout != time.Duration(test.timeout)*time.Second {
			t.Errorf("Failed %v Test: expected %v, received %v", idx, test.rateTime, client.RateLimiter.Limit())
		}

		if client.Client.Transport.(*http.Transport).TLSClientConfig.InsecureSkipVerify == test.validate {
			t.Errorf("Failed %v Test: expected %v, received %v", idx, test.validate, client.Client.Transport.(*http.Transport).TLSClientConfig.InsecureSkipVerify)
		}
	}

}
