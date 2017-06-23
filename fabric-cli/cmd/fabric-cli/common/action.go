/*
Copyright SecureKey Technologies Inc. All Rights Reserved.


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at


      http://www.apache.org/licenses/LICENSE-2.0


Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package common

import (
	"fmt"

	"github.com/hyperledger/fabric-sdk-go/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/events"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/orderer"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/peer"
	fcutil "github.com/hyperledger/fabric-sdk-go/pkg/util"
	bccspFactory "github.com/hyperledger/fabric/bccsp/factory"
	logging "github.com/op/go-logging"
	"github.com/spf13/pflag"
)

// ArgStruct is used for marshalling arguments to chaincode invocations
type ArgStruct struct {
	Args []string `json:"Args"`
}

// Action is implemented by all command actions
type Action interface {
	Initialize(flags *pflag.FlagSet) error
	Flags() *pflag.FlagSet
	Invoke() error
	EventHub() api.EventHub
	Peers() []api.Peer
	Client() api.FabricClient
	Printer() Printer
}

// ActionImpl is the base implementation of the Action interface.
type ActionImpl struct {
	flags    *pflag.FlagSet
	eventHub api.EventHub
	peers    []api.Peer
	client   api.FabricClient
	printer  Printer
}

// GetAdminUser returns the admin user for the given org
func (action *ActionImpl) GetAdminUser(orgID string) (api.User, error) {
	keyDir := fmt.Sprintf("peerOrganizations/%s.example.com/users/Admin@%s.example.com/keystore", orgID, orgID)
	certDir := fmt.Sprintf("peerOrganizations/%s.example.com/users/Admin@%s.example.com/signcerts", orgID, orgID)
	username := fmt.Sprintf("peer%sAdmin", orgID)
	user, err := fcutil.GetPreEnrolledUser(action.Client(), keyDir, certDir, username)
	if err != nil {
		return nil, fmt.Errorf("Error getting org admin user: %v", err)
	}
	return user, nil
}

// GetOrdererAdminUser returns the admin user for the orderer
func (action *ActionImpl) GetOrdererAdminUser() (api.User, error) {
	keyDir := "ordererOrganizations/example.com/users/Admin@example.com/keystore"
	certDir := "ordererOrganizations/example.com/users/Admin@example.com/signcerts"
	user, err := fcutil.GetPreEnrolledUser(action.Client(), keyDir, certDir, "ordererAdmin")
	if err != nil {
		return nil, fmt.Errorf("Error getting orderer admin user: %v", err)
	}
	return user, nil
}

// Initialize initializes the action using the given flags
func (action *ActionImpl) Initialize(flags *pflag.FlagSet) error {
	action.flags = flags

	cnfg, err := config.InitConfig(Config().ConfigFile())
	if err != nil {
		return err
	}

	getConfigImpl().config = cnfg

	logging.SetLevel(levelFromName(Config().LoggingLevel()), loggerName)

	// Initialize bccsp factories before calling get client
	err = bccspFactory.InitFactories(Config().GetCSPConfig())
	if err != nil {
		return fmt.Errorf("Failed getting ephemeral software-based BCCSP [%s]", err)
	}

	client, err := fcutil.GetClient(Config().User(), Config().Password(), userStatePath, Config())
	if err != nil {
		return fmt.Errorf("Create client failed: %v", err)
	}
	action.client = client

	peersConfig, err := Config().GetPeersConfig()
	if err != nil {
		return fmt.Errorf("Error getting peer configs: %v", err)
	}

	var allPeers []api.Peer
	for _, p := range peersConfig {
		endorser, err := peer.NewPeer(fmt.Sprintf("%s:%d", p.Host, p.Port),
			p.TLS.Certificate, p.TLS.ServerHostOverride, Config())
		if err != nil {
			return fmt.Errorf("NewPeer return error: %v", err)
		}
		allPeers = append(allPeers, endorser)
	}

	orgAdminUser, err := action.GetAdminUser("org1")
	if err != nil {
		return fmt.Errorf("Error getting org admin user: %v", err)
	}

	client.SetUserContext(orgAdminUser)
	eventHub, err := action.getEventHub(Config().PeerURL())
	if err != nil {
		return err
	}
	action.eventHub = eventHub

	if err := eventHub.Connect(); err != nil {
		return fmt.Errorf("Failed eventHub.Connect() [%s]", err)
	}

	for i, peer := range allPeers {
		Config().Logger().Debugf("Peer[%d]: Name: %s, URL: %s\n", i, peer.GetName(), peer.GetURL())
	}

	var peers []api.Peer
	if Config().PeerURL() != "" {
		peers, err = getPeers(allPeers, Config().PeerURL())
		if err != nil {
			return err
		}
	} else {
		peers = allPeers
	}

	action.peers = peers
	action.printer = NewPrinter(AsOutputFormat(Config().PrintFormat()))

	return nil
}

// Flags returns the flag-set
func (action *ActionImpl) Flags() *pflag.FlagSet {
	return action.flags
}

// NewChannel creates a new Channel
func (action *ActionImpl) NewChannel() (api.Channel, error) {
	o, err := orderer.NewOrderer(
		fmt.Sprintf("%s:%s", Config().GetOrdererHost(), Config().GetOrdererPort()),
		Config().GetOrdererTLSCertificate(), Config().GetOrdererTLSServerHostOverride(), Config())
	if err != nil {
		return nil, fmt.Errorf("NewOrderer return error: %v", err)
	}

	channel, err := channel.NewChannel(Config().ChannelID(), action.Client())
	if err != nil {
		return nil, fmt.Errorf("Could not get channel: %v", err)
	}

	channel.AddOrderer(o)
	for _, peer := range action.peers {
		channel.AddPeer(peer)
	}

	if err := channel.Initialize(nil); err != nil {
		return nil, fmt.Errorf("Error initializing channel: %v", err)
	}

	return channel, err
}

// Printer returns the Printer
func (action *ActionImpl) Printer() Printer {
	return action.printer
}

// EventHub returns the event hub
func (action *ActionImpl) EventHub() api.EventHub {
	return action.eventHub
}

// Peers returns the peers
func (action *ActionImpl) Peers() []api.Peer {
	return action.peers
}

// Client returns the Fabric client
func (action *ActionImpl) Client() api.FabricClient {
	return action.client
}

// PeerFromURL returns the peer for the given URL
func (action *ActionImpl) PeerFromURL(url string) api.Peer {
	for _, peer := range action.peers {
		if url == peer.GetURL() {
			return peer
		}
	}
	return nil
}

func levelFromName(levelName string) logging.Level {
	switch levelName {
	case "CRITICAL":
		return logging.CRITICAL
	case "ERROR":
		return logging.ERROR
	case "WARNING":
		return logging.WARNING
	case "INFO":
		return logging.INFO
	case "DEBUG":
		return logging.DEBUG
	default:
		return logging.CRITICAL
	}
}

func getPeers(allPeers []api.Peer, peerURL string) ([]api.Peer, error) {
	if peerURL == "" {
		return allPeers, nil
	}

	var selectedPeer api.Peer
	for _, peer := range allPeers {
		if peer.GetURL() == peerURL {
			selectedPeer = peer
			break
		}
	}
	if selectedPeer == nil {
		return nil, fmt.Errorf("Peer not found for URL: %s", peerURL)
	}

	return []api.Peer{selectedPeer}, nil
}

func (action *ActionImpl) getEventHub(peerURL string) (api.EventHub, error) {
	eventHub, err := events.NewEventHub(action.Client())
	if err != nil {
		return nil, fmt.Errorf("Error creating new event hub: %v", err)
	}
	foundEventHub := false
	peerConfig, err := Config().GetPeersConfig()
	if err != nil {
		return nil, fmt.Errorf("Error reading peer config: %v", err)
	}
	for _, p := range peerConfig {
		if p.EventHost != "" && p.EventPort != 0 {
			fmt.Printf("******* EventHub connect to peer (%s:%d) *******\n", p.EventHost, p.EventPort)
			eventHub.SetPeerAddr(fmt.Sprintf("%s:%d", p.EventHost, p.EventPort),
				p.TLS.Certificate, p.TLS.ServerHostOverride)
			foundEventHub = true
			break
		}
	}

	if !foundEventHub {
		return nil, fmt.Errorf("No EventHub configuration found")
	}

	return eventHub, nil
}
