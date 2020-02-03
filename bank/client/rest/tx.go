package rest

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"

	bankTypes "github.com/maticnetwork/heimdall/bank/types"
	restClient "github.com/maticnetwork/heimdall/client/rest"
	"github.com/maticnetwork/heimdall/types"
	"github.com/maticnetwork/heimdall/types/rest"
)

// RegisterRoutes - Central function to define routes that get registered by the main application
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc("/bank/accounts/{address}/transfers", SendRequestHandlerFn(cliCtx)).Methods("POST")
	r.HandleFunc("/bank/accounts/topup", TopupHandlerFn(cliCtx)).Methods("POST")
	r.HandleFunc("/bank/balances/{address}", QueryBalancesRequestHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc("/bank/accounts/fee/withdraw", WithdrawFeeHandlerFn(cliCtx)).Methods("POST")
}

// SendReq defines the properties of a send request's body.
type SendReq struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`

	Amount types.Coins `json:"amount" yaml:"amount"`
}

// SendRequestHandlerFn - http request handler to send coins to a address.
func SendRequestHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		// get to address
		toAddr := types.HexToHeimdallAddress(vars["address"])

		var req SendReq
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		// get from address
		fromAddr := types.HexToHeimdallAddress(req.BaseReq.From)

		msg := bankTypes.NewMsgSend(fromAddr, toAddr, req.Amount)
		restClient.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}

//
// Topup req
//

// TopupReq defines the properties of a topup request's body.
type TopupReq struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`

	ID       uint64 `json:"id" yaml:"id"`
	TxHash   string `json:"tx_hash" yaml:"tx_hash"`
	LogIndex uint64 `json:"log_index" yaml:"log_index"`
}

// TopupHandlerFn - http request handler to topup coins to a address.
func TopupHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req TopupReq
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		// get from address
		fromAddr := types.HexToHeimdallAddress(req.BaseReq.From)

		// get msg
		msg := bankTypes.NewMsgTopup(
			fromAddr,
			req.ID,
			types.HexToHeimdallHash(req.TxHash),
			req.LogIndex,
		)
		restClient.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}

//
// Withdraw Fee req
//

// WithdrawFeeReq defines the properties of a withdraw fee request's body.
type WithdrawFeeReq struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`

	ID             uint64 `json:"id" yaml:"id"`
	WithdrawAmount string `json:"withdraw_amount" yaml:"withdraw_amount"`
}

// WithdrawFeeHandlerFn - http request handler to withdraw fee coins from a address.
func WithdrawFeeHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req WithdrawFeeReq
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		// get from address & withdraw amount
		fromAddr := types.HexToHeimdallAddress(req.BaseReq.From)
		withdrawAmt, _ := types.NewIntFromString(req.WithdrawAmount)

		// get msg
		msg := bankTypes.NewMsgWithdrawFee(
			fromAddr,
			req.ID,
			withdrawAmt,
		)
		restClient.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}
