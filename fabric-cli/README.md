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

Navigate to folder cmd/fabric-cli. (If you don't have dep installed then you can get here: https://github.com/golang/dep.)

Populate generated files (not included in git):

```
$ make populate
```

Run the client:

```
$ go run fabric-cli.go <command> <sub-command> [options]
```

To display the available commands/options:

```
$ go run fabric-cli.go
```

## Compatability

This example is compatible with the following Hyperledger Fabric/SDK commit levels:

- fabric: v1.1.0
- fabric-sdk-go: master:fda0d9a039dd3afc034e970023e51cd7690bb6d7

### Quick Tour

Start the example HLF network locally

```
$ make example-network
```

Open another shell and go to the CLI directory

```
$ cd $GOPATH/src/github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli
```

Create channel 'mychannel'

```
$ go run fabric-cli.go channel create --cid mychannel --txfile ../../fabric-sdk-go/test/fixtures/fabric/v1.1/channel/mychannel.tx --config ../../test/fixtures/config/config_test_local.yaml
```

Join all peers to channel 'mychannel'

```
$ go run fabric-cli.go channel join --cid mychannel --config ../../test/fixtures/config/config_test_local.yaml
```

Install ExampleCC chaincode on all peers

```
$ go run fabric-cli.go chaincode install --cid=mychannel --ccp=github.com/example_cc --ccid=ExampleCC --v v0 --gopath ../../test/fixtures/testgopath --config ../../test/fixtures/config/config_test_local.yaml
```

Instantiate ExampleCC chaincode with endorsement policy AND('Org1MSP.member','Org2MSP.member')

```
$ go run fabric-cli.go chaincode instantiate --cid mychannel --ccp=github.com/example_cc --ccid ExampleCC --v v0 --args '{"Args":["A","1","B","2"]}' --policy "AND('Org1MSP.member','Org2MSP.member')" --config ../../test/fixtures/config/config_test_local.yaml
```

Query ExampleCC chaincode on a set of peers

```
$ go run fabric-cli.go chaincode query --ccid ExampleCC --args '{"Func":"query","Args":["A"]}' --peer localhost:7051,localhost:8051 --config ../../test/fixtures/config/config_test_local.yaml
```

Listen for block events (output in JSON):

```
$ go run fabric-cli.go event listenblock --format json --config ../../test/fixtures/config/config_test_local.yaml
```

Then invoke ExampleCC chaincode in another shell and observe block events in the first shell.

```
$ cd $GOPATH/src/github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli
$ go run fabric-cli.go chaincode invoke --ccid=ExampleCC --args '{"Func":"move","Args":["A","B","1"]}' --config ../../test/fixtures/config/config_test_local.yaml
```

To clean up example network artifacts run

```
$ make example-network-clean
```

and follow instructions. This process is not yet automated.

## CLI examples

### Channel

#### Create a channel

$ go run fabric-cli.go channel create --cid mychannel --txfile fixtures/channel/mychannel.tx --config fixtures/config/config_test.yaml

#### Join a peer to a channel

$ go run fabric-cli.go channel join --cid mychannel --peer grpcs://localhost:7051

#### Join all peers in org1 to a channel

$ go run fabric-cli.go channel join --cid mychannel --orgid org1

#### Join all peers to a channel

$ go run fabric-cli.go channel join --cid mychannel

### Query

#### Query info:

$ go run fabric-cli.go query info --cid mychannel

#### Query block by block number:

$ go run fabric-cli.go query block --cid mychannel --num 0

#### Query block by hash (replace the example hash with a valid hash, e.g. using the output from query info):

$ go run fabric-cli.go query block --cid mychannel --hash MKUvwa85E7OvITqBZYmf8yn9QIS5eZkal2xLTleK2AA

#### Query block output in JSON format:

$ go run fabric-cli.go query block --cid mychannel --num 0 --format json

#### Query transaction (replace txid with a valid transaction, e.g. using the output from query block):

$ go run fabric-cli.go query tx --cid mychannel --txid 29bd4fd03e657da488acfa8ae1740eebf4a6ee81399bfc501f192cb407d2328c

#### Query channels joined by a peer:
$ go run fabric-cli.go query channels --peer grpcs://localhost:7051

#### Query installed chaincodes on a peer:

$ go run fabric-cli.go query installed --peer grpcs://localhost:7051

## Chaincode

#### Install chaincode on a peer:

$ go run fabric-cli.go chaincode install --cid=mychannel --peer grpcs://localhost:7051 --ccp=github.com/user/somecc --ccid=somecc --v v0

#### Install chaincode on all peers of org1:

$ go run fabric-cli.go chaincode install --cid=mychannel --orgid org1 --ccp=github.com/user/somecc --ccid=somecc --v v0

#### Install chaincode on all peers:

$ go run fabric-cli.go chaincode install --cid=mychannel --ccp=github.com/user/somecc --ccid=somecc --v v0

#### Instantiate chaincode with default endorsement policy:

$ go run fabric-cli.go chaincode instantiate --cid mychannel --ccp github.com/user/somecc --ccid somecc --v v0 --args='{"Args":["arg1","arg2"]}'

#### Instantiate chaincode with specified endorsement policy:

$ go run fabric-cli.go chaincode instantiate --cid mychannel --ccp github.com/user/somecc --ccid somecc --v v0 --args='{"Args":["arg1","arg2"]}' --policy "AND('Org1MSP.member','Org2MSP.member')"

#### Instantiate chaincode with specified private data collection configuration:

$ go run fabric-cli.go chaincode instantiate --cid mychannel --ccp github.com/user/somecc --ccid somecc --v v0 --args='{"Args":["arg1","arg2"]}' --collconfig fixtures/config/pvtdatacollection.json

#### Upgrade chaincode:

$ go run fabric-cli.go chaincode upgrade --cid mychannel --ccp github.com/user/somecc --ccid somecc --v v1 --args='{"Args":["arg1","arg2"]}' --policy "AND('Org1MSP.member','Org2MSP.member')"

#### Retrieve chaincode deployment info:

$ go run fabric-cli.go chaincode info --cid mychannel --ccid somecc

#### Query chaincode on a set of peers:

$ go run fabric-cli.go chaincode query --ccid=somecc --args='{"Func":"query","Args":["a"]}' --peer grpcs://localhost:7051,grpcs://localhost:8051

#### Query chaincode and view payloads only:

$ go run fabric-cli.go chaincode query --ccid=somecc --args='{"Func":"query","Args":["a"]}' --peer grpcs://localhost:7051,grpcs://localhost:8051 --payload

#### Invoke chaincode on all peers:

$ go run fabric-cli.go chaincode invoke --ccid=somecc --args='{"Func":"add","Args":["a","1","11"]}'

#### Invoke chaincode on all peers in org1:

$ go run fabric-cli.go chaincode invoke --ccid=somecc --args='{"Func":"add","Args":["a","1","11"]}' --orgid org1

#### Invoke chaincode using a 'dynamic' selection provider that chooses a minimal set of peers required to satisfy the endorsement policy of the chaincode:

$ go run fabric-cli.go chaincode invoke --ccid=somecc --args='{"Func":"add","Args":["a","1","11"]}' --orgid org1,org2 --selectprovider=dynamic

#### Invoke chaincode 5 times:

$ go run fabric-cli.go chaincode invoke --ccid=somecc --args='{"Func":"add","Args":["a","1","11"]}' --iterations 5

#### Invoke chaincode 100 times in 8 Go routines with 3 attempts for each invocation (in case the invocation fails):

$ go run fabric-cli.go chaincode invoke --ccid=somecc --args='{"Func":"add","Args":["a","1","11"]}' --iterations 100 --concurrency 8 --attempts=3

#### Invoke chaincode with two sets of args, 100 times each in 8 Go routines with 3 attempts for each invocation (in case the invocation fails):

$ go run fabric-cli.go chaincode invoke --ccid=somecc --args='[{"Func":"add","Args":["a","1","11"]},{"Func":"add","Args":["b","1","12"]}]' --iterations 100 --concurrency 8 --attempts=3

## Event

#### Listen for block events (output in JSON):

$ go run fabric-cli.go event listenblock --format json

#### Listen for filtered block events (output in JSON):

$ go run fabric-cli.go event listenfilteredblock --format json

#### Listen for chaincode events:

$ go run fabric-cli.go event listencc --ccid=somecc --event=someevent

#### Listen for tx event:

$ go run fabric-cli.go event listentx --txid <txid>
