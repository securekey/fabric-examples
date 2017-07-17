/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"os"

	chaincode "github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/chaincode"
	channel "github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/channel"
	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/common"
	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/event"
	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/query"
	"github.com/spf13/cobra"
)

func newFabricCLICmd() *cobra.Command {
	mainCmd := &cobra.Command{
		Use: "fabric-cli",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			cmd.HelpFunc()(cmd, args)
		},
	}

	flags := mainCmd.PersistentFlags()
	common.Config().InitLoggingLevel(flags)
	common.Config().InitUserName(flags)
	common.Config().InitUserPassword(flags)
	common.Config().InitConfigFile(flags)
	common.Config().InitOrdererTLSCertificate(flags)
	common.Config().InitPrintFormat(flags)
	common.Config().InitOrgIDs(flags)

	mainCmd.AddCommand(chaincode.Cmd())
	mainCmd.AddCommand(query.Cmd())
	mainCmd.AddCommand(channel.Cmd())
	mainCmd.AddCommand(event.Cmd())

	return mainCmd
}

func main() {
	if newFabricCLICmd().Execute() != nil {
		os.Exit(1)
	}
}
