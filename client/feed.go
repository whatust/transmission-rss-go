package client

import (
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/whatust/transmission-rss/logger"
)

// FeedItem structure that wraps the torrent data
type FeedItem struct {
	Title   string `xml:"title"`
	Link    string `xml:"link"`
	Remake  string `xml:"remake"`
	Trusted string `xml:"trusted"`
}

// Channel ...
type Channel struct {
	Items []FeedItem `xml:"item"`
}

// Feed structure that wraps the list of torrents
type Feed struct {
	Channel Channel `xml:"channel"`
}

// ParseXML ...
func ParseXML(body []byte) *Feed {

	var feed Feed

	err := xml.Unmarshal([]byte(body), &feed)
	if err != nil {
		logger.Error("%v", err)
		return nil
	}

	if logger.IsLevelGreaterEqual(logger.DebugLevel) {
		xmlIndent, err := xml.MarshalIndent(feed, "  ", "    ")
		if err != nil {
			logger.Error("Error Marshaling Feed: %v\n", err)
		}
		logger.Debug("RSS Feed:\n%v\n", string(xmlIndent))
	} else {
		logger.Debug("RSS Feed:\n%v\n", feed)
	}

	return &feed
}

// ParseResponseXML ...
func ParseResponseXML(resp *http.Response, waitTime int) *Feed {

	if resp.StatusCode == http.StatusOK {

		data, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			logger.Error("Unable to read response body: %v", err)
			logger.Error("Waiting %v seconds until retry\n", waitTime)
			time.Sleep(time.Duration(waitTime) * time.Second)
			return nil
		}

		feed := ParseXML(data)

		if feed == nil {
			logger.Error("Waiting %v seconds until retry\n", waitTime)
			time.Sleep(time.Duration(waitTime) * time.Second)
			return nil
		}

		return feed
	}
	logger.Error("Response Status Code(%v)\n", resp.StatusCode)
	logger.Error("Waiting %v seconds until retry\n", waitTime)

	return nil
}