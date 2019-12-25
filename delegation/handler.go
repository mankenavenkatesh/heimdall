package delegation

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	hmCommon "github.com/maticnetwork/heimdall/common"
	"github.com/maticnetwork/heimdall/helper"
	"github.com/maticnetwork/heimdall/types"
)

// NewHandler new handler
func NewHandler(k Keeper, contractCaller helper.IContractCaller) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case MsgDelegatorJoin:
			return HandleMsgDelegatorJoin(ctx, msg, k, contractCaller)
		// case MsgDelegatorStakeUpdate:
		// 	return HandleMsgDelegatorStakeUpdate(ctx, msg, k, contractCaller)
		// case MsgDelegatorUnstake:
		// 	return HandleMsgDelegatorUnstake(ctx, msg, k, contractCaller)
		// case MsgDelegatorBond:
		// 	return HandleMsgDelegatorBond(ctx, msg, k, contractCaller)
		// case MsgDelegatorUnBond:
		// 	return HandleMsgDelegatorUnBond(ctx, msg, k, contractCaller)
		// case MsgDelegatorRebond:
		// 	return HandleMsgDelegatorRebond(ctx, msg, k, contractCaller)
		default:
			return sdk.ErrTxDecode("Invalid message in delegation module").Result()
		}
	}
}

// HandleMsgDelegatorJoin msg delegator join
func HandleMsgDelegatorJoin(ctx sdk.Context, msg MsgDelegatorJoin, k Keeper, contractCaller helper.IContractCaller) sdk.Result {
	k.Logger(ctx).Debug("Handling new delegator join", "msg", msg)

	// check if transaction is confirmed.
	if confirmed := contractCaller.IsTxConfirmed(msg.TxHash.EthHash()); !confirmed {
		return hmCommon.ErrWaitForConfirmation(k.Codespace()).Result()
	}

	// Fetch the Delegator from root chain contract using delegator ID from msg.
	delegator, err := contractCaller.GetDelegator(msg.ID)
	if err != nil {
		k.Logger(ctx).Error(
			"Unable to fetch delegator from rootchain",
			"error", err,
		)
		return hmCommon.ErrNoValidator(k.Codespace()).Result()
	}

	k.Logger(ctx).Debug("Fetched delegator from rootchain successfully", "delegator", delegator.String())

	// Check if delegator has been delegator before
	if _, err := k.GetDelegatorInfo(ctx, msg.ID); err != nil {
		k.Logger(ctx).Error("Delegator has been a delegator before, cannot join with same ID", "delegatorID", msg.ID)
		return hmCommon.ErrDelegatorAlreadyJoined(k.Codespace()).Result()
	}

	// create new delegator
	newDelegator := types.Delegator{
		ID:          delegator.ID,
		VotingPower: delegator.VotingPower,
		ValID:       delegator.ValID,
	}

	// add validator to store
	k.Logger(ctx).Debug("Adding new delegator to state", "delegator", newDelegator.String())
	err = k.AddDelegator(ctx, newDelegator)
	if err != nil {
		k.Logger(ctx).Error("Unable to add delegator to state", "error", err, "validator", newDelegator.String())
		return hmCommon.ErrDelegatorSave(k.Codespace()).Result()
	}

	// Add Delegator to DelegatorStore
	return sdk.Result{}
}

// // HandleMsgDelegatorStakeUpdate msg delegator stake update
// func HandleMsgDelegatorStakeUpdate(ctx sdk.Context, msg MsgDelegatorStakeUpdate, k Keeper, contractCaller helper.IContractCaller) sdk.Result {
// 	k.Logger(ctx).Debug("Handing new delegator join", "msg", msg)

// 	return sdk.Result{Tags: resTags}
// }

// // HandleMsgDelegatorUnstake msg delegator exit
// func HandleMsgDelegatorUnstake(ctx sdk.Context, msg MsgDelegatorUnstake, k Keeper, contractCaller helper.IContractCaller) sdk.Result {
// 	k.Logger(ctx).Debug("Handling delegator unstake", "msg", msg)

// 	return sdk.Result{Tags: resTags}
// }

// // HandleMsgDelegatorBond msg delegator Bond with Validator
// // 1. On Bonding event, Validator to whom delegator is bonded will send `MsgDelegatorBond` transaction to Heimdall.
// // 2. Delegator is updated with Validator ID.
// // 3. VotingPower of the bonded validator is updated.
// // 4. shares are added to Delegator proportional to his stake and exchange rate. // delegatorshares = (delegatorstake / exchangeRate)
// // 5. Exchange rate is calculated instantly.  //   ExchangeRate = (delegatedpower + delegatorRewardPool) / totaldelegatorshares
// // 6. TotalDelegatorShares of bonded validator is updated.
// // 7. DelegatedPower of bonded validator is updated.
// func HandleMsgDelegatorBond(ctx sdk.Context, msg MsgDelegatorBond, k Keeper, contractCaller helper.IContractCaller) sdk.Result {

// 	k.Logger(ctx).Debug("Handling delegator bond with validator", "Delegator", msg.ID, "Validator", msg.ValID)

// 	// get main tx receipt
// 	receipt, err := contractCaller.GetConfirmedTxReceipt(msg.TxHash.EthHash())
// 	if err != nil || receipt == nil {
// 		return hmCommon.ErrWaitForConfirmation(k.Codespace()).Result()
// 	}

// 	eventLog, err := contractCaller.DecodeDelegatorBondEvent(receipt, msg.LogIndex)
// 	if err != nil || eventLog == nil {
// 		k.Logger(ctx).Error("Error fetching log from txhash")
// 		return hmCommon.ErrInvalidMsg(k.Codespace(), "Unable to fetch delegator bond log for txHash").Result()
// 	}

// 	if eventLog.delegatorId.Uint64() != msg.ID.Uint64() {
// 		k.Logger(ctx).Error("Delegator ID in message doesnt match delegaot id in logs", "MsgID", msg.ID, "IdFromTx", eventLog.delegatorId.Uint64())
// 		return hmCommon.ErrInvalidMsg(k.Codespace(), "Invalid txhash, id's dont match. Id from tx hash is %v", eventLog.delegatorId.Uint64()).Result()
// 	}

// 	if eventLog.validatorId.Uint64() != msg.ValID.Uint64() {
// 		k.Logger(ctx).Error("Validator ID in message doesnt match Validator id in logs", "MsgID", msg.ValID, "IdFromTx", eventLog.validatorId.Uint64())
// 		return hmCommon.ErrInvalidMsg(k.Codespace(), "Invalid txhash, id's dont match. Id from tx hash is %v", eventLog.validatorId.Uint64()).Result()
// 	}

// 	// pull delegator from store
// 	delegator, ok := k.GetDelegatorInfo(ctx, msg.ID)
// 	if !ok {
// 		k.Logger(ctx).Error("Fetching of delegator from store failed", "delegatorId", msg.ID)
// 		return hmCommon.ErrNoDelegator(k.Codespace()).Result()
// 	}

// 	// last updated
// 	lastUpdated := (receipt.BlockNumber.Uint64() * stakingTypes.DefaultLogIndexUnit) + msg.LogIndex

// 	// check if incoming tx is older
// 	if lastUpdated <= delegator.LastUpdated {
// 		k.Logger(ctx).Error("Older invalid tx found")
// 		return hmCommon.ErrOldTx(k.Codespace()).Result()
// 	}

// 	// check if delegator is already bonded
// 	if delegator.ValID != 0 {
// 		k.Logger(ctx).Error("Delegator is already bonded")
// 		return hmCommon.ErrAlreadyBonded(k.Codespace()).Result()
// 	}

// 	k.BondDelegator(ctx, msg.ID, msg.ValID, lastUpdated)

// 	resTags := sdk.NewTags(
// 		tags.SignerUpdate, []byte(newSigner.String()),
// 		tags.UpdatedAt, []byte(strconv.FormatUint(validator.LastUpdated, 10)),
// 		tags.ValidatorID, []byte(strconv.FormatUint(validator.ID.Uint64(), 10)),
// 	)

// 	return sdk.Result{Tags: resTags}
// }

// // HandleMsgDelegatorUnBond msg delegator unbond with validator
// // ** stake calculations **
// // 1. On Bonding event, Validator will send MsgDelegatorUnBond transaction to heimdall.
// // 2. Delegator is updated with Validator ID = 0.
// // 3. VotingPower of bonded validator is reduced.
// // 4. DelegatedPower of the bonded validator is reduced after reward calculation.

// // ** reward calculations **
// // 1. Exchange rate is calculated instantly.  ExchangeRate = (delegatedpower + delegatorRewardPool) / totaldelegatorshares
// // 2. Based on exchange rate and no of shares delegator holds, totalReturns for delegator is calculated.  `totalReturns = exchangeRate * noOfShares`
// // 3. Delegator RewardAmount += totalReturns - delegatorVotingPower
// // 4. Add RewardAmount to DelegatorAccount .
// // 5. Reduce TotalDelegatorShares of bonded validator.
// // 6. Reduce DelgatorRewardPool of bonded validator.
// // 7. make shares = 0 on Delegator Account.
// func HandleMsgDelegatorUnBond(ctx sdk.Context, msg MsgDelegatorUnBond, k Keeper, contractCaller helper.IContractCaller) sdk.Result {
// 	k.Logger(ctx).Debug("Handling new delegator join", "msg", msg)

// 	return sdk.Result{Tags: resTags}
// }

// // HandleMsgDelegatorRebond msg delegator rebond with new validator
// func HandleMsgDelegatorRebond(ctx sdk.Context, msg MsgDelegatorRebond, k Keeper, contractCaller helper.IContractCaller) sdk.Result {
// 	k.Logger(ctx).Debug("Handling new delegator join", "msg", msg)

// 	return sdk.Result{Tags: resTags}
// }
