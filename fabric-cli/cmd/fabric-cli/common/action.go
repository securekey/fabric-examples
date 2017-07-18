/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package common

import (
	"fmt"

	"sync"

	"github.com/hyperledger/fabric-sdk-go/api/apifabca"
	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	deffab "github.com/hyperledger/fabric-sdk-go/def/fabapi"
	"github.com/hyperledger/fabric-sdk-go/pkg/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/events"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/orderer"
	logging "github.com/op/go-logging"
	"github.com/spf13/pflag"
)

// ArgStruct is used for marshalling arguments to chaincode invocations
type ArgStruct struct {
	Func string   `json:"Func"`
	Args []string `json:"Args"`
}

// Action is the base implementation of the Action interface.
type Action struct {
	flags             *pflag.FlagSet
	eventHub          apifabclient.EventHub
	peersByOrg        map[string][]apifabclient.Peer
	peers             []apifabclient.Peer
	orgIDByPeer       map[string]string
	client            apifabclient.FabricClient
	channelClient     apifabclient.Channel
	mspClient         apifabca.FabricCAClient
	printer           Printer
	channelClientInit sync.Once
	eventHubInit      sync.Once
	initError         error
}

// Initialize initializes the action using the given flags
func (action *Action) Initialize(flags *pflag.FlagSet) error {
	action.flags = flags

	cnfg, err := config.InitConfig(Config().ConfigFile())
	if err != nil {
		return err
	}

	getConfigImpl().config = cnfg

	// Create SDK setup for the integration tests
	sdkOptions := deffab.Options{
		ConfigManager: cnfg,
		OrgID:         Config().OrgIDs()[0], // FIXME: Should allow connection to multiple MSPs
		StateStoreOpts: deffab.StateStoreOpts{
			Path: "/tmp/enroll_user",
		},
	}

	sdk, err := deffab.NewSDK(sdkOptions)
	if err != nil {
		return fmt.Errorf("Error initializing SDK: %s", err)
	}

	logging.SetLevel(levelFromName(Config().LoggingLevel()), loggerName)

	action.client = sdk.SystemClient
	action.mspClient = sdk.MSPClient

	action.orgIDByPeer = make(map[string]string)

	var allPeers []apifabclient.Peer
	allPeersByOrg := make(map[string][]apifabclient.Peer)
	for _, orgID := range Config().OrgIDs() {
		peersConfig, err := Config().PeersConfig("peer" + orgID)
		if err != nil {
			return fmt.Errorf("Error getting peer configs for org [%s]: %v", orgID, err)
		}

		var peers []apifabclient.Peer
		for _, p := range peersConfig {
			endorser, err := deffab.NewPeer(
				fmt.Sprintf("%s:%d", p.Host, p.Port),
				p.TLS.Certificate, p.TLS.ServerHostOverride, Config())
			if err != nil {
				return fmt.Errorf("NewPeer return error: %v", err)
			}
			peers = append(peers, endorser)
			action.orgIDByPeer[endorser.URL()] = orgID
		}
		allPeersByOrg[orgID] = peers
		allPeers = append(allPeers, peers...)
	}

	if Config().Logger().IsEnabledFor(logging.DEBUG) {
		Config().Logger().Debug("All Peers:")
		for orgID, peers := range allPeersByOrg {
			Config().Logger().Debugf("Org: %s\n", orgID)
			for i, peer := range peers {
				Config().Logger().Debugf("-- Peer[%d]: MSPID: %s, Name: %s, URL: %s\n", i, peer.MSPID(), peer.Name(), peer.URL())
			}
		}
	}

	if Config().PeerURL() != "" {
		peer, err := getPeer(allPeers, Config().PeerURL())
		if err != nil {
			return err
		}

		Config().Logger().Debugf("Selected Peer: Name: %s, URL: %s\n", peer.Name(), peer.URL())
		orgID := action.orgIDByPeer[peer.URL()]

		action.peers = []apifabclient.Peer{peer}
		action.peersByOrg = map[string][]apifabclient.Peer{orgID: action.peers}
	} else {
		action.peers = allPeers
		action.peersByOrg = allPeersByOrg
	}

	action.printer = NewPrinter(AsOutputFormat(Config().PrintFormat()))

	action.SetUserContext(action.OrgUser(Config().OrgID()))

	return nil
}

// Terminate closes any open connections. This function should be called at the end of every command invocation.
func (action *Action) Terminate() {
	if action.eventHub != nil {
		Config().Logger().Info("Disconnecting event hub")
		action.eventHub.Disconnect()
	}
}

// Flags returns the flag-set
func (action *Action) Flags() *pflag.FlagSet {
	return action.flags
}

// ChannelClient creates a new ChannelClient
func (action *Action) ChannelClient() (apifabclient.Channel, error) {
	action.channelClientInit.Do(func() {
		if channelClient, err := action.newChannelClient(); err != nil {
			action.initError = err
		} else {
			action.channelClient = channelClient
		}
	})
	return action.channelClient, action.initError
}

// SetUserContext sets the current user for the client
// TODO: This function should disappear when the SDK introduces sessions
func (action *Action) SetUserContext(user apifabclient.User) *UserContext {
	context := newUserContext(action.Client())
	action.Client().SetUserContext(user)
	return context
}

// Printer returns the Printer
func (action *Action) Printer() Printer {
	return action.printer
}

// EventHub returns the event hub.
func (action *Action) EventHub() (apifabclient.EventHub, error) {
	action.eventHubInit.Do(func() {
		eventHub, err := action.newEventHub()
		if err != nil {
			action.initError = err
		} else {
			action.eventHub = eventHub
		}
	})
	return action.eventHub, action.initError
}

// Peers returns the peers
func (action *Action) Peers() []apifabclient.Peer {
	return action.peers
}

// PeersByOrg returns the peers mapped by organization
func (action *Action) PeersByOrg() map[string][]apifabclient.Peer {
	return action.peersByOrg
}

// OrgOfPeer returns the organization ID of the given peer
func (action *Action) OrgOfPeer(peerURL string) (string, error) {
	orgID, ok := action.orgIDByPeer[peerURL]
	if !ok {
		return "", fmt.Errorf("org not found for peer %s", peerURL)
	}
	return orgID, nil
}

// Client returns the Fabric client
func (action *Action) Client() apifabclient.FabricClient {
	return action.client
}

// User returns the enrolled user. If the user doesn't exist then a new user is enrolled.
func (action *Action) User() (apifabclient.User, error) {
	userName := Config().UserName()

	user, err := action.Client().LoadUserFromStateStore(userName)
	if err != nil {
		return nil, fmt.Errorf("unable to load user: %s: %s", userName, err)
	}

	if user == nil {
		mspID, err := Config().MspID(Config().OrgID())
		if err != nil {
			return nil, fmt.Errorf("Error reading MSP ID config: %s", err)
		}
		user, err = deffab.NewUser(Config(), action.mspClient, userName, Config().UserPassword(), mspID)
		if err != nil {
			return nil, fmt.Errorf("NewUser returned error: %v", err)
		}
		err = action.Client().SaveUserToStateStore(user, false)
		if err != nil {
			return nil, fmt.Errorf("SaveUserToStateStore returned error: %v", err)
		}
	}

	return user, nil
}

// OrgUser returns the pre-enrolled user for the given organization
func (action *Action) OrgUser(orgID string) apifabclient.User {
	user, err := getUser(action.Client(), orgID)
	if err != nil {
		panic(fmt.Errorf("Error getting user for org %s: %v", orgID, err))
	}
	return user
}

// OrgAdminUser returns the pre-enrolled administrative user for the given organization
func (action *Action) OrgAdminUser(orgID string) apifabclient.User {
	admin, err := getAdmin(action.Client(), orgID)
	if err != nil {
		panic(fmt.Errorf("Error getting admin user for org %s: %v", orgID, err))
	}
	return admin
}

// OrgOrdererAdminUser returns the pre-enrolled orderer admin user for the given organization
func (action *Action) OrgOrdererAdminUser(orgID string) apifabclient.User {
	ordererAdmin, err := getOrdererAdmin(action.Client(), orgID)
	if err != nil {
		panic(fmt.Errorf("Error getting orderer admin user for org %s: %v", orgID, err))
	}
	return ordererAdmin
}

// PeerFromURL returns the peer for the given URL
func (action *Action) PeerFromURL(url string) apifabclient.Peer {
	for _, peer := range action.peers {
		if url == peer.URL() {
			return peer
		}
	}
	return nil
}

// Orderers returns all Orderers from the set of configured Orderers
func (action *Action) Orderers() ([]apifabclient.Orderer, error) {
	ordererConfigs, err := Config().OrderersConfig()
	if err != nil {
		return nil, fmt.Errorf("Could not orderer configurations: %s", err)
	}

	orderers := make([]apifabclient.Orderer, len(ordererConfigs))
	for i, ordererConfig := range ordererConfigs {
		orderer, err := orderer.NewOrderer(
			AsURL(ordererConfig.Host, ordererConfig.Port),
			ordererConfig.TLS.Certificate, ordererConfig.TLS.ServerHostOverride, Config())
		if err != nil {
			return nil, fmt.Errorf("NewOrderer return error: %v", err)
		}
		orderers[i] = orderer
	}

	return orderers, nil
}

// RandomOrderer chooses a random Orderer from the set of configured Orderers
func (action *Action) RandomOrderer() (apifabclient.Orderer, error) {
	ordererConfig, err := Config().RandomOrdererConfig()
	if err != nil {
		return nil, fmt.Errorf("Could not get config for orderer: %s", err)
	}

	orderer, err := orderer.NewOrderer(
		AsURL(ordererConfig.Host, ordererConfig.Port),
		ordererConfig.TLS.Certificate, ordererConfig.TLS.ServerHostOverride, Config())
	if err != nil {
		return nil, fmt.Errorf("NewOrderer return error: %v", err)
	}

	return orderer, nil
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

func getPeer(allPeers []apifabclient.Peer, peerURL string) (apifabclient.Peer, error) {
	if peerURL == "" {
		return nil, nil
	}

	var selectedPeer apifabclient.Peer
	for _, peer := range allPeers {
		if peer.URL() == peerURL {
			selectedPeer = peer
			break
		}
	}
	if selectedPeer == nil {
		return nil, fmt.Errorf("Peer not found for URL: %s", peerURL)
	}
	return selectedPeer, nil
}

func (action *Action) getEventHub(orgID, peerURL string) (apifabclient.EventHub, error) {
	eventHub, err := events.NewEventHub(action.Client())
	if err != nil {
		return nil, fmt.Errorf("Error creating new event hub: %v", err)
	}
	foundEventHub := false

	peerConfig, err := Config().PeersConfig(orgID)
	if err != nil {
		return nil, fmt.Errorf("Error reading peer config: %v", err)
	}

	for _, p := range peerConfig {
		if peerURL == "" || AsURL(p.Host, p.Port) == peerURL {
			Config().Logger().Infof("******* EventHub connect to %s peer (%s:%d) *******\n", orgID, p.EventHost, p.EventPort)
			eventHub.SetPeerAddr(fmt.Sprintf("%s:%d", p.EventHost, p.EventPort), p.TLS.Certificate, p.TLS.ServerHostOverride)
			foundEventHub = true
			break
		}
	}

	if !foundEventHub {
		return nil, fmt.Errorf("No EventHub configuration found for peer %s", peerURL)
	}

	return eventHub, nil
}

func (action *Action) newEventHub() (apifabclient.EventHub, error) {
	Config().Logger().Debugf("initEventHub - Initializing %s...\n")

	var orgID string
	peerURL := Config().PeerURL()
	if peerURL != "" {
		var err error
		if orgID, err = action.OrgOfPeer(peerURL); err != nil {
			return nil, err
		}
	} else {
		orgID = Config().OrgID()
	}

	eventHub, err := action.getEventHub(orgID, peerURL)
	if err != nil {
		return nil, fmt.Errorf("unable to get event hub: %s", err)
	}

	if err := eventHub.Connect(); err != nil {
		return nil, fmt.Errorf("unable to connect to event hub: %s", err)
	}

	return eventHub, nil
}

func (action *Action) newChannelClient() (apifabclient.Channel, error) {
	channelClient, err := action.Client().NewChannel(Config().ChannelID())
	if err != nil {
		return nil, fmt.Errorf("error creating channel client: %v", err)
	}

	orderers, err := action.Orderers()
	if err != nil {
		return nil, fmt.Errorf("error retrieving orderers: %v", err)
	}

	for _, orderer := range orderers {
		channelClient.AddOrderer(orderer)
	}

	for _, peer := range action.Peers() {
		channelClient.AddPeer(peer)
	}

	context := action.SetUserContext(action.OrgAdminUser(Config().OrgID()))
	defer context.Restore()

	if err := channelClient.Initialize(nil); err != nil {
		return nil, fmt.Errorf("Error initializing channel: %v", err)
	}

	return channelClient, nil
}
