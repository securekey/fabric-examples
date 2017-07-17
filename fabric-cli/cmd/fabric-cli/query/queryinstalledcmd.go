/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package query

import (
	"fmt"

	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/common"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var queryInstalledCmd = &cobra.Command{
	Use:   "installed",
	Short: "Query installed chaincodes",
	Long:  "Queries the chaincodes installed to the specified peer",
	Run: func(cmd *cobra.Command, args []string) {
		if common.Config().PeerURL() == "" {
			fmt.Printf("\nMust specify the peer URL\n\n")
			cmd.HelpFunc()(cmd, args)
			return
		}
		action, err := newqueryInstalledAction(cmd.Flags())
		if err != nil {
			common.Config().Logger().Criticalf("Error while initializing queryInstalledAction: %v", err)
			return
		}

		defer action.Terminate()

		err = action.run()
		if err != nil {
			common.Config().Logger().Criticalf("Error while running queryInstalledAction: %v", err)
		}
	},
}

func getQueryInstalledCmd() *cobra.Command {
	common.Config().InitPeerURL(queryInstalledCmd.Flags())
	return queryInstalledCmd
}

type queryInstalledAction struct {
	common.Action
}

func newqueryInstalledAction(flags *pflag.FlagSet) (*queryInstalledAction, error) {
	action := &queryInstalledAction{}
	err := action.Initialize(flags)
	return action, err
}

func (action *queryInstalledAction) run() error {
	peer := action.PeerFromURL(common.Config().PeerURL())
	if peer == nil {
		return fmt.Errorf("unknown peer URL: %s", common.Config().PeerURL())
	}

	orgID, err := action.OrgOfPeer(peer.URL())
	if err != nil {
		return err
	}

	context := action.SetUserContext(action.OrgAdminUser(orgID))
	defer context.Restore()

	response, err := action.Client().QueryInstalledChaincodes(peer)
	if err != nil {
		return err
	}

	fmt.Printf("Chaincodes for peer [%s]\n", peer.URL())
	action.Printer().PrintChaincodes(response.Chaincodes)
	return nil
}
