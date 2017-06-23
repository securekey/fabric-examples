/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
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
