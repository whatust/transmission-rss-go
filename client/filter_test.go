package client

import (
	"fmt"
	"reflect"
	"regexp"
	"testing"

	"github.com/whatust/transmission-rss/config"
)

func TestFilterTorrent(t *testing.T) {

	var tests = []struct {
		item     FeedItem
		filter   *Filter
		expected bool
	}{
		{
			item: FeedItem{
				Title:   "Item1",
				Remake:  "No",
				Trusted: "No",
			},
			filter: &Filter{
				RegExp:       regexp.MustCompile("Item"),
				IgnoreRemake: false,
				OnlyTrusted:  false,
			},
			expected: true,
		},
		{
			item: FeedItem{
				Title:   "item1",
				Remake:  "No",
				Trusted: "No",
			},
			filter: &Filter{
				RegExp:       regexp.MustCompile("Item"),
				IgnoreRemake: false,
				OnlyTrusted:  false,
			},
			expected: false,
		},
		{
			item: FeedItem{
				Title:   "Item1",
				Remake:  "Yes",
				Trusted: "No",
			},
			filter: &Filter{
				RegExp:       regexp.MustCompile("Item"),
				IgnoreRemake: true,
				OnlyTrusted:  false,
			},
			expected: false,
		},
		{
			item: FeedItem{
				Title:   "Item1",
				Remake:  "Yes",
				Trusted: "No",
			},
			filter: &Filter{
				RegExp:       regexp.MustCompile("Item"),
				IgnoreRemake: false,
				OnlyTrusted:  true,
			},
			expected: false,
		},
		{
			item: FeedItem{
				Title:   "Item1",
				Remake:  "No",
				Trusted: "Yes",
			},
			filter: &Filter{
				RegExp:       regexp.MustCompile("Item"),
				IgnoreRemake: true,
				OnlyTrusted:  true,
			},
			expected: true,
		},
	}

	for idx, test := range tests {
		if output := FilterTorrent(test.item, test.filter); output != test.expected {
			t.Errorf("Test %v Failed: expected %v, received %v", idx, test.expected, output)
		}
	}
}


func TestCreateFilter(t *testing.T) {

	var tests = []struct{
		matcher config.Matcher
		expected *Filter
		expectedErr error
	}{
		{
			config.Matcher{
				RegExp: "example",
				DownloadPath: "/var/lib/transmission-daemon/downloads",
				IgnoreRemake: false,
				OnlyTrusted: false,
			},
			&Filter{
				RegExp: regexp.MustCompile("example"),
				DownloadPath: "/var/lib/transmission-daemon/downloads",
				IgnoreRemake: false,
				OnlyTrusted: false,
			},
			nil,
		},
		{
			config.Matcher{
				RegExp: "example",
				DownloadPath: "",
				IgnoreRemake: false,
				OnlyTrusted: false,
			},
			nil,
			fmt.Errorf("Download path  must be set"),
		},
	}

	for idx, test := range tests {

		filter, err := CreateFilter(test.matcher)

		if !reflect.DeepEqual(filter, test.expected) {
			t.Errorf("Test %v Failed: Incorrect filter\nGot     : %v\nExpected: %v", idx, filter, test.expected)
		}

		if (err == nil) != (test.expectedErr == nil) {
			t.Errorf("Test %v Failed: Incorrect error\nGot     : %v\nExpected: %v", idx, filter, test.expected)
		}
	}
}