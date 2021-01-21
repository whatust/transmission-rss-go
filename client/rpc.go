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
	Proxy          string
}

// Initialize rpc client
func (c *TransmissionClient) Initialize(conf *config.Config) error {

	c.Proxy = conf.Proxy
	c.TorrentPath = conf.TorrentPath
	if len(c.TorrentPath) > 0 {
		os.MkdirAll(c.TorrentPath, 0755)
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

var wg = sync.WaitGroup{}
var wc = sync.WaitGroup{}

// AddFeeds ...
func (c TransmissionClient) AddFeeds(confs []config.Feed, seen helper.SeenTorrent) {

	channel := make(chan TorrentReq, 50)

	wc.Add(1)
	go addTorrentURL(channel, &c.RPCClient, c.ConnectionConf, seen)

	client := NewRateClient(
		c.Proxy,
		true,
		c.ConnectionConf.Timeout,
		c.ConnectionConf.RateTime,
	)

	for _, conf := range confs {

		client.Client.Transport.(*http.Transport).TLSClientConfig.InsecureSkipVerify = !conf.ValidateCert

		logger.Info("Retriving feed from: %v", conf.URL)
		feed, err := c.RetriveFeed(client, conf.URL)
		if err != nil {
			logger.Error("Could not retrieve RSS feed: %v\n", err)
			continue
		}

		for _, matcher := range conf.Matchers {

			logger.Info("Processing match: %v\n", matcher.RegExp)

			filter, err := CreateFilter(matcher, conf)
			if err != nil {
				logger.Error("Error while creating torrent filter: %v\n", err)
				continue
			}

			for _, item := range feed.Channel.Items {

				wg.Add(1)
				go c.processItem(item, filter, channel, seen)
			}
		}
		wg.Wait()
	}

	close(channel)
	wc.Wait()
}

// TorrentReq ...
type TorrentReq struct {
	Link         string
	Title        string
	DownloadPath string
	TorrentPath  string
}

func (c TransmissionClient) processItem(item FeedItem, filter *Filter, channel chan<- TorrentReq, seen helper.SeenTorrent) {

	defer wg.Done()

	if !FilterTorrent(item, filter) {
		logger.Info("Torrent does not match filter: %v\n", item.Title)
		return
	}

	if seen.Contain(item.Title) {
		logger.Info("Torrent already seen: %v\n", item.Title)
		return
	}

	channel <- TorrentReq{
		Link:         item.Link,
		Title:        item.Title,
		DownloadPath: filter.DownloadPath,
		TorrentPath:  path.Join(c.TorrentPath, item.Title+".torrent"),
	}
}
