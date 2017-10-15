/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package query

import (
	"fmt"

	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/common"
	cliconfig "github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/config"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var queryInstalledCmd = &cobra.Command{
	Use:   "installed",
	Short: "Query installed chaincodes",
	Long:  "Queries the chaincodes installed to the specified peer",
	Run: func(cmd *cobra.Command, args []string) {
		action, err := newqueryInstalledAction(cmd.Flags())
		if err != nil {
			cliconfig.Config().Logger().Errorf("Error while initializing queryInstalledAction: %v", err)
			return
		}

		if len(action.Peers()) != 1 {
			fmt.Printf("\nMust specify exactly one peer URL\n\n")
			cmd.HelpFunc()(cmd, args)
			return
		}

		defer action.Terminate()

		err = action.run()
		if err != nil {
			cliconfig.Config().Logger().Errorf("Error while running queryInstalledAction: %v", err)
		}
	},
}

func getQueryInstalledCmd() *cobra.Command {
	cliconfig.InitPeerURL(queryInstalledCmd.Flags())
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
	user, err := action.OrgAdminUser(action.OrgID())
	if err != nil {
		return err
	}

	client, err := action.ClientForUser(action.OrgID(), user)
	if err != nil {
		return err
	}

	response, err := client.QueryInstalledChaincodes(action.Peer())
	if err != nil {
		return err
	}

	fmt.Printf("Chaincodes for peer [%s]\n", action.Peer().URL())
	action.Printer().PrintChaincodes(response.Chaincodes)
	return nil
}
