package tests

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"ledger-sats-stack/pkg/httpd"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
)

const (
	SuccessColor = "\033[1;32m%s\033[0m"
	NoticeColor  = "\033[1;36m%s\033[0m"
	WarningColor = "\033[1;33m%s\033[0m"
	ErrorColor   = "\033[1;31m%s\033[0m"
	DebugColor   = "\033[0;36m%s\033[0m"
)

func TestRegression(t *testing.T) {
	wire := httpd.GetWire(
		os.Getenv("BITCOIND_RPC_HOST"),
		os.Getenv("BITCOIND_RPC_USER"),
		os.Getenv("BITCOIND_RPC_PASSWORD"),
		os.Getenv("BITCOIND_RPC_ENABLE_TLS") == "true",
	)

	// Inject Gin router into an HTTP server
	ts := httptest.NewServer(httpd.GetRouter(wire))
	// Shut down the server and block until all requests have gone through

	fmt.Printf(SuccessColor, "[OK]\tSetup\n")

	defer func() {
		fmt.Printf(SuccessColor, "[OK]\tTeardown\n")
		defer wire.Shutdown()
		ts.Close()
	}()

	check := func(endpoint string) {
		remoteEndpoint := fmt.Sprintf(
			"http://bitcoin-mainnet.explorers.prod.aws.ledger.fr/%s",
			endpoint,
		)

		localEndpoint := fmt.Sprintf(
			"%s/%s",
			ts.URL,
			endpoint,
		)

		remoteResponse, err := http.Get(remoteEndpoint)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		localResponse, err := http.Get(localEndpoint)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if localResponse.StatusCode != remoteResponse.StatusCode {
			t.Fatalf("Expected status code %v, got %v", remoteResponse.StatusCode, localResponse.StatusCode)
		}

		var localResponseJSON interface{}
		var remoteResponseJSON interface{}

		localResponseBytes, _ := ioutil.ReadAll(localResponse.Body)
		remoteResponseBytes, _ := ioutil.ReadAll(remoteResponse.Body)

		json.Unmarshal([]byte(localResponseBytes), &localResponseJSON)
		json.Unmarshal([]byte(remoteResponseBytes), &remoteResponseJSON)

		if !reflect.DeepEqual(localResponseJSON, remoteResponseJSON) {
			fmt.Printf(WarningColor, fmt.Sprintf("\t%s\n", string(localResponseBytes)))
			fmt.Printf(WarningColor, fmt.Sprintf("\t%s\n", string(remoteResponseBytes)))
			fmt.Printf(ErrorColor, fmt.Sprintf("[FAIL]\t%sRegression found\n", endpoint))
			t.Fatalf("Regression found")
		}

		fmt.Printf(SuccessColor, fmt.Sprintf("[OK]\t%s\n", endpoint))
	}

	check("blockchain/v3/explorer/_health")
	check("blockchain/v3/blocks/current")
}
