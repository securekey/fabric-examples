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
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/common"
	cliconfig "github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/config"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var invokeCmd = &cobra.Command{
	Use:   "invoke",
	Short: "invoke chaincode.",
	Long:  "invoke chaincode",
	Run: func(cmd *cobra.Command, args []string) {
		if cliconfig.Config().ChaincodeID() == "" {
			fmt.Printf("\nMust specify the chaincode ID\n\n")
			cmd.HelpFunc()(cmd, args)
			return
		}
		action, err := newInvokeAction(cmd.Flags())
		if err != nil {
			cliconfig.Config().Logger().Errorf("Error while initializing invokeAction: %v", err)
			return
		}

		defer action.Terminate()

		err = action.invoke()
		if err != nil {
			cliconfig.Config().Logger().Errorf("Error while running invokeAction: %v", err)
		}
	},
}

func getInvokeCmd() *cobra.Command {
	flags := invokeCmd.Flags()
	cliconfig.Config().InitPeerURL(flags)
	cliconfig.Config().InitChannelID(flags)
	cliconfig.Config().InitChaincodeID(flags)
	cliconfig.Config().InitArgs(flags)
	cliconfig.Config().InitIterations(flags)
	cliconfig.Config().InitSleepTime(flags)
	cliconfig.Config().InitTimeout(flags)
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

	argBytes := []byte(cliconfig.Config().Args())
	args := &common.ArgStruct{}
	err = json.Unmarshal(argBytes, args)
	if err != nil {
		return fmt.Errorf("Error unmarshaling JSON arg string: %v", err)
	}

	if cliconfig.Config().Iterations() > 1 {
		go action.invokeMultiple(channelClient, args.Func, asBytes(args.Args), cliconfig.Config().Iterations())

		completed := false
		for !completed {
			select {
			case <-action.done:
				completed = true
			case <-time.After(cliconfig.Config().Timeout() * time.Millisecond):
				fmt.Printf("... completed %d out of %d\n", action.numInvoked, cliconfig.Config().Iterations())
			}
		}
	} else {
		if err := action.doInvoke(channelClient, args.Func, asBytes(args.Args)); err != nil {
			fmt.Printf("Error invoking chaincode: %v\n", err)
		} else {
			fmt.Println("Invocation successful!")
		}
	}

	return nil
}

func (action *invokeAction) invokeMultiple(chain apitxn.ChannelClient, fctn string, args [][]byte, iterations int) {
	fmt.Printf("Invoking CC %d times ...\n", iterations)
	for i := 0; i < iterations; i++ {
		if err := action.doInvoke(chain, fctn, args); err != nil {
			fmt.Printf("Error invoking chaincode: %v\n", err)
		} else {
			cliconfig.Config().Logger().Info("Invocation %d successful\n", i)
		}
		if (i+1) < iterations && cliconfig.Config().SleepTime() > 0 {
			time.Sleep(time.Duration(cliconfig.Config().SleepTime()) * time.Millisecond)
		}
		atomic.AddUint32(&action.numInvoked, 1)
	}
	fmt.Printf("Completed %d invocations\n", iterations)
	action.done <- true
}

func (action *invokeAction) ProcessTxProposalResponse(txProposalResponses []*apitxn.TransactionProposalResponse) ([]*apitxn.TransactionProposalResponse, error) {
	action.Printer().PrintTxProposalResponses(txProposalResponses)
	return txProposalResponses, nil
}

func (action *invokeAction) doInvoke(channel apitxn.ChannelClient, fctn string, args [][]byte) error {
	cliconfig.Config().Logger().Infof("Invoking chaincode: %s on channel: %s, peers: %s, function: %s, args: [%v]\n",
		cliconfig.Config().ChaincodeID(), cliconfig.Config().ChannelID(), asString(action.Peers()), fctn, args)

	txStatusEvents := make(chan apitxn.ExecuteTxResponse)
	txID, err := channel.ExecuteTxWithOpts(
		apitxn.ExecuteTxRequest{
			ChaincodeID: cliconfig.Config().ChaincodeID(),
			Fcn:         fctn,
			Args:        args,
		},
		apitxn.ExecuteTxOpts{
			TxFilter:           action,
			Notifier:           txStatusEvents,
			ProposalProcessors: action.ProposalProcessors(),
		},
	)
	if err != nil {
		return fmt.Errorf("SendTransactionProposal return error: %v", err)
	}

	cliconfig.Config().Logger().Debugf("invoke - Committing transaction - TxID: %s ...\n", txID)

	select {
	case s := <-txStatusEvents:
		if s.TxValidationCode == pb.TxValidationCode_VALID {
			return nil
		}
		return fmt.Errorf("invoke Error received from eventhub for txid(%s). Code: %s, Details: %s", txID, s.TxValidationCode, s.Error)
	case <-time.After(cliconfig.Config().Timeout() * time.Millisecond):
		return fmt.Errorf("timed out waiting to receive block event for txid(%s)", txID)
	}
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

func asBytes(args []string) [][]byte {
	bytes := make([][]byte, len(args))
	for i, arg := range args {
		bytes[i] = []byte(arg)
	}
	return bytes
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
