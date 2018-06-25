/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package config

import (
	"fmt"
	"strconv"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/deliverclient/seek"

	"strings"

	"github.com/spf13/pflag"

	"os"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
)

const (
	loggerName    = "fabriccli"
	userStatePath = "/tmp/enroll_user"

	// AutoDetectSelectionProvider indicates that a selection provider is to be automatically determined using channel capabilities
	AutoDetectSelectionProvider = "auto"

	// StaticSelectionProvider indicates that a static selection provider is to be used for selecting peers for invoke/query commands
	StaticSelectionProvider = "static"

	// DynamicSelectionProvider indicates that a dynamic selection provider is to be used for selecting peers for invoke/query commands
	DynamicSelectionProvider = "dynamic"

	// FabricSelectionProvider indicates that the Fabric selection provider is to be used for selecting peers for invoke/query commands
	FabricSelectionProvider = "fabric"
)

// Flags
const (
	UserFlag        = "user"
	userDescription = "The user"
	defaultUser     = ""

	PasswordFlag        = "pw"
	passwordDescription = "The password of the user"
	defaultPassword     = ""

	ChaincodeVersionFlag        = "v"
	chaincodeVersionDescription = "The chaincode version"
	defaultChaincodeVersion     = "v0"

	LoggingLevelFlag        = "logging-level"
	loggingLevelDescription = "Logging level - ERROR, WARN, INFO, DEBUG"
	defaultLoggingLevel     = "ERROR"

	OrgIDsFlag        = "orgid"
	orgIDsDescription = "A comma-separated list of organization IDs"
	defaultOrgIDs     = ""

	ChannelIDFlag        = "cid"
	channelIDDescription = "The channel ID"
	defaultChannelID     = ""

	ChaincodeIDFlag        = "ccid"
	chaincodeIDDescription = "The Chaincode ID"
	defaultChaincodeID     = ""

	ChaincodePathFlag        = "ccp"
	chaincodePathDescription = "The chaincode path"
	defaultChaincodePath     = ""

	ConfigFileFlag        = "config"
	configFileDescription = "The path of the config.yaml file"
	defaultConfigFile     = ""

	PeerURLFlag        = "peer"
	peerURLDescription = "A comma-separated list of peer targets, e.g. 'grpcs://localhost:7051,grpcs://localhost:8051'"
	defaultPeerURL     = ""

	OrdererFlag           = "orderer"
	ordererURLDescription = "The URL of the orderer, e.g. grpcs://localhost:7050"
	defaultOrdererURL     = ""

	PrintFormatFlag        = "format"
	printFormatDescription = "The output format - display, json, raw"

	WriterFlag        = "writer"
	writerDescription = "The writer - stdout, stderr, log"

	Base64Flag        = "base64"
	base64Description = "If true then binary values are encoded in base64 (only applies to 'display' format)"

	CertificateFileFlag    = "cacert"
	certificateDescription = "The path of the ca-cert.pem file"
	defaultCertificate     = ""

	ArgsFlag        = "args"
	argsDescription = `The args in JSON format. Example: {"Func":"function","Args":["arg1","arg2"]}. Note that $rand(N) may be used anywhere within the value of the arg in order to generate a random value between 0 and N. For example {"Func":"function","Args":["arg_$rand(100)","$rand(10)"]}.`

	IterationsFlag        = "iterations"
	iterationsDescription = "The number of times to invoke the chaincode"
	defaultIterations     = "1"

	SleepFlag            = "sleep"
	sleepTimeDescription = "The number of milliseconds to sleep between invocations of the chaincode."
	defaultSleepTime     = "0"

	TxFileFlag        = "txfile"
	txFileDescription = "The path of the channel.tx file"
	defaultTxFile     = ""

	ChaincodeEventFlag        = "event"
	chaincodeEventDescription = "The name of the chaincode event to listen for"
	defaultChaincodeEvent     = ""

	SeekTypeFlag        = "seek"
	seekTypeDescription = "The seek type. Possible values: oldest - delivers all blocks from the oldest block; newest - delivers the newest block; from - delivers from the block number as specified by the '--num' flag"
	defaultSeekType     = string(seek.Newest)

	TxIDFlag        = "txid"
	txIDDescription = "The transaction ID"
	defaultTxID     = ""

	BlockNumFlag        = "num"
	blockNumDescription = "The block number"
	defaultBlockNum     = "0"

	BlockHashFlag        = "hash"
	blockHashDescription = "The block hash"
	defaultBlockHash     = ""

	TraverseFlag        = "traverse"
	traverseDescription = "Blocks will be traversed starting with the given block in reverse order up to the given number of blocks"
	defaultTraverse     = "0"

	ChaincodePolicyFlag        = "policy"
	chaincodePolicyDescription = "The chaincode policy, e.g. OutOf(1,'Org1MSP.admin','Org2MSP.admin',AND('Org3MSP.member','Org4MSP.member'))"
	defaultChaincodePolicy     = ""

	CollectionConfigFileFlag        = "collconfig"
	collectionConfigFileDescription = "The path of the JSON file that contains the private data collection configuration for the chaincode"
	defaultCollectionConfigFile     = ""

	TimeoutFlag        = "timeout"
	timeoutDescription = "The timeout (in milliseconds) for the operation"
	defaultTimeout     = "5000"

	PrintPayloadOnlyFlag        = "payload"
	printPayloadOnlyDescription = "If specified then only the payload from the transaction proposal response(s) will be output"
	defaultPrintPayloadOnly     = "false"

	ConcurrencyFlag        = "concurrency"
	concurrencyDescription = "Specifies the number of concurrent requests sent on an invoke or a query chaincode request"
	defaultConcurrency     = "1"

	MaxAttemptsFlag        = "attempts"
	maxAttemptsDescription = "Specifies the maximum number of attempts to be made for a single chaincode invocation request. If >1 then retries will be attempted should transient errors occur."
	defaultMaxAttempts     = "3"

	InitialBackoffFlag        = "backoff"
	initialBackoffDescription = "The initial backoff is the time (in milliseconds) to wait before resubmitting an invocation after a transient error"
	defaultInitialBackoff     = "1000"

	MaxBackoffFlag        = "maxbackoff"
	maxBackoffDescription = "The maximum backoff time (in milliseconds)"
	defaultMaxBackoff     = "5000"

	BackoffFactorFlag        = "backofffactor"
	backoffFactorDescription = "The factor by which the backoff time is multiplied each time a retry fails. For example, if the initial backoff is 1s and factor is 2 then the next retry will have a backoff of 2s and a subsequent backoff will be 4s up to the maximum backoff"
	defaultBackoffFactor     = "2"

	VerboseFlag        = "verbose"
	verboseDescription = "If specified then the transaction proposal responses will be output when iterations > 1, otherwise transaction proposal responses are only output when iterations = 1"
	defaultVerbosity   = "false"

	SelectionProviderFlag        = "selectprovider"
	selectionProviderDescription = "The peer selection provider for invoke/query commands. The possible values are: (1) static - Selects all peers; (2) dynamic - Uses the built-in selection service from the SDK to select a minimal set of peers according to the endorsement policy of the chaincode; (3) fabric - Uses Fabric's Discovery Service to select a minimal set of peers according to the endorsement/collection policy of the chaincode; (4) auto (default) - Automatically determines which selection service to use based on channel capabilities."
	defaultSelectionProvider     = AutoDetectSelectionProvider

	GoPathFlag        = "gopath"
	goPathDescription = "GOPATH for chaincode install command. If not set, GOPATH is taken from the environment"
	defaultGoPath     = ""
)

var opts *options
var instance *CLIConfig

type options struct {
	certificate          string
	user                 string
	password             string
	loggingLevel         string
	orgIDsStr            string
	channelID            string
	chaincodeID          string
	chaincodePath        string
	chaincodeVersion     string
	peerURL              string
	ordererURL           string
	iterations           int
	sleepTime            int64
	configFile           string
	txFile               string
	txID                 string
	printFormat          string
	writer               string
	base64               bool
	args                 string
	chaincodeEvent       string
	seekType             string
	blockHash            string
	blockNum             uint64
	traverse             int
	chaincodePolicy      string
	collectionConfigFile string
	timeout              int64
	printPayloadOnly     bool
	concurrency          int
	maxAttempts          int
	initialBackoff       int64
	maxBackoff           int64
	backoffFactor        float64
	verbose              bool
	selectionProvider    string
	goPath               string
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
		concurrency:      1,
		args:             getEmptyArgs(),
	}
}

// CLIConfig overrides certain configuration values with those supplied on the command-line
type CLIConfig struct {
	core.ConfigProvider
	logger   *logging.Logger
	setFlags map[string]string
}

// InitConfig initializes the configuration
func InitConfig(flags *pflag.FlagSet) error {

	instance = &CLIConfig{
		logger:   logging.NewLogger(loggerName),
		setFlags: make(map[string]string),
	}
	flags.Visit(func(flag *pflag.Flag) {
		instance.setFlags[flag.Name] = flag.Value.String()
	})

	cnfg := config.FromFile(opts.configFile)

	instance.ConfigProvider = cnfg

	return nil
}

// IsFlagSet indicates whether or not the given flag is set
func IsFlagSet(name string) bool {
	_, ok := instance.setFlags[name]
	return ok
}

// Provider returns the config provider
func Provider() core.ConfigProvider {
	return instance.ConfigProvider
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
	flags.StringVar(&opts.loggingLevel, LoggingLevelFlag, defaultValue, description)
}

// InitConfigFile initializes the config file path from the provided arguments
func InitConfigFile(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(defaultConfigFile, configFileDescription, defaultValueAndDescription...)
	flags.StringVar(&opts.configFile, ConfigFileFlag, defaultValue, description)
}

// OrgID specifies the ID of the current organization. If multiple org IDs are specified then the first one is returned.
func (c *CLIConfig) OrgID() string {
	if len(c.OrgIDs()) == 0 {
		return ""
	}
	return c.OrgIDs()[0]
}

// OrgIDs returns a comma-separated list of organization IDs
func (c *CLIConfig) OrgIDs() []string {
	var orgIDs []string
	if len(strings.TrimSpace(opts.orgIDsStr)) > 0 {
		s := strings.Split(opts.orgIDsStr, ",")
		for _, orgID := range s {
			orgIDs = append(orgIDs, orgID)
		}
	}
	return orgIDs
}

// InitOrgIDs initializes the org IDs from the provided arguments
func InitOrgIDs(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(defaultOrgIDs, orgIDsDescription, defaultValueAndDescription...)
	flags.StringVar(&opts.orgIDsStr, OrgIDsFlag, defaultValue, description)
}

// ChannelID returns the channel ID
func (c *CLIConfig) ChannelID() string {
	return opts.channelID
}

// InitChannelID initializes the channel ID from the provided arguments
func InitChannelID(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(defaultChannelID, channelIDDescription, defaultValueAndDescription...)
	flags.StringVar(&opts.channelID, ChannelIDFlag, defaultValue, description)
}

// UserName returns the name of the enrolled user
func (c *CLIConfig) UserName() string {
	return opts.user
}

// InitUserName initializes the user name from the provided arguments
func InitUserName(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(defaultUser, userDescription, defaultValueAndDescription...)
	flags.StringVar(&opts.user, UserFlag, defaultValue, description)
}

// UserPassword is the password to use when enrolling a user
func (c *CLIConfig) UserPassword() string {
	return opts.password
}

// InitUserPassword initializes the user password from the provided arguments
func InitUserPassword(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(defaultPassword, passwordDescription, defaultValueAndDescription...)
	flags.StringVar(&opts.password, PasswordFlag, defaultValue, description)
}

// ChaincodeID returns the chaicode ID
func (c *CLIConfig) ChaincodeID() string {
	return opts.chaincodeID
}

// InitChaincodeID initializes the chaincode ID from the provided arguments
func InitChaincodeID(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(defaultChaincodeID, chaincodeIDDescription, defaultValueAndDescription...)
	flags.StringVar(&opts.chaincodeID, ChaincodeIDFlag, defaultValue, description)
}

// ChaincodeEvent the name of the chaincode event to listen for
func (c *CLIConfig) ChaincodeEvent() string {
	return opts.chaincodeEvent
}

// InitChaincodeEvent initializes the chaincode event name from the provided arguments
func InitChaincodeEvent(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(defaultChaincodeEvent, chaincodeEventDescription, defaultValueAndDescription...)
	flags.StringVar(&opts.chaincodeEvent, ChaincodeEventFlag, defaultValue, description)
}

// SeekType the seek type for Deliver Events. Possible values:
// - Oldest - will deliver all blocks from the oldest block and will continue listening for new blocks
// - Newest - will deliver the newest block and will continue listening for new blocks
// - FromBlock - Delivers from the specific block, as specified by the "--num" flag
func (c *CLIConfig) SeekType() seek.Type {
	return seek.Type(opts.seekType)
}

// InitSeekType initializes the seek type from the provided arguments
func InitSeekType(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(defaultSeekType, seekTypeDescription, defaultValueAndDescription...)
	flags.StringVar(&opts.seekType, SeekTypeFlag, defaultValue, description)
}

// ChaincodePath returns the source path of the chaincode to install/instantiate
func (c *CLIConfig) ChaincodePath() string {
	return opts.chaincodePath
}

// InitChaincodePath initializes the chaincode install source path from the provided arguments
func InitChaincodePath(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(defaultChaincodePath, chaincodePathDescription, defaultValueAndDescription...)
	flags.StringVar(&opts.chaincodePath, ChaincodePathFlag, defaultValue, description)
}

// ChaincodeVersion returns the version of the chaincode
func (c *CLIConfig) ChaincodeVersion() string {
	return opts.chaincodeVersion
}

// InitChaincodeVersion initializes the chaincode version from the provided arguments
func InitChaincodeVersion(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(defaultChaincodeVersion, chaincodeVersionDescription, defaultValueAndDescription...)
	flags.StringVar(&opts.chaincodeVersion, ChaincodeVersionFlag, defaultValue, description)
}

// PeerURL returns a comma-separated list of peers in the format host1:port1,host2:port2,...
func (c *CLIConfig) PeerURL() string {
	return opts.peerURL
}

// PeerURLs returns a list of peer URLs
func (c *CLIConfig) PeerURLs() []string {
	var urls []string
	if len(strings.TrimSpace(opts.peerURL)) > 0 {
		s := strings.Split(opts.peerURL, ",")
		for _, orgID := range s {
			urls = append(urls, orgID)
		}
	}
	return urls
}

// InitPeerURL initializes the peer URL from the provided arguments
func InitPeerURL(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(defaultPeerURL, peerURLDescription, defaultValueAndDescription...)
	flags.StringVar(&opts.peerURL, PeerURLFlag, defaultValue, description)
}

// OrdererURL returns the URL of the orderer
func (c *CLIConfig) OrdererURL() string {
	return opts.ordererURL
}

// InitOrdererURL initializes the orderer URL from the provided arguments
func InitOrdererURL(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(defaultOrdererURL, ordererURLDescription, defaultValueAndDescription...)
	flags.StringVar(&opts.ordererURL, OrdererFlag, defaultValue, description)
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
		fmt.Printf("Invalid number for %s: %s\n", IterationsFlag, defaultValue)
		os.Exit(-1)
	}
	flags.IntVar(&opts.iterations, IterationsFlag, i, description)
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
		fmt.Printf("Invalid number for %s: %s\n", SleepFlag, defaultValue)
		os.Exit(-1)
	}
	flags.Int64Var(&opts.sleepTime, SleepFlag, int64(i), description)
}

// BlockNum returns the block number (where 0 is the first block)
func (c *CLIConfig) BlockNum() uint64 {
	return opts.blockNum
}

// InitBlockNum initializes the bluck number from the provided arguments
func InitBlockNum(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(defaultBlockNum, blockNumDescription, defaultValueAndDescription...)
	i, err := strconv.ParseUint(defaultValue, 10, 64)
	if err != nil {
		fmt.Printf("Invalid number for %s: %s\n", BlockNumFlag, defaultValue)
		os.Exit(-1)
	}
	flags.Uint64Var(&opts.blockNum, BlockNumFlag, i, description)
}

// BlockHash specifies the hash of the block
func (c *CLIConfig) BlockHash() string {
	return opts.blockHash
}

// InitBlockHash initializes the block hash from the provided arguments
func InitBlockHash(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(defaultBlockHash, blockHashDescription, defaultValueAndDescription...)
	flags.StringVar(&opts.blockHash, BlockHashFlag, defaultValue, description)
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
		fmt.Printf("Invalid number for %s: %s\n", TimeoutFlag, defaultValue)
		i = 1
	}
	flags.IntVar(&opts.traverse, TraverseFlag, i, description)
}

// PrintFormat returns the print (output) format for a block
func (c *CLIConfig) PrintFormat() string {
	return opts.printFormat
}

// InitPrintFormat initializes the print format from the provided arguments
func InitPrintFormat(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription("display", printFormatDescription, defaultValueAndDescription...)
	flags.StringVar(&opts.printFormat, PrintFormatFlag, defaultValue, description)
}

// Writer returns the writer for output
func (c *CLIConfig) Writer() string {
	return opts.writer
}

// InitWriter initializes the print writer from the provided arguments
func InitWriter(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription("stdout", writerDescription, defaultValueAndDescription...)
	flags.StringVar(&opts.writer, WriterFlag, defaultValue, description)
}

// Base64 indicates whether binary values are to be encoded in base64. (Only applies to 'display' format.)
func (c *CLIConfig) Base64() bool {
	return opts.base64
}

// InitBase64 initializes the base64 flag from the provided arguments
func InitBase64(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription("false", writerDescription, defaultValueAndDescription...)
	flags.BoolVar(&opts.base64, Base64Flag, defaultValue == "true", description)
}

// OrdererTLSCertificate is the path of the orderer TLS certificate
func (c *CLIConfig) OrdererTLSCertificate() string {
	return opts.certificate
}

// InitOrdererTLSCertificate initializes the orderer TLS certificate from the provided arguments
func InitOrdererTLSCertificate(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(defaultCertificate, certificateDescription, defaultValueAndDescription...)
	flags.StringVar(&opts.certificate, CertificateFileFlag, defaultValue, description)
}

// Args returns the chaincode invocation arguments as a JSON string in the format, {"Func":"function","Args":["arg1","arg2",...]}
func (c *CLIConfig) Args() string {
	return opts.args
}

// InitArgs initializes the invoke/query args from the provided arguments
func InitArgs(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(getEmptyArgs(), argsDescription, defaultValueAndDescription...)
	flags.StringVar(&opts.args, ArgsFlag, defaultValue, description)
}

// TxFile is the path of the .tx file used to create a channel
func (c *CLIConfig) TxFile() string {
	return opts.txFile
}

// InitTxFile initializes the path of the .tx file used to create/update a channel from the provided arguments
func InitTxFile(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(defaultTxFile, txFileDescription, defaultValueAndDescription...)
	flags.StringVar(&opts.txFile, TxFileFlag, defaultValue, description)
}

// TxID returns the transaction ID
func (c *CLIConfig) TxID() string {
	return opts.txID
}

// InitTxID initializes the transaction D from the provided arguments
func InitTxID(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(defaultTxID, txIDDescription, defaultValueAndDescription...)
	flags.StringVar(&opts.txID, TxIDFlag, defaultValue, description)
}

// ChaincodePolicy returns the chaincode policy string, e.g Nof(1,(SignedBy(Org1Msp),SignedBy(Org2MSP)))
func (c *CLIConfig) ChaincodePolicy() string {
	return opts.chaincodePolicy
}

// InitChaincodePolicy initializes the chaincode policy from the provided arguments
func InitChaincodePolicy(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(defaultChaincodePolicy, chaincodePolicyDescription, defaultValueAndDescription...)
	flags.StringVar(&opts.chaincodePolicy, ChaincodePolicyFlag, defaultValue, description)
}

// CollectionConfigFile returns the path of the JSON file that contains the private data collection configuration for the chaincode to be instantiated/upgraded
func (c *CLIConfig) CollectionConfigFile() string {
	return opts.collectionConfigFile
}

// InitCollectionConfigFile initializes the collection config file from the provided arguments
func InitCollectionConfigFile(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(defaultCollectionConfigFile, collectionConfigFileDescription, defaultValueAndDescription...)
	flags.StringVar(&opts.collectionConfigFile, CollectionConfigFileFlag, defaultValue, description)
}

// Timeout returns the timeout (in milliseconds) for various operations
func (c *CLIConfig) Timeout(timeoutType fab.TimeoutType) time.Duration {
	// TODO use provided timoutType
	return time.Duration(opts.timeout) * time.Millisecond
}

// InitTimeout initializes the timeout from the provided arguments
func InitTimeout(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(defaultTimeout, timeoutDescription, defaultValueAndDescription...)
	i, err := strconv.Atoi(defaultValue)
	if err != nil {
		fmt.Printf("Invalid number for %s: %s\n", TimeoutFlag, defaultValue)
		i = 1000
	}
	flags.Int64Var(&opts.timeout, TimeoutFlag, int64(i), description)
}

// PrintPayloadOnly indicates whether only the payload or the entire
// transaction proposal response should be printed
func (c *CLIConfig) PrintPayloadOnly() bool {
	return opts.printPayloadOnly
}

// InitPrintPayloadOnly initializes the PrintPayloadOnly flag from the provided arguments
func InitPrintPayloadOnly(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(defaultPrintPayloadOnly, printPayloadOnlyDescription, defaultValueAndDescription...)
	flags.BoolVar(&opts.printPayloadOnly, PrintPayloadOnlyFlag, defaultValue == "true", description)
}

// Concurrency returns the number of concurrent invocations/queries
func (c *CLIConfig) Concurrency() uint16 {
	return uint16(opts.concurrency)
}

// InitConcurrency initializes the 'concurrency' flag from the provided arguments
func InitConcurrency(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(defaultConcurrency, traverseDescription, defaultValueAndDescription...)
	i, err := strconv.Atoi(defaultValue)
	if err != nil {
		fmt.Printf("Invalid number for %s: %s\n", TimeoutFlag, defaultValue)
		i = 1
	}
	flags.IntVar(&opts.concurrency, ConcurrencyFlag, i, description)
}

// MaxAttempts returns the maximum number of invocations attempts to be made
// for a single chaincode invocation request. If >1 then a retry will be attempted
// if a transient failure occurs.
func (c *CLIConfig) MaxAttempts() int {
	return opts.maxAttempts
}

// InitMaxAttempts initializes the 'maxAttempts' flag from the provided arguments
func InitMaxAttempts(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(defaultMaxAttempts, maxAttemptsDescription, defaultValueAndDescription...)
	i, err := strconv.Atoi(defaultValue)
	if err != nil {
		fmt.Printf("Invalid number for %s: %s\n", TimeoutFlag, defaultValue)
		i = 1
	}
	flags.IntVar(&opts.maxAttempts, MaxAttemptsFlag, i, description)
}

// InitialBackoff returns the time (in milliseconds) to wait
// before resubmitting an invocation after a transient error
func (c *CLIConfig) InitialBackoff() time.Duration {
	return time.Duration(opts.initialBackoff) * time.Millisecond
}

// InitInitialBackoff initializes the initial backoff from the provided arguments
func InitInitialBackoff(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(defaultInitialBackoff, initialBackoffDescription, defaultValueAndDescription...)
	i, err := strconv.Atoi(defaultValue)
	if err != nil {
		fmt.Printf("Invalid number for %s: %s\n", TimeoutFlag, defaultValue)
		i = 1000
	}
	flags.Int64Var(&opts.initialBackoff, InitialBackoffFlag, int64(i), description)
}

// MaxBackoff returns the number
func (c *CLIConfig) MaxBackoff() time.Duration {
	return time.Duration(opts.maxBackoff) * time.Millisecond
}

// InitMaxBackoff initializes the maximum backoff from the provided arguments
func InitMaxBackoff(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(defaultMaxBackoff, maxBackoffDescription, defaultValueAndDescription...)
	i, err := strconv.Atoi(defaultValue)
	if err != nil {
		fmt.Printf("Invalid number for %s: %s\n", TimeoutFlag, defaultValue)
		i = 1000
	}
	flags.Int64Var(&opts.maxBackoff, MaxBackoffFlag, int64(i), description)
}

// BackoffFactor returns the factor by which the backoff time is multiplied each time a retry fails. For example, if the initial
// backoff is 1s and factor is 2 then the next retry will have a backoff of 2s and a subsequent backoff will be 4s up to the maximum backoff
func (c *CLIConfig) BackoffFactor() float64 {
	return opts.backoffFactor
}

// InitBackoffFactor initializes the backoff factor from the provided arguments
func InitBackoffFactor(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(defaultBackoffFactor, backoffFactorDescription, defaultValueAndDescription...)
	i, err := strconv.Atoi(defaultValue)
	if err != nil {
		fmt.Printf("Invalid number for %s: %s\n", TimeoutFlag, defaultValue)
		i = 1000
	}
	flags.Float64Var(&opts.backoffFactor, BackoffFactorFlag, float64(i), description)
}

// Verbose indicates whether or not to print the transaction proposal responses
// when Iterations > 1
func (c *CLIConfig) Verbose() bool {
	return opts.verbose
}

// InitVerbosity initializes the Verbose flag from the provided arguments
func InitVerbosity(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(defaultVerbosity, verboseDescription, defaultValueAndDescription...)
	flags.BoolVar(&opts.verbose, VerboseFlag, defaultValue == "true", description)
}

// SelectionProvider returns the peer selection provider - either static or dynamic
func (c *CLIConfig) SelectionProvider() string {
	return opts.selectionProvider
}

// InitSelectionProvider initializes the peer selection provider from the provided arguments
func InitSelectionProvider(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(defaultSelectionProvider, selectionProviderDescription, defaultValueAndDescription...)
	flags.StringVar(&opts.selectionProvider, SelectionProviderFlag, defaultValue, description)
}

// InitGoPath initializes the gopath from the provided arguments
func InitGoPath(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(defaultGoPath, goPathDescription, defaultValueAndDescription...)
	flags.StringVar(&opts.goPath, GoPathFlag, defaultValue, description)
}

// GoPath returns the gopath
func (c *CLIConfig) GoPath() string {
	gopath := opts.goPath
	if !IsFlagSet(GoPathFlag) {
		gopath = os.Getenv("GOPATH")
	}
	return gopath
}

// IsLoggingEnabledFor indicates whether the logger is enabled for the given logging level
func (c *CLIConfig) IsLoggingEnabledFor(level logging.Level) bool {
	return logging.IsEnabledFor(loggerName, level)
}

// Utility functions...

func getEmptyArgs() string {
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
