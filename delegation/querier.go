package delegation

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// NewQuerier returns querier for Delegation Rest endpoints
func NewQuerier(keeper Keeper) sdk.Querier {

	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, sdk.Error) {
		switch path[0] {
		default:
			return nil, sdk.ErrUnknownRequest("unknown auth query endpoint")
		}
	}
}
