/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package query

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	fabricCommon "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/common"
	cliconfig "github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/config"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var queryBlockCmd = &cobra.Command{
	Use:   "block",
	Short: "Query block",
	Long:  "Queries a block",
	Run: func(cmd *cobra.Command, args []string) {
		if cliconfig.Config().BlockNum() < 0 && cliconfig.Config().BlockHash() == "" {
			fmt.Printf("\nMust specify either the block number or the block hash\n\n")
			cmd.HelpFunc()(cmd, args)
			return
		}
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
	cliconfig.Config().InitChannelID(flags)
	cliconfig.Config().InitBlockNum(flags)
	cliconfig.Config().InitBlockHash(flags)
	cliconfig.Config().InitTraverse(flags)
	cliconfig.Config().InitPeerURL(flags, "", "The URL of the peer on which to install the chaincode, e.g. grpcs://localhost:7051")
	return queryBlockCmd
}

type queryBlockAction struct {
	common.Action
}

func newQueryBlockAction(flags *pflag.FlagSet) (*queryBlockAction, error) {
	action := &queryBlockAction{}
	err := action.Initialize(flags)
	return action, err
}

func (action *queryBlockAction) invoke() error {
	channelClient, err := action.AdminChannelClient()
	if err != nil {
		return fmt.Errorf("Error getting admin channel client: %v", err)
	}

	var block *fabricCommon.Block
	if cliconfig.Config().BlockNum() >= 0 {
		var err error
		block, err = channelClient.QueryBlock(cliconfig.Config().BlockNum())
		if err != nil {
			return err
		}
	} else if cliconfig.Config().BlockHash() != "" {
		var err error

		hashBytes, err := Base64URLDecode(cliconfig.Config().BlockHash())
		if err != nil {
			return err
		}

		block, err = channelClient.QueryBlockByHash(hashBytes)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("must specify either a block number of a block hash")
	}

	action.Printer().PrintBlock(block)

	action.traverse(channelClient, block, cliconfig.Config().Traverse()-1)

	return nil
}

func (action *queryBlockAction) traverse(chain apifabclient.Channel, currentBlock *fabricCommon.Block, num int) error {
	if num <= 0 {
		return nil
	}

	block, err := chain.QueryBlockByHash(currentBlock.Header.PreviousHash)
	if err != nil {
		return err
	}

	action.Printer().PrintBlock(block)

	if block.Header.PreviousHash != nil {
		return action.traverse(chain, block, num-1)
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
