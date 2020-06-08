package regression

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"ledger-sats-stack/pkg/httpd"
	utils "ledger-sats-stack/tests"
	"net/http/httptest"
	"os"
	"path"
	"reflect"
	"testing"
)

func TestTransactionsRegression(t *testing.T) {
	// Setup phase
	xrpc := httpd.GetXRPC(
		os.Getenv("BITCOIND_RPC_HOST"),
		os.Getenv("BITCOIND_RPC_USER"),
		os.Getenv("BITCOIND_RPC_PASSWORD"),
		os.Getenv("BITCOIND_RPC_ENABLE_TLS") == "true",
	)
	// Inject Gin router into an HTTP server
	ts := httptest.NewServer(httpd.GetRouter(xrpc))

	for _, testCase := range TransactionTestCases {
		t.Run(testCase, func(t *testing.T) {
			baseEndpoint := fmt.Sprintf("blockchain/v3/transactions/%s", testCase)
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

	// Teardown phase
	xrpc.Shutdown()
	ts.Close()
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

func loadRecordingOrGetUrl(transactionHash string) ([]byte, error) {
	cassette := path.Join("fixtures", "transactions", fmt.Sprintf("%s.json", transactionHash))
	if _, err := os.Stat(cassette); os.IsNotExist(err) {
		baseEndpoint := fmt.Sprintf("blockchain/v3/transactions/%s", transactionHash)
		endpoint := fmt.Sprintf("http://bitcoin-mainnet.explorers.prod.aws.ledger.fr/%s", baseEndpoint)
		responseBytes, err := utils.GetResponseBytes(endpoint)
		if err != nil {
			return nil, err
		}

		// prettify JSON
		jsonArray, err := utils.LoadJSONArray(responseBytes)
		if err != nil {
			return nil, err
		}
		indentedResponseBytes, err := json.MarshalIndent(jsonArray, "", "  ")
		if err != nil {
			return nil, err
		}
		if err := ioutil.WriteFile(cassette, indentedResponseBytes, 0644); err != nil {
			return nil, err
		}
		return responseBytes, nil
	}
	return ioutil.ReadFile(cassette)
}
