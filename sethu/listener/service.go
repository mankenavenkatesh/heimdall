package listener

import (
	"bytes"
	"log"

	"github.com/cosmos/cosmos-sdk/client"
	cliContext "github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/maticnetwork/bor/accounts/abi"
	"github.com/spf13/viper"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/tendermint/tendermint/libs/common"
	httpClient "github.com/tendermint/tendermint/rpc/client"
)


// ListenerService starts and stops all chain event listeners
type ListenerService struct {
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

	listeners []Listener
}

// NewListenerService returns new service object for listneing to events
func NewListenerService(cdc *codec.Codec, queueConnector *QueueConnector, httpClient *httpClient.HTTP) *ListenerService {
	// create logger
	logger := Logger.With("module", ListenerService)

	cliCtx := cliContext.NewCLIContext().WithCodec(cdc)
	cliCtx.BroadcastMode = client.BroadcastAsync
	cliCtx.TrustNode = true

	// creating listener object
	listenerService := &ListenerService{
		storageClient: getBridgeDBInstance(viper.GetString(BridgeDBFlag)),

		cliCtx:         cliCtx,
		queueConnector: queueConnector,
		httpClient:     httpClient,
	}

	listenerService.BaseService = *common.NewBaseService(logger, ListenerService, listenerService)
	return listener
}

// OnStart starts new block subscription
func (listenerService *ListenerService) OnStart() error {
	listenerService.BaseService.OnStart() // Always call the overridden method.

	// start chain listeners
	for listener := range listenerService.listeners {
		listener.Start()
	}

	return nil
}

// OnStop stops all necessary go routines
func (listenerService *ListenerService) OnStop() {
	listenerService.BaseService.OnStop() // Always call the overridden method.

	// close db
	closeBridgeDBInstance()

	// stop chain listeners
	for listener := range listenerService.listeners {
		listener.Stop()
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
