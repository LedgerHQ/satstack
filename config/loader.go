package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/mitchellh/go-homedir"
)

func LoadConfig() (*Configuration, error) {
	home, err := homedir.Dir()
	if err != nil {
		return nil, fmt.Errorf("home directory not found: %w", err)
	}

	configLookupPaths := []string{
		// TODO: Add Ledger Live user data folder
		"lss.json",
		path.Join(home, "lss.json"),
	}

	var configPath string
	for _, maybePath := range configLookupPaths {
		if fileExists(maybePath) {
			configPath = maybePath
			break
		}
	}

	if configPath == "" {
		return nil, ErrConfigFileNotFound
	}

	configuration, err := loadFromPath(configPath)
	if err != nil {
		return nil, fmt.Errorf("malformed JSON config: %w", err)
	}

	if err := configuration.validate(); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	return configuration, nil
}

// fileExists checks if a file exists and is not a directory before we
// try using it to prevent further errors.
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func loadFromPath(path string) (*Configuration, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer func() {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}()

	decoder := json.NewDecoder(file)

	configuration := &Configuration{}
	err = decoder.Decode(configuration)
	if err != nil {
		return nil, err
	}

	return configuration, nil
}
