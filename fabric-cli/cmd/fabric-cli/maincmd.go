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

package main

import (
	chaincode "github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/chaincode"
	channel "github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/channel"
	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/common"
	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/event"
	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/query"
	"github.com/spf13/cobra"
)

// NewFabricCLICmd returns the fabriccli command
func NewFabricCLICmd() *cobra.Command {
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
	common.Config().InitUser(flags)
	common.Config().InitPassword(flags)
	common.Config().InitConfigFile(flags)
	common.Config().InitCertificate(flags)
	common.Config().InitPrintFormat(flags)

	mainCmd.AddCommand(chaincode.Cmd())
	mainCmd.AddCommand(query.Cmd())
	mainCmd.AddCommand(channel.Cmd())
	mainCmd.AddCommand(event.Cmd())

	return mainCmd
}
