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

	"github.com/securekey/fabric-examples/fabric-sdk-go/fabric-cli/common"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var queryInstalledCmd = &cobra.Command{
	Use:   "installed",
	Short: "Query installed chaincodes",
	Long:  "Queries the chaincodes installed to the specified peer",
	Run: func(cmd *cobra.Command, args []string) {
		if peerURL == "" {
			fmt.Printf("\nMust specify the peer URL\n\n")
			cmd.HelpFunc()(cmd, args)
			return
		}
		action, err := newqueryInstalledAction(cmd.Flags())
		if err != nil {
			common.Logger.Criticalf("Error while initializing queryInstalledAction: %v", err)
			return
		}

		err = action.run()
		if err != nil {
			common.Logger.Criticalf("Error while running queryInstalledAction: %v", err)
			return
		}
	},
}

// getQueryChannelsCmd returns the Query installed CCs action command
func getQueryInstalledCmd() *cobra.Command {
	flags := queryInstalledCmd.Flags()
	flags.StringVar(&peerURL, common.PeerFlag, "", "The URL of the peer to query, e.g. localhost:7051")
	return queryInstalledCmd
}

type queryInstalledAction struct {
	common.ActionImpl
}

func newqueryInstalledAction(flags *pflag.FlagSet) (*queryInstalledAction, error) {
	action := &queryInstalledAction{}
	err := action.Initialize(flags)
	return action, err
}

func (action *queryInstalledAction) run() error {
	peer := action.PeerFromURL(peerURL)
	if peer == nil {
		return fmt.Errorf("unknown peer URL: %s", peerURL)
	}

	response, err := action.Client().QueryInstalledChaincodes(peer)
	if err != nil {
		return err
	}

	fmt.Printf("Chaincodes for peer [%s]\n", peerURL)
	action.Printer().PrintChaincodes(response.Chaincodes)
	return nil
}
