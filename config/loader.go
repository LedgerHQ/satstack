package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"runtime"

	log "github.com/sirupsen/logrus"

	"github.com/mitchellh/go-homedir"
)

// Load reads the config file from disk and returns a Configuration.
//
// It searches for the config file in a standard set of directories, in the
// following order:
//  1. Ledger Live user data folder.
//  2. Current directory.
//  3. User's home directory.
//
// The filename is always expected to be lss.json.
func Load() (*Configuration, error) {
	paths, err := configLookupPaths()
	if err != nil {
		return nil, err
	}

	var configPath string
	for _, maybePath := range paths {
		if fileExists(maybePath) {
			configPath = maybePath
			break
		}
	}

	if configPath == "" {
		return nil, ErrConfigFileNotFound
	}

	log.WithField("path", configPath).Info("Config file detected")

	configuration, err := loadFromPath(configPath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ErrMalformed, err)
	}

	if err := configuration.validate(); err != nil {
		return nil, fmt.Errorf("%s: %w", ErrValidation, err)
	}

	return configuration, nil
}

func LoadRescanConf() (*ConfigurationRescan, error) {
	paths, err := configRescanLookupPaths()
	if err != nil {
		return nil, err
	}

	var configPath string
	for _, maybePath := range paths {
		if fileExists(maybePath) {
			configPath = maybePath
			break
		}
	}

	if configPath == "" {
		return nil, ErrConfigFileNotFound
	}

	log.WithField("path", configPath).Info("Rescan Config file detected")

	configuration, err := loadFromPathRescan(configPath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ErrMalformed, err)
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

func loadFromPathRescan(path string) (*ConfigurationRescan, error) {
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

	configuration := &ConfigurationRescan{}
	err = decoder.Decode(configuration)
	if err != nil {
		return nil, err
	}

	return configuration, nil
}

func configLookupPaths() ([]string, error) {
	home, err := homedir.Dir()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ErrHomeNotFound, err)
	}

	return []string{
		path.Join(liveUserDataFolder(home), "lss.json"),
		"lss.json",
		path.Join(home, ".satstack", "lss.json"),
		path.Join(home, "lss.json"),
	}, nil
}

func configRescanLookupPaths() ([]string, error) {
	home, err := homedir.Dir()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ErrHomeNotFound, err)
	}

	return []string{
		path.Join(liveUserDataFolder(home), "lss_rescan.json"),
		"lss_rescan.json",
		path.Join(home, ".satstack", "lss_rescan.json"),
		path.Join(home, "lss_rescan.json"),
	}, nil
}

func liveUserDataFolder(home string) string {
	switch runtime.GOOS {
	case "linux":
		return path.Join(home, ".config", "Ledger Live")
	case "darwin":
		return path.Join(home, "Library", "Application Support", "Ledger Live")
	case "windows":
		return path.Join(home, "AppData", "Roaming", "Ledger Live")
	default:
		return path.Join(home, ".config", "Ledger Live")
	}
}
