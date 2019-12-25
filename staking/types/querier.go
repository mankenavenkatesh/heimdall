package types

import (
	"math/big"

	hmTyps "github.com/maticnetwork/heimdall/types"
)

// query endpoints supported by the staking Querier
const (
	QueryValidatorStatus      = "validator-status"
	QueryProposerBonusPercent = "proposer-bonus-percent"
)

// QueryValidatorStatusParams defines the params for querying val status.
type QueryValidatorStatusParams struct {
	SignerAddress []byte
}

// ValidatorSlashParams defines the params for slashing a validator
type ValidatorSlashParams struct {
	ValID       hmTyps.ValidatorID
	SlashAmount *big.Int
}

// NewQueryValidatorStatusParams creates a new instance of QueryValidatorStatusParams.
func NewQueryValidatorStatusParams(signerAddress []byte) QueryValidatorStatusParams {
	return QueryValidatorStatusParams{SignerAddress: signerAddress}
}

// NewValidatorSlashParams creates a new instance of ValidatorSlashParams.
func NewValidatorSlashParams(validatorID hmTyps.ValidatorID, amountToSlash *big.Int) ValidatorSlashParams {
	return ValidatorSlashParams{ValID: validatorID, SlashAmount: amountToSlash}
}
