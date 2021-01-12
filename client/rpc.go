package client

import (
	"io"
	"os"
	"path"
	"io/ioutil"
	"crypto/tls"
	"golang.org/x/time/rate"
	"sync"
	"bytes"
	"fmt"
	"time"
	"strconv"
	"net/url"
	"net/http"
	"encoding/json"
	"encoding/base64"
	"github.com/whatust/transmission-rss/config"
	"github.com/whatust/transmission-rss/helper"
	"github.com/whatust/transmission-rss/logger"
)

// RSSClient methods to interact with the tranmission RPC server
type RSSClient interface {
	Initialize() error
	AddFeeds([]config.Feed, helper.SeenTorrent)
}

// TransmissionClient wraps client methods and sessions
type TransmissionClient struct {
	Creds config.Creds
	Server config.Server
	Connect config.Connect
	URL string
	sessionID string
	clientRSS Client
}

// Initialize rpc client
func(c *TransmissionClient) Initialize() error {

	if c.Server.SaveTorrent {
		os.MkdirAll(c.Server.TorrentPath, 0755)
	}

	var scheme string = "http"

	if c.Server.TLS {
		scheme = scheme + "s"
	}

	URL := url.URL {
		Scheme: scheme,
		Host: c.Server.Host + ":" + strconv.Itoa(c.Server.Port),
		Path: c.Server.RPCPath,
	}
	c.URL = URL.String()

	logger.Info("Initializing Server: %v\n", c.URL)

	var proxy func(*http.Request) (*url.URL, error);

	if len(c.Server.Proxy) != 0 {

		proxyURL, err := url.Parse(c.Server.Proxy)
		if err != nil {
			logger.Warn("Could not parse proxy address: %v", err)
		} else {
			proxy = http.ProxyURL(proxyURL)
		}
	}

	c.clientRSS = &RateClient {
		Client: &http.Client {
			Transport: &http.Transport {
				TLSClientConfig : &tls.Config {
					InsecureSkipVerify: !c.Server.ValidateCert,
				},
				Proxy: proxy,
				TLSHandshakeTimeout: 10 * time.Second,
			},
			Timeout: time.Duration(c.Connect.Timeout) * time.Second,
		},
		RateLimiter: rate.NewLimiter(rate.Every(time.Duration(c.Server.RateTime) * time.Millisecond), 1),
	}

	sessionID, err := c.getSessionID()
	c.sessionID = sessionID

	return err
}

// RetriveFeed ...
func (c TransmissionClient) RetriveFeed(url string) (*Feed, error) {

	logger.Info("Retriving RSS feed..")

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	for i := 0; i < c.Connect.Retries; i++ {

		resp, err := c.clientRSS.Do(req)

		if err != nil {
			logger.Error("Response error: %v", err)
			logger.Error("Waiting %v seconds until retry\n", c.Connect.WaitTime)
			time.Sleep(time.Duration(c.Connect.WaitTime) * time.Second)
		} else {
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusOK {

				data, err := ioutil.ReadAll(resp.Body)

				if err != nil {
					logger.Error("Unable to read response body.")
					logger.Error("Waiting %v seconds until retry\n", c.Connect.WaitTime)
					time.Sleep(time.Duration(c.Connect.WaitTime) * time.Second)
					continue
				}

				feed, err := ParseXML(data)

				if err != nil {
					logger.Error("Unable to parse XML.")
					logger.Error("Waiting %v seconds until retry\n", c.Connect.WaitTime)
					time.Sleep(time.Duration(c.Connect.WaitTime) * time.Second)
					continue
				}

				return feed, nil
			}

			logger.Error("Response Status Code(%v)\n", resp.StatusCode)
			logger.Error("Waiting %v seconds until retry\n", c.Connect.WaitTime)
		}
	}

	return nil, fmt.Errorf("All retries failed could not retrieve RSS Feed")
}
func (c TransmissionClient) getSessionID() (string, error) {

	var sessionID string

	logger.Info("Getting session ID from: %v", c.URL)

	req, err := http.NewRequest("GET", c.URL, nil)
	if err != nil {
		return "", err
	}
	req.SetBasicAuth(c.Creds.Username, c.Creds.Password)

	fmt.Printf("Retries: %v", c.Connect.Retries)
	
	for i := 0; i < c.Connect.Retries; i++ {
		logger.Info("Getting session ID")

		resp, err := c.clientRSS.Do(req)

		if err != nil || resp.StatusCode != 409 {
			logger.Error("Unable to get sessionID: %v", err)
			logger.Error("Waiting %v seconds until retry\n", c.Connect.WaitTime)
			time.Sleep(time.Duration(c.Connect.WaitTime) * time.Second)
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
func (c TransmissionClient) AddFeeds(confs []config.Feed, seen helper.SeenTorrent) {

	for _, conf := range confs {

		logger.Info("Processing feed: %v\n", conf.URL)

		feed, err := c.RetriveFeed(conf.URL)
		if err != nil {
			logger.Error("Could not retrieve RSS feed: %v\n", err)
			continue
		}

		c.addFeed(conf, feed, seen)
	}
}

var wg = sync.WaitGroup{}

// AddFeed adds all torrents that match filter to transmission
func (c TransmissionClient) addFeed(conf config.Feed, feed *Feed, seen helper.SeenTorrent) {

	var proxy func(*http.Request) (*url.URL, error);

	if len(conf.Proxy) != 0 {

		proxyURL, err := url.Parse(conf.Proxy)
		if err != nil {
			logger. Warn("Could not parse proxy address: %v", err)
		} else  {
			proxy = http.ProxyURL(proxyURL)
		}
	}

	logger.Info("RateTime: %v\n", c.Connect.RateTime)

	clt := RateClient {
		Client: &http.Client{
			Transport: &http.Transport {
				TLSClientConfig: &tls.Config {
					InsecureSkipVerify: !conf.ValidateCert,
				},
			Proxy: proxy,
			TLSHandshakeTimeout: 10 * time.Second,
			},
		},
		RateLimiter: rate.NewLimiter(rate.Every(time.Duration(c.Connect.RateTime) * time.Millisecond), 1),
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
			go c.processItem(item, filter, clt, seen)
		}
	}
	wg.Wait()
}

func (c TransmissionClient) processItem(item FeedItem, filter *Filter, clt Client, seen helper.SeenTorrent) {

	defer wg.Done()

	if !FilterTorrent(item, filter) {
		logger.Info("Torrent does not match filter: %v\n", item.Title)
		return
	}

	if seen.Contain(item.Title) {
		logger.Info("Torrent already seen: %v\n", item.Title)
		return
	}

	if c.Server.SaveTorrent {

		torrentPath := path.Join(c.Server.TorrentPath, item.Title + ".torrent")

		if _, err := os.Stat(torrentPath); os.IsNotExist(err) {

			err := c.getTorrent(item.Link, clt, torrentPath)
			if err != nil {
				logger.Error("Unable to download torrent: %v\n", err)
				return
			}
		}

		_, err := c.addTorrent(torrentPath, filter.DownloadPath)
		if err != nil {
			logger.Error("Unable to add torrent: %v\n", err)
			return
		}
		seen.AddSeen(item.Title)

	} else {
		_, err := c.addTorrentURL(item.Link, filter.DownloadPath)
		if err != nil {
			logger.Error("Unable to add torrent: %v\n", err)
			return
		}
		seen.AddSeen(item.Title)
	}

}

type argumentsURL struct {
	Paused bool				`json:"paused"`
	DownloadDir string		`json:"download-dir"`
	Filename string			`json:"filename"`
}

type argumentsTorrent struct {
	Paused bool				`json:"paused"`
	DownloadDir string		`json:"download-dir"`
	MetaInfo string			`json:"metainfo"`
}

type addURL struct {
	Method string				`json:"method"`
	Arguments argumentsURL		`json:"arguments"`
}

type addTorrent struct {
	Method string				`json:"method"`
	Arguments argumentsTorrent	`json:"arguments"`
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

	data := addURL{
		Method: "torrent-add",
		Arguments: argumentsURL {
			Paused: true,
			DownloadDir: path,
			Filename: url,
		},
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		logger.Error("Unable to marshal POST JSON: %v\n", err)
		return "", err
	}

	req, err := http.NewRequest("POST", c.URL, bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Error("Unable to create POST request: %v\n", err)
		return "", err
	}

	req.SetBasicAuth(c.Creds.Username, c.Creds.Password)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Transmission-Session-Id", c.sessionID)

	resp, err := c.clientRSS.Do(req)

	if err != nil {
		logger.Error("Error during POST request: %v\n", err)
		return "", err
	}
	defer resp.Body.Close()

	var body respTorrent
	err = json.NewDecoder(resp.Body).Decode(&body)

	if err != nil {
		logger.Error("Unable to parse response body JSON:%v\n%v\n", err, resp.Body)
		return "", err
	}

	if body.Result != "success" {
		logger.Error("Unable to add torrent: %v\n", body)
		return "", err
	}

	hashString := body.Arguments.TorrentAdded.HashString

	if len(hashString) == 0 {

		hashString = body.Arguments.TorrentDuplicate.HashString
		/*if len(hashString) != 0 {
			logger.Info("Torrent from %v and hash %v duplicated\n", url, hashString)
		}*/
	}

	return hashString, nil
}

func(c TransmissionClient) getTorrent(link string, clt Client, torrentFilename string) error {

	req, err := http.NewRequest("GET", link, nil)
	if err!= nil {
		logger.Error("Unable to create GET request: %v\n", err)
		return err
	}

	resp, err := clt.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	logger.Info("Response: %v\n", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Status code different from 200 (%v)", resp.StatusCode)
	}

	out, err := os.Create(torrentFilename)
	if err != nil {
		return err
	}
	defer out.Close()

	err = os.Chmod(torrentFilename, 0644)
	if err != nil {
		return err
	}

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	logger.Debug("Retriving torrent:\n %v\n", resp)
	return nil
}

func(c TransmissionClient) addTorrent(torrentPath string, path string) (string, error) {

	torrentData, err := ioutil.ReadFile(torrentPath)
	if err != nil {
		return "", err
	}

	torrent := base64.StdEncoding.EncodeToString([]byte(torrentData))

	data := addTorrent{
		Method: "torrent-add",
		Arguments: argumentsTorrent {
			Paused: true,
			DownloadDir: path,
			MetaInfo: torrent,
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

	resp, err := c.clientRSS.Do(req)
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
		return "", fmt.Errorf("%v: %v", torrentPath, body.Result)
	}

	hashString := body.Arguments.TorrentAdded.HashString
	if len(hashString) == 0 {

		hashString = body.Arguments.TorrentDuplicate.HashString
		if len(hashString) != 0 {
			logger.Debug("Torrent from %v and hash %v duplicated", path, hashString)
		}
	}

	return hashString, nil
}
