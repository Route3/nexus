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

	if len(b.Uncles) == 0 {
		vv.Set(ar.NewNullArray())
	} else {
		v1 := ar.NewArray()
		for _, uncle := range b.Uncles {
			v1.Set(uncle.MarshalRLPWith(ar))
		}
		vv.Set(v1)
	}

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

func (r Receipts) MarshalRLPTo(dst []byte) []byte {
	return MarshalRLPTo(r.MarshalRLPWith, dst)
}

func (r *Receipts) MarshalRLPWith(a *fastrlp.Arena) *fastrlp.Value {
	vv := a.NewArray()
	for _, rr := range *r {
		vv.Set(rr.MarshalRLPWith(a))
	}

	return vv
}

func (r *Receipt) MarshalRLP() []byte {
	return r.MarshalRLPTo(nil)
}

func (r *Receipt) MarshalRLPTo(dst []byte) []byte {
	return MarshalRLPTo(r.MarshalRLPWith, dst)
}

// MarshalRLPWith marshals a receipt with a specific fastrlp.Arena
func (r *Receipt) MarshalRLPWith(a *fastrlp.Arena) *fastrlp.Value {
	vv := a.NewArray()

	if r.Status != nil {
		vv.Set(a.NewUint(uint64(*r.Status)))
	} else {
		vv.Set(a.NewBytes(r.Root[:]))
	}

	vv.Set(a.NewUint(r.CumulativeGasUsed))
	vv.Set(a.NewCopyBytes(r.LogsBloom[:]))
	vv.Set(r.MarshalLogsWith(a))

	return vv
}

// MarshalLogsWith marshals the logs of the receipt to RLP with a specific fastrlp.Arena
func (r *Receipt) MarshalLogsWith(a *fastrlp.Arena) *fastrlp.Value {
	if len(r.Logs) == 0 {
		// There are no receipts, write the RLP null array entry
		return a.NewNullArray()
	}

	logs := a.NewArray()

	for _, l := range r.Logs {
		logs.Set(l.MarshalRLPWith(a))
	}

	return logs
}

func (l *Log) MarshalRLPWith(a *fastrlp.Arena) *fastrlp.Value {
	v := a.NewArray()
	v.Set(a.NewBytes(l.Address.Bytes()))

	topics := a.NewArray()
	for _, t := range l.Topics {
		topics.Set(a.NewBytes(t.Bytes()))
	}

	v.Set(topics)
	v.Set(a.NewBytes(l.Data))

	return v
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
