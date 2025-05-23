package types

import (
	"fmt"

	"github.com/umbracle/fastrlp"
)

type RLPStoreUnmarshaler interface {
	UnmarshalStoreRLP(input []byte) error
}

func (b *Body) UnmarshalRLP(input []byte) error {
	return UnmarshalRlp(b.UnmarshalRLPFrom, input)
}

func (b *Body) UnmarshalRLPFrom(p *fastrlp.Parser, v *fastrlp.Value) error {
	tuple, err := v.GetElems()
	if err != nil {
		return err
	}

	if len(tuple) < 3 {
		return fmt.Errorf("incorrect number of elements to decode header, expected 3 but found %d", len(tuple))
	}

	// execution payload
	b.ExecutionPayload = &Payload{}
	if err := b.ExecutionPayload.UnmarshalRLPFrom(p, tuple[2]); err != nil {
		return err
	}

	return nil
}
