package helper

import (
	"context"
	"errors"
	"math/big"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/maticnetwork/heimdall/contracts/delegationmanager"
	"github.com/maticnetwork/heimdall/contracts/rootchain"
	"github.com/maticnetwork/heimdall/contracts/stakemanager"
	"github.com/maticnetwork/heimdall/contracts/statereceiver"
	"github.com/maticnetwork/heimdall/contracts/statesender"
	"github.com/maticnetwork/heimdall/contracts/validatorset"
	"github.com/maticnetwork/heimdall/types"
)

// IContractCaller represents contract caller
type IContractCaller interface {
	GetHeaderInfo(headerID uint64) (root common.Hash, start, end, createdAt uint64, proposer types.HeimdallAddress, err error)
	GetValidatorInfo(valID types.ValidatorID) (validator types.Validator, err error)
	GetLastChildBlock() (uint64, error)
	CurrentHeaderBlock() (uint64, error)
	GetBalance(address common.Address) (*big.Int, error)
	SendCheckpoint(voteSignBytes []byte, sigs []byte, txData []byte)
	GetCheckpointSign(ctx sdk.Context, txHash common.Hash) ([]byte, []byte, []byte, error)
	GetMainChainBlock(*big.Int) (*ethTypes.Header, error)
	GetMaticChainBlock(*big.Int) (*ethTypes.Header, error)
	IsTxConfirmed(common.Hash) bool
	GetConfirmedTxReceipt(common.Hash) (*ethTypes.Receipt, error)
	GetBlockNumberFromTxHash(common.Hash) (*big.Int, error)
	DecodeValidatorTopupFeesEvent(*ethTypes.Receipt, uint64) (*stakemanager.StakemanagerTopupFees, error)
	DecodeValidatorStakeUpdateEvent(*ethTypes.Receipt, uint64) (*stakemanager.StakemanagerStakeUpdate, error)
	DecodeNewHeaderBlockEvent(*ethTypes.Receipt, uint64) (*rootchain.RootchainNewHeaderBlock, error)
	DecodeSignerUpdateEvent(*ethTypes.Receipt, uint64) (*stakemanager.StakemanagerSignerChange, error)
	GetMainTxReceipt(common.Hash) (*ethTypes.Receipt, error)
	GetMaticTxReceipt(common.Hash) (*ethTypes.Receipt, error)

	// bor related contracts
	CurrentSpanNumber() (Number *big.Int)
	GetSpanDetails(id *big.Int) (*big.Int, *big.Int, *big.Int, error)
	CurrentStateCounter() (Number *big.Int)
	EncodeStateSyncedEvent(*ethTypes.Log) (*statesender.StatesenderStateSynced, error)
}

// ContractCaller contract caller
type ContractCaller struct {
	MainChainClient  *ethclient.Client
	MainChainRPC     *rpc.Client
	MaticChainClient *ethclient.Client

	RootChainInstance         *rootchain.Rootchain
	StakeManagerInstance      *stakemanager.Stakemanager
	DelegationManagerInstance *delegationmanager.Delegationmanager
	ValidatorSetInstance      *validatorset.Validatorset
	StateSenderInstance       *statesender.Statesender
	StateReceiverInstance     *statereceiver.Statereceiver

	RootChainABI         abi.ABI
	StakeManagerABI      abi.ABI
	DelegationManagerABI abi.ABI
	ValidatorSetABI      abi.ABI
	StateReceiverABI     abi.ABI
	StateSenderABI       abi.ABI
}

type txExtraInfo struct {
	BlockNumber *string         `json:"blockNumber,omitempty"`
	BlockHash   *common.Hash    `json:"blockHash,omitempty"`
	From        *common.Address `json:"from,omitempty"`
}

type rpcTransaction struct {
	tx *ethTypes.Transaction
	txExtraInfo
}

// NewContractCaller contract caller
func NewContractCaller() (contractCallerObj ContractCaller, err error) {
	contractCallerObj.MainChainClient = GetMainClient()
	contractCallerObj.MaticChainClient = GetMaticClient()
	contractCallerObj.MainChainRPC = GetMainChainRPCClient()

	//
	// Root chain instance
	//

	if contractCallerObj.RootChainInstance, err = rootchain.NewRootchain(GetRootChainAddress(), contractCallerObj.MainChainClient); err != nil {
		return
	}

	if contractCallerObj.DelegationManagerInstance, err = delegationmanager.NewDelegationmanager((GetDelegationManagerAddress()), contractCallerObj.MainChainClient); err != nil {
		return
	}

	if contractCallerObj.StakeManagerInstance, err = stakemanager.NewStakemanager(GetStakeManagerAddress(), contractCallerObj.MainChainClient); err != nil {
		return
	}

	if contractCallerObj.ValidatorSetInstance, err = validatorset.NewValidatorset(GetValidatorSetAddress(), contractCallerObj.MaticChainClient); err != nil {
		return
	}

	if contractCallerObj.StateSenderInstance, err = statesender.NewStatesender(GetStateSenderAddress(), contractCallerObj.MainChainClient); err != nil {
		return
	}

	if contractCallerObj.StateReceiverInstance, err = statereceiver.NewStatereceiver(GetStateReceiverAddress(), contractCallerObj.MaticChainClient); err != nil {
		return
	}

	//
	// ABIs
	//

	if contractCallerObj.RootChainABI, err = getABI(string(rootchain.RootchainABI)); err != nil {
		return
	}

	if contractCallerObj.StakeManagerABI, err = getABI(string(stakemanager.StakemanagerABI)); err != nil {
		return
	}

	if contractCallerObj.DelegationManagerABI, err = getABI(string(delegationmanager.DelegationmanagerABI)); err != nil {
		return
	}

	if contractCallerObj.ValidatorSetABI, err = getABI(string(validatorset.ValidatorsetABI)); err != nil {
		return
	}

	if contractCallerObj.StateReceiverABI, err = getABI(string(statereceiver.StatereceiverABI)); err != nil {
		return
	}

	if contractCallerObj.StateSenderABI, err = getABI(string(statesender.StatesenderABI)); err != nil {
		return
	}

	return
}

// GetHeaderInfo get header info from header id
func (c *ContractCaller) GetHeaderInfo(headerID uint64) (
	root common.Hash,
	start uint64,
	end uint64,
	createdAt uint64,
	proposer types.HeimdallAddress,
	err error,
) {
	// get header from rootchain
	headerBlock, err := c.RootChainInstance.HeaderBlocks(nil, big.NewInt(0).SetUint64(headerID))
	if err != nil {
		Logger.Error("Unable to fetch header block from rootchain", "headerBlockIndex", headerID)
		return root, start, end, createdAt, proposer, errors.New("Unable to fetch header block")
	}

	return headerBlock.Root,
		headerBlock.Start.Uint64(),
		headerBlock.End.Uint64(),
		headerBlock.CreatedAt.Uint64(),
		types.BytesToHeimdallAddress(headerBlock.Proposer.Bytes()),
		nil
}

// GetLastChildBlock fetch current child block
func (c *ContractCaller) GetLastChildBlock() (uint64, error) {
	GetLastChildBlock, err := c.RootChainInstance.GetLastChildBlock(nil)
	if err != nil {
		Logger.Error("Could not fetch current child block from rootchain contract", "Error", err)
		return 0, err
	}
	return GetLastChildBlock.Uint64(), nil
}

// CurrentHeaderBlock fetches current header block
func (c *ContractCaller) CurrentHeaderBlock() (uint64, error) {
	currentHeaderBlock, err := c.RootChainInstance.CurrentHeaderBlock(nil)
	if err != nil {
		Logger.Error("Could not fetch current header block from rootchain contract", "Error", err)
		return 0, err
	}
	return currentHeaderBlock.Uint64(), nil
}

// GetBalance get balance of account (returns big.Int balance wont fit in uint64)
func (c *ContractCaller) GetBalance(address common.Address) (*big.Int, error) {
	balance, err := c.MainChainClient.BalanceAt(context.Background(), address, nil)
	if err != nil {
		Logger.Error("Unable to fetch balance of account from root chain", "Error", err, "Address", address.String())
		return big.NewInt(0), err
	}

	return balance, nil
}

// GetValidatorInfo get validator info
func (c *ContractCaller) GetValidatorInfo(valID types.ValidatorID) (validator types.Validator, err error) {
	amount, startEpoch, endEpoch, signer, status, err := c.StakeManagerInstance.GetStakerDetails(nil, big.NewInt(int64(valID)))
	if err != nil {
		Logger.Error("Error fetching validator information from stake manager", "error", err, "validatorId", valID, "status", status)
		return
	}

	newAmount, err := GetPowerFromAmount(amount)
	if err != nil {
		return
	}

	// newAmount
	validator = types.Validator{
		ID:          valID,
		VotingPower: newAmount.Int64(),
		StartEpoch:  startEpoch.Uint64(),
		EndEpoch:    endEpoch.Uint64(),
		Signer:      types.BytesToHeimdallAddress(signer.Bytes()),
	}

	return validator, nil
}

// get main chain block header
func (c *ContractCaller) GetMainChainBlock(blockNum *big.Int) (header *ethTypes.Header, err error) {
	latestBlock, err := c.MainChainClient.HeaderByNumber(context.Background(), blockNum)
	if err != nil {
		Logger.Error("Unable to connect to main chain", "Error", err)
		return
	}
	return latestBlock, nil
}

// get child chain block header
func (c *ContractCaller) GetMaticChainBlock(blockNum *big.Int) (header *ethTypes.Header, err error) {
	latestBlock, err := c.MaticChainClient.HeaderByNumber(context.Background(), blockNum)
	if err != nil {
		Logger.Error("Unable to connect to matic chain", "Error", err)
		return
	}
	return latestBlock, nil
}

// GetBlockNumberFromTxHash gets block number of transaction
func (c *ContractCaller) GetBlockNumberFromTxHash(tx common.Hash) (*big.Int, error) {
	var rpcTx rpcTransaction
	if err := c.MainChainRPC.CallContext(context.Background(), &rpcTx, "eth_getTransactionByHash", tx); err != nil {
		return nil, err
	}

	if rpcTx.BlockNumber == nil {
		return nil, errors.New("No tx found")
	}

	blkNum := big.NewInt(0)
	blkNum, ok := blkNum.SetString(*rpcTx.BlockNumber, 0)
	if !ok {
		return nil, errors.New("unable to set string")
	}
	return blkNum, nil
}

// IsTxConfirmed is tx confirmed
func (c *ContractCaller) IsTxConfirmed(tx common.Hash) bool {
	// get main tx receipt
	receipt, err := c.GetConfirmedTxReceipt(tx)
	if receipt == nil || err != nil {
		return false
	}

	return true
}

// GetConfirmedTxReceipt returns confirmed tx receipt
func (c *ContractCaller) GetConfirmedTxReceipt(tx common.Hash) (*ethTypes.Receipt, error) {
	// get main tx receipt
	receipt, err := c.GetMainTxReceipt(tx)
	if err != nil {
		return nil, err
	}
	Logger.Debug("Tx included in block", "block", receipt.BlockNumber.Uint64(), "tx", tx)

	// get main chain block
	latestBlk, err := c.GetMainChainBlock(nil)
	if err != nil {
		Logger.Error("error getting latest block from main chain", "Error", err)
		return nil, err
	}
	Logger.Debug("Latest block on main chain obtained", "Block", latestBlk.Number.Uint64())

	diff := latestBlk.Number.Uint64() - receipt.BlockNumber.Uint64()
	if diff < GetConfig().ConfirmationBlocks {
		return nil, errors.New("Not enough confirmations")
	}

	return receipt, nil
}

// DecodeValidatorTopupFeesEvent represents topup for fees tokens
func (c *ContractCaller) DecodeValidatorTopupFeesEvent(receipt *ethTypes.Receipt, logIndex uint64) (*stakemanager.StakemanagerTopupFees, error) {
	event := new(stakemanager.StakemanagerTopupFees)

	found := false
	for _, vLog := range receipt.Logs {
		if uint64(vLog.Index) == logIndex {
			found = true
			if err := UnpackLog(&c.StakeManagerABI, event, "TopupFees", vLog); err != nil {
				return nil, err
			}
			break
		}
	}

	if !found {
		return nil, errors.New("Event not found")
	}

	return event, nil
}

// DecodeValidatorStakeUpdateEvent represents validator stake update event
func (c *ContractCaller) DecodeValidatorStakeUpdateEvent(receipt *ethTypes.Receipt, logIndex uint64) (*stakemanager.StakemanagerStakeUpdate, error) {
	event := new(stakemanager.StakemanagerStakeUpdate)

	found := false
	for _, vLog := range receipt.Logs {
		if uint64(vLog.Index) == logIndex {
			found = true
			if err := UnpackLog(&c.StakeManagerABI, event, "StakeUpdate", vLog); err != nil {
				return nil, err
			}
			break
		}
	}

	if !found {
		return nil, errors.New("Event not found")
	}

	return event, nil
}

// DecodeNewHeaderBlockEvent represents new header block event
func (c *ContractCaller) DecodeNewHeaderBlockEvent(receipt *ethTypes.Receipt, logIndex uint64) (*rootchain.RootchainNewHeaderBlock, error) {
	event := new(rootchain.RootchainNewHeaderBlock)

	found := false
	for _, vLog := range receipt.Logs {
		if uint64(vLog.Index) == logIndex {
			found = true
			if err := UnpackLog(&c.RootChainABI, event, "NewHeaderBlock", vLog); err != nil {
				return nil, err
			}
			break
		}
	}

	if !found {
		return nil, errors.New("Event not found")
	}

	return event, nil
}

// DecodeSignerUpdateEvent represents sig update event
func (c *ContractCaller) DecodeSignerUpdateEvent(receipt *ethTypes.Receipt, logIndex uint64) (*stakemanager.StakemanagerSignerChange, error) {
	event := new(stakemanager.StakemanagerSignerChange)

	found := false
	for _, vLog := range receipt.Logs {
		if uint64(vLog.Index) == logIndex {
			found = true
			if err := UnpackLog(&c.StakeManagerABI, event, "SignerChange", vLog); err != nil {
				return nil, err
			}
			break
		}
	}

	if !found {
		return nil, errors.New("Event not found")
	}

	return event, nil
}

// CurrentSpanNumber get current span
func (c *ContractCaller) CurrentSpanNumber() (Number *big.Int) {
	result, err := c.ValidatorSetInstance.CurrentSpanNumber(nil)
	if err != nil {
		Logger.Error("Unable to get current span number", "Error", err)
		return nil
	}

	return result
}

// GetSpanDetails get span details
func (c *ContractCaller) GetSpanDetails(id *big.Int) (
	*big.Int,
	*big.Int,
	*big.Int,
	error,
) {
	d, err := c.ValidatorSetInstance.GetSpan(nil, id)
	return d.Number, d.StartBlock, d.EndBlock, err
}

// CurrentStateCounter get state counter
func (c *ContractCaller) CurrentStateCounter() (Number *big.Int) {
	result, err := c.StateSenderInstance.Counter(nil)
	if err != nil {
		Logger.Error("Unable to get current counter number", "Error", err)
		return nil
	}

	return result
}

// GetMainTxReceipt returns main tx receipt
func (c *ContractCaller) GetMainTxReceipt(txHash common.Hash) (*ethTypes.Receipt, error) {
	return c.getTxReceipt(c.MainChainClient, txHash)
}

// GetMaticTxReceipt returns matic tx receipt
func (c *ContractCaller) GetMaticTxReceipt(txHash common.Hash) (*ethTypes.Receipt, error) {
	return c.getTxReceipt(c.MaticChainClient, txHash)
}

func (c *ContractCaller) getTxReceipt(client *ethclient.Client, txHash common.Hash) (*ethTypes.Receipt, error) {
	return client.TransactionReceipt(context.Background(), txHash)
}

// EncodeStateSyncedEvent encode state sync data
func (c *ContractCaller) EncodeStateSyncedEvent(log *ethTypes.Log) (*statesender.StatesenderStateSynced, error) {
	event := new(statesender.StatesenderStateSynced)
	if err := UnpackLog(&c.StateSenderABI, event, "StateSynced", log); err != nil {
		return nil, err
	}
	return event, nil
}

//
// private abi methods
//

func getABI(data string) (abi.ABI, error) {
	return abi.JSON(strings.NewReader(data))
}

// GetCheckpointSign returns sigs input of committed checkpoint tranasction
func (c *ContractCaller) GetCheckpointSign(ctx sdk.Context, txHash common.Hash) ([]byte, []byte, []byte, error) {
	mainChainClient := GetMainClient()
	transaction, isPending, err := mainChainClient.TransactionByHash(ctx, txHash)
	if err != nil {
		Logger.Error("Error while Fetching Transaction By hash from MainChain", "error", err)
		return []byte{}, []byte{}, []byte{}, err
	} else if isPending {
		return []byte{}, []byte{}, []byte{}, errors.New("Transaction is still pending")
	}

	payload := transaction.Data()
	abi := c.RootChainABI
	return UnpackSigAndVotes(payload, abi)
}
