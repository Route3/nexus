package crypto

import (
	"fmt"
	"math/big"

	"github.com/umbracle/fastrlp"
)

type FrontierSigner struct {
}

var signerPool fastrlp.ArenaPool

// Magic numbers from Ethereum, used in v calculation
var (
	big27 = big.NewInt(27)
	big35 = big.NewInt(35)
)

// calculateV returns the V value for transactions pre EIP155
func (f *FrontierSigner) CalculateV(parity byte) []byte {
	reference := big.NewInt(int64(parity))
	reference.Add(reference, big27)

	return reference.Bytes()
}

// NewEIP155Signer returns a new EIP155Signer object
func NewEIP155Signer(chainID uint64) *EIP155Signer {
	return &EIP155Signer{chainID: chainID}
}

type EIP155Signer struct {
	chainID uint64
}

// calculateV returns the V value for transaction signatures. Based on EIP155
func (e *EIP155Signer) CalculateV(parity byte) []byte {
	reference := big.NewInt(int64(parity))
	reference.Add(reference, big35)

	mulOperand := big.NewInt(0).Mul(big.NewInt(int64(e.chainID)), big.NewInt(2))

	reference.Add(reference, mulOperand)

	return reference.Bytes()
}

// encodeSignature generates a signature value based on the R, S and V value
func encodeSignature(R, S *big.Int, V byte) ([]byte, error) {
	if !ValidateSignatureValues(V, R, S) {
		return nil, fmt.Errorf("invalid txn signature")
	}

	sig := make([]byte, 65)
	copy(sig[32-len(R.Bytes()):32], R.Bytes())
	copy(sig[64-len(S.Bytes()):64], S.Bytes())
	sig[64] = V

	return sig, nil
}
