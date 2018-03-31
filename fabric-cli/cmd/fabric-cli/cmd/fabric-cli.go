/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cmd

import (
	"os"

	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/chaincode"
	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/channel"
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
	cliconfig.InitConfigFile(flags)
	cliconfig.InitLoggingLevel(flags)
	cliconfig.InitUserName(flags)
	cliconfig.InitUserPassword(flags)
	cliconfig.InitOrdererTLSCertificate(flags)
	cliconfig.InitPrintFormat(flags)
	cliconfig.InitWriter(flags)
	cliconfig.InitBase64(flags)
	cliconfig.InitOrgIDs(flags)

	mainCmd.AddCommand(chaincode.Cmd())
	mainCmd.AddCommand(query.Cmd())
	mainCmd.AddCommand(channel.Cmd())
	mainCmd.AddCommand(event.Cmd())

	return mainCmd
}

func Execute() {
	if newFabricCLICmd().Execute() != nil {
		os.Exit(1)
	}
}
