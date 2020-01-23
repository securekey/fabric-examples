/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package channel

import (
	"fmt"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/pkg/errors"
	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/action"
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
	cliconfig.InitChannelID(flags)
	cliconfig.InitOrdererURL(flags)
	cliconfig.InitPeerURL(flags)
	return chainJoinCmd
}

type channelJoinAction struct {
	action.Action
}

func newChannelJoinAction(flags *pflag.FlagSet) (*channelJoinAction, error) {
	action := &channelJoinAction{}
	err := action.Initialize(flags)
	if err != nil {
		return nil, err
	}
	if len(action.Peers()) == 0 {
		return nil, errors.Errorf("at least one peer is required for join")
	}
	return action, nil
}

func (a *channelJoinAction) invoke() error {
	fmt.Printf("Attempting to join channel: %s\n", cliconfig.Config().ChannelID())

	var lastErr error
	for orgID, peers := range a.PeersByOrg() {
		fmt.Printf("Joining channel %s on org[%s] peers:\n", cliconfig.Config().ChannelID(), orgID)
		for _, peer := range peers {
			fmt.Printf("-- %s\n", peer.URL())
		}
		err := a.joinChannel(orgID, peers)
		if err != nil {
			lastErr = err
		}
	}
	return lastErr
}

func (a *channelJoinAction) joinChannel(orgID string, peers []fab.Peer) error {
	cliconfig.Config().Logger().Debugf("Joining channel [%s]...\n", cliconfig.Config().ChannelID())

	fmt.Printf("==========> JOIN ORG: %s\n", orgID)

	resMgmtClient, err := a.ResourceMgmtClientForOrg(orgID)
	if err != nil {
		return err
	}

	orderer, err := a.RandomOrderer()
	if err != nil {
		return err
	}

	err = resMgmtClient.JoinChannel(cliconfig.Config().ChannelID(), resmgmt.WithTargets(peers...), resmgmt.WithOrderer(orderer))
	if err != nil {
		return errors.WithMessage(err, "Could not join channel: %v")
	}

	fmt.Printf("Channel %s joined!\n", cliconfig.Config().ChannelID())

	return nil
}
