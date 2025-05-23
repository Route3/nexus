package types

import (
	"fmt"
	"math/big"

	"github.com/umbracle/fastrlp"
)

type RLPUnmarshaler interface {
	UnmarshalRLP(input []byte) error
}

type unmarshalRLPFunc func(p *fastrlp.Parser, v *fastrlp.Value) error

func UnmarshalRlp(obj unmarshalRLPFunc, input []byte) error {
	pr := fastrlp.DefaultParserPool.Get()

	v, err := pr.Parse(input)
	if err != nil {
		fastrlp.DefaultParserPool.Put(pr)

		return err
	}

	if err := obj(pr, v); err != nil {
		fastrlp.DefaultParserPool.Put(pr)

		return err
	}

	fastrlp.DefaultParserPool.Put(pr)

	return nil
}

func (b *Block) UnmarshalRLP(input []byte) error {
	return UnmarshalRlp(b.UnmarshalRLPFrom, input)
}

func (b *Block) UnmarshalRLPFrom(p *fastrlp.Parser, v *fastrlp.Value) error {
	elems, err := v.GetElems()
	if err != nil {
		return err
	}

	if len(elems) < 4 {
		return fmt.Errorf("incorrect number of elements to decode block, expected 3 but found %d", len(elems))
	}

	// header
	b.Header = &Header{}
	if err := b.Header.UnmarshalRLPFrom(p, elems[0]); err != nil {
		return err
	}

	// payload
	b.ExecutionPayload = &Payload{}
	if err := b.ExecutionPayload.UnmarshalRLPFrom(p, elems[3]); err != nil {
		return err
	}

	return nil
}

func (h *Header) UnmarshalRLP(input []byte) error {
	return UnmarshalRlp(h.UnmarshalRLPFrom, input)
}

func (h *Header) UnmarshalRLPFrom(_ *fastrlp.Parser, v *fastrlp.Value) error {
	elems, err := v.GetElems()
	if err != nil {
		return err
	}

	if len(elems) < 15 {
		return fmt.Errorf("incorrect number of elements to decode header, expected 15 but found %d", len(elems))
	}

	// parentHash
	if err = elems[0].GetHash(h.ParentHash[:]); err != nil {
		return err
	}
	// sha3uncles
	if err = elems[1].GetHash(h.Sha3Uncles[:]); err != nil {
		return err
	}
	// miner
	if h.Miner, err = elems[2].GetBytes(h.Miner[:]); err != nil {
		return err
	}
	// stateroot
	if err = elems[3].GetHash(h.StateRoot[:]); err != nil {
		return err
	}
	// txroot
	if err = elems[4].GetHash(h.TxRoot[:]); err != nil {
		return err
	}
	// receiptroot
	if err = elems[5].GetHash(h.ReceiptsRoot[:]); err != nil {
		return err
	}
	// logsBloom
	if _, err = elems[6].GetBytes(h.LogsBloom[:0], 256); err != nil {
		return err
	}
	// difficulty
	if h.Difficulty, err = elems[7].GetUint64(); err != nil {
		return err
	}
	// number
	if h.Number, err = elems[8].GetUint64(); err != nil {
		return err
	}
	// gasLimit
	if h.GasLimit, err = elems[9].GetUint64(); err != nil {
		return err
	}
	// gasused
	if h.GasUsed, err = elems[10].GetUint64(); err != nil {
		return err
	}
	// timestamp
	if h.Timestamp, err = elems[11].GetUint64(); err != nil {
		return err
	}
	// extraData
	if h.ExtraData, err = elems[12].GetBytes(h.ExtraData[:0]); err != nil {
		return err
	}
	// mixHash
	if err = elems[13].GetHash(h.MixHash[:0]); err != nil {
		return err
	}
	// nonce
	nonce, err := elems[14].GetUint64()
	if err != nil {
		return err
	}
	h.SetNonce(nonce)

	// payload hash
	if err = elems[15].GetHash(h.PayloadHash[:0]); err != nil {
		return err
	}

	// compute the hash after the decoding
	h.ComputeHash()

	return err
}

func (p *Payload) UnmarshalRLPFrom(_ *fastrlp.Parser, v *fastrlp.Value) error {
	elems, err := v.GetElems()
	if err != nil {
		return err
	}

	if len(elems) != 13 {
		return fmt.Errorf("incorrect number of elements to decode payload, expected 13 but found %d", len(elems))
	}

	if err = elems[0].GetHash(p.ParentHash[:]); err != nil {
		return err
	}

	if err = elems[1].GetAddr(p.FeeRecipient[:]); err != nil {
		return err
	}

	if err = elems[2].GetHash(p.StateRoot[:]); err != nil {
		return err
	}

	if err = elems[3].GetHash(p.ReceiptsRoot[:]); err != nil {
		return err
	}

	// perhaps do it like in header?: if _, err = elems[6].GetBytes(h.LogsBloom[:0], 256); err != nil
	if _, err = elems[4].GetBytes(p.LogsBloom[:0], 256); err != nil {
		return err
	}

	if p.Number, err = elems[5].GetUint64(); err != nil {
		return err
	}

	if p.GasLimit, err = elems[6].GetUint64(); err != nil {
		return err
	}

	if p.GasUsed, err = elems[7].GetUint64(); err != nil {
		return err
	}

	if p.Timestamp, err = elems[8].GetUint64(); err != nil {
		return err
	}

	if p.ExtraData, err = elems[9].GetBytes(p.ExtraData[:0]); err != nil {
		return err
	}

	p.BaseFeePerGas = new(big.Int)
	if err := elems[10].GetBigInt(p.BaseFeePerGas); err != nil {
		return err
	}

	if err = elems[11].GetHash(p.BlockHash[:]); err != nil {
		return err
	}

	// transactions
	transactions, err := elems[12].GetElems()
	if err != nil {
		return err
	}

	p.Transactions = make([][]byte, len(transactions))

	for i, transaction := range transactions {

		p.Transactions[i], err = transaction.GetBytes(p.Transactions[i])
		if err != nil {
			return err
		}
	}

	return err
}
