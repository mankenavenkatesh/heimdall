package bank

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/maticnetwork/heimdall/auth"
	"github.com/maticnetwork/heimdall/bank/types"
	hmCommon "github.com/maticnetwork/heimdall/common"
	"github.com/maticnetwork/heimdall/helper"
	hmTypes "github.com/maticnetwork/heimdall/types"
)

// NewHandler returns a handler for "bank" type messages.
func NewHandler(k Keeper, contractCaller helper.IContractCaller) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case types.MsgSend:
			return handleMsgSend(ctx, k, msg)
		case types.MsgMultiSend:
			return handleMsgMultiSend(ctx, k, msg)
		case types.MsgTopup:
			return handleMsgTopup(ctx, k, msg, contractCaller)
		case types.MsgWithdrawFee:
			return handleMsgWithdrawFee(ctx, k, msg)
		default:
			errMsg := "Unrecognized bank Msg type: %s" + msg.Type()
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

// Handle MsgSend.
func handleMsgSend(ctx sdk.Context, k Keeper, msg types.MsgSend) sdk.Result {
	if !k.GetSendEnabled(ctx) {
		return types.ErrSendDisabled(k.Codespace()).Result()
	}

	err := k.SendCoins(ctx, msg.FromAddress, msg.ToAddress, msg.Amount)
	if err != nil {
		return err.Result()
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		),
	)

	return sdk.Result{
		Events: ctx.EventManager().Events(),
	}
}

// Handle MsgMultiSend.
func handleMsgMultiSend(ctx sdk.Context, k Keeper, msg types.MsgMultiSend) sdk.Result {
	// NOTE: totalIn == totalOut should already have been checked
	if !k.GetSendEnabled(ctx) {
		return types.ErrSendDisabled(k.Codespace()).Result()
	}

	err := k.InputOutputCoins(ctx, msg.Inputs, msg.Outputs)
	if err != nil {
		return err.Result()
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		),
	)

	return sdk.Result{
		Events: ctx.EventManager().Events(),
	}
}

// Handle MsgMintFeeToken
func handleMsgTopup(ctx sdk.Context, k Keeper, msg types.MsgTopup, contractCaller helper.IContractCaller) sdk.Result {
	if !k.GetSendEnabled(ctx) {
		return types.ErrSendDisabled(k.Codespace()).Result()
	}

	// get main tx receipt
	receipt, err := contractCaller.GetConfirmedTxReceipt(msg.TxHash.EthHash())
	if err != nil || receipt == nil {
		return hmCommon.ErrWaitForConfirmation(k.Codespace()).Result()
	}

	// get event log for topup
	eventLog, err := contractCaller.DecodeValidatorTopupFeesEvent(receipt, msg.LogIndex)
	if err != nil || eventLog == nil {
		k.Logger(ctx).Error("Error fetching log from txhash")
		return hmCommon.ErrInvalidMsg(k.Codespace(), "Unable to fetch logs for txHash").Result()
	}

	if eventLog.ValidatorId.Uint64() != msg.ID.Uint64() {
		k.Logger(ctx).Error("ID in message doesn't match id in logs", "MsgID", msg.ID, "IdFromTx", eventLog.ValidatorId)
		return hmCommon.ErrInvalidMsg(k.Codespace(), "Invalid txhash and id don't match. Id from tx hash is %v", eventLog.ValidatorId.Uint64()).Result()
	}

	// fetch validator from mainchain
	validator, err := contractCaller.GetValidatorInfo(msg.ID)
	if err != nil {
		k.Logger(ctx).Error(
			"Unable to fetch validator from rootchain",
			"error", err,
		)
		return hmCommon.ErrNoValidator(k.Codespace()).Result()
	}

	// validator topup
	topupObject, err := k.GetValidatorTopup(ctx, validator.Signer)
	if err != nil {
		return types.ErrNoValidatorTopup(k.Codespace()).Result()
	}

	// create topup object
	if topupObject == nil {
		topupObject = &types.ValidatorTopup{
			ID:          validator.ID,
			TotalTopups: hmTypes.Coins{hmTypes.Coin{Denom: "matic", Amount: hmTypes.NewInt(0)}},
		}
	}

	// create topup amount
	topupAmount := hmTypes.Coins{hmTypes.Coin{Denom: "matic", Amount: hmTypes.NewIntFromBigInt(eventLog.Fee)}}

	// sequence id
	sequence := (receipt.BlockNumber.Uint64() * hmTypes.DefaultLogIndexUnit) + msg.LogIndex

	// check if incoming tx already exists
	if k.HasTopupSequence(ctx, sequence) {
		k.Logger(ctx).Error("Older invalid tx found")
		return hmCommon.ErrOldTx(k.Codespace()).Result()
	}

	// add total topups amount
	topupObject.TotalTopups = topupObject.TotalTopups.Add(topupAmount)

	// increase coins in account
	if _, ec := k.AddCoins(ctx, validator.Signer, topupAmount); ec != nil {
		return ec.Result()
	}

	// transfer fees to sender (proposer)
	if ec := k.SendCoins(ctx, validator.Signer, msg.FromAddress, auth.FeeWantedPerTx); ec != nil {
		return ec.Result()
	}

	// save old validator
	if err := k.SetValidatorTopup(ctx, validator.Signer, *topupObject); err != nil {
		k.Logger(ctx).Error("Unable to update signer", "error", err, "validatorId", validator.ID)
		return hmCommon.ErrSignerUpdateError(k.Codespace()).Result()
	}
	// save topup
	k.SetTopupSequence(ctx, sequence)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeTopup,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(types.AttributeKeyValidatorID, strconv.FormatUint(uint64(msg.ID), 10)),
			sdk.NewAttribute(types.AttributeKeyTopupAmount, strconv.FormatUint(eventLog.Fee.Uint64(), 10)),
		),
	})

	return sdk.Result{
		Events: ctx.EventManager().Events(),
	}
}

// Handle MsgWithdrawFee.
func handleMsgWithdrawFee(ctx sdk.Context, k Keeper, msg types.MsgWithdrawFee) sdk.Result {

	// check if fee is already withdrawn
	coins := k.GetCoins(ctx, msg.FromAddress)
	maticBalance := coins.AmountOf("matic")
	k.Logger(ctx).Info("Fee balance for ", "fromAddress", msg.FromAddress, "validatorId", msg.ID, "balance", maticBalance.BigInt().String())
	if maticBalance.IsZero() {
		return types.ErrNoBalanceToWithdraw(k.Codespace()).Result()
	}

	amountToWithdraw := msg.WithdrawAmount
	if amountToWithdraw.GTE(maticBalance) {
		return types.ErrInvalidWithdrawAmount(k.Codespace()).Result()
	}

	// withdraw coins of validator.
	balanceAfterWithdraw := hmTypes.Coins{hmTypes.Coin{Denom: "matic", Amount: maticBalance.Sub(amountToWithdraw)}}
	if err := k.SetCoins(ctx, msg.FromAddress, balanceAfterWithdraw); err != nil {
		k.Logger(ctx).Error("Error while setting Fee balance ", "fromAddress", msg.FromAddress, "validatorId", msg.ID, "balance", balanceAfterWithdraw, "err", err)
		return err.Result()
	}

	// Add Fee to Dividend Account
	k.AddFeeToDividendAccount(ctx, msg.ID, amountToWithdraw.BigInt())

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeFeeWithdraw,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(types.AttributeKeyValidatorID, strconv.FormatUint(uint64(msg.ID), 10)),
			sdk.NewAttribute(types.AttributeKeyFeeWithdrawAmount, strconv.FormatUint(amountToWithdraw.BigInt().Uint64(), 10)),
		),
	})

	return sdk.Result{
		Events: ctx.EventManager().Events(),
	}
}
