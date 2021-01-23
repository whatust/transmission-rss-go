package client

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/whatust/transmission-rss/config"
)

func TestNewRateClient(t *testing.T) {

	var tests = []struct {
		proxy    string
		validate bool
		timeout  int
		rateTime int
	}{
		{"", false, 0, 0},
		{"http://localhost:8080\n", false, 0, 0},
		{"http://localhost:8080\r", false, 0, 0},
		{"http://localhost:8080\t", false, 0, 0},
		{"http://localhost: 8080", false, 0, 0},
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

func TestAddTorrent(t *testing.T) {

	var tests = []struct{
		sentData []byte 
		retData string
		statusCode int
		expected error
	}{
		{ []byte(""), "", 200, fmt.Errorf("Unable to parse response body JSON") },
		{ []byte(""), "", 429, fmt.Errorf("Status code different from 200 (%v)", 429) },
		{ []byte(""), "{\"result\":\"success\"}", 200, nil },
		{ []byte(""), "{\"result\":\"error\"}", 200, fmt.Errorf("Unable to add torrent") },
		{ []byte(""), "{\"arguments\":{\"torrent-added\":{\"hashString\":\"hashstring\",\"id\":1,\"name\":\"title\"}},\"result\":\"success\"}", 200, nil },
		{ []byte(""), "{\"arguments\":{\"torrent-duplicate\":{\"hashString\":\"hashstring\",\"id\":1,\"name\":\"title\"}},\"result\":\"success\"}", 200, nil },
	}

	for idx, test := range tests {

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
			if test.statusCode != 200 {
				w.WriteHeader(test.statusCode)
			}else{
				fmt.Fprintf(w, test.retData)
			}
		}))

		clientRPC := RPCClient{
			URL: server.URL,
			SessionID: "sessionID",
			Creds: config.Creds{
				Username: "transmission",
				Password: "transmission",
			},
			Client: server.Client(),
		}
		err := addTorrent(test.sentData, &clientRPC)

		if (err == nil) != (test.expected == nil) {
			t.Errorf("Test %v Failed: expected %v, received %v", idx, test.expected, err)
		}
	}
}
