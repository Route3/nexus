package ibft

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/Route3/go-ibft/messages"
	"github.com/apex-fusion/nexus/consensus/ibft/signer"
	"github.com/apex-fusion/nexus/helper/hex"
	"github.com/apex-fusion/nexus/types"
)

func (i *backendIBFT) BuildProposal(blockNumber uint64, context context.Context) []byte {
	var (
		latestHeader      = i.blockchain.Header()
		latestBlockNumber = latestHeader.Number
	)

	if latestBlockNumber+1 != blockNumber {
		i.logger.Error(
			"unable to build block, due to lack of parent block",
			"num",
			latestBlockNumber,
		)

		return nil
	}

	block, err := i.buildBlock(latestHeader, context)
	if err != nil {
		i.logger.Error("cannot build block", "num", blockNumber, "err", err)

		return nil
	}

	return block.MarshalRLP()
}

func (i *backendIBFT) InsertBlock(
	proposal []byte,
	committedSeals []*messages.CommittedSeal,
) {
	newBlock := &types.Block{}
	if err := newBlock.UnmarshalRLP(proposal); err != nil {
		i.logger.Error("cannot unmarshal proposal", "err", err)

		return
	}

	committedSealsMap := make(map[types.Address][]byte, len(committedSeals))

	for _, cm := range committedSeals {
		committedSealsMap[types.BytesToAddress(cm.Signer)] = cm.Signature
	}

	// Copy extra data for debugging purposes
	extraDataOriginal := newBlock.Header.ExtraData
	extraDataBackup := make([]byte, len(extraDataOriginal))
	copy(extraDataBackup, extraDataOriginal)

	// Push the committed seals to the header
	header, err := i.currentSigner.WriteCommittedSeals(newBlock.Header, committedSealsMap)
	if err != nil {
		i.logger.Error("cannot write committed seals", "err", err)

		return
	}

	// WriteCommittedSeals alters the extra data before writing the block
	// It doesn't handle errors while pushing changes which can result in
	// corrupted extra data.
	// We don't know exact circumstance of the unmarshalRLP error
	// This is a safety net to help us narrow down and also recover before
	// writing the block
	if err := i.ValidateExtraDataFormat(newBlock.Header); err != nil {
		// Format committed seals to make them more readable
		committedSealsStr := make([]string, len(committedSealsMap))
		for i, seal := range committedSeals {
			committedSealsStr[i] = fmt.Sprintf("{signer=%v signature=%v}",
				hex.EncodeToHex(seal.Signer),
				hex.EncodeToHex(seal.Signature))
		}

		i.logger.Error("cannot write block: corrupted extra data",
			"err", err,
			"before", hex.EncodeToHex(extraDataBackup),
			"after", hex.EncodeToHex(header.ExtraData),
			"committedSeals", committedSealsStr)

		return
	}

	newBlock.Header = header

	// Save the block locally
	if err := i.blockchain.WriteBlock(newBlock, "consensus"); err != nil {
		i.logger.Error("cannot write block", "err", err)

		return
	}

	i.updateMetrics(newBlock)

	i.logger.Info(
		"block committed",
		"number", newBlock.Number(),
		"hash", newBlock.Hash(),
		"validation_type", i.currentSigner.Type(),
		"validators", i.currentValidators.Len(),
		"committed", len(committedSeals),
	)

	if err := i.currentHooks.PostInsertBlock(newBlock); err != nil {
		i.logger.Error(
			"failed to call PostInsertBlock hook",
			"height", newBlock.Number(),
			"hash", newBlock.Hash(),
			"err", err,
		)

		return
	}
}

func (i *backendIBFT) ID() []byte {
	return i.currentSigner.Address().Bytes()
}

func (i *backendIBFT) MaximumFaultyNodes() uint64 {
	return uint64(CalcMaxFaultyNodes(i.currentValidators))
}

func (i *backendIBFT) Quorum(blockNumber uint64) uint64 {
	validators, err := i.forkManager.GetValidators(blockNumber)
	if err != nil {
		i.logger.Error(
			"failed to get validators when calculation quorum",
			"height", blockNumber,
			"err", err,
		)

		// return Math.MaxInt32 to prevent overflow when casting to int in go-ibft package
		return math.MaxInt32
	}

	quorumFn := i.quorumSize(blockNumber)

	return uint64(quorumFn(validators))
}

// buildBlock builds the block, based on the passed in snapshot and parent header
func (i *backendIBFT) buildBlock(parent *types.Header, ctx context.Context) (*types.Block, error) {
	header := &types.Header{
		ParentHash: parent.Hash,
		Number:     parent.Number + 1,
		Miner:      types.ZeroAddress.Bytes(),
		Nonce:      types.Nonce{},
		MixHash:    signer.IstanbulDigest,
		// this is required because blockchain needs difficulty to organize blocks and forks
		Difficulty:   parent.Number + 1,
		StateRoot:    types.EmptyRootHash, // this avoids needing state for now
		ReceiptsRoot: types.EmptyRootHash,
		Sha3Uncles:   types.EmptyUncleHash,
		GasLimit:     parent.GasLimit, // Inherit from parent for now, will need to adjust dynamically later.
		TxRoot:       types.EmptyRootHash,
	}

	// calculate gas limit based on parent header
	gasLimit, err := i.blockchain.CalculateGasLimit(header.Number)
	if err != nil {
		return nil, err
	}

	header.GasLimit = gasLimit

	// TODO: Check if we need the modify header hooks for consensus compatibility
	if err := i.currentHooks.ModifyHeader(header, i.currentSigner.Address()); err != nil {
		return nil, err
	}

	// set the timestamp
	potentialTimestamp := i.calcHeaderTimestamp(parent.Timestamp, time.Now())
	header.Timestamp = uint64(potentialTimestamp.Unix())

	parentCommittedSeals, err := i.extractParentCommittedSeals(parent)
	if err != nil {
		return nil, err
	}

	i.currentSigner.InitIBFTExtra(header, i.currentValidators, parentCommittedSeals)

	timeUntilTimestamp := time.Until(potentialTimestamp)
	if timeUntilTimestamp > 0 {
		time.Sleep(timeUntilTimestamp)
	}

	payloadResponse, err := i.blockchain.EngineClient.GetPayloadV3(i.blockchain.GetPayloadId(), ctx)
	if err != nil {
		i.logger.Error("cannot get engine's payload", "err", err)

		return nil, err
	}

	header.PayloadHash = payloadResponse.Result.ExecutionPayload.BlockHash

	var block types.Block

	if i.blockchain.Config().Forks.IsBelgrade(header.Number) {
		i.logger.Debug("Belgrade Fork detected, setting header state root from payload")
		header.StateRoot = payloadResponse.Result.ExecutionPayload.StateRoot
	} else {
		i.logger.Debug("BG not activated!")
	}

	header.GasUsed = uint64(payloadResponse.Result.ExecutionPayload.GasUsed)

	// write the seal of the block after all the fields are completed
	header, err = i.currentSigner.WriteProposerSeal(header)
	if err != nil {
		return nil, err
	}

	block.Header = header
	block.ExecutionPayload = &payloadResponse.Result.ExecutionPayload

	// compute the hash, this is only a provisional hash since the final one
	// is sealed after all the committed seals
	block.Header.ComputeHash()

	parentBeaconBlockRoot := parent.Hash.String()
	_, err = i.blockchain.EngineClient.NewPayloadV3(block.ExecutionPayload, parentBeaconBlockRoot)
	if err != nil {
		i.logger.Error("payload verification failed", "err", err)

		return nil, err
	}

	return &block, nil
}

type status uint8

const (
	success status = iota
	fail
	skip
)

// extractCommittedSeals extracts CommittedSeals from header
func (i *backendIBFT) extractCommittedSeals(
	header *types.Header,
) (signer.Seals, error) {
	signer, err := i.forkManager.GetSigner(header.Number)
	if err != nil {
		return nil, err
	}

	extra, err := signer.GetIBFTExtra(header)
	if err != nil {
		return nil, err
	}

	return extra.CommittedSeals, nil
}

// extractParentCommittedSeals extracts ParentCommittedSeals from header
func (i *backendIBFT) extractParentCommittedSeals(
	header *types.Header,
) (signer.Seals, error) {
	if header.Number == 0 {
		return nil, nil
	}

	return i.extractCommittedSeals(header)
}

// calcHeaderTimestamp calculates the new block timestamp, based
// on the block time and parent timestamp
func (i *backendIBFT) calcHeaderTimestamp(parentUnix uint64, currentTime time.Time) time.Time {
	var (
		parentTimestamp    = time.Unix(int64(parentUnix), 0)
		potentialTimestamp = parentTimestamp.Add(i.blockTime)
	)

	if potentialTimestamp.Before(currentTime) {
		// The deadline for creating this next block
		// has passed, round it to the nearest
		// multiple of block time
		// t........t+blockT...x (t+blockT.x; now).....t+blockT (potential)

		potentialTimestamp = currentTime.Add(i.blockTime / 2).Round(i.blockTime)
	}

	return potentialTimestamp
}
