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
	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/common"
	"github.com/spf13/cobra"
)

var queryCmd = &cobra.Command{
	Use:   "query",
	Short: "Query commands",
	Long:  "Query commands",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.HelpFunc()(cmd, args)
	},
}

// Cmd returns the query command
func Cmd() *cobra.Command {
	common.Config().InitChannelID(queryCmd.Flags())

	queryCmd.AddCommand(getQueryBlockCmd())
	queryCmd.AddCommand(getQueryInfoCmd())
	queryCmd.AddCommand(getQueryTXCmd())
	queryCmd.AddCommand(getQueryChannelsCmd())
	queryCmd.AddCommand(getQueryInstalledCmd())

	return queryCmd
}
