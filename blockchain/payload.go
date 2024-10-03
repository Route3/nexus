package blockchain

import "github.com/apex-fusion/nexus/types"

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
func (b *Blockchain) GetLatestPayloadHash() types.Hash {
	if b.Header().Number == 0 {
		return types.StringToHash(b.executionGenesisHash)
	}

	return b.Header().PayloadHash
}
