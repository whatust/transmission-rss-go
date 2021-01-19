package client

import (
	"encoding/xml"
	"fmt"
	"github.com/whatust/transmission-rss/logger"
	"io/ioutil"
	"net/http"
)

// RetriveFeed ...
func RetriveFeed(url string) (*Feed, error) {

	resp, err := http.Get(url)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Network: %v", resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	feed, err := ParseXML(data)

	if err != nil {
		return nil, err
	}

	return feed, nil
}

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

// RetriveTorrent ...
func RetriveTorrent(url string) error {

	resp, err := http.Get(url)

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	logger.Debug("Retriving torrent:\n %v\n", resp)

	return nil
}
