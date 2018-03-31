/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package query

import (
	"fmt"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/pkg/errors"
	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/action"
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
	cliconfig.InitChannelID(flags)
	cliconfig.InitTxID(flags)
	cliconfig.InitPeerURL(flags)
	return queryTXCmd
}

type queryTXAction struct {
	action.Action
}

func newQueryTXAction(flags *pflag.FlagSet) (*queryTXAction, error) {
	action := &queryTXAction{}
	err := action.Initialize(flags)

	return action, err
}

func (a *queryTXAction) run() error {
	ledgerClient, err := a.LedgerClient()
	if err != nil {
		return errors.Errorf("Error getting ledger client: %v", err)
	}

	tx, err := ledgerClient.QueryTransaction(fab.TransactionID(cliconfig.Config().TxID()))
	if err != nil {
		return err
	}

	fmt.Printf("Transaction %s in channel %s\n", cliconfig.Config().TxID(), cliconfig.Config().ChannelID())
	a.Printer().PrintProcessedTransaction(tx)

	return nil
}
