package staking

import (
	"errors"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/maticnetwork/heimdall/types"
	hmTypes "github.com/maticnetwork/heimdall/types"
)

// GenesisValidator genesis validator
type GenesisValidator struct {
	ID         hmTypes.ValidatorID   `json:"id"`
	StartEpoch uint64                `json:"start_epoch"`
	EndEpoch   uint64                `json:"end_epoch"`
	Power      uint64                `json:"power"` // aka Amount
	PubKey     hmTypes.PubKey        `json:"pub_key"`
	Signer     types.HeimdallAddress `json:"signer"`
}

// HeimdallValidator converts genesis validator validator to Heimdall validator
func (v *GenesisValidator) HeimdallValidator() hmTypes.Validator {
	return hmTypes.Validator{
		ID:          v.ID,
		PubKey:      v.PubKey,
		VotingPower: int64(v.Power),
		StartEpoch:  v.StartEpoch,
		EndEpoch:    v.EndEpoch,
		Signer:      v.Signer,
	}
}

// GenesisState is the checkpoint state that must be provided at genesis.
type GenesisState struct {
	Validators           []*hmTypes.Validator       `json:"validators" yaml:"validators"`
	CurrentValSet        hmTypes.ValidatorSet       `json:"current_val_set" yaml:"current_val_set"`
	ValidatorAccounts    []hmTypes.ValidatorAccount `json:"val_accounts" yaml:"val_accounts"`
	ProposerBonusPercent int64                      `json:"proposer_bonus_percent" yaml:"proposer_bonus_percent"`
}

// NewGenesisState creates a new genesis state.
func NewGenesisState(
	validators []*hmTypes.Validator,
	currentValSet hmTypes.ValidatorSet,
	validatorAccounts []hmTypes.ValidatorAccount,
	proposerBonusPercent int64,

) GenesisState {
	return GenesisState{
		Validators:           validators,
		CurrentValSet:        currentValSet,
		ValidatorAccounts:    validatorAccounts,
		ProposerBonusPercent: proposerBonusPercent,
	}
}

// DefaultGenesisState returns a default genesis state
func DefaultGenesisState(validators []*hmTypes.Validator, currentValSet hmTypes.ValidatorSet) GenesisState {
	var validatorAccounts []hmTypes.ValidatorAccount
	for _, val := range validators {
		valAccount := hmTypes.ValidatorAccount{
			ID:            val.ID,
			RewardAmount:  big.NewInt(0).String(),
			SlashedAmount: big.NewInt(0).String(),
		}
		validatorAccounts = append(validatorAccounts, valAccount)
	}
	return NewGenesisState(validators, currentValSet, validatorAccounts, DefaultProposerBonusPercent)
}

// InitGenesis sets distribution information for genesis.
func InitGenesis(ctx sdk.Context, keeper Keeper, data GenesisState) {
	// get current val set
	var vals []*hmTypes.Validator
	if len(data.CurrentValSet.Validators) == 0 {
		vals = data.Validators
	} else {
		vals = data.CurrentValSet.Validators
	}

	// result
	resultValSet := hmTypes.NewValidatorSet(vals)
	// add validators in store
	for _, validator := range resultValSet.Validators {
		// Add individual validator to state
		keeper.AddValidator(ctx, *validator)

	}

	// update validator set in store
	if err := keeper.UpdateValidatorSetInStore(ctx, *resultValSet); err != nil {
		panic(err)
	}

	// Add genesis validator accounts
	for _, validator := range resultValSet.Validators {
		// check if validator exists in data.ValidatorAccounts
		isExist := false
		// Add validator account from genesis
		for _, valAccount := range data.ValidatorAccounts {
			if valAccount.ID == validator.ID {
				isExist = true
				if err := keeper.AddValidatorAccount(ctx, valAccount); err != nil {
					panic((err))
				}
			}
		}
		// Create Validator Account if not set in genesis
		if !isExist {
			validatorAccount := types.ValidatorAccount{
				ID:            validator.ID,
				RewardAmount:  big.NewInt(0).String(),
				SlashedAmount: big.NewInt(0).String(),
			}
			if err := keeper.AddValidatorAccount(ctx, validatorAccount); err != nil {
				panic((err))
			}
		}
	}

	keeper.SetProposerBonusPercent(ctx, data.ProposerBonusPercent)
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, keeper Keeper) GenesisState {
	// return new genesis state
	return NewGenesisState(
		keeper.GetAllValidators(ctx),
		keeper.GetValidatorSet(ctx),
		keeper.GetAllValidatorAccounts(ctx),
		keeper.GetProposerBonusPercent(ctx),
	)
}

// ValidateGenesis performs basic validation of bor genesis data returning an
// error for any failed validation criteria.
func ValidateGenesis(data GenesisState) error {
	for _, validator := range data.Validators {
		if !validator.ValidateBasic() {
			return errors.New("Invalid validator")
		}
	}

	return nil
}
