package client

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/whatust/transmission-rss/config"
)

func TestSessionID(t *testing.T) {

	var tests = []struct {
		statusCode int
		retData string
		expected string
		expectedError error
	}{
		{ 409, "", "sessionID", nil, },
	}

	for idx, test := range tests{

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
			w.WriteHeader(test.statusCode)
			w.Header().Add("X-Transmission-Session-Id", "sessionID")
		}))

		client := TransmissionClient{
			RPCClient: RPCClient{
				URL: server.URL,
				Client: NewRateClient("", false, 0, 0 ),
				Creds: config.Creds{
					Username: "transmission",
					Password: "transmission",
				},
			},
			ConnectionConf: config.Connect{
				Retries: 10,
				WaitTime: 0,
			},
		}

		sessionID, err := client.getSessionID()

		if err == nil {
			if sessionID != test.expected {
				t.Errorf("Test %v Failed:\nGot:     %v\nExpected: %v", idx, sessionID, test.expected)
			}
		} else {
			if test.expectedError != nil {
				t.Errorf("Test %v Failed:\nGot:     %v\nExpected: %v", idx, err, test.expectedError)
			}
		}

	}
	
}

func TestInitialize(t *testing.T) {

	var tests = []struct{
		conf *config.Config
		expected TransmissionClient
		expectedError error
	}{
		{
			&config.Config{
				Proxy: "",
				TorrentPath: "",
				Server: config.Server{
					RPCPath: "/transmission/rpc",
					TLS: false,
					ValidateCert: true,
					RateTime: 600,
				},
				Creds: config.Creds{
					Username: "transmission",
					Password: "transmission",
				},
				Connect: config.Connect{
					Timeout: 10,
					Retries: 10,
					WaitTime: 0,
				},
			},
			TransmissionClient{
				Proxy: "",
				TorrentPath: "",
				RPCClient: RPCClient{
					URL: "",
					Creds: config.Creds {
						Username: "transmission",
						Password: "transmission",
					},
					SessionID: "sessionId",
				},
				ConnectionConf: config.Connect{
					Timeout: 10,
					Retries: 10,
					WaitTime: 0,
				},
			},
			nil,
		},
	}

	for idx, test := range tests {

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
			w.Header().Add("X-Transmission-Session-Id", "sessionID")
			w.WriteHeader(409)
		}))

		myClient := TransmissionClient{}
		var client RSSClient = &myClient

		if len(test.conf.Server.Host) == 0 {

			url := strings.Split(server.URL, "://")
			s := strings.Split(url[1], ":")

			test.conf.Server.Host = s[0]
			port, _ := strconv.Atoi(s[1])
			test.conf.Server.Port = port

			test.expected.RPCClient.URL = server.URL+test.conf.Server.RPCPath
		}

		err := client.Initialize(test.conf)

		if err == nil {
			if test.expected.RPCClient.URL != client.(*TransmissionClient).RPCClient.URL ||
				test.expected.Proxy != client.(*TransmissionClient).Proxy ||
				test.expected.RPCClient.Creds != client.(*TransmissionClient).RPCClient.Creds ||
				test.expected.ConnectionConf != client.(*TransmissionClient).ConnectionConf {
				t.Errorf("Test %v Failed:\nGot:      %v\nExpected: %v", idx, client, test.expected)
			}
		}else{
			if test.expectedError == nil {
				t.Errorf("Test %v Failed:\nGot:      %v\nExpected: %v", idx, err, test.expectedError)
			}
		}
	}
}