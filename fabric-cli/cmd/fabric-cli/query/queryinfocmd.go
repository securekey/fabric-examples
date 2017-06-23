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

var queryInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Query info",
	Long:  "Queries general info",
	Run: func(cmd *cobra.Command, args []string) {
		action, err := newQueryInfoAction(cmd.Flags())
		if err != nil {
			common.Config().Logger().Criticalf("Error while initializing queryInfoAction: %v", err)
			return
		}

		err = action.run()
		if err != nil {
			common.Config().Logger().Criticalf("Error while running queryInfoAction: %v", err)
			return
		}
	},
}

// getQueryInfoCmd returns the Query block action command
func getQueryInfoCmd() *cobra.Command {
	flags := queryInfoCmd.Flags()
	common.Config().InitTxID(flags)
	common.Config().InitChannelID(flags)
	common.Config().InitPeerURL(flags)
	return queryInfoCmd
}

type queryInfoAction struct {
	common.ActionImpl
}

func newQueryInfoAction(flags *pflag.FlagSet) (*queryInfoAction, error) {
	action := &queryInfoAction{}
	err := action.Initialize(flags)
	return action, err
}

func (action *queryInfoAction) run() error {
	chain, err := action.NewChannel()
	if err != nil {
		return fmt.Errorf("Error initializing chain: %v", err)
	}

	info, err := chain.QueryInfo()
	if err != nil {
		return err
	}

	action.Printer().PrintBlockchainInfo(info)

	return nil
}
