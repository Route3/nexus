package types

import (
	"github.com/umbracle/fastrlp"
)

type RLPMarshaler interface {
	MarshalRLPTo(dst []byte) []byte
}

type marshalRLPFunc func(ar *fastrlp.Arena) *fastrlp.Value

func MarshalRLPTo(obj marshalRLPFunc, dst []byte) []byte {
	ar := fastrlp.DefaultArenaPool.Get()
	dst = obj(ar).MarshalTo(dst)
	fastrlp.DefaultArenaPool.Put(ar)

	return dst
}

func (b *Block) MarshalRLP() []byte {
	return b.MarshalRLPTo(nil)
}

func (b *Block) MarshalRLPTo(dst []byte) []byte {
	return MarshalRLPTo(b.MarshalRLPWith, dst)
}

func (b *Block) MarshalRLPWith(ar *fastrlp.Arena) *fastrlp.Value {
	vv := ar.NewArray()
	vv.Set(b.Header.MarshalRLPWith(ar))

	vv.Set(ar.NewNullArray()) // Backwards compatibility for Transactions in block
	vv.Set(ar.NewNullArray()) // Backwards compatibility for Uncles in block

	if b.ExecutionPayload != nil {
		vv.Set(b.ExecutionPayload.MarshalRLPWith(ar))
	}
	return vv
}

func (h *Header) MarshalRLP() []byte {
	return h.MarshalRLPTo(nil)
}

func (h *Header) MarshalRLPTo(dst []byte) []byte {
	return MarshalRLPTo(h.MarshalRLPWith, dst)
}

// MarshalRLPWith marshals the header to RLP with a specific fastrlp.Arena
func (h *Header) MarshalRLPWith(arena *fastrlp.Arena) *fastrlp.Value {
	vv := arena.NewArray()

	vv.Set(arena.NewBytes(h.ParentHash.Bytes()))
	vv.Set(arena.NewBytes(h.Sha3Uncles.Bytes()))
	vv.Set(arena.NewCopyBytes(h.Miner[:]))
	vv.Set(arena.NewBytes(h.StateRoot.Bytes()))
	vv.Set(arena.NewBytes(h.TxRoot.Bytes()))
	vv.Set(arena.NewBytes(h.ReceiptsRoot.Bytes()))
	vv.Set(arena.NewCopyBytes(h.LogsBloom[:]))

	vv.Set(arena.NewUint(h.Difficulty))
	vv.Set(arena.NewUint(h.Number))
	vv.Set(arena.NewUint(h.GasLimit))
	vv.Set(arena.NewUint(h.GasUsed))
	vv.Set(arena.NewUint(h.Timestamp))

	vv.Set(arena.NewCopyBytes(h.ExtraData))
	vv.Set(arena.NewBytes(h.MixHash.Bytes()))
	vv.Set(arena.NewCopyBytes(h.Nonce[:]))

	vv.Set(arena.NewBytes(h.PayloadHash.Bytes()))

	return vv
}

func (p *Payload) MarshalRLPWith(arena *fastrlp.Arena) *fastrlp.Value {
	// We only encode the values that we are using.
	// Missing fields in Payload struct (random, withdrawalsRoot) are for engine API compatibility only

	vv := arena.NewArray()
	vv.Set(arena.NewBytes(p.ParentHash.Bytes()))
	vv.Set(arena.NewBytes(p.FeeRecipient.Bytes()))
	vv.Set(arena.NewBytes(p.StateRoot.Bytes()))
	vv.Set(arena.NewBytes(p.ReceiptsRoot.Bytes()))
	vv.Set(arena.NewCopyBytes(p.LogsBloom[:]))
	vv.Set(arena.NewUint(p.Number))
	vv.Set(arena.NewUint(p.GasLimit))
	vv.Set(arena.NewUint(p.GasUsed))
	vv.Set(arena.NewUint(p.Timestamp))
	vv.Set(arena.NewBytes(p.ExtraData))
	vv.Set(arena.NewBigInt(p.BaseFeePerGas))
	vv.Set(arena.NewBytes(p.BlockHash.Bytes()))

	if len(p.Transactions) == 0 {
		vv.Set(arena.NewNullArray())
	} else {
		v0 := arena.NewArray()
		for _, tx := range p.Transactions {
			v0.Set(arena.NewBytes(tx))
		}
		vv.Set(v0)
	}

	return vv
}
