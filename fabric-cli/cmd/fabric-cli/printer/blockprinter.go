/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package printer

import (
	"encoding/base64"
	"fmt"
	"reflect"
	"strings"

	"net/http"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/core/common/ccprovider"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/rwsetutil"
	ledgerUtil "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/core/ledger/util"
	fabriccmn "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/ledger/rwset/kvrwset"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/msp"
	ab "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/orderer"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
	utils "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/utils"
	"github.com/pkg/errors"
)

const (
	// AnchorPeersKey is the key name for the AnchorPeers ConfigValue
	AnchorPeersKey = "AnchorPeers"

	// ReadersPolicyKey is the key used for the read policy
	ReadersPolicyKey = "Readers"

	// WritersPolicyKey is the key used for the read policy
	WritersPolicyKey = "Writers"

	// AdminsPolicyKey is the key used for the read policy
	AdminsPolicyKey = "Admins"

	// MSPKey is the org key used for MSP configuration
	MSPKey = "MSP"

	// ConsensusTypeKey is the cb.ConfigItem type key name for the ConsensusType message
	ConsensusTypeKey = "ConsensusType"

	// BatchSizeKey is the cb.ConfigItem type key name for the BatchSize message
	BatchSizeKey = "BatchSize"

	// BatchTimeoutKey is the cb.ConfigItem type key name for the BatchTimeout message
	BatchTimeoutKey = "BatchTimeout"

	// ChannelRestrictionsKey is the key name for the ChannelRestrictions message
	ChannelRestrictionsKey = "ChannelRestrictions"

	// KafkaBrokersKey is the cb.ConfigItem type key name for the KafkaBrokers message
	KafkaBrokersKey = "KafkaBrokers"

	// ChannelCreationPolicyKey is the key for the ChannelCreationPolicy value
	ChannelCreationPolicyKey = "ChannelCreationPolicy"

	// ConsortiumKey is the key for the cb.ConfigValue for the Consortium message
	ConsortiumKey = "Consortium"

	// HashingAlgorithmKey is the cb.ConfigItem type key name for the HashingAlgorithm message
	HashingAlgorithmKey = "HashingAlgorithm"

	// BlockDataHashingStructureKey is the cb.ConfigItem type key name for the BlockDataHashingStructure message
	BlockDataHashingStructureKey = "BlockDataHashingStructure"

	// OrdererAddressesKey is the cb.ConfigItem type key name for the OrdererAddresses message
	OrdererAddressesKey = "OrdererAddresses"

	// ChannelGroupKey is the name of the channel group
	ChannelGroupKey = "Channel"

	// CapabilitiesKey is the key for capabilities
	CapabilitiesKey = "Capabilities"
)

// Printer is used for printing various data structures
type Printer interface {
	// PrintBlockchainInfo outputs BlockchainInfo
	PrintBlockchainInfo(info *fabriccmn.BlockchainInfo)

	// PrintBlock outputs a Block
	PrintBlock(block *fabriccmn.Block)

	// PrintFilteredBlock outputs a Block
	PrintFilteredBlock(block *pb.FilteredBlock)

	// PrintChannels outputs the array of ChannelInfo
	PrintChannels(channels []*pb.ChannelInfo)

	// PrintChaincodes outputs the given array of ChaincodeInfo
	PrintChaincodes(chaincodes []*pb.ChaincodeInfo)

	// PrintProcessedTransaction outputs a ProcessedTransaction
	PrintProcessedTransaction(tx *pb.ProcessedTransaction)

	// PrintChaincodeData outputs ChaincodeData
	PrintChaincodeData(ccdata *ccprovider.ChaincodeData)

	// PrintTxProposalResponses outputs the proposal responses
	PrintTxProposalResponses(responses []*fab.TransactionProposalResponse, payloadOnly bool)

	// PrintResponses outputs responses
	PrintResponses(response []*pb.Response)

	// PrintChaincodeEvent outputs a chaincode event
	PrintChaincodeEvent(event *fab.CCEvent)

	// Print outputs a formatted string
	Print(frmt string, vars ...interface{})
}

// BlockPrinter is an implementation of BlockPrinter
type BlockPrinter struct {
	printer
}

// NewBlockPrinter returns a new Printer of the given OutputFormat and WriterType
func NewBlockPrinter(format OutputFormat, writerType WriterType) *BlockPrinter {
	return &BlockPrinter{
		printer: *newPrinter(format, writerType),
	}
}

// NewBlockPrinterWithOpts returns a new Printer of the given OutputFormat and WriterType
func NewBlockPrinterWithOpts(format OutputFormat, writerType WriterType, opts *FormatterOpts) *BlockPrinter {
	return &BlockPrinter{
		printer: *newPrinterWithOpts(format, writerType, opts),
	}
}

// PrintBlockchainInfo prints BlockchainInfo
func (p *BlockPrinter) PrintBlockchainInfo(info *fabriccmn.BlockchainInfo) {
	if p.Formatter == nil {
		fmt.Printf("%s\n", info)
		return
	}

	p.PrintHeader()
	p.Field("Height", info.Height)
	p.Field("CurrentBlockHash", Base64URLEncode(info.CurrentBlockHash))
	p.Field("PreviousBlockHash", Base64URLEncode(info.PreviousBlockHash))
	p.PrintFooter()
}

// PrintBlock prints a Block
func (p *BlockPrinter) PrintBlock(block *fabriccmn.Block) {
	if p.Formatter == nil {
		fmt.Printf("%s\n", block)
		return
	}

	p.PrintHeader()
	p.Element("Header")
	p.Field("Number", block.Header.Number)
	p.Field("PreviousHash", Base64URLEncode(block.Header.PreviousHash))
	p.Field("DataHash", Base64URLEncode(block.Header.DataHash))
	p.ElementEnd()

	p.Element("Metadata")
	p.PrintBlockMetadata(block.Metadata)
	p.ElementEnd()

	p.Element("Data")
	p.Array("Data")
	for i := range block.Data.Data {
		p.Item("Envelope", i)
		p.PrintEnvelope(utils.ExtractEnvelopeOrPanic(block, i))
		p.ItemEnd()
	}
	p.ArrayEnd()
	p.ElementEnd()
	p.PrintFooter()
}

// PrintFilteredBlock prints a FilteredBlock
func (p *BlockPrinter) PrintFilteredBlock(block *pb.FilteredBlock) {
	if p.Formatter == nil {
		fmt.Printf("%s\n", block)
		return
	}

	p.PrintHeader()
	p.Field("ChannelID", block.ChannelId)
	p.Field("Number", block.Number)

	p.Element("FilteredTransactions")
	p.Array("FilteredTransactions")
	for _, tx := range block.FilteredTransactions {
		p.PrintFilteredTransaction(tx)
	}
	p.ArrayEnd()
	p.ElementEnd()
	p.PrintFooter()
}

// PrintChannels prints the array of ChannelInfo
func (p *BlockPrinter) PrintChannels(channels []*pb.ChannelInfo) {
	if p.Formatter == nil {
		fmt.Printf("%s\n", channels)
		return
	}

	p.PrintHeader()
	p.Array("Channels")
	for _, channel := range channels {
		p.Field("ChannelId", channel.ChannelId)
	}
	p.ArrayEnd()
	p.PrintFooter()
}

// PrintChaincodes prints the array of ChaincodeInfo
func (p *BlockPrinter) PrintChaincodes(chaincodes []*pb.ChaincodeInfo) {
	if p.Formatter == nil {
		fmt.Printf("%s\n", chaincodes)
		return
	}

	p.PrintHeader()
	p.Array("")
	for _, ccInfo := range chaincodes {
		p.Item("ChaincodeInfo", ccInfo.Name)
		p.PrintChaincodeInfo(ccInfo)
		p.ItemEnd()
	}
	p.ArrayEnd()
	p.PrintFooter()
}

// PrintProcessedTransaction prints a ProcessedTransaction
func (p *BlockPrinter) PrintProcessedTransaction(tx *pb.ProcessedTransaction) {
	if p.Formatter == nil {
		fmt.Printf("%s\n", tx)
		return
	}

	p.PrintHeader()
	p.Print("ValidationCode: %s", pb.TxValidationCode(tx.ValidationCode))
	p.PrintEnvelope(tx.TransactionEnvelope)
	p.PrintFooter()
}

// PrintChaincodeData prints the given ChaincodeData
func (p *BlockPrinter) PrintChaincodeData(ccData *ccprovider.ChaincodeData) {
	if p.Formatter == nil {
		fmt.Printf("%s\n", ccData)
		return
	}

	p.PrintHeader()

	p.Field("Id", ccData.Id)
	p.Field("Name", ccData.Name)
	p.Field("Version", ccData.Version)
	p.Field("Escc", ccData.Escc)
	p.Field("Vscc", ccData.Vscc)

	cdsData := &ccprovider.CDSData{}
	unmarshalOrPanic(ccData.Data, cdsData)
	p.Element("Data")
	p.PrintCDSData(cdsData)
	p.ElementEnd()

	policy := &fabriccmn.SignaturePolicyEnvelope{}
	unmarshalOrPanic(ccData.Policy, policy)
	p.Element("Policy")
	p.PrintSignaturePolicyEnvelope(policy)
	p.ElementEnd()

	instPolicy := &fabriccmn.SignaturePolicyEnvelope{}
	unmarshalOrPanic(ccData.InstantiationPolicy, instPolicy)
	p.Element("InstantiationPolicy")
	p.PrintSignaturePolicyEnvelope(instPolicy)
	p.ElementEnd()

	p.PrintFooter()
}

// PrintTxProposalResponses prints the given transaction proposal responses
func (p *BlockPrinter) PrintTxProposalResponses(responses []*fab.TransactionProposalResponse, payloadOnly bool) {
	if p.Formatter == nil {
		for i, response := range responses {
			fmt.Printf("Response[%d]: %v\n", i, response)
		}
		return
	}

	p.PrintHeader()
	p.Array("")
	for i, response := range responses {
		p.Item("Response", i)
		p.PrintTxProposalResponse(response, payloadOnly)
		p.ItemEnd()
	}
	p.ArrayEnd()
	p.PrintFooter()
}

// PrintChaincodeEvent prints the given ChaincodeEvent
func (p *BlockPrinter) PrintChaincodeEvent(event *fab.CCEvent) {
	if p.Formatter == nil {
		fmt.Printf("%v\n", event)
		return
	}

	p.PrintHeader()
	p.Field("ChaincodeID", event.ChaincodeID)
	p.Field("EventName", event.EventName)
	//p.Field("ChannelID", event.ChannelID)
	p.Field("TxID", event.TxID)
	p.Field("Payload", event.Payload)
	p.PrintFooter()
}

// PrintTxProposalResponse prints the TransactionProposalResponse
func (p *BlockPrinter) PrintTxProposalResponse(response *fab.TransactionProposalResponse, payloadOnly bool) {
	if payloadOnly {
		if response.Status != http.StatusOK {
			p.Field("Err", response.Status)
		} else if response.ProposalResponse == nil || response.ProposalResponse.Response == nil {
			p.Field("Response", nil)
		} else {
			p.Field("Payload", response.ProposalResponse.Response.Payload)
		}
	} else {
		p.Field("Endorser", response.Endorser)
		p.Field("Status", response.Status)
		p.Element("ProposalResponse")
		p.PrintProposalResponse(response.ProposalResponse)
		p.ElementEnd()
	}
}

// PrintProposalResponse prints a ProposalResponse
func (p *BlockPrinter) PrintProposalResponse(response *pb.ProposalResponse) {
	if response == nil {
		return
	}
	p.Element("Response")
	p.PrintResponse(response.Response)
	p.ElementEnd()
	p.Field("Payload", response.Payload)

	prp := &pb.ProposalResponsePayload{}
	unmarshalOrPanic(response.Payload, prp)

	p.Element("ProposalResponsePayload")
	p.PrintProposalResponsePayload(prp)
	p.ElementEnd()

	p.Element("Endorsement")
	p.PrintEndorsement(response.Endorsement)
	p.ElementEnd()
}

// PrintResponses prints an array of Response
func (p *BlockPrinter) PrintResponses(responses []*pb.Response) {
	if p.Formatter == nil {
		for i, response := range responses {
			fmt.Printf("Response[%d]: %v\n", i, response)
		}
		return
	}

	p.PrintHeader()
	p.Array("")
	for i, response := range responses {
		p.Item("Response", i)
		p.PrintResponse(response)
		p.ItemEnd()
	}
	p.ArrayEnd()
	p.PrintFooter()
}

// PrintResponse prints the Response
func (p *BlockPrinter) PrintResponse(response *pb.Response) {
	p.Field("Message", response.Message)
	p.Field("Status", response.Status)
	p.Field("Payload", response.Payload)
}

// PrintCDSData prints the chaincode deployment spec data (CDSData)
func (p *BlockPrinter) PrintCDSData(cdsData *ccprovider.CDSData) {
	p.Field("CodeHash", Base64URLEncode(cdsData.CodeHash))
	p.Field("MetaDataHash", Base64URLEncode(cdsData.MetaDataHash))
}

// PrintEnvelope prints the given Envelope
func (p *BlockPrinter) PrintEnvelope(envelope *fabriccmn.Envelope) {
	p.Field("Signature", envelope.Signature)

	payload := utils.ExtractPayloadOrPanic(envelope)
	p.Element("Payload")
	p.PrintPayload(payload)
	p.ElementEnd()
}

// PrintPayload prints a Payload
func (p *BlockPrinter) PrintPayload(payload *fabriccmn.Payload) {
	p.Element("Header")

	chdr, err := utils.UnmarshalChannelHeader(payload.Header.ChannelHeader)
	if err != nil {
		panic(err)
	}

	p.Element("ChannelHeader")
	p.PrintChannelHeader(chdr)
	p.ElementEnd()

	sigHeader, err := utils.GetSignatureHeader(payload.Header.SignatureHeader)
	if err != nil {
		panic(err)
	}

	p.Element("SignatureHeader")
	p.PrintSignatureHeader(sigHeader)
	p.ElementEnd()

	p.ElementEnd() // Header

	p.Element("Data")
	p.Field("Type", fabriccmn.HeaderType(chdr.Type))
	p.PrintData(fabriccmn.HeaderType(chdr.Type), payload.Data)
	p.ElementEnd()
}

// PrintChannelHeader prints the ChannelHeader
func (p *BlockPrinter) PrintChannelHeader(chdr *fabriccmn.ChannelHeader) {
	p.Field("Type", fabriccmn.HeaderType(chdr.Type))
	p.Field("ChannelId", chdr.ChannelId)
	p.Field("Epoch", chdr.Epoch)

	ccHdrExt := &pb.ChaincodeHeaderExtension{}
	unmarshalOrPanic(chdr.Extension, ccHdrExt)
	p.Element("Extension")
	p.PrintChaincodeHeaderExtension(ccHdrExt)
	p.ElementEnd()

	p.Field("Timestamp", chdr.Timestamp)
	p.Field("TxId", chdr.TxId)
	p.Field("Version", chdr.Version)
}

// PrintChaincodeHeaderExtension prints the ChaincodeHeaderExtension
func (p *BlockPrinter) PrintChaincodeHeaderExtension(ccHdrExt *pb.ChaincodeHeaderExtension) {
	p.Element("ChaincodeId")
	p.PrintChaincodeID(ccHdrExt.ChaincodeId)
	p.ElementEnd()
	p.Field("PayloadVisibility", ccHdrExt.PayloadVisibility)
}

// PrintChaincodeID prints the ChaincodeID
func (p *BlockPrinter) PrintChaincodeID(ccID *pb.ChaincodeID) {
	if ccID == nil {
		return
	}
	p.Field("Name", ccID.Name)
	p.Field("Version", ccID.Version)
	p.Field("Path", ccID.Path)
}

// PrintChaincodeInfo prints ChaincodeInfo
func (p *BlockPrinter) PrintChaincodeInfo(ccInfo *pb.ChaincodeInfo) {
	p.Field("Name", ccInfo.Name)
	p.Field("Path", ccInfo.Path)
	p.Field("Version", ccInfo.Version)
	p.Field("Escc", ccInfo.Escc)
	p.Field("Vscc", ccInfo.Vscc)
	p.Field("Input", ccInfo.Input)
}

// PrintSignatureHeader prints a SignatureHeader
func (p *BlockPrinter) PrintSignatureHeader(sigHdr *fabriccmn.SignatureHeader) {
	p.Field("Nonce", sigHdr.Nonce)
	p.Field("Creator", sigHdr.Creator)
}

// PrintData prints the block of data formatted according to the given HeaderType
func (p *BlockPrinter) PrintData(headerType fabriccmn.HeaderType, data []byte) {
	if headerType == fabriccmn.HeaderType_CONFIG {
		envelope := &fabriccmn.ConfigEnvelope{}
		if err := proto.Unmarshal(data, envelope); err != nil {
			panic(errors.Errorf("Bad envelope: %v", err))
		}
		p.Print("Config Envelope:")
		p.PrintConfigEnvelope(envelope)
	} else if headerType == fabriccmn.HeaderType_CONFIG_UPDATE {
		envelope := &fabriccmn.ConfigUpdateEnvelope{}
		if err := proto.Unmarshal(data, envelope); err != nil {
			panic(errors.Errorf("Bad envelope: %v", err))
		}
		p.Print("Config Update Envelope:")
		p.PrintConfigUpdateEnvelope(envelope)
	} else if headerType == fabriccmn.HeaderType_ENDORSER_TRANSACTION {
		tx, err := utils.GetTransaction(data)
		if err != nil {
			panic(errors.Errorf("Bad envelope: %v", err))
		}
		p.Print("Transaction:")
		p.PrintTransaction(tx)
	} else {
		p.Field("Unsupported Envelope", Base64URLEncode(data))
	}
}

// PrintConfigEnvelope prints the ConfigEnvelope
func (p *BlockPrinter) PrintConfigEnvelope(envelope *fabriccmn.ConfigEnvelope) {
	p.Element("Config")
	p.PrintConfig(envelope.Config)
	p.ElementEnd()
	p.Element("LastUpdate")
	p.PrintEnvelope(envelope.LastUpdate)
	p.ElementEnd()
}

// PrintConfigUpdateEnvelope prints a ConfigUpdateEnvelope
func (p *BlockPrinter) PrintConfigUpdateEnvelope(envelope *fabriccmn.ConfigUpdateEnvelope) {
	p.Array("Signatures")
	for i, sig := range envelope.Signatures {
		p.Item("Config Signature", i)
		p.PrintConfigSignature(sig)
		p.ItemEnd()
	}
	p.ArrayEnd()

	configUpdate := &fabriccmn.ConfigUpdate{}
	if err := proto.Unmarshal(envelope.ConfigUpdate, configUpdate); err != nil {
		panic(err)
	}

	p.Element("ConfigUpdate")
	p.PrintConfigUpdate(configUpdate)
	p.ElementEnd()
}

// PrintTransaction prints a Transaction
func (p *BlockPrinter) PrintTransaction(tx *pb.Transaction) {
	p.Array("Actions")
	for i, action := range tx.Actions {
		p.Item("Action", i)
		p.PrintTXAction(action)
		p.ItemEnd()
	}
	p.ArrayEnd()
}

// PrintFilteredTransaction prints a FilteredTransaction
func (p *BlockPrinter) PrintFilteredTransaction(tx *pb.FilteredTransaction) {
	p.Field("Txid", tx.GetTxid())
	p.Field("Type", tx.GetType())
	p.Field("TxValidationCode", tx.GetTxValidationCode())
	p.Element("ChaincodeEvents")
	p.Array("ChaincodeEvents")
	txActions := tx.GetTransactionActions()
	for _, ccAction := range txActions.ChaincodeActions {
		p.PrintFilteredCCAction(ccAction)
	}
	p.ArrayEnd()
	p.ElementEnd()
}

// PrintTXAction prinbts a transaction action
func (p *BlockPrinter) PrintTXAction(action *pb.TransactionAction) {
	p.Element("Header")

	sigHeader, err := utils.GetSignatureHeader(action.Header)
	if err != nil {
		panic(err)
	}

	p.PrintSignatureHeader(sigHeader)
	p.ElementEnd()

	p.Element("Payload")

	chaPayload, err := utils.GetChaincodeActionPayload(action.Payload)
	if err != nil {
		panic(err)
	}

	p.PrintChaincodeActionPayload(chaPayload)
	p.ElementEnd()
}

// PrintFilteredCCAction prinbts a chaincode action
func (p *BlockPrinter) PrintFilteredCCAction(action *pb.FilteredChaincodeAction) {
	ccEvent := action.GetChaincodeEvent()
	p.Field("Txid", ccEvent.GetTxId())
	p.Field("ChaincodeId", ccEvent.GetChaincodeId())
	p.Field("EventName", ccEvent.GetEventName())
	p.Field("Payload", ccEvent.GetPayload())
}

// PrintChaincodeActionPayload prints a ChaincodeActionPayload
func (p *BlockPrinter) PrintChaincodeActionPayload(chaPayload *pb.ChaincodeActionPayload) {

	cpp := &pb.ChaincodeProposalPayload{}
	err := proto.Unmarshal(chaPayload.ChaincodeProposalPayload, cpp)
	if err != nil {
		panic(err)
	}

	p.Element("ChaincodeProposalPayload")
	p.PrintChaincodeProposalPayload(cpp)
	p.ElementEnd()

	p.Element("Action")
	p.PrintAction(chaPayload.Action)
	p.ElementEnd()
}

// PrintChaincodeProposalPayload prints a ChaincodeProposalPayload
func (p *BlockPrinter) PrintChaincodeProposalPayload(cpp *pb.ChaincodeProposalPayload) {
	cis := &pb.ChaincodeInvocationSpec{}
	err := proto.Unmarshal(cpp.Input, cis)
	if err != nil {
		panic(err)
	}

	p.Element("Input")
	p.PrintChaincodeInvocationSpec(cis)
	p.ElementEnd()

	p.Array("TransientMap")
	for key, value := range cpp.TransientMap {
		p.Item("Key", key)
		p.Field("Value", value)
		p.ItemEnd()
	}
	p.ArrayEnd()
}

// PrintChaincodeInvocationSpec prints a ChaincodeProposalPayload
func (p *BlockPrinter) PrintChaincodeInvocationSpec(cis *pb.ChaincodeInvocationSpec) {
	p.Element("ChaincodeSpec")
	p.PrintChaincodeSpec(cis.ChaincodeSpec)
	p.ElementEnd()
}

// PrintAction prints a ChaincodeEndorsedAction
func (p *BlockPrinter) PrintAction(action *pb.ChaincodeEndorsedAction) {
	p.Array("Endorsements")
	for i, endorsement := range action.Endorsements {
		p.Item("Endorsement", i)
		p.PrintEndorsement(endorsement)
		p.ItemEnd()
	}
	p.ArrayEnd()

	prp := &pb.ProposalResponsePayload{}
	unmarshalOrPanic(action.ProposalResponsePayload, prp)

	p.Element("ProposalResponsePayload")
	p.PrintProposalResponsePayload(prp)
	p.ElementEnd()
}

// PrintProposalResponsePayload prints a ProposalResponsePayload
func (p *BlockPrinter) PrintProposalResponsePayload(prp *pb.ProposalResponsePayload) {
	p.Field("ProposalHash", Base64URLEncode(prp.ProposalHash))

	chaincodeAction := &pb.ChaincodeAction{}
	unmarshalOrPanic(prp.Extension, chaincodeAction)
	p.Element("Extension")
	p.PrintChaincodeAction(chaincodeAction)
	p.ElementEnd()
}

// PrintChaincodeAction prints a ChaincodeAction
func (p *BlockPrinter) PrintChaincodeAction(chaincodeAction *pb.ChaincodeAction) {
	p.Element("Response")
	p.PrintChaincodeResponse(chaincodeAction.Response)
	p.ElementEnd()

	p.Element("Results")
	if len(chaincodeAction.Results) > 0 {
		txRWSet := &rwsetutil.TxRwSet{}
		if err := txRWSet.FromProtoBytes(chaincodeAction.Results); err != nil {
			panic(err)
		}

		p.PrintTxReadWriteSet(txRWSet)
	}
	p.ElementEnd()

	p.Element("Events")
	if len(chaincodeAction.Events) > 0 {
		chaincodeEvent := &pb.ChaincodeEvent{}
		unmarshalOrPanic(chaincodeAction.Events, chaincodeEvent)
		p.PrintChaincodeEventFromBlock(chaincodeEvent)
	}
	p.ElementEnd()
}

// PrintTxReadWriteSet prints a transaction read-write set (TxRwSet)
func (p *BlockPrinter) PrintTxReadWriteSet(txRWSet *rwsetutil.TxRwSet) {
	p.Array("NsRWs")
	for i, nsRWSet := range txRWSet.NsRwSets {
		p.Item("TxRwSet", i)
		p.PrintNsReadWriteSet(nsRWSet)
		p.ItemEnd()
	}
	p.ArrayEnd()
}

// PrintNsReadWriteSet prints a namespaced read-write set (NsRwSet)
func (p *BlockPrinter) PrintNsReadWriteSet(nsRWSet *rwsetutil.NsRwSet) {
	p.Field("NameSpace", nsRWSet.NameSpace)

	p.Element("KvRwSet")
	p.PrintKvRwSet(nsRWSet.KvRwSet)
	p.ElementEnd()

	p.Element("CollHashedRwSets")
	p.PrintCollHashedRwSets(nsRWSet.CollHashedRwSets)
	p.ElementEnd()
}

// PrintKvRwSet prints a key-value read-write set
func (p *BlockPrinter) PrintKvRwSet(kvRWSet *kvrwset.KVRWSet) {
	p.Array("Reads")
	for i, r := range kvRWSet.Reads {
		p.Item("Read", i)
		p.PrintRead(r)
		p.ItemEnd()
	}
	p.ArrayEnd()

	p.Array("Writes")
	for i, w := range kvRWSet.Writes {
		p.Item("Write", i)
		p.PrintWrite(w)
		p.ItemEnd()
	}
	p.ArrayEnd()

	p.Array("RangeQueriesInfo")
	for i, rqi := range kvRWSet.RangeQueriesInfo {
		p.Item("RangeQueryInfo", i)
		p.PrintRangeQueryInfo(rqi)
		p.ItemEnd()
	}
	p.ArrayEnd()
}

// PrintCollHashedRwSets prints an array of collection hashed read-write set
func (p *BlockPrinter) PrintCollHashedRwSets(collHashedRwSets []*rwsetutil.CollHashedRwSet) {
	p.Array("CollHashedRwSets")
	for i, w := range collHashedRwSets {
		p.Item("CollHashedRwSet", i)
		p.PrintCollHashedRwSet(w)
		p.ItemEnd()
	}
	p.ArrayEnd()
}

// PrintCollHashedRwSet prints a collection hashed read-write set
func (p *BlockPrinter) PrintCollHashedRwSet(collHashedRwSet *rwsetutil.CollHashedRwSet) {
	p.Field("CollectionName", collHashedRwSet.CollectionName)
	p.Field("PvtRwSetHash", Base64URLEncode(collHashedRwSet.PvtRwSetHash))

	p.Element("HashedRwSet")
	p.PrintHashedRwSet(collHashedRwSet.HashedRwSet)
	p.ElementEnd()
}

// PrintHashedRwSet prints a HashedRWSet
func (p *BlockPrinter) PrintHashedRwSet(hashedRwSet *kvrwset.HashedRWSet) {
	p.PrintHashedReads(hashedRwSet.HashedReads)
	p.PrintHashedWrites(hashedRwSet.HashedWrites)
}

// PrintHashedReads prints an array of key-value read hashes (KVReadHash)
func (p *BlockPrinter) PrintHashedReads(hashedReads []*kvrwset.KVReadHash) {
	p.Array("HashedReads")
	for i, r := range hashedReads {
		p.Item("HashedRead", i)
		p.PrintHashedRead(r)
		p.ItemEnd()
	}
	p.ArrayEnd()
}

// PrintHashedWrites prints an array of key-value write hashes (KVWriteHash)
func (p *BlockPrinter) PrintHashedWrites(hashedWrites []*kvrwset.KVWriteHash) {
	p.Array("HashedWrites")
	for i, r := range hashedWrites {
		p.Item("HashedWrite", i)
		p.PrintHashedWrite(r)
		p.ItemEnd()
	}
	p.ArrayEnd()
}

// PrintHashedRead prints a key-value read hash (KVReadHash)
func (p *BlockPrinter) PrintHashedRead(hashedRead *kvrwset.KVReadHash) {
	p.Field("KeyHash", Base64URLEncode(hashedRead.KeyHash))
	p.PrintVersion(hashedRead.Version)
}

// PrintHashedWrite prints a key-value write hash (KVWriteHash)
func (p *BlockPrinter) PrintHashedWrite(hashedWrite *kvrwset.KVWriteHash) {
	p.Field("KeyHash", Base64URLEncode(hashedWrite.KeyHash))
	p.Field("ValueHash", Base64URLEncode(hashedWrite.ValueHash))
	p.Field("IsDelete", hashedWrite.IsDelete)
}

// PrintRangeQueryInfo prints a RangeQueryInfo
func (p *BlockPrinter) PrintRangeQueryInfo(rqi *kvrwset.RangeQueryInfo) {
	p.Field("StartKey", rqi.StartKey)
	p.Field("EndKey", rqi.EndKey)
	p.Field("ItrExhausted", rqi.ItrExhausted)
	p.Element("ReadsInfo")
	p.PrintReadsInfo(rqi.ReadsInfo)
	p.ElementEnd()
}

// PrintReadsInfo prints a RangeQuery reads info
func (p *BlockPrinter) PrintReadsInfo(ri interface{}) {
	switch x := ri.(type) {
	case *kvrwset.RangeQueryInfo_RawReads:
		p.PrintRawReads(x.RawReads)
	case *kvrwset.RangeQueryInfo_ReadsMerkleHashes:
		p.PrintReadsMerkleHashes(x.ReadsMerkleHashes)
	case nil:
	default:
		p.Print("unknown type: %v", reflect.TypeOf(ri))
	}
}

// PrintRawReads prints QueryReads
func (p *BlockPrinter) PrintRawReads(qr *kvrwset.QueryReads) {
	p.Array("QueryReads")
	for i, r := range qr.KvReads {
		p.Item("Read", i)
		p.PrintRead(r)
		p.ItemEnd()
	}
	p.ArrayEnd()
}

// PrintReadsMerkleHashes prints QueryReadsMerkleSummary
func (p *BlockPrinter) PrintReadsMerkleHashes(qr *kvrwset.QueryReadsMerkleSummary) {
	p.Field("MaxDegree", qr.MaxDegree)
	p.Field("MaxLevel", qr.MaxLevel)

	p.Array("MaxLevelHashes")
	for i, r := range qr.MaxLevelHashes {
		p.ItemValue("Hash", i, Base64URLEncode(r))
	}
	p.ArrayEnd()
}

// PrintRead prints a
func (p *BlockPrinter) PrintRead(r *kvrwset.KVRead) {
	p.Field("Key", r.Key)
	p.Element("Version")
	p.PrintVersion(r.Version)
	p.ElementEnd()
}

// PrintVersion print a Version
func (p *BlockPrinter) PrintVersion(version *kvrwset.Version) {
	if version == nil {
		return
	}

	p.Field("BlockNum", version.BlockNum)
	p.Field("TxNum", version.TxNum)
}

// PrintWrite prints a key-value write (KVWrite)
func (p *BlockPrinter) PrintWrite(w *kvrwset.KVWrite) {
	p.Field("Key", w.Key)
	p.Field("IsDelete", w.IsDelete)
	p.Field("Value", w.Value)
}

// PrintChaincodeResponse prints a response
func (p *BlockPrinter) PrintChaincodeResponse(response *pb.Response) {
	p.Field("Message", response.Message)
	p.Field("Status", response.Status)
	p.Field("Payload", response.Payload)
}

// PrintChaincodeEventFromBlock prints a ChaincodeEvent
func (p *BlockPrinter) PrintChaincodeEventFromBlock(chaincodeEvent *pb.ChaincodeEvent) {
	p.Field("ChaincodeId", chaincodeEvent.ChaincodeId)
	p.Field("EventName", chaincodeEvent.EventName)
	p.Field("TxID", chaincodeEvent.TxId)
	p.Field("Payload", chaincodeEvent.Payload)
}

// PrintEndorsement prints an Endorsement
func (p *BlockPrinter) PrintEndorsement(endorsement *pb.Endorsement) {
	p.Field("Endorser", endorsement.Endorser)
	p.Field("Signature", endorsement.Signature)
}

// PrintConfig prints a Config
func (p *BlockPrinter) PrintConfig(config *fabriccmn.Config) {
	p.Field("Sequence", config.Sequence)
	p.Element("ChannelGroup")
	p.PrintConfigGroup(config.ChannelGroup)
	p.ElementEnd()
}

// PrintConfigGroup prints a ConfigGroup
func (p *BlockPrinter) PrintConfigGroup(configGroup *fabriccmn.ConfigGroup) {
	if configGroup == nil {
		return
	}

	p.Field("Version", configGroup.Version)
	p.Field("ModPolicy", configGroup.ModPolicy)
	p.Array("Groups")
	for key, grp := range configGroup.Groups {
		p.Item("Group", key)
		p.PrintConfigGroup(grp)
		p.ItemEnd()
	}
	p.ArrayEnd()

	p.Array("Values")
	for key, val := range configGroup.Values {
		p.Item("Value", key)
		p.PrintConfigValue(key, val)
		p.ItemEnd()
	}
	p.ArrayEnd()

	p.Array("Policies")
	for key, policy := range configGroup.Policies {
		p.Item("Policy", key)
		p.PrintConfigPolicy(policy)
		p.ItemEnd()
	}
	p.ArrayEnd()
}

// PrintConfigPolicy prints a ConfigPolicy
func (p *BlockPrinter) PrintConfigPolicy(policy *fabriccmn.ConfigPolicy) {
	p.Field("ModPolicy", policy.ModPolicy)
	p.Field("Version", policy.Version)

	if policy.Policy == nil {
		return
	}

	policyType := fabriccmn.Policy_PolicyType(policy.Policy.Type)
	switch policyType {
	case fabriccmn.Policy_SIGNATURE:
		sigPolicyEnv := &fabriccmn.SignaturePolicyEnvelope{}
		unmarshalOrPanic(policy.Policy.Value, sigPolicyEnv)
		p.Print("Signature Policy:")
		p.PrintSignaturePolicyEnvelope(sigPolicyEnv)
		break

	case fabriccmn.Policy_MSP:
		p.Print("Policy_MSP: TODO")
		break

	case fabriccmn.Policy_IMPLICIT_META:
		impMetaPolicy := &fabriccmn.ImplicitMetaPolicy{}
		unmarshalOrPanic(policy.Policy.Value, impMetaPolicy)
		p.Print("Implicit Meta Policy:")
		p.PrintImplicitMetaPolicy(impMetaPolicy)
		break

	default:
		break
	}
}

// PrintCapabilities prints capabilities
func (p *BlockPrinter) PrintCapabilities(capabilities *fabriccmn.Capabilities) {
	for capability := range capabilities.Capabilities {
		p.Field("Capability", capability)
	}
}

// PrintImplicitMetaPolicy prints an ImplicitMetaPolicy
func (p *BlockPrinter) PrintImplicitMetaPolicy(impMetaPolicy *fabriccmn.ImplicitMetaPolicy) {
	p.Field("Rule", impMetaPolicy.Rule)
	p.Field("Sub-policy", impMetaPolicy.SubPolicy)
}

// PrintSignaturePolicyEnvelope prints a SignaturePolicyEnvelope
func (p *BlockPrinter) PrintSignaturePolicyEnvelope(sigPolicyEnv *fabriccmn.SignaturePolicyEnvelope) {
	p.Element("Rule")
	p.PrintSignaturePolicy(sigPolicyEnv.Rule)
	p.ElementEnd()

	p.Array("Identities")
	for i, identity := range sigPolicyEnv.Identities {
		p.Item("Identity", i)
		p.PrintMSPPrincipal(identity)
		p.ItemEnd()
	}
	p.ArrayEnd()
}

// PrintMSPPrincipal prints a MSPPrincipal
func (p *BlockPrinter) PrintMSPPrincipal(principal *msp.MSPPrincipal) {
	p.Field("PrincipalClassification", principal.PrincipalClassification)
	p.Element("Principal")
	switch principal.PrincipalClassification {
	case msp.MSPPrincipal_ROLE:
		// Principal contains the msp role
		mspRole := &msp.MSPRole{}
		unmarshalOrPanic(principal.Principal, mspRole)
		p.Field("Role", mspRole.Role)
		p.Field("MspIdentifier", mspRole.MspIdentifier)
	case msp.MSPPrincipal_IDENTITY:
		p.Value(principal.Principal)
	case msp.MSPPrincipal_ORGANIZATION_UNIT:
		// Principal contains the OrganizationUnit
		unit := &msp.OrganizationUnit{}
		unmarshalOrPanic(principal.Principal, unit)

		p.Field("MspIdentifier", unit.MspIdentifier)
		p.Field("OrganizationalUnitIdentifier", unit.OrganizationalUnitIdentifier)
		p.Field("CertifiersIdentifier", unit.CertifiersIdentifier)
	default:
		p.Value("unknown PrincipalClassification")
	}
	p.ElementEnd()
}

// PrintSignaturePolicy prints a SignaturePolicy
func (p *BlockPrinter) PrintSignaturePolicy(sigPolicy *fabriccmn.SignaturePolicy) {
	switch t := sigPolicy.Type.(type) {
	case *fabriccmn.SignaturePolicy_SignedBy:
		p.PrintSignaturePolicySignedBy(t)
		break
	case *fabriccmn.SignaturePolicy_NOutOf_:
		p.PrintSignaturePolicyNOutOf(t.NOutOf)
		break
	}
}

// PrintSignaturePolicySignedBy prints a SignaturePolicy_SignedBy policy
func (p *BlockPrinter) PrintSignaturePolicySignedBy(sigPolicy *fabriccmn.SignaturePolicy_SignedBy) {
	p.Field("SignaturePolicy_SignedBy", sigPolicy.SignedBy)
}

// PrintSignaturePolicyNOutOf prints a SignaturePolicy_NOutOf policy
func (p *BlockPrinter) PrintSignaturePolicyNOutOf(sigPolicy *fabriccmn.SignaturePolicy_NOutOf) {
	p.Print("SignaturePolicy_NOutOf")
	p.Field("N", sigPolicy.N)
	p.Array("Rules")
	for i, policy := range sigPolicy.Rules {
		p.Item("Rule", i)
		p.PrintSignaturePolicy(policy)
		p.ItemEnd()
	}
	p.ArrayEnd()
}

// PrintConfigValue prints a ConfigValue
func (p *BlockPrinter) PrintConfigValue(name string, value *fabriccmn.ConfigValue) {
	p.Field("Version", value.Version)
	p.Field("ModPolicy", value.ModPolicy)

	switch name {
	case AnchorPeersKey:
		anchorPeers := &pb.AnchorPeers{}
		unmarshalOrPanic(value.Value, anchorPeers)
		p.Element("AnchorPeers")
		p.PrintAnchorPeers(anchorPeers)
		p.ElementEnd()
		break

	case MSPKey:
		mspConfig := &msp.MSPConfig{}
		unmarshalOrPanic(value.Value, mspConfig)
		p.Element("MSPConfig")
		p.PrintMSPConfig(mspConfig)
		p.ElementEnd()
		break

	case ConsensusTypeKey:
		consensusType := &ab.ConsensusType{}
		unmarshalOrPanic(value.Value, consensusType)
		p.Field("ConsensusType", consensusType.Type)
		break

	case BatchSizeKey:
		batchSize := &ab.BatchSize{}
		unmarshalOrPanic(value.Value, batchSize)
		p.Element("BatchSize")
		p.PrintBatchSize(batchSize)
		p.ElementEnd()
		break

	case BatchTimeoutKey:
		batchTimeout := &ab.BatchTimeout{}
		unmarshalOrPanic(value.Value, batchTimeout)
		p.Field("Timeout", batchTimeout.Timeout)
		break

	case ChannelRestrictionsKey:
		chRestrictions := &ab.ChannelRestrictions{}
		unmarshalOrPanic(value.Value, chRestrictions)
		p.Element("ChannelRestrictions")
		p.Field("MaxCount", chRestrictions.MaxCount)
		p.ElementEnd()
		break

	case ChannelCreationPolicyKey:
		creationPolicy := &fabriccmn.ConfigPolicy{}
		unmarshalOrPanic(value.Value, creationPolicy)
		p.Element("ChannelCreationPolicy")
		p.PrintConfigPolicy(creationPolicy)
		p.ElementEnd()
		break

	// case ChainCreationPolicyNamesKey:
	// 	chainCreationPolicyNames := &ab.ChainCreationPolicyNames{}
	// 	unmarshalOrPanic(value.Value, chainCreationPolicyNames)
	// 	p.Print("ChainCreationPolicyNames")
	// 	p.Field("Names", chainCreationPolicyNames.Names)
	// 	break

	case HashingAlgorithmKey:
		hashingAlgorithm := &fabriccmn.HashingAlgorithm{}
		unmarshalOrPanic(value.Value, hashingAlgorithm)
		p.Element("HashingAlgorithm")
		p.Field("Name", hashingAlgorithm.Name)
		p.ElementEnd()
		break

	case BlockDataHashingStructureKey:
		hashingStructure := &fabriccmn.BlockDataHashingStructure{}
		unmarshalOrPanic(value.Value, hashingStructure)
		p.Element("BlockDataHashingStructure")
		p.Field("Width", hashingStructure.Width)
		p.ElementEnd()
		break

	case OrdererAddressesKey:
		addresses := &fabriccmn.OrdererAddresses{}
		unmarshalOrPanic(value.Value, addresses)
		p.Element("OrdererAddresses")
		p.Field("Addresses", addresses.Addresses)
		p.ElementEnd()
		break

	case ConsortiumKey:
		consortium := &fabriccmn.Consortium{}
		unmarshalOrPanic(value.Value, consortium)
		p.Element("Consortium")
		p.Field("Name", consortium.Name)
		p.ElementEnd()
		break

	case CapabilitiesKey:
		capabilities := &fabriccmn.Capabilities{}
		unmarshalOrPanic(value.Value, capabilities)
		p.Element("Capabilities")
		p.PrintCapabilities(capabilities)
		p.ElementEnd()
		break

	default:
		p.Print("!!!!!!! Don't know how to Print config value: %s", name)
		break
	}
}

// PrintBatchSize prints a BatchSize
func (p *BlockPrinter) PrintBatchSize(batchSize *ab.BatchSize) {
	p.Field("MaxMessageCount", batchSize.MaxMessageCount)
	p.Field("AbsoluteMaxBytes", batchSize.AbsoluteMaxBytes)
	p.Field("PreferredMaxBytes", batchSize.PreferredMaxBytes)
}

// ProviderType indicates the type of an identity provider
type ProviderType int

// The ProviderType of a member relative to the member API
const (
	FABRIC ProviderType = iota // MSP is of FABRIC type
	IDEMIX                     // MSP is of IDEMIX type
	OTHER                      // MSP is of OTHER TYPE
)

// PrintMSPConfig prints a MSPConfig
func (p *BlockPrinter) PrintMSPConfig(mspConfig *msp.MSPConfig) {
	switch ProviderType(mspConfig.Type) {
	case FABRIC:
		p.Print("Type: FABRIC")

		config := &msp.FabricMSPConfig{}
		unmarshalOrPanic(mspConfig.Config, config)

		p.Print("Config:")
		p.PrintFabricMSPConfig(config)
		break

	default:
		p.Print("Type: OTHER")
		break
	}
}

// PrintFabricMSPConfig prints a FabricMSPConfig
func (p *BlockPrinter) PrintFabricMSPConfig(mspConfig *msp.FabricMSPConfig) {
	p.Field("Name", mspConfig.Name)
	p.Array("Admins")
	for i, admCert := range mspConfig.Admins {
		p.ItemValue("Admin Cert", i, admCert)
	}
	p.ArrayEnd()
}

// PrintAnchorPeers prints AnchorPeers
func (p *BlockPrinter) PrintAnchorPeers(anchorPeers *pb.AnchorPeers) {
	p.Array("AnchorPeers")
	for i, anchorPeer := range anchorPeers.AnchorPeers {
		p.Item("Anchor Peer", i)
		p.PrintAnchorPeer(anchorPeer)
		p.ItemEnd()
	}
	p.ArrayEnd()
}

// PrintAnchorPeer prints an AnchorPeer
func (p *BlockPrinter) PrintAnchorPeer(anchorPeer *pb.AnchorPeer) {
	p.Field("Host", anchorPeer.Host)
	p.Field("Port", anchorPeer.Port)
}

// PrintConfigSignature prints ConfigSignature
func (p *BlockPrinter) PrintConfigSignature(sig *fabriccmn.ConfigSignature) {
	p.Field("Signature", sig.Signature)

	sigHeader, err := utils.GetSignatureHeader(sig.SignatureHeader)
	if err != nil {
		panic(err)
	}

	p.Element("SignatureHeader")
	p.PrintSignatureHeader(sigHeader)
	p.ElementEnd()
}

// PrintConfigUpdate prints ConfigUpdate
func (p *BlockPrinter) PrintConfigUpdate(configUpdate *fabriccmn.ConfigUpdate) {
	p.Field("ChannelId", configUpdate.ChannelId)
	p.Element("ReadSet")
	p.PrintConfigGroup(configUpdate.ReadSet)
	p.ElementEnd()
	p.Element("WriteSet")
	p.PrintConfigGroup(configUpdate.WriteSet)
	p.ElementEnd()
}

// PrintBlockMetadata prints BlockMetadata
func (p *BlockPrinter) PrintBlockMetadata(blockMetaData *fabriccmn.BlockMetadata) {
	p.PrintSignaturesMetadata(getMetadataOrPanic(blockMetaData, fabriccmn.BlockMetadataIndex_SIGNATURES))
	p.PrintLastConfigMetadata(getMetadataOrPanic(blockMetaData, fabriccmn.BlockMetadataIndex_LAST_CONFIG))
	p.PrintTransactionsFilterMetadata(ledgerUtil.TxValidationFlags(blockMetaData.Metadata[fabriccmn.BlockMetadataIndex_TRANSACTIONS_FILTER]))
	p.PrintOrdererMetadata(getMetadataOrPanic(blockMetaData, fabriccmn.BlockMetadataIndex_ORDERER))
}

// PrintSignaturesMetadata prints signature Metadata
func (p *BlockPrinter) PrintSignaturesMetadata(metadata *fabriccmn.Metadata) {
	p.Array("Signatures")
	for i, metadataSignature := range metadata.Signatures {
		shdr, err := utils.GetSignatureHeader(metadataSignature.SignatureHeader)
		if err != nil {
			panic(errors.Errorf("Failed unmarshalling meta data signature header. Error: %v", err))
		}
		p.Item("Metadata Signature", i)
		p.PrintSignatureHeader(shdr)
		p.ItemEnd()
	}
	p.ArrayEnd()
}

// PrintLastConfigMetadata prints last-config MetaData
func (p *BlockPrinter) PrintLastConfigMetadata(metadata *fabriccmn.Metadata) {
	lastConfig := &fabriccmn.LastConfig{}
	err := proto.Unmarshal(metadata.Value, lastConfig)
	if err != nil {
		panic(errors.Errorf("Failed unmarshalling meta data last config. Error: %v", err))
	}

	p.Field("Last Config Index", lastConfig.Index)
}

// PrintTransactionsFilterMetadata prints TxValidationFlags
func (p *BlockPrinter) PrintTransactionsFilterMetadata(txFilter ledgerUtil.TxValidationFlags) {
	p.Array("TransactionFilters")
	for i := 0; i < len(txFilter); i++ {
		p.ItemValue("TxValidationFlag", i, txFilter.Flag(i))
	}
	p.ArrayEnd()
}

// PrintOrdererMetadata prints orderer Metadata
func (p *BlockPrinter) PrintOrdererMetadata(metadata *fabriccmn.Metadata) {
	p.Field("Orderer Metadata", metadata.Value)
}

// PrintChaincodeDeploymentSpec prints ChaincodeDeploymentSpec
func (p *BlockPrinter) PrintChaincodeDeploymentSpec(depSpec *pb.ChaincodeDeploymentSpec) {
	p.Element("ChaincodeSpec")
	p.PrintChaincodeSpec(depSpec.ChaincodeSpec)
	p.ElementEnd()
}

// PrintChaincodeSpec prints ChaincodeSpec
func (p *BlockPrinter) PrintChaincodeSpec(ccSpec *pb.ChaincodeSpec) {
	p.Element("ChaincodeId")
	p.PrintChaincodeID(ccSpec.ChaincodeId)
	p.ElementEnd()

	p.Field("Timeout", ccSpec.Timeout)
	p.Field("Type", ccSpec.Type)

	p.Element("Input")
	p.PrintChaincodeInput(ccSpec.Input)
	p.ElementEnd()
}

// PrintChaincodeInput prints ChaincodeInput
func (p *BlockPrinter) PrintChaincodeInput(ccInput *pb.ChaincodeInput) {
	p.Array("Args")
	for i, value := range ccInput.Args {
		p.ItemValue("Arg", i, value)
	}
	p.ArrayEnd()
	p.Array("Decorations")
	for key, value := range ccInput.Decorations {
		p.ItemValue("Decoration", key, value)
	}
	p.ArrayEnd()
}

func getMetadataOrPanic(blockMetaData *fabriccmn.BlockMetadata, index fabriccmn.BlockMetadataIndex) *fabriccmn.Metadata {
	metaData := &fabriccmn.Metadata{}
	err := proto.Unmarshal(blockMetaData.Metadata[index], metaData)
	if err != nil {
		panic(errors.Errorf("Unable to unmarshal meta data at index %d", index))
	}
	return metaData
}

func unmarshalOrPanic(buf []byte, pb proto.Message) {
	err := proto.Unmarshal(buf, pb)
	if err != nil {
		panic(err)
	}
}

// Base64URLEncode encodes the byte array into a base64 string
func Base64URLEncode(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}

// Base64URLDecode decodes the base64 string into a byte array
func Base64URLDecode(data string) ([]byte, error) {
	//check if it has padding or not
	if strings.HasSuffix(data, "=") {
		return base64.URLEncoding.DecodeString(data)
	}
	return base64.RawURLEncoding.DecodeString(data)
}
