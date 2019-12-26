package delegation

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	hmCommon "github.com/maticnetwork/heimdall/common"
	delegationTypes "github.com/maticnetwork/heimdall/delegation/types"
	"github.com/maticnetwork/heimdall/types"
)

var cdc = codec.New()

//
// Delegator Join
//

var _ sdk.Msg = &MsgDelegatorJoin{}

type MsgDelegatorJoin struct {
	From     types.HeimdallAddress `json:"from"`
	ID       types.DelegatorID     `json:"id"`
	TxHash   types.HeimdallHash    `json:"tx_hash"`
	LogIndex uint64                `json:"log_index"`
}

// NewMsgDelegatorJoin creates new delegator-join
func NewMsgDelegatorJoin(
	from types.HeimdallAddress,
	id uint64,
	txhash types.HeimdallHash,
	logIndex uint64,
) MsgDelegatorJoin {

	return MsgDelegatorJoin{
		From:     from,
		ID:       types.NewDelegatorID(id),
		TxHash:   txhash,
		LogIndex: logIndex,
	}
}

func (msg MsgDelegatorJoin) Type() string {
	return "delegator-join"
}

func (msg MsgDelegatorJoin) Route() string {
	return delegationTypes.RouterKey
}

func (msg MsgDelegatorJoin) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{types.HeimdallAddressToAccAddress(msg.From)}
}

func (msg MsgDelegatorJoin) GetSignBytes() []byte {
	b, err := cdc.MarshalJSON(msg)
	if err != nil {
		panic(err)
	}
	return sdk.MustSortJSON(b)
}

func (msg MsgDelegatorJoin) ValidateBasic() sdk.Error {
	if msg.ID <= 0 {
		return hmCommon.ErrInvalidMsg(hmCommon.DefaultCodespace, "Invalid delegator ID %v", msg.ID)
	}

	if msg.From.Empty() {
		return hmCommon.ErrInvalidMsg(hmCommon.DefaultCodespace, "Invalid proposer %v", msg.From.String())
	}

	return nil
}

// //
// // Delegator stake update
// //

// var _ sdk.Msg = &MsgDelegatorStakeUpdate{}

// type MsgDelegatorStakeUpdate struct {
// 	From     types.HeimdallAddress `json:"from"`
// 	ID       types.DelegatorID     `json:"id"`
// 	TxHash   types.HeimdallHash    `json:"tx_hash"`
// 	LogIndex uint64                `json:"log_index"`
// }

// // NewMsgDelegatorStakeUpdate creates new delegator-stake-update
// func NewMsgDelegatorStakeUpdate(
// 	id uint64,
// 	txhash types.HeimdallHash,
// 	logIndex uint64,
// ) MsgDelegatorStakeUpdate {

// 	return MsgDelegatorStakeUpdate{
// 		ID:       types.NewDelegatorID(id),
// 		TxHash:   txhash,
// 		LogIndex: logIndex,
// 	}
// }

// func (msg MsgDelegatorStakeUpdate) Type() string {
// 	return "delegator-stake-update"
// }

// func (msg MsgDelegatorStakeUpdate) Route() string {
// 	return delegationTypes.RouterKey
// }

// //
// // Delegator Join
// //

// var _ sdk.Msg = &MsgDelegatorUnstake{}

// type MsgDelegatorUnstake struct {
// 	From     types.HeimdallAddress `json:"from"`
// 	ID       types.DelegatorID     `json:"id"`
// 	TxHash   types.HeimdallHash    `json:"tx_hash"`
// 	LogIndex uint64                `json:"log_index"`
// }

// // NewMsgDelegatorUnstake creates new delegator-unstake
// func NewMsgDelegatorUnstake(
// 	id uint64,
// 	txhash types.HeimdallHash,
// 	logIndex uint64,
// ) MsgDelegatorUnstake {

// 	return MsgDelegatorUnstake{
// 		ID:       types.NewDelegatorID(id),
// 		TxHash:   txhash,
// 		LogIndex: logIndex,
// 	}
// }

// func (msg MsgDelegatorUnstake) Type() string {
// 	return "delegator-unstake"
// }

// func (msg MsgDelegatorUnstake) Route() string {
// 	return delegationTypes.RouterKey
// }

// //
// Delegator Bond
//

var _ sdk.Msg = &MsgDelegatorBond{}

type MsgDelegatorBond struct {
	From     types.HeimdallAddress `json:"from"`
	ID       types.DelegatorID     `json:"id"`
	TxHash   types.HeimdallHash    `json:"tx_hash"`
	LogIndex uint64                `json:"log_index"`
}

// NewMsgDelegatorBond creates new delegator-bond
func NewMsgDelegatorBond(
	from types.HeimdallAddress,
	id uint64,
	txhash types.HeimdallHash,
	logIndex uint64,
) MsgDelegatorBond {

	return MsgDelegatorBond{
		From:     from,
		ID:       types.NewDelegatorID(id),
		TxHash:   txhash,
		LogIndex: logIndex,
	}
}

func (msg MsgDelegatorBond) Type() string {
	return "delegator-bond"
}

func (msg MsgDelegatorBond) Route() string {
	return delegationTypes.RouterKey
}

func (msg MsgDelegatorBond) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{types.HeimdallAddressToAccAddress(msg.From)}
}

func (msg MsgDelegatorBond) GetSignBytes() []byte {
	b, err := cdc.MarshalJSON(msg)
	if err != nil {
		panic(err)
	}
	return sdk.MustSortJSON(b)
}

func (msg MsgDelegatorBond) ValidateBasic() sdk.Error {
	if msg.ID <= 0 {
		return hmCommon.ErrInvalidMsg(hmCommon.DefaultCodespace, "Invalid delegator ID %v", msg.ID)
	}

	if msg.From.Empty() {
		return hmCommon.ErrInvalidMsg(hmCommon.DefaultCodespace, "Invalid proposer %v", msg.From.String())
	}

	return nil
}

//
// Delegator Unbond
//

var _ sdk.Msg = &MsgDelegatorUnBond{}

type MsgDelegatorUnBond struct {
	From     types.HeimdallAddress `json:"from"`
	ID       types.DelegatorID     `json:"id"`
	TxHash   types.HeimdallHash    `json:"tx_hash"`
	LogIndex uint64                `json:"log_index"`
}

// NewMsgDelegatorUnBond creates new delegator-unbond
func NewMsgDelegatorUnBond(
	from types.HeimdallAddress,
	id uint64,
	txhash types.HeimdallHash,
	logIndex uint64,
) MsgDelegatorUnBond {

	return MsgDelegatorUnBond{
		From:     from,
		ID:       types.NewDelegatorID(id),
		TxHash:   txhash,
		LogIndex: logIndex,
	}
}

func (msg MsgDelegatorUnBond) Type() string {
	return "delegator-unbond"
}

func (msg MsgDelegatorUnBond) Route() string {
	return delegationTypes.RouterKey
}

func (msg MsgDelegatorUnBond) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{types.HeimdallAddressToAccAddress(msg.From)}
}

func (msg MsgDelegatorUnBond) GetSignBytes() []byte {
	b, err := cdc.MarshalJSON(msg)
	if err != nil {
		panic(err)
	}
	return sdk.MustSortJSON(b)
}

func (msg MsgDelegatorUnBond) ValidateBasic() sdk.Error {
	if msg.ID <= 0 {
		return hmCommon.ErrInvalidMsg(hmCommon.DefaultCodespace, "Invalid delegator ID %v", msg.ID)
	}

	if msg.From.Empty() {
		return hmCommon.ErrInvalidMsg(hmCommon.DefaultCodespace, "Invalid proposer %v", msg.From.String())
	}

	return nil
}

// //
// // Delegator ReBond
// //

// var _ sdk.Msg = &MsgDelegatorRebond{}

// type MsgDelegatorRebond struct {
// 	From     types.HeimdallAddress `json:"from"`
// 	ID       types.DelegatorID     `json:"id"`
// 	TxHash   types.HeimdallHash    `json:"tx_hash"`
// 	LogIndex uint64                `json:"log_index"`
// }

// // NewMsgDelegatorRebond creates new delegator-restake
// func NewMsgDelegatorRebond(
// 	id uint64,
// 	txhash types.HeimdallHash,
// 	logIndex uint64,
// ) MsgDelegatorRebond {

// 	return MsgDelegatorRebond{
// 		ID:       types.NewDelegatorID(id),
// 		TxHash:   txhash,
// 		LogIndex: logIndex,
// 	}
// }

// func (msg MsgDelegatorRebond) Type() string {
// 	return "delegator-rebond"
// }

// func (msg MsgDelegatorRebond) Route() string {
// 	return delegationTypes.RouterKey
// }
