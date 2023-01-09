package config

import (
	"encoding/json"
	"os"

	log "github.com/sirupsen/logrus"
)

// WriteRescanConf writes the rescan information into a file
// when it does not exist it saves it to the same location
// where the lss.json is stored
func WriteRescanConf(data *ConfigurationRescan) error {
	paths, err := configRescanLookupPaths()
	if err != nil {
		return err
	}

	var configPath string
	for _, maybePath := range paths {
		if fileExists(maybePath) {
			configPath = maybePath
			break
		}
	}

	if configPath == "" {
		// if the file does not exist, save to home dir
		// check where the lss.json lies and take the same path
		lssPath, err := configLookupPaths()
		if err != nil {
			return err
		}

		for index, maybePath := range lssPath {
			if fileExists(maybePath) {
				configPath = paths[index]
				break
			}
		}
	}
	// This should never happen, in case we have no lss.json
	// we should fail before
	if configPath == "" {
		return ErrConfigFileNotFound
	}

	// Writing to file

	file, _ := json.MarshalIndent(*data, "", " ")
	ferr := os.WriteFile(configPath, file, 0644)
	if ferr != nil {
		log.Error("Error savng last timestamp to file %s: %s", configPath, ferr)
		return err
	}

	log.WithField("path", configPath).Info("RescanConfigFile successfully saved")

	return nil
}
