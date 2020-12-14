package client

import (
	//"io/ioutil"
	"bytes"
	"fmt"
	"time"
	"strconv"
	"net/url"
	"net/http"
	"encoding/json"
	"github.com/whatust/transmission-rss/config"
	"github.com/whatust/transmission-rss/helper"
	"github.com/whatust/transmission-rss/logger"
)

// Client methods to interact with the tranmission RPC server
type Client interface {
	Initialize() error
	AddFeeds([]config.Feed, helper.SeenSet)
}

// TransmissionClient wraps client methods and sessions
type TransmissionClient struct {
	Creds config.Creds
	Server config.Server
	URL string
	sessionID string
	client *http.Client
}

// Initialize rpc client
func(c *TransmissionClient) Initialize() error {

	var scheme string = "http"

	if c.Server.TLS {
		scheme = scheme + "s"
	}

	URL := url.URL {
		Scheme: scheme,
		Host: c.Server.Host + ":" +strconv.Itoa(c.Server.Port),
		Path: c.Server.RPCPath,
	}
	c.URL = URL.String()

	c.client = &http.Client{}
	logger.Info("Initialize Server: %v\n", c.URL)

	if len(c.Server.Proxy) != 0 {

		proxyURL, err := url.Parse(c.Server.Proxy)
		if err != nil {
			logger.Warn("Could not parse proxy address: %v", err)
		} else {
			c.client.Transport = &http.Transport{Proxy: http.ProxyURL(proxyURL)}
		}
	}

	sessionID, err := c.getSessionID()
	c.sessionID = sessionID

	return err
}

func (c TransmissionClient) getSessionID() (string, error) {

	var sessionID string

	req, err := http.NewRequest("GET", c.URL, nil)
	if err != nil {
		return "", err
	}
	req.SetBasicAuth(c.Creds.Username, c.Creds.Password)
	
	for i := 0; i < c.Server.Retries; i++ {

		resp, err := c.client.Do(req)

		if err != nil || resp.StatusCode != 409 {
			logger.Error("Unable to get sessionID: %v", err)
			logger.Error("Waiting %v seconds util retry\n", c.Server.WaitTime)
			time.Sleep(time.Duration(c.Server.WaitTime) * time.Second)
		} else {
			sessionID = resp.Header.Get("X-Transmission-Session-Id")
			break;
		}
	}

	if len(sessionID) == 0 {
		return "", fmt.Errorf("All retries failed could not retrieve sessionID")
	}
	logger.Info("SessionID: %v", sessionID)

	return sessionID, nil
}

// AddFeeds ...
func (c TransmissionClient) AddFeeds(confs []config.Feed, seenTorrent helper.SeenSet) {

	for _, conf := range confs {

		logger.Info("Processing feed: %v\n", conf)

		feed, err := RetriveFeed(conf.URL)
		if err != nil {
			logger.Error("Could not retrieve RSS feed: %v\n", err)
			continue
		}

		c.getMatcher(feed.Channel.Items, conf, seenTorrent)
	}
}

func (c TransmissionClient) getMatcher(feed []FeedItem, conf config.Feed, seenTorrent helper.SeenSet) {

	for _, matcher := range conf.Matchers {

		logger.Info("Processing feed: %v\n", matcher)

		filter, err := CreateFilter(matcher, conf)
		if err != nil {
			logger.Error("Error while creating torrent filter: %v\n", err)
			continue
		}

		c.addFeed(feed, filter, seenTorrent)
	}
}

// AddFeed adds all torrents that match filter to transmission
func (c TransmissionClient) addFeed(feed []FeedItem, filter *Filter, seenTorrent helper.SeenTorrent) {

	for _, item := range feed {

		if !FilterTorrent(item, filter) {
			logger.Info("Torrent does not match filter: %v\n", item.Title)
			continue
		}

		if seenTorrent.Contain(item.Title) {
			logger.Info("Torrent already seen: %v\n", item.Title)
			continue
		}

		_, err := c.addTorrentURL(item.Link, filter.DownloadPath)
		if err != nil {
			logger.Error("Unable to add torrent: %v\n", err)
			continue
		}

		seenTorrent.AddSeen(item.Title)
		logger.Debug("Added torrent: %v\n", item.Title)
	}
}

type arguments struct {
	Paused bool				`json:"paused"`
	DownloadDir string		`json:"download-dir"`
	Filename string			`json:"filename"`
}

type addTorrent struct {
	Method string				`json:"method"`
	Arguments arguments			`json:"arguments"`
}

type respTorrent struct {
	Arguments struct{
		TorrentAdded struct {
			ID int				`json:"id"`
			HashString string	`json:"hashString"`
			Name string			`json:"name"`
		} 						`json:"torrent-added"`
		TorrentDuplicate struct {
			ID int				`json:"id"`
			HashString string	`json:"hashString"`
			Name string			`json:"name"`
		} 						`json:"torrent-duplicate"`
	}							`json:"arguments"`
	Result string				`json:"result"`
}

// AddTorrentURL adds a torrent to transmission server from an URL
func(c TransmissionClient) addTorrentURL(url string, path string) (string, error) {

	data := addTorrent{
		Method: "torrent-add",
		Arguments: arguments {
			Paused: false,
			DownloadDir: path,
			Filename: url,
		},
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", c.URL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.SetBasicAuth(c.Creds.Username, c.Creds.Password)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Transmission-Session-Id", c.sessionID)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var body respTorrent
	err = json.NewDecoder(resp.Body).Decode(&body)

	if err != nil {
		return "", err
	}

	if body.Result != "success" {
		return "", fmt.Errorf("Unable to add torrent: %v", body.Result)
	}

	hashString := body.Arguments.TorrentAdded.HashString

	if len(hashString) == 0 {

		hashString = body.Arguments.TorrentDuplicate.HashString
		if len(hashString) != 0 {
			logger.Info("Torrent from %v and hash %v duplicated", url, hashString)
		}
	}

	return hashString, nil
}

func(c TransmissionClient) addTorrent(filePath string) error {

	_, err := http.NewRequest("POST", c.URL, nil)
	if err != nil {
		return err
	}


	return nil
}
