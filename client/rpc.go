package client

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"sync"
	"time"

	"github.com/whatust/transmission-rss/config"
	"github.com/whatust/transmission-rss/helper"
	"github.com/whatust/transmission-rss/logger"
)

// RSSClient methods to interact with the tranmission RPC server
type RSSClient interface {
	Initialize(*config.Config) error
	AddFeeds([]config.Feed, helper.SeenTorrent)
}

// RPCClient ...
type RPCClient struct {
	URL       string
	SessionID string
	Creds     config.Creds
	Client    Client
}

// TransmissionClient wraps client methods and sessions
type TransmissionClient struct {
	RPCClient      RPCClient
	ConnectionConf config.Connect
	TorrentPath    string
}

// Initialize rpc client
func (c *TransmissionClient) Initialize(conf *config.Config) error {

	if conf.SaveTorrent {
		os.MkdirAll(conf.TorrentPath, 0755)
	}

	var scheme string = "http"

	if conf.Server.TLS {
		scheme = scheme + "s"
	}

	URL := url.URL{
		Scheme: scheme,
		Host:   conf.Server.Host + ":" + strconv.Itoa(conf.Server.Port),
		Path:   conf.Server.RPCPath,
	}
	c.RPCClient.URL = URL.String()

	logger.Info("Initializing Server: %v\n", c.RPCClient.URL)

	c.RPCClient.Client = NewRateClient(
		conf.Server.Proxy,
		conf.Server.ValidateCert,
		conf.Connect.Timeout,
		conf.Server.RateTime,
	)

	c.RPCClient.Creds = conf.Creds
	c.ConnectionConf = conf.Connect

	sessionID, err := c.getSessionID()
	c.RPCClient.SessionID = sessionID

	return err
}

// RetriveFeed ...
func (c TransmissionClient) RetriveFeed(client Client, url string) (*Feed, error) {

	logger.Info("Retriving RSS feed..")

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	for i := 0; i < c.ConnectionConf.Retries; i++ {

		resp, err := client.Do(req)

		if err != nil {
			logger.Error("Response error: %v", err)
			logger.Error("Waiting %v seconds until retry\n", c.ConnectionConf.WaitTime)
			time.Sleep(time.Duration(c.ConnectionConf.WaitTime) * time.Second)
		} else {
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusOK {

				data, err := ioutil.ReadAll(resp.Body)

				if err != nil {
					logger.Error("Unable to read response body.")
					logger.Error("Waiting %v seconds until retry\n", c.ConnectionConf.WaitTime)
					time.Sleep(time.Duration(c.ConnectionConf.WaitTime) * time.Second)
					continue
				}

				feed, err := ParseXML(data)

				if err != nil {
					logger.Error("Unable to parse XML.")
					logger.Error("Waiting %v seconds until retry\n", c.ConnectionConf.WaitTime)
					time.Sleep(time.Duration(c.ConnectionConf.WaitTime) * time.Second)
					continue
				}

				return feed, nil
			}

			logger.Error("Response Status Code(%v)\n", resp.StatusCode)
			logger.Error("Waiting %v seconds until retry\n", c.ConnectionConf.WaitTime)
		}
	}

	return nil, fmt.Errorf("All retries failed could not retrieve RSS Feed")
}
func (c TransmissionClient) getSessionID() (string, error) {

	var sessionID string

	logger.Info("Getting session ID from: %v", c.RPCClient.URL)

	req, err := http.NewRequest("GET", c.RPCClient.URL, nil)
	if err != nil {
		return "", err
	}
	req.SetBasicAuth(c.RPCClient.Creds.Username, c.RPCClient.Creds.Password)

	for i := 0; i < c.ConnectionConf.Retries; i++ {
		logger.Info("Getting session ID")

		resp, err := c.RPCClient.Client.Do(req)

		if err != nil || resp.StatusCode != 409 {
			logger.Error("Unable to get sessionID: %v", err)
			logger.Error("Waiting %v seconds until retry\n", c.ConnectionConf.WaitTime)
			time.Sleep(time.Duration(c.ConnectionConf.WaitTime) * time.Second)
		} else {
			sessionID = resp.Header.Get("X-Transmission-Session-Id")
			break
		}
	}

	if len(sessionID) == 0 {
		return "", fmt.Errorf("All retries failed could not retrieve sessionID")
	}
	logger.Info("SessionID: %v", sessionID)

	return sessionID, nil
}

// AddFeeds ...
func (c TransmissionClient) AddFeeds(confs []config.Feed, seen helper.SeenTorrent) {

	for _, conf := range confs {

		logger.Info("Processing feed: %v\n", conf.URL)

		client := NewRateClient(
			conf.Proxy,
			conf.ValidateCert,
			c.ConnectionConf.Timeout,
			c.ConnectionConf.RateTime,
		)

		feed, err := c.RetriveFeed(client, conf.URL)
		if err != nil {
			logger.Error("Could not retrieve RSS feed: %v\n", err)
			continue
		}

		c.addFeed(conf, feed, client, seen)
	}
}

var wg = sync.WaitGroup{}

// AddFeed adds all torrents that match filter to transmission
func (c TransmissionClient) addFeed(conf config.Feed, feed *Feed, client *RateClient, seen helper.SeenTorrent) {

	for _, matcher := range conf.Matchers {

		logger.Info("Processing match: %v\n", matcher.RegExp)

		filter, err := CreateFilter(matcher, conf)
		if err != nil {
			logger.Error("Error while creating torrent filter: %v\n", err)
			continue
		}

		for _, item := range feed.Channel.Items {

			wg.Add(1)
			go c.processItem(item, filter, client, seen)
		}
	}
	wg.Wait()
}

func (c TransmissionClient) processItem(item FeedItem, filter *Filter, client *RateClient, seen helper.SeenTorrent) {

	defer wg.Done()

	if !FilterTorrent(item, filter) {
		logger.Info("Torrent does not match filter: %v\n", item.Title)
		return
	}

	if seen.Contain(item.Title) {
		logger.Info("Torrent already seen: %v\n", item.Title)
		return
	}

	if len(c.TorrentPath) != 0 {

		torrentPath := path.Join(c.TorrentPath, item.Title+".torrent")

		if _, err := os.Stat(torrentPath); os.IsNotExist(err) {

			err := getTorrent(item.Link, client.Client, torrentPath)
			if err != nil {
				logger.Error("Unable to download torrent: %v\n", err)
				return
			}
		}

		_, err := addTorrent(torrentPath, &c.RPCClient, filter.DownloadPath)
		if err != nil {
			logger.Error("Unable to add torrent: %v\n", err)
			return
		}
		seen.AddSeen(item.Title)

	} else {
		_, err := addTorrentURL(item.Link, &c.RPCClient, filter.DownloadPath)
		if err != nil {
			logger.Error("Unable to add torrent: %v\n", err)
			return
		}
		seen.AddSeen(item.Title)
	}
}
