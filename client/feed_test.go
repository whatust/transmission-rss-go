package client

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestParseXML(t *testing.T) {

	var tests = []struct {
		filename string
		expected *Feed
	}{
		{"../test/feed/feed1.xml",
			&Feed{
				Channel: Channel{
					Items: []FeedItem{
						{
							Title:   "title1",
							Link:    "http://example1.com",
							Remake:  "Yes",
							Trusted: "Yes",
						},
						{
							Title:   "title2",
							Link:    "http://example2.com",
							Remake:  "Yes",
							Trusted: "No",
						},
						{
							Title:   "title3",
							Link:    "http://example3.com",
							Remake:  "No",
							Trusted: "No",
						},
						{
							Title:   "title4",
							Link:    "http://example4.com",
							Remake:  "No",
							Trusted: "Yes",
						},
					},
				},
			},
		},
		{"../test/feed/feed2.xml", nil },
	}

	for idx, test := range tests {

		data, err := ioutil.ReadFile(test.filename)
		if err != nil {
			fmt.Printf("%v\n", err)
		}

		feed := ParseXML([]byte(data))

		if !reflect.DeepEqual(feed, test.expected) {
			t.Errorf("Test %v Failed: Feed do not match \nGot      %v, \nExpected %v", idx, feed, test.expected)
		}
	}
}

func TestParseResponseXML(t *testing.T) {

	var tests = []struct {
		statuscode int
		filename   string
		expected   *Feed
	}{
		{200, "../test/feed/feed1.xml",
			&Feed{
				Channel: Channel{
					Items: []FeedItem{
						{
							Title:   "title1",
							Link:    "http://example1.com",
							Remake:  "Yes",
							Trusted: "Yes",
						},
						{
							Title:   "title2",
							Link:    "http://example2.com",
							Remake:  "Yes",
							Trusted: "No",
						},
						{
							Title:   "title3",
							Link:    "http://example3.com",
							Remake:  "No",
							Trusted: "No",
						},
						{
							Title:   "title4",
							Link:    "http://example4.com",
							Remake:  "No",
							Trusted: "Yes",
						},
					},
				},
			},
		},
		{429, "../test/feed/unknown1.xml", nil},
		{200, "../test/feed/unknown1.xml", nil},
		{429, "../test/feed/feed1.xml", nil},
	}

	for idx, test := range tests {

		data, err := ioutil.ReadFile(test.filename)
		if err != nil {
			fmt.Printf("%v\n", err)
		}

		handler := func(w http.ResponseWriter, r *http.Request) {
			io.WriteString(w, string(data))
		}

		req := httptest.NewRequest("GET", "http://example.com", nil)
		w := httptest.NewRecorder()
		handler(w, req)

		resp := w.Result()
		resp.StatusCode = test.statuscode

		feed := ParseResponseXML(resp, 0)

		if !reflect.DeepEqual(feed, test.expected) {
			t.Errorf("Test %v Failed: Feed do not match \nGot      %v, \nExpected %v", idx, feed, test.expected)
		}
	}
}
