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

var queryChannelsCmd = &cobra.Command{
	Use:   "channels",
	Short: "Query channels",
	Long:  "Queries the channels of the specified peer",
	Run: func(cmd *cobra.Command, args []string) {
		action, err := newQueryChannelsAction(cmd.Flags())
		if err != nil {
			cliconfig.Config().Logger().Errorf("Error while initializing queryChannelsAction: %v", err)
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
			cliconfig.Config().Logger().Errorf("Error while running queryChannelsAction: %v", err)
		}
	},
}

func getQueryChannelsCmd() *cobra.Command {
	cliconfig.InitPeerURL(queryChannelsCmd.Flags())
	return queryChannelsCmd
}

type queryChannelsAction struct {
	action.Action
}

func newQueryChannelsAction(flags *pflag.FlagSet) (*queryChannelsAction, error) {
	action := &queryChannelsAction{}
	err := action.Initialize(flags)
	return action, err
}

func (a *queryChannelsAction) run() error {
	user, err := a.OrgAdminUser(a.OrgID())
	if err != nil {
		return err
	}

	client, err := a.ResourceMgmtClientForUser(user)
	if err != nil {
		return errors.Errorf("error getting fabric client: %s", err)
	}

	response, err := client.QueryChannels(resmgmt.WithTargets(a.Peer()))
	if err != nil {
		return err
	}

	fmt.Printf("Channels for peer [%s]\n", a.Peer().URL())

	a.Printer().PrintChannels(response.Channels)

	return nil
}
