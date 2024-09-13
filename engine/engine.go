package engine

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"strconv"
	"time"

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

	_, err = engineClient.ExchangeTransitionConfigurationV1()
	if err != nil {
		return nil, err
	}

	//TODO move this call somewhere else
	fcuResp, err := engineClient.ForkChoiceUpdatedV1_init("0x9e7be6b5ccb576cec0ab66a64639aab41e8edf604a93ccaa5c0073410c1e780d", "")
	if err != nil {
		return nil, err
	}
	fmt.Println("fcuResp:", fcuResp)

	//TODO remove --------------------------------------------------->
	execResp, err := engineClient.GetPayloadV1(fcuResp.Result.PayloadID)

	fmt.Println("execResp:", execResp)

	for i := 0; i < 5; i++ {

		fcuResp, err = engineClient.ForkChoiceUpdatedV1(execResp.Result.ParentHash, "")

		fmt.Println("fcuResp:", fcuResp)

		execResp, err = engineClient.GetPayloadV1(fcuResp.Result.PayloadID)

		GetPayloadV1ResponseToPayload(execResp)

		fmt.Println("execResp:", execResp)
	}

	panic("<------- done ")

	//TODO <--------------------------------------------------- remove

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

func (c *Client) GetPayloadV1(payloadId string) (responseData *GetPayloadV1Response, err error) {
	c.logger.Debug("Running GetPayloadV1")
	requestData := GetPayloadV1Request{
		RequestBase: getRequestBase(GetPayloadV1Method),
		// Params: []string {"0x01eab440f90637f9"}, //`payloadId` received in the first `forkChoiceUpdatedV1` response
		Params: []string {payloadId},
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

	blockTimestamp :=  "0x" + fmt.Sprintf("%X", time.Now().Unix())      
	
	fmt.Println("blockTimestamp:", blockTimestamp)

	requestData := ForkchoiceUpdatedV1Request{
		RequestBase: getRequestBase(ForkchoiceUpdatedV1Method),
		Params: []ForkchoiceUpdatedV1Param{
			ForkchoiceStateParam{
				HeadBlockHash:      blockHash,
				SafeBlockHash:      blockHash,
				FinalizedBlockHash: blockHash,
			},
			ForkchoicePayloadAttributes{
				Timestamp:             blockTimestamp,                                
				PrevRandao:            "0x0000000000000000000000000000000000000000000000000000000000000000", // TODO
				SuggestedFeeRecipient: "0x0000000000000000000000000000000000000000",
			},
		},
	}

	err = c.handleRequest(&requestData, &responseData)

	return
}

func (c *Client) ForkChoiceUpdatedV1_init(blockHash string, suggestedFeeRecipient string) (responseData *ForkchoiceUpdatedV1Response, err error) {
	c.logger.Debug("Running ForkchoiceUpdatedV1")

	requestData := ForkchoiceUpdatedV1Request{
		RequestBase: getRequestBase(ForkchoiceUpdatedV1Method),
		Params: []ForkchoiceUpdatedV1Param{
			ForkchoiceStateParam{
				HeadBlockHash:      blockHash,
				SafeBlockHash:      blockHash,
				FinalizedBlockHash: blockHash,
			},
			ForkchoicePayloadAttributes{
				Timestamp:             "0x66dbe07b",                           
				PrevRandao:            "0x0000000000000000000000000000000000000000000000000000000000000000", // TODO
				SuggestedFeeRecipient: "0x0000000000000000000000000000000000000000",
			},
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

func GetPayloadV1ResponseToPayload (resp *GetPayloadV1Response) (payload *types.Payload, err error) {
	//TODO handle potential conversion errors

	payload = new (types.Payload)

	payload.BaseFeePerGas = new (big.Int)
	payload.BaseFeePerGas.SetString(resp.Result.BaseFeePerGas[2:], 16)

	payload.BlockHash = types.StringToHash(resp.Result.BlockHash)

	payload.ExtraData, _ = hex.DecodeString(resp.Result.ExtraData[2:])

	payload.FeeRecipient = types.StringToAddress(resp.Result.FeeRecipient)

	payload.GasLimit, _ = strconv.ParseUint(resp.Result.GasLimit, 16, 64)

	payload.GasUsed, _ = strconv.ParseUint(resp.Result.GasUsed, 16, 64)

	payload.LogsBloom, _ = hex.DecodeString(resp.Result.LogsBloom[2:])

	payload.Number, _ = strconv.ParseUint(resp.Result.BlockNumber, 16, 64)

	payload.ParentHash = types.StringToHash(resp.Result.ParentHash)

	// TODO check if neccesary
	// payload.Random = types.StringToHash(resp.Result.PrevRandao)

	payload.ReceiptsRoot = types.StringToHash(resp.Result.ReceiptsRoot)

	payload.StateRoot = types.StringToHash(resp.Result.StateRoot)

	payload.Timestamp, _ = strconv.ParseUint(resp.Result.Timestamp, 16, 64)

	// TODO: handle this ?
	// payload.Transactions = resp.Result.???

	fmt.Println("- payload ------>", payload)

	return
}