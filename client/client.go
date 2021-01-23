package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/whatust/transmission-rss/config"
	"github.com/whatust/transmission-rss/helper"
	"github.com/whatust/transmission-rss/logger"
	"golang.org/x/time/rate"
)

// RateClient ...
type RateClient struct {
	Client      *http.Client
	RateLimiter *rate.Limiter
}

// Client ...
type Client interface {
	Do(req *http.Request) (*http.Response, error)
}

// Do function from http client with rate limit
func (c RateClient) Do(req *http.Request) (*http.Response, error) {

	ctx := context.Background()
	err := c.RateLimiter.Wait(ctx)

	if err != nil {
		return nil, err
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// NewRateClient ...
func NewRateClient(proxyAddr string, validateCert bool, timeout int, rateTime int) *RateClient {

	var proxy func(*http.Request) (*url.URL, error)

	if len(proxyAddr) != 0 {

		proxyURL, err := url.Parse(proxyAddr)
		if err != nil {
			logger.Warn("Could not parse proxy address: %v", err)
		} else {
			proxy = http.ProxyURL(proxyURL)
		}
	}

	client := &RateClient{
		Client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: !validateCert,
				},
				Proxy: proxy,
			},
			Timeout: time.Duration(timeout) * time.Second,
		},
		RateLimiter: rate.NewLimiter(rate.Every(time.Duration(rateTime)*time.Millisecond), 3),
	}

	return client
}

type argumentsURL struct {
	Paused      bool   `json:"paused"`
	DownloadDir string `json:"download-dir"`
	Filename    string `json:"filename"`
}

type argumentsTorrent struct {
	Paused      bool   `json:"paused"`
	DownloadDir string `json:"download-dir"`
	MetaInfo    string `json:"metainfo"`
}

type addURL struct {
	Method    string       `json:"method"`
	Arguments argumentsURL `json:"arguments"`
}

type addFile struct {
	Method    string           `json:"method"`
	Arguments argumentsTorrent `json:"arguments"`
}

type respTorrent struct {
	Arguments struct {
		TorrentAdded struct {
			ID         int    `json:"id"`
			HashString string `json:"hashString"`
			Name       string `json:"name"`
		} `json:"torrent-added"`
		TorrentDuplicate struct {
			ID         int    `json:"id"`
			HashString string `json:"hashString"`
			Name       string `json:"name"`
		} `json:"torrent-duplicate"`
	} `json:"arguments"`
	Result string `json:"result"`
}

func addTorrentURL(items <-chan TorrentReq, client *RPCClient, connection config.Connect, seen helper.SeenTorrent) {

	defer wc.Done()

	for item := range items {

		data := addURL{
			Method: "torrent-add",
			Arguments: argumentsURL{
				Paused:      true,
				DownloadDir: item.DownloadPath,
				Filename:    item.Link,
			},
		}

		jsonData, _ := json.Marshal(data)

		var i int
		for ; i < connection.Retries; i++ {
			err := addTorrent(jsonData, client)
			if err != nil {
				logger.Error("%v", err)
				logger.Error("Waiting %v seconds until retry\n", connection.WaitTime)
				time.Sleep(time.Duration(connection.WaitTime) * time.Second)
			} else {
				seen.AddSeen(item.Title)
				break
			}
		}

		if i == connection.Retries {
			logger.Error("All %v retries failed could not add torrent %v\n", i, item.Link)
		}
	}
}

func addTorrentFile(items <-chan TorrentReq, clientRPC *RPCClient, connection config.Connect, client Client, seen helper.SeenTorrent) {

	for item := range items {

		var torrentData []byte

		for i := 0; i < connection.Retries; i++ {

			if _, err := os.Stat(item.TorrentPath); os.IsNotExist(err) {

				err := getTorrent(item.Link, client, item.TorrentPath)
				if err != nil {
					logger.Error("%v\n", err)
				}
			}

			var err error
			torrentData, err = ioutil.ReadFile(item.TorrentPath)
			if err != nil {
				logger.Error("%v\n", err)
			}

		}

		torrent := base64.StdEncoding.EncodeToString([]byte(torrentData))

		data := addFile{
			Method: "torrent-add",
			Arguments: argumentsTorrent{
				Paused:      true,
				DownloadDir: item.DownloadPath,
				MetaInfo:    torrent,
			},
		}

		jsonData, _ := json.Marshal(data)

		for i := 0; i < connection.Retries; i++ {
			err := addTorrent(jsonData, clientRPC)
			if err != nil {
				logger.Error("%v", err)
			} else {
				seen.AddSeen(item.Title)
				break
			}
		}
		logger.Error("All retries failed could not add torrent %v\n", item.Link)
	}
}

func addTorrent(data []byte, client *RPCClient) error {

	req, err := http.NewRequest("POST", client.URL, bytes.NewBuffer(data))
	if err != nil {
		logger.Error("Unable to create POST request: %v\n", err)
		return err
	}

	req.SetBasicAuth(client.Creds.Username, client.Creds.Password)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Transmission-Session-Id", client.SessionID)

	resp, err := client.Client.Do(req)

	if err != nil {
		logger.Error("Error during POST request: %v\n", err)
		return err
	}
	defer resp.Body.Close()

	var body respTorrent
	err = json.NewDecoder(resp.Body).Decode(&body)
	if err != nil {
		logger.Error("Unable to parse response body JSON:%v\n%v\n", err, resp.Body)
		return err
	}

	if body.Result != "success" {
		return fmt.Errorf("Unable to add torrent: %v", body)
	}

	hashString := body.Arguments.TorrentAdded.HashString

	if len(hashString) == 0 {

		hashString = body.Arguments.TorrentDuplicate.HashString
		if len(hashString) != 0 {
			logger.Info("Torrent from hash %v duplicated\n", hashString)
		}
	}

	return nil
}
