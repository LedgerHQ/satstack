package regression

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

const (
	WarningColor = "\033[1;33m%s\033[0m"
	ErrorColor   = "\033[1;31m%s\033[0m"
)

func GetResponseBytes(endpoint string) ([]byte, error) {
	response, err := http.Get(endpoint)
	if err != nil {
		return nil, err
	}
	responseBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	return responseBytes, nil
}

func LoadJSON(bytes []byte) (interface{}, error) {
	var jsonArrayValue interface{}
	if err := json.Unmarshal(bytes, &jsonArrayValue); err != nil {
		return nil, err
	}
	return jsonArrayValue, nil
}

func LoadJSONArray(bytes []byte) ([]interface{}, error) {
	var jsonArrayValue []interface{}
	if err := json.Unmarshal(bytes, &jsonArrayValue); err != nil {
		return nil, err
	}
	return jsonArrayValue, nil
}
