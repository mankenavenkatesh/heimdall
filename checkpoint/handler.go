package checkpoint

import (
	"bytes"
	"strconv"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ethCmn "github.com/ethereum/go-ethereum/common"
	"github.com/maticnetwork/heimdall/checkpoint/tags"
	"github.com/maticnetwork/heimdall/common"
	hmCommon "github.com/maticnetwork/heimdall/common"
	"github.com/maticnetwork/heimdall/helper"
	"github.com/maticnetwork/heimdall/types"
	hmTypes "github.com/maticnetwork/heimdall/types"
)

// NewHandler creates new handler for handling messages for checkpoint module
func NewHandler(k Keeper, contractCaller helper.IContractCaller) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case MsgCheckpoint:
			return handleMsgCheckpoint(ctx, msg, k, contractCaller)
		case MsgCheckpointAck:
			return handleMsgCheckpointAck(ctx, msg, k, contractCaller)
		case MsgCheckpointNoAck:
			return handleMsgCheckpointNoAck(ctx, msg, k)
		default:
			return sdk.ErrTxDecode("Invalid message in checkpoint module").Result()
		}
	}
}

// handleMsgCheckpoint Validates checkpoint transaction
func handleMsgCheckpoint(ctx sdk.Context, msg MsgCheckpoint, k Keeper, contractCaller helper.IContractCaller) sdk.Result {
	k.Logger(ctx).Debug("Validating checkpoint data", "TxData", msg)

	if msg.TimeStamp == 0 || msg.TimeStamp > uint64(time.Now().UTC().Unix()) {
		k.Logger(ctx).Error("Checkpoint timestamp must be in near past", "CurrentTime", time.Now().UTC().Unix(), "CheckpointTime", msg.TimeStamp, "Condition", msg.TimeStamp >= uint64(time.Now().UTC().Unix()))
		return common.ErrBadTimeStamp(k.Codespace()).Result()
	}

	checkpointBuffer, err := k.GetCheckpointFromBuffer(ctx)
	if err == nil {
		if msg.TimeStamp == 0 || checkpointBuffer.TimeStamp == 0 || ((msg.TimeStamp > checkpointBuffer.TimeStamp) && msg.TimeStamp-checkpointBuffer.TimeStamp >= uint64(helper.GetConfig().CheckpointBufferTime.Seconds())) {
			k.Logger(ctx).Debug("Checkpoint has been timed out, flushing buffer", "CheckpointTimestamp", msg.TimeStamp, "PrevCheckpointTimestamp", checkpointBuffer.TimeStamp)
			k.FlushCheckpointBuffer(ctx)
		} else {
			// calulates remaining time for buffer to be flushed
			checkpointTime := time.Unix(int64(checkpointBuffer.TimeStamp), 0)
			expiryTime := checkpointTime.Add(helper.GetConfig().CheckpointBufferTime)
			diff := expiryTime.Sub(time.Now().UTC()).Seconds()
			k.Logger(ctx).Error("Checkpoint already exits in buffer", "Checkpoint", checkpointBuffer.String(), "Expires", expiryTime)
			return common.ErrNoACK(k.Codespace(), diff).Result()
		}
	}
	// k.Logger(ctx).Debug("Received checkpoint from buffer", "Checkpoint", checkpointBuffer.String())

	// validate checkpoint
	if !ValidateCheckpoint(msg.StartBlock, msg.EndBlock, msg.RootHash) {
		k.Logger(ctx).Error("RootHash is not valid",
			"StartBlock", msg.StartBlock,
			"EndBlock", msg.EndBlock,
			"RootHash", msg.RootHash)
		return common.ErrBadBlockDetails(k.Codespace()).Result()
	}

	k.Logger(ctx).Debug("Valid Roothash in checkpoint", "StartBlock", msg.StartBlock, "EndBlock", msg.EndBlock)

	// fetch last checkpoint from store
	if lastCheckpoint, err := k.GetLastCheckpoint(ctx); err == nil {
		// make sure new checkpoint is after tip
		if lastCheckpoint.EndBlock > msg.StartBlock {
			k.Logger(ctx).Error("Checkpoint already exists",
				"currentTip", lastCheckpoint.EndBlock,
				"startBlock", msg.StartBlock)
			return common.ErrOldCheckpoint(k.Codespace()).Result()
		}
		if lastCheckpoint.EndBlock+1 != msg.StartBlock {
			k.Logger(ctx).Error("Checkpoint not in countinuity",
				"currentTip", lastCheckpoint.EndBlock,
				"startBlock", msg.StartBlock)
			return common.ErrDisCountinuousCheckpoint(k.Codespace()).Result()
		}
		// make sure latest AccountRootHash matches
		if !bytes.Equal(lastCheckpoint.AccountRootHash.Bytes(), msg.AccountRootHash.Bytes()) {
			k.Logger(ctx).Error("AccountRootHash of LastCheckpoint", lastCheckpoint.AccountRootHash,
				"doesn't match with AccountRootHash of msg", msg.AccountRootHash)
			return common.ErrBadBlockDetails(k.Codespace()).Result()
		}
	} else if err.Error() == common.ErrNoCheckpointFound(k.Codespace()).Error() && msg.StartBlock != 0 {
		k.Logger(ctx).Error("First checkpoint to start from block 1", "Error", err)
		return common.ErrBadBlockDetails(k.Codespace()).Result()
	} else if err.Error() == common.ErrNoCheckpointFound(k.Codespace()).Error() && msg.StartBlock == 0 {
		// Check if genesis AccountRootHash matches
		genesisValidatorAccounts := k.sk.GetAllValidatorAccounts(ctx)
		genesisaccountRootHash, err := GetAccountRootHash(genesisValidatorAccounts)
		if err != nil {
			k.Logger(ctx).Error("Error calculating genesis accountroothash", err)
			return common.ErrComputeGenesisAccountRoot(k.Codespace()).Result()
		}
		if !bytes.Equal(genesisaccountRootHash, msg.AccountRootHash.Bytes()) {
			k.Logger(ctx).Error("Genesis AccountRootHash", types.BytesToHeimdallHash(genesisaccountRootHash).String(),
				"doesn't match with Genesis AccountRootHash of msg", msg.AccountRootHash)
			return common.ErrAccountRootMismatch(k.Codespace()).Result()
		}
	}
	k.Logger(ctx).Debug("Valid checkpoint tip")
	k.Logger(ctx).Debug("AccountRootHash matches")

	// check proposer in message
	if !bytes.Equal(msg.Proposer.Bytes(), k.sk.GetValidatorSet(ctx).Proposer.Signer.Bytes()) {
		k.Logger(ctx).Error("Invalid proposer in message",
			"currentProposer", k.sk.GetValidatorSet(ctx).Proposer.Signer.String(),
			"checkpointProposer", msg.Proposer.String())
		return common.ErrBadProposerDetails(k.Codespace(), k.sk.GetValidatorSet(ctx).Proposer.Signer).Result()
	}
	k.Logger(ctx).Debug("Valid proposer in checkpoint")

	// check if proposer has min ether
	// balance, _ := contractCaller.GetBalance(msg.Proposer.EthAddress())
	// if balance.Cmp(helper.MinBalance) == -1 {
	// 	k.Logger(ctx).Error("Proposer doesnt have enough ether to send checkpoint tx", "Balance", balance, "RequiredBalance", helper.MinBalance)
	// 	return common.ErrLowBalance(k.Codespace(), msg.Proposer.String()).Result()
	// }

	// add checkpoint to buffer
	// Add AccountRootHash to CheckpointBuffer
	k.SetCheckpointBuffer(ctx, hmTypes.CheckpointBlockHeader{
		StartBlock:      msg.StartBlock,
		EndBlock:        msg.EndBlock,
		RootHash:        msg.RootHash,
		AccountRootHash: msg.AccountRootHash,
		Proposer:        msg.Proposer,
		TimeStamp:       msg.TimeStamp,
	})

	checkpoint, _ := k.GetCheckpointFromBuffer(ctx)
	k.Logger(ctx).Debug("Adding good checkpoint to buffer to await ACK", "checkpointStored", checkpoint.String())

	resTags := sdk.NewTags(
		tags.Proposer, []byte(msg.Proposer.String()),
		tags.StartBlock, []byte(strconv.FormatUint(uint64(msg.StartBlock), 10)),
		tags.EndBlock, []byte(strconv.FormatUint(uint64(msg.EndBlock), 10)),
	)

	// send tags
	return sdk.Result{Tags: resTags}
}

// handleMsgCheckpointAck Validates if checkpoint submitted on chain is valid
func handleMsgCheckpointAck(ctx sdk.Context, msg MsgCheckpointAck, k Keeper, contractCaller helper.IContractCaller) sdk.Result {
	k.Logger(ctx).Debug("Validating Checkpoint ACK", "Tx", msg)

	// make call to headerBlock with header number
	root, start, end, createdAt, proposer, err := contractCaller.GetHeaderInfo(msg.HeaderBlock)
	if err != nil {
		k.Logger(ctx).Error("Unable to fetch header from rootchain contract", "Error", err, "headerBlockIndex", msg.HeaderBlock)
		return common.ErrBadAck(k.Codespace()).Result()
	}

	// check confirmation
	latestBlock, err := contractCaller.GetMainChainBlock(nil)
	if err != nil {
		k.Logger(ctx).Error("Unable to connect to mainchain", "Error", err)
		return common.ErrNoConn(k.Codespace()).Result()
	}

	if latestBlock.Number.Uint64()-createdAt < helper.GetConfig().ConfirmationBlocks {
		k.Logger(ctx).Error("Not enough confirmations", "latestBlock", latestBlock.Number.Uint64(), "txBlock", createdAt)
		return common.ErrWaitForConfirmation(k.Codespace()).Result()
	}

	k.Logger(ctx).Debug("HeaderBlock fetched",
		"headerBlock", msg.HeaderBlock,
		"start", start,
		"end", end,
		"roothash", root,
		"proposer", proposer,
		"createdAt", createdAt,
		"latest", latestBlock.Number.Uint64(),
	)

	// get last checkpoint from buffer
	headerBlock, err := k.GetCheckpointFromBuffer(ctx)
	if err != nil {
		k.Logger(ctx).Error("Unable to get checkpoint", "error", err)
		return common.ErrBadAck(k.Codespace()).Result()
	}
	if start != headerBlock.StartBlock {
		k.Logger(ctx).Error("Invalid start block", "startExpected", headerBlock.StartBlock, "startReceived", start)
		return common.ErrBadAck(k.Codespace()).Result()
	} else if start == headerBlock.StartBlock && end == headerBlock.EndBlock && !bytes.Equal(root.Bytes(), headerBlock.RootHash.Bytes()) {
		k.Logger(ctx).Error("Invalid ACK",
			"startExpected", headerBlock.StartBlock,
			"startReceived", start,
			"endExpected", headerBlock.EndBlock,
			"endReceived", end,
			"rootExpected", headerBlock.RootHash.String(),
			"rootRecieved", root.String())
		return common.ErrBadAck(k.Codespace()).Result()
	}
	if headerBlock.EndBlock > end {
		k.Logger(ctx).Info("Adjusting endBlock to one already submitted on chain", "OldEndBlock", headerBlock.EndBlock, "AdjustedEndBlock", end)
		headerBlock.EndBlock = end
		headerBlock.RootHash = types.HeimdallHash(root)
		// TODO proposer also needs to be changed
	}

	// Get Tx hash from ack msg
	txHash := msg.TxHash

	// Fetch all the signatures from tx input data and calculate signer rewards
	voteBytes, sigInput, _, err := contractCaller.GetCheckpointSign(ctx, ethCmn.Hash(txHash))
	if err != nil {
		k.Logger(ctx).Error("Error while fetching signers from transaction", "error", err)
		return common.ErrFetchCheckpointSigners(k.Codespace()).Result()
	}

	// get main tx receipt
	receipt, err := contractCaller.GetConfirmedTxReceipt(msg.TxHash.EthHash())
	if err != nil || receipt == nil {
		return hmCommon.ErrWaitForConfirmation(k.Codespace()).Result()
	}

	eventLog, err := contractCaller.DecodeNewHeaderBlockEvent(receipt, msg.LogIndex)
	if err != nil || eventLog == nil {
		k.Logger(ctx).Error("Error fetching log from txhash")
		return hmCommon.ErrInvalidMsg(k.Codespace(), "Unable to fetch logs for txHash").Result()
	}

	k.Logger(ctx).Info("Fetched checkpoint reward from event", eventLog.Reward)

	// Calculate Signer Rewards
	signerRewards, err := k.sk.CalculateSignerRewards(ctx, voteBytes, sigInput, eventLog.Reward)
	if err != nil {
		k.Logger(ctx).Error("Error while calculating Signer Rewards", "error", err)
		return common.ErrComputeCheckpointRewards(k.Codespace()).Result()
	}

	// update store with new rewards
	k.sk.UpdateValidatorRewards(ctx, signerRewards)
	k.Logger(ctx).Info("Signer Rewards updated to store")

	// Calculate new account root hash
	valRewardMap := k.sk.GetAllValidatorAccounts(ctx)
	k.Logger(ctx).Debug("validatoraccounts of all validators", "valAccounts", valAccounts)
	accountRoot, err := GetAccountRootHash(valAccounts)
	k.Logger(ctx).Info("Validator Account root hash generated", "AccountRootHash", types.BytesToHeimdallHash(accountRoot).String())

	// Add new AccountRootHash root hash to bufferedcheckpoint header block
	headerBlock.AccountRootHash = types.BytesToHeimdallHash(accountRoot)

	// Add checkpoint to headerBlocks
	k.AddCheckpoint(ctx, msg.HeaderBlock, *headerBlock)
	k.Logger(ctx).Info("Checkpoint added to store", "headerBlock", headerBlock.String())

	// flush buffer
	k.FlushCheckpointBuffer(ctx)
	k.Logger(ctx).Debug("Checkpoint buffer flushed after receiving checkpoint ack", "checkpoint", headerBlock)

	// update ack count
	k.UpdateACKCount(ctx)
	k.Logger(ctx).Debug("Valid ack received", "CurrentACKCount", k.GetACKCount(ctx)-1, "UpdatedACKCount", k.GetACKCount(ctx))

	// --- Update to new proposer

	// increment accum
	k.sk.IncrementAccum(ctx, 1)

	//log new proposer
	vs := k.sk.GetValidatorSet(ctx)
	newProposer := vs.GetProposer()
	k.Logger(ctx).Debug(
		"New proposer selected",
		"validator", newProposer.Signer.String(),
		"signer", newProposer.Signer.String(),
		"power", newProposer.VotingPower,
	)

	resTags := sdk.NewTags(
		tags.HeaderIndex, []byte(strconv.FormatUint(uint64(msg.HeaderBlock), 10)),
	)

	return sdk.Result{Tags: resTags}
}

// Validate checkpoint no-ack transaction
func handleMsgCheckpointNoAck(ctx sdk.Context, msg MsgCheckpointNoAck, k Keeper) sdk.Result {
	k.Logger(ctx).Debug("Validating checkpoint no-ack", "TxData", msg)
	// current time
	currentTime := time.Unix(int64(msg.TimeStamp), 0) // buffer time
	bufferTime := helper.GetConfig().CheckpointBufferTime

	// fetch last checkpoint from store
	// TODO figure out how to handle this error
	lastCheckpoint, _ := k.GetLastCheckpoint(ctx)
	lastCheckpointTime := time.Unix(int64(lastCheckpoint.TimeStamp), 0)

	// if last checkpoint is not present or last checkpoint happens before checkpoint buffer time -- thrown an error
	if lastCheckpointTime.After(currentTime) || (currentTime.Sub(lastCheckpointTime) < bufferTime) {
		k.Logger(ctx).Debug("Invalid No ACK -- ongoing buffer period")
		return common.ErrInvalidNoACK(k.Codespace()).Result()
	}

	// check last no ack - prevents repetitive no-ack
	lastAck := k.GetLastNoAck(ctx)
	lastAckTime := time.Unix(int64(lastAck), 0)

	if lastAckTime.After(currentTime) || (currentTime.Sub(lastAckTime) < bufferTime) {
		k.Logger(ctx).Debug("Too many no-ack")
		return common.ErrTooManyNoACK(k.Codespace()).Result()
	}

	// set last no ack
	k.SetLastNoAck(ctx, uint64(currentTime.Unix()))
	k.Logger(ctx).Debug("Last No-ACK time set", "LastNoAck", k.GetLastNoAck(ctx))

	// --- Update to new proposer

	// increment accum
	k.sk.IncrementAccum(ctx, 1)

	//log new proposer
	vs := k.sk.GetValidatorSet(ctx)
	newProposer := vs.GetProposer()
	k.Logger(ctx).Debug(
		"New proposer selected",
		"validator", newProposer.Signer.String(),
		"signer", newProposer.Signer.String(),
		"power", newProposer.VotingPower,
	)

	resTags := sdk.NewTags(
		tags.NewProposer, []byte(newProposer.Signer.String()),
	)

	// --- End
	return sdk.Result{Tags: resTags}
}
