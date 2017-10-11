/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package printer

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	fabriccmn "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/ledger/rwset/kvrwset"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/msp"
	ab "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/orderer"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
	"github.com/securekey/fabric-examples/fabric-cli/internal/github.com/hyperledger/fabric/common/configtx"
	"github.com/securekey/fabric-examples/fabric-cli/internal/github.com/hyperledger/fabric/core/common/ccprovider"
	"github.com/securekey/fabric-examples/fabric-cli/internal/github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/rwsetutil"
	ledgerUtil "github.com/securekey/fabric-examples/fabric-cli/internal/github.com/hyperledger/fabric/core/ledger/util"
	fabricMsp "github.com/securekey/fabric-examples/fabric-cli/internal/github.com/hyperledger/fabric/msp"
	fabricUtils "github.com/securekey/fabric-examples/fabric-cli/internal/github.com/hyperledger/fabric/protos/utils"
)

const (
	indentSize = 3

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

	// GroupKey is the name of the channel group
	ChannelGroupKey = "Channel"
)

// Printer is used for printing various data structures
type Printer interface {
	// PrintBlockchainInfo outputs BlockchainInfo
	PrintBlockchainInfo(info *fabriccmn.BlockchainInfo)

	// PrintBlock outputs a Block
	PrintBlock(block *fabriccmn.Block)

	// PrintChannels outputs the array of ChannelInfo
	PrintChannels(channels []*pb.ChannelInfo)

	// PrintChaincodes outputs the given array of ChaincodeInfo
	PrintChaincodes(chaincodes []*pb.ChaincodeInfo)

	// PrintProcessedTransaction outputs a ProcessedTransaction
	PrintProcessedTransaction(tx *pb.ProcessedTransaction)

	// PrintChaincodeData outputs ChaincodeData
	PrintChaincodeData(ccdata *ccprovider.ChaincodeData)

	// PrintTxProposalResponses outputs the proposal responses
	PrintTxProposalResponses(responses []*apitxn.TransactionProposalResponse)

	// PrintResponses outputs responses
	PrintResponses(response []*pb.Response)

	// PrintChaincodeEvent outputs a chaincode event
	PrintChaincodeEvent(event *apifabclient.ChaincodeEvent)

	// Print outputs a formatted string
	Print(frmt string, vars ...interface{})
}

// PrinterImpl is an implementation of Printer
type PrinterImpl struct {
	Formatter Formatter
}

// NewPrinter returns a new Printer of the given OutputFormat and WriterType
func NewPrinter(format OutputFormat, writerType WriterType) *PrinterImpl {
	return &PrinterImpl{Formatter: NewFormatter(format, writerType)}
}

func (p *PrinterImpl) PrintBlockchainInfo(info *fabriccmn.BlockchainInfo) {
	if p.Formatter == nil {
		fmt.Printf("%s\n", info)
		return
	}

	p.PrintHeader()
	// p.Print("Block Info for chain %s", Config().ChannelID())
	p.Field("Height", info.Height)
	p.Field("CurrentBlockHash", Base64URLEncode(info.CurrentBlockHash))
	p.Field("PreviousBlockHash", Base64URLEncode(info.PreviousBlockHash))
	p.PrintFooter()
}

func (p *PrinterImpl) PrintBlock(block *fabriccmn.Block) {
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
	numEnvelopes := len(block.Data.Data)
	for i := 0; i < numEnvelopes; i++ {
		p.Item("Envelope", i)
		p.PrintEnvelope(fabricUtils.ExtractEnvelopeOrPanic(block, i))
		p.ItemEnd()
	}
	p.ArrayEnd()
	p.ElementEnd()
	p.PrintFooter()
}

func (p *PrinterImpl) PrintChannels(channels []*pb.ChannelInfo) {
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

func (p *PrinterImpl) PrintChaincodes(chaincodes []*pb.ChaincodeInfo) {
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

func (p *PrinterImpl) PrintProcessedTransaction(tx *pb.ProcessedTransaction) {
	if p.Formatter == nil {
		fmt.Printf("%s\n", tx)
		return
	}

	p.PrintHeader()
	p.Print("ValidationCode: %s", pb.TxValidationCode(tx.ValidationCode))
	p.PrintEnvelope(tx.TransactionEnvelope)
	p.PrintFooter()
}

func (p *PrinterImpl) PrintChaincodeData(ccData *ccprovider.ChaincodeData) {
	if p.Formatter == nil {
		fmt.Printf("%s\n", ccData)
		return
	}

	p.PrintHeader()

	p.Field("Id", Base64URLEncode(ccData.Id))
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

func (p *PrinterImpl) PrintTxProposalResponses(responses []*apitxn.TransactionProposalResponse) {
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
		p.PrintTxProposalResponse(response)
		p.ItemEnd()
	}
	p.ArrayEnd()
	p.PrintFooter()
}

func (p *PrinterImpl) PrintChaincodeEvent(event *apifabclient.ChaincodeEvent) {
	if p.Formatter == nil {
		fmt.Printf("%v\n", event)
		return
	}

	p.PrintHeader()
	p.Field("ChaincodeID", event.ChaincodeID)
	p.Field("EventName", event.EventName)
	p.Field("ChannelID", event.ChannelID)
	p.Field("TxID", event.TxID)
	p.Field("Payload", Base64URLEncode(event.Payload))
	p.PrintFooter()
}

func (p *PrinterImpl) PrintTxProposalResponse(response *apitxn.TransactionProposalResponse) {
	p.Field("Endorser", response.Endorser)
	p.Field("Err", response.Err)
	p.Field("Status", response.Status)
	p.Element("ProposalResponse")
	p.PrintProposalResponse(response.ProposalResponse)
	p.ElementEnd()
}

func (p *PrinterImpl) PrintProposalResponse(response *pb.ProposalResponse) {
	if response == nil {
		return
	}
	p.Element("Response")
	p.PrintResponse(response.Response)
	p.ElementEnd()
	p.Field("Payload", string(response.Payload))
	p.Element("Endorsement")
	p.PrintEndorsement(response.Endorsement)
	p.ElementEnd()
}

func (p *PrinterImpl) PrintResponses(responses []*pb.Response) {
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

func (p *PrinterImpl) PrintResponse(response *pb.Response) {
	p.Field("Message", response.Message)
	p.Field("Status", response.Status)
	p.Field("Payload", string(response.Payload))
}

func (p *PrinterImpl) PrintCDSData(cdsData *ccprovider.CDSData) {
	p.Field("CodeHash", Base64URLEncode(cdsData.CodeHash))
	p.Field("MetaDataHash", Base64URLEncode(cdsData.MetaDataHash))
}

func (p *PrinterImpl) PrintEnvelope(envelope *fabriccmn.Envelope) {
	p.Field("Signature", Base64URLEncode(envelope.Signature))

	payload := fabricUtils.ExtractPayloadOrPanic(envelope)
	p.Element("Payload")
	p.PrintPayload(payload)
	p.ElementEnd()
}

func (p *PrinterImpl) PrintPayload(payload *fabriccmn.Payload) {
	p.Element("Header")

	chdr, err := fabricUtils.UnmarshalChannelHeader(payload.Header.ChannelHeader)
	if err != nil {
		panic(err)
	}

	p.Element("ChannelHeader")
	p.PrintChannelHeader(chdr)
	p.ElementEnd()

	sigHeader, err := fabricUtils.GetSignatureHeader(payload.Header.SignatureHeader)
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

func (p *PrinterImpl) PrintChannelHeader(chdr *fabriccmn.ChannelHeader) {
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

func (p *PrinterImpl) PrintChaincodeHeaderExtension(ccHdrExt *pb.ChaincodeHeaderExtension) {
	p.Element("ChaincodeId")
	p.PrintChaincodeID(ccHdrExt.ChaincodeId)
	p.ElementEnd()
	p.Field("PayloadVisibility", ccHdrExt.PayloadVisibility)
}

func (p *PrinterImpl) PrintChaincodeID(ccID *pb.ChaincodeID) {
	if ccID == nil {
		return
	}
	p.Field("Name", ccID.Name)
	p.Field("Version", ccID.Version)
	p.Field("Path", ccID.Path)
}

func (p *PrinterImpl) PrintChaincodeInfo(ccInfo *pb.ChaincodeInfo) {
	p.Field("Name", ccInfo.Name)
	p.Field("Path", ccInfo.Path)
	p.Field("Version", ccInfo.Version)
	p.Field("Escc", ccInfo.Escc)
	p.Field("Vscc", ccInfo.Vscc)
	p.Field("Input", ccInfo.Input)
}

func (p *PrinterImpl) PrintSignatureHeader(sigHdr *fabriccmn.SignatureHeader) {
	p.Field("Nonce", Base64URLEncode(sigHdr.Nonce))
	p.Field("Creator", Base64URLEncode(sigHdr.Creator))
}

func (p *PrinterImpl) PrintData(headerType fabriccmn.HeaderType, data []byte) {
	if headerType == fabriccmn.HeaderType_CONFIG {
		envelope := &fabriccmn.ConfigEnvelope{}
		if err := proto.Unmarshal(data, envelope); err != nil {
			panic(fmt.Errorf("Bad envelope: %v", err))
		}
		p.Print("Config Envelope:")
		p.PrintConfigEnvelope(envelope)
	} else if headerType == fabriccmn.HeaderType_CONFIG_UPDATE {
		envelope := &fabriccmn.ConfigUpdateEnvelope{}
		if err := proto.Unmarshal(data, envelope); err != nil {
			panic(fmt.Errorf("Bad envelope: %v", err))
		}
		p.Print("Config Update Envelope:")
		p.PrintConfigUpdateEnvelope(envelope)
	} else if headerType == fabriccmn.HeaderType_ENDORSER_TRANSACTION {
		tx, err := fabricUtils.GetTransaction(data)
		if err != nil {
			panic(fmt.Errorf("Bad envelope: %v", err))
		}
		p.Print("Transaction:")
		p.PrintTransaction(tx)
	} else {
		p.Field("Unsupported Envelope", Base64URLEncode(data))
	}
}

func (p *PrinterImpl) PrintConfigEnvelope(envelope *fabriccmn.ConfigEnvelope) {
	p.Element("Config")
	p.PrintConfig(envelope.Config)
	p.ElementEnd()
	p.Element("LastUpdate")
	p.PrintEnvelope(envelope.LastUpdate)
	p.ElementEnd()
}

func (p *PrinterImpl) PrintConfigUpdateEnvelope(envelope *fabriccmn.ConfigUpdateEnvelope) {
	p.Array("Signatures")
	for i, sig := range envelope.Signatures {
		p.Item("Config Signature", i)
		p.PrintConfigSignature(sig)
		p.ItemEnd()
	}
	p.ArrayEnd()

	configUpdate, err := configtx.UnmarshalConfigUpdate(envelope.ConfigUpdate)
	if err != nil {
		panic(err)
	}

	p.Element("ConfigUpdate")
	p.PrintConfigUpdate(configUpdate)
	p.ElementEnd()
}

func (p *PrinterImpl) PrintTransaction(tx *pb.Transaction) {
	p.Array("Actions")
	for i, action := range tx.Actions {
		p.Item("Action", i)
		p.PrintTXAction(action)
		p.ItemEnd()
	}
	p.ArrayEnd()
}

func (p *PrinterImpl) PrintTXAction(action *pb.TransactionAction) {
	p.Element("Header")

	sigHeader, err := fabricUtils.GetSignatureHeader(action.Header)
	if err != nil {
		panic(err)
	}

	p.PrintSignatureHeader(sigHeader)
	p.ElementEnd()

	p.Element("Payload")

	chaPayload, err := fabricUtils.GetChaincodeActionPayload(action.Payload)
	if err != nil {
		panic(err)
	}

	p.PrintChaincodeActionPayload(chaPayload)
	p.ElementEnd()
}

func (p *PrinterImpl) PrintChaincodeActionPayload(chaPayload *pb.ChaincodeActionPayload) {
	p.Element("Action")
	p.PrintAction(chaPayload.Action)
	p.ElementEnd()
}

func (p *PrinterImpl) PrintAction(action *pb.ChaincodeEndorsedAction) {
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

func (p *PrinterImpl) PrintProposalResponsePayload(prp *pb.ProposalResponsePayload) {
	p.Field("ProposalHash", Base64URLEncode(prp.ProposalHash))

	chaincodeAction := &pb.ChaincodeAction{}
	unmarshalOrPanic(prp.Extension, chaincodeAction)
	p.Element("Extension")
	p.PrintChaincodeAction(chaincodeAction)
	p.ElementEnd()
}

func (p *PrinterImpl) PrintChaincodeAction(chaincodeAction *pb.ChaincodeAction) {
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

func (p *PrinterImpl) PrintTxReadWriteSet(txRWSet *rwsetutil.TxRwSet) {
	p.Array("NsRWs")
	for i, nsRWSet := range txRWSet.NsRwSets {
		p.Item("TxRwSet", i)
		p.PrintNsReadWriteSet(nsRWSet)
		p.ItemEnd()
	}
	p.ArrayEnd()
}

func (p *PrinterImpl) PrintNsReadWriteSet(nsRWSet *rwsetutil.NsRwSet) {
	p.Field("NameSpace", nsRWSet.NameSpace)

	p.Element("KvRwSet")
	p.PrintKvRwSet(nsRWSet.KvRwSet)
	p.ElementEnd()

	p.Element("CollHashedRwSets")
	p.PrintCollHashedRwSets(nsRWSet.CollHashedRwSets)
	p.ElementEnd()
}

func (p *PrinterImpl) PrintKvRwSet(kvRWSet *kvrwset.KVRWSet) {
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

func (p *PrinterImpl) PrintCollHashedRwSets(collHashedRwSets []*rwsetutil.CollHashedRwSet) {
	p.Array("CollHashedRwSets")
	for i, w := range collHashedRwSets {
		p.Item("CollHashedRwSet", i)
		p.PrintCollHashedRwSet(w)
		p.ItemEnd()
	}
	p.ArrayEnd()
}

func (p *PrinterImpl) PrintCollHashedRwSet(collHashedRwSet *rwsetutil.CollHashedRwSet) {
	p.Field("CollectionName", collHashedRwSet.CollectionName)
	p.Field("PvtRwSetHash", Base64URLEncode(collHashedRwSet.PvtRwSetHash))

	p.Element("HashedRwSet")
	p.PrintHashedRwSet(collHashedRwSet.HashedRwSet)
	p.ElementEnd()
}

func (p *PrinterImpl) PrintHashedRwSet(hashedRwSet *kvrwset.HashedRWSet) {
	p.PrintHashedReads(hashedRwSet.HashedReads)
	p.PrintHashedWrites(hashedRwSet.HashedWrites)
}

func (p *PrinterImpl) PrintHashedReads(hashedReads []*kvrwset.KVReadHash) {
	p.Array("HashedReads")
	for i, r := range hashedReads {
		p.Item("HashedRead", i)
		p.PrintHashedRead(r)
		p.ItemEnd()
	}
	p.ArrayEnd()
}

func (p *PrinterImpl) PrintHashedWrites(hashedWrites []*kvrwset.KVWriteHash) {
	p.Array("HashedWrites")
	for i, r := range hashedWrites {
		p.Item("HashedWrite", i)
		p.PrintHashedWrite(r)
		p.ItemEnd()
	}
	p.ArrayEnd()
}

func (p *PrinterImpl) PrintHashedRead(hashedRead *kvrwset.KVReadHash) {
	p.Field("KeyHash", Base64URLEncode(hashedRead.KeyHash))
	p.PrintVersion(hashedRead.Version)
}

func (p *PrinterImpl) PrintHashedWrite(hashedWrite *kvrwset.KVWriteHash) {
	p.Field("KeyHash", Base64URLEncode(hashedWrite.KeyHash))
	p.Field("ValueHash", Base64URLEncode(hashedWrite.ValueHash))
	p.Field("IsDelete", hashedWrite.IsDelete)
}

func (p *PrinterImpl) PrintRangeQueryInfo(rqi *kvrwset.RangeQueryInfo) {
	p.Field("TODO", nil)
}

func (p *PrinterImpl) PrintRead(r *kvrwset.KVRead) {
	p.Field("Key", r.Key)
	p.Element("Version")
	p.PrintVersion(r.Version)
	p.ElementEnd()
}

func (p *PrinterImpl) PrintVersion(version *kvrwset.Version) {
	if version == nil {
		return
	}

	p.Field("BlockNum", version.BlockNum)
	p.Field("TxNum", version.TxNum)
}

func (p *PrinterImpl) PrintWrite(w *kvrwset.KVWrite) {
	p.Field("Key", w.Key)
	p.Field("IsDelete", w.IsDelete)
	p.Field("Value", string(w.Value))
}

func (p *PrinterImpl) PrintChaincodeResponse(response *pb.Response) {
	p.Field("Message", response.Message)
	p.Field("Status", response.Status)
	p.Field("Payload", string(response.Payload))
}

func (p *PrinterImpl) PrintChaincodeEventFromBlock(chaincodeEvent *pb.ChaincodeEvent) {
	p.Field("ChaincodeId", chaincodeEvent.ChaincodeId)
	p.Field("EventName", chaincodeEvent.EventName)
	p.Field("TxID", chaincodeEvent.TxId)
	p.Field("Payload", string(chaincodeEvent.Payload))
}

func (p *PrinterImpl) PrintEndorsement(endorsement *pb.Endorsement) {
	p.Field("Endorser", Base64URLEncode(endorsement.Endorser))
	p.Field("Signature", Base64URLEncode(endorsement.Signature))
}

func (p *PrinterImpl) PrintConfig(config *fabriccmn.Config) {
	p.Field("Sequence", config.Sequence)
	p.Element("ChannelGroup")
	p.PrintConfigGroup(config.ChannelGroup)
	p.ElementEnd()
}

func (p *PrinterImpl) PrintConfigGroup(configGroup *fabriccmn.ConfigGroup) {
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

func (p *PrinterImpl) PrintConfigPolicy(policy *fabriccmn.ConfigPolicy) {
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

func (p *PrinterImpl) PrintImplicitMetaPolicy(impMetaPolicy *fabriccmn.ImplicitMetaPolicy) {
	p.Field("Rule", impMetaPolicy.Rule)
	p.Field("Sub-policy", impMetaPolicy.SubPolicy)
}

func (p *PrinterImpl) PrintSignaturePolicyEnvelope(sigPolicyEnv *fabriccmn.SignaturePolicyEnvelope) {
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

func (p *PrinterImpl) PrintMSPPrincipal(principal *msp.MSPPrincipal) {
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
		p.Value(Base64URLEncode(principal.Principal))
	case msp.MSPPrincipal_ORGANIZATION_UNIT:
		// Principal contains the OrganizationUnit
		unit := &msp.OrganizationUnit{}
		unmarshalOrPanic(principal.Principal, unit)

		p.Field("MspIdentifier", unit.MspIdentifier)
		p.Field("OrganizationalUnitIdentifier", unit.OrganizationalUnitIdentifier)
		p.Field("CertifiersIdentifier", Base64URLEncode(unit.CertifiersIdentifier))
	default:
		p.Value("unknown PrincipalClassification")
	}
	p.ElementEnd()
}

func (p *PrinterImpl) PrintSignaturePolicy(sigPolicy *fabriccmn.SignaturePolicy) {
	switch t := sigPolicy.Type.(type) {
	case *fabriccmn.SignaturePolicy_SignedBy:
		p.PrintSignaturePolicySignedBy(t)
		break
	case *fabriccmn.SignaturePolicy_NOutOf_:
		p.PrintSignaturePolicyNOutOf(t.NOutOf)
		break
	}
}

func (p *PrinterImpl) PrintSignaturePolicySignedBy(sigPolicy *fabriccmn.SignaturePolicy_SignedBy) {
	p.Field("SignaturePolicy_SignedBy", sigPolicy.SignedBy)
}

func (p *PrinterImpl) PrintSignaturePolicyNOutOf(sigPolicy *fabriccmn.SignaturePolicy_NOutOf) {
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

func (p *PrinterImpl) PrintConfigValue(name string, value *fabriccmn.ConfigValue) {
	p.Field("Version", value.Version)
	p.Field("ModPolicy", value.ModPolicy)

	switch name {
	case AnchorPeersKey:
		anchorPeers := &pb.AnchorPeers{}
		unmarshalOrPanic(value.Value, anchorPeers)
		p.Element("Anchor Peers:")
		p.PrintAnchorPeers(anchorPeers)
		p.ElementEnd()
		break

	case MSPKey:
		mspConfig := &msp.MSPConfig{}
		unmarshalOrPanic(value.Value, mspConfig)
		p.Element("MSP Config:")
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
		p.Element("Batch Size:")
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
		p.Element("Channel Restrictions:")
		p.Field("MaxCount", chRestrictions.MaxCount)
		p.ElementEnd()
		break

	case ChannelCreationPolicyKey:
		creationPolicy := &fabriccmn.ConfigPolicy{}
		unmarshalOrPanic(value.Value, creationPolicy)
		p.Element("Channel Creation Policy:")
		p.PrintConfigPolicy(creationPolicy)
		p.ElementEnd()
		break

	// case ChainCreationPolicyNamesKey:
	// 	chainCreationPolicyNames := &ab.ChainCreationPolicyNames{}
	// 	unmarshalOrPanic(value.Value, chainCreationPolicyNames)
	// 	p.Print("Chain Creation Policy Names:")
	// 	p.Field("Names", chainCreationPolicyNames.Names)
	// 	break

	case HashingAlgorithmKey:
		hashingAlgorithm := &fabriccmn.HashingAlgorithm{}
		unmarshalOrPanic(value.Value, hashingAlgorithm)
		p.Element("Hashing Algorithm:")
		p.Field("Name", hashingAlgorithm.Name)
		p.ElementEnd()
		break

	case BlockDataHashingStructureKey:
		hashingStructure := &fabriccmn.BlockDataHashingStructure{}
		unmarshalOrPanic(value.Value, hashingStructure)
		p.Element("Block Data Hashing Structure:")
		p.Field("Width", hashingStructure.Width)
		p.ElementEnd()
		break

	case OrdererAddressesKey:
		addresses := &fabriccmn.OrdererAddresses{}
		unmarshalOrPanic(value.Value, addresses)
		p.Element("Orderer Addresses:")
		p.Field("Addresses", addresses.Addresses)
		p.ElementEnd()
		break

	case ConsortiumKey:
		consortium := &fabriccmn.Consortium{}
		unmarshalOrPanic(value.Value, consortium)
		p.Element("Consortium:")
		p.Field("Name", consortium.Name)
		p.ElementEnd()
		break

	default:
		p.Print("!!!!!!! Don't know how to Print config value: %s", name)
		break
	}
}

func (p *PrinterImpl) PrintBatchSize(batchSize *ab.BatchSize) {
	p.Field("MaxMessageCount", batchSize.MaxMessageCount)
	p.Field("AbsoluteMaxBytes", batchSize.AbsoluteMaxBytes)
	p.Field("PreferredMaxBytes", batchSize.PreferredMaxBytes)
}

func (p *PrinterImpl) PrintMSPConfig(mspConfig *msp.MSPConfig) {
	switch fabricMsp.ProviderType(mspConfig.Type) {
	case fabricMsp.FABRIC:
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

func (p *PrinterImpl) PrintFabricMSPConfig(mspConfig *msp.FabricMSPConfig) {
	p.Field("Name", mspConfig.Name)
	p.Array("Admins")
	for i, admCert := range mspConfig.Admins {
		p.ItemValue("Admin Cert", i, Base64URLEncode(admCert))
	}
	p.ArrayEnd()
}

func (p *PrinterImpl) PrintAnchorPeers(anchorPeers *pb.AnchorPeers) {
	p.Array("AnchorPeers")
	for i, anchorPeer := range anchorPeers.AnchorPeers {
		p.Item("Anchor Peer", i)
		p.PrintAnchorPeer(anchorPeer)
		p.ItemEnd()
	}
	p.ArrayEnd()
}

func (p *PrinterImpl) PrintAnchorPeer(anchorPeer *pb.AnchorPeer) {
	p.Field("Host", anchorPeer.Host)
	p.Field("Port", anchorPeer.Port)
}

func (p *PrinterImpl) PrintConfigSignature(sig *fabriccmn.ConfigSignature) {
	p.Field("Signature", Base64URLEncode(sig.Signature))

	sigHeader, err := fabricUtils.GetSignatureHeader(sig.SignatureHeader)
	if err != nil {
		panic(err)
	}

	p.Element("SignatureHeader")
	p.PrintSignatureHeader(sigHeader)
	p.ElementEnd()
}

func (p *PrinterImpl) PrintConfigUpdate(configUpdate *fabriccmn.ConfigUpdate) {
	p.Field("ChannelId", configUpdate.ChannelId)
	p.Element("ReadSet")
	p.PrintConfigGroup(configUpdate.ReadSet)
	p.ElementEnd()
	p.Element("WriteSet")
	p.PrintConfigGroup(configUpdate.WriteSet)
	p.ElementEnd()
}

func (p *PrinterImpl) PrintBlockMetadata(blockMetaData *fabriccmn.BlockMetadata) {
	p.PrintSignaturesMetadata(getMetadataOrPanic(blockMetaData, fabriccmn.BlockMetadataIndex_SIGNATURES))
	p.PrintLastConfigMetadata(getMetadataOrPanic(blockMetaData, fabriccmn.BlockMetadataIndex_LAST_CONFIG))
	p.PrintTransactionsFilterMetadata(ledgerUtil.TxValidationFlags(blockMetaData.Metadata[fabriccmn.BlockMetadataIndex_TRANSACTIONS_FILTER]))
	p.PrintOrdererMetadata(getMetadataOrPanic(blockMetaData, fabriccmn.BlockMetadataIndex_ORDERER))
}

func (p *PrinterImpl) PrintSignaturesMetadata(metadata *fabriccmn.Metadata) {
	p.Array("Signatures")
	for i, metadataSignature := range metadata.Signatures {
		shdr, err := fabricUtils.GetSignatureHeader(metadataSignature.SignatureHeader)
		if err != nil {
			panic(fmt.Errorf("Failed unmarshalling meta data signature header. Error: %v", err))
		}
		p.Item("Metadata Signature", i)
		p.PrintSignatureHeader(shdr)
		p.ItemEnd()
	}
	p.ArrayEnd()
}

func (p *PrinterImpl) PrintLastConfigMetadata(metadata *fabriccmn.Metadata) {
	lastConfig := &fabriccmn.LastConfig{}
	err := proto.Unmarshal(metadata.Value, lastConfig)
	if err != nil {
		panic(fmt.Errorf("Failed unmarshalling meta data last config. Error: %v", err))
	}

	p.Field("Last Config Index", lastConfig.Index)
}

func (p *PrinterImpl) PrintTransactionsFilterMetadata(txFilter ledgerUtil.TxValidationFlags) {
	p.Array("TransactionFilters")
	for i := 0; i < len(txFilter); i++ {
		p.ItemValue("TxValidationFlag", i, txFilter.Flag(i))
	}
	p.ArrayEnd()
}

func (p *PrinterImpl) PrintOrdererMetadata(metadata *fabriccmn.Metadata) {
	p.Field("Orderer Metadata", metadata.Value)
}

func (p *PrinterImpl) PrintChaincodeDeploymentSpec(depSpec *pb.ChaincodeDeploymentSpec) {
	p.Field("EffectiveDate", depSpec.EffectiveDate)
	p.Element("ChaincodeSpec")
	p.PrintChaincodeSpec(depSpec.ChaincodeSpec)
	p.ElementEnd()
}

func (p *PrinterImpl) PrintChaincodeSpec(ccSpec *pb.ChaincodeSpec) {
	p.Element("ChaincodeId")
	p.PrintChaincodeID(ccSpec.ChaincodeId)
	p.ElementEnd()
	p.Field("Timeout", ccSpec.Timeout)
	p.Field("Type", ccSpec.Type)
	p.Element("Input")
	p.Field("Args", ccSpec.Input.Args)
	p.ElementEnd()
}

func (p *PrinterImpl) Print(frmt string, vars ...interface{}) {
	p.Formatter.Print(frmt, vars...)
}

func (p *PrinterImpl) Field(Field string, value interface{}) {
	p.Formatter.Field(Field, value)
}

func (p *PrinterImpl) Element(element string) {
	p.Formatter.Element(element)
}

func (p *PrinterImpl) ElementEnd() {
	p.Formatter.ElementEnd()
}

func (p *PrinterImpl) Array(element string) {
	p.Formatter.Array(element)
}

func (p *PrinterImpl) ArrayEnd() {
	p.Formatter.ArrayEnd()
}

func (p *PrinterImpl) Item(element string, index interface{}) {
	p.Formatter.Item(element, index)
}

func (p *PrinterImpl) ItemEnd() {
	p.Formatter.ItemEnd()
}

func (p *PrinterImpl) ItemValue(element string, index interface{}, value interface{}) {
	p.Formatter.ItemValue(element, index, value)
}

func (p *PrinterImpl) Value(value interface{}) {
	p.Formatter.Value(value)
}

func (p *PrinterImpl) PrintHeader() {
	p.Formatter.PrintHeader()
}

func (p *PrinterImpl) PrintFooter() {
	p.Formatter.PrintFooter()
}

func getMetadataOrPanic(blockMetaData *fabriccmn.BlockMetadata, index fabriccmn.BlockMetadataIndex) *fabriccmn.Metadata {
	metaData := &fabriccmn.Metadata{}
	err := proto.Unmarshal(blockMetaData.Metadata[index], metaData)
	if err != nil {
		panic(fmt.Errorf("Unable to unmarshal meta data at index %d", index))
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
