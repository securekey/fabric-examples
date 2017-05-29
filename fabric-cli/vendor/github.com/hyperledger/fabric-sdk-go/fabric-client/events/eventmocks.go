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

package events

import (
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	consumer "github.com/hyperledger/fabric-sdk-go/fabric-client/events/consumer"
	"github.com/hyperledger/fabric-sdk-go/fabric-client/util"
	"github.com/hyperledger/fabric/bccsp"
	"github.com/hyperledger/fabric/bccsp/factory"
	ledger_util "github.com/hyperledger/fabric/core/ledger/util"
	fcConsumer "github.com/hyperledger/fabric/events/consumer"
	"github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric/protos/peer"
)

type mockEventClientMockEventRegistration struct {
	Action string
	ies    []*pb.Interest
}

type mockEventClient struct {
	PeerAddress string
	RegTimeout  time.Duration
	Adapter     fcConsumer.EventAdapter

	Started       int
	Stopped       int
	Registrations []mockEventClientMockEventRegistration

	events chan pb.Event
}

type mockEventClientFactory struct {
	clients []*mockEventClient
}

func (mecf *mockEventClientFactory) newEventsClient(peerAddress string, certificate string, serverHostOverride string, regTimeout time.Duration, adapter fcConsumer.EventAdapter) (consumer.EventsClient, error) {
	client := &mockEventClient{
		PeerAddress: peerAddress,
		RegTimeout:  regTimeout,
		Adapter:     adapter,
		events:      make(chan pb.Event),
	}
	mecf.clients = append(mecf.clients, client)
	return client, nil
}

// MockEvent mocks an event
func (mec *mockEventClient) MockEvent(msg *pb.Event) (bool, error) {
	if mec.Started > mec.Stopped {
		return mec.Adapter.Recv(msg)
	}

	mec.events <- *msg
	return true, nil
}

// RegisterAsync does not register anything anywhere but acts like all is well
func (mec *mockEventClient) RegisterAsync(ies []*pb.Interest) error {
	mec.Registrations = append(mec.Registrations, mockEventClientMockEventRegistration{
		Action: "register",
		ies:    ies,
	})
	return nil
}

// UnregisterAsync does not unregister anything anywhere but acts like all is well
func (mec *mockEventClient) UnregisterAsync(ies []*pb.Interest) error {
	mec.Registrations = append(mec.Registrations, mockEventClientMockEventRegistration{
		Action: "register",
		ies:    ies,
	})
	return nil
}

// Recv will return mock events sent to the event channel. Warning! This might block indefinitely
func (mec *mockEventClient) Recv() (*pb.Event, error) {
	event := <-mec.events
	return &event, nil
}

// Start does not start anything
func (mec *mockEventClient) Start() error {
	mec.Started++
	return nil
}

// Stop does not stop anything
func (mec *mockEventClient) Stop() error {
	mec.Stopped++
	return nil
}

func createMockedEventHub(t *testing.T) (*eventHub, *mockEventClientFactory) {
	eventHub, ok := NewEventHub().(*eventHub)
	if !ok {
		t.Fatalf("Could not create eventHub")
		return nil, nil
	}

	var clientFactory mockEventClientFactory
	eventHub.eventsClientFactory = &clientFactory

	eventHub.SetPeerAddr("mock://mock", "", "")

	err := eventHub.Connect()
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
		return nil, nil
	}

	return eventHub, &clientFactory
}

// MockTxEventBuilder builds a mock TX event block
type MockTxEventBuilder struct {
	ChannelID string
	TxID      string
}

// MockCCEventBuilder builds a mock chaincode event
type MockCCEventBuilder struct {
	CCID      string
	EventName string
	Payload   []byte
}

// MockCCBlockEventBuilder builds a mock CC event block
type MockCCBlockEventBuilder struct {
	CCID      string
	EventName string
	ChannelID string
	TxID      string
	Payload   []byte
}

// Build builds a mock TX event block
func (b *MockTxEventBuilder) Build() *pb.Event_Block {
	return &pb.Event_Block{
		Block: &common.Block{
			Header:   &common.BlockHeader{},
			Metadata: b.buildBlockMetadata(),
			Data: &common.BlockData{
				Data: [][]byte{util.MarshalOrPanic(b.buildEnvelope())},
			},
		},
	}
}

func (b *MockTxEventBuilder) buildBlockMetadata() *common.BlockMetadata {
	return &common.BlockMetadata{
		Metadata: [][]byte{
			[]byte{},
			[]byte{},
			b.buildTransactionsFilterMetaDataBytes(),
			[]byte{},
		},
	}
}

func (b *MockTxEventBuilder) buildTransactionsFilterMetaDataBytes() []byte {
	return []byte(ledger_util.TxValidationFlags{uint8(pb.TxValidationCode_VALID)})
}

// Build builds a mock chaincode event
func (b *MockCCEventBuilder) Build() *pb.Event_ChaincodeEvent {
	return &pb.Event_ChaincodeEvent{
		ChaincodeEvent: &pb.ChaincodeEvent{
			ChaincodeId: b.CCID,
			EventName:   b.EventName,
			Payload:     b.Payload,
		},
	}
}

func (b *MockTxEventBuilder) buildEnvelope() *common.Envelope {
	return &common.Envelope{
		Payload: util.MarshalOrPanic(b.buildPayload()),
	}
}

func (b *MockTxEventBuilder) buildPayload() *common.Payload {
	return &common.Payload{
		Header: &common.Header{
			ChannelHeader: util.MarshalOrPanic(b.buildChannelHeader()),
		},
	}
}

func (b *MockTxEventBuilder) buildChannelHeader() *common.ChannelHeader {
	return &common.ChannelHeader{
		TxId:      b.TxID,
		ChannelId: b.ChannelID,
	}
}

// Build builds a mock chaincode event block
func (b *MockCCBlockEventBuilder) Build() *pb.Event_Block {
	return &pb.Event_Block{
		Block: &common.Block{
			Header:   &common.BlockHeader{},
			Metadata: b.buildBlockMetadata(),
			Data: &common.BlockData{
				Data: [][]byte{util.MarshalOrPanic(b.buildEnvelope())},
			},
		},
	}
}

func (b *MockCCBlockEventBuilder) buildBlockMetadata() *common.BlockMetadata {
	return &common.BlockMetadata{
		Metadata: [][]byte{
			[]byte{},
			[]byte{},
			b.buildTransactionsFilterMetaDataBytes(),
			[]byte{},
		},
	}
}

func (b *MockCCBlockEventBuilder) buildEnvelope() *common.Envelope {
	return &common.Envelope{
		Payload: util.MarshalOrPanic(b.buildPayload()),
	}
}

func (b *MockCCBlockEventBuilder) buildTransactionsFilterMetaDataBytes() []byte {
	return []byte(ledger_util.TxValidationFlags{uint8(pb.TxValidationCode_VALID)})
}

func (b *MockCCBlockEventBuilder) buildPayload() *common.Payload {
	fmt.Printf("MockCCBlockEventBuilder.buildPayload\n")
	return &common.Payload{
		Header: &common.Header{
			ChannelHeader: util.MarshalOrPanic(b.buildChannelHeader()),
		},
		Data: util.MarshalOrPanic(b.buildTransaction()),
	}
}

func (b *MockCCBlockEventBuilder) buildChannelHeader() *common.ChannelHeader {
	return &common.ChannelHeader{
		Type:      int32(common.HeaderType_ENDORSER_TRANSACTION),
		TxId:      b.TxID,
		ChannelId: b.ChannelID,
	}
}

func (b *MockCCBlockEventBuilder) buildTransaction() *pb.Transaction {
	return &pb.Transaction{
		Actions: []*pb.TransactionAction{b.buildTransactionAction()},
	}
}

func (b *MockCCBlockEventBuilder) buildTransactionAction() *pb.TransactionAction {
	return &pb.TransactionAction{
		Header:  []byte{},
		Payload: util.MarshalOrPanic(b.buildChaincodeActionPayload()),
	}
}

func (b *MockCCBlockEventBuilder) buildChaincodeActionPayload() *pb.ChaincodeActionPayload {
	return &pb.ChaincodeActionPayload{
		Action: b.buildChaincodeEndorsedAction(),
		ChaincodeProposalPayload: []byte{},
	}
}

func (b *MockCCBlockEventBuilder) buildChaincodeEndorsedAction() *pb.ChaincodeEndorsedAction {
	return &pb.ChaincodeEndorsedAction{
		ProposalResponsePayload: util.MarshalOrPanic(b.buildProposalResponsePayload()),
		Endorsements:            []*pb.Endorsement{},
	}
}

func (b *MockCCBlockEventBuilder) buildProposalResponsePayload() *pb.ProposalResponsePayload {
	return &pb.ProposalResponsePayload{
		ProposalHash: []byte("somehash"),
		Extension:    util.MarshalOrPanic(b.buildChaincodeAction()),
	}
}

func (b *MockCCBlockEventBuilder) buildChaincodeAction() *pb.ChaincodeAction {
	return &pb.ChaincodeAction{
		Events: util.MarshalOrPanic(b.buildChaincodeEvent()),
	}
}

func (b *MockCCBlockEventBuilder) buildChaincodeEvent() *pb.ChaincodeEvent {
	return &pb.ChaincodeEvent{
		ChaincodeId: b.CCID,
		EventName:   b.EventName,
		TxId:        b.TxID,
		Payload:     b.Payload,
	}
}

func generateTxID() string {
	nonce, err := util.GenerateRandomNonce()
	if err != nil {
		panic(fmt.Errorf("error generating nonce: %v", err))
	}
	digest, err := factory.GetDefault().Hash(
		nonce,
		&bccsp.SHA256Opts{})
	if err != nil {
		panic(fmt.Errorf("error hashing nonce: %v", err))
	}
	return hex.EncodeToString(digest)
}
