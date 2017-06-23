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

	"strings"

	"sync"

	"github.com/hyperledger/fabric-sdk-go/api"
	bccspFactory "github.com/hyperledger/fabric/bccsp/factory"
	logging "github.com/op/go-logging"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
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
	peerURLDescription = "The URL of the peer to connect to, e.g. localhost:7051"
	defaultPeerURL     = ""

	ordererFlag           = "orderer"
	ordererURLDescription = "The URL of the orderer, e.g. localhost:7050"
	defaultOrdererURL     = ""

	printFormatFlag        = "format"
	printFormatDescription = "The output format - display, json, raw"

	certificateFileFlag    = "cacert"
	certificateDescription = "The path of the ca-cert.pem file"
	defaultCertificate     = ""

	argsFlag        = "args"
	argsDescription = "The args in JSON format. Example: {\"Args\":[\"arg1\",\"arg2\"]}"

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

	txIDFlag        = "tx"
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
)

var configInstance *cliConfig
var configInit sync.Once

// CLIConfig extendsthe fabric API config and provides additional configuration options
type CLIConfig interface {
	api.Config

	Logger() *logging.Logger

	LoggingLevel() string
	InitLoggingLevel(flags *pflag.FlagSet, defaultValueAndDescription ...string)

	ConfigFile() string
	InitConfigFile(flags *pflag.FlagSet, defaultValueAndDescription ...string)

	ChannelID() string
	InitChannelID(flags *pflag.FlagSet, defaultValueAndDescription ...string)

	User() string
	InitUser(flags *pflag.FlagSet, defaultValueAndDescription ...string)

	Password() string
	InitPassword(flags *pflag.FlagSet, defaultValueAndDescription ...string)

	ChaincodeID() string
	InitChaincodeID(flags *pflag.FlagSet, defaultValueAndDescription ...string)

	ChaincodePath() string
	InitChaincodePath(flags *pflag.FlagSet, defaultValueAndDescription ...string)

	ChaincodeVersion() string
	InitChaincodeVersion(flags *pflag.FlagSet, defaultValueAndDescription ...string)

	ChaincodeEvent() string
	InitChaincodeEvent(flags *pflag.FlagSet, defaultValueAndDescription ...string)

	PeerURL() string
	InitPeerURL(flags *pflag.FlagSet, defaultValueAndDescription ...string)

	OrdererURL() string
	InitOrdererURL(flags *pflag.FlagSet, defaultValueAndDescription ...string)

	Certificate() string
	InitCertificate(flags *pflag.FlagSet, defaultValueAndDescription ...string)

	Iterations() int
	InitIterations(flags *pflag.FlagSet, defaultValueAndDescription ...string)

	BlockNum() int
	InitBlockNum(flags *pflag.FlagSet, defaultValueAndDescription ...string)

	BlockHash() string
	InitBlockHash(flags *pflag.FlagSet, defaultValueAndDescription ...string)

	Traverse() int
	InitTraverse(flags *pflag.FlagSet, defaultValueAndDescription ...string)

	SleepTime() int64
	InitSleepTime(flags *pflag.FlagSet, defaultValueAndDescription ...string)

	PrintFormat() string
	InitPrintFormat(flags *pflag.FlagSet, defaultValueAndDescription ...string)

	Args() string
	InitArgs(flags *pflag.FlagSet, defaultValueAndDescription ...string)

	TxFile() string
	InitTxFile(flags *pflag.FlagSet, defaultValueAndDescription ...string)

	TxID() string
	InitTxID(flags *pflag.FlagSet, defaultValueAndDescription ...string)
}

// cliConfig overrides certain configuration values with those supplied on the command-line
type cliConfig struct {
	config           api.Config
	logger           *logging.Logger
	certificate      string
	user             string
	password         string
	loggingLevel     string
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
	args             string
	chaincodeEvent   string
	blockHash        string
	blockNum         int
	traverse         int
}

func getConfigImpl() *cliConfig {
	configInit.Do(func() {
		configInstance = &cliConfig{
			logger:           logging.MustGetLogger(loggerName),
			user:             defaultUser,
			password:         defaultPassword,
			loggingLevel:     defaultLoggingLevel,
			channelID:        defaultChannelID,
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

func (c *cliConfig) ChannelID() string {
	return c.channelID
}

func (c *cliConfig) InitChannelID(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(defaultChannelID, channelIDDescription, defaultValueAndDescription...)
	flags.StringVar(&c.channelID, channelIDFlag, defaultValue, description)
}

func (c *cliConfig) User() string {
	return c.user
}

func (c *cliConfig) InitUser(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
	defaultValue, description := getDefaultValueAndDescription(defaultUser, userDescription, defaultValueAndDescription...)
	flags.StringVar(&c.user, userFlag, defaultValue, description)
}

func (c *cliConfig) Password() string {
	return c.password
}

func (c *cliConfig) InitPassword(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
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

func (c *cliConfig) Certificate() string {
	return c.certificate
}

func (c *cliConfig) InitCertificate(flags *pflag.FlagSet, defaultValueAndDescription ...string) {
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

// Implementation of api.Config...

func (c *cliConfig) GetServerURL() string {
	return c.config.GetServerURL()
}

func (c *cliConfig) GetServerCertFiles() []string {
	return c.config.GetServerCertFiles()
}

func (c *cliConfig) GetFabricCAClientKeyFile() string {
	return c.config.GetFabricCAClientKeyFile()
}

func (c *cliConfig) GetFabricCAClientCertFile() string {
	return c.config.GetFabricCAClientCertFile()
}

func (c *cliConfig) GetFabricCATLSEnabledFlag() bool {
	return c.config.GetFabricCATLSEnabledFlag()
}

func (c *cliConfig) GetFabricClientViper() *viper.Viper {
	return c.config.GetFabricClientViper()
}

func (c *cliConfig) GetPeersConfig() ([]api.PeerConfig, error) {
	return c.config.GetPeersConfig()
}

func (c *cliConfig) IsTLSEnabled() bool {
	return c.config.IsTLSEnabled()
}

func (c *cliConfig) GetTLSCACertPool(tlsCertificate string) (*x509.CertPool, error) {
	return c.config.GetTLSCACertPool(tlsCertificate)
}

func (c *cliConfig) GetTLSCACertPoolFromRoots(ordererRootCAs [][]byte) (*x509.CertPool, error) {
	return c.config.GetTLSCACertPoolFromRoots(ordererRootCAs)
}

func (c *cliConfig) IsSecurityEnabled() bool {
	return c.config.IsSecurityEnabled()
}

func (c *cliConfig) TcertBatchSize() int {
	return c.config.TcertBatchSize()
}

func (c *cliConfig) GetSecurityAlgorithm() string {
	return c.config.GetSecurityAlgorithm()
}

func (c *cliConfig) GetSecurityLevel() int {
	return c.config.GetSecurityLevel()
}

func (c *cliConfig) GetOrdererHost() string {
	if c.ordererURL == "" {
		return c.config.GetOrdererHost()
	}
	return strings.Split(c.ordererURL, ":")[0]
}

func (c *cliConfig) GetOrdererPort() string {
	if c.ordererURL == "" {
		return c.config.GetOrdererPort()
	}
	s := strings.Split(c.ordererURL, ":")
	if len(s) > 1 {
		return s[1]
	}
	return c.config.GetOrdererPort()
}

func (c *cliConfig) GetOrdererTLSServerHostOverride() string {
	return c.config.GetOrdererTLSServerHostOverride()
}

func (c *cliConfig) GetOrdererTLSCertificate() string {
	if c.certificate == "" {
		return c.config.GetOrdererTLSCertificate()
	}
	return c.certificate
}

func (c *cliConfig) GetFabricCAID() string {
	return c.config.GetFabricCAID()
}

func (c *cliConfig) GetFabricCAName() string {
	return c.config.GetFabricCAName()
}

func (c *cliConfig) GetKeyStorePath() string {
	return c.config.GetKeyStorePath()
}

func (c *cliConfig) GetFabricCAHomeDir() string {
	return c.config.GetFabricCAHomeDir()
}

func (c *cliConfig) GetFabricCAMspDir() string {
	return c.config.GetFabricCAMspDir()
}

func (c *cliConfig) GetCryptoConfigPath() string {
	return c.config.GetCryptoConfigPath()
}

func (c *cliConfig) GetCSPConfig() *bccspFactory.FactoryOpts {
	return c.config.GetCSPConfig()
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
