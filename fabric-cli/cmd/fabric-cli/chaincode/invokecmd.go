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

package chaincode

import (
	"encoding/json"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/hyperledger/fabric-sdk-go/api"
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

		err = action.invoke()
		if err != nil {
			common.Config().Logger().Criticalf("Error while running invokeAction: %v", err)
			return
		}
	},
}

// Cmd returns the invoke command
func getInvokeCmd() *cobra.Command {
	flags := invokeCmd.Flags()
	common.Config().InitPeerURL(flags)
	common.Config().InitChannelID(flags)
	common.Config().InitChaincodeID(flags)
	common.Config().InitArgs(flags)
	common.Config().InitIterations(flags)
	common.Config().InitSleepTime(flags)
	return invokeCmd
}

type invokeAction struct {
	common.ActionImpl
	numInvoked uint32
	done       chan bool
}

func newInvokeAction(flags *pflag.FlagSet) (*invokeAction, error) {
	action := &invokeAction{done: make(chan bool)}
	err := action.Initialize(flags)
	return action, err
}

func (action *invokeAction) invoke() error {
	chain, err := action.NewChannel()
	if err != nil {
		return fmt.Errorf("Error initializing chain: %v", err)
	}

	argBytes := []byte(common.Config().Args())
	args := &common.ArgStruct{}
	err = json.Unmarshal(argBytes, args)
	if err != nil {
		return fmt.Errorf("Error unmarshaling JSON arg string: %v", err)
	}

	if common.Config().Iterations() > 1 {
		go action.invokeMultiple(chain, args.Args, common.Config().Iterations())

		completed := false
		for !completed {
			select {
			case <-action.done:
				completed = true
			case <-time.After(5 * time.Second):
				fmt.Printf("... completed %d out of %d\n", action.numInvoked, common.Config().Iterations())
			}
		}
	} else {
		if err := action.doInvoke(chain, args.Args); err != nil {
			fmt.Printf("Error invoking chaincode: %v\n", err)
		}
	}

	return nil
}

func (action *invokeAction) invokeMultiple(chain api.Channel, args []string, iterations int) {
	fmt.Printf("Invoking CC %d times ...\n", iterations)
	for i := 0; i < iterations; i++ {
		if err := action.doInvoke(chain, args); err != nil {
			fmt.Printf("Error invoking chaincode: %v\n", err)
		}
		if (i+1) < iterations && common.Config().SleepTime() > 0 {
			time.Sleep(time.Duration(common.Config().SleepTime()) * time.Millisecond)
		}
		atomic.AddUint32(&action.numInvoked, 1)
	}
	fmt.Printf("Completed %d invocations\n", iterations)
	action.done <- true
}

func (action *invokeAction) doInvoke(chain api.Channel, args []string) error {
	common.Config().Logger().Infof("Invoking chaincode: %s or channel: %s, with args: [%v]\n", common.Config().ChaincodeID(), common.Config().ChannelID(), args)

	signedProposal, err := chain.CreateTransactionProposal(common.Config().ChaincodeID(), common.Config().ChannelID(), args, true, nil)
	if err != nil {
		return fmt.Errorf("SendTransactionProposal return error: %v", err)
	}

	transactionProposalResponses, err := chain.SendTransactionProposal(signedProposal, 0, action.Peers())
	if err != nil {
		return fmt.Errorf("SendTransactionProposal return error: %v", err)
	}

	var proposalErr error
	var responses []*api.TransactionProposalResponse
	for _, v := range transactionProposalResponses {
		if v.Err != nil {
			common.Config().Logger().Errorf("invoke - TxID: %s, Endorser %s returned error: %v\n", signedProposal.TransactionID, v.Endorser, v.Err)
			proposalErr = fmt.Errorf("invoke Endorser %s return error: %v", v.Endorser, v.Err)
		} else {
			responses = append(responses, v)
			common.Config().Logger().Debugf("invoke - TxID: %s, Endorser %s returned ProposalResponse: %v\n", signedProposal.TransactionID, v.Endorser, v.ProposalResponse.Response.Payload)
		}
	}

	if len(responses) == 0 {
		return proposalErr
	}

	common.Config().Logger().Debugf("invoke - Creating transaction - TxID: %s ...\n", signedProposal.TransactionID)

	tx, err := chain.CreateTransaction(responses)
	if err != nil {
		return fmt.Errorf("CreateTransaction return error: %v", err)
	}

	common.Config().Logger().Debugf("invoke - Sending transaction - TxID: %s ...\n", signedProposal.TransactionID)
	transactionResponses, err := chain.SendTransaction(tx)
	if err != nil {
		common.Config().Logger().Criticalf("invoke - Unregistering Tx Event for txId: %s since the transaction was not able to be sent ...\n", signedProposal.TransactionID)
		return fmt.Errorf("SendTransaction returned error: %v", err)
	}

	for _, v := range transactionResponses {
		if v.Err != nil {
			common.Config().Logger().Criticalf("Unregistering TX Event for txId: %s since received error on transaction response", signedProposal.TransactionID)
			return fmt.Errorf("Orderer %s return error: %v", v.Orderer, v.Err)
		}
	}
	done := make(chan bool)
	fail := make(chan error)

	action.EventHub().RegisterTxEvent(signedProposal.TransactionID, func(txID string, code pb.TxValidationCode, err error) {
		if err != nil {
			fail <- err
		} else {
			fmt.Printf("invoke receive success event for txid(%s)\n", txID)
			done <- true
		}

	})

	select {
	case <-done:
	case <-fail:
		return fmt.Errorf("invoke Error received from eventhub for txid(%s) error(%v)", signedProposal.TransactionID, fail)
	case <-time.After(time.Second * 60):
		return fmt.Errorf("timed out waiting to receive block event for txid(%s)", signedProposal.TransactionID)
	}

	common.Config().Logger().Infof("Invocation successful!\n")
	return nil
}
