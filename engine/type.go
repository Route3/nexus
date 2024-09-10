package engine

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

type GetPayloadV1Request struct {
	RequestBase
	Params []string `json:"params"`
}

type GetPayloadV1ResponseResult struct {
	ParentHash    string `json:"parentHash"`
	FeeRecipient  string `json:"feeRecipient"`
	StateRoot     string `json:"stateRoot"`
	ReceiptsRoot  string `json:"receiptsRoot"`
	LogsBloom     string `json:"logsBloom"`
	PrevRandao    string `json:"prevRandao"`
	BlockNumber   string `json:"blockNumber"`
	GasLimit      string `json:"gasLimit"`
	GasUsed       string `json:"gasUsed"`
	Timestamp     string `json:"timestamp"`
	ExtraData     string `json:"extraData"`
	BaseFeePerGas string `json:"baseFeePerGas"`
	BlockHash     string `json:"blockHash"`
}

type GetPayloadV1Response struct {
	Result GetPayloadV1ResponseResult `json:"result"`
}

type NewPayloadV1RequestParams struct {
	ParentHash    string   `json:"parentHash"    gencodec:"required"`
	FeeRecipient  string   `json:"feeRecipient"  gencodec:"required"`
	StateRoot     string   `json:"stateRoot"     gencodec:"required"`
	ReceiptsRoot  string   `json:"receiptsRoot"  gencodec:"required"`
	LogsBloom     string   `json:"logsBloom"     gencodec:"required"`
	Random        string   `json:"prevRandao"    gencodec:"required"` // TODO:see if really needed
	Number        string   `json:"blockNumber"   gencodec:"required"`
	GasLimit      string   `json:"gasLimit"      gencodec:"required"`
	GasUsed       string   `json:"gasUsed"       gencodec:"required"`
	Timestamp     string   `json:"timestamp"     gencodec:"required"`
	ExtraData     string   `json:"extraData"     gencodec:"required"`
	BaseFeePerGas string   `json:"baseFeePerGas" gencodec:"required"`
	BlockHash     string   `json:"blockHash"     gencodec:"required"`
	Transactions  []string `json:"transactions"  gencodec:"required"`
}

type NewPayloadV1Request struct {
	RequestBase
	Params []NewPayloadV1RequestParams `json:"params"`
}

type NewPayloadV1ResponseResult struct {
	Status          string  `json:"status"`
	LatestValidHash string  `json:"latestValidHash"`
	ValidationError *string `json:"validationError"`
}

type NewPayloadV1Response struct {
	Result NewPayloadV1ResponseResult `json:"result"`
}

type ForkchoiceUpdatedV1Param interface {
	isForkchoiceUpdatedV1Param() bool
}

type ForkchoiceStateParam struct {
	HeadBlockHash      string `json:"headBlockHash"`
	SafeBlockHash      string `json:"safeBlockHash"`
	FinalizedBlockHash string `json:"finalizedBlockHash"`
}

func (s ForkchoiceStateParam) isForkchoiceUpdatedV1Param() bool {
	return true
}

type ForkchoicePayloadAttributes struct {
	Timestamp             string `json:"timestamp"`
	PrevRandao            string `json:"prevRandao"`
	SuggestedFeeRecipient string `json:"suggestedFeeRecipient"`
}

func (s ForkchoicePayloadAttributes) isForkchoiceUpdatedV1Param() bool {
	return true
}

type ForkchoiceUpdatedV1Request struct {
	RequestBase
	Params []ForkchoiceUpdatedV1Param `json:"params"`
}

type ForkchoiceUpdatedV1ResponsePayloadStatus struct {
	Status          string  `json:"status"`
	LatestValidHash string  `json:"latestValidHash"`
	ValidationError *string `json:"validationError"`
}

type ForkchoiceUpdatedV1ResponseResult struct {
	PayloadStatus ForkchoiceUpdatedV1ResponsePayloadStatus `json:"payloadStatus"`
	PayloadID     string                                   `json:"payloadId"`
}

type ForkchoiceUpdatedV1Response struct {
	Result ForkchoiceUpdatedV1ResponseResult `json:"result"`
}

type ExchangeCapabilitiesRequest struct {
	RequestBase
	Params [][]string `json:"params"`
}

type ExchangeCapabilitiesResponse struct {
	Result []string `json:"result"`
}
