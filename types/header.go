package types

import (
	"encoding/binary"
	encodingHex "encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"sync/atomic"

	"github.com/apex-fusion/nexus/helper/hex"
)

// Header represents a block header in the Ethereum blockchain.
type Header struct {
	ParentHash   Hash
	Sha3Uncles   Hash
	Miner        []byte
	StateRoot    Hash
	TxRoot       Hash
	ReceiptsRoot Hash
	LogsBloom    Bloom
	Difficulty   uint64
	Number       uint64
	GasLimit     uint64
	GasUsed      uint64
	Timestamp    uint64
	ExtraData    []byte
	MixHash      Hash
	Nonce        Nonce
	Hash         Hash
	PayloadHash  Hash
}

func (h *Header) Equal(hh *Header) bool {
	return h.Hash == hh.Hash
}

func (h *Header) HasBody() bool {
	return h.TxRoot != EmptyRootHash || h.Sha3Uncles != EmptyUncleHash
}

func (h *Header) HasReceipts() bool {
	return h.ReceiptsRoot != EmptyRootHash
}

func (h *Header) SetNonce(i uint64) {
	binary.BigEndian.PutUint64(h.Nonce[:], i)
}

func (h *Header) IsGenesis() bool {
	return h.Hash != ZeroHash && h.Number == 0
}

type Nonce [8]byte

func (n Nonce) String() string {
	return hex.EncodeToHex(n[:])
}

// MarshalText implements encoding.TextMarshaler
func (n Nonce) MarshalText() ([]byte, error) {
	return []byte(n.String()), nil
}

func (h *Header) Copy() *Header {
	newHeader := &Header{
		ParentHash:   h.ParentHash,
		Sha3Uncles:   h.Sha3Uncles,
		StateRoot:    h.StateRoot,
		TxRoot:       h.TxRoot,
		ReceiptsRoot: h.ReceiptsRoot,
		MixHash:      h.MixHash,
		Hash:         h.Hash,
		LogsBloom:    h.LogsBloom,
		Nonce:        h.Nonce,
		Difficulty:   h.Difficulty,
		Number:       h.Number,
		GasLimit:     h.GasLimit,
		GasUsed:      h.GasUsed,
		Timestamp:    h.Timestamp,
		PayloadHash:  h.PayloadHash,
	}

	newHeader.Miner = make([]byte, len(h.Miner))
	copy(newHeader.Miner[:], h.Miner[:])

	newHeader.ExtraData = make([]byte, len(h.ExtraData))
	copy(newHeader.ExtraData[:], h.ExtraData[:])

	return newHeader
}

type Body struct {
	Transactions     []*Transaction
	Uncles           []*Header
	ExecutionPayload *Payload
}

type Block struct {
	Header           *Header
	Transactions     []*Transaction
	Uncles           []*Header
	ExecutionPayload *Payload
	// Cache
	size atomic.Value // *uint64
}

type Payload struct {
	ParentHash    Hash     `json:"parentHash"    gencodec:"required"`
	FeeRecipient  Address  `json:"feeRecipient"  gencodec:"required"`
	StateRoot     Hash     `json:"stateRoot"     gencodec:"required"`
	ReceiptsRoot  Hash     `json:"receiptsRoot"  gencodec:"required"`
	LogsBloom     Bloom    `json:"logsBloom"     gencodec:"required"`
	Random        Hash     `json:"prevRandao"    gencodec:"required"` // TODO:see if really needed
	Number        uint64   `json:"blockNumber"   gencodec:"required"`
	GasLimit      uint64   `json:"gasLimit"      gencodec:"required"`
	GasUsed       uint64   `json:"gasUsed"       gencodec:"required"`
	Timestamp     uint64   `json:"timestamp"     gencodec:"required"`
	ExtraData     []byte   `json:"extraData"     gencodec:"required"`
	BaseFeePerGas *big.Int `json:"baseFeePerGas" gencodec:"required"`
	BlockHash     Hash     `json:"blockHash"     gencodec:"required"`
	Transactions  [][]byte `json:"transactions"  gencodec:"required"`
	/*Withdrawals   []*types.Withdrawal `json:"withdrawals"`
	BlobGasUsed   *uint64             `json:"blobGasUsed"`
	ExcessBlobGas *uint64             `json:"excessBlobGas"`*/
}

type _RawPayload struct {
	ParentHash    Hash     `json:"parentHash"`
	FeeRecipient  Address  `json:"feeRecipient"`
	StateRoot     Hash     `json:"stateRoot"`
	ReceiptsRoot  Hash     `json:"receiptsRoot"`
	LogsBloom     string   `json:"logsBloom"`
	PrevRandao    string   `json:"prevRandao"`
	BlockNumber   string   `json:"blockNumber"`
	GasLimit      string   `json:"gasLimit"`
	GasUsed       string   `json:"gasUsed"`
	Timestamp     string   `json:"timestamp"`
	ExtraData     string   `json:"extraData"`
	BaseFeePerGas string   `json:"baseFeePerGas"`
	BlockHash     Hash     `json:"blockHash"`
	Transactions  []string `json:"transactions"`
}

func (p *Payload) MarshalJSON() ([]byte, error) {
	logsBloom, err := p.LogsBloom.MarshalText()
	if err != nil {
		return nil, err
	}

	transactions := make([]string, len(p.Transactions))
	for i, transaction := range p.Transactions {
		transactions[i] = hex.EncodeToHex(transaction)
	}

	return json.Marshal(&struct {
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
	}{
		ParentHash:    p.ParentHash.String(),
		FeeRecipient:  p.FeeRecipient.String(),
		StateRoot:     p.StateRoot.String(),
		ReceiptsRoot:  p.ReceiptsRoot.String(),
		LogsBloom:     string(logsBloom),
		Random:        p.Random.String(),
		Number:        hex.EncodeUint64(p.Number),
		GasLimit:      hex.EncodeUint64(p.GasLimit),
		GasUsed:       hex.EncodeUint64(p.GasUsed),
		Timestamp:     hex.EncodeUint64(p.Timestamp),
		ExtraData:     hex.EncodeToHex(p.ExtraData),
		BaseFeePerGas: hex.EncodeBig(p.BaseFeePerGas),
		BlockHash:     p.BlockHash.String(),
		Transactions:  transactions,
	})
}

func (p *Payload) UnmarshalJSON(data []byte) error {
	var rawPayload _RawPayload
	err := json.Unmarshal(data, &rawPayload)

	p.BaseFeePerGas = hex.DecodeHexToBig(string(hex.DropHexPrefix([]byte(rawPayload.BaseFeePerGas))))
	p.BlockHash = rawPayload.BlockHash
	p.ExtraData, err = hex.DecodeString(string(hex.DropHexPrefix([]byte(rawPayload.ExtraData))))

	if err != nil {
		return fmt.Errorf("failed to unmarshal payload.extraData")
	}
	p.FeeRecipient = rawPayload.FeeRecipient
	p.GasLimit, err = hex.DecodeUint64(rawPayload.GasLimit)
	if err != nil {
		return fmt.Errorf("failed to decode payload.GasLimit")
	}
	p.GasUsed, err = hex.DecodeUint64(rawPayload.GasUsed)
	if err != nil {
		return fmt.Errorf("failed to decode payload.GasUsed")
	}

	// Logs bloom decoding
	var logsBloom Bloom
	input := hex.DropHexPrefix([]byte(rawPayload.LogsBloom))
	if _, err := encodingHex.Decode(logsBloom[:], input); err != nil {
		return fmt.Errorf("failed to decode payload.logsBloom")
	}
	p.LogsBloom = logsBloom

	p.Number, err = hex.DecodeUint64(rawPayload.BlockNumber)
	if err != nil {
		return fmt.Errorf("failed to decode payload.BlockNumber")
	}
	p.ParentHash = rawPayload.ParentHash
	p.ReceiptsRoot = rawPayload.ReceiptsRoot
	p.StateRoot = rawPayload.StateRoot
	p.Timestamp, err = hex.DecodeUint64(rawPayload.Timestamp)
	if err != nil {
		return fmt.Errorf("failed to decode payload.Timestamp")
	}

	// Transaction decoding
	p.Transactions = [][]byte{}

	for _, transaction := range rawPayload.Transactions {
		decoded, err := hex.DecodeHex(transaction)
		if err != nil {
			return fmt.Errorf("failed to decode payload.Transactions")
		}
		p.Transactions = append(p.Transactions, decoded)
	}

	return err
}

func (b *Block) Hash() Hash {
	return b.Header.Hash
}

func (b *Block) Number() uint64 {
	return b.Header.Number
}

func (b *Block) ParentHash() Hash {
	return b.Header.ParentHash
}

func (b *Block) Body() *Body {
	return &Body{
		Transactions:     b.Transactions,
		Uncles:           b.Uncles,
		ExecutionPayload: b.ExecutionPayload,
	}
}

func (b *Block) Size() uint64 {
	sizePtr := b.size.Load()
	if sizePtr == nil {
		bytes := b.MarshalRLP()
		size := uint64(len(bytes))
		b.size.Store(&size)

		return size
	}

	sizeVal, ok := sizePtr.(*uint64)
	if !ok {
		return 0
	}

	return *sizeVal
}

func (b *Block) String() string {
	str := fmt.Sprintf(`Block(#%v):`, b.Number())

	return str
}

// WithSeal returns a new block with the data from b but the header replaced with
// the sealed one.
func (b *Block) WithSeal(header *Header) *Block {
	cpy := *header

	return &Block{
		Header:       &cpy,
		Transactions: b.Transactions,
		Uncles:       b.Uncles,
	}
}
