package processor

import (
	"bytes"

	"github.com/cosmos/cosmos-sdk/client"
	cliContext "github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/maticnetwork/bor/accounts/abi"
	"github.com/spf13/viper"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/tendermint/tendermint/libs/common"
	httpClient "github.com/tendermint/tendermint/rpc/client"
)

// ProcessorService starts and stops all event processors
type ProcessorService struct {
	// Base service
	common.BaseService

	// storage client
	storageClient *leveldb.DB

	// cli context
	cliCtx cliContext.CLIContext

	// queue connector
	queueConnector *QueueConnector

	// http client to subscribe to
	httpClient *httpClient.HTTP

	processors []Processor
}

// NewProcessorService returns new service object for listneing to events
func NewProcessorService(cdc *codec.Codec, queueConnector *QueueConnector, httpClient *httpClient.HTTP) *ProcessorService {
	// create logger
	logger := Logger.With("module", ProcessorService)

	cliCtx := cliContext.NewCLIContext().WithCodec(cdc)
	cliCtx.BroadcastMode = client.BroadcastAsync
	cliCtx.TrustNode = true

	// creating processor object
	processorService := &ProcessorService{
		storageClient: getBridgeDBInstance(viper.GetString(BridgeDBFlag)),

		cliCtx:         cliCtx,
		queueConnector: queueConnector,
		httpClient:     httpClient,
	}

	processorService.BaseService = *common.NewBaseService(logger, ProcessorService, processorService)
	return processor
}

// OnStart starts new block subscription
func (processorService *ProcessorService) OnStart() error {
	processorService.BaseService.OnStart() // Always call the overridden method.

	// start chain processors
	for processor := range processorService.processors {
		processor.Start()
	}

	return nil
}

// OnStop stops all necessary go routines
func (processorService *ProcessorService) OnStop() {
	processorService.BaseService.OnStop() // Always call the overridden method.

	// close db
	closeBridgeDBInstance()

	// stop chain processors
	for processor := range processorService.processors {
		processor.Stop()
	}

}

//
// Utils
//

// EventByID looks up a event by the topic id
func EventByID(abiObject *abi.ABI, sigdata []byte) *abi.Event {
	for _, event := range abiObject.Events {
		if bytes.Equal(event.Id().Bytes(), sigdata) {
			return &event
		}
	}
	return nil
}
