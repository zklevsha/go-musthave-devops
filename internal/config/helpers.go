package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

// loadServerConfig загружает конфигурацию сервера из json файла
func loadServerConfig(path string) (ServerConfigJSON, error) {
	jsonFile, err := os.Open(path)
	if err != nil {
		e := fmt.Errorf("failed open json file %s: %s", path, err.Error())
		return ServerConfigJSON{}, e
	}
	defer jsonFile.Close()

	var config ServerConfigJSON
	err = json.NewDecoder(jsonFile).Decode(&config)
	if err != nil {
		e := fmt.Errorf("failed to decode json file: %s", err.Error())
		return ServerConfigJSON{}, e
	}
	return config, nil
}

//loadAgentConfig загружает конфигурацию агента из json файла
func loadAgentConfig(path string) (AgentConfigJSON, error) {
	jsonFile, err := os.Open(path)
	if err != nil {
		e := fmt.Errorf("failed open json file %s: %s", path, err.Error())
		return AgentConfigJSON{}, e
	}
	defer jsonFile.Close()

	var config AgentConfigJSON
	err = json.NewDecoder(jsonFile).Decode(&config)
	if err != nil {
		e := fmt.Errorf("failed to decode json file: %s", err.Error())
		return AgentConfigJSON{}, e
	}
	return config, nil
}

// isFlagPassed checks if specific flag were passed at startup
func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}
