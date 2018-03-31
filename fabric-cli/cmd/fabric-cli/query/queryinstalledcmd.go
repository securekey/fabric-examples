/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package query

import (
	"fmt"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/pkg/errors"
	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/action"
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

		if len(cliconfig.Config().PeerURLs()) != 1 {
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
	action.Action
}

func newqueryInstalledAction(flags *pflag.FlagSet) (*queryInstalledAction, error) {
	action := &queryInstalledAction{}
	err := action.Initialize(flags)
	return action, err
}

func (a *queryInstalledAction) run() error {

	url := cliconfig.Config().PeerURLs()
	if len(url) != 1 {
		return errors.New("must specify exactly one peer URL")
	}
	peer, ok := a.PeerFromURL(url[0])
	if !ok {
		return fmt.Errorf("invalid peer URL: %s", url)
	}

	user, err := a.OrgAdminUser(a.OrgID())
	if err != nil {
		return err
	}

	client, err := a.ResourceMgmtClientForUser(user)
	if err != nil {
		return err
	}

	response, err := client.QueryInstalledChaincodes(resmgmt.WithTargets(peer))
	if err != nil {
		return err
	}

	fmt.Printf("Chaincodes for peer [%s]\n", a.Peer().URL())
	a.Printer().PrintChaincodes(response.Chaincodes)
	return nil
}
