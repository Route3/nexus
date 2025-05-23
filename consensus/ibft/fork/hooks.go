package fork

import (
	"errors"

	"github.com/apex-fusion/nexus/consensus/ibft/hook"
	stakingHelper "github.com/apex-fusion/nexus/helper/staking"
	"github.com/apex-fusion/nexus/types"
	"github.com/apex-fusion/nexus/validators"
	"github.com/apex-fusion/nexus/validators/store"
)

var (
	ErrTxInLastEpochOfBlock = errors.New("block must not have transactions in the last of epoch")
)

// HeaderModifier is an interface for the struct that modifies block header for additional process
type HeaderModifier interface {
	ModifyHeader(*types.Header, types.Address) error
	VerifyHeader(*types.Header) error
	ProcessHeader(*types.Header) error
}

// registerHeaderModifierHooks registers hooks to modify header by validator store
func registerHeaderModifierHooks(
	hooks *hook.Hooks,
	validatorStore store.ValidatorStore,
) {
	if modifier, ok := validatorStore.(HeaderModifier); ok {
		hooks.ModifyHeaderFunc = modifier.ModifyHeader
		hooks.VerifyHeaderFunc = modifier.VerifyHeader
		hooks.ProcessHeaderFunc = modifier.ProcessHeader
	}
}

// Updatable is an interface for the struct that updates validators in the middle
type Updatable interface {
	// UpdateValidatorSet updates validators forcibly
	// in order that new validators are available from the given height
	UpdateValidatorSet(validators.Validators, uint64) error
}

// registerUpdateValidatorsHooks registers hooks to update validators in the middle
func registerUpdateValidatorsHooks(
	hooks *hook.Hooks,
	validatorStore store.ValidatorStore,
	validators validators.Validators,
	fromHeight uint64,
) {
	if us, ok := validatorStore.(Updatable); ok {
		hooks.PostInsertBlockFunc = func(b *types.Block) error {
			if fromHeight != b.Number()+1 {
				return nil
			}

			// update validators if the block height is the one before beginning height
			return us.UpdateValidatorSet(validators, fromHeight)
		}
	}
}

// registerPoSVerificationHooks registers that hooks to prevent the last epoch block from having transactions
func registerTxInclusionGuardHooks(hooks *hook.Hooks, epochSize uint64) {
	isLastEpoch := func(height uint64) bool {
		return height > 0 && height%epochSize == 0
	}

	hooks.ShouldWriteTransactionFunc = func(height uint64) bool {
		return !isLastEpoch(height)
	}

	hooks.VerifyBlockFunc = func(block *types.Block) error {
		if isLastEpoch(block.Number()) {
			return ErrTxInLastEpochOfBlock
		}

		return nil
	}
}

// registerStakingContractDeploymentHooks registers hooks
// to deploy or update staking contract
func registerStakingContractDeploymentHooks(
	hooks *hook.Hooks,
	fork *IBFTFork,
) {
}

// getPreDeployParams returns PredeployParams for Staking Contract from IBFTFork
func getPreDeployParams(fork *IBFTFork) stakingHelper.PredeployParams {
	params := stakingHelper.PredeployParams{
		MinValidatorCount: stakingHelper.MinValidatorCount,
		MaxValidatorCount: stakingHelper.MaxValidatorCount,
	}

	if fork.MinValidatorCount != nil {
		params.MinValidatorCount = fork.MinValidatorCount.Value
	}

	if fork.MaxValidatorCount != nil {
		params.MaxValidatorCount = fork.MaxValidatorCount.Value
	}

	return params
}
