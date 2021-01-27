package client

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
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

func TestRetrieveFeed(t *testing.T) {

	var tests = []struct{
		statusCode int
		filename string
		retries int
		waitTime int
		connectError bool
		expected *Feed
		expectedErr error
	}{ { 
			200,
			"../test/feed/feed1.xml",
			10, 3, false,
			&Feed{
				Channel: Channel{
					Items: []FeedItem {
						{
							Title: "title1",
							Link: "http://example1.com",
							Remake: "Yes",
							Trusted: "Yes",
						},
						{
							Title: "title2",
							Link: "http://example2.com",
							Remake: "Yes",
							Trusted: "No",
						},
						{
							Title: "title3",
							Link: "http://example3.com",
							Remake: "No",
							Trusted: "No",
						},
						{
							Title: "title4",
							Link: "http://example4.com",
							Remake: "No",
							Trusted: "Yes",
						},
					},
				},
			},
			nil,
		},
	}

	for idx, test := range tests{

		tclient := TransmissionClient {
			ConnectionConf: config.Connect{
				Retries: test.retries,
				WaitTime: test.waitTime,
			},
		}

		data, err:= ioutil.ReadFile(test.filename)
		if err != nil {
			fmt.Printf("%v\n", err)
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r* http.Request){
			if test.statusCode != 200 {
				w.WriteHeader(test.statusCode)
			}else{
				binary.Write(w, binary.LittleEndian, data)
			}
		}))

		feed, err := tclient.RetriveFeed(server.Client(), server.URL)

		if err == nil {
			if !reflect.DeepEqual(feed, test.expected) {
				t.Errorf("Test %v Failed:\nGot:         %v\nExpected: %v\n", idx, feed, test.expected)
			}
		} else {
			if test.expectedErr == nil {
				t.Errorf("Test %v Failed:\nGot:         %v\nExpected: %v\n", idx, err, test.expectedErr)
			}
		}
	}
}