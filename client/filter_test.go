package client

import (
	"regexp"
	"testing"
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
