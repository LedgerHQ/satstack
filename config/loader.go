package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"runtime"

	"github.com/mitchellh/go-homedir"
)

// Load reads the config file from disk and returns a Configuration.
//
// It searches for the config file in a standard set of directories, in the
// following order:
//   1. Ledger Live user data folder.
//   2. Current directory.
//   3. User's home directory.
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

func configLookupPaths() ([]string, error) {
	home, err := homedir.Dir()
	if err != nil {
		return nil, fmt.Errorf("home directory not found: %w", err)
	}

	return []string{
		liveUserDataFolder(home),
		"lss.json",
		path.Join(home, "lss.json"),
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
