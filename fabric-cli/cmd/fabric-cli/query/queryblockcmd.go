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

package query

import (
	"fmt"

	"github.com/hyperledger/fabric-sdk-go/api"
	fabricCommon "github.com/hyperledger/fabric/protos/common"
	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/common"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var queryBlockCmd = &cobra.Command{
	Use:   "block",
	Short: "Query block",
	Long:  "Queries a block",
	Run: func(cmd *cobra.Command, args []string) {
		if common.Config().BlockNum() < 0 && common.Config().BlockHash() == "" {
			fmt.Printf("\nMust specify either the block number or the block hash\n\n")
			cmd.HelpFunc()(cmd, args)
			return
		}
		action, err := newQueryBlockAction(cmd.Flags())
		if err != nil {
			common.Config().Logger().Criticalf("Error while initializing queryBlockAction: %v", err)
			return
		}

		err = action.invoke()
		if err != nil {
			common.Config().Logger().Criticalf("Error while running queryBlockAction: %v", err)
			return
		}
	},
}

// getQueryBlockCmd returns the Query block action command
func getQueryBlockCmd() *cobra.Command {
	flags := queryBlockCmd.Flags()
	common.Config().InitChannelID(flags)
	common.Config().InitBlockNum(flags)
	common.Config().InitBlockHash(flags)
	common.Config().InitTraverse(flags)
	common.Config().InitPeerURL(flags, "", "The URL of the peer on which to install the chaincode, e.g. localhost:7051")
	return queryBlockCmd
}

type queryBlockAction struct {
	common.ActionImpl
}

func newQueryBlockAction(flags *pflag.FlagSet) (*queryBlockAction, error) {
	action := &queryBlockAction{}
	err := action.Initialize(flags)
	return action, err
}

func (action *queryBlockAction) invoke() error {
	chain, err := action.NewChannel()
	if err != nil {
		return fmt.Errorf("Error initializing chain: %v", err)
	}

	var block *fabricCommon.Block
	if common.Config().BlockNum() >= 0 {
		var err error
		block, err = chain.QueryBlock(common.Config().BlockNum())
		if err != nil {
			return err
		}
	} else if common.Config().BlockHash() != "" {
		var err error

		hashBytes, err := common.Base64URLDecode(common.Config().BlockHash())
		if err != nil {
			return err
		}

		block, err = chain.QueryBlockByHash(hashBytes)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("must specify either a block number of a block hash")
	}

	action.Printer().PrintBlock(block)

	action.traverse(chain, block, common.Config().Traverse()-1)

	return nil
}

func (action *queryBlockAction) traverse(chain api.Channel, currentBlock *fabricCommon.Block, num int) error {
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
