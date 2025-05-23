package dev

import (
	"fmt"
	"time"

	"github.com/apex-fusion/nexus/blockchain"
	"github.com/apex-fusion/nexus/consensus"
	"github.com/apex-fusion/nexus/helper/progress"
	"github.com/apex-fusion/nexus/types"
	"github.com/hashicorp/go-hclog"
)

const (
	devConsensus = "dev-consensus"
)

// Dev consensus protocol seals any new transaction immediately
type Dev struct {
	logger hclog.Logger

	notifyCh chan struct{}
	closeCh  chan struct{}

	interval uint64

	blockchain *blockchain.Blockchain
}

// Factory implements the base factory method
func Factory(
	params *consensus.Params,
) (consensus.Consensus, error) {
	logger := params.Logger.Named("dev")

	d := &Dev{
		logger:     logger,
		notifyCh:   make(chan struct{}),
		closeCh:    make(chan struct{}),
		blockchain: params.Blockchain,
	}

	rawInterval, ok := params.Config.Config["interval"]
	if ok {
		interval, ok := rawInterval.(uint64)
		if !ok {
			return nil, fmt.Errorf("interval expected int")
		}

		d.interval = interval
	}

	return d, nil
}

// Initialize initializes the consensus
func (d *Dev) Initialize() error {

	return nil
}

// Start starts the consensus mechanism
func (d *Dev) Start() error {
	go d.run()

	return nil
}

func (d *Dev) nextNotify() chan struct{} {
	if d.interval == 0 {
		d.interval = 1
	}

	go func() {
		<-time.After(time.Duration(d.interval) * time.Second)
		d.notifyCh <- struct{}{}
	}()

	return d.notifyCh
}

func (d *Dev) run() {
	d.logger.Info("consensus started")

	for {
		// wait until there is a new txn
		select {
		case <-d.nextNotify():
		case <-d.closeCh:
			return
		}

		// There are new transactions in the pool, try to seal them
		header := d.blockchain.Header()
		if err := d.writeNewBlock(header); err != nil {
			d.logger.Error("failed to mine block", "err", err)
		}
	}
}

// writeNewBLock generates a new block based on transactions from the pool,
// and writes them to the blockchain
func (d *Dev) writeNewBlock(parent *types.Header) error {
	// Generate the base block
	num := parent.Number
	header := &types.Header{
		ParentHash: parent.Hash,
		Number:     num + 1,
		GasLimit:   parent.GasLimit, // Inherit from parent for now, will need to adjust dynamically later.
		Timestamp:  uint64(time.Now().Unix()),
	}

	// calculate gas limit based on parent header
	gasLimit, err := d.blockchain.CalculateGasLimit(header.Number)
	if err != nil {
		return err
	}

	header.GasLimit = gasLimit

	return nil
}

// REQUIRED BASE INTERFACE METHODS //

func (d *Dev) VerifyHeader(header *types.Header) error {
	// All blocks are valid
	return nil
}

func (d *Dev) ProcessHeaders(headers []*types.Header) error {
	return nil
}

func (d *Dev) GetBlockCreator(header *types.Header) (types.Address, error) {
	return types.BytesToAddress(header.Miner), nil
}

// PreCommitState a hook to be called before finalizing state transition on inserting block
func (d *Dev) PreCommitState(_header *types.Header) error {
	return nil
}

func (d *Dev) GetSyncProgression() *progress.Progression {
	return nil
}

func (d *Dev) Close() error {
	close(d.closeCh)

	return nil
}
