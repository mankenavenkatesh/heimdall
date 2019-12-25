package types

import (
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/codec"
)

// Delegator heimdall delegator
type Delegator struct {
	ID          DelegatorID `json:"ID"`
	VotingPower int64       `json:"power"` // TODO add 10^-18 here so that we dont overflow easily
	LastUpdated uint64      `json:"last_updated"`
	ValID       ValidatorID `json:"val_id'`
}

// --------

// DelegatorID  delegator ID and helper functions
type DelegatorID uint64

// NewDelegatorID generate new delegator ID
func NewDelegatorID(id uint64) DelegatorID {
	return DelegatorID(id)
}

// Bytes get bytes of delegatorID
func (delegatorID DelegatorID) Bytes() []byte {
	return []byte(strconv.Itoa(delegatorID.Int()))
}

// Int converts delegator ID to int
func (delegatorID DelegatorID) Int() int {
	return int(delegatorID)
}

// Uint64 converts delegator ID to int
func (delegatorID DelegatorID) Uint64() uint64 {
	return uint64(delegatorID)
}

// MarshallDelegator - amino marshall delegator
func MarshallDelegator(cdc *codec.Codec, delegator Delegator) (bz []byte, err error) {
	bz, err = cdc.MarshalBinaryBare(delegator)
	if err != nil {
		return bz, err
	}
	return bz, nil
}

// UnmarshallDelegator - amono unmarshall delegator
func UnmarshallDelegator(cdc *codec.Codec, value []byte) (Delegator, error) {
	var delegator Delegator
	// unmarshall validator and return
	err := cdc.UnmarshalBinaryBare(value, &delegator)
	if err != nil {
		return delegator, err
	}
	return delegator, nil
}

// String - return string representration of delegator
func (v *Delegator) String() string {
	if v == nil {
		return "nil-Delegator"
	}
	return fmt.Sprintf("Delegator{%v %v VP:%v A:%v}",
		v.ID,
		v.VotingPower)
}
