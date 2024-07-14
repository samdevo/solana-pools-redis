package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	RedisAddress  string `json:"redis_address"`
	GeyserAddress string `json:"geyser_address"`
}

func LoadConfig() (Config, error) {
	file, err := os.Open("config/config.json")
	if err != nil {
		return Config{}, err
	}
	defer file.Close()

	var config Config
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return Config{}, err
	}

	return config, nil
}
