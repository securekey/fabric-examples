/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chaincode

import (
	"encoding/json"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	pb "github.com/hyperledger/fabric/protos/peer"
	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/common"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var invokeCmd = &cobra.Command{
	Use:   "invoke",
	Short: "invoke chaincode.",
	Long:  "invoke chaincode",
	Run: func(cmd *cobra.Command, args []string) {
		if common.Config().ChaincodeID() == "" {
			fmt.Printf("\nMust specify the chaincode ID\n\n")
			cmd.HelpFunc()(cmd, args)
			return
		}
		action, err := newInvokeAction(cmd.Flags())
		if err != nil {
			common.Config().Logger().Criticalf("Error while initializing invokeAction: %v", err)
			return
		}

		defer action.Terminate()

		err = action.invoke()
		if err != nil {
			common.Config().Logger().Criticalf("Error while running invokeAction: %v", err)
		}
	},
}

func getInvokeCmd() *cobra.Command {
	flags := invokeCmd.Flags()
	common.Config().InitPeerURL(flags)
	common.Config().InitChannelID(flags)
	common.Config().InitChaincodeID(flags)
	common.Config().InitArgs(flags)
	common.Config().InitIterations(flags)
	common.Config().InitSleepTime(flags)
	common.Config().InitTimeout(flags)
	return invokeCmd
}

type invokeAction struct {
	common.Action
	numInvoked uint32
	done       chan bool
}

type txStatus struct {
	txID string
	code pb.TxValidationCode
	err  error
}

func newInvokeAction(flags *pflag.FlagSet) (*invokeAction, error) {
	action := &invokeAction{done: make(chan bool)}
	err := action.Initialize(flags)
	return action, err
}

func (action *invokeAction) invoke() error {
	channelClient, err := action.ChannelClient()
	if err != nil {
		return fmt.Errorf("Error getting channel client: %v", err)
	}

	argBytes := []byte(common.Config().Args())
	args := &common.ArgStruct{}
	err = json.Unmarshal(argBytes, args)
	if err != nil {
		return fmt.Errorf("Error unmarshaling JSON arg string: %v", err)
	}

	if common.Config().Iterations() > 1 {
		go action.invokeMultiple(channelClient, args.Func, args.Args, common.Config().Iterations())

		completed := false
		for !completed {
			select {
			case <-action.done:
				completed = true
			case <-time.After(common.Config().Timeout() * time.Millisecond):
				fmt.Printf("... completed %d out of %d\n", action.numInvoked, common.Config().Iterations())
			}
		}
	} else {
		if err := action.doInvoke(channelClient, args.Func, args.Args); err != nil {
			fmt.Printf("Error invoking chaincode: %v\n", err)
		} else {
			fmt.Println("Invocation successful!")
		}
	}

	return nil
}

func (action *invokeAction) invokeMultiple(chain apifabclient.Channel, fctn string, args []string, iterations int) {
	fmt.Printf("Invoking CC %d times ...\n", iterations)
	for i := 0; i < iterations; i++ {
		if err := action.doInvoke(chain, fctn, args); err != nil {
			fmt.Printf("Error invoking chaincode: %v\n", err)
		} else {
			common.Config().Logger().Info("Invocation %d successful\n", i)
		}
		if (i+1) < iterations && common.Config().SleepTime() > 0 {
			time.Sleep(time.Duration(common.Config().SleepTime()) * time.Millisecond)
		}
		atomic.AddUint32(&action.numInvoked, 1)
	}
	fmt.Printf("Completed %d invocations\n", iterations)
	action.done <- true
}

func (action *invokeAction) doInvoke(channel apifabclient.Channel, fctn string, args []string) error {
	common.Config().Logger().Infof("Invoking chaincode: %s on channel: %s, peers: %s, function: %s, args: [%v]\n",
		common.Config().ChaincodeID(), common.Config().ChannelID(), asString(action.Peers()), fctn, args)

	targets := make([]apitxn.ProposalProcessor, len(action.Peers()))
	for i, p := range action.Peers() {
		targets[i] = p
	}

	transactionProposalResponses, txnID, err := channel.SendTransactionProposal(apitxn.ChaincodeInvokeRequest{
		Targets:      targets,
		Fcn:          fctn,
		Args:         args,
		TransientMap: nil,
		ChaincodeID:  common.Config().ChaincodeID(),
	})
	if err != nil {
		return fmt.Errorf("SendTransactionProposal return error: %v", err)
	}

	action.Printer().PrintTxProposalResponses(transactionProposalResponses)

	var proposalErr error
	var responses []*apitxn.TransactionProposalResponse
	for _, v := range transactionProposalResponses {
		if v.Err != nil {
			proposalErr = v.Err
		} else {
			responses = append(responses, v)
		}
	}

	if len(responses) == 0 {
		if proposalErr == nil {
			return fmt.Errorf("no responses were received")
		}
		return proposalErr
	}

	status, err := action.registerTxEvent(txnID)
	if err != nil {
		return err
	}

	common.Config().Logger().Debugf("invoke - Committing transaction - TxID: %s ...\n", txnID.ID)
	if err = action.commit(channel, responses); err != nil {
		common.Config().Logger().Criticalf("invoke - Unregistering Tx Event for txId: %s since the transaction was not able to be sent ...\n", txnID.ID)
		return err
	}

	select {
	case s := <-status:
		if s.code == pb.TxValidationCode_VALID {
			return nil
		}
		return fmt.Errorf("invoke Error received from eventhub for txid(%s). Code: %s, Details: %s", txnID.ID, s.code, s.err)
	case <-time.After(common.Config().Timeout() * time.Millisecond):
		return fmt.Errorf("timed out waiting to receive block event for txid(%s)", txnID.ID)
	}
}

func (action *invokeAction) commit(channel apifabclient.Channel, responses []*apitxn.TransactionProposalResponse) error {
	tx, err := channel.CreateTransaction(responses)
	if err != nil {
		return fmt.Errorf("CreateTransaction return error: %v", err)
	}

	_, err = channel.SendTransaction(tx)
	if err != nil {
		return fmt.Errorf("SendTransaction returned error: %v", err)
	}

	return nil
}

func (action *invokeAction) registerTxEvent(txnID apitxn.TransactionID) (chan txStatus, error) {
	eventHub, err := action.EventHub()
	if err != nil {
		return nil, err
	}

	status := make(chan txStatus)

	eventHub.RegisterTxEvent(txnID, func(txID string, code pb.TxValidationCode, err error) {
		status <- txStatus{txID: txID, code: code, err: err}
	})

	return status, nil
}

func asString(peers []apifabclient.Peer) string {
	str := "["
	for i, peer := range peers {
		if peer.Name() != "" {
			str += peer.Name()
		} else {
			str += peer.URL()
		}
		if i+1 < len(peers) {
			str += ","
		}
	}
	str += "]"
	return str
}
