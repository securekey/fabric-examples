# Fabric CLI Sample Application

Fabric CLI is a sample command-line interface built using Fabric SDK GO. It provides the following functionality:

Channel:

- Create - Create a channel
- Join - Join a peer to a channel

Query:

- Info - Displays information about the block, including the block height
- Block - Displays the contents of a given block
- Tx - Displays the contents of a transaction within a block
- Channels - Displays all channels for a peer
- Installed - Displays the chaincodes installed on a peer

Chaincode:

- Install - Installs chaincode
- Instantiate - Instantiates chaincode
- Upgrade - Upgrades chaincode
- Invoke - Invokes chaincode with a transaction proposal and a commit
- Query - Invokes chaincode without a commit
- Info - Retrieves details about the chaincode

Events:

- Listen Block - Listens for block events and displays the block when a new block is created
- Listen Filtered Block - Listens for filtered block events and displays the filtered block when a new block is created
- Listen TX - Listens for transaction events
- Listen CC - Listens for chaincode events for a specified chaincode

## Running

Navigate to folder cmd/fabric-cli. (If you don't have dep installed then you can get here: <https://github.com/golang/dep>.)

Populate generated files (not included in git):

```bash
make populate
```

Run the client:

```bash
go run fabric-cli.go <command> <sub-command> [options]
```

To display the available commands/options:

```bash
go run fabric-cli.go
```

## Compatability

This example is compatible with the following Hyperledger Fabric/SDK commit levels:

- fabric: v1.1.0, v1.2.0
- fabric-sdk-go: master:37201e914412e03b048f398ba3cdbf101515f9a3

### Quick Tour

Start the example HLF network locally

```bash
make example-network
```

Open another shell and go to the CLI directory

```bash
cd $GOPATH/src/github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli
```

Create channel 'mychannel'

```bash
go run fabric-cli.go channel create --cid mychannel --txfile ../../fabric-sdk-go/test/fixtures/fabric/v1.1/channel/mychannel.tx --config ../../test/fixtures/config/config_test_local.yaml
```

Join all peers to channel 'mychannel'

```bash
go run fabric-cli.go channel join --cid mychannel --config ../../test/fixtures/config/config_test_local.yaml
```

Install ExampleCC chaincode on all peers

```bash
go run fabric-cli.go chaincode install --ccp=github.com/securekey/example_cc --ccid=ExampleCC --v v0 --gopath ../../test/fixtures/testdata --config ../../test/fixtures/config/config_test_local.yaml
```

Instantiate ExampleCC chaincode with endorsement policy AND('Org1MSP.member','Org2MSP.member')

```bash
go run fabric-cli.go chaincode instantiate --cid mychannel --ccp=github.com/securekey/example_cc --ccid ExampleCC --v v0 --args '{"Args":["A","1","B","2"]}' --policy "AND('Org1MSP.member','Org2MSP.member')" --config ../../test/fixtures/config/config_test_local.yaml
```

Query ExampleCC chaincode on a set of peers

```bash
go run fabric-cli.go chaincode query --cid mychannel --ccid ExampleCC --args '{"Func":"query","Args":["A"]}' --peer localhost:7051,localhost:8051 --payload --config ../../test/fixtures/config/config_test_local.yaml
```

Listen for block events (output in JSON):

```bash
go run fabric-cli.go event listenblock --cid mychannel --format json --config ../../test/fixtures/config/config_test_local.yaml
```

Then invoke ExampleCC chaincode in another shell and observe block events in the first shell.

```bash
go run fabric-cli.go chaincode invoke --cid mychannel --ccid=ExampleCC --args '{"Func":"move","Args":["A","B","1"]}' --peer localhost:7051,localhost:8051 --base64 --config ../../test/fixtures/config/config_test_local.yaml
```

To clean up example network artifacts run

```bash
make example-network-clean
```

and follow instructions. This process is not yet automated.

## CLI examples

### Channel

#### Create a channel

```bash
go run fabric-cli.go channel create --cid mychannel --txfile ../../fabric-sdk-go/test/fixtures/fabric/v1.1/channel/mychannel.tx --config ../../test/fixtures/config/config_test_local.yaml
```

#### Update anchor peers

```bash
go run fabric-cli.go channel create --cid mychannel --txfile ../../fabric-sdk-go/test/fixtures/fabric/v1.1/channel/mychannelOrg1MSPanchors.tx --orgid=org1 --config ../../test/fixtures/config/config_test_local.yaml
go run fabric-cli.go channel create --cid mychannel --txfile ../../fabric-sdk-go/test/fixtures/fabric/v1.1/channel/mychannelOrg2MSPanchors.tx --orgid=org2 --config ../../test/fixtures/config/config_test_local.yaml
```

#### Join a peer to a channel

```bash
go run fabric-cli.go channel join --cid mychannel --peer localhost:7051 --config ../../test/fixtures/config/config_test_local.yaml
```

#### Join all peers in org1 to a channel

```bash
go run fabric-cli.go channel join --cid mychannel --orgid org1 --config ../../test/fixtures/config/config_test_local.yaml
```

#### Join all peers to a channel

```bash
go run fabric-cli.go channel join --cid mychannel --config ../../test/fixtures/config/config_test_local.yaml
```

### Query

#### Query info

```bash
go run fabric-cli.go query info --cid mychannel --config ../../test/fixtures/config/config_test_local.yaml
```

#### Query block by block number

```bash
go run fabric-cli.go query block --cid mychannel --num 0 --base64 --config ../../test/fixtures/config/config_test_local.yaml
```

#### Query block by hash (replace the example hash with a valid hash, e.g. using the output from query info)

```bash
go run fabric-cli.go query block --cid mychannel --hash BNNsxK_Xyz2d3Yj2g6M2t3aOYkHCxvoPeIGmTWdOJ9w --base64 --config ../../test/fixtures/config/config_test_local.yaml
```

#### Query block output in JSON format

```bash
go run fabric-cli.go query block --cid mychannel --num 0 --format json --config ../../test/fixtures/config/config_test_local.yaml
```

#### Query transaction (replace txid with a valid transaction, e.g. using the output from query block)

```bash
go run fabric-cli.go query tx --cid mychannel --txid 0ed409872e0e6a6aa745df10d5e71e33d7e160b84519c2ad89281e65b6561364 --base64 --config ../../test/fixtures/config/config_test_local.yaml
```

#### Query channels joined by a peer

```bash
go run fabric-cli.go query channels --peer localhost:7051 --config ../../test/fixtures/config/config_test_local.yaml
```

#### Query installed chaincodes on a peer

```bash
go run fabric-cli.go query installed --peer localhost:7051 --config ../../test/fixtures/config/config_test_local.yaml
```

## Chaincode

### Install Chaincode

#### Install chaincode on a peer

```bash
go run fabric-cli.go chaincode install --peer localhost:7051 --ccp=github.com/securekey/example_cc --gopath ../../test/fixtures/testdata --ccid=examplecc --v v0 --config ../../test/fixtures/config/config_test_local.yaml
```

#### Install chaincode on all peers of org1

```bash
go run fabric-cli.go chaincode install --orgid org1 --ccp=github.com/securekey/example_cc --gopath ../../test/fixtures/testdata --ccid=examplecc --v v0 --config ../../test/fixtures/config/config_test_local.yaml
```

#### Install chaincode on all peers

```bash
go run fabric-cli.go chaincode install --ccp=github.com/securekey/example_cc --gopath ../../test/fixtures/testdata --ccid=examplecc --v v0 --config ../../test/fixtures/config/config_test_local.yaml
```

### Instantiate and Upgrade Chaincode

#### Instantiate chaincode with default endorsement policy (any one)

```bash
go run fabric-cli.go chaincode instantiate --cid mychannel --ccp github.com/securekey/example_cc --ccid examplecc --v v0 --args='{"Args":["A","1","B","2"]}' --config ../../test/fixtures/config/config_test_local.yaml
```

#### Instantiate chaincode with specified endorsement policy

```bash
go run fabric-cli.go chaincode instantiate --cid mychannel --ccp github.com/securekey/example_cc --ccid examplecc --v v0 --args='{"Args":["A","1","B","2"]}' --policy "AND('Org1MSP.member','Org2MSP.member')" --config ../../test/fixtures/config/config_test_local.yaml
```

#### Instantiate chaincode with specified private data collection configuration (Fabric 1.2 and greater)

```bash
go run fabric-cli.go chaincode instantiate --cid mychannel --ccp github.com/securekey/example2_cc --ccid example2cc --v v0 --policy "AND('Org1MSP.member','Org2MSP.member')" --collconfig ../../test/fixtures/config/pvtdatacollection.json --config ../../test/fixtures/config/config_test_local.yaml
```

#### Upgrade chaincode

```bash
go run fabric-cli.go chaincode install --cid=mychannel --ccp=github.com/securekey/example_cc --gopath ../../test/fixtures/testdata --ccid=examplecc --v v1 --config ../../test/fixtures/config/config_test_local.yaml
go run fabric-cli.go chaincode upgrade --cid mychannel --ccp github.com/securekey/example_cc --ccid examplecc --v v1 --args='{"Args":["A","1","B","2"]}' --policy "OutOf(2,'Org1MSP.member','Org2MSP.member')" --config ../../test/fixtures/config/config_test_local.yaml
```

### Chaincode Info

#### Retrieve chaincode deployment info

```bash
go run fabric-cli.go chaincode info --cid mychannel --ccid examplecc --base64 --config ../../test/fixtures/config/config_test_local.yaml
```

### Query Chaincode

#### Query chaincode on a set of peers

```bash
go run fabric-cli.go chaincode query --cid mychannel --ccid=examplecc --args='{"Func":"query","Args":["A"]}' --peer localhost:7051,localhost:8051 --base64 --config ../../test/fixtures/config/config_test_local.yaml
```

#### Query chaincode and view payloads only

```bash
go run fabric-cli.go chaincode query --cid mychannel --ccid=examplecc --args='{"Func":"query","Args":["A"]}' --peer localhost:7051,localhost:8051 --payload --config ../../test/fixtures/config/config_test_local.yaml
```

### Invoke Chaincode

#### Invoke chaincode on all peers in org1

```bash
go run fabric-cli.go chaincode invoke --cid mychannel --ccid=examplecc --args='{"Func":"move","Args":["A","B","1"]}' --orgid org1 --base64 --config ../../test/fixtures/config/config_test_local.yaml
```

#### Invoke chaincode using a 'dynamic' selection provider that chooses a minimal set of peers required to satisfy the endorsement policy of the chaincode

```bash
go run fabric-cli.go chaincode invoke --cid mychannel --ccid=examplecc --args='{"Func":"move","Args":["A","B","1"]}' --selectprovider=dynamic --base64 --config ../../test/fixtures/config/config_test_local.yaml
```

#### Invoke chaincode using Fabric's discovery service to choose a minimal set of peers required to satisfy the endorsement policy of the chaincode (Requires Fabric 1.2)

```bash
go run fabric-cli.go chaincode invoke --cid mychannel --ccid=examplecc --args='{"Func":"move","Args":["A","B","1"]}' --selectprovider=fabric --base64 --config ../../test/fixtures/config/config_test_local.yaml
```

#### Invoke chaincode using a selection provider automatically determined from channel capabilities ('dynamic' for v1.1; 'fabric' for v1.2)

```bash
go run fabric-cli.go chaincode invoke --cid mychannel --ccid=examplecc --args='{"Func":"move","Args":["A","B","1"]}' --base64 --config ../../test/fixtures/config/config_test_local.yaml
```

#### Invoke chaincode 5 times

```bash
go run fabric-cli.go chaincode invoke --cid mychannel --ccid=examplecc --args='{"Func":"move","Args":["A","B","1"]}' --iterations 5 --config ../../test/fixtures/config/config_test_local.yaml
```

#### Invoke chaincode 100 times in 8 Go routines with a maximum of 5 attempts for each invocation (in case the invocation fails)

```bash
go run fabric-cli.go chaincode invoke --cid mychannel --ccid=examplecc --args='{"Func":"move","Args":["A","B","1"]}' --iterations 100 --concurrency 8 --attempts 5 --backoff 1000 --backofffactor 1.5 --maxbackoff 5000 --config ../../test/fixtures/config/config_test_local.yaml
```

#### Invoke chaincode with two sets of args, 100 times each in 8 Go routines with 3 attempts for each invocation (in case the invocation fails)

```bash
go run fabric-cli.go chaincode invoke --cid mychannel --ccid=examplecc --args='[{"Func":"move","Args":["A","B","1"]},{"Func":"move","Args":["B","A","1"]}]' --iterations 100 --concurrency 8 --attempts=3 --config ../../test/fixtures/config/config_test_local.yaml
```

#### Invoke chaincode 100 times in 8 Go routines using randomly generated keys and values

```bash
go run fabric-cli.go chaincode invoke --cid mychannel --ccid=example2cc --args='{"Func":"putprivate","Args":["coll1","Key_$rand(500)","Val_$rand(1000)"]}' --iterations 100 --concurrency 8 --config ../../test/fixtures/config/config_test_local.yaml
```

## Event

### Block Events

### Listen for block events on a specific peer (output in JSON)

```bash
go run fabric-cli.go event listenblock --cid mychannel --peer localhost:7051 --format json --config ../../test/fixtures/config/config_test_local.yaml
```

### Listen for block events starting from block number 20

```bash
go run fabric-cli.go event listenblock --cid mychannel --peer localhost:7051 --seek from --num 20 --base64 --config ../../test/fixtures/config/config_test_local.yaml
```

### Listen for filtered block events (output in JSON)

```bash
go run fabric-cli.go event listenfilteredblock --cid mychannel --format json --config ../../test/fixtures/config/config_test_local.yaml
```

### Listen for chaincode events

```bash
go run fabric-cli.go event listencc --cid mychannel --ccid=examplecc --event=.* --config ../../test/fixtures/config/config_test_local.yaml
```

### Listen for tx event

```bash
go run fabric-cli.go event listentx --cid mychannel --txid <txid> --config ../../test/fixtures/config/config_test_local.yaml
```
