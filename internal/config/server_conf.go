package config

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

func GetServerConfig() ServerConfig {
	var config ServerConfig
	var addressF, sIntervalF, sFIleF, keyF, DSNf, privateKeyPathF, configPathF string
	var restoreF bool

	flag.StringVar(&addressF, "a", "",
		fmt.Sprintf("server socket (default: %s)", serverAddressDefault))
	flag.StringVar(&sIntervalF, "i", "",
		fmt.Sprintf("store interval (default: %s)", storeIntervalDefault))
	flag.StringVar(&sFIleF, "f", storeFileDefault,
		fmt.Sprintf("store file (default: %s)", storeFileDefault))
	flag.BoolVar(&restoreF, "r", restoreDefault, "restore from file at start")
	flag.StringVar(&keyF, "k", "", "key for HMAC (if not set responses will not be signed and hash from agent will not be checked)")
	flag.StringVar(&DSNf, "d", "", "database connection string (postgres://username:password@localhost:5432/database_name)")
	flag.StringVar(&privateKeyPathF, "crypto-key", "", "path to private key to decryt messages with")
	flag.StringVar(&configPathF, "c", "", "configuration path to use")
	flag.Parse()

	addressEnv := os.Getenv("ADDRESS")
	sIntervalEnv := os.Getenv("STORE_INTERVAL")
	sFileEnv := os.Getenv("STORE_FILE")
	restoreEnv := os.Getenv("RESTORE")
	keyEnv := os.Getenv("KEY")
	DSNenv := os.Getenv("DATABASE_DSN")
	privateKeyPathEnv := os.Getenv("CRYPTO_KEY")
	configPathEnv := os.Getenv("CONFIG")

	// checking config file
	var configJSON ServerConfigJSON
	var err error
	if configPathEnv != "" {
		configJSON, err = loadServerConfig(configPathEnv)
		if err != nil {
			log.Printf("WARN failed to read config file %s: %s. File will be ignored", configPathEnv, err.Error())
		}
	} else if configPathF != "" {
		configJSON, err = loadServerConfig(configPathF)
		if err != nil {
			log.Printf("WARN failed to read config file %s: %s. File will be ignored", configPathEnv, err.Error())
		}
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

	// storeFile
	if sFileEnv != "" {
		config.StoreFile = sFileEnv
	} else if sFIleF != "" {
		config.StoreFile = sFIleF
	} else if configJSON.StoreFile != "" {
		config.StoreFile = configJSON.StoreFile
	} else {
		config.StoreFile = storeFileDefault
	}

	// restore
	if restoreEnv != "" {
		restore, err := strconv.ParseBool(restoreEnv)
		if err != nil {
			log.Printf("WARN failed to get `restore` value from env var (%s): %s. Default value (%t) will be used",
				restoreEnv, err.Error(), restoreDefault)
			config.Restore = restoreDefault
		} else {
			config.Restore = restore
		}
	} else if isFlagPassed("r") {
		config.Restore = restoreF
	} else if configJSON.Restore != nil {
		config.Restore = *configJSON.Restore
	} else {
		config.Restore = restoreDefault
	}

	// storeInterval
	if sIntervalEnv != "" {
		d, err := time.ParseDuration(sIntervalEnv)
		if err != nil {
			log.Printf("WARN can`t parse STORE_INTERVAL env variable (%s): %s. Default value will be used (%s)",
				sIntervalEnv, err.Error(), storeIntervalDefault)
			config.StoreInterval = storeIntervalDefault
		} else {
			config.StoreInterval = d
		}
	} else if sIntervalF != "" {
		d, err := time.ParseDuration(sIntervalF)
		if err != nil {
			log.Printf("WARN can`t parse  '-i' flag (%s): %s. Default value will be used (%s)",
				sIntervalF, err.Error(), storeIntervalDefault)
			config.StoreInterval = storeIntervalDefault
		} else {
			config.StoreInterval = d
		}
	} else if configJSON.StoreInterval != "" {
		d, err := time.ParseDuration(configJSON.StoreInterval)
		if err != nil {
			log.Printf("WARN can`t parse  'store_interval' configuration attribute (%s): %s."+
				"Default value will be used (%s)",
				configJSON.StoreInterval, err.Error(), storeIntervalDefault)
			config.StoreInterval = storeIntervalDefault
		} else {
			config.StoreInterval = d
		}
	} else {
		config.StoreInterval = storeIntervalDefault
	}

	// key
	if keyEnv != "" {
		config.Key = keyEnv
	} else if keyF != "" {
		config.Key = keyF
	} else if configJSON.Key != "" {
		config.Key = configJSON.Key
	}

	// DSN
	if DSNenv != "" {
		config.DSN = DSNenv
	} else if DSNf != "" {
		config.DSN = DSNf
	} else {
		config.DSN = configJSON.DSN
	}

	// UseDB
	if config.DSN != "" {
		config.UseDB = true
	} else {
		config.UseDB = false
	}

	// PrivateKeyPath
	if privateKeyPathEnv != "" {
		config.PrivateKeyPath = privateKeyPathEnv
	} else if privateKeyPathF != "" {
		config.PrivateKeyPath = privateKeyPathF
	} else {
		config.PrivateKeyPath = configJSON.PrivateKeyPath
	}

	return config
}
