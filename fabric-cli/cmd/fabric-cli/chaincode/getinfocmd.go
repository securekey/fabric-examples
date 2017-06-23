/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chaincode

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric/core/common/ccprovider"
	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/common"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	lifecycleSCC = "lscc"
)

var getInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Get chaincode info",
	Long:  "Retrieves details about the chaincode",
	Run: func(cmd *cobra.Command, args []string) {
		if common.Config().ChaincodeID() == "" {
			fmt.Printf("\nMust specify the chaincode ID\n\n")
			cmd.HelpFunc()(cmd, args)
			return
		}
		action, err := newGetInfoAction(cmd.Flags())
		if err != nil {
			common.Config().Logger().Criticalf("Error while initializing getAction: %v", err)
			return
		}

		err = action.invoke()
		if err != nil {
			common.Config().Logger().Criticalf("Error while running getAction: %v", err)
			return
		}
	},
}

// Cmd returns the install command
func getGetInfoCmd() *cobra.Command {
	flags := getInfoCmd.Flags()
	common.Config().InitPeerURL(flags)
	common.Config().InitChannelID(flags)
	common.Config().InitChaincodeID(flags)
	return getInfoCmd
}

type getInfoAction struct {
	common.ActionImpl
}

func newGetInfoAction(flags *pflag.FlagSet) (*getInfoAction, error) {
	action := &getInfoAction{}
	err := action.Initialize(flags)
	if len(action.Peers()) == 0 {
		return nil, fmt.Errorf("a peer must be specified")
	}
	return action, err
}

func (action *getInfoAction) invoke() error {
	channel, err := action.NewChannel()
	if err != nil {
		return fmt.Errorf("Error initializing channel: %v", err)
	}

	var args []string
	args = append(args, "getccdata")
	args = append(args, common.Config().ChannelID())
	args = append(args, common.Config().ChaincodeID())

	cdbytes, err := common.QueryChaincode(channel, action.Peers(), lifecycleSCC, common.Config().ChannelID(), args)
	if err != nil {
		return fmt.Errorf("Error instantiating chaincode: %v", err)
	}

	ccData := &ccprovider.ChaincodeData{}
	err = proto.Unmarshal(cdbytes, ccData)
	if err != nil {
		return fmt.Errorf("Error unmarshalling chaincode data: %v", err)
	}

	action.Printer().PrintChaincodeData(ccData)

	return nil
}
