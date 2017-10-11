/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"os"

	chaincode "github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/chaincode"
	channel "github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/channel"
	cliconfig "github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/config"
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
	cliconfig.Config().InitLoggingLevel(flags)
	cliconfig.Config().InitUserName(flags)
	cliconfig.Config().InitUserPassword(flags)
	cliconfig.Config().InitOrdererTLSCertificate(flags)
	cliconfig.Config().InitPrintFormat(flags)
	cliconfig.Config().InitWriter(flags)
	cliconfig.Config().InitOrgIDs(flags)

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
