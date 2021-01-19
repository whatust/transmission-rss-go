package client

import (
	//"fmt"
	//"net/http"
	//"io/ioutil"
	"encoding/xml"

	"github.com/whatust/transmission-rss/logger"
)

// FeedItem structure that wraps the torrent data
type FeedItem struct {
	Title   string `xml:"title"`
	Link    string `xml:"link"`
	Remake  string `xml:"remake"`
	Trusted string `xml:"trusted"`
}

// Feed structure that wraps the list of torrents
type Feed struct {
	Channel struct {
		Items []FeedItem `xml:"item"`
	} `xml:"channel"`
}

// ParseXML ...
func ParseXML(resp []byte) (*Feed, error) {

	var feed Feed

	err := xml.Unmarshal([]byte(resp), &feed)
	if err != nil {
		return nil, err
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

	return &feed, nil
}
