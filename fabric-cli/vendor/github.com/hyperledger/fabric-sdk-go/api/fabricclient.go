/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package api

import (
	"github.com/hyperledger/fabric/bccsp"
	"github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// FabricClient ...
/*
 * Main interaction handler with end user. A client instance provides a handler to interact
 * with a network of peers, orderers and optionally member services. An application using the
 * SDK may need to interact with multiple networks, each through a separate instance of the Client.
 *
 * Each client when initially created should be initialized with configuration data from the
 * consensus service, which includes a list of trusted roots, orderer certificates and IP addresses,
 * and a list of peer certificates and IP addresses that it can access. This must be done out of band
 * as part of bootstrapping the application environment. It is also the responsibility of the application
 * to maintain the configuration of a client as the SDK does not persist this object.
 *
 * Each Client instance can maintain several {@link Channel} instances representing channels and the associated
 * private ledgers.
 *
 *
 */
type FabricClient interface {
	NewChannel(name string) (Channel, error)
	GetChannel(name string) Channel
	ExtractChannelConfig(configEnvelope []byte) ([]byte, error)
	SignChannelConfig(config []byte) (*common.ConfigSignature, error)
	CreateChannel(request *CreateChannelRequest) error
	QueryChannelInfo(name string, peers []Peer) (Channel, error)
	SetStateStore(stateStore KeyValueStore)
	GetStateStore() KeyValueStore
	SetCryptoSuite(cryptoSuite bccsp.BCCSP)
	GetCryptoSuite() bccsp.BCCSP
	SaveUserToStateStore(user User, skipPersistence bool) error
	LoadUserFromStateStore(name string) (User, error)
	InstallChaincode(chaincodeName string, chaincodePath string, chaincodeVersion string, chaincodePackage []byte, targets []Peer) ([]*TransactionProposalResponse, string, error)
	QueryChannels(peer Peer) (*pb.ChannelQueryResponse, error)
	QueryInstalledChaincodes(peer Peer) (*pb.ChaincodeQueryResponse, error)
	GetIdentity() ([]byte, error)
	GetUserContext() User
	SetUserContext(user User)
	GetConfig() Config
}

// CreateChannelRequest requests channel creation on the network
type CreateChannelRequest struct {
	// required - The name of the new channel
	Name string
	// required - The Orderer to send the update request
	Orderer Orderer
	// optional - the envelope object containing all
	// required settings and signatures to initialize this channel.
	// This envelope would have been created by the command
	// line tool "configtx"
	Envelope []byte
	// optional - ConfigUpdate object built by the
	// buildChannelConfig() method of this package
	Config []byte
	// optional - the list of collected signatures
	// required by the channel create policy when using the `config` parameter.
	// see signChannelConfig() method of this package
	Signatures []*common.ConfigSignature
	// optional - transaction ID
	// required when using the `config` parameter
	TxID string
	// optional - nonce
	// required when using the `config` parameter
	Nonce []byte
}
