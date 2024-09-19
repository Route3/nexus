package blockchain

import "fmt"

// func (b *Blockchain) setCurrentHeader(h *types.Header, diff *big.Int) {

// SetPayloadId sets the value of payloadId, using a write lock
func (b *Blockchain) SetPayloadId(payloadId string) {
	b.payloadIdMutex.Lock()
	defer b.payloadIdMutex.Unlock()

	b.payloadId = payloadId
}

// GetPayloadId gets the value of payloadId, using a read lock
func (b *Blockchain) GetPayloadId() string {
	b.payloadIdMutex.RLock()

	defer b.payloadIdMutex.RUnlock()
	return b.payloadId
}

// GetLatestPayloadHash returns the hash of the latest payload in the blockchain database
func (b *Blockchain) GetLatestPayloadHash() string {
	if b.Header().Number == 0 {
		return b.executionGenesisHash
	}

	return b.Header().PayloadHash.String() // TODO: Typify PayloadHash to use Hash instead of String everywhere
}
