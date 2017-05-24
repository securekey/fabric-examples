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
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-sdk-go/config"
	fabricClient "github.com/hyperledger/fabric-sdk-go/fabric-client"
	"github.com/hyperledger/fabric-sdk-go/fabric-client/events"
	fcUtil "github.com/hyperledger/fabric-sdk-go/fabric-client/helpers"
	bccspFactory "github.com/hyperledger/fabric/bccsp/factory"
	logging "github.com/op/go-logging"
	"github.com/spf13/pflag"
)

const (
	// PeerFlag is the flag that specifies the URL of the peer to connect to
	PeerFlag = "peer"

	// OrdererFlag is the flag that specifies the URL of the orderer to connect to
	OrdererFlag = "orderer"

	// ChannelIDFlag is the flag that specifies the ID of the channel
	ChannelIDFlag = "cid"

	// ChaincodeIDFlag is the flag that specifies the ID of the chaincode
	ChaincodeIDFlag = "ccid"

	// ChaincodePathFlag is the flag that specifies the path of the chaincode relative to the GO src directory
	ChaincodePathFlag = "ccp"

	// ChaincodeVersionFlag is the flag that specifies the version of the chaincode
	ChaincodeVersionFlag = "v"

	// ArgsFlag is the flag that specifies a JSON string containing the chaincode arguments.
	// For example, --args='{"Args":"function","arg1","arg2"}'
	ArgsFlag = "args"

	// PrettyPrintFlag is the flag that specifies whether or not to print out blocks in a human-readable way
	PrettyPrintFlag = "prettyprint"

	// IterationsFlag is the flag that specifies the number of times to perform a task (such as invoke chaincode)
	IterationsFlag = "iterations"

	// SleepFlag is the flag that specifies the number of milliseconds to sleep between invocations
	SleepFlag = "sleep"

	// Private constants ...
	defaultUser         = "admin"
	defaultPassword     = "adminpw"
	userStatePath       = "/tmp/enroll_user"
	defaultCCVersion    = "v0"
	defaultLoggingLevel = "CRITICAL"
	defaultChannelID    = "testchannel"
	loggerName          = "fabriccli"
	defaultConfigFile   = "fixtures/config/config_test.yaml"
	defaultCertificate  = "fixtures/tls/orderer/ca-cert.pem"
)

var Logger = logging.MustGetLogger(loggerName)
var User = defaultUser
var Password = defaultPassword
var LoggingLevel = defaultLoggingLevel
var ChannelID = defaultChannelID
var ChaincodeID string
var ChaincodePath string
var ChaincodeVersion = defaultCCVersion
var OrdererURL string
var Args = getEmptyArgs()
var PrettyPrint = true
var Iterations = 1
var SleepTime int64
var ConfigFile = defaultConfigFile
var Certificate = defaultCertificate
var PrintFormat string

// ArgStruct is used for marshalling arguments to chaincode invocations
type ArgStruct struct {
	Args []string `json:"Args"`
}

// Action is implemented by all command actions
type Action interface {
	Initialize(flags *pflag.FlagSet) error
	Flags() *pflag.FlagSet
	Invoke() error
	EventHub() events.EventHub
	Peers() []fabricClient.Peer
	Client() fabricClient.Client
	Printer() Printer
}

// ActionImpl is the base implementation of the Action interface.
type ActionImpl struct {
	flags    *pflag.FlagSet
	eventHub events.EventHub
	peers    []fabricClient.Peer
	client   fabricClient.Client
	printer  Printer
}

// Initialize initializes the action using the given flags
func (action *ActionImpl) Initialize(flags *pflag.FlagSet) error {
	if err := config.InitConfig(ConfigFile); err != nil {
		return err
	}

	level := levelFromName(LoggingLevel)
	Logger.Infof("******* Setting logging level of %s to %s, %d\n", loggerName, LoggingLevel, level)
	logging.SetLevel(level, loggerName)

	// Initialize bccsp factories before calling get client
	if err := bccspFactory.InitFactories(&bccspFactory.FactoryOpts{
		ProviderName: "SW",
		SwOpts: &bccspFactory.SwOpts{
			HashFamily: config.GetSecurityAlgorithm(),
			SecLevel:   config.GetSecurityLevel(),
			FileKeystore: &bccspFactory.FileKeystoreOpts{
				KeyStorePath: config.GetKeyStorePath(),
			},
			Ephemeral: false,
		},
	}); err != nil {
		return fmt.Errorf("Failed getting ephemeral software-based BCCSP [%s]", err)
	}

	client, err := fcUtil.GetClient(User, Password, userStatePath)
	if err != nil {
		return fmt.Errorf("Create client failed: %v", err)
	}

	peersConfig, err := config.GetPeersConfig()
	if err != nil {
		return fmt.Errorf("Error getting peer configs: %v", err)
	}

	var allPeers []fabricClient.Peer
	for _, p := range peersConfig {
		endorser, err := fabricClient.NewPeer(fmt.Sprintf("%s:%d", p.Host, p.Port),
			p.TLS.Certificate, p.TLS.ServerHostOverride)
		if err != nil {
			return fmt.Errorf("NewPeer return error: %v", err)
		}
		allPeers = append(allPeers, endorser)
	}

	peerURL, _ := flags.GetString(PeerFlag)

	eventHub, err := getEventHub(peerURL)
	if err != nil {
		return err
	}

	if err := eventHub.Connect(); err != nil {
		return fmt.Errorf("Failed eventHub.Connect() [%s]", err)
	}

	for i, peer := range allPeers {
		Logger.Debugf("Peer[%d]: Name: %s, URL: %s\n", i, peer.GetName(), peer.GetURL())
	}

	var peers []fabricClient.Peer
	if peerURL != "" {
		peers, err = getPeers(allPeers, peerURL)
		if err != nil {
			return err
		}
	} else {
		peers = allPeers
	}

	action.flags = flags
	action.client = client
	action.eventHub = eventHub
	action.peers = peers
	action.printer = NewPrinter(AsOutputFormat(PrintFormat))

	return nil
}

// Flags returns the flag-set
func (action *ActionImpl) Flags() *pflag.FlagSet {
	return action.flags
}

// NewChain creates a new Chain
func (action *ActionImpl) NewChain() (fabricClient.Chain, error) {
	serverHostOverride := "orderer0"

	orderer, err := fabricClient.NewOrderer(OrdererURL, Certificate, serverHostOverride)
	if err != nil {
		return nil, fmt.Errorf("JoinNewOrderer return error: %v", err)
	}

	chain, err := fabricClient.NewChain(ChannelID, action.client)
	if err != nil {
		return nil, fmt.Errorf("Could not get chain: %v", err)
	}

	chain.AddOrderer(orderer)
	for _, peer := range action.peers {
		chain.AddPeer(peer)
	}

	if err := chain.Initialize(nil); err != nil {
		return nil, fmt.Errorf("Error initializing chain: %v", err)
	}

	return chain, err
}

// Printer returns the Printer
func (action *ActionImpl) Printer() Printer {
	return action.printer
}

// EventHub returns the event hub
func (action *ActionImpl) EventHub() events.EventHub {
	return action.eventHub
}

// Peers returns the peers
func (action *ActionImpl) Peers() []fabricClient.Peer {
	return action.peers
}

// Client returns the Fabric client
func (action *ActionImpl) Client() fabricClient.Client {
	return action.client
}

// PeerFromURL returns the peer for the given URL
func (action *ActionImpl) PeerFromURL(url string) fabricClient.Peer {
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

func getPeers(allPeers []fabricClient.Peer, peerURL string) ([]fabricClient.Peer, error) {
	if peerURL == "" {
		return allPeers, nil
	}

	var selectedPeer fabricClient.Peer
	for _, peer := range allPeers {
		if peer.GetURL() == peerURL {
			selectedPeer = peer
			break
		}
	}
	if selectedPeer == nil {
		return nil, fmt.Errorf("Peer not found for URL: %s", peerURL)
	}

	return []fabricClient.Peer{selectedPeer}, nil
}

func getEventHub(peerURL string) (events.EventHub, error) {
	eventHub := events.NewEventHub()
	foundEventHub := false
	peersConfig, err := config.GetPeersConfig()
	if err != nil {
		return nil, err
	}

	for _, p := range peersConfig {
		if p.EventHost != "" {
			url := fmt.Sprintf("%s:%d", p.Host, p.Port)
			if peerURL == "" || peerURL == url {
				fmt.Printf("Connecting to URL (%s:%d)\n", p.EventHost, p.EventPort)
				eventHub.SetPeerAddr(fmt.Sprintf("%s:%d", p.EventHost, p.EventPort), p.TLS.Certificate, p.TLS.ServerHostOverride)
				foundEventHub = true
				break
			}
		}
	}

	if !foundEventHub {
		return nil, fmt.Errorf("No EventHub configuration found")
	}

	return eventHub, nil
}

func getEmptyArgs() string {
	argBytes, err := json.Marshal(&ArgStruct{})
	if err != nil {
		panic(fmt.Errorf("error marshaling empty args struct: %v", err))
	}
	return string(argBytes)
}
