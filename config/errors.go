package config

import "errors"

var (
	// ErrMissingKey indicates that a key was expected in the config,
	// but was not found.
	ErrMissingKey = errors.New("missing key")

	// ErrConfigFileNotFound indicates that no config file was found in any of
	// the standard paths.
	ErrConfigFileNotFound = errors.New("config file not found")

	// ErrMalformed indicates that a config file was found, but the
	// JSON contents could not be decoded. This does not indicate a validation
	// error.
	ErrMalformed = errors.New("malformed JSON")

	// ErrValidation indicates a validation error in the config.
	ErrValidation = errors.New("validation error")

	// ErrHomeNotFound indicates that an error was encountered while obtaining
	// the user's home directory.
	ErrHomeNotFound = errors.New("home directory not found")
)
