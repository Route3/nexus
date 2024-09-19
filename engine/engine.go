package engine

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	hexutils "github.com/apex-fusion/nexus/helper/hex"
	"github.com/apex-fusion/nexus/types"
	"github.com/hashicorp/go-hclog"
)

const (
	JSONRPC                                 = "2.0"
	ExchangeTransitionConfigurationV1Method = "engine_exchangeTransitionConfigurationV1"
	ExchangeCapabilitiesMethod              = "engine_exchangeCapabilities"
	ForkchoiceUpdatedV1Method               = "engine_forkchoiceUpdatedV3"
	GetPayloadV1Method                      = "engine_getPayloadV3"
	NewPayloadV3Method                      = "engine_newPayloadV3"
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

	return engineClient, nil
}

func (c *Client) Init(latestPayloadHash string) (payloadId string, err error) {
	_, err = c.ExchangeCapabilities(make([]string, 0))
	if err != nil {
		return
	}

	_, err = c.ExchangeTransitionConfigurationV1()
	if err != nil {
		return
	}

	res, err := c.ForkChoiceUpdatedV1(latestPayloadHash, true)
	if err != nil {
		return
	}

	return res.Result.PayloadID, nil
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

	fmt.Println("httpClient: body response:", string(body))

	// Check if HTTP.status == 200 but some Geth error occured
	var potentialErrResp EngineResponseError

	err = json.Unmarshal(body, &potentialErrResp)

	if potentialErrResp.Error.Code != 0 {
		return fmt.Errorf("engine err: %v", potentialErrResp.Error)
	}

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
		Params:      []string{payloadId},
	}

	err = c.handleRequest(&requestData, &responseData)

	return
}

func (c *Client) NewPayloadV3(executionPayload *types.Payload, beaconBlockRoot string) (responseData *NewPayloadV3Response, err error) {
	c.logger.Debug("Running NewPayloadV3")

	params := []NewPayloadV3RequestParams{
		NewPayloadV3ExecutionPayloadParam{*executionPayload},
		make(NewPayloadV3ExpectedBlobVersionedHashes, 0),
		NewPayloadV3ParentBeaconBlockRoot(beaconBlockRoot),
	}

	requestData := NewPayloadV3Request{
		RequestBase: getRequestBase(NewPayloadV3Method),
		Params:      params,
	}

	err = c.handleRequest(&requestData, &responseData)

	return
}

func (c *Client) ForkChoiceUpdatedV1(blockHash string, buildPayload bool) (responseData *ForkchoiceUpdatedV1Response, err error) {
	c.logger.Debug("Running ForkchoiceUpdatedV1", "blockHash", blockHash)

	blockTimestamp := "0x" + fmt.Sprintf("%X", time.Now().Unix())

	params := []ForkchoiceUpdatedV1Param{
		ForkchoiceStateParam{
			HeadBlockHash:      blockHash,
			SafeBlockHash:      blockHash,
			FinalizedBlockHash: blockHash,
		},
		nil,
	}

	if buildPayload {
		params[1] = ForkchoicePayloadAttributes{
			Timestamp:             blockTimestamp,
			PrevRandao:            "0x0000000000000000000000000000000000000000000000000000000000000000", // TODO
			SuggestedFeeRecipient: "0x0000000000000000000000000000000000000000",
		}
	}
	requestData := ForkchoiceUpdatedV1Request{
		RequestBase: getRequestBase(ForkchoiceUpdatedV1Method),
		Params:      params,
	}

	err = c.handleRequest(&requestData, &responseData)
	if err != nil {
		return
	}

	if responseData.Result.PayloadStatus.Status != "VALID" {
		err = fmt.Errorf("engine error: payload status is not VALID! actual value:", responseData.Result.PayloadStatus.Status)
	}

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

func GetPayloadV1ResponseToPayload(resp *GetPayloadV1Response) (payload *types.Payload, err error) {
	// TODO: handle potential conversion errors and implement this as a json.Unmarshal method

	payload = new(types.Payload)

	payload.BaseFeePerGas = hexutils.DecodeHexToBig(string(hexutils.DropHexPrefix([]byte(resp.Result.BaseFeePerGas)))) // TODO: Make it prettier
	payload.BlockHash = types.StringToHash(resp.Result.BlockHash)
	payload.ExtraData, _ = hexutils.DecodeString(string(hexutils.DropHexPrefix([]byte(resp.Result.ExtraData)))) // TODO: Make it prettier
	payload.FeeRecipient = types.StringToAddress(resp.Result.FeeRecipient)
	payload.GasLimit, _ = hexutils.DecodeUint64(resp.Result.GasLimit)
	payload.GasUsed, _ = hexutils.DecodeUint64(resp.Result.GasUsed)

	// Logs bloom encoding
	var logsBloom types.Bloom
	input := hexutils.DropHexPrefix([]byte(resp.Result.LogsBloom))
	if _, err := hex.Decode(logsBloom[:], input); err != nil {
		return nil, err
	}
	payload.LogsBloom = logsBloom

	payload.Number, _ = hexutils.DecodeUint64(resp.Result.BlockNumber)
	payload.ParentHash = types.StringToHash(resp.Result.ParentHash)
	payload.ReceiptsRoot = types.StringToHash(resp.Result.ReceiptsRoot)
	payload.StateRoot = types.StringToHash(resp.Result.StateRoot)
	payload.Timestamp, _ = hexutils.DecodeUint64(resp.Result.Timestamp)
	payload.Transactions = [][]byte{}

	for _, transaction := range resp.Result.Transactions {
		decoded, err := hexutils.DecodeHex(transaction)
		if err != nil {
			return nil, err
		}
		payload.Transactions = append(payload.Transactions, decoded)
	}

	return
}
