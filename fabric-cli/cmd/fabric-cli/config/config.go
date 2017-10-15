/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package config

import (
	"fmt"
	"strconv"
	"time"

	"strings"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/pkg/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
	"github.com/spf13/pflag"
)

const (
	loggerName    = "fabriccli"
	userStatePath = "/tmp/enroll_user"
)

// Flags
const (
	userFlag        = "user"
	userDescription = "The user"
	defaultUser     = ""

	passwordFlag        = "pw"
	passwordDescription = "The password of the user"
	defaultPassword     = ""

	chaincodeVersionFlag        = "v"
	chaincodeVersionDescription = "The chaincode version"
	defaultChaincodeVersion     = "v0"

	loggingLevelFlag        = "logging-level"
	loggingLevelDescription = "Logging level - ERROR, WARN, INFO, DEBUG"
	defaultLoggingLevel     = "ERROR"

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
	peerURLDescription = "A comma-separated list of peer targets, e.g. 'grpcs://localhost:7051,grpcs://localhost:8051'"
	defaultPeerURL     = ""

	ordererFlag           = "orderer"
	ordererURLDescription = "The URL of the orderer, e.g. grpcs://localhost:7050"
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

var opts *options
var instance *CLIConfig

type options struct {
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

func init() {
	opts = &options{
		user:             defaultUser,
		password:         defaultPassword,
		loggingLevel:     defaultLoggingLevel,
		channelID:        defaultChannelID,
		orgIDsStr:        defaultOrgIDs,
		chaincodeVersion: defaultChaincodeVersion,
		iterations:       1,
		args:             getEmptyArgs(),
	}
}

// CLIConfig overrides certain configuration values with those supplied on the command-line
type CLIConfig struct {
	apiconfig.Config
	logger *logging.Logger
}

// InitConfig initializes the configuration
func InitConfig(flags *pflag.FlagSet) error {
	// var configFile string
	// defaultValue, description := getDefaultValueAndDescription(defaultConfigFile, configFileDescription, defaultConfigFile)
	// flags.StringVar(&configFile, configFileFlag, defaultValue, description)

	instance = &CLIConfig{
		logger: logging.NewLogger(loggerName),
	}

	cnfg, err := config.InitConfig(opts.configFile)
	if err != nil {
		return err
	}

	instance.Config = cnfg

	return nil
}

// Config returns the CLI configuration
func Config() *CLIConfig {
	return instance
}

// Logger returns the Logger for the CLI tool
func (c *CLIConfig) Logger() *logging.Logger {
	return c.logger
}

// LoggingLevel specifies the logging level (DEBUG, INFO, WARNING, ERROR, or CRITICAL)
func (c *CLIConfig) LoggingLevel() string {
	return opts.loggingLevel
}

// InitLoggingLevel initializes the logging level from the provided arguments
func InitLoggingLevel(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(defaultLoggingLevel, loggingLevelDescription, defaultValueAndDescription...)
	flags.StringVar(&opts.loggingLevel, loggingLevelFlag, defaultValue, description)
}

// InitConfigFile initializes the config file path from the provided arguments
func InitConfigFile(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(defaultConfigFile, configFileDescription, defaultValueAndDescription...)
	flags.StringVar(&opts.configFile, configFileFlag, defaultValue, description)
}

// OrgID specifies the ID of the current organization. If multiple org IDs are specified then the first one is returned.
func (c *CLIConfig) OrgID() string {
	return c.OrgIDs()[0]
}

// OrgIDs returns a comma-separated list of organization IDs
func (c *CLIConfig) OrgIDs() []string {
	var orgIDs []string
	s := strings.Split(opts.orgIDsStr, ",")
	for _, orgID := range s {
		orgIDs = append(orgIDs, orgID)
	}
	return orgIDs
}

// InitOrgIDs initializes the org IDs from the provided arguments
func InitOrgIDs(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(defaultOrgIDs, orgIDsDescription, defaultValueAndDescription...)
	flags.StringVar(&opts.orgIDsStr, orgIDsFlag, defaultValue, description)
}

// ChannelID returns the channel ID
func (c *CLIConfig) ChannelID() string {
	return opts.channelID
}

// InitChannelID initializes the channel ID from the provided arguments
func InitChannelID(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(defaultChannelID, channelIDDescription, defaultValueAndDescription...)
	flags.StringVar(&opts.channelID, channelIDFlag, defaultValue, description)
}

// UserName returns the name of the enrolled user
func (c *CLIConfig) UserName() string {
	return opts.user
}

// InitUserName initializes the user name from the provided arguments
func InitUserName(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(defaultUser, userDescription, defaultValueAndDescription...)
	flags.StringVar(&opts.user, userFlag, defaultValue, description)
}

// UserPassword is the password to use when enrolling a user
func (c *CLIConfig) UserPassword() string {
	return opts.password
}

// InitUserPassword initializes the user password from the provided arguments
func InitUserPassword(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(defaultPassword, passwordDescription, defaultValueAndDescription...)
	flags.StringVar(&opts.password, passwordFlag, defaultValue, description)
}

// ChaincodeID returns the chaicode ID
func (c *CLIConfig) ChaincodeID() string {
	return opts.chaincodeID
}

// InitChaincodeID initializes the chaincode ID from the provided arguments
func InitChaincodeID(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(defaultChaincodeID, chaincodeIDDescription, defaultValueAndDescription...)
	flags.StringVar(&opts.chaincodeID, chaincodeIDFlag, defaultValue, description)
}

// ChaincodeEvent the name of the chaincode event to listen for
func (c *CLIConfig) ChaincodeEvent() string {
	return opts.chaincodeEvent
}

// InitChaincodeEvent initializes the chaincode event name from the provided arguments
func InitChaincodeEvent(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(defaultChaincodeEvent, chaincodeEventDescription, defaultValueAndDescription...)
	flags.StringVar(&opts.chaincodeEvent, chaincodeEventFlag, defaultValue, description)
}

// ChaincodePath returns the source path of the chaincode to install/instantiate
func (c *CLIConfig) ChaincodePath() string {
	return opts.chaincodePath
}

// InitChaincodePath initializes the chaincode install source path from the provided arguments
func InitChaincodePath(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(defaultChaincodePath, chaincodePathDescription, defaultValueAndDescription...)
	flags.StringVar(&opts.chaincodePath, chaincodePathFlag, defaultValue, description)
}

// ChaincodeVersion returns the version of the chaincode
func (c *CLIConfig) ChaincodeVersion() string {
	return opts.chaincodeVersion
}

// InitChaincodeVersion initializes the chaincode version from the provided arguments
func InitChaincodeVersion(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(defaultChaincodeVersion, chaincodeVersionDescription, defaultValueAndDescription...)
	flags.StringVar(&opts.chaincodeVersion, chaincodeVersionFlag, defaultValue, description)
}

// PeerURL returns a comma-separated list of peers in the format host1:port1,host2:port2,...
func (c *CLIConfig) PeerURL() string {
	return opts.peerURL
}

// InitPeerURL initializes the peer URL from the provided arguments
func InitPeerURL(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(defaultPeerURL, peerURLDescription, defaultValueAndDescription...)
	flags.StringVar(&opts.peerURL, peerURLFlag, defaultValue, description)
}

// OrdererURL returns the URL of the orderer
func (c *CLIConfig) OrdererURL() string {
	return opts.ordererURL
}

// InitOrdererURL initializes the orderer URL from the provided arguments
func InitOrdererURL(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(defaultOrdererURL, ordererURLDescription, defaultValueAndDescription...)
	flags.StringVar(&opts.ordererURL, ordererFlag, defaultValue, description)
}

// Iterations returns the number of times that a chaincode should be invoked
func (c *CLIConfig) Iterations() int {
	return opts.iterations
}

// InitIterations initializes the number of query/invoke iterations from the provided arguments
func InitIterations(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(defaultIterations, iterationsDescription, defaultValueAndDescription...)
	i, err := strconv.Atoi(defaultValue)
	if err != nil {
		fmt.Printf("Invalid number: %s\n", defaultValue)
		i = 1
	}
	flags.IntVar(&opts.iterations, iterationsFlag, i, description)
}

// SleepTime returns the number of milliseconds to sleep between invocations of a chaincode
func (c *CLIConfig) SleepTime() int64 {
	return opts.sleepTime
}

// InitSleepTime initializes the sleep time from the provided arguments
func InitSleepTime(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(defaultSleepTime, sleepTimeDescription, defaultValueAndDescription...)
	i, err := strconv.Atoi(defaultValue)
	if err != nil {
		fmt.Printf("Invalid number: %s\n", defaultValue)
		i = 1
	}
	flags.Int64Var(&opts.sleepTime, sleepFlag, int64(i), description)
}

// BlockNum returns the block number (where 0 is the first block)
func (c *CLIConfig) BlockNum() int {
	return opts.blockNum
}

// InitBlockNum initializes the bluck number from the provided arguments
func InitBlockNum(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(defaultBlockNum, blockNumDescription, defaultValueAndDescription...)
	i, err := strconv.Atoi(defaultValue)
	if err != nil {
		fmt.Printf("Invalid number: %s\n", defaultValue)
		i = 1
	}
	flags.IntVar(&opts.blockNum, blockNumFlag, i, description)
}

// BlockHash specifies the hash of the block
func (c *CLIConfig) BlockHash() string {
	return opts.blockHash
}

// InitBlockHash initializes the block hash from the provided arguments
func InitBlockHash(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(defaultBlockHash, blockHashDescription, defaultValueAndDescription...)
	flags.StringVar(&opts.blockHash, blockHashFlag, defaultValue, description)
}

// Traverse returns the number of blocks to traverse backwards in the query block command
func (c *CLIConfig) Traverse() int {
	return opts.traverse
}

// InitTraverse initializes the 'traverse' flag from the provided arguments
func InitTraverse(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(defaultTraverse, traverseDescription, defaultValueAndDescription...)
	i, err := strconv.Atoi(defaultValue)
	if err != nil {
		fmt.Printf("Invalid number: %s\n", defaultValue)
		i = 1
	}
	flags.IntVar(&opts.traverse, traverseFlag, i, description)
}

// PrintFormat returns the print (output) format for a block
func (c *CLIConfig) PrintFormat() string {
	return opts.printFormat
}

// InitPrintFormat initializes the print format from the provided arguments
func InitPrintFormat(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription("display", printFormatDescription, defaultValueAndDescription...)
	flags.StringVar(&opts.printFormat, printFormatFlag, defaultValue, description)
}

// Writer returns the writer for output
func (c *CLIConfig) Writer() string {
	return opts.writer
}

// InitWriter initializes the print writer from the provided arguments
func InitWriter(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription("stdout", writerDescription, defaultValueAndDescription...)
	flags.StringVar(&opts.writer, writerFlag, defaultValue, description)
}

// OrdererTLSCertificate is the path of the orderer TLS certificate
func (c *CLIConfig) OrdererTLSCertificate() string {
	return opts.certificate
}

// InitOrdererTLSCertificate initializes the orderer TLS certificate from the provided arguments
func InitOrdererTLSCertificate(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(defaultCertificate, certificateDescription, defaultValueAndDescription...)
	flags.StringVar(&opts.certificate, certificateFileFlag, defaultValue, description)
}

// Args returns the chaincode invocation arguments as a JSON string in the format, {"Func":"function","Args":["arg1","arg2",...]}
func (c *CLIConfig) Args() string {
	return opts.args
}

// InitArgs initializes the invoke/query args from the provided arguments
func InitArgs(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(getEmptyArgs(), argsDescription, defaultValueAndDescription...)
	flags.StringVar(&opts.args, argsFlag, defaultValue, description)
}

// TxFile is the path of the .tx file used to create a channel
func (c *CLIConfig) TxFile() string {
	return opts.txFile
}

// InitTxFile initializes the path of the .tx file used to create/update a channel from the provided arguments
func InitTxFile(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(defaultTxFile, txFileDescription, defaultValueAndDescription...)
	flags.StringVar(&opts.txFile, txFileFlag, defaultValue, description)
}

// TxID returns the transaction ID
func (c *CLIConfig) TxID() string {
	return opts.txID
}

// InitTxID initializes the transaction D from the provided arguments
func InitTxID(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(defaultTxID, txIDDescription, defaultValueAndDescription...)
	flags.StringVar(&opts.txID, txIDFlag, defaultValue, description)
}

// ChaincodePolicy returns the chaincode policy string, e.g Nof(1,(SignedBy(Org1Msp),SignedBy(Org2MSP)))
func (c *CLIConfig) ChaincodePolicy() string {
	return opts.chaincodePolicy
}

// InitChaincodePolicy initializes the chaincode policy from the provided arguments
func InitChaincodePolicy(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(defaultChaincodePolicy, chaincodePolicyDescription, defaultValueAndDescription...)
	flags.StringVar(&opts.chaincodePolicy, chaincodePolicyFlag, defaultValue, description)
}

// Timeout returns the timeout (in milliseconds) for various operations
func (c *CLIConfig) Timeout() time.Duration {
	return time.Duration(opts.timeout) * time.Millisecond
}

// InitTimeout initializes the timeout from the provided arguments
func InitTimeout(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(defaultTimeout, timeoutDescription, defaultValueAndDescription...)
	i, err := strconv.Atoi(defaultValue)
	if err != nil {
		fmt.Printf("Invalid number: %s\n", defaultValue)
		i = 1000
	}
	flags.Int64Var(&opts.timeout, timeoutFlag, int64(i), description)
}

// Overrides of apifabclient.Config...

// OrderersConfig returns the configuration of all of the defined orderers
func (c *CLIConfig) OrderersConfig() ([]apiconfig.OrdererConfig, error) {
	overridden := false

	configs, err := c.Config.OrderersConfig()
	if err != nil {
		return nil, err
	}

	defaultConfig := configs[0]

	url := defaultConfig.URL

	if c.OrdererURL() != "" {
		overridden = true
		url = c.OrdererURL()
	}

	certificate := defaultConfig.TLSCACerts.Path
	pem := defaultConfig.TLSCACerts.Pem

	if opts.certificate != "" {
		overridden = true
		certificate = c.OrdererTLSCertificate()
	}

	if !overridden {
		return c.Config.OrderersConfig()
	}

	return []apiconfig.OrdererConfig{
		apiconfig.OrdererConfig{
			URL: url,
			TLSCACerts: apiconfig.TLSConfig{
				Path: certificate,
				Pem:  pem,
			},
		},
	}, nil
}

// RandomOrdererConfig returns the configuration of a randomly selected orderer
func (c *CLIConfig) RandomOrdererConfig() (*apiconfig.OrdererConfig, error) {
	orderers, err := c.OrderersConfig()
	if err != nil {
		return nil, err
	}
	return &orderers[0], nil
}

// IsLoggingEnabledFor indicates whether the logger is enabled for the given logging level
func (c *CLIConfig) IsLoggingEnabledFor(level logging.Level) bool {
	return logging.IsEnabledFor(level, loggerName)
}

// Utility functions...

func getEmptyArgs() string {
	// argBytes, err := json.Marshal(&common.ArgStruct{})
	// if err != nil {
	// 	panic(errors.Errorf("error marshaling empty args struct: %v", err))
	// }
	// return string(argBytes)
	return "{}"
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
