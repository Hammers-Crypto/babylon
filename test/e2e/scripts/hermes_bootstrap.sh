#!/bin/bash

set -ex

# initialize Hermes relayer configuration
mkdir -p /root/.hermes/
touch /root/.hermes/config.toml

echo $BBN_A_E2E_VAL_MNEMONIC > /root/.hermes/BBN_A_MNEMONIC.txt
echo $BBN_B_E2E_VAL_MNEMONIC > /root/.hermes/BBN_B_MNEMONIC.txt
# setup Hermes relayer configuration
tee /root/.hermes/config.toml <<EOF
[global]
log_level = 'debug'
[mode]
[mode.clients]
enabled = true
refresh = true
misbehaviour = true
[mode.connections]
enabled = false
[mode.channels]
enabled = true
[mode.packets]
enabled = true
clear_interval = 100
clear_on_start = true
tx_confirmation = true
[rest]
enabled = true
host = '0.0.0.0'
port = 3031
[telemetry]
enabled = true
host = '127.0.0.1'
port = 3001
[[chains]]
type = "CosmosSdk"
id = '$BBN_A_E2E_CHAIN_ID'
rpc_addr = 'http://$BBN_A_E2E_VAL_HOST:26657'
grpc_addr = 'http://$BBN_A_E2E_VAL_HOST:9090'
event_source = { mode = 'push', url = 'ws://$BBN_A_E2E_VAL_HOST:26657/websocket', batch_delay = '500ms' }
rpc_timeout = '10s'
account_prefix = 'bbn'
key_name = 'val01-bbn-a'
store_prefix = 'ibc'
max_gas = 50000000
gas_price = { price = 0.01, denom = 'ubbn' }
gas_multiplier = 1.5
clock_drift = '1m' # to accomdate docker containers
trusting_period = '14days'
trust_threshold = { numerator = '1', denominator = '3' }
[[chains]]
type = "CosmosSdk"
id = '$BBN_B_E2E_CHAIN_ID'
rpc_addr = 'http://$BBN_B_E2E_VAL_HOST:26657'
grpc_addr = 'http://$BBN_B_E2E_VAL_HOST:9090'
event_source = { mode = 'push', url = 'ws://$BBN_B_E2E_VAL_HOST:26657/websocket', batch_delay = '500ms' }
rpc_timeout = '10s'
account_prefix = 'bbn'
key_name = 'val01-bbn-b'
store_prefix = 'ibc'
max_gas = 50000000
gas_price = { price = 0.01, denom = 'ubbn' }
gas_multiplier = 1.5
clock_drift = '1m' # to accomdate docker containers
trusting_period = '14days'
trust_threshold = { numerator = '1', denominator = '3' }
EOF

# import keys
hermes keys add --chain ${BBN_A_E2E_CHAIN_ID} --key-name "val01-bbn-a" --mnemonic-file /root/.hermes/BBN_A_MNEMONIC.txt
hermes keys add --chain ${BBN_B_E2E_CHAIN_ID} --key-name "val01-bbn-b" --mnemonic-file /root/.hermes/BBN_B_MNEMONIC.txt

# start Hermes relayer
hermes start
