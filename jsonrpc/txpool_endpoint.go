package jsonrpc

import (
	"github.com/apex-fusion/nexus/types"
)

type ContentResponse struct {
	Pending map[types.Address]map[uint64]*txpoolTransaction `json:"pending"`
	Queued  map[types.Address]map[uint64]*txpoolTransaction `json:"queued"`
}

type InspectResponse struct {
	Pending         map[string]map[string]string `json:"pending"`
	Queued          map[string]map[string]string `json:"queued"`
	CurrentCapacity uint64                       `json:"currentCapacity"`
	MaxCapacity     uint64                       `json:"maxCapacity"`
}

type StatusResponse struct {
	Pending uint64 `json:"pending"`
	Queued  uint64 `json:"queued"`
}

type txpoolTransaction struct {
	Nonce       argUint64      `json:"nonce"`
	GasPrice    argBig         `json:"gasPrice"`
	Gas         argUint64      `json:"gas"`
	To          *types.Address `json:"to"`
	Value       argBig         `json:"value"`
	Input       argBytes       `json:"input"`
	Hash        types.Hash     `json:"hash"`
	From        types.Address  `json:"from"`
	BlockHash   types.Hash     `json:"blockHash"`
	BlockNumber interface{}    `json:"blockNumber"`
	TxIndex     interface{}    `json:"transactionIndex"`
}
