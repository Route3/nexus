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
	JSONRPC                    = "2.0"
	ExchangeCapabilitiesMethod = "engine_exchangeCapabilities"
	ForkchoiceUpdatedV3Method  = "engine_forkchoiceUpdatedV3"
	GetPayloadV3Method         = "engine_getPayloadV3"
	NewPayloadV3Method         = "engine_newPayloadV3"
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

func (c *Client) Init(latestPayloadHash types.Hash, parentBeaconBlockRoot string) (payloadId string, err error) {
	_, err = c.ExchangeCapabilities(make([]string, 0))
	if err != nil {
		return
	}

	res, err := c.ForkChoiceUpdatedV3(latestPayloadHash, parentBeaconBlockRoot, true)
	if err != nil {
		return
	}

	return res.Result.PayloadID, nil
}

func (c *Client) retryIndefinitely(requestData interface{}, responseData interface{}) {
	for {
		err := c.handleRequest(requestData, responseData)

		// If no error, stop retrying
		if err == nil {
			break
		}

		c.logger.Error("engine API error, retrying indefinitely", "error", err)

		time.Sleep(2 * time.Second)
	}
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

func (c *Client) GetPayloadV3(payloadId string) (responseData *GetPayloadV3Response, err error) {
	c.logger.Debug("Running GetPayloadV3")
	requestData := GetPayloadV3Request{
		RequestBase: getRequestBase(GetPayloadV3Method),
		Params:      []string{payloadId},
	}

	c.retryIndefinitely(&requestData, &responseData)

	return
}

func (c *Client) NewPayloadV3(payload *types.Payload, beaconBlockRoot string) (responseData *NewPayloadV3Response, err error) {
	c.logger.Debug("Running NewPayloadV3")

	executionPayload := &NewPayloadV3ExecutionPayloadParam{
		ParentHash:      payload.ParentHash.String(),
		FeeRecipient:    payload.FeeRecipient.String(),
		StateRoot:       payload.StateRoot.String(),
		ReceiptsRoot:    payload.ReceiptsRoot.String(),
		LogsBloom:       payload.LogsBloom.String(),
		Random:          payload.Random.String(),
		Number:          hexutils.EncodeUint64(payload.Number),
		GasLimit:        hexutils.EncodeUint64(payload.GasLimit),
		GasUsed:         hexutils.EncodeUint64(payload.GasUsed),
		Timestamp:       hexutils.EncodeUint64(payload.Timestamp),
		ExtraData:       hexutils.EncodeToHex(payload.ExtraData),
		BaseFeePerGas:   hexutils.EncodeBig(payload.BaseFeePerGas),
		BlockHash:       payload.BlockHash.String(),
		Transactions:    make([]string, 0),
		Withdrawals:     make([]string, 0),
		ExcessBlobGas:   "0x0",
		BlobGasUsed:     "0x0",
		DepositRequests: nil,
	}

	for _, transaction := range payload.Transactions {
		decoded := hexutils.EncodeToHex(transaction)
		executionPayload.Transactions = append(executionPayload.Transactions, decoded)
	}

	params := []NewPayloadV3RequestParams{
		executionPayload,
		make(NewPayloadV3ExpectedBlobVersionedHashes, 0),
		NewPayloadV3ParentBeaconBlockRoot(beaconBlockRoot),
	}

	requestData := NewPayloadV3Request{
		RequestBase: getRequestBase(NewPayloadV3Method),
		Params:      params,
	}

	c.retryIndefinitely(&requestData, &responseData)

	return
}

func (c *Client) ForkChoiceUpdatedV3(blockHash types.Hash, parentBeaconBlockRoot string, buildPayload bool) (responseData *ForkchoiceUpdatedV3Response, err error) {
	c.logger.Debug("Running ForkChoiceUpdatedV3", "blockHash", blockHash)

	blockTimestamp := "0x" + fmt.Sprintf("%X", time.Now().Unix())

	params := []ForkchoiceUpdatedV3Param{
		ForkchoiceStateParam{
			HeadBlockHash:      blockHash.String(),
			SafeBlockHash:      blockHash.String(),
			FinalizedBlockHash: blockHash.String(),
		},
		nil,
	}

	if buildPayload {
		params[1] = ForkchoicePayloadAttributes{
			Timestamp:             blockTimestamp,
			PrevRandao:            "0x0000000000000000000000000000000000000000000000000000000000000000", // TODO
			SuggestedFeeRecipient: "0x0000000000000000000000000000000000000000",
			Withdrawals:           make([]string, 0),
			ParentBeaconBlockroot: parentBeaconBlockRoot,
		}
	}
	requestData := ForkchoiceUpdatedV3Request{
		RequestBase: getRequestBase(ForkchoiceUpdatedV3Method),
		Params:      params,
	}

	c.retryIndefinitely(&requestData, &responseData)

	if responseData.Result.PayloadStatus.Status == "SYNCING" {
		c.logger.Error("payload status is not VALID!", "status", responseData.Result.PayloadStatus.Status)
		return nil, fmt.Errorf("payload status is not VALID!")
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

func GetPayloadV3ResponseToPayload(resp *GetPayloadV3Response) (payload *types.Payload, err error) {
	// TODO: handle potential conversion errors and implement this as a json.Unmarshal method

	payload = new(types.Payload)
	rawPayload := resp.Result.ExecutionPayload

	payload.BaseFeePerGas = hexutils.DecodeHexToBig(string(hexutils.DropHexPrefix([]byte(rawPayload.BaseFeePerGas)))) // TODO: Make it prettier
	payload.BlockHash = types.StringToHash(rawPayload.BlockHash)
	payload.ExtraData, _ = hexutils.DecodeString(string(hexutils.DropHexPrefix([]byte(rawPayload.ExtraData)))) // TODO: Make it prettier
	payload.FeeRecipient = types.StringToAddress(rawPayload.FeeRecipient)
	payload.GasLimit, _ = hexutils.DecodeUint64(rawPayload.GasLimit)
	payload.GasUsed, _ = hexutils.DecodeUint64(rawPayload.GasUsed)

	// Logs bloom encoding
	var logsBloom types.Bloom
	input := hexutils.DropHexPrefix([]byte(rawPayload.LogsBloom))
	if _, err := hex.Decode(logsBloom[:], input); err != nil {
		return nil, err
	}
	payload.LogsBloom = logsBloom

	payload.Number, _ = hexutils.DecodeUint64(rawPayload.BlockNumber)
	payload.ParentHash = types.StringToHash(rawPayload.ParentHash)
	payload.ReceiptsRoot = types.StringToHash(rawPayload.ReceiptsRoot)
	payload.StateRoot = types.StringToHash(rawPayload.StateRoot)
	payload.Timestamp, _ = hexutils.DecodeUint64(rawPayload.Timestamp)
	payload.Transactions = [][]byte{}

	for _, transaction := range rawPayload.Transactions {
		decoded, err := hexutils.DecodeHex(transaction)
		if err != nil {
			return nil, err
		}
		payload.Transactions = append(payload.Transactions, decoded)
	}

	return
}
