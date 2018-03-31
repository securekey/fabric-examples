/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cmd

import (
	"os"
	"testing"

	"strings"

	"fmt"

	"github.com/hyperledger/fabric-sdk-go/test/metadata"
)

const (
	cryptoConfigPath = "../fabric-examples/fabric-cli/cmd/fabric-cli/fixtures/fabric/v1/crypto-config"
)

func run(cmd string) {
	metadata.CryptoConfigPath = cryptoConfigPath
	os.Args = strings.Split(cmd, " ")
	os.Setenv("GRPC_TRACE", "all")
	os.Setenv("GRPC_VERBOSITY", "DEBUG")
	os.Setenv("GRPC_GO_LOG_SEVERITY_LEVEL", "INFO")
	Cmd()
}

func header(h string) {
	fmt.Println("****************************************")
	fmt.Println("*")
	fmt.Printf("*     %s\n", h)
	fmt.Println("*")
	fmt.Println("****************************************")
}

func TestCreate_a_channel(t *testing.T) {
	header("Create a channel")
	run("fabric-cli channel create --cid mychannel --txfile ../fixtures/fabric/v1.1/channel/mychannel.tx --config ../fixtures/config/config_test.yaml")
}

func TestJoin_org1_peer_to_a_channel(t *testing.T) {
	header("Join a peer to a channel")
	run("fabric-cli.go channel join --cid mychannel --peer localhost:7051 --config ../fixtures/config/config_test.yaml")
}

func TestJoin_org2_peer_to_a_channel(t *testing.T) {
	header("Join a peer to a channel")
	run("fabric-cli.go channel join --cid mychannel --peer localhost:8051 --config ../fixtures/config/config_test.yaml")
}

func TestJoin_all_peers_in_org1_to_a_channel(t *testing.T) {
	header("Join all peers in org1 to a channel")
	run("fabric-cli.go channel join --cid mychannel --orgid org1 --config ../fixtures/config/config_test.yaml")
}

func TestJoin_all_peers_in_org2_to_a_channel(t *testing.T) {
	header("Join all peers in org1 to a channel")
	run("fabric-cli.go channel join --cid mychannel --orgid org2 --config ../fixtures/config/config_test.yaml")
}

func TestJoin_all_peers_to_a_channel(t *testing.T) {
	header("Join all peers to a channel")
	run("fabric-cli.go channel join --cid mychannel --config ../fixtures/config/config_test.yaml")
}

func TestOther(t *testing.T) {

	///// Query
	//
	////// Query info:
	//
	//run("fabric-cli.go query info --cid mychannel")
	//
	////// Query block by block number:
	//
	//run("fabric-cli.go query block --cid mychannel --num 0")
	//
	////// Query block by hash:
	//
	//run("fabric-cli.go query block --cid mychannel --hash MKUvwa85E7OvITqBZYmf8yn9QIS5eZkal2xLTleK2AA")
	//
	////// Query block output in JSON format:
	//
	//run("fabric-cli.go query block --cid mychannel --num 0 --format json")
	//
	////// Query transaction:
	//
	//run("fabric-cli.go query tx --cid mychannel --txid 29bd4fd03e657da488acfa8ae1740eebf4a6ee81399bfc501f192cb407d2328c")
	//
	////// Query channels joined by a peer:
	//
	//run("fabric-cli.go query channels --peer grpcs://localhost:7051")
	//
	////// Query installed chaincodes on a peer:
	//
	//run("fabric-cli.go query installed --peer grpcs://localhost:7051")
	//
	//// Chaincode
	//
	////// Install chaincode on a peer:
	//
	//run("fabric-cli.go chaincode install --cid=mychannel --peer grpcs://localhost:7051 --ccp=github.com/user/somecc --ccid=somecc --v v0")
	//
	////// Install chaincode on all peers of org1:
	//
	//run("fabric-cli.go chaincode install --cid=mychannel --orgid org1 --ccp=github.com/user/somecc --ccid=somecc --v v0")
	//
	////// Install chaincode on all peers:
	//
	//run("fabric-cli.go chaincode install --cid=mychannel --ccp=github.com/user/somecc --ccid=somecc --v v0")
	//
	////// Instantiate chaincode with default endorsement policy:
	//
	//run("fabric-cli.go chaincode instantiate --cid mychannel --ccp github.com/user/somecc --ccid somecc --v v0 --args='{\"Args\":[\"arg1\",\"arg2\"]}")
	//
	////// Instantiate chaincode with specified endorsement policy:
	//
	//run("fabric-cli.go chaincode instantiate --cid mychannel --ccp github.com/user/somecc --ccid somecc --v v0 --args='{\"Args\":[\"arg1\",\"arg2\"]}' --policy \"AND('Org1MSP.member','Org2MSP.member')\"")
	//
	////// Instantiate chaincode with specified private data collection configuration:
	//
	//run("fabric-cli.go chaincode instantiate --cid mychannel --ccp github.com/user/somecc --ccid somecc --v v0 --args='{\"Args\":[\"arg1\",\"arg2\"]}' --collconfig fixtures/config/pvtdatacollection.json")
	//
	////// Upgrade chaincode:
	//
	//run("fabric-cli.go chaincode upgrade --cid mychannel --ccp github.com/user/somecc --ccid somecc --v v1 --args='{\"Args\":[\"arg1\",\"arg2\"]}' --policy \"AND('Org1MSP.member','Org2MSP.member')\"")
	//
	////// Retrieve chaincode deployment info:
	//
	//run("fabric-cli.go chaincode info --cid mychannel --ccid somecc")
	//
	////// Query chaincode on a set of peers:
	//
	//run("fabric-cli.go chaincode query --ccid=somecc --args='{\"Func\":\"query\",\"Args\":[\"a\"]}' --peer grpcs://localhost:7051,grpcs://localhost:8051")
	//
	////// Query chaincode and view payloads only:
	//
	//run("fabric-cli.go chaincode query --ccid=somecc --args='{\"Func\":\"query\",\"Args\":[\"a\"]}' --peer grpcs://localhost:7051,grpcs://localhost:8051 --payload")
	//
	////// Invoke chaincode on all peers:
	//
	//run("fabric-cli.go chaincode invoke --ccid=somecc --args='{\"Func\":\"add\",\"Args\":[\"a\",\"1\",\"11\"]}'")
	//
	////// Invoke chaincode on all peers in org1:
	//
	//run("fabric-cli.go chaincode invoke --ccid=somecc --args='{\"Func\":\"add\",\"Args\":[\"a\",\"1\",\"11\"]}' --orgid org1")
	//
	////// Invoke chaincode using a 'dynamic' selection provider that chooses a minimal set of peers required to satisfy the endorsement policy of the chaincode:
	//
	//run("fabric-cli.go chaincode invoke --ccid=somecc --args='{\"Func\":\"add\",\"Args\":[\"a\",\"1\",\"11\"]}' --orgid org1,org2 --selectprovider=dynamic")
	//
	////// Invoke chaincode 5 times:
	//
	//run("fabric-cli.go chaincode invoke --ccid=somecc --args='{\"Func\":\"add\",\"Args\":[\"a\",\"1\",\"11\"]}' --iterations 5")
	//
	////// Invoke chaincode 100 times in 8 Go routines with 3 attempts for each invocation (in case the invocation fails):
	//
	//run("fabric-cli.go chaincode invoke --ccid=somecc --args='{\"Func\":\"add\",\"Args\":[\"a\",\"1\",\"11\"]}' --iterations 100 --concurrency 8 --attempts=3")
	//
	////// Invoke chaincode with two sets of args, 100 times each in 8 Go routines with 3 attempts for each invocation (in case the invocation fails):
	//
	//run("fabric-cli.go chaincode invoke --ccid=somecc --args='[{\"Func\":\"add\",\"Args\":[\"a\",\"1\",\"11\"]},{\"Func\":\"add\",\"Args\":[\"b\",\"1\",\"12\"]}]' --iterations 100 --concurrency 8 --attempts=3")
	//
	//// Event
	//
	////// Listen for block events (output in JSON):
	//
	//run("fabric-cli.go event listenblock --format json")
	//
	////// Listen for chaincode events:
	//
	//run("fabric-cli.go event listencc --ccid=somecc --event=someevent")

}
