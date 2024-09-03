package engine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const (
	JSONRPC                    = "2.0"
	ExchangeCapabilitiesMethod = "engine_exchangeCapabilities"
	ForkchoiceUpdatedV1Method  = "engine_forkchoiceUpdatedV1"
	GetPayloadV1Method         = "engine_getPayloadV1"
	NewPayloadV1Method         = "engine_newPayloadV1"
)

type Client struct {
	client *http.Client
	url    *url.URL
	token  string
}

func NewClient(rawUrl string, token string, jwtId string) (*Client, error) {
	url, err := url.Parse(rawUrl)
	if err != nil {
		return nil, err
	}

	authTransport := &jwtTransport{
		underlyingTransport: http.DefaultTransport,
		jwtSecret:           []byte(token),
		jwtId:               jwtId,
	}

	client := &http.Client{
		Timeout:   DefaultRPCHTTPTimeout,
		Transport: authTransport,
	}

	return &Client{
		client,
		url,
		token,
	}, nil
}

func getRequestBase(method string) RequestBase {
	return RequestBase{
		JsonRPC: JSONRPC,
		Method:  method,
		ID:      0,
	}
}

// Encode the request data to JSON
func (c *Client) handleRequest(requestData interface{}, responseData interface{}) error {
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return fmt.Errorf("failed to marshal request data: %v", err)
	}

	resp, err := c.client.Post(c.url.String(), "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status code: %d", resp.StatusCode)
	}

	// Read the entire response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}

	err = json.Unmarshal(body, &responseData)
	if err != nil {
		return fmt.Errorf("failed to unmarshal response: %v", err)
	}

	return nil
}

func (c *Client) GetPayloadV1(payloadId string) (responseData *GetPayloadV1Response, err error) {
	requestData := GetPayloadV1Request{
		RequestBase: getRequestBase(GetPayloadV1Method),
		Params:      []string{payloadId},
	}

	err = c.handleRequest(&requestData, &responseData)

	return
}

func (c *Client) NewPayloadV1(executionPayload NewPayloadV1RequestParams) (responseData *NewPayloadV1Response, err error) {
	requestData := NewPayloadV1Request{
		RequestBase: getRequestBase(NewPayloadV1Method),
		Params:      []NewPayloadV1RequestParams{executionPayload},
	}

	err = c.handleRequest(&requestData, &responseData)

	return
}

func (c *Client) ForkchoiceUpdatedV1(blockHash string, suggestedFeeRecipient string) (responseData *ForkchoiceUpdatedV1Response, err error) {
	requestData := ForkchoiceUpdatedV1Request{
		RequestBase: getRequestBase(ForkchoiceUpdatedV1Method),
		Params: []ForkchoiceUpdatedV1Param{
			ForkchoiceStateParam{
				// HeadBlockHash:      blockHash,
				// SafeBlockHash:      blockHash,
				// FinalizedBlockHash: blockHash,
				HeadBlockHash:      "0x3559e851470f6e7bbed1db474980683e8c315bfce99b2a6ef47c057c04de7858",
				SafeBlockHash:      "0x3559e851470f6e7bbed1db474980683e8c315bfce99b2a6ef47c057c04de7858",
				FinalizedBlockHash: "0x3b8fb240d288781d4aac94d3fd16809ee413bc99294a085798a589dae51ddd4a",
			},
			ForkchoicePayloadAttributes{
				Timestamp:             "0x5",                                                                // TODO
				PrevRandao:            "0x0000000000000000000000000000000000000000000000000000000000000000", // TODO
				SuggestedFeeRecipient: suggestedFeeRecipient,
			},
		},
	}

	err = c.handleRequest(&requestData, &responseData)

	return
}

func (c *Client) ExchangeCapabilities(consesusCapabilites []string) (responseData *ExchangeCapabilitiesResponse, err error) {
	requestData := ExchangeCapabilitiesRequest{
		RequestBase: getRequestBase(ExchangeCapabilitiesMethod),
		Params:      [][]string{consesusCapabilites},
	}

	err = c.handleRequest(&requestData, &responseData)

	return
}
