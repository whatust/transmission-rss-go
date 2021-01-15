package client

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"fmt"
	"net/http"
	"github.com/whatust/transmission-rss/logger"
)

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

type addFile struct {
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

func getTorrent(link string, clt Client, torrentFilename string) error {

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

// AddTorrentURL adds a torrent to transmission server from an URL
func addTorrentURL(url string, client *RPCClient, path string) (string, error) {

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
		logger.Error("%v\n", data)
		return "", err
	}

	req, err := http.NewRequest("POST", client.URL, bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Error("Unable to create POST request: %v\n", err)
		return "", err
	}

	req.SetBasicAuth(client.Creds.Username, client.Creds.Password)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Transmission-Session-Id", client.SessionID)

	resp, err := client.Client.Do(req)

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
		if len(hashString) != 0 {
			logger.Info("Torrent from %v and hash %v duplicated\n", url, hashString)
		}
	}

	return hashString, nil
}

func addTorrent(torrentPath string, client *RPCClient, path string) (string, error) {

	torrentData, err := ioutil.ReadFile(torrentPath)
	if err != nil {
		return "", err
	}

	torrent := base64.StdEncoding.EncodeToString([]byte(torrentData))

	data := addFile{
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

	req, err := http.NewRequest("POST", client.URL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.SetBasicAuth(client.Creds.Username, client.Creds.Password)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Transmission-Session-Id", client.SessionID)

	resp, err := client.Client.Do(req)
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