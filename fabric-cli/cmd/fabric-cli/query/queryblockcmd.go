/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package query

import (
	"encoding/base64"
	"strings"

	fabricCommon "github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	"github.com/pkg/errors"
	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/action"
	cliconfig "github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/config"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var queryBlockCmd = &cobra.Command{
	Use:   "block",
	Short: "Query block",
	Long:  "Queries a block",
	Run: func(cmd *cobra.Command, args []string) {
		action, err := newQueryBlockAction(cmd.Flags())
		if err != nil {
			cliconfig.Config().Logger().Errorf("Error while initializing queryBlockAction: %v", err)
			return
		}

		defer action.Terminate()

		err = action.invoke()
		if err != nil {
			cliconfig.Config().Logger().Errorf("Error while running queryBlockAction: %v", err)
		}
	},
}

func getQueryBlockCmd() *cobra.Command {
	flags := queryBlockCmd.Flags()
	cliconfig.InitChannelID(flags)
	cliconfig.InitBlockNum(flags)
	cliconfig.InitBlockHash(flags)
	cliconfig.InitTraverse(flags)
	cliconfig.InitPeerURL(flags, "", "The URL of the peer on which to install the chaincode, e.g. grpcs://localhost:7051")
	return queryBlockCmd
}

type queryBlockAction struct {
	action.Action
}

func newQueryBlockAction(flags *pflag.FlagSet) (*queryBlockAction, error) {
	action := &queryBlockAction{}
	err := action.Initialize(flags)
	return action, err
}

func (a *queryBlockAction) invoke() error {
	ledgerClient, err := a.LedgerClient()
	if err != nil {
		return errors.Errorf("Error getting admin channel client: %v", err)
	}

	var block *fabricCommon.Block
	if cliconfig.IsFlagSet(cliconfig.BlockNumFlag) {
		var err error
		block, err = ledgerClient.QueryBlock(cliconfig.Config().BlockNum())
		if err != nil {
			return err
		}
	} else if cliconfig.IsFlagSet(cliconfig.BlockHashFlag) {
		var err error

		hashBytes, err := Base64URLDecode(cliconfig.Config().BlockHash())
		if err != nil {
			return err
		}

		block, err = ledgerClient.QueryBlockByHash(hashBytes)
		if err != nil {
			return err
		}
	} else {
		return errors.Errorf("must specify either a block number of a block hash")
	}

	a.Printer().PrintBlock(block)

	a.traverse(ledgerClient, block, cliconfig.Config().Traverse()-1)

	return nil
}

func (a *queryBlockAction) traverse(ledgerClient *ledger.Client, currentBlock *fabricCommon.Block, num int) error {
	if num <= 0 {
		return nil
	}

	block, err := ledgerClient.QueryBlockByHash(currentBlock.Header.PreviousHash)
	if err != nil {
		return err
	}

	a.Printer().PrintBlock(block)

	if block.Header.PreviousHash != nil {
		return a.traverse(ledgerClient, block, num-1)
	}
	return nil
}

// Base64URLDecode decodes the base64 string into a byte array
func Base64URLDecode(data string) ([]byte, error) {
	//check if it has padding or not
	if strings.HasSuffix(data, "=") {
		return base64.URLEncoding.DecodeString(data)
	}
	return base64.RawURLEncoding.DecodeString(data)
}
