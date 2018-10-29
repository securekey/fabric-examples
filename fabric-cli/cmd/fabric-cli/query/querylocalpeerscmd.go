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

var queryLocalPeersCmd = &cobra.Command{
	Use:   "localpeers",
	Short: "Query local peers",
	Long:  "Queries the peers for the specified org",
	Run: func(cmd *cobra.Command, args []string) {
		action, err := newQueryLocalPeersAction(cmd.Flags())
		if err != nil {
			cliconfig.Config().Logger().Errorf("Error while initializing queryLocalPeersAction: %v", err)
			return
		}
		defer action.Terminate()

		if cliconfig.Config().OrgID() == "" {
			fmt.Printf("\nMust specify org ID\n\n")
			cmd.HelpFunc()(cmd, args)
			return
		}

		err = action.run()
		if err != nil {
			cliconfig.Config().Logger().Errorf("Error while running queryLocalPeersAction: %v", err)
		}
	},
}

func getQueryLocalPeersCmd() *cobra.Command {
	flags := queryLocalPeersCmd.Flags()
	cliconfig.InitOrgIDs(flags)
	cliconfig.InitPeerURL(flags)
	return queryLocalPeersCmd
}

type queryLocalPeersAction struct {
	action.Action
}

func newQueryLocalPeersAction(flags *pflag.FlagSet) (*queryLocalPeersAction, error) {
	action := &queryLocalPeersAction{}
	err := action.Initialize(flags)
	return action, err
}

func (a *queryLocalPeersAction) run() error {
	localContext, err := a.LocalContext()
	if err != nil {
		return err
	}

	peers, err := localContext.LocalDiscoveryService().GetPeers()
	if err != nil {
		return err
	}

	a.Printer().PrintPeers(peers)

	return nil
}
