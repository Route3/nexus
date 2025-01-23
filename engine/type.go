package engine

import (
	"github.com/apex-fusion/nexus/types"
)

type RequestBase struct {
	JsonRPC string `json:"jsonrpc,omitempty"`
	Method  string `json:"method"`
	ID      uint   `json:"id"`
}

type ExchangeTransitionConfigurationV1RequestParams struct {
	TerminalTotalDifficulty string `json:"terminalTotalDifficulty" gencodec:"required"`
	TerminalBlockHash       string `json:"terminalBlockHash" gencodec:"required"`
	TerminalBlockNumber     string `json:"terminalBlockNumber" gencodec:"required"`
}

type ExchangeTransitionConfigurationV1Request struct {
	RequestBase
	Params []ExchangeTransitionConfigurationV1RequestParams `json:"params"`
}

type ExchangeTransitionConfigurationV1Response struct {
	TerminalTotalDifficulty string `json:"terminalTotalDifficulty" gencodec:"required"`
	TerminalBlockHash       string `json:"terminalBlockHash" gencodec:"required"`
	TerminalBlockNumber     string `json:"terminalBlockNumber" gencodec:"required"`
}

type GetPayloadV3Request struct {
	RequestBase
	Params []string `json:"params"`
}

type PayloadVersion struct {
	PayloadID string `json:"payloadId"`
}

type GetPayloadV3ResponseResult struct {
	ExecutionPayload types.Payload `json:"executionPayload"`
}

type GetPayloadV3Response struct {
	Result GetPayloadV3ResponseResult `json:"result"`
}

type NewPayloadV3Request struct {
	RequestBase
	Params []NewPayloadV3RequestParams `json:"params"`
}

type NewPayloadV3RequestParams interface {
	isNewPayloadV3RequestParams() bool
}

type NewPayloadV3ExecutionPayloadParam struct {
	ParentHash      string   `json:"parentHash"    gencodec:"required"`
	FeeRecipient    string   `json:"feeRecipient"  gencodec:"required"`
	StateRoot       string   `json:"stateRoot"     gencodec:"required"`
	ReceiptsRoot    string   `json:"receiptsRoot"  gencodec:"required"`
	LogsBloom       string   `json:"logsBloom"     gencodec:"required"`
	Random          string   `json:"prevRandao"    gencodec:"required"`
	Number          string   `json:"blockNumber"   gencodec:"required"`
	GasLimit        string   `json:"gasLimit"      gencodec:"required"`
	GasUsed         string   `json:"gasUsed"       gencodec:"required"`
	Timestamp       string   `json:"timestamp"     gencodec:"required"`
	ExtraData       string   `json:"extraData"     gencodec:"required"`
	BaseFeePerGas   string   `json:"baseFeePerGas" gencodec:"required"`
	BlockHash       string   `json:"blockHash"     gencodec:"required"`
	Transactions    []string `json:"transactions"  gencodec:"required"`
	Withdrawals     []string `json:"withdrawals"  gencodec:"required"`
	ExcessBlobGas   string   `json:"excessBlobGas" gencodec:"required"`
	BlobGasUsed     string   `json:"blobGasUsed" gencodec:"required"`
	DepositRequests *string  `json:"depositRequests" gencodec:"required"`
}

func (s NewPayloadV3ExecutionPayloadParam) isNewPayloadV3RequestParams() bool {
	return true
}

type NewPayloadV3ExpectedBlobVersionedHashes []string

func (s NewPayloadV3ExpectedBlobVersionedHashes) isNewPayloadV3RequestParams() bool {
	return true
}

type NewPayloadV3ParentBeaconBlockRoot string

func (s NewPayloadV3ParentBeaconBlockRoot) isNewPayloadV3RequestParams() bool {
	return true
}

type NewPayloadV3ResponseResult struct {
	Status          string  `json:"status"`
	LatestValidHash string  `json:"latestValidHash"`
	ValidationError *string `json:"validationError"`
}

type NewPayloadV3Response struct {
	Result NewPayloadV3ResponseResult `json:"result"`
}

type ForkchoiceUpdatedV3Param interface {
	isForkchoiceUpdatedV3Param() bool
}

type ForkchoiceStateParam struct {
	HeadBlockHash      string `json:"headBlockHash"`
	SafeBlockHash      string `json:"safeBlockHash"`
	FinalizedBlockHash string `json:"finalizedBlockHash"`
}

func (s ForkchoiceStateParam) isForkchoiceUpdatedV3Param() bool {
	return true
}

type ForkchoicePayloadAttributes struct {
	Timestamp             string   `json:"timestamp"`
	PrevRandao            string   `json:"prevRandao"`
	SuggestedFeeRecipient string   `json:"suggestedFeeRecipient"`
	Withdrawals           []string `json:"withdrawals"`
	ParentBeaconBlockroot string   `json:"parentBeaconBlockRoot"`
}

func (s ForkchoicePayloadAttributes) isForkchoiceUpdatedV3Param() bool {
	return true
}

type ForkchoiceUpdatedV3Request struct {
	RequestBase
	Params []ForkchoiceUpdatedV3Param `json:"params"`
}

type ForkchoiceUpdatedV3ResponsePayloadStatus struct {
	Status          string  `json:"status"`
	LatestValidHash string  `json:"latestValidHash"`
	ValidationError *string `json:"validationError"`
}

type ForkchoiceUpdatedV3ResponseResult struct {
	PayloadStatus ForkchoiceUpdatedV3ResponsePayloadStatus `json:"payloadStatus"`
	PayloadID     string                                   `json:"payloadId"`
}

type ForkchoiceUpdatedV3Response struct {
	Result ForkchoiceUpdatedV3ResponseResult `json:"result"`
}

type ExchangeCapabilitiesRequest struct {
	RequestBase
	Params [][]string `json:"params"`
}

type ExchangeCapabilitiesResponse struct {
	Result []string `json:"result"`
}

type EngineErrorBody struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type EngineResponseError struct {
	Error EngineErrorBody `json:"error"`
}

type EngineConfig struct {
	EngineTokenPath string `json:"engineTokenPath"`
	EngineURL       string `json:"engineURL"`
	EngineJWTID     string `json:"engineJWTID"`
}
