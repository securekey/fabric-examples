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

var queryChannelsCmd = &cobra.Command{
	Use:   "channels",
	Short: "Query channels",
	Long:  "Queries the channels of the specified peer",
	Run: func(cmd *cobra.Command, args []string) {
		action, err := newQueryChannelsAction(cmd.Flags())
		if err != nil {
			common.Config().Logger().Criticalf("Error while initializing queryChannelsAction: %v", err)
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
			common.Config().Logger().Criticalf("Error while running queryChannelsAction: %v", err)
		}
	},
}

func getQueryChannelsCmd() *cobra.Command {
	common.Config().InitPeerURL(queryChannelsCmd.Flags())
	return queryChannelsCmd
}

type queryChannelsAction struct {
	common.Action
}

func newQueryChannelsAction(flags *pflag.FlagSet) (*queryChannelsAction, error) {
	action := &queryChannelsAction{}
	err := action.Initialize(flags)
	return action, err
}

func (action *queryChannelsAction) run() error {
	context := action.SetUserContext(action.OrgAdminUser(action.OrgID()))
	defer context.Restore()

	response, err := action.Client().QueryChannels(action.Peer())
	if err != nil {
		return err
	}

	fmt.Printf("Channels for peer [%s]\n", action.Peer().URL())

	action.Printer().PrintChannels(response.Channels)

	return nil
}
