package dummy

import (
	"github.com/apex-fusion/nexus/blockchain"
	"github.com/apex-fusion/nexus/consensus"
	"github.com/apex-fusion/nexus/helper/progress"
	"github.com/apex-fusion/nexus/types"
	"github.com/hashicorp/go-hclog"
)

type Dummy struct {
	logger     hclog.Logger
	notifyCh   chan struct{}
	closeCh    chan struct{}
	blockchain *blockchain.Blockchain
}

func Factory(params *consensus.Params) (consensus.Consensus, error) {
	return nil, nil
}

// Initialize initializes the consensus
func (d *Dummy) Initialize() error {

	return nil
}

func (d *Dummy) Start() error {
	go d.run()

	return nil
}

func (d *Dummy) VerifyHeader(header *types.Header) error {
	// All blocks are valid
	return nil
}

func (d *Dummy) ProcessHeaders(headers []*types.Header) error {
	return nil
}

func (d *Dummy) GetBlockCreator(header *types.Header) (types.Address, error) {
	return types.BytesToAddress(header.Miner), nil
}

// PreCommitState a hook to be called before finalizing state transition on inserting block
func (d *Dummy) PreCommitState(_header *types.Header) error {
	return nil
}

func (d *Dummy) GetSyncProgression() *progress.Progression {
	return nil
}

func (d *Dummy) Close() error {
	close(d.closeCh)

	return nil
}

func (d *Dummy) run() {
	d.logger.Info("started")
	// do nothing
	<-d.closeCh
}
