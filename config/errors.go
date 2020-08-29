package config

import "errors"

var (
	// ErrMissingAccounts indicates that the configuration file did not
	// provide any accounts to import.
	ErrMissingAccounts = errors.New("no accounts in config file")

	// ErrMissingKeyInConfig indicates that a key was expected in the config,
	// but was not found.
	ErrMissingKeyInConfig = errors.New("missing key in config")

	// ErrConfigFileNotFound indicates that no config file was found in any of
	// the standard paths.
	ErrConfigFileNotFound = errors.New("config file not found")

	// ErrMalformedConfig indicates that a config file was found, but the
	// JSON contents could not be decoded. This does not indicate a validation
	// error.
	ErrMalformedConfig = errors.New("malformed JSON config")

	// ErrValidation indicates a validation error in the config.
	ErrValidation = errors.New("validation error")

	// ErrHomeNotFound indicates that an error was encountered while obtaining
	// the user's home directory.
	ErrHomeNotFound = errors.New("home directory not found")
)
