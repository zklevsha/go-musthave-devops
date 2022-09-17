// Package config implements parsing logic for Agent and Server configuration parameters
package config

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"
)

func GetAgentConfig(args []string) AgentConfig {
	var config AgentConfig

	f := flag.NewFlagSet("agent", flag.ExitOnError)
	var addressF, reportF, pollF, keyF, publicKeyPathF, configPathF, gAddressF string
	f.StringVar(&addressF, "a", "",
		fmt.Sprintf("server`s socket (default: %s)", serverAddressDefault))
	f.StringVar(&reportF, "r", "",
		fmt.Sprintf("report interval (default: %s)", serverAddressDefault))
	f.StringVar(&pollF, "p", "",
		fmt.Sprintf("poll interval (default: %s)", pollIntervalDefault))
	f.StringVar(&keyF, "k", "", "key for HMAC (if not set messages will not be signed)")
	f.StringVar(&publicKeyPathF, "crypto-key", "",
		"server`s public key to encrypt the messages with (if not set messages will not be encrypted)")
	f.StringVar(&configPathF, "c", "", "configuration file to use")
	f.StringVar(&gAddressF, "g", "",
		"server`s gRPC socket (if not set, metrics will be sent via REST)")
	f.Parse(args)

	pollEnv := os.Getenv("POLL_INTERVAL")
	reportEnv := os.Getenv("REPORT_INTERVAL")
	addressEnv := os.Getenv("ADDRESS")
	keyEnv := os.Getenv("KEY")
	publicKeyPathEnv := os.Getenv("CRYPTO_KEY")
	configPathEnv := os.Getenv("CONFIG")
	gAddressEnv := os.Getenv("GRPC_ADDRESS")

	// checking config file
	var configJSON AgentConfigJSON
	var err error
	if configPathEnv != "" {
		configJSON, err = loadAgentConfig(configPathEnv)
		if err != nil {
			log.Printf("WARN failed to read config file %s: %s. File will be ignored", configPathEnv, err.Error())
		}
	} else if configPathF != "" {
		configJSON, err = loadAgentConfig(configPathF)
		if err != nil {
			log.Printf("WARN failed to read config file %s: %s. File will be ignored", configPathF, err.Error())
		}
	}

	// pollInterval
	if pollEnv != "" {
		poll, err := time.ParseDuration(pollEnv)
		if err != nil {
			log.Printf("WARN failed to parse poll interval from 'POLL_INTERVAL'"+
				"enviroment variable (%s). Default value (%s) will be used", pollEnv, pollIntervalDefault)
			config.PollInterval = pollIntervalDefault
		} else {
			config.PollInterval = poll
		}
	} else if pollF != "" {
		poll, err := time.ParseDuration(pollF)
		if err != nil {
			log.Printf("WARN failed to parse poll interval from '-p' flag (%s). "+
				"Default value (%s) will be used", pollF, pollIntervalDefault)
			config.PollInterval = pollIntervalDefault
		} else {
			config.PollInterval = poll
		}
	} else if configJSON.PollInterval != "" {
		poll, err := time.ParseDuration(configJSON.PollInterval)
		if err != nil {
			log.Printf("WARN failed to parse poll interval from 'poll_interval' config attribute(%s). "+
				"Default value (%s) will be used", configJSON.PollInterval, pollIntervalDefault)
			config.PollInterval = pollIntervalDefault
		} else {
			config.PollInterval = poll
		}
	} else {
		config.PollInterval = pollIntervalDefault
	}

	// reportInterval
	if reportEnv != "" {
		report, err := time.ParseDuration(reportEnv)
		if err != nil {
			log.Printf("WARN failed to parse report interval from 'REPORT_INTERVAL'"+
				"enviroment variable (%s). Default value (%s) will be used", reportEnv, reportIntervalDefault)
			config.ReportInterval = reportIntervalDefault
		} else {
			config.ReportInterval = report
		}
	} else if reportF != "" {
		report, err := time.ParseDuration(reportF)
		if err != nil {
			log.Printf("WARN failed to parse report interval from '-r' flag (%s). "+
				"Default value (%s) will be used", reportF, reportIntervalDefault)
			config.ReportInterval = reportIntervalDefault
		} else {
			config.ReportInterval = report
		}

	} else if configJSON.ReportInterval != "" {
		report, err := time.ParseDuration(configJSON.ReportInterval)
		if err != nil {
			log.Printf("WARN failed to parse poll interval from 'report_interval' config attribute(%s). "+
				"Default value (%s) will be used", configJSON.ReportInterval, reportIntervalDefault)
			config.ReportInterval = reportIntervalDefault
		} else {
			config.ReportInterval = report
		}
	} else {
		config.ReportInterval = reportIntervalDefault
	}

	// address
	if addressEnv != "" {
		config.ServerAddress = addressEnv
	} else if addressF != "" {
		config.ServerAddress = addressF
	} else if configJSON.ServerAddress != "" {
		config.ServerAddress = configJSON.ServerAddress
	} else {
		config.ServerAddress = serverAddressDefault
	}

	// key
	if keyEnv != "" {
		config.Key = keyEnv
	} else if keyF != "" {
		config.Key = keyF
	} else {
		config.Key = configJSON.Key
	}

	// PublicKeyPath
	if publicKeyPathEnv != "" {
		config.PublicKeyPath = publicKeyPathEnv
	} else if publicKeyPathF != "" {
		config.PublicKeyPath = publicKeyPathF
	} else {
		config.PublicKeyPath = configJSON.PublicKeyPath
	}

	// gRPC address
	if gAddressEnv != "" {
		config.GRPCAddress = gAddressEnv
	} else if gAddressF != "" {
		config.GRPCAddress = gAddressF
	} else if configJSON.GRPCAddress != "" {
		config.GRPCAddress = configJSON.GRPCAddress
	}

	return config
}
