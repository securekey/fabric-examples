/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package action

import (
	"encoding/json"
	"math/rand"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/factory/defcore"

	"github.com/pkg/errors"

	"strings"

	"io"

	"fmt"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/event"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	mspapi "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	contextImpl "github.com/hyperledger/fabric-sdk-go/pkg/context"
	cryptosuiteimpl "github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite/bccsp/multisuite"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/orderer"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	cliconfig "github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/config"
	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/printer"
	"github.com/spf13/pflag"
)

const (
	defaultUser = "User1" // pre-enrolled user
	adminUser   = "Admin"
)

// ArgStruct is used for marshalling arguments to chaincode invocations
type ArgStruct struct {
	Func string   `json:"Func"`
	Args []string `json:"Args"`
}

// Action is the base implementation of the Action interface.
type Action struct {
	flags          *pflag.FlagSet
	sdk            *fabsdk.FabricSDK
	endpointConfig fab.EndpointConfig
	peersByOrg     map[string][]fab.Peer
	peers          []fab.Peer
	orgIDByPeer    map[string]string
	printer        printer.Printer
	initError      error
	Writer         io.Writer
	sessions       map[string]context.ClientProvider
}

// Initialize initializes the action using the given flags
func (action *Action) Initialize(flags *pflag.FlagSet) error {

	action.sessions = make(map[string]context.ClientProvider)
	action.flags = flags

	if err := cliconfig.InitConfig(flags); err != nil {
		return err
	}

	var opts []fabsdk.Option
	if cliconfig.Config().SelectionProvider() != cliconfig.AutoDetectSelectionProvider {
		svcPackage, err := newServiceProviderFactory()
		if err != nil {
			return err
		}
		opts = append(opts, fabsdk.WithServicePkg(svcPackage))
	}
	opts = append(opts, fabsdk.WithCorePkg(&cryptoSuiteProviderFactory{}))

	sdk, err := fabsdk.New(cliconfig.Provider(), opts...)
	if err != nil {
		return errors.Errorf("Error initializing SDK: %s", err)
	}
	action.sdk = sdk

	ctx, err := sdk.Context()()
	if err != nil {
		return errors.WithMessage(err, "Error creating anonymous provider")
	}

	action.endpointConfig = ctx.EndpointConfig()

	networkConfig := action.endpointConfig.NetworkConfig()

	level := levelFromName(cliconfig.Config().LoggingLevel())

	logging.SetLevel("", level)

	action.orgIDByPeer = make(map[string]string)

	var allPeers []fab.Peer
	allPeersByOrg := make(map[string][]fab.Peer)
	for orgID := range networkConfig.Organizations {
		peersConfig, ok := action.endpointConfig.PeersConfig(orgID)
		if !ok {
			return errors.Errorf("failed to get peer configs for org [%s]", orgID)
		}

		cliconfig.Config().Logger().Debugf("Peers for org [%s]: %v\n", orgID, peersConfig)

		var peers []fab.Peer
		for _, p := range peersConfig {
			endorser, err := ctx.InfraProvider().CreatePeerFromConfig(&fab.NetworkPeer{PeerConfig: p})
			if err != nil {
				return errors.Wrapf(err, "failed to create peer from config")
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
				cliconfig.Config().Logger().Debugf("-- Peer[%d]: MSPID: %s, URL: %s\n", i, peer.MSPID(), peer.URL())
			}
		}
	}

	// Filter peers by specified peers/orgs
	peers, err := action.getPeers(allPeers, cliconfig.Config().PeerURLs(), cliconfig.Config().OrgIDs())
	if err != nil {
		return err
	}

	// Organize peers by orgs
	peersByOrg := make(map[string][]fab.Peer)
	cliconfig.Config().Logger().Debugf("Selected Peers:\n")
	for _, peer := range peers {
		cliconfig.Config().Logger().Debugf("- URL: %s\n", peer.URL())
		orgID := action.orgIDByPeer[peer.URL()]
		if orgID == "" {
			return errors.Errorf("unable to find org for peer: %s", peer.URL())
		}
		peersByOrg[orgID] = append(peersByOrg[orgID], peer)
	}

	action.peers = peers
	action.peersByOrg = peersByOrg

	action.printer = printer.NewBlockPrinterWithOpts(
		printer.AsOutputFormat(cliconfig.Config().PrintFormat()),
		printer.AsWriterType(cliconfig.Config().Writer()),
		&printer.FormatterOpts{Base64Encode: cliconfig.Config().Base64()})

	return nil
}

// Terminate closes any open connections. This function should be called at the end of every command invocation.
func (action *Action) Terminate() {
	if action.sdk != nil {
		cliconfig.Config().Logger().Info("Closing SDK")
		action.sdk.Close()
	}
}

// Flags returns the flag-set
func (action *Action) Flags() *pflag.FlagSet {
	return action.flags
}

// EndpointConfig returns the endpoint configuration
func (action *Action) EndpointConfig() fab.EndpointConfig {
	return action.endpointConfig
}

// ChannelClient creates a new channel client
func (action *Action) ChannelClient(...channel.ClientOption) (*channel.Client, error) {
	user, err := action.User()
	if err != nil {
		return nil, errors.Errorf("error getting user: %s", err)
	}
	session, err := action.context(user)
	if err != nil {
		return nil, errors.Errorf("error getting session for user [%s,%s]: %v", user.Identifier().MSPID, user.Identifier().ID, err)
	}
	channelProvider := func() (context.Channel, error) {
		return contextImpl.NewChannel(session, cliconfig.Config().ChannelID())
	}
	return channel.New(channelProvider)
}

// OrgAdminChannelClient creates a new channel client for the given org in order to perform administrative functions
func (action *Action) OrgAdminChannelClient(orgID string) (*channel.Client, error) {
	channelID := cliconfig.Config().ChannelID()
	cliconfig.Config().Logger().Debugf("Creating new channel client for channel [%s] and org [%s] ...", channelID, orgID)

	user, err := action.OrgAdminUser(orgID)
	if err != nil {
		return nil, err
	}

	channelClient, err := action.ClientForUser(channelID, user)
	if err != nil {
		return nil, errors.Errorf("error creating fabric client: %s", err)
	}

	return channelClient, nil
}

// AdminChannelClient creates a new channel client for performing administrative functions
func (action *Action) AdminChannelClient() (*channel.Client, error) {
	return action.OrgAdminChannelClient(action.OrgID())
}

// Printer returns the Printer
func (action *Action) Printer() printer.Printer {
	return action.printer
}

// LocalContext creates a new local context
func (action *Action) LocalContext() (context.Local, error) {
	user, err := action.User()
	if err != nil {
		return nil, errors.Errorf("error getting user: %s", err)
	}
	contextProvider, err := action.context(user)
	if err != nil {
		return nil, errors.Errorf("error getting context for user [%s,%s]: %v", user.Identifier().MSPID, user.Identifier().ID, err)
	}
	return contextImpl.NewLocal(contextProvider)
}

// ChannelProvider returns the ChannelProvider
func (action *Action) ChannelProvider() (context.ChannelProvider, error) {
	channelID := cliconfig.Config().ChannelID()
	user, err := action.User()
	if err != nil {
		return nil, err
	}
	cliconfig.Config().Logger().Debugf("creating channel provider for user [%s] in org [%s]...", user.Identifier().ID, user.Identifier().MSPID)
	clientContext, err := action.context(user)
	if err != nil {
		return nil, errors.Errorf("error getting client context for user [%s,%s]: %v", user.Identifier().MSPID, user.Identifier().ID, err)
	}
	channelProvider := func() (context.Channel, error) {
		return contextImpl.NewChannel(clientContext, channelID)
	}
	return channelProvider, nil
}

// EventClient returns the event hub.
func (action *Action) EventClient(opts ...event.ClientOption) (*event.Client, error) {
	channelProvider, err := action.ChannelProvider()
	if err != nil {
		return nil, errors.Errorf("error creating channel provider: %v", err)
	}
	c, err := event.New(channelProvider, opts...)
	if err != nil {
		return nil, errors.Errorf("error creating new event client: %v", err)
	}
	return c, nil
}

// LedgerClient returns the Fabric client for the current user
func (action *Action) LedgerClient() (*ledger.Client, error) {
	channelProvider, err := action.ChannelProvider()
	if err != nil {
		return nil, errors.Errorf("error creating channel provider: %v", err)
	}
	c, err := ledger.New(channelProvider)
	if err != nil {
		return nil, errors.Errorf("error creating new ledger client: %v", err)
	}
	return c, nil
}

// Peer returns the first peer in the list of selected peers
func (action *Action) Peer() fab.Peer {
	if len(action.peers) == 0 {
		return nil
	}
	return action.peers[0]
}

// Peers returns the peers
func (action *Action) Peers() []fab.Peer {
	return action.peers
}

// PeersByOrg returns the peers mapped by organization
func (action *Action) PeersByOrg() map[string][]fab.Peer {
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
func (action *Action) Client(channelID string) (*channel.Client, error) {
	user, err := action.User()
	if err != nil {
		return nil, err
	}
	return action.ClientForUser(channelID, user)
}

// ResourceMgmtClient returns a resource management client for the current user
func (action *Action) ResourceMgmtClient() (*resmgmt.Client, error) {
	return action.ResourceMgmtClientForOrg(action.OrgID())
}

// ResourceMgmtClientForOrg returns a resource management client for the given org
func (action *Action) ResourceMgmtClientForOrg(orgID string) (*resmgmt.Client, error) {
	user, err := action.OrgAdminUser(orgID)
	if err != nil {
		return nil, err
	}
	return action.ResourceMgmtClientForUser(user)
}

// ClientForUser returns the Channel client for the given user
func (action *Action) ClientForUser(channelID string, user mspapi.SigningIdentity) (*channel.Client, error) {
	cliconfig.Config().Logger().Debugf("create resmgmt client for user [%s] in org [%s]...", user.Identifier().ID, user.Identifier().MSPID)
	session, err := action.context(user)
	if err != nil {
		return nil, errors.Errorf("error getting session for user [%s,%s]: %v", user.Identifier().MSPID, user.Identifier().ID, err)
	}
	channelProvider := func() (context.Channel, error) {
		return contextImpl.NewChannel(session, channelID)
	}
	c, err := channel.New(channelProvider)
	if err != nil {
		return nil, errors.Errorf("error creating new resmgmt client for user [%s,%s]: %v", user.Identifier().MSPID, user.Identifier().ID, err)
	}
	return c, nil
}

// ResourceMgmtClientForUser returns the Fabric client for the given user
func (action *Action) ResourceMgmtClientForUser(user mspapi.SigningIdentity) (*resmgmt.Client, error) {
	cliconfig.Config().Logger().Debugf("create resmgmt client for user [%s] in org [%s]...", user.Identifier().ID, user.Identifier().MSPID)
	session, err := action.context(user)
	if err != nil {
		return nil, errors.Errorf("error getting session for user [%s,%s]: %v", user.Identifier().MSPID, user.Identifier().ID, err)
	}
	c, err := resmgmt.New(session)
	if err != nil {
		return nil, errors.Errorf("error creating new resmgmt client for user [%s,%s]: %v", user.Identifier().MSPID, user.Identifier().ID, err)
	}
	return c, nil
}

// ChannelMgmtClientForUser returns the Fabric client for the given user
func (action *Action) ChannelMgmtClientForUser(channelID string, user mspapi.SigningIdentity) (*channel.Client, error) {
	cliconfig.Config().Logger().Debugf("create channel client for user [%s] in org [%s]...", user.Identifier().ID, user.Identifier().MSPID)
	session, err := action.context(user)
	if err != nil {
		return nil, errors.Errorf("error getting session for user [%s,%s]: %v", user.Identifier().MSPID, user.Identifier().ID, err)
	}
	channelProvider := func() (context.Channel, error) {
		return contextImpl.NewChannel(session, channelID)
	}
	c, err := channel.New(channelProvider)
	if err != nil {
		return nil, errors.Errorf("error creating new channel client for user [%s,%s]: %v", user.Identifier().MSPID, user.Identifier().ID, err)
	}
	return c, nil
}

func (action *Action) context(user mspapi.SigningIdentity) (context.ClientProvider, error) {
	key := user.Identifier().MSPID + "_" + user.Identifier().ID
	session := action.sessions[key]
	if session == nil {
		session = action.sdk.Context(fabsdk.WithIdentity(user))
		cliconfig.Config().Logger().Debugf("Created session for user [%s] in org [%s]", user.Identifier().ID, user.Identifier().MSPID)
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
	networkConfig := action.endpointConfig.NetworkConfig()

	for orgID, orgConfig := range networkConfig.Organizations {
		if mspID == orgConfig.MSPID {
			return orgID, nil
		}
	}
	return "", errors.Errorf("unable to find org ID for MSP [%s]", mspID)
}

// User returns the enrolled user. If the user doesn't exist then a new user is enrolled.
func (action *Action) User() (mspapi.SigningIdentity, error) {
	userName := cliconfig.Config().UserName()
	if userName == "" {
		userName = defaultUser
	}
	return action.OrgUser(action.OrgID(), userName)
}

func (action *Action) newUser(orgID, username, pwd string) (mspapi.SigningIdentity, error) {

	cliconfig.Config().Logger().Infof("Enrolling user %s...\n", username)

	mspClient, err := msp.New(action.sdk.Context(), msp.WithOrg(orgID))
	if err != nil {
		return nil, errors.Errorf("error creating MSP client: %s", err)
	}

	cliconfig.Config().Logger().Infof("Creating new user %s...\n", username)
	err = mspClient.Enroll(username, msp.WithSecret(pwd))
	if err != nil {
		return nil, errors.Errorf("Enroll returned error: %v", err)
	}

	user, err := mspClient.GetSigningIdentity(username)
	if err != nil {
		return nil, errors.Errorf("GetSigningIdentity returned error: %v", err)
	}

	cliconfig.Config().Logger().Infof("Returning user [%s], MSPID [%s]\n", user.Identifier().ID, user.Identifier().MSPID)

	return user, nil
}

// OrgUser returns an already enrolled user for the given organization
func (action *Action) OrgUser(orgID, username string) (mspapi.SigningIdentity, error) {
	if username == "" {
		return nil, errors.Errorf("no username specified")
	}
	mspClient, err := msp.New(action.sdk.Context(), msp.WithOrg(orgID))
	if err != nil {
		return nil, errors.Errorf("error creating MSP client: %s", err)
	}

	user, err := mspClient.GetSigningIdentity(username)
	if err != nil {
		return nil, errors.Errorf("GetSigningIdentity returned error: %v", err)
	}

	cliconfig.Config().Logger().Infof("Returning user [%s], MSPID [%s]\n", user.Identifier().ID, user.Identifier().MSPID)

	return user, nil
}

// OrgAdminUser returns the pre-enrolled administrative user for the given organization
func (action *Action) OrgAdminUser(orgID string) (mspapi.SigningIdentity, error) {
	userName := cliconfig.Config().UserName()
	if userName == "" {
		userName = adminUser
	}
	return action.OrgUser(orgID, userName)
}

// PeerFromURL returns the peer for the given URL
func (action *Action) PeerFromURL(url string) (fab.Peer, bool) {
	for _, peer := range action.peers {
		if url == peer.URL() {
			return peer, true
		}
	}
	return nil, false
}

// Orderers returns all Orderers from the set of configured Orderers
func (action *Action) Orderers() ([]fab.Orderer, error) {
	ordererConfigs := action.endpointConfig.OrderersConfig()
	ordererURL := cliconfig.Config().OrdererURL()

	var orderers []fab.Orderer
	for _, ordererConfig := range ordererConfigs {
		if ordererURL == "" || ordererConfig.URL == ordererURL {
			newOrderer, err := orderer.New(action.endpointConfig, orderer.FromOrdererConfig(&ordererConfig))
			if err != nil {
				return nil, errors.WithMessage(err, "creating orderer failed")
			}
			orderers = append(orderers, newOrderer)
		}
	}

	return orderers, nil
}

// RandomOrderer chooses a random Orderer from the set of configured Orderers
func (action *Action) RandomOrderer() (fab.Orderer, error) {
	orderers, err := action.Orderers()
	if err != nil {
		return nil, err
	}
	if len(orderers) == 0 {
		return nil, errors.New("No orders found")
	}
	return orderers[rand.Intn(len(orderers))], nil
}

// ArgsArray returns an array of args used in chaincode invocations
func ArgsArray() ([]ArgStruct, error) {
	var argsArray []ArgStruct
	argBytes := []byte(cliconfig.Config().Args())
	if strings.HasPrefix(cliconfig.Config().Args(), "[") {
		if err := json.Unmarshal(argBytes, &argsArray); err != nil {
			return nil, errors.Errorf("Error unmarshaling JSON arg string: %v", err)
		}
	} else {
		args := ArgStruct{}
		if err := json.Unmarshal(argBytes, &args); err != nil {
			return nil, errors.Errorf("Error unmarshaling JSON arg string: %v", err)
		}
		argsArray = append(argsArray, args)
	}
	return argsArray, nil
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
		return logging.ERROR
	}
}

func (action *Action) getPeers(allPeers []fab.Peer, peerURLs []string, orgIDs []string) ([]fab.Peer, error) {
	selectAll := false
	if len(peerURLs) == 0 && len(orgIDs) == 0 {
		selectAll = true
	}
	var selectedPeers []fab.Peer
	var allPeerURLs []string
	for _, peer := range allPeers {
		allPeerURLs = append(allPeerURLs, peer.URL())
		orgID := action.orgIDByPeer[peer.URL()]
		if selectAll || containsString(peerURLs, peer.URL()) || len(peerURLs) == 0 && containsString(orgIDs, orgID) {
			selectedPeers = append(selectedPeers, peer)
		}
	}
	for _, url := range peerURLs {
		if !containsString(allPeerURLs, url) {
			return nil, fmt.Errorf("invalid peer URL: %s", url)
		}
	}
	return selectedPeers, nil
}

// PeerConfig returns the PeerConfig for the first peer in the current org
func (action *Action) PeerConfig() (*fab.PeerConfig, error) {
	peersConfig, ok := action.endpointConfig.PeersConfig(action.OrgID())
	if !ok {
		return nil, errors.Errorf("Error reading peers config for %s: %v", action.OrgID())
	}

	peer := action.Peer()

	for _, p := range peersConfig {
		if peer.URL() == "" || p.URL == peer.URL() {
			return &p, nil
		}
	}

	return nil, errors.Errorf("No configuration found for peer %s", peer.URL())
}

// CreateDiscoveryService returns a new DiscoveryService for the given channel.
// This is an implementation of the DiscoveryProvider interface
func (action *Action) CreateDiscoveryService(channelID string) (fab.DiscoveryService, error) {
	return action, nil
}

// GetPeers returns the peers in context.
// This is an implementation of the DiscoveryService interface
func (action *Action) GetPeers() ([]fab.Peer, error) {
	return action.Peers(), nil
}

func containsString(sarr []string, s string) bool {
	for _, str := range sarr {
		if s == str {
			return true
		}
	}
	return false
}

// cryptoSuiteProviderFactory will provide custom cryptosuite (bccsp.BCCSP)
type cryptoSuiteProviderFactory struct {
	defcore.ProviderFactory
}

// CreateCryptoSuiteProvider returns a new default implementation of BCCSP
func (f *cryptoSuiteProviderFactory) CreateCryptoSuiteProvider(config core.CryptoSuiteConfig) (core.CryptoSuite, error) {
	return cryptosuiteimpl.GetSuiteByConfig(config)
}
