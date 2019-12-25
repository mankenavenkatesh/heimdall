package delegation

import (
	"github.com/cosmos/cosmos-sdk/codec"
	authTypes "github.com/maticnetwork/heimdall/auth/types"
)

// TODO we most likely dont need to register to amino as we are using RLP to encode

func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgDelegatorJoin{}, "delegation/MsgDelegatorJoin", nil)
}

func RegisterPulp(pulp *authTypes.Pulp) {
	pulp.RegisterConcrete(MsgDelegatorJoin{})
}

var cdcEmpty = codec.New()

func init() {
	RegisterCodec(cdcEmpty)
	codec.RegisterCrypto(cdcEmpty)
}
