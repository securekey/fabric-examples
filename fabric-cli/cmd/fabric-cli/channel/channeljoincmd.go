/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package channel

import (
	"fmt"

	"github.com/hyperledger/fabric-sdk-go/api"
	"github.com/hyperledger/fabric/common/crypto"
	protos_utils "github.com/hyperledger/fabric/protos/utils"
	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/common"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var chainJoinCmd = &cobra.Command{
	Use:   "join",
	Short: "Join Channel",
	Long:  "Join a channel",
	Run: func(cmd *cobra.Command, args []string) {
		action, err := newChannelJoinAction(cmd.Flags())
		if err != nil {
			common.Config().Logger().Criticalf("Error while initializing channelJoinAction: %v", err)
			return
		}

		err = action.invoke()
		if err != nil {
			common.Config().Logger().Criticalf("Error while running channelJoinAction: %v", err)
			return
		}
	},
}

// getChannelJoinCmd returns the chainJoinAction command
func getChannelJoinCmd() *cobra.Command {
	flags := chainJoinCmd.Flags()
	common.Config().InitChannelID(flags)
	common.Config().InitOrdererURL(flags)
	common.Config().InitPeerURL(flags)
	return chainJoinCmd
}

type channelJoinAction struct {
	common.ActionImpl
}

func newChannelJoinAction(flags *pflag.FlagSet) (*channelJoinAction, error) {
	action := &channelJoinAction{}
	err := action.Initialize(flags)
	if err != nil {
		return nil, err
	}
	if len(action.Peers()) == 0 {
		return nil, fmt.Errorf("at least one peer is required for join")
	}
	return action, nil
}

func (action *channelJoinAction) invoke() error {
	channel, err := action.NewChannel()
	if err != nil {
		return err
	}

	fmt.Printf("Attempting to join channel: %s\n", common.Config().ChannelID())

	err = action.joinChannel(channel)
	if err != nil {
		return fmt.Errorf("Could not join channel: %v", err)
	}
	fmt.Println("Channel joined!")

	return nil
}

func (action *channelJoinAction) joinChannel(channel api.Channel) error {
	nonce, err := crypto.GetRandomNonce()
	if err != nil {
		return fmt.Errorf("Could not compute nonce: %s", err)
	}

	signingIdentity, err := action.Client().GetIdentity()
	if err != nil {
		return fmt.Errorf("Could not get signing identity: %s", err)
	}

	txID, err := protos_utils.ComputeProposalTxID(nonce, signingIdentity)
	if err != nil {
		return fmt.Errorf("Could not compute TxID: %s", err)
	}

	genesisBlockRequest := &api.GenesisBlockRequest{
		TxID:  txID,
		Nonce: nonce,
	}
	genesisBlock, err := channel.GetGenesisBlock(genesisBlockRequest)
	if err != nil {
		return fmt.Errorf("Error getting genesis block: %v", err)
	}

	nonce, err = crypto.GetRandomNonce()
	if err != nil {
		return fmt.Errorf("Could not compute nonce: %s", err)
	}

	txID, err = protos_utils.ComputeProposalTxID(nonce, signingIdentity)
	if err != nil {
		return fmt.Errorf("Could not compute TxID: %s", err)
	}

	req := &api.JoinChannelRequest{
		Targets:      action.Peers(),
		GenesisBlock: genesisBlock,
		TxID:         txID,
		Nonce:        nonce,
	}

	return channel.JoinChannel(req)
}
