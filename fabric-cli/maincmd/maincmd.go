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

package maincmd

import (
	chaincode "github.com/securekey/fabric-examples/fabric-cli/chaincode"
	channel "github.com/securekey/fabric-examples/fabric-cli/channel"
	"github.com/securekey/fabric-examples/fabric-cli/common"
	"github.com/securekey/fabric-examples/fabric-cli/event"
	"github.com/securekey/fabric-examples/fabric-cli/query"
	"github.com/spf13/cobra"
)

const (
	loggingLevelFlag    = "logging-level"
	userFlag            = "user"
	passwordFlag        = "pw"
	configFileFlag      = "config"
	certificateFileFlag = "cacert"
	outputFormatFlag    = "format"
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

	mainFlags := mainCmd.PersistentFlags()
	mainFlags.StringVar(&common.LoggingLevel, loggingLevelFlag, "CRITICAL", "Logging level - CRITICAL, ERROR, WARNING, INFO, DEBUG")
	mainFlags.StringVar(&common.User, userFlag, common.User, "The enrollment user")
	mainFlags.StringVar(&common.Password, passwordFlag, common.Password, "The password of the enrollment user")
	mainFlags.StringVar(&common.ConfigFile, configFileFlag, common.ConfigFile, "The path of the config.yaml file")
	mainFlags.StringVar(&common.Certificate, certificateFileFlag, common.Certificate, "The path of the ca-cert.pem file")
	mainFlags.StringVar(&common.PrintFormat, outputFormatFlag, common.DISPLAY.String(), "The output format - display, json, raw")

	mainCmd.AddCommand(chaincode.Cmd())
	mainCmd.AddCommand(query.Cmd())
	mainCmd.AddCommand(channel.Cmd())
	mainCmd.AddCommand(event.Cmd())

	return mainCmd
}
