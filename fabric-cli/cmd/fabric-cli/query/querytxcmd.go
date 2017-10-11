/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package query

import (
	"fmt"

	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/common"
	cliconfig "github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/config"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var queryTXCmd = &cobra.Command{
	Use:   "tx",
	Short: "Query transaction",
	Long:  "Queries a transaction",
	Run: func(cmd *cobra.Command, args []string) {
		if cliconfig.Config().TxID() == "" {
			fmt.Printf("\nMust specify the transaction ID\n\n")
			cmd.HelpFunc()(cmd, args)
			return
		}
		action, err := newQueryTXAction(cmd.Flags())
		if err != nil {
			cliconfig.Config().Logger().Errorf("Error while initializing queryTXAction: %v", err)
			return
		}

		defer action.Terminate()

		err = action.run()
		if err != nil {
			cliconfig.Config().Logger().Errorf("Error while running queryTXAction: %v", err)
		}
	},
}

func getQueryTXCmd() *cobra.Command {
	flags := queryTXCmd.Flags()
	cliconfig.Config().InitChannelID(flags)
	cliconfig.Config().InitTxID(flags)
	cliconfig.Config().InitPeerURL(flags)
	return queryTXCmd
}

type queryTXAction struct {
	common.Action
}

func newQueryTXAction(flags *pflag.FlagSet) (*queryTXAction, error) {
	action := &queryTXAction{}
	err := action.Initialize(flags)

	return action, err
}

func (action *queryTXAction) run() error {
	channelClient, err := action.AdminChannelClient()
	if err != nil {
		return fmt.Errorf("Error getting admin channel client: %v", err)
	}

	tx, err := channelClient.QueryTransaction(cliconfig.Config().TxID())
	if err != nil {
		return err
	}

	fmt.Printf("Transaction %s in channel %s\n", cliconfig.Config().TxID(), cliconfig.Config().ChannelID())
	action.Printer().PrintProcessedTransaction(tx)

	return nil
}
