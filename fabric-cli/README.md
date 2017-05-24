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
- Invoke - Invokes chaincode with a transaction proposal and a commit
- Query - Invokes chaincode without a commit
- Info - Retrieves details about the chaincode

Events:
- Listen Block - Listens for block events and displays the block when a new block is created
- Listen TX - Listens for transaction events
- Listen CC - Listens for chaincode events for a specified chaincode

## Running

go run fabriccli.go <command> <sub-command> [options]

To display the available commands/options:

go run fabriccli.go

## Compatability

This example is compatible with the following Hyperledger Fabric commit levels:
- fabric: v1.0.0-alpha
- fabric-ca: v1.0.0-alpha
