package consensus

import (
	"github.com/apex-fusion/nexus/types"
)

// BuildBlockParams are parameters passed into the BuildBlock helper method
type BuildBlockParams struct {
	Header   *types.Header
	Receipts []*types.Receipt
	Payload  *types.Payload
}

// BuildBlock is a utility function that builds a block, based on the passed in header, transactions and receipts
func BuildBlock(params BuildBlockParams) *types.Block {
	header := params.Header
	payload := params.Payload

	header.TxRoot = types.EmptyRootHash

	header.ReceiptsRoot = types.EmptyRootHash
	
	header.Sha3Uncles = types.EmptyUncleHash
	header.ComputeHash()

	return &types.Block{
		Header:           header,
		ExecutionPayload: payload,
	}
}
