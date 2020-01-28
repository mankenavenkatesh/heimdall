package processor

type BaseProcessor struct {
	preProcess()  // called upon event from Rootchain/Matic
	postProcess() // called upon event from heimdall

	// Validations
	isProposer() // helper method
	isValid()    // helper method

	// Conversions
	DecodeEvent()       // helper method
	createMsg()         // helper method
	createTransaction() // helper method

	// Broadcasts
	BroadcastToHeimdall() // helper method
	BroadcastToMatic()    // helper method
	BradcastToRootchain() // helper method
}
