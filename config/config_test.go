package config

import (
	"fmt"
	"reflect"
	"testing"
)

func TestGetConfig(t *testing.T) {

	var tests = []struct {
		filename    string
		expected    *Config
		expectedErr error
	}{
		{"../test/config/unknown.yml", nil, fmt.Errorf("open ../test/config/unknown.yml: no such file or directory")},
		{"../test/config/wrong.yml", nil, fmt.Errorf("yaml") },
		{"../test/config/minimal.yml",
			&Config{
				Server: Server{
					Host:         "localhost",
					Port:         9091,
					TLS:          false,
					RateTime:     600,
					RPCPath:      "/transmission/rpc",
					Proxy:        "",
					ValidateCert: true,
				},
				Log: Log{
					LogPath:    "/var/log/transmission-rss-log.log",
					Level:      "Info",
					MaxSize:    10000,
					MaxAge:     10,
					MaxBackups: 1,
					Compress:   false,
					LocalTime:  true,
					Formatter:  "JSON",
				},
				Creds: Creds{
					Username: "transmission",
					Password: "transmission",
				},
				Connect: Connect{
					Retries:  10,
					WaitTime: 5,
					Timeout:  10,
					RateTime: 600,
				},
				SeenFile:    "/etc/transmission-rss-seen.log",
				RSSFile:     "/etc/transmission-rss-feeds.yml",
				TorrentPath: "",
			},
			nil,
		},
	}

	for idx, test := range tests {
		config, err := GetConfig(test.filename)

		if config != nil {
			if !reflect.DeepEqual(config, test.expected) {
				t.Errorf("Test %v Failed: Configs do not match for file %v", idx, test.filename)
			}
		} else {
			if config != test.expected {
				t.Errorf("test %v Failed: Config returned nil expected config", idx)
			}
		}

		if (err == nil) != (test.expectedErr == nil) {
			t.Errorf("Test %v Failed: Errors do not match %v, expected %v", idx, err, test.expectedErr)
		}
	}
}

func TestGetFeedsConfig(t *testing.T) {

	var tests = []struct {
		filename    string
		expected    *FeedConfig
		expectedErr error
	}{
		{"../test/feed/unknwon.yml", nil, fmt.Errorf("open ../test/config/unknown.yml: no such file or directory")},
		{"../test/feed/wrong_feed.yml", nil, fmt.Errorf("yaml") },
		{"../test/feed/feed1.yml",
			&FeedConfig{
				Feeds: []Feed{
					{
						URL:            "https:feed1.com",
						SeedRatioLimit: 0,
						Matchers: []Matcher{
							{
								RegExp:       "regexp0",
								DownloadPath: "/var/lib/transmission-daemon/downloads",
								IgnoreRemake: true,
								OnlyTrusted:  true,
							},
							{
								RegExp:       "regexp1",
								DownloadPath: "/var/lib/transmission-daemon/downloads",
								IgnoreRemake: false,
								OnlyTrusted:  true,
							},
							{
								RegExp:       "regexp2",
								DownloadPath: "/var/lib/transmission-daemon/downloads",
								IgnoreRemake: false,
								OnlyTrusted:  false,
							},
						},
						Proxy:        "",
						ValidateCert: true,
					},
					{
						URL:            "http:feed2.org",
						SeedRatioLimit: 1,
						Matchers: []Matcher{
							{
								RegExp:       "regexp3",
								DownloadPath: "/var/lib/transmission-daemon/downloads",
								IgnoreRemake: true,
								OnlyTrusted:  false,
							},
						},
						Proxy:        "http://localhost:8080",
						ValidateCert: false,
					},
				},
			},
			nil,
		},
	}

	for idx, test := range tests {

		feed, err := GetFeedsConfig(test.filename)

		fmt.Printf("Feed: %v\n", feed)
		fmt.Printf("Test: %v\n", test.expected)

		if feed != nil {
			if !reflect.DeepEqual(feed, test.expected) {
				t.Errorf("Test %v Failed: GetFeedsConfig do not match for file %v", idx, test.filename)
			}
		} else {
			if feed != test.expected {
				t.Errorf("Test %v Failed: GetFeedsConfig returned nil expected feed", idx)
			}
		}

		if (err == nil) != (test.expectedErr == nil) {
			t.Errorf("Test %v Failed: Errors do not match %v, expected %v", idx, err, test.expectedErr)
		}
	}
}
