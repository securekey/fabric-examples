/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package common

import (
	"sync"

	"github.com/hyperledger/fabric-sdk-go/pkg/errors"

	"strings"

	"io"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	deffab "github.com/hyperledger/fabric-sdk-go/def/fabapi"
	"github.com/hyperledger/fabric-sdk-go/def/fabapi/context"
	"github.com/hyperledger/fabric-sdk-go/def/fabapi/opt"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/events"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/orderer"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
	cliconfig "github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/config"
	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/printer"
	"github.com/spf13/pflag"
)

const (
	// FIXME: Make configurable
	defaultUser  = "User1" // pre-enrolled user
	adminUser    = "admin"
	ordererOrgID = "ordererorg"
)

// ArgStruct is used for marshalling arguments to chaincode invocations
type ArgStruct struct {
	Func string   `json:"Func"`
	Args []string `json:"Args"`
}

// Action is the base implementation of the Action interface.
type Action struct {
	flags        *pflag.FlagSet
	eventHub     apifabclient.EventHub
	peersByOrg   map[string][]apifabclient.Peer
	peers        []apifabclient.Peer
	orgIDByPeer  map[string]string
	sdk          *deffab.FabricSDK
	printer      printer.Printer
	eventHubInit sync.Once
	initError    error
	Writer       io.Writer
	sessions     map[string]context.Session
}

// Initialize initializes the action using the given flags
func (action *Action) Initialize(flags *pflag.FlagSet) error {
	action.sessions = make(map[string]context.Session)
	action.flags = flags

	if err := cliconfig.InitConfig(flags); err != nil {
		return err
	}

	sdkOptions := deffab.Options{
		ProviderFactory: NewProviderFactory(cliconfig.Config()),
		StateStoreOpts: opt.StateStoreOpts{
			Path: "/tmp/enroll_user",
		},
	}

	sdk, err := deffab.NewSDK(sdkOptions)
	if err != nil {
		return errors.Errorf("Error initializing SDK: %s", err)
	}
	action.sdk = sdk

	level := levelFromName(cliconfig.Config().LoggingLevel())

	logging.SetLevel(level, "")

	action.orgIDByPeer = make(map[string]string)

	var allPeers []apifabclient.Peer
	allPeersByOrg := make(map[string][]apifabclient.Peer)
	for _, orgID := range cliconfig.Config().OrgIDs() {
		peersConfig, err := cliconfig.Config().PeersConfig(orgID)
		if err != nil {
			return errors.Errorf("Error getting peer configs for org [%s]: %v", orgID, err)
		}

		cliconfig.Config().Logger().Debugf("Peers for org [%s]: %v\n", orgID, peersConfig)

		var peers []apifabclient.Peer
		for _, p := range peersConfig {
			serverHostOverride := ""
			if str, ok := p.GRPCOptions["ssl-target-name-override"].(string); ok {
				serverHostOverride = str
			}
			endorser, err := deffab.NewPeer(
				p.URL,
				p.TLSCACerts.Path, serverHostOverride, cliconfig.Config())
			if err != nil {
				return errors.Errorf("NewPeer return error: %v", err)
			}
			peers = append(peers, endorser)
			action.orgIDByPeer[endorser.URL()] = orgID
		}
		allPeersByOrg[orgID] = peers
		allPeers = append(allPeers, peers...)
	}

	if cliconfig.Config().IsLoggingEnabledFor(logging.DEBUG) {
		cliconfig.Config().Logger().Debug("All Peers:")
		for orgID, peers := range allPeersByOrg {
			cliconfig.Config().Logger().Debugf("Org: %s\n", orgID)
			for i, peer := range peers {
				cliconfig.Config().Logger().Debugf("-- Peer[%d]: MSPID: %s, Name: %s, URL: %s\n", i, peer.MSPID(), peer.Name(), peer.URL())
			}
		}
	}

	if cliconfig.Config().PeerURL() != "" {
		peers, err := getPeers(allPeers, cliconfig.Config().PeerURL())
		if err != nil {
			return err
		}

		action.peers = peers
		action.peersByOrg = make(map[string][]apifabclient.Peer)

		cliconfig.Config().Logger().Debugf("Selected Peers:\n")
		for _, peer := range peers {
			cliconfig.Config().Logger().Debugf("- Name: %s, URL: %s\n", peer.Name(), peer.URL())
			orgID := action.orgIDByPeer[peer.URL()]
			if orgID == "" {
				return errors.Errorf("unable to find org for peer: %s", peer.URL())
			}
			action.peersByOrg[orgID] = append(action.peersByOrg[orgID], peer)
		}
	} else {
		action.peers = allPeers
		action.peersByOrg = allPeersByOrg
	}

	action.printer = printer.NewPrinter(printer.AsOutputFormat(cliconfig.Config().PrintFormat()), printer.AsWriterType(cliconfig.Config().Writer()))

	// context, err := sdk.NewContext(cliconfig.Config().OrgID())
	// if err != nil {
	// 	return errors.Errorf("Error getting a context for org: %s", err)
	// }
	// action.context = context

	// // action.SetUserContext(action.OrgUser(cliconfig.Config().OrgID()))
	// user, err := action.User()
	// if err != nil {
	// 	return errors.Errorf("Error getting user: %s", err)
	// }

	// action.SetUserContext(user)

	return nil
}

// Terminate closes any open connections. This function should be called at the end of every command invocation.
func (action *Action) Terminate() {
	if action.eventHub != nil {
		cliconfig.Config().Logger().Info("Disconnecting event hub")
		action.eventHub.Disconnect()
	}
}

// Flags returns the flag-set
func (action *Action) Flags() *pflag.FlagSet {
	return action.flags
}

// ChannelClient creates a new channel client
func (action *Action) ChannelClient() (apitxn.ChannelClient, error) {
	user, err := action.User()
	if err != nil {
		return nil, errors.Errorf("error getting user: %s", err)
	}

	session, err := action.session(action.OrgID(), user)
	if err != nil {
		return nil, errors.Errorf("error getting user session: %s", err)
	}

	return action.sdk.SessionFactory.NewChannelClient(action.sdk, session, action.sdk.ConfigProvider(), cliconfig.Config().ChannelID())
}

// OrgAdminChannelClient creates a new channel client for the given org in order to perform administrative functions
func (action *Action) OrgAdminChannelClient(orgID string) (apifabclient.Channel, error) {
	channelID := cliconfig.Config().ChannelID()
	cliconfig.Config().Logger().Debugf("Creating new channel client for channel [%s]...", channelID)

	user, err := action.OrgAdminUser(orgID)
	if err != nil {
		return nil, err
	}

	fabClient, err := action.ClientForUser(orgID, user)
	if err != nil {
		return nil, errors.Errorf("error creating fabric client: %s", err)
	}

	channelClient, err := fabClient.NewChannel(channelID)
	if err != nil {
		return nil, errors.Errorf("error creating channel client: %v", err)
	}

	orderers, err := action.Orderers()
	if err != nil {
		return nil, errors.Errorf("error retrieving orderers: %v", err)
	}

	for _, orderer := range orderers {
		channelClient.AddOrderer(orderer)
	}

	for _, peer := range action.Peers() {
		channelClient.AddPeer(peer)
	}

	if err := channelClient.Initialize(nil); err != nil {
		return nil, errors.Errorf("Error initializing channel: %v", err)
	}

	return channelClient, nil
}

// AdminChannelClient creates a new channel client for performing administrative functions
func (action *Action) AdminChannelClient() (apifabclient.Channel, error) {
	return action.OrgAdminChannelClient(action.OrgID())
}

// Printer returns the Printer
func (action *Action) Printer() printer.Printer {
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

// Peer returns the first peer in the list of selected peers
func (action *Action) Peer() apifabclient.Peer {
	if len(action.peers) == 0 {
		return nil
	}
	return action.peers[0]
}

// Peers returns the peers
func (action *Action) Peers() []apifabclient.Peer {
	return action.peers
}

// ProposalProcessor returns the first proposal processor in the list of selected processors
func (action *Action) ProposalProcessor() apitxn.ProposalProcessor {
	return action.Peer()
}

// ProposalProcessors returns the proposal processors
func (action *Action) ProposalProcessors() []apitxn.ProposalProcessor {
	targets := make([]apitxn.ProposalProcessor, len(action.Peers()))
	for i, p := range action.Peers() {
		targets[i] = p
	}
	return targets
}

// PeersByOrg returns the peers mapped by organization
func (action *Action) PeersByOrg() map[string][]apifabclient.Peer {
	return action.peersByOrg
}

// OrgOfPeer returns the organization ID of the given peer
func (action *Action) OrgOfPeer(peerURL string) (string, error) {
	orgID, ok := action.orgIDByPeer[peerURL]
	if !ok {
		return "", errors.Errorf("org not found for peer %s", peerURL)
	}
	return orgID, nil
}

// Client returns the Fabric client for the current user
func (action *Action) Client() (apifabclient.FabricClient, error) {
	user, err := action.User()
	if err != nil {
		return nil, err
	}
	return action.ClientForUser(action.OrgID(), user)
}

// ClientForUser returns the Fabric client for the given user
func (action *Action) ClientForUser(orgID string, user apifabclient.User) (apifabclient.FabricClient, error) {
	cliconfig.Config().Logger().Debugf("Create admin channel client for user [%s] in org [%s]...", user.Name(), orgID)
	session, err := action.session(orgID, user)
	if err != nil {
		return nil, errors.Errorf("error getting session for user [%s,%s]: %s", orgID, user.Name(), err)
	}

	cliconfig.Config().Logger().Infof("Creating new system client with user session[%s:%s:%s]\n", orgID, session.Identity().Name(), session.Identity().MspID())

	return action.sdk.NewSystemClient(session)
}

func (action *Action) session(orgID string, user apifabclient.User) (context.Session, error) {
	key := orgID + "_" + user.Name()
	session := action.sessions[key]
	if session == nil {
		var err error
		session, err = action.newSession(orgID, user)
		if err != nil {
			return nil, errors.Errorf("error creating session for user [%s] in org [%s]: %s", user.Name(), orgID, err)
		}
		cliconfig.Config().Logger().Debugf("Created session for user [%s] in org [%s]", user.Name(), orgID)
		action.sessions[key] = session
	}
	return session, nil
}

// OrgID returns the organization ID of the first peer in the list of peers
func (action *Action) OrgID() string {
	if len(action.Peers()) == 0 {
		// This shouldn't happen since we should already have passed validation
		panic("no peers to choose from!")
	}

	peer := action.Peers()[0]
	orgID, err := action.OrgOfPeer(peer.URL())
	if err != nil {
		// This shouldn't happen since we should already have passed validation
		panic(err)
	}
	return orgID
}

// GetOrgID returns the organization ID for the given MSP ID
func (action *Action) GetOrgID(mspID string) (string, error) {
	networkConfig, err := cliconfig.Config().NetworkConfig()
	if err != nil {
		return "", err
	}
	for orgID, orgConfig := range networkConfig.Organizations {
		if mspID == orgConfig.MspID {
			return orgID, nil
		}
	}
	return "", errors.Errorf("unable to find org ID for MSP [%s]", mspID)
}

// User returns the enrolled user. If the user doesn't exist then a new user is enrolled.
func (action *Action) User() (apifabclient.User, error) {
	userName := cliconfig.Config().UserName()
	if userName == "" {
		userName = defaultUser
	}
	return action.OrgUser(action.OrgID(), userName)
}

func (action *Action) newUser(orgID, userName, pwd string) (apifabclient.User, error) {
	cliconfig.Config().Logger().Infof("Loading user %s...\n", userName)

	var user apifabclient.User
	// user, err := action.Client().LoadUserFromStateStore(userName)
	// if err != nil {
	// 	return nil, errors.Errorf("unable to load user: %s: %s", userName, err)
	// }

	if user == nil {
		cliconfig.Config().Logger().Infof("Enrolling user %s...\n", userName)
		mspID, err := cliconfig.Config().MspID(orgID)
		if err != nil {
			return nil, errors.Errorf("error reading MSP ID config: %s", err)
		}

		caClient, err := deffab.NewCAClient(orgID, cliconfig.Config())
		if err != nil {
			return nil, errors.Errorf("error creating CA client: %s", err)
		}

		cliconfig.Config().Logger().Infof("Creating new user %s...\n", userName)
		user, err = deffab.NewUser(cliconfig.Config(), caClient, userName, pwd, mspID)
		if err != nil {
			return nil, errors.Errorf("NewUser returned error: %v", err)
		}

		// cliconfig.Config().Logger().Infof("Saving user to state store %s...\n", userName)
		// err = action.Client().SaveUserToStateStore(user, false)
		// if err != nil {
		// 	return nil, errors.Errorf("SaveUserToStateStore returned error: %v", err)
		// }
	}

	cliconfig.Config().Logger().Infof("Returning user [%s], MSPID [%s]\n", user.Name(), user.MspID())

	return user, nil
}

// OrgUser returns the pre-enrolled user for the given organization
func (action *Action) OrgUser(orgID, userName string) (apifabclient.User, error) {
	if userName == "" {
		return nil, errors.Errorf("no user name specified")
	}

	user, err := action.sdk.NewPreEnrolledUser(orgID, userName)
	if err == nil {
		return user, nil
	}
	cliconfig.Config().Logger().Debugf("Error getting pre-enrolled user for org %s: %v. Trying enrolled user...", orgID, err)

	// FIXME: Password should be passed in?
	return action.newUser(orgID, userName, cliconfig.Config().UserPassword())
}

// OrgAdminUser returns the pre-enrolled administrative user for the given organization
func (action *Action) OrgAdminUser(orgID string) (apifabclient.User, error) {
	userName := cliconfig.Config().UserName()
	if userName == "" {
		userName = adminUser
	}
	return action.OrgUser(orgID, userName)
}

// OrdererAdminUser returns the pre-enrolled administrative user for the orderer organization
func (action *Action) OrdererAdminUser() (apifabclient.User, error) {
	userName := cliconfig.Config().UserName()
	if userName == "" {
		userName = adminUser
	}
	return action.OrgUser(ordererOrgID, userName)
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
	ordererConfigs, err := cliconfig.Config().OrderersConfig()
	if err != nil {
		return nil, errors.Errorf("Could not orderer configurations: %s", err)
	}

	orderers := make([]apifabclient.Orderer, len(ordererConfigs))
	for i, ordererConfig := range ordererConfigs {
		serverHostOverride := ""
		if str, ok := ordererConfig.GRPCOptions["ssl-target-name-override"].(string); ok {
			serverHostOverride = str
		}
		orderer, err := orderer.NewOrderer(
			ordererConfig.URL,
			ordererConfig.TLSCACerts.Path, serverHostOverride, cliconfig.Config())
		if err != nil {
			return nil, errors.Errorf("NewOrderer return error: %v", err)
		}
		orderers[i] = orderer
	}

	return orderers, nil
}

// RandomOrderer chooses a random Orderer from the set of configured Orderers
func (action *Action) RandomOrderer() (apifabclient.Orderer, error) {
	ordererConfig, err := cliconfig.Config().RandomOrdererConfig()
	if err != nil {
		return nil, errors.Errorf("Could not get config for orderer: %s", err)
	}

	serverHostOverride := ""
	if str, ok := ordererConfig.GRPCOptions["ssl-target-name-override"].(string); ok {
		serverHostOverride = str
	}
	orderer, err := orderer.NewOrderer(
		ordererConfig.URL,
		ordererConfig.TLSCACerts.Path, serverHostOverride, cliconfig.Config())
	if err != nil {
		return nil, errors.Errorf("NewOrderer return error: %v", err)
	}

	return orderer, nil
}

func levelFromName(levelName string) logging.Level {
	switch levelName {
	case "ERROR":
		return logging.ERROR
	case "WARNING":
		return logging.WARNING
	case "INFO":
		return logging.INFO
	case "DEBUG":
		return logging.DEBUG
	default:
		return logging.ERROR
	}
}

func getPeers(allPeers []apifabclient.Peer, peerURLs string) ([]apifabclient.Peer, error) {
	if peerURLs == "" {
		return nil, nil
	}

	s := strings.Split(peerURLs, ",")

	var selectedPeers []apifabclient.Peer
	for _, peer := range allPeers {
		if containsString(s, peer.URL()) {
			selectedPeers = append(selectedPeers, peer)
		}
	}
	if len(selectedPeers) != len(s) {
		return nil, errors.Errorf("one or more peers is invalid: %s", peerURLs)
	}
	return selectedPeers, nil
}

func (action *Action) getEventHub() (apifabclient.EventHub, error) {
	fabClient, err := action.Client()
	if err != nil {
		return nil, errors.Errorf("error getting fabric client: %s", err)
	}

	eventHub, err := events.NewEventHub(fabClient)
	if err != nil {
		return nil, errors.Errorf("Error creating new event hub: %v", err)
	}

	peerConfig, err := action.peerConfig()
	if err != nil {
		return nil, err
	}

	cliconfig.Config().Logger().Infof("Connecting to event hub at %s ...\n", peerConfig.URL)

	serverHostOverride := ""
	if str, ok := peerConfig.GRPCOptions["ssl-target-name-override"].(string); ok {
		serverHostOverride = str
	}

	eventHub.SetPeerAddr(peerConfig.EventURL, peerConfig.TLSCACerts.Path, serverHostOverride)

	return eventHub, nil
}

func (action *Action) peerConfig() (*apiconfig.PeerConfig, error) {
	peersConfig, err := cliconfig.Config().PeersConfig(action.OrgID())
	if err != nil {
		return nil, errors.Errorf("Error reading peers config for %s: %v", action.OrgID(), err)
	}

	peer := action.Peer()

	for _, p := range peersConfig {
		if peer.URL() == "" || p.URL == peer.URL() {
			return &p, nil
		}
	}

	return nil, errors.Errorf("No configuration found for peer %s", peer.URL())
}

func (action *Action) newEventHub() (apifabclient.EventHub, error) {
	cliconfig.Config().Logger().Debugf("initEventHub - Initializing...\n")

	eventHub, err := action.getEventHub()
	if err != nil {
		return nil, errors.Errorf("unable to get event hub: %s", err)
	}

	// // Set the user to the org admin since the 'register' message must be signed by a user in the peer's MSP
	// context := action.SetUserContext(action.OrgAdminUser(action.OrgID()))
	// defer context.Restore()

	if err := eventHub.Connect(); err != nil {
		return nil, errors.Errorf("unable to connect to event hub: %s", err)
	}

	return eventHub, nil
}

func (action *Action) newSession(orgID string, user apifabclient.User) (context.Session, error) {
	cliconfig.Config().Logger().Debugf("... got user [%s]. Creating context for org [%s]...\n", user.Name(), orgID)
	context, err := action.sdk.NewContext(orgID)
	if err != nil {
		return nil, errors.Errorf("Error getting a context for org: %s", err)
	}

	cliconfig.Config().Logger().Debugf("... created context for org [%s]. Creating session for org [%s], user [%s]...\n", orgID, orgID, user.Name())
	session, err := action.sdk.NewSession(context, user)
	if err != nil {
		return nil, errors.Errorf("NewSession returned error: %v", err)
	}
	cliconfig.Config().Logger().Debugf("... successfully created session for org [%s], user [%s].\n", orgID, user.Name())

	return session, nil
}

func containsString(sarr []string, s string) bool {
	for _, str := range sarr {
		if s == str {
			return true
		}
	}
	return false
}
