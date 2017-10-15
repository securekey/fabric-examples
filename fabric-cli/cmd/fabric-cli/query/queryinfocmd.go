/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package query

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/common"
	cliconfig "github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/config"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var queryInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Query info",
	Long:  "Queries general info",
	Run: func(cmd *cobra.Command, args []string) {
		action, err := newQueryInfoAction(cmd.Flags())
		if err != nil {
			cliconfig.Config().Logger().Errorf("Error while initializing queryInfoAction: %v", err)
			return
		}

		defer action.Terminate()

		err = action.run()
		if err != nil {
			cliconfig.Config().Logger().Errorf("Error while running queryInfoAction: %v", err)
		}
	},
}

func getQueryInfoCmd() *cobra.Command {
	flags := queryInfoCmd.Flags()
	cliconfig.InitTxID(flags)
	cliconfig.InitChannelID(flags)
	cliconfig.InitPeerURL(flags)
	return queryInfoCmd
}

type queryInfoAction struct {
	common.Action
}

func newQueryInfoAction(flags *pflag.FlagSet) (*queryInfoAction, error) {
	action := &queryInfoAction{}
	err := action.Initialize(flags)
	return action, err
}

func (action *queryInfoAction) run() error {
	channelClient, err := action.AdminChannelClient()
	if err != nil {
		return errors.Errorf("Error getting admin channel client: %v", err)
	}

	info, err := channelClient.QueryInfo()
	if err != nil {
		return err
	}

	action.Printer().PrintBlockchainInfo(info)

	return nil
}
