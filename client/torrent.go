package client

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/whatust/transmission-rss/logger"
)

func getTorrent(link string, clt Client, torrentFilename string) error {

	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
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
