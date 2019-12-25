package helper

import (
	"bytes"
	"text/template"

	"github.com/spf13/viper"
	cmn "github.com/tendermint/tendermint/libs/common"
)

// Note: any changes to the comments/variables/mapstructure
// must be reflected in the appropriate struct in helper/config.go
const defaultConfigTemplate = `# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml

##### RPC configrations #####

# RPC endpoint for ethereum chain
eth_RPC_URL = "{{ .EthRPCUrl }}"

# RPC endpoint for bor chain
bor_RPC_URL = "{{ .BorRPCUrl }}"


# RPC endpoint for tendermint
tendermint_RPC_URL = "{{ .TendermintRPCUrl }}"

##### Chain ID configration #####

# Bor chain ID
bor_chain_id = "{{ .BorChainID }}" 


##### MQTT and Rest Server Config #####

# MQTT endpoint
amqp_url = "{{ .AmqpURL }}" 

# Heimdall REST server endpoint
heimdall_rest_server = "{{ .HeimdallServerURL }}" 

##### Contract Addresses #####

### Eth Chain Contracts 
stakemanager_contract = "{{ .StakeManagerAddress }}" 
rootchain_contract = "{{ .RootchainAddress }}" 
delegationmanager_contract = "{{ .DelegationManagerAddress }}" 
state_sender_contract = "{{ .StateSenderAddress }}" 

### Bor Chain Contracts
state_receiver_contract = "{{ .StateReceiverAddress }}" 
validator_set_contract = "{{ .ValidatorSetAddress }}" 


##### Intervals #####
child_chain_block_interval = "{{ .ChildBlockInterval }}" 

## Bridge Poll Intervals 
checkpoint_poll_interval = "{{ .CheckpointerPollInterval }}" 
syncer_poll_interval = "{{ .SyncerPollInterval }}"
noack_poll_interval = "{{ .NoACKPollInterval }}"


##### Checkpoint Length Config #####
avg_checkpoint_length = "{{ .AvgCheckpointLength }}"
max_checkpoint_length = "{{ .MaxCheckpointLength }}"

##### Timeout Config #####

no_ack_wait_time = "{{ .NoACKWaitTime }}"
checkpoint_buffer_time = "{{ .CheckpointBufferTime }}"

##### Transaction Confirmations  #####

confirmation_blocks = "{{ .ConfirmationBlocks }}"

`

var configTemplate *template.Template

func init() {
	var err error
	tmpl := template.New("appConfigFileTemplate")
	if configTemplate, err = tmpl.Parse(defaultConfigTemplate); err != nil {
		panic(err)
	}
}

// ParseConfig retrieves the default environment configuration for the
// application.
func ParseConfig() (*Configuration, error) {
	conf := GetDefaultHeimdallConfig()
	err := viper.Unmarshal(conf)
	return &conf, err
}

// WriteConfigFile renders config using the template and writes it to
// configFilePath.
func WriteConfigFile(configFilePath string, config *Configuration) {
	var buffer bytes.Buffer

	if err := configTemplate.Execute(&buffer, config); err != nil {
		panic(err)
	}

	cmn.MustWriteFile(configFilePath, buffer.Bytes(), 0644)
}
