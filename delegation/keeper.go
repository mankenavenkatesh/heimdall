package delegation

import (
	"encoding/hex"
	"errors"
	"math/big"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	delegationTypes "github.com/maticnetwork/heimdall/delegation/types"
	"github.com/maticnetwork/heimdall/helper"
	"github.com/maticnetwork/heimdall/staking"
	"github.com/maticnetwork/heimdall/types"
	"github.com/tendermint/tendermint/libs/log"
)

var (
	DelegatorKey           = []byte{0x41} // prefix key to iterate through Delegators
	DelegatorAccountMapKey = []byte{0x42} // prefix for each key for Delegator Account Map
)

// Keeper stores all related data
type Keeper struct {
	cdc *codec.Codec
	// staking keeper
	sk staking.Keeper
	// The (unexposed) keys used to access the stores from the Context.
	storeKey sdk.StoreKey
	// codespace
	codespace sdk.CodespaceType
	// param space
	paramSpace params.Subspace
}

// NewKeeper create new keeper
func NewKeeper(
	cdc *codec.Codec,
	stakingKeeper staking.Keeper,
	storeKey sdk.StoreKey,
	paramSpace params.Subspace,
	codespace sdk.CodespaceType,
) Keeper {
	keeper := Keeper{
		cdc:        cdc,
		sk:         stakingKeeper,
		storeKey:   storeKey,
		paramSpace: paramSpace,
		codespace:  codespace,
	}
	return keeper
}

// Codespace returns the codespace
func (k Keeper) Codespace() sdk.CodespaceType {
	return k.codespace
}

// Logger returns a module-specific logger
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", delegationTypes.ModuleName)
}

// GetDelegatorKey - returns delegator key
func GetDelegatorKey(delegatorID types.DelegatorID) []byte {
	return append(DelegatorKey, delegatorID.Bytes()...)
}

// GetDelegatorAccountMapKey returns delegator account map
func GetDelegatorAccountMapKey(delegatorID []byte) []byte {
	return append(DelegatorAccountMapKey, delegatorID...)
}

// AddDelegator - Adds Delegator indexed with Delegator ID
func (k *Keeper) AddDelegator(ctx sdk.Context, delegator types.Delegator) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := types.MarshallDelegator(k.cdc, delegator)
	if err != nil {
		return err
	}

	store.Set(GetDelegatorKey(delegator.ID), bz)
	k.Logger(ctx).Debug("Delegator stored", "key", hex.EncodeToString(GetDelegatorKey(delegator.ID)), "delegator", delegator.String())
	return nil
}

// GetDelegatorInfo returns Delegator
func (k *Keeper) GetDelegatorInfo(ctx sdk.Context, delegatorID types.DelegatorID) (delegator types.Delegator, err error) {
	store := ctx.KVStore(k.storeKey)

	// check if delegator exists
	key := GetDelegatorKey(delegatorID)
	if !store.Has(key) {
		return delegator, errors.New("Delegator not found")
	}

	// unmarshall delegator and return
	delegator, err = types.UnmarshallDelegator(k.cdc, store.Get(key))
	if err != nil {
		return delegator, err
	}
	// return true if delegator
	return delegator, nil
}

// AddDelegatorAccount adds DelegatorAccount index with DelegatorID
func (k *Keeper) AddDelegatorAccount(ctx sdk.Context, delegatorAccount types.DelegatorAccount) error {
	store := ctx.KVStore(k.storeKey)
	// marshall validator account
	bz, err := types.MarshallDelegatorAccount(k.cdc, delegatorAccount)
	if err != nil {
		return err
	}

	store.Set(GetDelegatorAccountMapKey(delegatorAccount.ID.Bytes()), bz)
	k.Logger(ctx).Debug("DelegatorAccount Stored", "key", hex.EncodeToString(GetDelegatorAccountMapKey(delegatorAccount.ID.Bytes())), "delegatorAccount", delegatorAccount.String())
	return nil
}

// GetDelegatorAccountByID will return DelegatorAccount of valID
func (k *Keeper) GetDelegatorAccountByID(ctx sdk.Context, delegatorID types.DelegatorID) (delegatorAccount types.DelegatorAccount, err error) {

	// check if delegator account exists
	if !k.CheckIfDelegatorAccountExists(ctx, delegatorID) {
		return delegatorAccount, errors.New("Validator Account not found")
	}

	store := ctx.KVStore(k.storeKey)
	key := GetDelegatorAccountMapKey(delegatorID.Bytes())

	// unmarshall delegator account and return
	delegatorAccount, err = types.UnMarshallDelegatorAccount(k.cdc, store.Get(key))
	if err != nil {
		return delegatorAccount, err
	}

	return delegatorAccount, nil
}

// CheckIfDelegatorAccountExists will return true if delegator account exists
func (k *Keeper) CheckIfDelegatorAccountExists(ctx sdk.Context, delegatorID types.DelegatorID) (ok bool) {
	store := ctx.KVStore(k.storeKey)
	key := GetDelegatorAccountMapKey(delegatorID.Bytes())
	if !store.Has(key) {
		return false
	}
	return true
}

// GetAllDelegatorAccounts returns all DelegatorAccountss
func (k *Keeper) GetAllDelegatorAccounts(ctx sdk.Context) (delegatorAccounts []types.DelegatorAccount) {
	// iterate through delegatorAccounts and create delegatorAccounts update array
	k.IterateDelegatorAccountsAndApplyFn(ctx, func(delegatorAccount types.DelegatorAccount) error {
		// append to list of delegatorUpdates
		delegatorAccounts = append(delegatorAccounts, delegatorAccount)
		return nil
	})

	return
}

// IterateDelegatorAccountsAndApplyFn iterate delegatorAccounts and apply the given function.
func (k *Keeper) IterateDelegatorAccountsAndApplyFn(ctx sdk.Context, f func(delegatorAccount types.DelegatorAccount) error) {
	store := ctx.KVStore(k.storeKey)

	// get validator iterator
	iterator := sdk.KVStorePrefixIterator(store, DelegatorAccountMapKey)
	defer iterator.Close()

	// loop through delegatorAccounts
	for ; iterator.Valid(); iterator.Next() {
		// unmarshall delegatorAccount
		delegatorAccount, _ := types.UnMarshallDelegatorAccount(k.cdc, iterator.Value())
		// call function and return if required
		if err := f(delegatorAccount); err != nil {
			return
		}
	}
}

// 1. Delegator is updated with Validator ID.
// 2. VotingPower of the bonded validator is updated.
// 3. shares are added to Delegator proportional to his stake and exchange rate. // delegatorshares = (delegatorstake / exchangeRate)
// 4. Exchange rate is calculated instantly.  //   ExchangeRate = (delegatedpower + delegatorRewardPool) / totaldelegatorshares
// 5. TotalDelegatorShares of bonded validator is updated.
// 6. DelegatedPower of bonded validator is updated.
func (k *Keeper) BondDelegator(ctx sdk.Context, delegatorID types.DelegatorID, valID types.ValidatorID, amount *big.Int, lastUpdated uint64) error {

	// pull delegator from store
	delegator, err := k.GetDelegatorInfo(ctx, delegatorID)
	if err != nil {
		k.Logger(ctx).Error("Fetching of delegator from store failed", "delegatorId", delegatorID)
		return errors.New("Delegator not found")
	}

	delegatorAccount, err := k.GetDelegatorAccountByID(ctx, delegatorID)
	if err != nil {
		k.Logger(ctx).Error("Fetching of delegator account from store failed", "delegatorId", delegatorID)
		return errors.New("DelegatorAccount not found")
	}

	// 2. update validator ID of delegator.
	delegator.ValID = valID

	// update last udpated
	delegator.LastUpdated = lastUpdated

	// 3. VotingPower of the bonded validator is updated.
	// pull validator from store
	validator, ok := k.sk.GetValidatorFromValID(ctx, valID)
	if !ok {
		k.Logger(ctx).Error("Fetching of bonded validator from store failed", "validatorId", valID)
		return errors.New("Validator not found")
	}

	p, err := helper.GetPowerFromAmount(amount)
	if err != nil {
		k.Logger(ctx).Error("Unable to convert amount to power", "Amount", amount)
		return errors.New("Unable to convert amount to power")
	}

	// 4. shares are added to Delegator proportional to his stake and exchange rate.
	// delegatorshares = (delegatorstake / exchangeRate)
	delegatorshares := float32(p.Int64()) / validator.ExchangeRate()

	// add shares to delegator account
	delegatorAccount.Shares += delegatorshares

	// 6. TotalDelegatorShares of bonded validator is updated.
	validator.TotalDelegatorShares += delegatorshares

	validator.VotingPower += p.Int64()

	// 7. DelegatedPower of bonded validator is updated.
	validator.DelegatedPower += p.Int64()

	// save delegator
	err = k.AddDelegator(ctx, delegator)
	if err != nil {
		k.Logger(ctx).Error("Unable to update delegator", "error", err, "DelegatorID", delegator.ID)
		return errors.New("Delegator updation failed")
	}

	// save delegator account
	err = k.AddDelegatorAccount(ctx, delegatorAccount)
	if err != nil {
		k.Logger(ctx).Error("Unable to update delegatorAccount", "error", err, "DelegatorID", delegator.ID)
		return errors.New("DelegatorAccount updation failed")
	}

	// save validator
	err = k.sk.AddValidator(ctx, validator)
	if err != nil {
		k.Logger(ctx).Error("Unable to update validator", "error", err, "ValidatorID", validator.ID)
		return errors.New("Validator updation failed")
	}
	return nil
}

// HandleMsgDelegatorUnBond msg delegator unbond with validator
// ** stake calculations **
// 1. On Bonding event, Validator will send MsgDelegatorUnBond transaction to heimdall.
// 2. Delegator is updated with Validator ID = 0.
// 3. VotingPower of bonded validator is reduced.
// 4. DelegatedPower of the bonded validator is reduced after reward calculation.

// ** reward calculations **
// 1. Exchange rate is calculated instantly.  ExchangeRate = (delegatedpower + delegatorRewardPool) / totaldelegatorshares
// 2. Based on exchange rate and no of shares delegator holds, totalReturns for delegator is calculated.  `totalReturns = exchangeRate * noOfShares`
// 3. Delegator RewardAmount += totalReturns - delegatorVotingPower
// 4. Add RewardAmount to DelegatorAccount .
// 5. Reduce TotalDelegatorShares of bonded validator.
// 6. Reduce DelgatorRewardPool of bonded validator.
// 7. make shares = 0 on Delegator Account.
func (k *Keeper) UnBondDelegator(ctx sdk.Context, delegatorID types.DelegatorID, lastUpdated uint64) error {

	// pull delegator from store
	delegator, err := k.GetDelegatorInfo(ctx, delegatorID)
	if err != nil {
		k.Logger(ctx).Error("Fetching of delegator from store failed", "delegatorId", delegatorID)
		return errors.New("Delegator not found")
	}

	if delegator.ValID == 0 {
		k.Logger(ctx).Error("Delegator already unbonded", "delegatorId", delegatorID)
		return errors.New("Delegator already unbonded")
	}

	delegatorAccount, err := k.GetDelegatorAccountByID(ctx, delegatorID)
	if err != nil {
		k.Logger(ctx).Error("Fetching of delegator account from store failed", "delegatorId", delegatorID)
		return errors.New("DelegatorAccount not found")
	}

	valID := delegator.ValID
	// 3. VotingPower of the bonded validator is updated.
	// pull validator from store
	validator, ok := k.sk.GetValidatorFromValID(ctx, valID)
	if !ok {
		k.Logger(ctx).Error("Fetching of bonded validator from store failed", "validatorId", valID)
		return errors.New("Validator not found")
	}

	// Get shares of delegator account
	delegatorshares := delegatorAccount.Shares

	// 6. TotalDelegatorShares of bonded validator is updated.
	validator.TotalDelegatorShares -= delegatorshares

	validator.VotingPower -= delegator.VotingPower

	// calculate rewards.
	totalDelegatorReturns := validator.ExchangeRate() * delegatorshares

	rewardAmount := totalDelegatorReturns - float32(delegator.VotingPower)

	validator.DelgatorRewardPool = int64(float32(validator.DelgatorRewardPool) - rewardAmount)

	// 7. DelegatedPower of bonded validator is updated.
	validator.DelegatedPower -= delegator.VotingPower

	// save validator
	err = k.sk.AddValidator(ctx, validator)
	if err != nil {
		k.Logger(ctx).Error("Unable to update validator", "error", err, "ValidatorID", validator.ID)
		return errors.New("error adding validator to store")
	}

	// 2. update validator ID of delegator.
	delegator.ValID = 0

	// update last udpated
	delegator.LastUpdated = lastUpdated

	delegatorAccount.Shares = 0

	// save delegator account
	err = k.AddDelegatorAccount(ctx, delegatorAccount)
	if err != nil {
		k.Logger(ctx).Error("Unable to update delegatorAccount", "error", err, "DelegatorID", delegator.ID)
		return errors.New("DelegatorAccount updation failed")
	}

	// save delegator
	err = k.AddDelegator(ctx, delegator)
	if err != nil {
		k.Logger(ctx).Error("Unable to update delegator", "error", err, "DelegatorID", delegator.ID)
		return errors.New("error adding delegator to store")
	}

	return nil
}
