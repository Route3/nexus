package contract

import (
	"fmt"

	"github.com/apex-fusion/nexus/contracts/staking"
	"github.com/apex-fusion/nexus/crypto"
	"github.com/apex-fusion/nexus/types"
	"github.com/apex-fusion/nexus/validators"
)

// FetchValidators fetches validators from a contract switched by validator type
func FetchValidators(
	validatorType validators.ValidatorType,
	from types.Address,
) (validators.Validators, error) {
	switch validatorType {
	case validators.ECDSAValidatorType:
		return FetchECDSAValidators(from)
	case validators.BLSValidatorType:
		return FetchBLSValidators(from)
	}

	return nil, fmt.Errorf("unsupported validator type: %s", validatorType)
}

// FetchECDSAValidators queries a contract for validator addresses and returns ECDSAValidators
func FetchECDSAValidators(
	from types.Address,
) (validators.Validators, error) {
	valAddrs, err := staking.QueryValidators(nil, from)
	if err != nil {
		return nil, err
	}

	ecdsaValidators := validators.NewECDSAValidatorSet()
	for _, addr := range valAddrs {
		if err := ecdsaValidators.Add(validators.NewECDSAValidator(addr)); err != nil {
			return nil, err
		}
	}

	return ecdsaValidators, nil
}

// FetchBLSValidators queries a contract for validator addresses & BLS Public Keys and returns ECDSAValidators
func FetchBLSValidators(
	from types.Address,
) (validators.Validators, error) {
	valAddrs, err := staking.QueryValidators(nil, from)
	if err != nil {
		return nil, err
	}

	blsPublicKeys, err := staking.QueryBLSPublicKeys(nil, from)
	if err != nil {
		return nil, err
	}

	blsValidators := validators.NewBLSValidatorSet()

	for idx := range valAddrs {
		// ignore the validator whose BLS Key is not set
		// because BLS validator needs to have both Address and BLS Public Key set
		// in the contract
		if _, err := crypto.UnmarshalBLSPublicKey(blsPublicKeys[idx]); err != nil {
			continue
		}

		if err := blsValidators.Add(validators.NewBLSValidator(
			valAddrs[idx],
			blsPublicKeys[idx],
		)); err != nil {
			return nil, err
		}
	}

	return blsValidators, nil
}
