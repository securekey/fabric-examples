/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package channel

import (
	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/common"
	"github.com/spf13/cobra"
)

var channelCmd = &cobra.Command{
	Use:   "channel",
	Short: "Channel commands",
	Long:  "Channel commands",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.HelpFunc()(cmd, args)
	},
}

// Cmd returns the channel command
func Cmd() *cobra.Command {
	common.Config().InitChannelID(channelCmd.Flags())

	channelCmd.AddCommand(getChannelCreateCmd())
	channelCmd.AddCommand(getChannelJoinCmd())

	return channelCmd
}
