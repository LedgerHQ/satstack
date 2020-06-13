package integration

import (
	"fmt"
	"ledger-sats-stack/httpd"
	utils "ledger-sats-stack/tests"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestFeesIntegration(t *testing.T) {
	// Setup phase
	xrpc := httpd.GetBus(
		os.Getenv("BITCOIND_RPC_HOST"),
		os.Getenv("BITCOIND_RPC_USER"),
		os.Getenv("BITCOIND_RPC_PASSWORD"),
		os.Getenv("BITCOIND_RPC_ENABLE_TLS") == "true",
	)
	// Inject Gin router into an HTTP server
	ts := httptest.NewServer(httpd.GetRouter(xrpc))

	endpoint := fmt.Sprintf("%s/blockchain/v3/%s/fees", ts.URL, xrpc.Currency)
	responseBytes, err := utils.GetResponseBytes(endpoint)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	responseJSON, err := utils.LoadJSON(responseBytes)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	fees := responseJSON.(map[string]interface{})
	if !(fees["6"].(float64) <= fees["3"].(float64) && fees["3"].(float64) <= fees["2"].(float64)) {
		t.Fatalf("Fees are inconsistent: %v", fees)
	}

	if fees["last_updated"].(float64) > float64(time.Now().Unix()) {
		t.Fatalf("last_updated value is greater than current timestamp: %v", fees)
	}

	// Teardown phase
	xrpc.Shutdown()
	ts.Close()
}
