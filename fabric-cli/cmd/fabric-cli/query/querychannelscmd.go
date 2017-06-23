/*
Copyright SecureKey Technologies Inc. All Rights Reserved.


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at


      http://www.apache.org/licenses/LICENSE-2.0


Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
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
		if common.Config().PeerURL() == "" {
			fmt.Printf("\nMust specify the peer URL\n\n")
			cmd.HelpFunc()(cmd, args)
			return
		}
		action, err := newQueryChannelsAction(cmd.Flags())
		if err != nil {
			common.Config().Logger().Criticalf("Error while initializing queryChannelsAction: %v", err)
			return
		}

		err = action.run()
		if err != nil {
			common.Config().Logger().Criticalf("Error while running queryChannelsAction: %v", err)
			return
		}
	},
}

// getQueryChannelsCmd returns the Query block action command
func getQueryChannelsCmd() *cobra.Command {
	common.Config().InitPeerURL(queryChannelsCmd.Flags())
	return queryChannelsCmd
}

type queryChannelsAction struct {
	common.ActionImpl
}

func newQueryChannelsAction(flags *pflag.FlagSet) (*queryChannelsAction, error) {
	action := &queryChannelsAction{}
	err := action.Initialize(flags)
	return action, err
}

func (action *queryChannelsAction) run() error {
	peer := action.PeerFromURL(common.Config().PeerURL())
	if peer == nil {
		return fmt.Errorf("unknown peer URL: %s", common.Config().PeerURL())
	}

	response, err := action.Client().QueryChannels(peer)
	if err != nil {
		return err
	}

	fmt.Printf("Channels for peer [%s]\n", common.Config().PeerURL())

	action.Printer().PrintChannels(response.Channels)

	return nil
}
