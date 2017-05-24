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

	fabricClient "github.com/hyperledger/fabric-sdk-go/fabric-client"
	pb "github.com/hyperledger/fabric/protos/peer"
	"github.com/securekey/fabric-examples/fabric-cli/common"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var invokeCmd = &cobra.Command{
	Use:   "invoke",
	Short: "invoke chaincode.",
	Long:  "invoke chaincode",
	Run: func(cmd *cobra.Command, args []string) {
		if common.ChaincodeID == "" {
			fmt.Printf("\nMust specify the chaincode ID\n\n")
			cmd.HelpFunc()(cmd, args)
			return
		}
		action, err := newInvokeAction(cmd.Flags())
		if err != nil {
			common.Logger.Criticalf("Error while initializing invokeAction: %v", err)
			return
		}

		err = action.invoke()
		if err != nil {
			common.Logger.Criticalf("Error while running invokeAction: %v", err)
			return
		}
	},
}

// Cmd returns the invoke command
func getInvokeCmd() *cobra.Command {
	flags := invokeCmd.Flags()
	flags.String(common.PeerFlag, "", "The URL of the peer on which to invoke the chaincode, e.g. localhost:7051")
	flags.StringVar(&common.ChannelID, common.ChannelIDFlag, common.ChannelID, "The channel ID")
	flags.StringVar(&common.ChaincodeID, common.ChaincodeIDFlag, "", "The chaincode ID")
	flags.StringVar(&common.Args, common.ArgsFlag, common.Args, "The args in JSON format. Example: {\"Args\":\"invoke\",\"arg1\",\"arg2\"}")
	flags.IntVar(&common.Iterations, common.IterationsFlag, 1, "The number of times to invoke the chaincode.")
	flags.Int64Var(&common.SleepTime, common.SleepFlag, int64(100), "The number of milliseconds to sleep between invocations of the chaincode.")
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
	chain, err := action.NewChain()
	if err != nil {
		return fmt.Errorf("Error initializing chain: %v", err)
	}

	argBytes := []byte(common.Args)
	args := &common.ArgStruct{}
	err = json.Unmarshal(argBytes, args)
	if err != nil {
		return fmt.Errorf("Error unmarshaling JSON arg string: %v", err)
	}

	if common.Iterations > 1 {
		go action.invokeMultiple(chain, args.Args, common.Iterations)

		completed := false
		for !completed {
			select {
			case <-action.done:
				completed = true
			case <-time.After(5 * time.Second):
				fmt.Printf("... completed %d out of %d\n", action.numInvoked, common.Iterations)
			}
		}
	} else {
		if err := action.doInvoke(chain, args.Args); err != nil {
			fmt.Printf("Error invoking chaincode: %v\n", err)
		}
	}

	return nil
}

func (action *invokeAction) invokeMultiple(chain fabricClient.Chain, args []string, iterations int) {
	fmt.Printf("Invoking CC %d times ...\n", common.Iterations)
	for i := 0; i < common.Iterations; i++ {
		if err := action.doInvoke(chain, args); err != nil {
			fmt.Printf("Error invoking chaincode: %v\n", err)
		}
		if (i+1) < common.Iterations && common.SleepTime > 0 {
			time.Sleep(time.Duration(common.SleepTime) * time.Millisecond)
		}
		atomic.AddUint32(&action.numInvoked, 1)
	}
	fmt.Printf("Completed %d invocations\n", common.Iterations)
	action.done <- true
}

func (action *invokeAction) doInvoke(chain fabricClient.Chain, args []string) error {
	common.Logger.Infof("Invoking chaincode: %s or channel: %s, with args: [%v]\n", common.ChaincodeID, common.ChannelID, args)

	signedProposal, err := chain.CreateTransactionProposal(common.ChaincodeID, common.ChannelID, args, true, nil)
	if err != nil {
		return fmt.Errorf("SendTransactionProposal return error: %v", err)
	}

	transactionProposalResponses, err := chain.SendTransactionProposal(signedProposal, 0, action.Peers())
	if err != nil {
		return fmt.Errorf("SendTransactionProposal return error: %v", err)
	}

	var proposalErr error
	var responses []*fabricClient.TransactionProposalResponse
	for _, v := range transactionProposalResponses {
		if v.Err != nil {
			common.Logger.Errorf("invoke - TxID: %s, Endorser %s returned error: %v\n", signedProposal.TransactionID, v.Endorser, v.Err)
			proposalErr = fmt.Errorf("invoke Endorser %s return error: %v", v.Endorser, v.Err)
		} else {
			responses = append(responses, v)
			common.Logger.Debugf("invoke - TxID: %s, Endorser %s returned ProposalResponse: %v\n", signedProposal.TransactionID, v.Endorser, v.GetResponsePayload())
		}
	}

	if len(responses) == 0 {
		return proposalErr
	}

	common.Logger.Debugf("invoke - Creating transaction - TxID: %s ...\n", signedProposal.TransactionID)

	tx, err := chain.CreateTransaction(responses)
	if err != nil {
		return fmt.Errorf("CreateTransaction return error: %v", err)
	}

	common.Logger.Debugf("invoke - Sending transaction - TxID: %s ...\n", signedProposal.TransactionID)
	transactionResponses, err := chain.SendTransaction(tx)
	if err != nil {
		common.Logger.Criticalf("invoke - Unregistering Tx Event for txId: %s since the transaction was not able to be sent ...\n", signedProposal.TransactionID)
		return fmt.Errorf("SendTransaction returned error: %v", err)
	}

	for _, v := range transactionResponses {
		if v.Err != nil {
			common.Logger.Criticalf("Unregistering TX Event for txId: %s since received error on transaction response", signedProposal.TransactionID)
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

	common.Logger.Infof("Invocation successful!\n")
	return nil
}
