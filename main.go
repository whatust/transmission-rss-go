package main

import (
	"time"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"github.com/akamensky/argparse"
	"github.com/whatust/transmission-rss/logger"
	"github.com/whatust/transmission-rss/helper"
	"github.com/whatust/transmission-rss/config"
	"github.com/whatust/transmission-rss/client"
)

// Logger pointer to the logger object
func main () {

	parser := argparse.NewParser(
		"transmission-rss",
		"Automatically adds new torrents from a RSS feeds to transmission using its RPC client.",
	)

	configFile := parser.String(
		"c",
		"config",
		&argparse.Options{
			Required: true,
			Help: "Path to the config YAML file used for transmission client.",
		},
	)
	dry := parser.Flag(
		"d",
		"dry-run",
		&argparse.Options{
			Required: false,
			Help: "Prints out the added torrent files without send it to the RPC client.",
		},
	)
	daemon := parser.Flag(
		"f",
		"daemon",
		&argparse.Options{
			Required: false,
			Help: "Daemonize process.",
		},
	)

	// Parse input arguments
	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Fprintln(os.Stderr, parser.Usage(err))
		os.Exit(1)
	}

	// Load configurations
	conf, err := config.GetConfig(*configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configurations: %v\n", err)
	}

	// Configure logger
	logger.ConfigLogger(conf.Log)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not configure log layer: %v\n", err)
		os.Exit(1)
	}

	// Create signal handler when daemonized
	if *daemon {
		go signalHandler()
	}

	// Create transmission client
	myClient := client.TransmissionClient{}
	var client client.RSSClient = &myClient
	
	err = client.Initialize(conf)
	if err != nil {
		logger.Error("Could not initialize RPC client: %v", err)
		os.Exit(1)
	}
	logger.Info("Client Initialized\n")

	// Load seen torrents
	var seenTorrent helper.SeenSet = helper.SeenSet{
		Old: make(map[string]struct{}),
		New: make(map[string]struct{}),
	}

	err = seenTorrent.LoadSeen(conf.SeenFile)
	if err != nil {
		logger.Error("Could not load seen torrents: %v\n", err)
	}

	// Load RSS feed list
	feedConfig, err := config.GetFeedsConfig(conf.RSSFile)
	if err != nil {
		logger.Error("Could not parse RSS feed list: %v", err)
		os.Exit(1)
	}

	for true {

		// Populate torrent from the feed list
		client.AddFeeds(feedConfig.Feeds, &seenTorrent)

		// Save updates to seen torrents file
		err = seenTorrent.SaveSeen(conf.SeenFile)
		if err != nil {
			logger.Error("Unable to save seen torrents: %v\n", err)
		}

		if !*daemon {
			break
		}

		time.Sleep(time.Second * 300)
	}

	logger.Info("Dry Run: %v", *dry)
}

func cleanup() {

}

func signalHandler() {

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	s := <-sigs

	logger.Info("Signal received: %v\n", s)

	cleanup()
	os.Exit(1)
}
