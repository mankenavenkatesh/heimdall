package common

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/maticnetwork/heimdall/helper"
	"github.com/maticnetwork/heimdall/types"
)

type CodeType = sdk.CodeType

const (
	DefaultCodespace sdk.CodespaceType = "1"

	CodeInvalidMsg CodeType = 1400

	CodeInvalidProposerInput     CodeType = 1500
	CodeInvalidBlockInput        CodeType = 1501
	CodeInvalidACK               CodeType = 1502
	CodeNoACK                    CodeType = 1503
	CodeBadTimeStamp             CodeType = 1504
	CodeInvalidNoACK             CodeType = 1505
	CodeTooManyNoAck             CodeType = 1506
	CodeLowBal                   CodeType = 1507
	CodeNoCheckpoint             CodeType = 1508
	CodeOldCheckpoint            CodeType = 1509
	CodeDisCountinuousCheckpoint CodeType = 1510

	CodeOldValidator       CodeType = 2500
	CodeNoValidator        CodeType = 2501
	CodeValSignerMismatch  CodeType = 2502
	CodeValidatorExitDeny  CodeType = 2503
	CodeValAlreadyUnbonded CodeType = 2504
	CodeSignerSynced       CodeType = 2505
	CodeValSave            CodeType = 2506
	CodeValAlreadyJoined   CodeType = 2507
	CodeSignerUpdateError  CodeType = 2508
	CodeNoConn             CodeType = 2509
	CodeWaitFrConfirmation CodeType = 2510

	CodeSpanNotCountinuous CodeType = 3501
	CodeUnableToFreezeSet  CodeType = 3502
	CodeSpanNotFound       CodeType = 3503
	CodeValSetMisMatch     CodeType = 3504
	CodeProducerMisMatch   CodeType = 3505

	CodeFetchCheckpointSigners       CodeType = 4501
	CodeErrComputeSignerRewards      CodeType = 4502
	CodeErrComputeGenesisAccountRoot CodeType = 4503
	CodeAccountRootMismatch          CodeType = 4504
	CodeErrComputeCheckpointReward   CodeType = 4505

	CodeOldDelegator       CodeType = 6500
	CodeNoDelegator        CodeType = 6501
	CodeDelegatorExitDeny  CodeType = 6502
	CodeDelAlreadyJoined   CodeType = 6503
	CodeDelAlreadyBonded   CodeType = 6504
	CodeDelAlreadyUnbonded CodeType = 6505
	CodeDelSave            CodeType = 6506
)

// -------- Invalid msg

func ErrInvalidMsg(codespace sdk.CodespaceType, format string, args ...interface{}) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidMsg, format, args...)
}

// -------- Checkpoint Errors

func ErrBadProposerDetails(codespace sdk.CodespaceType, proposer types.HeimdallAddress) sdk.Error {
	return newError(codespace, CodeInvalidProposerInput, fmt.Sprintf("Proposer is not valid, current proposer is %v", proposer.String()))
}

func ErrBadBlockDetails(codespace sdk.CodespaceType) sdk.Error {
	return newError(codespace, CodeInvalidBlockInput, "Wrong roothash for given start and end block numbers")
}

func ErrBadAck(codespace sdk.CodespaceType) sdk.Error {
	return newError(codespace, CodeInvalidACK, "Ack Not Valid")
}

func ErrOldCheckpoint(codespace sdk.CodespaceType) sdk.Error {
	return newError(codespace, CodeOldCheckpoint, "Checkpoint already received for given start and end block")
}

func ErrDisCountinuousCheckpoint(codespace sdk.CodespaceType) sdk.Error {
	return newError(codespace, CodeDisCountinuousCheckpoint, "Checkpoint not in countinuity")
}

func ErrNoACK(codespace sdk.CodespaceType, timeRemaining float64) sdk.Error {
	return newError(codespace, CodeNoACK, fmt.Sprintf("Checkpoint Already Exists In Buffer, ACK expected ,expires %v", timeRemaining))
}

func ErrNoConn(codespace sdk.CodespaceType) sdk.Error {
	return newError(codespace, CodeNoConn, "Unable to connect to chain")
}

func ErrWaitForConfirmation(codespace sdk.CodespaceType) sdk.Error {
	return newError(codespace, CodeWaitFrConfirmation, fmt.Sprintf("Please wait for %v confirmations before sending transaction", helper.GetConfig().ConfirmationBlocks))
}

func ErrNoCheckpointFound(codespace sdk.CodespaceType) sdk.Error {
	return newError(codespace, CodeNoCheckpoint, "Checkpoint Not Found")
}

func ErrInvalidNoACK(codespace sdk.CodespaceType) sdk.Error {
	return newError(codespace, CodeInvalidNoACK, "Invalid no-ack")
}

func ErrTooManyNoACK(codespace sdk.CodespaceType) sdk.Error {
	return newError(codespace, CodeTooManyNoAck, "Too many no-acks")
}

func ErrBadTimeStamp(codespace sdk.CodespaceType) sdk.Error {
	return newError(codespace, CodeBadTimeStamp, "Invalid time stamp. It must be in near past.")
}

func ErrLowBalance(codespace sdk.CodespaceType, address string) sdk.Error {
	return newError(codespace, CodeLowBal, fmt.Sprintf("Min bal %v required for sending checkpoint TX for address %v", helper.MinBalance, address))
}

// ----------- Staking Errors

func ErrOldValidator(codespace sdk.CodespaceType) sdk.Error {
	return newError(codespace, CodeOldValidator, "Start Epoch behind Current Epoch")
}

func ErrNoValidator(codespace sdk.CodespaceType) sdk.Error {
	return newError(codespace, CodeNoValidator, "Validator information not found")
}

func ErrValSignerMismatch(codespace sdk.CodespaceType) sdk.Error {
	return newError(codespace, CodeValSignerMismatch, "Signer Address doesnt match pubkey address")
}

func ErrValIsNotCurrentVal(codespace sdk.CodespaceType) sdk.Error {
	return newError(codespace, CodeValidatorExitDeny, "Validator is not in validator set, exit not possible")
}

func ErrValUnbonded(codespace sdk.CodespaceType) sdk.Error {
	return newError(codespace, CodeValAlreadyUnbonded, "Validator already unbonded , cannot exit")
}

func ErrSignerUpdateError(codespace sdk.CodespaceType) sdk.Error {
	return newError(codespace, CodeSignerUpdateError, "Signer update error")
}

func ErrOldTx(codespace sdk.CodespaceType) sdk.Error {
	return newError(codespace, CodeSignerUpdateError, "Old txhash not allowed")
}

func ErrValidatorAlreadySynced(codespace sdk.CodespaceType) sdk.Error {
	return newError(codespace, CodeSignerSynced, "No signer update found, invalid message")
}

func ErrValidatorSave(codespace sdk.CodespaceType) sdk.Error {
	return newError(codespace, CodeValSave, "Cannot save validator")
}

func ErrValidatorNotDeactivated(codespace sdk.CodespaceType) sdk.Error {
	return newError(codespace, CodeValSave, "Validator Not Deactivated")
}

func ErrValidatorAlreadyJoined(codespace sdk.CodespaceType) sdk.Error {
	return newError(codespace, CodeValAlreadyJoined, "Validator already joined")
}

// ----------- Reward Errors
func ErrFetchCheckpointSigners(codespace sdk.CodespaceType) sdk.Error {
	return newError(codespace, CodeFetchCheckpointSigners, "Error Fetching checkpoint signatures from tx")
}

func ErrComputeCheckpointRewards(codespace sdk.CodespaceType) sdk.Error {
	return newError(codespace, CodeErrComputeCheckpointReward, "Error while computing checkpoint reward")
}

func ErrComputeGenesisAccountRoot(codespace sdk.CodespaceType) sdk.Error {
	return newError(codespace, CodeErrComputeGenesisAccountRoot, "Error while computing Genesis Account Root")
}

func ErrAccountRootMismatch(codespace sdk.CodespaceType) sdk.Error {
	return newError(codespace, CodeAccountRootMismatch, "Account Root hash mismatch")
}

// ----------- Delegation Errors
func ErrOldDelegator(codespace sdk.CodespaceType) sdk.Error {
	return newError(codespace, CodeOldDelegator, "Start Epoch behind Current Epoch")
}

func ErrNoDelegator(codespace sdk.CodespaceType) sdk.Error {
	return newError(codespace, CodeNoDelegator, "Delegator information not found")
}

func ErrDelegatorSave(codespace sdk.CodespaceType) sdk.Error {
	return newError(codespace, CodeDelSave, "Cannot save Delegator")
}

func ErrDelegatorAlreadyJoined(codespace sdk.CodespaceType) sdk.Error {
	return newError(codespace, CodeDelAlreadyJoined, "Delegator already joined")
}

func ErrAlreadyBonded(codespace sdk.CodespaceType) sdk.Error {
	return newError(codespace, CodeDelAlreadyBonded, "Delegator already bonded")
}

// Bor Errors --------------------------------

func ErrSpanNotInCountinuity(codespace sdk.CodespaceType) sdk.Error {
	return newError(codespace, CodeSpanNotCountinuous, "Span not countinuous")
}

func ErrSpanNotFound(codespace sdk.CodespaceType) sdk.Error {
	return newError(codespace, CodeSpanNotFound, "Span not found")
}

func ErrUnableToFreezeValSet(codespace sdk.CodespaceType) sdk.Error {
	return newError(codespace, CodeUnableToFreezeSet, "Unable to freeze validator set for next span")
}

func ErrValSetMisMatch(codespace sdk.CodespaceType) sdk.Error {
	return newError(codespace, CodeValSetMisMatch, "Validator set mismatch")
}

func ErrProducerMisMatch(codespace sdk.CodespaceType) sdk.Error {
	return newError(codespace, CodeProducerMisMatch, "Producer set mismatch")
}

func codeToDefaultMsg(code CodeType) string {
	switch code {
	case CodeInvalidBlockInput:
		return "Invalid Block Input"
	default:
		return sdk.CodeToDefaultMsg(code)
	}
}

func msgOrDefaultMsg(msg string, code CodeType) string {
	if msg != "" {
		return msg
	}
	return codeToDefaultMsg(code)
}

func newError(codespace sdk.CodespaceType, code CodeType, msg string) sdk.Error {
	msg = msgOrDefaultMsg(msg, code)
	return sdk.NewError(codespace, code, msg)
}
