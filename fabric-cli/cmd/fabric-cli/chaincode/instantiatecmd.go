/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chaincode

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/hyperledger/fabric-sdk-go/api"
	pb "github.com/hyperledger/fabric/protos/peer"
	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/common"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var instantiateCmd = &cobra.Command{
	Use:   "instantiate",
	Short: "Instantiate chaincode.",
	Long:  "Instantiates the chaincode",
	Run: func(cmd *cobra.Command, args []string) {
		if common.Config().ChaincodeID() == "" {
			fmt.Printf("\nMust specify the chaincode ID\n\n")
			cmd.HelpFunc()(cmd, args)
			return
		}
		if common.Config().ChaincodePath() == "" {
			fmt.Printf("\nMust specify the path of the chaincode\n\n")
			cmd.HelpFunc()(cmd, args)
			return
		}
		action, err := newInstantiateAction(cmd.Flags())
		if err != nil {
			common.Config().Logger().Criticalf("Error while initializing instantiateAction: %v", err)
			return
		}

		err = action.invoke()
		if err != nil {
			common.Config().Logger().Criticalf("Error while running instantiateAction: %v", err)
			return
		}
	},
}

// Cmd returns the install command
func getInstantiateCmd() *cobra.Command {
	flags := instantiateCmd.Flags()
	common.Config().InitPeerURL(flags)
	common.Config().InitChannelID(flags)
	common.Config().InitChaincodeID(flags)
	common.Config().InitChaincodePath(flags)
	common.Config().InitChaincodeVersion(flags)
	common.Config().InitArgs(flags)
	return instantiateCmd
}

type instantiateAction struct {
	common.ActionImpl
}

func newInstantiateAction(flags *pflag.FlagSet) (*instantiateAction, error) {
	action := &instantiateAction{}
	err := action.Initialize(flags)
	if len(action.Peers()) == 0 {
		return nil, fmt.Errorf("a peer must be specified")
	}
	return action, err
}

func (action *instantiateAction) invoke() error {
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

	peer := action.Peers()[0]
	fmt.Printf("instantiating chaincode on peer: %v\n", peer.GetURL())

	err = instantiateChaincode(chain, action.EventHub(), common.Config().ChannelID(), common.Config().ChaincodeID(), common.Config().ChaincodeVersion(), common.Config().ChaincodePath(), args.Args, []api.Peer{peer})
	if err != nil {
		return fmt.Errorf("Error instantiating chaincode: %v", err)
	}

	fmt.Printf("...successfuly instantiated chaincode %s on channel %s.\n", common.Config().ChaincodeID(), common.Config().ChannelID())

	return nil
}

func instantiateChaincode(chain api.Channel, eventHub api.EventHub, chainID string, chaincodeID string, chainCodeVersion string, chainCodePath string, args []string, targetPeers []api.Peer) error {
	err := instantiateCC(chain, eventHub, chainID, chaincodeID, chainCodeVersion, chainCodePath, args, targetPeers)
	if err != nil {
		if strings.Contains(err.Error(), "Chaincode exists "+chaincodeID) {
			// Ignore
			common.Config().Logger().Infof("Chaincode %s already instantiated.", chaincodeID)
			return nil
		}
		return fmt.Errorf("instantiateCC returned error: %v", err)
	}

	return nil
}

func instantiateCC(chain api.Channel, eventHub api.EventHub, chainID string, chainCodeID string, chainCodeVersion string, chainCodePath string, args []string, targetPeers []api.Peer) error {
	transactionProposalResponse, txID, err := chain.SendInstantiateProposal(chainCodeID, chainID, args, chainCodePath, chainCodeVersion, targetPeers)
	if err != nil {
		return fmt.Errorf("SendInstantiateProposal return error: %v", err)
	}

	for _, v := range transactionProposalResponse {
		if v.Err != nil {
			return fmt.Errorf("SendInstantiateProposal Endorser %s return error: %v", v.Endorser, v.Err)
		}
		fmt.Printf("SendInstantiateProposal Endorser '%s' return ProposalResponse status:%v\n", v.Endorser, v.Status)
	}

	tx, err := chain.CreateTransaction(transactionProposalResponse)
	if err != nil {
		return fmt.Errorf("CreateTransaction return error: %v", err)

	}
	transactionResponse, err := chain.SendTransaction(tx)
	if err != nil {
		return fmt.Errorf("SendTransaction return error: %v", err)

	}
	for _, v := range transactionResponse {
		if v.Err != nil {
			return fmt.Errorf("Orderer %s return error: %v", v.Orderer, v.Err)
		}
	}
	done := make(chan bool)
	fail := make(chan error)

	eventHub.RegisterTxEvent(txID, func(txID string, code pb.TxValidationCode, err error) {
		if err != nil {
			fail <- err
		} else {
			fmt.Printf("instantiateCC receive success event for txid(%s)\n", txID)
			done <- true
		}

	})

	select {
	case <-done:
	case <-fail:
		return fmt.Errorf("instantiateCC Error received from eventhub for txid(%s) error(%v)", txID, fail)
	case <-time.After(time.Second * 60):
		return fmt.Errorf("timed out waiting to receive block event for txid(%s)", txID)
	}
	return nil

}
