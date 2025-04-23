package types

import (
	"github.com/umbracle/fastrlp"
)

type RLPStoreMarshaler interface {
	MarshalStoreRLPTo(dst []byte) []byte
}

func (b *Body) MarshalRLPTo(dst []byte) []byte {
	return MarshalRLPTo(b.MarshalRLPWith, dst)
}

func (b *Body) MarshalRLPWith(ar *fastrlp.Arena) *fastrlp.Value {
	vv := ar.NewArray()

	vv.Set(ar.NewNullArray()) // Backwards compatibility for Transactions in Body
	vv.Set(ar.NewNullArray()) // Backwards compatibility for Uncles in Body

	if b.ExecutionPayload != nil {
		vv.Set(b.ExecutionPayload.MarshalRLPWith(ar))
	}

	return vv
}
