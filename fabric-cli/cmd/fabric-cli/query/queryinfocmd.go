/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package query

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/action"
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

		if cliconfig.Config().ChannelID() == "" {
			fmt.Printf("\nMust specify channel ID\n\n")
			cmd.HelpFunc()(cmd, args)
			return
		}

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
	action.Action
}

func newQueryInfoAction(flags *pflag.FlagSet) (*queryInfoAction, error) {
	action := &queryInfoAction{}
	err := action.Initialize(flags)
	return action, err
}

func (a *queryInfoAction) run() error {
	channelClient, err := a.LedgerClient()
	if err != nil {
		return errors.Errorf("Error getting admin ledger client: %v", err)
	}

	info, err := channelClient.QueryInfo()
	if err != nil {
		return err
	}

	a.Printer().PrintBlockchainInfo(info.BCI)

	return nil
}
