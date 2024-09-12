package engine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/apex-fusion/nexus/types"
	"github.com/hashicorp/go-hclog"
)

const (
	JSONRPC                                 = "2.0"
	ExchangeTransitionConfigurationV1Method = "engine_exchangeTransitionConfigurationV1"
	ExchangeCapabilitiesMethod              = "engine_exchangeCapabilities"
	ForkchoiceUpdatedV1Method               = "engine_forkchoiceUpdatedV1"
	GetPayloadV1Method                      = "engine_getPayloadV1"
	NewPayloadV1Method                      = "engine_newPayloadV1"
)

type Client struct {
	logger hclog.Logger
	client *http.Client
	url    *url.URL
	token  []byte
}

func NewClient(logger hclog.Logger, rawUrl string, token []byte, jwtId string) (*Client, error) {
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

	engineClient := &Client{
		logger.Named("engine"),
		client,
		url,
		token,
	}

	fmt.Println("------")
	fmt.Println("------")
	fmt.Println(rawUrl)
	fmt.Println(token)
	fmt.Println(jwtId)
	fmt.Println("------")
	fmt.Println("------")

	_, err = engineClient.ExchangeCapabilities(make([]string, 0))
	if err != nil {
		return nil, err
	}

	_, err = engineClient.ForkChoiceUpdatedV1("0x9e7be6b5ccb576cec0ab66a64639aab41e8edf604a93ccaa5c0073410c1e780d", "")
	if err != nil {
		return nil, err
	}

	_, err = engineClient.ExchangeTransitionConfigurationV1()
	if err != nil {
		return nil, err
	}


	return engineClient, nil
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

	fmt.Println(string(body))
	err = json.Unmarshal(body, &responseData)
	if err != nil {
		return fmt.Errorf("failed to unmarshal response: %v", err)
	}

	return nil
}

func (c *Client) ExchangeTransitionConfigurationV1() (responseData *ExchangeTransitionConfigurationV1Response, err error) {
	c.logger.Debug("Running ExchangeTransitionConfigurationV1")

	requestData := ExchangeTransitionConfigurationV1Request{
		RequestBase: getRequestBase(ExchangeTransitionConfigurationV1Method),
		Params: []ExchangeTransitionConfigurationV1RequestParams{
			{
				TerminalTotalDifficulty: "0x0",
				TerminalBlockHash:       "0x0000000000000000000000000000000000000000000000000000000000000000",
				TerminalBlockNumber:     "0x1",
			},
		},
	}

	err = c.handleRequest(&requestData, &responseData)

	return
}

func (c *Client) GetPayloadV1() (responseData *GetPayloadV1Response, err error) {
	c.logger.Debug("Running GetPayloadV1")
	requestData := GetPayloadV1Request{
		RequestBase: getRequestBase(GetPayloadV1Method),
		Params: []string {"0x01eab440f90637f9"}, //`payloadId` received in the first `forkChoiceUpdatedV1` response
	}

	err = c.handleRequest(&requestData, &responseData)

	return
}

func (c *Client) NewPayloadV1(executionPayload types.Payload) (responseData *NewPayloadV1Response, err error) {
	c.logger.Debug("Running NewPayloadV1")
	requestData := NewPayloadV1Request{
		RequestBase: getRequestBase(NewPayloadV1Method),
		Params:      []types.Payload{executionPayload},
	}

	err = c.handleRequest(&requestData, &responseData)

	return
}

func (c *Client) ForkChoiceUpdatedV1(blockHash string, suggestedFeeRecipient string) (responseData *ForkchoiceUpdatedV1Response, err error) {
	c.logger.Debug("Running ForkchoiceUpdatedV1")
	requestData := ForkchoiceUpdatedV1Request{
		RequestBase: getRequestBase(ForkchoiceUpdatedV1Method),
		Params: []ForkchoiceUpdatedV1Param{
			ForkchoiceStateParam{
				HeadBlockHash:      blockHash,
				SafeBlockHash:      blockHash,
				FinalizedBlockHash: blockHash,
				// HeadBlockHash:      "0x3559e851470f6e7bbed1db474980683e8c315bfce99b2a6ef47c057c04de7858",
				// SafeBlockHash:      "0x3559e851470f6e7bbed1db474980683e8c315bfce99b2a6ef47c057c04de7858",
				// FinalizedBlockHash: "0x3b8fb240d288781d4aac94d3fd16809ee413bc99294a085798a589dae51ddd4a",
			},
			// ForkchoicePayloadAttributes{
			// 	Timestamp:             "0x66dbe07b",                                                                // TODO
			// 	PrevRandao:            "0x0000000000000000000000000000000000000000000000000000000000000000", // TODO
			// 	SuggestedFeeRecipient: "0x0000000000000000000000000000000000000000",
			// },
		},
	}

	err = c.handleRequest(&requestData, &responseData)

	return
}

func (c *Client) ExchangeCapabilities(consesusCapabilites []string) (responseData *ExchangeCapabilitiesResponse, err error) {
	c.logger.Debug("Running ExchangeCapabilities")
	requestData := ExchangeCapabilitiesRequest{
		RequestBase: getRequestBase(ExchangeCapabilitiesMethod),
		Params:      [][]string{consesusCapabilites},
	}

	err = c.handleRequest(&requestData, &responseData)

	return
}
