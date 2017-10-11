/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package channel

import (
	"fmt"

	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	channel "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/channel"
	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/common"
	cliconfig "github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/config"
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
			cliconfig.Config().Logger().Errorf("Error while initializing channelJoinAction: %v", err)
			return
		}

		defer action.Terminate()

		err = action.invoke()
		if err != nil {
			cliconfig.Config().Logger().Errorf("Error while running channelJoinAction: %v", err)
		}
	},
}

func getChannelJoinCmd() *cobra.Command {
	flags := chainJoinCmd.Flags()
	cliconfig.Config().InitChannelID(flags)
	cliconfig.Config().InitOrdererURL(flags)
	cliconfig.Config().InitPeerURL(flags)
	return chainJoinCmd
}

type channelJoinAction struct {
	common.Action
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
	fmt.Printf("Attempting to join channel: %s\n", cliconfig.Config().ChannelID())

	for orgID, peers := range action.PeersByOrg() {
		fmt.Printf("Joining channel %s on org[%s] peers:\n", cliconfig.Config().ChaincodeID(), orgID)
		for _, peer := range peers {
			fmt.Printf("-- %s\n", peer.URL())
		}
		err := action.joinChannel(orgID, peers)
		if err != nil {
			return err
		}
	}
	return nil
}

func (action *channelJoinAction) joinChannel(orgID string, peers []apifabclient.Peer) error {
	cliconfig.Config().Logger().Debugf("Joining channel [%s]...\n", cliconfig.Config().ChannelID())

	channelClient, err := action.OrgAdminChannelClient(orgID)
	if err != nil {
		return fmt.Errorf("Error getting admin channel client: %v", err)
	}

	// FIXME: Remove this when SDK includes a SystemChannelClient
	channelContext := channelClient.(*channel.Channel).ClientContext()
	txnID, err := channelContext.NewTxnID()
	if err != nil {
		return err
	}

	genesisBlock, err := channelClient.GenesisBlock(&apifabclient.GenesisBlockRequest{TxnID: txnID})
	if err != nil {
		return fmt.Errorf("Error getting genesis block: %v", err)
	}

	ctx := channelClient.(*channel.Channel).ClientContext()
	cliconfig.Config().Logger().Errorf("*****Channel client context: Name [%s], MSPID [%s]\n", ctx.UserContext().Name(), ctx.UserContext().MspID())

	if err = channelClient.JoinChannel(&apifabclient.JoinChannelRequest{
		Targets:      peers,
		GenesisBlock: genesisBlock,
		TxnID:        txnID,
	}); err != nil {
		return fmt.Errorf("Could not join channel: %v", err)
	}

	fmt.Printf("Channel %s joined!\n", cliconfig.Config().ChannelID())

	return nil
}
