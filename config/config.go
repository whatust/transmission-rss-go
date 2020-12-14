package config

import (
	"gopkg.in/yaml.v3"
	"os"
)

// Server struct used to parse yaml file
type Server struct {
	Host string					`yaml:"host"`
	Port int					`yaml:"port"`
	TLS bool					`yaml:"tls"`
	RPCPath string				`yaml:"rpcPath"`
	Retries int					`yaml:"retries"`
	WaitTime int				`yaml:"waitTime"`
	Proxy string				`yaml:"proxy"`
	ProxyPort string			`yaml:"proxyPort"`
}

// Log struct used to parse yaml file
type Log struct {
	LogPath string 				`yaml:"logPath"`
	Level string				`yaml:"level"`
	MaxSize int					`yaml:"maxSize"`
	MaxBackups int				`yaml:"maxBackups"`
	MaxAge int					`yaml:"maxAge"`
	Compress bool				`yaml:"compress"`
	LocalTime bool				`yaml:"localTime"`
	Formatter string			`yaml:"formatter"`
}

// Creds struct used to parse yaml file
type Creds struct {
	Username string				`yaml:"username"`
	Password string 			`yaml:"password"`
}

// Config struct used to parse yaml file
type Config struct {
	ServerConfig Server			`yaml:"server"`
	LogConfig Log				`yaml:"log"`
	CredsConfig Creds			`yaml:"login"`
	SeenFile string				`yaml:"seenFile"`
	RSSFile string				`yaml:"rssFile"`
	UIDType string				`yaml:"uID"`
}

// GetConfig parse config file for transmission-rss
func GetConfig(configFilename string) (*Config, error) {

	configFile, err := os.OpenFile(configFilename, os.O_RDONLY, 0644)

	if err != nil {
		return nil, err
	}
	defer configFile.Close()

	var config Config
	decoder := yaml.NewDecoder(configFile)
	err = decoder.Decode(&config)

	if err != nil {
		return nil, err
	}

	return &config, nil
}

// Matcher struct used to parse yaml file
type Matcher struct {
	RegExp string				`yaml:"regexp"`
	DownloadPath string			`yaml:"downloadPath"`
	ValidateCert bool			`yaml:"validateCert"`
	IgnoreRemake bool			`yaml:"ignoreRemake"`
	OnlyTrusted bool			`yaml:"onlyTrusted"`
}

// Feed strcut used to parse yaml file
type Feed struct {
	URL string 					`yaml:"url"`
	DefaultDownloadPath string	`yaml:"defaultDownloadPath"`
	DefaultIgnoreRemake string	`yaml:"defaultIgnoreRemake"`
	DefaultValidateCert string	`yaml:"defaultValidateCert"`
	SeedRatioLimit int			`yaml:"seedRationLimit"`
	Matchers []Matcher			`yaml:"matchers"`
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
