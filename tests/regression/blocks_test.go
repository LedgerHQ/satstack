package regression

import (
	"encoding/json"
	"fmt"
	"ledger-sats-stack/pkg/httpd"
	utils "ledger-sats-stack/tests"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
)

func TestBlocksRegression(t *testing.T) {
	// Setup phase
	xrpc := httpd.GetXRPC(
		os.Getenv("BITCOIND_RPC_HOST"),
		os.Getenv("BITCOIND_RPC_USER"),
		os.Getenv("BITCOIND_RPC_PASSWORD"),
		os.Getenv("BITCOIND_RPC_ENABLE_TLS") == "true",
	)
	// Inject Gin router into an HTTP server
	ts := httptest.NewServer(httpd.GetRouter(xrpc))

	for _, testCase := range BlocksTestCases {
		t.Run(testCase, func(t *testing.T) {
			baseEndpoint := fmt.Sprintf("blockchain/v3/blocks/%s", testCase)
			localEndpoint := fmt.Sprintf("%s/%s", ts.URL, baseEndpoint)
			remoteEndpoint := fmt.Sprintf("http://bitcoin-mainnet.explorers.prod.aws.ledger.fr/%s", baseEndpoint)

			localResponseBytes, err := utils.GetResponseBytes(localEndpoint)
			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}
			remoteResponseBytes, err := utils.GetResponseBytes(remoteEndpoint)
			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}

			var localResponseJSON, remoteResponseJSON interface{}
			var localErr, remoteErr error
			if testCase == "current" {
				localResponseJSON, localErr = utils.LoadJSON(localResponseBytes)
				remoteResponseJSON, remoteErr = utils.LoadJSON(remoteResponseBytes)

				localCurrentHeight := localResponseJSON.(map[string]interface{})["height"]
				remoteCurrentHeight := remoteResponseJSON.(map[string]interface{})["height"]
				if localCurrentHeight != remoteCurrentHeight {
					return
				}
			} else {
				localResponseJSON, localErr = utils.LoadJSONArray(localResponseBytes)
				remoteResponseJSON, remoteErr = utils.LoadJSONArray(remoteResponseBytes)
			}
			if localErr != nil {
				t.Fatalf("Expected no error, got %v", err)
			}
			if remoteErr != nil {
				t.Fatalf("Expected no error, got %v", err)
			}

			if !reflect.DeepEqual(localResponseJSON, remoteResponseJSON) {
				localOutput, _ := json.Marshal(localResponseJSON)
				remoteOutput, _ := json.Marshal(remoteResponseJSON)
				fmt.Printf(utils.WarningColor, fmt.Sprintf("\tLocal  -> %s\n", string(localOutput)))
				fmt.Printf(utils.WarningColor, fmt.Sprintf("\tRemote -> %s\n", string(remoteOutput)))
				fmt.Printf(utils.ErrorColor, fmt.Sprintf("[FAIL]\t%s\n", baseEndpoint))
				t.Errorf("Regression found\n")
			}
		})
	}
	// Teardown phase
	xrpc.Shutdown()
	ts.Close()
}
