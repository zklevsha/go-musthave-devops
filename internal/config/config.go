package config

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"
)

const pollIntervalDefault = time.Duration(2 * time.Second)
const reportIntervalDefault = time.Duration(10 * time.Second)
const serverAddressDefault = "127.0.0.1:8080"

type AgentConfig struct {
	PollInterval   time.Duration
	ReportInterval time.Duration
	ServerAddress  string
}

func parseInterval(env string, flag string) (time.Duration, error) {
	if env != "" && flag != "" {
		i, errEnv := time.ParseDuration(env)
		if errEnv == nil {
			return i, nil
		}
		log.Printf("WARN main failed to convert env var to time.Duration: %s.", errEnv.Error())
		i, errFlag := time.ParseDuration(flag)
		if errFlag == nil {
			return i, nil
		}
		return time.Duration(0),
			fmt.Errorf("failed to convert both env var and flag to time.Duration: errEnv=%s, errFlag=%s", errEnv, errFlag)
	}

	if env != "" {
		i, err := time.ParseDuration(env)
		if err != nil {
			return time.Duration(0),
				fmt.Errorf("failed to parse flag %s: %s", flag, err.Error())
		}
		return i, nil
	}

	if flag != "" {
		i, err := time.ParseDuration(flag)
		if err != nil {
			return time.Duration(0),
				fmt.Errorf("failed to convert env var to time.Duration: %s", err.Error())
		}
		return i, nil
	}

	return time.Duration(0), fmt.Errorf("both flag and env are empty")

}

func GetAgentConfig() AgentConfig {
	var config AgentConfig

	var aFlag, rFlag, pFlag string
	flag.StringVar(&aFlag, "a", "", "server socket (default: 127.0.0.1:8080)")
	flag.StringVar(&rFlag, "r", "", "report interval (default: 10s)")
	flag.StringVar(&pFlag, "p", "", "poll interval (default: 2s)")
	flag.Parse()

	pollEnv := os.Getenv("POLL_INTERVAL")
	reportEnv := os.Getenv("REPORT_INTERVAL")
	addressEnv := os.Getenv("ADDRESS")

	// pollInterval
	if pollEnv != "" || pFlag != "" {
		pollInterval, err := parseInterval(pollEnv, pFlag)
		if err != nil {
			log.Printf("WARN cant parse pollInterval (env:%s, flag: %s): %s. Default value will be used (%s)",
				pollEnv, pFlag, err.Error(), pollIntervalDefault)
			config.PollInterval = pollIntervalDefault
		} else {
			config.PollInterval = pollInterval
		}
	} else {
		config.PollInterval = pollIntervalDefault
	}

	// reportInterval
	if reportEnv != "" || rFlag != "" {
		reportInterval, err := parseInterval(reportEnv, rFlag)
		if err != nil {
			log.Printf("WARN cant parse reportInterval (env:%s, flag: %s): %s. Default value will be used (%s)",
				reportEnv, rFlag, err.Error(), reportIntervalDefault)
			config.ReportInterval = reportIntervalDefault
		} else {
			config.ReportInterval = reportInterval
		}
	} else {
		config.ReportInterval = reportIntervalDefault
	}

	// address
	if addressEnv != "" {
		config.ServerAddress = addressEnv
	} else if aFlag != "" {
		config.ServerAddress = aFlag
	} else {
		config.ServerAddress = serverAddressDefault
	}

	return config
}
