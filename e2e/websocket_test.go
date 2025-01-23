package e2e

import (
	"encoding/json"
	"testing"

	"github.com/apex-fusion/nexus/jsonrpc"
	"github.com/gorilla/websocket"
)

type testWSRequest struct {
	JSONRPC string   `json:"jsonrpc"`
	Params  []string `json:"params"`
	Method  string   `json:"method"`
	ID      int      `json:"id"`
}

func constructWSRequest(id int, method string, params []string) ([]byte, error) {
	request := testWSRequest{
		JSONRPC: "2.0",
		Method:  method,
		ID:      id,
		Params:  params,
	}

	return json.Marshal(request)
}

func getWSResponse(t *testing.T, ws *websocket.Conn, request []byte) jsonrpc.SuccessResponse {
	t.Helper()

	if wsError := ws.WriteMessage(websocket.TextMessage, request); wsError != nil {
		t.Fatalf("Unable to write message to WS connection: %v", wsError)
	}

	_, response, wsError := ws.ReadMessage()

	if wsError != nil {
		t.Fatalf("Unable to read message from WS connection: %v", wsError)
	}

	var res jsonrpc.SuccessResponse
	if wsError = json.Unmarshal(response, &res); wsError != nil {
		t.Fatalf("Unable to unmarshal WS response: %v", wsError)
	}

	return res
}
