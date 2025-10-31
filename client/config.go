package main

import (
	"flag"
	"fmt"
	"strings"
	"time"
)

type Config struct {
	ClientID        string
	ClientPort      string // Port this client listens
	PrintServerAddr string
	PeerAddresses   []string // List of peer client addresses
	RequestInterval time.Duration
	AutoMode        bool // Enable automatic print request generation
}

func ParseConfig() (*Config, error) {
	config := &Config{}

	var peersFlag string

	flag.StringVar(&config.ClientID, "id", "", "Client ID (required)")
	flag.StringVar(&config.ClientPort, "port", "", "Port to listen on (required, e.g., 50052)")
	flag.StringVar(&config.PrintServerAddr, "server", "localhost:50051", "Print server address")
	flag.StringVar(&peersFlag, "peers", "", "Comma-separated list of peer addresses (e.g., localhost:50053,localhost:50054)")
	flag.BoolVar(&config.AutoMode, "auto", false, "Enable automatic print request generation")

	var intervalSeconds int
	flag.IntVar(&intervalSeconds, "interval", 10, "Print request interval in seconds")

	flag.Parse()

	// Validate required fields
	if config.ClientID == "" {
		return nil, fmt.Errorf("client ID is required (use -id flag)")
	}
	if config.ClientPort == "" {
		return nil, fmt.Errorf("client port is required (use -port flag)")
	}

	if peersFlag != "" {
		config.PeerAddresses = strings.Split(peersFlag, ",")
		for i := range config.PeerAddresses {
			config.PeerAddresses[i] = strings.TrimSpace(config.PeerAddresses[i])
		}
	}

	config.RequestInterval = time.Duration(intervalSeconds) * time.Second

	return config, nil
}
