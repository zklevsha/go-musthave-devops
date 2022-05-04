package config

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

const pollIntervalDefault = time.Duration(2 * time.Second)
const reportIntervalDefault = time.Duration(10 * time.Second)
const serverAddressDefault = "127.0.0.1:8080"
const storeIntervalDefault = time.Duration(300 * time.Second)
const storeFileDefault = "/tmp/devops-metrics-db.json"
const restoreDefault = true

type AgentConfig struct {
	PollInterval   time.Duration
	ReportInterval time.Duration
	ServerAddress  string
}

type ServerConfig struct {
	ServerAddress string
	StoreInterval time.Duration
	StoreFile     string
	Restore       bool
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

	var addressF, reportF, pollF string
	flag.StringVar(&addressF, "a", serverAddressDefault,
		fmt.Sprintf("server socket (default: %s)", serverAddressDefault))
	flag.StringVar(&reportF, "r", reportIntervalDefault.String(),
		fmt.Sprintf("report interval (default: %s)", reportIntervalDefault))
	flag.StringVar(&pollF, "p", pollIntervalDefault.String(),
		fmt.Sprintf("poll interval (default: %s)", pollIntervalDefault))
	flag.Parse()

	pollEnv := os.Getenv("POLL_INTERVAL")
	reportEnv := os.Getenv("REPORT_INTERVAL")
	addressEnv := os.Getenv("ADDRESS")

	// pollInterval
	pollInterval, err := parseInterval(pollEnv, pollF)
	if err != nil {
		log.Printf("WARN cant parse pollInterval (env:%s, flag: %s): %s. Default value will be used (%s)",
			pollEnv, pollF, err.Error(), pollIntervalDefault)
		config.PollInterval = pollIntervalDefault
	} else {
		config.PollInterval = pollInterval
	}

	// reportInterval
	reportInterval, err := parseInterval(reportEnv, reportF)
	if err != nil {
		log.Printf("WARN can`t parse reportInterval (env:%s, flag: %s): %s. Default value will be used (%s)",
			reportEnv, reportF, err.Error(), reportIntervalDefault)
		config.ReportInterval = reportIntervalDefault
	} else {
		config.ReportInterval = reportInterval
	}

	// address
	if addressEnv != "" {
		config.ServerAddress = addressEnv
	} else {
		config.ServerAddress = addressF
	}

	return config
}

func GetServerConfig() ServerConfig {
	var config ServerConfig
	var addressF, sIntervalF, sFIleF string
	var restoreF bool

	flag.StringVar(&addressF, "a", serverAddressDefault,
		fmt.Sprintf("server socket (default: %s)", serverAddressDefault))
	flag.StringVar(&sIntervalF, "i", storeIntervalDefault.String(),
		fmt.Sprintf("store interval (default: %s)", storeIntervalDefault))
	flag.StringVar(&sFIleF, "f", storeFileDefault,
		fmt.Sprintf("store file (default: %s)", storeFileDefault))
	flag.BoolVar(&restoreF, "r", restoreDefault, "restore from file at start")
	flag.Parse()

	addressEnv := os.Getenv("ADDRESS")
	sIntervalEnv := os.Getenv("STORE_INTERVAL")
	sFileEnv := os.Getenv("STORE_FILE")
	restoreEnv := os.Getenv("RESTORE")

	// address
	if addressEnv != "" {
		config.ServerAddress = addressEnv
	} else {
		config.ServerAddress = addressF
	}

	// storeFile
	if sFileEnv != "" {
		config.StoreFile = sFileEnv
	} else {
		config.StoreFile = sFIleF
	}

	// restore
	if restoreEnv != "" {
		restore, err := strconv.ParseBool(restoreEnv)
		if err != nil {
			log.Printf("WARN failed to get `restore` value from env var (%s): %s", restoreEnv, err.Error())
			config.Restore = restoreF
		} else {
			config.Restore = restore
		}
	} else {
		config.Restore = restoreF
	}

	// storeInterval
	sInterval, err := parseInterval(sIntervalEnv, sIntervalF)
	if err != nil {
		log.Printf("WARN can`t parse storeInterval (env:%s, flag: %s): %s. Default value will be used (%s)",
			sIntervalEnv, sIntervalF, err.Error(), storeIntervalDefault)
		config.StoreInterval = storeIntervalDefault
	} else {
		config.StoreInterval = sInterval
	}
	return config

}
