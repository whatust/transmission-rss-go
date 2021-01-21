package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

// Server struct used to parse yaml file
type Server struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	TLS      bool   `yaml:"tls"`
	RateTime int    `yaml:"rateTime"`
	RPCPath  string `yaml:"rpcPath"`
	Proxy    string `yaml:"proxy"`
	//ProxyPort    string `yaml:"proxyPort"`
	ValidateCert bool `yaml:"validateCert"`
}

// Connect struct used to parse yaml file
type Connect struct {
	Retries  int `yaml:"retries"`
	WaitTime int `yaml:"waitTime"`
	Timeout  int `yaml:"timeout"`
	RateTime int `yaml:"rateTime"`
}

// Log struct used to parse yaml file
type Log struct {
	LogPath    string `yaml:"logPath"`
	Level      string `yaml:"level"`
	MaxSize    int    `yaml:"maxSize"`
	MaxBackups int    `yaml:"maxBackups"`
	MaxAge     int    `yaml:"maxAge"`
	Compress   bool   `yaml:"compress"`
	LocalTime  bool   `yaml:"localTime"`
	Formatter  string `yaml:"formatter"`
}

// Creds struct used to parse yaml file
type Creds struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// Config struct used to parse yaml file
type Config struct {
	Server   Server  `yaml:"server"`
	Log      Log     `yaml:"log"`
	Creds    Creds   `yaml:"login"`
	Connect  Connect `yaml:"connection"`
	SeenFile string  `yaml:"seenFile"`
	RSSFile  string  `yaml:"rssFile"`
	UIDType  string  `yaml:"uID"`
	//SaveTorrent bool    `yaml:"saveTorrent"`
	TorrentPath string `yaml:"torrentPath"`
}

// NewConfig ...
func NewConfig() Config {

	config := Config{
		Server: Server{
			RateTime:     600,
			ValidateCert: true,
			RPCPath:      "/transmission/rpc",
			Port:         9091,
			Proxy:        "",
			Host:         "localhost",
			TLS:          false,
		},
		Connect: Connect{
			Retries:  10,
			WaitTime: 5,
			Timeout:  10,
			RateTime: 600,
		},
		Log: Log{
			Level:      "Info",
			MaxSize:    10000,
			MaxBackups: 1,
			MaxAge:     10,
			LocalTime:  true,
			Formatter:  "JSON",
			Compress:   false,
			LogPath: "/var/log/transmission-rss-log.log",
		},
		SeenFile: "/etc/transmission-rss-seen.log",
		RSSFile:  "/etc/transmission-rss-feeds.yml",
	}
	return config
}

// GetConfig parse config file for transmission-rss
func GetConfig(configFilename string) (*Config, error) {

	configFile, err := os.OpenFile(configFilename, os.O_RDONLY, 0644)

	if err != nil {
		return nil, err
	}
	defer configFile.Close()

	config := NewConfig()
	decoder := yaml.NewDecoder(configFile)
	err = decoder.Decode(&config)

	if err != nil {
		return nil, err
	}

	return &config, nil
}

// Matcher struct used to parse yaml file
type Matcher struct {
	RegExp       string `yaml:"regexp"`
	DownloadPath string `yaml:"downloadPath"`
	IgnoreRemake bool   `yaml:"ignoreRemake"`
	OnlyTrusted  bool   `yaml:"onlyTrusted"`
}

// Feed strcut used to parse yaml file
type Feed struct {
	URL                 string    `yaml:"url"`
	DefaultDownloadPath string    `yaml:"defaultDownloadPath"`
	DefaultIgnoreRemake string    `yaml:"defaultIgnoreRemake"`
	DefaultValidateCert string    `yaml:"defaultValidateCert"`
	SeedRatioLimit      int       `yaml:"seedRationLimit"`
	Matchers            []Matcher `yaml:"matchers"`
	Proxy               string    `yaml:"proxy"`
	ValidateCert        bool      `yaml:"validateCert"`
}

// FeedConfig struct used to parse yaml file
type FeedConfig struct {
	Feeds []Feed
}

// GetFeedsConfig parse feed file for transmission-rss
func GetFeedsConfig(RSSFilename string) (*FeedConfig, error) {

	rssFile, err := os.OpenFile(RSSFilename, os.O_RDONLY, 0644)

	if err != nil {
		return nil, err
	}
	defer rssFile.Close()

	var rssFeed FeedConfig
	decoder := yaml.NewDecoder(rssFile)
	err = decoder.Decode(&rssFeed)

	if err != nil {
		return nil, err
	}

	return &rssFeed, nil
}
