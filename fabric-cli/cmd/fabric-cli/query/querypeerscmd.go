/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package query

import (
	"fmt"

	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/action"
	cliconfig "github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/config"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var queryPeersCmd = &cobra.Command{
	Use:   "peers",
	Short: "Query peers",
	Long:  "Queries the peers for the specified channel",
	Run: func(cmd *cobra.Command, args []string) {
		action, err := newQueryPeersAction(cmd.Flags())
		if err != nil {
			cliconfig.Config().Logger().Errorf("Error while initializing queryPeersAction: %v", err)
			return
		}
		defer action.Terminate()

		if cliconfig.Config().ChannelID() == "" {
			fmt.Printf("\nMust specify channel ID\n\n")
			cmd.HelpFunc()(cmd, args)
			return
		}

		err = action.run()
		if err != nil {
			cliconfig.Config().Logger().Errorf("Error while running queryPeersAction: %v", err)
		}
	},
}

func getQueryPeersCmd() *cobra.Command {
	flags := queryPeersCmd.Flags()
	cliconfig.InitChannelID(flags)
	cliconfig.InitPeerURL(flags)
	return queryPeersCmd
}

type queryPeersAction struct {
	action.Action
}

func newQueryPeersAction(flags *pflag.FlagSet) (*queryPeersAction, error) {
	action := &queryPeersAction{}
	err := action.Initialize(flags)
	return action, err
}

func (a *queryPeersAction) run() error {
	chProvider, err := a.ChannelProvider()
	if err != nil {
		return err
	}

	chContext, err := chProvider()
	if err != nil {
		return err
	}

	discovery, err := chContext.ChannelService().Discovery()
	if err != nil {
		return err
	}

	peers, err := discovery.GetPeers()
	if err != nil {
		return err
	}

	a.Printer().PrintPeers(peers)

	return nil
}
