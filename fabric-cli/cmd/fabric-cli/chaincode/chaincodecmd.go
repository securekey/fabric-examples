/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chaincode

import (
	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/common"
	"github.com/spf13/cobra"
)

var chaincodeCmd = &cobra.Command{
	Use:   "chaincode",
	Short: "Chaincode commands",
	Long:  "Chaincode commands",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.HelpFunc()(cmd, args)
	},
}

// Cmd returns the chaincode command
func Cmd() *cobra.Command {
	common.Config().InitChannelID(chaincodeCmd.Flags())

	chaincodeCmd.AddCommand(getInstallCmd())
	chaincodeCmd.AddCommand(getInstantiateCmd())
	chaincodeCmd.AddCommand(getInvokeCmd())
	chaincodeCmd.AddCommand(getQueryCmd())
	chaincodeCmd.AddCommand(getGetInfoCmd())

	return chaincodeCmd
}
