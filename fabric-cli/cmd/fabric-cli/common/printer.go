/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package common

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	fabricConfig "github.com/hyperledger/fabric/common/config"
	"github.com/hyperledger/fabric/common/configtx"
	"github.com/hyperledger/fabric/core/common/ccprovider"
	"github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/rwsetutil"
	ledgerUtil "github.com/hyperledger/fabric/core/ledger/util"
	fabricMsp "github.com/hyperledger/fabric/msp"
	fabricCommon "github.com/hyperledger/fabric/protos/common"
	"github.com/hyperledger/fabric/protos/ledger/rwset/kvrwset"
	"github.com/hyperledger/fabric/protos/msp"
	ab "github.com/hyperledger/fabric/protos/orderer"
	"github.com/hyperledger/fabric/protos/peer"
	fabricPeer "github.com/hyperledger/fabric/protos/peer"
	fabricUtils "github.com/hyperledger/fabric/protos/utils"
)

const (
	indentSize = 3
)

// Printer is used for printing various data structures
type Printer interface {
	// PrintBlockchainInfo outputs BlockchainInfo
	PrintBlockchainInfo(info *fabricCommon.BlockchainInfo)

	// PrintBlock outputs a Block
	PrintBlock(block *fabricCommon.Block)

	// PrintChannels outputs the array of ChannelInfo
	PrintChannels(channels []*fabricPeer.ChannelInfo)

	// PrintChaincodes outputs the given array of ChaincodeInfo
	PrintChaincodes(chaincodes []*fabricPeer.ChaincodeInfo)

	// PrintProcessedTransaction outputs a ProcessedTransaction
	PrintProcessedTransaction(tx *fabricPeer.ProcessedTransaction)

	// PrintChaincodeData outputs ChaincodeData
	PrintChaincodeData(ccdata *ccprovider.ChaincodeData)
}

type printer struct {
	formatter Formatter
}

// NewPrinter returns a new Printer of the given OutputFormat
func NewPrinter(format OutputFormat) Printer {
	var f Formatter
	switch format {
	case JSON:
		f = &jsonFormatter{}
		break
	case DISPLAY:
		f = &displayFormatter{}
		break
	}
	return &printer{formatter: f}
}

func (p *printer) PrintBlockchainInfo(info *fabricCommon.BlockchainInfo) {
	if p.formatter == nil {
		fmt.Printf("%s\n", info)
		return
	}

	p.PrintHeader()
	p.print("Block Info for chain %s", Config().ChannelID())
	p.field("Height", info.Height)
	p.field("CurrentBlockHash", Base64URLEncode(info.CurrentBlockHash))
	p.field("PreviousBlockHash", Base64URLEncode(info.PreviousBlockHash))
	p.PrintFooter()
}

func (p *printer) PrintBlock(block *fabricCommon.Block) {
	if p.formatter == nil {
		fmt.Printf("%s\n", block)
		return
	}

	p.PrintHeader()
	p.element("Header")
	p.field("Number", block.Header.Number)
	p.field("Hash", Base64URLEncode(block.Header.Hash()))
	p.field("PreviousHash", Base64URLEncode(block.Header.PreviousHash))
	p.field("DataHash", Base64URLEncode(block.Header.DataHash))
	p.elementEnd()

	p.element("Metadata")
	p.printBlockMetadata(block.Metadata)
	p.elementEnd()

	p.element("Data")
	p.array("Data")
	numEnvelopes := len(block.Data.Data)
	for i := 0; i < numEnvelopes; i++ {
		p.item("Envelope", i)
		p.printEnvelope(fabricUtils.ExtractEnvelopeOrPanic(block, i))
		p.itemEnd()
	}
	p.arrayEnd()
	p.elementEnd()
	p.PrintFooter()
}

func (p *printer) PrintChannels(channels []*fabricPeer.ChannelInfo) {
	if p.formatter == nil {
		fmt.Printf("%s\n", channels)
		return
	}

	p.PrintHeader()
	p.array("Channels")
	for _, channel := range channels {
		p.field("ChannelId", channel.ChannelId)
	}
	p.arrayEnd()
	p.PrintFooter()
}

func (p *printer) PrintChaincodes(chaincodes []*fabricPeer.ChaincodeInfo) {
	if p.formatter == nil {
		fmt.Printf("%s\n", chaincodes)
		return
	}

	p.PrintHeader()
	p.array("")
	for _, ccInfo := range chaincodes {
		p.item("ChaincodeInfo", ccInfo.Name)
		p.printChaincodeInfo(ccInfo)
		p.itemEnd()
	}
	p.arrayEnd()
	p.PrintFooter()
}

func (p *printer) PrintProcessedTransaction(tx *fabricPeer.ProcessedTransaction) {
	if p.formatter == nil {
		fmt.Printf("%s\n", tx)
		return
	}

	p.PrintHeader()
	p.print("ValidationCode\n", tx.ValidationCode)
	p.printEnvelope(tx.TransactionEnvelope)
	p.PrintFooter()
}

func (p *printer) PrintChaincodeData(ccData *ccprovider.ChaincodeData) {
	if p.formatter == nil {
		fmt.Printf("%s\n", ccData)
		return
	}

	p.PrintHeader()

	p.field("Id", ccData.Id)
	p.field("Name", ccData.Name)
	p.field("Version", ccData.Version)
	p.field("Escc", ccData.Escc)
	p.field("Vscc", ccData.Vscc)

	cdsData := &ccprovider.CDSData{}
	unmarshalOrPanic(ccData.Data, cdsData)
	p.element("Data")
	p.printCDSData(cdsData)
	p.elementEnd()

	policy := &fabricCommon.SignaturePolicyEnvelope{}
	unmarshalOrPanic(ccData.Policy, policy)
	p.element("Policy")
	p.printSignaturePolicyEnvelope(policy)
	p.elementEnd()

	instPolicy := &fabricCommon.SignaturePolicyEnvelope{}
	unmarshalOrPanic(ccData.InstantiationPolicy, instPolicy)
	p.element("InstantiationPolicy")
	p.printSignaturePolicyEnvelope(instPolicy)
	p.elementEnd()

	p.PrintFooter()
}

func (p *printer) printCDSData(cdsData *ccprovider.CDSData) {
	p.field("CodeHash", Base64URLEncode(cdsData.CodeHash))
	p.field("MetaDataHash", Base64URLEncode(cdsData.MetaDataHash))
}

func (p *printer) printEnvelope(envelope *fabricCommon.Envelope) {
	p.field("Signature", Base64URLEncode(envelope.Signature))

	payload := fabricUtils.ExtractPayloadOrPanic(envelope)
	p.element("Payload")
	p.printPayload(payload)
	p.elementEnd()
}

func (p *printer) printPayload(payload *fabricCommon.Payload) {
	p.element("Header")

	chdr, err := fabricUtils.UnmarshalChannelHeader(payload.Header.ChannelHeader)
	if err != nil {
		panic(err)
	}

	p.element("ChannelHeader")
	p.printChannelHeader(chdr)
	p.elementEnd()

	sigHeader, err := fabricUtils.GetSignatureHeader(payload.Header.SignatureHeader)
	if err != nil {
		panic(err)
	}

	p.element("SignatureHeader")
	p.printSignatureHeader(sigHeader)
	p.elementEnd()

	p.elementEnd() // Header

	p.element("Data")
	p.field("Type", fabricCommon.HeaderType(chdr.Type))
	p.printData(fabricCommon.HeaderType(chdr.Type), payload.Data)
	p.elementEnd()
}

func (p *printer) printChannelHeader(chdr *fabricCommon.ChannelHeader) {
	p.field("Type", chdr.Type)
	p.field("ChannelId", chdr.ChannelId)
	p.field("Epoch", chdr.Epoch)

	ccHdrExt := &peer.ChaincodeHeaderExtension{}
	unmarshalOrPanic(chdr.Extension, ccHdrExt)
	p.element("Extension")
	p.printChaincodeHeaderExtension(ccHdrExt)
	p.elementEnd()

	p.field("Timestamp", chdr.Timestamp)
	p.field("TxId", chdr.TxId)
	p.field("Version", chdr.Version)
}

func (p *printer) printChaincodeHeaderExtension(ccHdrExt *peer.ChaincodeHeaderExtension) {
	p.element("ChaincodeId")
	p.printChaincodeID(ccHdrExt.ChaincodeId)
	p.elementEnd()
	p.field("PayloadVisibility", ccHdrExt.PayloadVisibility)
}

func (p *printer) printChaincodeID(ccID *peer.ChaincodeID) {
	if ccID == nil {
		return
	}
	p.field("Name", ccID.Name)
	p.field("Version", ccID.Version)
	p.field("Path", ccID.Path)
}

func (p *printer) printChaincodeInfo(ccInfo *peer.ChaincodeInfo) {
	p.field("Name", ccInfo.Name)
	p.field("Path", ccInfo.Path)
	p.field("Version", ccInfo.Version)
	p.field("Escc", ccInfo.Escc)
	p.field("Vscc", ccInfo.Vscc)
	p.field("Input", ccInfo.Input)
}

func (p *printer) printSignatureHeader(sigHdr *fabricCommon.SignatureHeader) {
	p.field("Nonce", Base64URLEncode(sigHdr.Nonce))
	p.field("Creator", Base64URLEncode(sigHdr.Creator))
}

func (p *printer) printData(headerType fabricCommon.HeaderType, data []byte) {
	if headerType == fabricCommon.HeaderType_CONFIG {
		envelope := &fabricCommon.ConfigEnvelope{}
		if err := proto.Unmarshal(data, envelope); err != nil {
			panic(fmt.Errorf("Bad envelope: %v", err))
		}
		p.print("Config Envelope:")
		p.printConfigEnvelope(envelope)
	} else if headerType == fabricCommon.HeaderType_CONFIG_UPDATE {
		envelope := &fabricCommon.ConfigUpdateEnvelope{}
		if err := proto.Unmarshal(data, envelope); err != nil {
			panic(fmt.Errorf("Bad envelope: %v", err))
		}
		p.print("Config Update Envelope:")
		p.printConfigUpdateEnvelope(envelope)
	} else if headerType == fabricCommon.HeaderType_ENDORSER_TRANSACTION {
		tx, err := fabricUtils.GetTransaction(data)
		if err != nil {
			panic(fmt.Errorf("Bad envelope: %v", err))
		}
		p.print("Transaction:")
		p.printTransaction(tx)
	} else {
		p.field("Unsupported Envelope", Base64URLEncode(data))
	}
}

func (p *printer) printConfigEnvelope(envelope *fabricCommon.ConfigEnvelope) {
	p.element("Config")
	p.printConfig(envelope.Config)
	p.elementEnd()
	p.element("LastUpdate")
	p.printEnvelope(envelope.LastUpdate)
	p.elementEnd()
}

func (p *printer) printConfigUpdateEnvelope(envelope *fabricCommon.ConfigUpdateEnvelope) {
	p.array("Signatures")
	for i, sig := range envelope.Signatures {
		p.item("Config Signature", i)
		p.printConfigSignature(sig)
		p.itemEnd()
	}
	p.arrayEnd()

	configUpdate, err := configtx.UnmarshalConfigUpdate(envelope.ConfigUpdate)
	if err != nil {
		panic(err)
	}

	p.element("ConfigUpdate")
	p.printConfigUpdate(configUpdate)
	p.elementEnd()
}

func (p *printer) printTransaction(tx *peer.Transaction) {
	p.array("Actions")
	for i, action := range tx.Actions {
		p.item("Action", i)
		p.printTXAction(action)
		p.itemEnd()
	}
	p.arrayEnd()
}

func (p *printer) printTXAction(action *peer.TransactionAction) {
	p.element("Header")

	sigHeader, err := fabricUtils.GetSignatureHeader(action.Header)
	if err != nil {
		panic(err)
	}

	p.printSignatureHeader(sigHeader)
	p.elementEnd()

	p.element("Payload")

	chaPayload, err := fabricUtils.GetChaincodeActionPayload(action.Payload)
	if err != nil {
		panic(err)
	}

	p.printChaincodeActionPayload(chaPayload)
	p.elementEnd()
}

func (p *printer) printChaincodeActionPayload(chaPayload *peer.ChaincodeActionPayload) {
	p.element("Action")
	p.printAction(chaPayload.Action)
	p.elementEnd()
}

func (p *printer) printAction(action *peer.ChaincodeEndorsedAction) {
	p.array("Endorsements")
	for i, endorsement := range action.Endorsements {
		p.item("Endorsement", i)
		p.printEndorsement(endorsement)
		p.itemEnd()
	}
	p.arrayEnd()

	prp := &peer.ProposalResponsePayload{}
	unmarshalOrPanic(action.ProposalResponsePayload, prp)

	p.element("ProposalResponsePayload")
	p.printProposalResponsePayload(prp)
	p.elementEnd()
}

func (p *printer) printProposalResponsePayload(prp *peer.ProposalResponsePayload) {
	p.field("ProposalHash", Base64URLEncode(prp.ProposalHash))

	chaincodeAction := &peer.ChaincodeAction{}
	unmarshalOrPanic(prp.Extension, chaincodeAction)
	p.element("Extension")
	p.printChaincodeAction(chaincodeAction)
	p.elementEnd()
}

func (p *printer) printChaincodeAction(chaincodeAction *peer.ChaincodeAction) {
	p.element("Response")
	p.printChaincodeResponse(chaincodeAction.Response)
	p.elementEnd()

	p.element("Results")
	if len(chaincodeAction.Results) == 0 {
		p.elementEnd()
		return
	}

	txRWSet := &rwsetutil.TxRwSet{}
	if err := txRWSet.FromProtoBytes(chaincodeAction.Results); err != nil {
		panic(err)
	}

	p.printTxReadWriteSet(txRWSet)
	p.elementEnd()

	p.element("Events")
	if len(chaincodeAction.Events) > 0 {
		chaincodeEvent := &peer.ChaincodeEvent{}
		unmarshalOrPanic(chaincodeAction.Events, chaincodeEvent)
		p.printChaincodeEvent(chaincodeEvent)
	}
	p.elementEnd()
}

func (p *printer) printTxReadWriteSet(txRWSet *rwsetutil.TxRwSet) {
	p.array("NsRWs")
	for i, nsRWSet := range txRWSet.NsRwSets {
		p.item("TxRwSet", i)
		p.printNsReadWriteSet(nsRWSet)
		p.itemEnd()
	}
	p.arrayEnd()
}

func (p *printer) printNsReadWriteSet(nsRWSet *rwsetutil.NsRwSet) {
	p.field("NameSpace", nsRWSet.NameSpace)
	p.array("RangeQueriesInfo")
	for i, rqi := range nsRWSet.KvRwSet.RangeQueriesInfo {
		p.item("RangeQueryInfo", i)
		p.printRangeQueryInfo(rqi)
		p.itemEnd()
	}
	p.arrayEnd()

	p.array("Reads")
	for i, r := range nsRWSet.KvRwSet.Reads {
		p.item("Read", i)
		p.printRead(r)
		p.itemEnd()
	}
	p.arrayEnd()

	p.array("Writes")
	for i, w := range nsRWSet.KvRwSet.Writes {
		p.item("Write", i)
		p.printWrite(w)
		p.itemEnd()
	}
	p.arrayEnd()
}

func (p *printer) printRangeQueryInfo(rqi *kvrwset.RangeQueryInfo) {
	p.field("TODO", nil)
}

func (p *printer) printRead(r *kvrwset.KVRead) {
	p.field("Key", r.Key)
	p.element("Version")
	p.printVersion(r.Version)
	p.elementEnd()
}

func (p *printer) printVersion(version *kvrwset.Version) {
	if version == nil {
		return
	}

	p.field("BlockNum", version.BlockNum)
	p.field("TxNum", version.TxNum)
}

func (p *printer) printWrite(w *kvrwset.KVWrite) {
	p.field("Key", w.Key)
	p.field("IsDelete", w.IsDelete)
	p.field("Value", w.Value)
}

func (p *printer) printChaincodeResponse(response *peer.Response) {
	p.field("Message", response.Message)
	p.field("Status", response.Status)
	p.field("Payload", Base64URLEncode(response.Payload))
}

func (p *printer) printChaincodeEvent(chaincodeEvent *peer.ChaincodeEvent) {
	p.field("ChaincodeId", chaincodeEvent.ChaincodeId)
	p.field("EventName", chaincodeEvent.EventName)
	p.field("TxID", chaincodeEvent.TxId)
	p.field("Payload", Base64URLEncode(chaincodeEvent.Payload))
}

func (p *printer) printEndorsement(endorsement *peer.Endorsement) {
	p.field("Endorser", Base64URLEncode(endorsement.Endorser))
	p.field("Signature", Base64URLEncode(endorsement.Signature))
}

func (p *printer) printConfig(config *fabricCommon.Config) {
	p.field("Sequence", config.Sequence)
	p.element("ChannelGroup")
	p.printConfigGroup(config.ChannelGroup)
	p.elementEnd()
}

func (p *printer) printConfigGroup(configGroup *fabricCommon.ConfigGroup) {
	if configGroup == nil {
		return
	}

	p.field("Version", configGroup.Version)
	p.field("ModPolicy", configGroup.ModPolicy)
	p.array("Groups")
	for key, grp := range configGroup.Groups {
		p.item("Group", key)
		p.printConfigGroup(grp)
		p.itemEnd()
	}
	p.arrayEnd()

	p.array("Values")
	for key, val := range configGroup.Values {
		p.item("Value", key)
		p.printConfigValue(key, val)
		p.itemEnd()
	}
	p.arrayEnd()

	p.array("Policies")
	for key, policy := range configGroup.Policies {
		p.item("Policy", key)
		p.printConfigPolicy(policy)
		p.itemEnd()
	}
	p.arrayEnd()
}

func (p *printer) printConfigPolicy(policy *fabricCommon.ConfigPolicy) {
	p.field("ModPolicy", policy.ModPolicy)
	p.field("Version", policy.Version)

	policyType := fabricCommon.Policy_PolicyType(policy.Policy.Type)
	switch policyType {
	case fabricCommon.Policy_SIGNATURE:
		sigPolicyEnv := &fabricCommon.SignaturePolicyEnvelope{}
		unmarshalOrPanic(policy.Policy.Value, sigPolicyEnv)
		p.print("Signature Policy:")
		p.printSignaturePolicyEnvelope(sigPolicyEnv)
		break

	case fabricCommon.Policy_MSP:
		p.print("Policy_MSP: TODO")
		break

	case fabricCommon.Policy_IMPLICIT_META:
		impMetaPolicy := &fabricCommon.ImplicitMetaPolicy{}
		unmarshalOrPanic(policy.Policy.Value, impMetaPolicy)
		p.print("Implicit Meta Policy:")
		p.printImplicitMetaPolicy(impMetaPolicy)
		break

	default:
		break
	}

}

func (p *printer) printImplicitMetaPolicy(impMetaPolicy *fabricCommon.ImplicitMetaPolicy) {
	p.field("Rule", impMetaPolicy.Rule)
	p.field("Sub-policy", impMetaPolicy.SubPolicy)
}

func (p *printer) printSignaturePolicyEnvelope(sigPolicyEnv *fabricCommon.SignaturePolicyEnvelope) {
	p.element("Rule")
	p.printSignaturePolicy(sigPolicyEnv.Rule)
	p.elementEnd()

	p.array("Identities")
	for i, identity := range sigPolicyEnv.Identities {
		p.item("Identity", i)
		p.printMSPPrincipal(identity)
		p.itemEnd()
	}
	p.arrayEnd()
}

func (p *printer) printMSPPrincipal(principal *msp.MSPPrincipal) {
	p.field("PrincipalClassification", principal.PrincipalClassification)
	p.element("Principal")
	switch principal.PrincipalClassification {
	case msp.MSPPrincipal_ROLE:
		// Principal contains the msp role
		mspRole := &msp.MSPRole{}
		unmarshalOrPanic(principal.Principal, mspRole)
		p.field("Role", mspRole.Role)
		p.field("MspIdentifier", mspRole.MspIdentifier)
	case msp.MSPPrincipal_IDENTITY:
		p.value(Base64URLEncode(principal.Principal))
	case msp.MSPPrincipal_ORGANIZATION_UNIT:
		// Principal contains the OrganizationUnit
		unit := &msp.OrganizationUnit{}
		unmarshalOrPanic(principal.Principal, unit)

		p.field("MspIdentifier", unit.MspIdentifier)
		p.field("OrganizationalUnitIdentifier", unit.OrganizationalUnitIdentifier)
		p.field("CertifiersIdentifier", Base64URLEncode(unit.CertifiersIdentifier))
	default:
		p.value("unknown PrincipalClassification")
	}
	p.elementEnd()
}

func (p *printer) printSignaturePolicy(sigPolicy *fabricCommon.SignaturePolicy) {
	switch t := sigPolicy.Type.(type) {
	case *fabricCommon.SignaturePolicy_SignedBy:
		p.printSignaturePolicySignedBy(t)
		break
	case *fabricCommon.SignaturePolicy_NOutOf_:
		p.printSignaturePolicyNOutOf(t.NOutOf)
		break
	}
}

func (p *printer) printSignaturePolicySignedBy(sigPolicy *fabricCommon.SignaturePolicy_SignedBy) {
	p.field("SignaturePolicy_SignedBy", sigPolicy.SignedBy)
}

func (p *printer) printSignaturePolicyNOutOf(sigPolicy *fabricCommon.SignaturePolicy_NOutOf) {
	p.print("SignaturePolicy_NOutOf")
	p.field("N", sigPolicy.N)
	p.array("Rules")
	for i, policy := range sigPolicy.Rules {
		p.item("Rule", i)
		p.printSignaturePolicy(policy)
		p.itemEnd()
	}
	p.arrayEnd()
}

func (p *printer) printConfigValue(name string, value *fabricCommon.ConfigValue) {
	p.field("Version", value.Version)
	p.field("ModPolicy", value.ModPolicy)

	switch name {
	case fabricConfig.AnchorPeersKey:
		anchorPeers := &peer.AnchorPeers{}
		unmarshalOrPanic(value.Value, anchorPeers)
		p.element("Anchor Peers:")
		p.printAnchorPeers(anchorPeers)
		p.elementEnd()
		break

	case fabricConfig.MSPKey:
		mspConfig := &msp.MSPConfig{}
		unmarshalOrPanic(value.Value, mspConfig)
		p.element("MSP Config:")
		p.printMSPConfig(mspConfig)
		p.elementEnd()
		break

	case fabricConfig.ConsensusTypeKey:
		consensusType := &ab.ConsensusType{}
		unmarshalOrPanic(value.Value, consensusType)
		p.field("ConsensusType", consensusType.Type)
		break

	case fabricConfig.BatchSizeKey:
		batchSize := &ab.BatchSize{}
		unmarshalOrPanic(value.Value, batchSize)
		p.element("Batch Size:")
		p.printBatchSize(batchSize)
		p.elementEnd()
		break

	case fabricConfig.BatchTimeoutKey:
		batchTimeout := &ab.BatchTimeout{}
		unmarshalOrPanic(value.Value, batchTimeout)
		p.field("Timeout", batchTimeout.Timeout)
		break

	case fabricConfig.ChannelRestrictionsKey:
		chRestrictions := &ab.ChannelRestrictions{}
		unmarshalOrPanic(value.Value, chRestrictions)
		p.element("Channel Restrictions:")
		p.field("MaxCount", chRestrictions.MaxCount)
		p.elementEnd()
		break

	case fabricConfig.ChannelCreationPolicyKey:
		creationPolicy := &fabricCommon.ConfigPolicy{}
		unmarshalOrPanic(value.Value, creationPolicy)
		p.element("Channel Creation Policy:")
		p.printConfigPolicy(creationPolicy)
		p.elementEnd()
		break

	// case fabricConfig.ChainCreationPolicyNamesKey:
	// 	chainCreationPolicyNames := &ab.ChainCreationPolicyNames{}
	// 	unmarshalOrPanic(value.Value, chainCreationPolicyNames)
	// 	p.print("Chain Creation Policy Names:")
	// 	p.field("Names", chainCreationPolicyNames.Names)
	// 	break

	case fabricConfig.HashingAlgorithmKey:
		hashingAlgorithm := &fabricCommon.HashingAlgorithm{}
		unmarshalOrPanic(value.Value, hashingAlgorithm)
		p.element("Hashing Algorithm:")
		p.field("Name", hashingAlgorithm.Name)
		p.elementEnd()
		break

	case fabricConfig.BlockDataHashingStructureKey:
		hashingStructure := &fabricCommon.BlockDataHashingStructure{}
		unmarshalOrPanic(value.Value, hashingStructure)
		p.element("Block Data Hashing Structure:")
		p.field("Width", hashingStructure.Width)
		p.elementEnd()
		break

	case fabricConfig.OrdererAddressesKey:
		addresses := &fabricCommon.OrdererAddresses{}
		unmarshalOrPanic(value.Value, addresses)
		p.element("Orderer Addresses:")
		p.field("Addresses", addresses.Addresses)
		p.elementEnd()
		break

	case fabricConfig.ConsortiumKey:
		consortium := &fabricCommon.Consortium{}
		unmarshalOrPanic(value.Value, consortium)
		p.element("Consortium:")
		p.field("Name", consortium.Name)
		p.elementEnd()
		break

	default:
		p.print("!!!!!!! Don't know how to print config value: %s", name)
		break
	}
}

func (p *printer) printBatchSize(batchSize *ab.BatchSize) {
	p.field("MaxMessageCount", batchSize.MaxMessageCount)
	p.field("AbsoluteMaxBytes", batchSize.AbsoluteMaxBytes)
	p.field("PreferredMaxBytes", batchSize.PreferredMaxBytes)
}

func (p *printer) printMSPConfig(mspConfig *msp.MSPConfig) {
	switch fabricMsp.ProviderType(mspConfig.Type) {
	case fabricMsp.FABRIC:
		p.print("Type: FABRIC")

		config := &msp.FabricMSPConfig{}
		unmarshalOrPanic(mspConfig.Config, config)

		p.print("Config:")
		p.printFabricMSPConfig(config)
		break

	default:
		p.print("Type: OTHER")
		break
	}
}

func (p *printer) printFabricMSPConfig(mspConfig *msp.FabricMSPConfig) {
	p.field("Name", mspConfig.Name)
	p.array("Admins")
	for i, admCert := range mspConfig.Admins {
		p.itemValue("Admin Cert", i, Base64URLEncode(admCert))
	}
	p.arrayEnd()
}

func (p *printer) printAnchorPeers(anchorPeers *peer.AnchorPeers) {
	p.array("AnchorPeers")
	for i, anchorPeer := range anchorPeers.AnchorPeers {
		p.item("Anchor Peer", i)
		p.printAnchorPeer(anchorPeer)
		p.itemEnd()
	}
	p.arrayEnd()
}

func (p *printer) printAnchorPeer(anchorPeer *peer.AnchorPeer) {
	p.field("Host", anchorPeer.Host)
	p.field("Port", anchorPeer.Port)
}

func (p *printer) printConfigSignature(sig *fabricCommon.ConfigSignature) {
	p.field("Signature", Base64URLEncode(sig.Signature))

	sigHeader, err := fabricUtils.GetSignatureHeader(sig.SignatureHeader)
	if err != nil {
		panic(err)
	}

	p.element("SignatureHeader")
	p.printSignatureHeader(sigHeader)
	p.elementEnd()
}

func (p *printer) printConfigUpdate(configUpdate *fabricCommon.ConfigUpdate) {
	p.field("ChannelId", configUpdate.ChannelId)
	p.element("ReadSet")
	p.printConfigGroup(configUpdate.ReadSet)
	p.elementEnd()
	p.element("WriteSet")
	p.printConfigGroup(configUpdate.WriteSet)
	p.elementEnd()
}

func (p *printer) printBlockMetadata(blockMetaData *fabricCommon.BlockMetadata) {
	p.printSignaturesMetadata(getMetadataOrPanic(blockMetaData, fabricCommon.BlockMetadataIndex_SIGNATURES))
	p.printLastConfigMetadata(getMetadataOrPanic(blockMetaData, fabricCommon.BlockMetadataIndex_LAST_CONFIG))
	p.printTransactionsFilterMetadata(ledgerUtil.TxValidationFlags(blockMetaData.Metadata[fabricCommon.BlockMetadataIndex_TRANSACTIONS_FILTER]))
	p.printOrdererMetadata(getMetadataOrPanic(blockMetaData, fabricCommon.BlockMetadataIndex_ORDERER))
}

func (p *printer) printSignaturesMetadata(metadata *fabricCommon.Metadata) {
	p.array("Signatures")
	for i, metadataSignature := range metadata.Signatures {
		shdr, err := fabricUtils.GetSignatureHeader(metadataSignature.SignatureHeader)
		if err != nil {
			panic(fmt.Errorf("Failed unmarshalling meta data signature header. Error: %v", err))
		}
		p.item("Metadata Signature", i)
		p.printSignatureHeader(shdr)
		p.itemEnd()
	}
	p.arrayEnd()
}

func (p *printer) printLastConfigMetadata(metadata *fabricCommon.Metadata) {
	lastConfig := &fabricCommon.LastConfig{}
	err := proto.Unmarshal(metadata.Value, lastConfig)
	if err != nil {
		panic(fmt.Errorf("Failed unmarshalling meta data last config. Error: %v", err))
	}

	p.field("Last Config Index", lastConfig.Index)
}

func (p *printer) printTransactionsFilterMetadata(txFilter ledgerUtil.TxValidationFlags) {
	p.array("TransactionFilters")
	for i := 0; i < len(txFilter); i++ {
		p.itemValue("TxValidationFlag", i, txFilter.Flag(i))
	}
	p.arrayEnd()
}

func (p *printer) printOrdererMetadata(metadata *fabricCommon.Metadata) {
	p.field("Orderer Metadata", metadata.Value)
}

func (p *printer) printChaincodeDeploymentSpec(depSpec *fabricPeer.ChaincodeDeploymentSpec) {
	p.field("EffectiveDate", depSpec.EffectiveDate)
	p.element("ChaincodeSpec")
	p.printChaincodeSpec(depSpec.ChaincodeSpec)
	p.elementEnd()
}

func (p *printer) printChaincodeSpec(ccSpec *fabricPeer.ChaincodeSpec) {
	p.element("ChaincodeId")
	p.printChaincodeID(ccSpec.ChaincodeId)
	p.elementEnd()
	p.field("Timeout", ccSpec.Timeout)
	p.field("Type", ccSpec.Type)
	p.element("Input")
	p.field("Args", ccSpec.Input.Args)
	p.elementEnd()
}

func (p *printer) print(frmt string, vars ...interface{}) {
	p.formatter.Print(frmt, vars...)
}

func (p *printer) field(field string, value interface{}) {
	p.formatter.Field(field, value)
}

func (p *printer) element(element string) {
	p.formatter.Element(element)
}

func (p *printer) elementEnd() {
	p.formatter.ElementEnd()
}

func (p *printer) array(element string) {
	p.formatter.Array(element)
}

func (p *printer) arrayEnd() {
	p.formatter.ArrayEnd()
}

func (p *printer) item(element string, index interface{}) {
	p.formatter.Item(element, index)
}

func (p *printer) itemEnd() {
	p.formatter.ItemEnd()
}

func (p *printer) itemValue(element string, index interface{}, value interface{}) {
	p.formatter.ItemValue(element, index, value)
}

func (p *printer) value(value interface{}) {
	p.formatter.Value(value)
}

func (p *printer) PrintHeader() {
	p.formatter.PrintHeader()
}

func (p *printer) PrintFooter() {
	p.formatter.PrintFooter()
}

func getMetadataOrPanic(blockMetaData *fabricCommon.BlockMetadata, index fabricCommon.BlockMetadataIndex) *fabricCommon.Metadata {
	metaData := &fabricCommon.Metadata{}
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
