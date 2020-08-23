package regression

import (
	"encoding/json"
	"fmt"
	"ledger-sats-stack/bus"
	"ledger-sats-stack/httpd"
	"ledger-sats-stack/httpd/svc"
	utils "ledger-sats-stack/tests"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
)

func TestBlocksRegression(t *testing.T) {
	b, err := bus.New(
		os.Getenv("RPC_URL"),
		os.Getenv("RPC_USER"),
		os.Getenv("RPC_PASS"),
		os.Getenv("RPC_NOTLS") == "true",
	)
	if err != nil {
		t.Fatalf("Failed to initialize Bus: %v", err)
	}
	defer b.Close()

	s := &svc.Service{
		Bus: b,
	}

	// Inject Gin router into an HTTP server
	engine := httpd.GetRouter(s)
	ts := httptest.NewServer(engine)
	defer ts.Close()

	for _, testCase := range BlocksTestCases {
		t.Run(testCase, func(t *testing.T) {
			baseEndpoint := fmt.Sprintf("blockchain/v3/btc/blocks/%s", testCase)
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
}
