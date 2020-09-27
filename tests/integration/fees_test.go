package integration

import (
	"fmt"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/onyb/sat-stack/bus"
	"github.com/onyb/sat-stack/httpd"
	"github.com/onyb/sat-stack/httpd/svc"
	utils "github.com/onyb/sat-stack/tests"
)

func TestFeesIntegration(t *testing.T) {
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

	endpoint := fmt.Sprintf("%s/blockchain/v3/%s/fees", ts.URL, b.Currency)
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
}
