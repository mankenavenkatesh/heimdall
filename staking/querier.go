package staking

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/maticnetwork/heimdall/auth/types"
	stakingTypes "github.com/maticnetwork/heimdall/staking/types"
)

// query endpoints supported by the staking Querier
const (
	QuerySlashValidator = "slash-validator"
)

// NewQuerier returns querier for staking Rest endpoints
func NewQuerier(keeper Keeper) sdk.Querier {

	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, sdk.Error) {
		switch path[0] {
		case stakingTypes.QueryValidatorStatus:
			return handlerQueryValidatorStatus(ctx, req, keeper)
		case stakingTypes.QueryProposerBonusPercent:
			return handlerQueryProposerBonusPercent(ctx, req, keeper)
		case QuerySlashValidator:
			return querySlashValidator(ctx, req, keeper)
		default:
			return nil, sdk.ErrUnknownRequest("unknown auth query endpoint")
		}
	}
}

func handlerQueryValidatorStatus(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	var params stakingTypes.QueryValidatorStatusParams
	if err := keeper.cdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}

	// get validator status by signer address
	status := keeper.IsCurrentValidatorByAddress(ctx, params.SignerAddress)

	// json record
	bz, err := codec.MarshalJSONIndent(keeper.cdc, status)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}
	return bz, nil
}

func handlerQueryProposerBonusPercent(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	// GetProposerBonusPercent
	proposerBonusPercent := keeper.GetProposerBonusPercent(ctx)

	// json record
	bz, err := codec.MarshalJSONIndent(keeper.cdc, proposerBonusPercent)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}
	return bz, nil
}

func querySlashValidator(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	var params stakingTypes.ValidatorSlashParams

	if err := types.ModuleCdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}

	err := keeper.SlashValidator(ctx, params.ValID, params.SlashAmount)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not Slash validator", err.Error()))
	}
	return nil, nil
}
