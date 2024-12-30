package config

import (
	"encoding/json"
	"log"
	"os"
)

const (
	config_file_name = ".gatorconfig.json"
)

var FilePath string = getMainPath() + "/internal/config/" + config_file_name

type Config struct {
	DBURL           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func Read() (Config, error) {

	file, err := os.Open(FilePath)
	if err != nil {
		return Config{}, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	config := Config{}
	err = decoder.Decode(&config)
	if err != nil {
		return Config{}, err
	}

	return config, err

}

func write(cfg Config) error {

	file, err := os.Create(FilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(cfg)
	if err != nil {
		return err
	}

	return nil
}

func (cfg *Config) SetUser(username string) error {
	cfg.CurrentUserName = username
	return write(*cfg)
}

func getMainPath() string {
	path, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	return path
}
