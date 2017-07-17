/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package query

import (
	"fmt"

	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/common"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var queryTXCmd = &cobra.Command{
	Use:   "tx",
	Short: "Query transaction",
	Long:  "Queries a transaction",
	Run: func(cmd *cobra.Command, args []string) {
		if common.Config().TxID() == "" {
			fmt.Printf("\nMust specify the transaction ID\n\n")
			cmd.HelpFunc()(cmd, args)
			return
		}
		action, err := newQueryTXAction(cmd.Flags())
		if err != nil {
			common.Config().Logger().Criticalf("Error while initializing queryTXAction: %v", err)
			return
		}

		defer action.Terminate()

		err = action.run()
		if err != nil {
			common.Config().Logger().Criticalf("Error while running queryTXAction: %v", err)
		}
	},
}

func getQueryTXCmd() *cobra.Command {
	flags := queryTXCmd.Flags()
	common.Config().InitChannelID(flags)
	common.Config().InitTxID(flags)
	common.Config().InitPeerURL(flags)
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
	channelClient, err := action.ChannelClient()
	if err != nil {
		return fmt.Errorf("Error getting channel client: %v", err)
	}

	context := action.SetUserContext(action.OrgAdminUser(common.Config().OrgID()))
	defer context.Restore()

	tx, err := channelClient.QueryTransaction(common.Config().TxID())
	if err != nil {
		return err
	}

	fmt.Printf("Transaction %s in chain %s\n", common.Config().TxID(), common.Config().ChannelID())
	action.Printer().PrintProcessedTransaction(tx)

	return nil
}
