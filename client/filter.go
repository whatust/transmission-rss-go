package client

import (
	"fmt"
	"github.com/whatust/transmission-rss/config"
	"github.com/whatust/transmission-rss/logger"
	"regexp"
)

// Filter ...
type Filter struct {
	RegExp       *regexp.Regexp
	DownloadPath string
	ValidateCert bool
	IgnoreRemake bool
	OnlyTrusted  bool
}

// CreateFilter creates filter to match torrent
func CreateFilter(matcher config.Matcher, def config.Feed) (*Filter, error) {

	filter := Filter{
		RegExp:       regexp.MustCompile(matcher.RegExp),
		DownloadPath: matcher.DownloadPath,
		ValidateCert: matcher.ValidateCert,
		IgnoreRemake: matcher.IgnoreRemake,
		OnlyTrusted:  matcher.OnlyTrusted,
	}

	if len(filter.DownloadPath) == 0 {
		filter.DownloadPath = def.DefaultDownloadPath
	}

	if len(filter.DownloadPath) == 0 {
		return nil, fmt.Errorf("Download path must be set")
	}

	return &filter, nil
}

// FilterTorrent ...
func FilterTorrent(torrent FeedItem, filter *Filter) bool {

	if !filter.IgnoreRemake && torrent.Remake == "Yes" {
		logger.Debug("Ignoring remake: %v\n", torrent.Title)
		return false
	}

	if filter.OnlyTrusted && torrent.Trusted == "Yes" {
		logger.Debug("Ignoring untrusted torrent: %v\n", torrent.Title)
		return false
	}

	matched := filter.RegExp.Match([]byte(torrent.Title))

	if !matched {
		logger.Debug(
			"Title does not match regex:\nRegex:%v\nTitle:%v\n",
			filter.RegExp,
			torrent.Title,
		)
		return false
	}

	return true
}
