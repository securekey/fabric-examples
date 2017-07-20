/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package common

import (
	"crypto/x509"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"strings"

	"sync"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	bccspFactory "github.com/hyperledger/fabric/bccsp/factory"
	logging "github.com/op/go-logging"
	"github.com/spf13/pflag"
)

const (
	loggerName    = "fabriccli"
	userStatePath = "/tmp/enroll_user"
)

// Flags
const (
	userFlag        = "user"
	userDescription = "The enrollment user"
	defaultUser     = "admin"

	passwordFlag        = "pw"
	passwordDescription = "The password of the enrollment user"
	defaultPassword     = "adminpw"

	chaincodeVersionFlag        = "v"
	chaincodeVersionDescription = "The chaincode version"
	defaultChaincodeVersion     = "v0"

	loggingLevelFlag        = "logging-level"
	loggingLevelDescription = "Logging level - CRITICAL, ERROR, WARNING, INFO, DEBUG"
	defaultLoggingLevel     = "CRITICAL"

	orgIDsFlag        = "orgid"
	orgIDsDescription = "A comma-separated list of organization IDs"
	defaultOrgIDs     = "org1,org2"

	channelIDFlag        = "cid"
	channelIDDescription = "The channel ID"
	defaultChannelID     = "mychannel"

	chaincodeIDFlag        = "ccid"
	chaincodeIDDescription = "The Chaincode ID"
	defaultChaincodeID     = ""

	chaincodePathFlag        = "ccp"
	chaincodePathDescription = "The chaincode path"
	defaultChaincodePath     = ""

	configFileFlag        = "config"
	configFileDescription = "The path of the config.yaml file"
	defaultConfigFile     = "fixtures/config/config_test.yaml"

	peerURLFlag        = "peer"
	peerURLDescription = "A comma-separated list of peer targets, e.g. 'localhost:7051,localhost:8051'"
	defaultPeerURL     = ""

	ordererFlag           = "orderer"
	ordererURLDescription = "The URL of the orderer, e.g. localhost:7050"
	defaultOrdererURL     = ""

	printFormatFlag        = "format"
	printFormatDescription = "The output format - display, json, raw"

	writerFlag        = "writer"
	writerDescription = "The writer - stdout, stderr, log"

	certificateFileFlag    = "cacert"
	certificateDescription = "The path of the ca-cert.pem file"
	defaultCertificate     = ""

	argsFlag        = "args"
	argsDescription = "The args in JSON format. Example: {\"Func\":\"function\",\"Args\":[\"arg1\",\"arg2\"]}"

	iterationsFlag        = "iterations"
	iterationsDescription = "The number of times to invoke the chaincode"
	defaultIterations     = "1"

	sleepFlag            = "sleep"
	sleepTimeDescription = "The number of milliseconds to sleep between invocations of the chaincode."
	defaultSleepTime     = "100"

	txFileFlag        = "txfile"
	txFileDescription = "The path of the channel.tx file"
	defaultTxFile     = "fixtures/channel/mychannel.tx"

	chaincodeEventFlag        = "event"
	chaincodeEventDescription = "The name of the chaincode event to listen for"
	defaultChaincodeEvent     = ""

	txIDFlag        = "txid"
	txIDDescription = "The transaction ID"
	defaultTxID     = ""

	blockNumFlag        = "num"
	blockNumDescription = "The block number"
	defaultBlockNum     = "-1"

	blockHashFlag        = "hash"
	blockHashDescription = "The block hash"
	defaultBlockHash     = ""

	traverseFlag        = "traverse"
	traverseDescription = "Blocks will be traversed starting with the given block in reverse order up to the given number of blocks"
	defaultTraverse     = "0"

	chaincodePolicyFlag        = "policy"
	chaincodePolicyDescription = "The chaincode policy, e.g. OR('Org1MSP.admin','Org2MSP.admin',AND('Org1MSP.member','Org2MSP.member'))"
	defaultChaincodePolicy     = ""

	timeoutFlag        = "timeout"
	timeoutDescription = "The timeout (in milliseconds) for the operation"
	defaultTimeout     = "3000"
)

var configInstance *cliConfig
var configInit sync.Once

// CLIConfig extendsthe fabric API config and provides additional configuration options
type CLIConfig interface {
	apiconfig.Config

	// Logger returns the Logger for the CLI tool
	Logger() *logging.Logger

	// LoggingLevel specifies the logging level (DEBUG, INFO, WARNING, ERROR, or CRITICAL)
	LoggingLevel() string
	InitLoggingLevel(flags *pflag.FlagSet, defaultValueAndDescription ...string)

	// ConfigFile specified the path of the configuration file
	ConfigFile() string
	InitConfigFile(flags *pflag.FlagSet, defaultValueAndDescription ...string)

	// OrgID specifies the ID of the current organization. If multiple org IDs are specified then the first one is returned.
	OrgID() string

	// OrgIDs returns a comma-separated list of organization IDs
	OrgIDs() []string
	InitOrgIDs(flags *pflag.FlagSet, defaultValueAndDescription ...string)

	// ChannelID returns the channel ID
	ChannelID() string
	InitChannelID(flags *pflag.FlagSet, defaultValueAndDescription ...string)

	// UserName returns the name of the enrolled user
	UserName() string
	InitUserName(flags *pflag.FlagSet, defaultValueAndDescription ...string)

	// UserPassword is the password to use when enrolling a user
	UserPassword() string
	InitUserPassword(flags *pflag.FlagSet, defaultValueAndDescription ...string)

	// ChaincodeID returns the chaicode ID
	ChaincodeID() string
	InitChaincodeID(flags *pflag.FlagSet, defaultValueAndDescription ...string)

	// ChaincodePath returns the source path of the chaincode to install/instantiate
	ChaincodePath() string
	InitChaincodePath(flags *pflag.FlagSet, defaultValueAndDescription ...string)

	// ChaincodeVersion returns the version of the chaincode
	ChaincodeVersion() string
	InitChaincodeVersion(flags *pflag.FlagSet, defaultValueAndDescription ...string)

	// ChaincodeEvent the name of the chaincode event to listen for
	ChaincodeEvent() string
	InitChaincodeEvent(flags *pflag.FlagSet, defaultValueAndDescription ...string)

	// PeerURL returns a comma-separated list of peers in the format host1:port1,host2:port2,...
	PeerURL() string
	InitPeerURL(flags *pflag.FlagSet, defaultValueAndDescription ...string)

	// OrdererURL returns the URL of the orderer
	OrdererURL() string
	InitOrdererURL(flags *pflag.FlagSet, defaultValueAndDescription ...string)

	// OrdererTLSCertificate is the path of the orderer TLS certificate
	OrdererTLSCertificate() string
	InitOrdererTLSCertificate(flags *pflag.FlagSet, defaultValueAndDescription ...string)

	// Iterations returns the number of times that a chaincode should be invoked
	Iterations() int
	InitIterations(flags *pflag.FlagSet, defaultValueAndDescription ...string)

	// SleepTime returns the number of milliseconds to sleep between invocations of a chaincode
	SleepTime() int64
	InitSleepTime(flags *pflag.FlagSet, defaultValueAndDescription ...string)

	// BlockNum returns the block number (where 0 is the first block)
	BlockNum() int
	InitBlockNum(flags *pflag.FlagSet, defaultValueAndDescription ...string)

	// BlockHash specifies the hash of the block
	BlockHash() string
	InitBlockHash(flags *pflag.FlagSet, defaultValueAndDescription ...string)

	// Traverse returns the number of blocks to traverse backwards in the query block command
	Traverse() int
	InitTraverse(flags *pflag.FlagSet, defaultValueAndDescription ...string)

	// PrintFormat returns the print (output) format for a block
	PrintFormat() string
	InitPrintFormat(flags *pflag.FlagSet, defaultValueAndDescription ...string)

	// Writer returns the writer for output
	Writer() string
	InitWriter(flags *pflag.FlagSet, defaultValueAndDescription ...string)

	// Args returns the chaincode invocation arguments as a JSON string in the format, {"Func":"function","Args":["arg1","arg2",...]}
	Args() string
	InitArgs(flags *pflag.FlagSet, defaultValueAndDescription ...string)

	// TxFile is the path of the .tx file used to create a channel
	TxFile() string
	InitTxFile(flags *pflag.FlagSet, defaultValueAndDescription ...string)

	// TxID returns the transaction ID
	TxID() string
	InitTxID(flags *pflag.FlagSet, defaultValueAndDescription ...string)

	// ChaincodePolicy returns the chaincode policy string, e.g Nof(1,(SignedBy(Org1Msp),SignedBy(Org2MSP)))
	ChaincodePolicy() string
	InitChaincodePolicy(flags *pflag.FlagSet, defaultValueAndDescription ...string)

	// Timeout returns the timeout (in milliseconds) for various operations
	Timeout() time.Duration
	InitTimeout(flags *pflag.FlagSet, defaultValueAndDescription ...string)
}

// cliConfig overrides certain configuration values with those supplied on the command-line
type cliConfig struct {
	config           apiconfig.Config
	logger           *logging.Logger
	certificate      string
	user             string
	password         string
	loggingLevel     string
	orgIDsStr        string
	channelID        string
	chaincodeID      string
	chaincodePath    string
	chaincodeVersion string
	peerURL          string
	ordererURL       string
	iterations       int
	sleepTime        int64
	configFile       string
	txFile           string
	txID             string
	printFormat      string
	writer           string
	args             string
	chaincodeEvent   string
	blockHash        string
	blockNum         int
	traverse         int
	chaincodePolicy  string
	timeout          int64
}

func getConfigImpl() *cliConfig {
	configInit.Do(func() {
		configInstance = &cliConfig{
			logger:           logging.MustGetLogger(loggerName),
			user:             defaultUser,
			password:         defaultPassword,
			loggingLevel:     defaultLoggingLevel,
			channelID:        defaultChannelID,
			orgIDsStr:        defaultOrgIDs,
			chaincodeVersion: defaultChaincodeVersion,
			iterations:       1,
			configFile:       defaultConfigFile,
			args:             getEmptyArgs(),
		}
	})
	return configInstance
}

// Config returns the CLI configuration
func Config() CLIConfig {
	return getConfigImpl()
}

// Implementation of CLIConfig ...

func (c *cliConfig) Logger() *logging.Logger {
	return c.logger
}

func (c *cliConfig) LoggingLevel() string {
	return c.loggingLevel
}

func (c *cliConfig) InitLoggingLevel(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(defaultLoggingLevel, loggingLevelDescription, defaultValueAndDescription...)
	flags.StringVar(&c.loggingLevel, loggingLevelFlag, defaultValue, description)
}

func (c *cliConfig) ConfigFile() string {
	return c.configFile
}

func (c *cliConfig) InitConfigFile(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(defaultConfigFile, configFileDescription, defaultValueAndDescription...)
	flags.StringVar(&c.configFile, configFileFlag, defaultValue, description)
}

func (c *cliConfig) OrgID() string {
	return c.OrgIDs()[0]
}

func (c *cliConfig) OrgIDs() []string {
	var orgIDs []string
	s := strings.Split(c.orgIDsStr, ",")
	for _, orgID := range s {
		orgIDs = append(orgIDs, orgID)
	}
	return orgIDs
}

func (c *cliConfig) InitOrgIDs(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(defaultOrgIDs, orgIDsDescription, defaultValueAndDescription...)
	flags.StringVar(&c.orgIDsStr, orgIDsFlag, defaultValue, description)
}

func (c *cliConfig) ChannelID() string {
	return c.channelID
}

func (c *cliConfig) InitChannelID(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(defaultChannelID, channelIDDescription, defaultValueAndDescription...)
	flags.StringVar(&c.channelID, channelIDFlag, defaultValue, description)
}

func (c *cliConfig) UserName() string {
	return c.user
}

func (c *cliConfig) InitUserName(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(defaultUser, userDescription, defaultValueAndDescription...)
	flags.StringVar(&c.user, userFlag, defaultValue, description)
}

func (c *cliConfig) UserPassword() string {
	return c.password
}

func (c *cliConfig) InitUserPassword(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(defaultPassword, passwordDescription, defaultValueAndDescription...)
	flags.StringVar(&c.password, passwordFlag, defaultValue, description)
}

func (c *cliConfig) ChaincodeID() string {
	return c.chaincodeID
}

func (c *cliConfig) InitChaincodeID(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(defaultChaincodeID, chaincodeIDDescription, defaultValueAndDescription...)
	flags.StringVar(&c.chaincodeID, chaincodeIDFlag, defaultValue, description)
}

func (c *cliConfig) ChaincodeEvent() string {
	return c.chaincodeEvent
}

func (c *cliConfig) InitChaincodeEvent(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(defaultChaincodeEvent, chaincodeEventDescription, defaultValueAndDescription...)
	flags.StringVar(&c.chaincodeEvent, chaincodeEventFlag, defaultValue, description)
}

func (c *cliConfig) ChaincodePath() string {
	return c.chaincodePath
}

func (c *cliConfig) InitChaincodePath(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(defaultChaincodePath, chaincodePathDescription, defaultValueAndDescription...)
	flags.StringVar(&c.chaincodePath, chaincodePathFlag, defaultValue, description)
}

func (c *cliConfig) ChaincodeVersion() string {
	return c.chaincodeVersion
}

func (c *cliConfig) InitChaincodeVersion(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(defaultChaincodeVersion, chaincodeVersionDescription, defaultValueAndDescription...)
	flags.StringVar(&c.chaincodeVersion, chaincodeVersionFlag, defaultValue, description)
}

func (c *cliConfig) PeerURL() string {
	return c.peerURL
}

func (c *cliConfig) InitPeerURL(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(defaultPeerURL, peerURLDescription, defaultValueAndDescription...)
	flags.StringVar(&c.peerURL, peerURLFlag, defaultValue, description)
}

func (c *cliConfig) OrdererURL() string {
	return c.ordererURL
}

func (c *cliConfig) InitOrdererURL(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(defaultOrdererURL, ordererURLDescription, defaultValueAndDescription...)
	flags.StringVar(&c.ordererURL, ordererFlag, defaultValue, description)
}

func (c *cliConfig) Iterations() int {
	return c.iterations
}

func (c *cliConfig) InitIterations(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(defaultIterations, iterationsDescription, defaultValueAndDescription...)
	i, err := strconv.Atoi(defaultValue)
	if err != nil {
		fmt.Printf("Invalid number: %s\n", defaultValue)
		i = 1
	}
	flags.IntVar(&c.iterations, iterationsFlag, i, description)
}

func (c *cliConfig) SleepTime() int64 {
	return c.sleepTime
}

func (c *cliConfig) InitSleepTime(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(defaultSleepTime, sleepTimeDescription, defaultValueAndDescription...)
	i, err := strconv.Atoi(defaultValue)
	if err != nil {
		fmt.Printf("Invalid number: %s\n", defaultValue)
		i = 1
	}
	flags.Int64Var(&c.sleepTime, sleepFlag, int64(i), description)
}

func (c *cliConfig) BlockNum() int {
	return c.blockNum
}

func (c *cliConfig) InitBlockNum(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(defaultBlockNum, blockNumDescription, defaultValueAndDescription...)
	i, err := strconv.Atoi(defaultValue)
	if err != nil {
		fmt.Printf("Invalid number: %s\n", defaultValue)
		i = 1
	}
	flags.IntVar(&c.blockNum, blockNumFlag, i, description)
}

func (c *cliConfig) BlockHash() string {
	return c.blockHash
}

func (c *cliConfig) InitBlockHash(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(defaultBlockHash, blockHashDescription, defaultValueAndDescription...)
	flags.StringVar(&c.blockHash, blockHashFlag, defaultValue, description)
}

func (c *cliConfig) Traverse() int {
	return c.traverse
}

func (c *cliConfig) InitTraverse(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(defaultTraverse, traverseDescription, defaultValueAndDescription...)
	i, err := strconv.Atoi(defaultValue)
	if err != nil {
		fmt.Printf("Invalid number: %s\n", defaultValue)
		i = 1
	}
	flags.IntVar(&c.traverse, traverseFlag, i, description)
}

func (c *cliConfig) PrintFormat() string {
	return c.printFormat
}

func (c *cliConfig) InitPrintFormat(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(DISPLAY.String(), printFormatDescription, defaultValueAndDescription...)
	flags.StringVar(&c.printFormat, printFormatFlag, defaultValue, description)
}

func (c *cliConfig) Writer() string {
	return c.writer
}

func (c *cliConfig) InitWriter(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(STDOUT.String(), writerDescription, defaultValueAndDescription...)
	flags.StringVar(&c.writer, writerFlag, defaultValue, description)
}

func (c *cliConfig) OrdererTLSCertificate() string {
	return c.certificate
}

func (c *cliConfig) InitOrdererTLSCertificate(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(defaultCertificate, certificateDescription, defaultValueAndDescription...)
	flags.StringVar(&c.certificate, certificateFileFlag, defaultValue, description)
}

func (c *cliConfig) Args() string {
	return c.args
}

func (c *cliConfig) InitArgs(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(getEmptyArgs(), argsDescription, defaultValueAndDescription...)
	flags.StringVar(&c.args, argsFlag, defaultValue, description)
}

func (c *cliConfig) TxFile() string {
	return c.txFile
}

func (c *cliConfig) InitTxFile(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(defaultTxFile, txFileDescription, defaultValueAndDescription...)
	flags.StringVar(&c.txFile, txFileFlag, defaultValue, description)
}

func (c *cliConfig) TxID() string {
	return c.txID
}

func (c *cliConfig) InitTxID(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(defaultTxID, txIDDescription, defaultValueAndDescription...)
	flags.StringVar(&c.txID, txIDFlag, defaultValue, description)
}

func (c *cliConfig) ChaincodePolicy() string {
	return c.chaincodePolicy
}

func (c *cliConfig) InitChaincodePolicy(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(defaultChaincodePolicy, chaincodePolicyDescription, defaultValueAndDescription...)
	flags.StringVar(&c.chaincodePolicy, chaincodePolicyFlag, defaultValue, description)
}

func (c *cliConfig) Timeout() time.Duration {
	return time.Duration(c.timeout)
}

func (c *cliConfig) InitTimeout(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(defaultTimeout, timeoutDescription, defaultValueAndDescription...)
	i, err := strconv.Atoi(defaultValue)
	if err != nil {
		fmt.Printf("Invalid number: %s\n", defaultValue)
		i = 1000
	}
	flags.Int64Var(&c.timeout, timeoutFlag, int64(i), description)
}

// Implementation of apifabclient.Config...

func (c *cliConfig) NetworkConfig() (*apiconfig.NetworkConfig, error) {
	return c.config.NetworkConfig()
}

func (c *cliConfig) CAConfig(org string) (*apiconfig.CAConfig, error) {
	return c.config.CAConfig(org)
}

func (c *cliConfig) MspID(org string) (string, error) {
	return c.config.MspID(org)
}

func (c *cliConfig) CAServerCertFiles(org string) ([]string, error) {
	return c.config.CAServerCertFiles(org)
}

func (c *cliConfig) CAClientKeyFile(org string) (string, error) {
	return c.config.CAClientKeyFile(org)
}

func (c *cliConfig) CAClientCertFile(org string) (string, error) {
	return c.config.CAClientCertFile(org)
}

func (c *cliConfig) IsTLSEnabled() bool {
	return c.config.IsTLSEnabled()
}

func (c *cliConfig) PeersConfig(org string) ([]apiconfig.PeerConfig, error) {
	return c.config.PeersConfig(org)
}

func (c *cliConfig) PeerConfig(org, name string) (*apiconfig.PeerConfig, error) {
	return c.config.PeerConfig(org, name)
}

func (c *cliConfig) TLSCACertPool(tlsCertificate string) (*x509.CertPool, error) {
	return c.config.TLSCACertPool(tlsCertificate)
}

func (c *cliConfig) SetTLSCACertPool(pool *x509.CertPool) {
	c.config.SetTLSCACertPool(pool)
}

func (c *cliConfig) IsSecurityEnabled() bool {
	return c.config.IsSecurityEnabled()
}

func (c *cliConfig) TcertBatchSize() int {
	return c.config.TcertBatchSize()
}

func (c *cliConfig) SecurityAlgorithm() string {
	return c.config.SecurityAlgorithm()
}

func (c *cliConfig) SecurityLevel() int {
	return c.config.SecurityLevel()
}

func (c *cliConfig) OrderersConfig() ([]apiconfig.OrdererConfig, error) {
	overridden := false

	configs, err := c.config.OrderersConfig()
	if err != nil {
		return nil, err
	}

	defaultConfig := configs[0]

	host := defaultConfig.Host
	port := defaultConfig.Port

	if c.OrdererURL() != "" {
		overridden = true
		s := strings.Split(c.OrdererURL(), ":")
		host = s[0]
		if len(s) > 1 {
			if p, err := strconv.Atoi(s[1]); err == nil {
				port = p
			} else {
				return nil, fmt.Errorf("invalid port %s: %s", s[1], err)
			}
		}
	}

	certificate := defaultConfig.TLS.Certificate

	if c.certificate != "" {
		overridden = true
		certificate = c.OrdererTLSCertificate()
	}

	if !overridden {
		return c.config.OrderersConfig()
	}

	return []apiconfig.OrdererConfig{
		apiconfig.OrdererConfig{
			Host: host,
			Port: port,
			TLS: apiconfig.TLSConfig{
				Certificate:        certificate,
				ServerHostOverride: defaultConfig.TLS.ServerHostOverride,
			},
		},
	}, nil
}

func (c *cliConfig) RandomOrdererConfig() (*apiconfig.OrdererConfig, error) {
	orderers, err := c.OrderersConfig()
	if err != nil {
		return nil, err
	}
	return &orderers[0], nil
}

func (c *cliConfig) OrdererConfig(name string) (*apiconfig.OrdererConfig, error) {
	return c.config.OrdererConfig(name)
}

func (c *cliConfig) KeyStorePath() string {
	return c.config.KeyStorePath()
}

func (c *cliConfig) CAKeyStorePath() string {
	return c.config.CAKeyStorePath()
}

func (c *cliConfig) CryptoConfigPath() string {
	return c.config.CryptoConfigPath()
}

func (c *cliConfig) CSPConfig() *bccspFactory.FactoryOpts {
	return c.config.CSPConfig()
}

func (c *cliConfig) TimeoutOrDefault(conn apiconfig.ConnectionType) time.Duration {
	return c.config.TimeoutOrDefault(conn)
}

// Utility functions...

func getEmptyArgs() string {
	argBytes, err := json.Marshal(&ArgStruct{})
	if err != nil {
		panic(fmt.Errorf("error marshaling empty args struct: %v", err))
	}
	return string(argBytes)
}

func getDefaultValueAndDescription(defaultValue string, defaultDescription string, overrides ...string) (value, description string) {
	if len(overrides) > 0 {
		value = overrides[0]
	} else {
		value = defaultValue
	}
	if len(overrides) > 1 {
		description = overrides[1]
	} else {
		description = defaultDescription
	}
	return value, description
}
