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

func TestTransactionsRegression(t *testing.T) {
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

	for _, testCase := range TransactionTestCases {
		t.Run(testCase, func(t *testing.T) {
			baseEndpoint := fmt.Sprintf("blockchain/v3/btc/transactions/%s", testCase)
			localEndpoint := fmt.Sprintf("%s/%s", ts.URL, baseEndpoint)
			remoteEndpoint := fmt.Sprintf("http://bitcoin-mainnet.explorers.prod.aws.ledger.fr/%s", baseEndpoint)

			localResponseBytes, err := utils.GetResponseBytes(localEndpoint)
			if err != nil {
				fmt.Printf(utils.ErrorColor, fmt.Sprintf("Could not fetch local endpoint: %v", err))
				t.Skip()
			}
			localResponseJSON, err := utils.LoadJSONArray(localResponseBytes)
			if err != nil {
				errorJSON, _ := utils.LoadJSON(localResponseBytes)
				fmt.Printf(utils.ErrorColor, fmt.Sprintf("Could not parse local response: %v", errorJSON))
				t.Skip()
			}

			remoteResponseBytes, err := utils.GetResponseBytes(remoteEndpoint)
			if err != nil {
				fmt.Printf(utils.ErrorColor, fmt.Sprintf("Could not fetch remote endpoint: %v", err))
				t.Skip()
			}
			remoteResponseJSON, err := utils.LoadJSONArray(remoteResponseBytes)
			if err != nil {
				errorJSON, _ := utils.LoadJSON(remoteResponseBytes)
				fmt.Printf(utils.ErrorColor, fmt.Sprintf("Could not parse remote response: %v", errorJSON))
				t.Skip()
			}

			// Transform remote JSON
			//   - remove keys for which value changes over time.
			//   - sanitize for known bugs in Ledger Blockchain Explorer
			for _, transaction := range localResponseJSON {
				deleteConfirmations(transaction)
				deleteInputIndexes(transaction)
			}
			for idx, transaction := range remoteResponseJSON {
				deleteConfirmations(transaction)
				deleteInputIndexes(transaction)
				normalizeInputsOrder(localResponseJSON[idx], transaction)
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

func normalizeInputsOrder(localTransaction interface{}, remoteTransaction interface{}) {
	localTxn := localTransaction.(map[string]interface{})
	remoteTxn := remoteTransaction.(map[string]interface{})

	equivalent := true
	for _, each := range localTxn["inputs"].([]interface{}) {
		if !containsInput(remoteTxn["inputs"].([]interface{}), each) {
			equivalent = false
		}
	}
	if equivalent {
		remoteTxn["inputs"] = localTxn["inputs"]
	}
}

func deleteInputIndexes(transaction interface{}) {
	typedTransaction := transaction.(map[string]interface{})
	for _, input := range typedTransaction["inputs"].([]interface{}) {
		i := input.(map[string]interface{})
		delete(i, "input_index")
	}
}

func deleteConfirmations(transaction interface{}) {
	typedTransaction := transaction.(map[string]interface{})
	delete(typedTransaction, "confirmations")
}

func containsInput(inputs []interface{}, input interface{}) bool {
	for _, each := range inputs {
		if reflect.DeepEqual(each, input) {
			return true
		}
	}
	return false
}
