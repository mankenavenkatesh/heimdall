package types

import (
	"fmt"
	"sort"

	"github.com/cosmos/cosmos-sdk/codec"
)

// DelegatorAccount contains Rewards, Shares, Slashed Amount
type DelegatorAccount struct {
	ID            DelegatorID `json:"ID"`
	Shares        float32     `json:"shares"`
	RewardAmount  string      `json:"rewardAmount"`
	SlashedAmount string      `json:"slashedAmount"`
}

func (da *DelegatorAccount) String() string {
	if da == nil {
		return "nil-DelegatorAccount"
	}

	return fmt.Sprintf("DelegatorAccount{%v %v %v %v}",
		da.ID,
		da.Shares,
		da.RewardAmount,
		da.SlashedAmount)
}

// MarshallDelegatorAccount - amino Marshall DelegatorAccount
func MarshallDelegatorAccount(cdc *codec.Codec, delegatorAccount DelegatorAccount) (bz []byte, err error) {
	bz, err = cdc.MarshalBinaryBare(delegatorAccount)
	if err != nil {
		return bz, err
	}

	return bz, nil
}

// UnMarshallDelegatorAccount - amino Unmarshall DelegatorAccount
func UnMarshallDelegatorAccount(cdc *codec.Codec, value []byte) (DelegatorAccount, error) {

	var delegatorAccount DelegatorAccount
	err := cdc.UnmarshalBinaryBare(value, &delegatorAccount)
	if err != nil {
		return delegatorAccount, err
	}
	return delegatorAccount, nil
}

// SortDelegatorAccountByID - Sorts Delegator Accounts By Delegator ID
func SortDelegatorAccountByID(delegatorAccounts []DelegatorAccount) []DelegatorAccount {
	sort.Slice(delegatorAccounts, func(i, j int) bool { return delegatorAccounts[i].ID < delegatorAccounts[j].ID })
	return delegatorAccounts
}
