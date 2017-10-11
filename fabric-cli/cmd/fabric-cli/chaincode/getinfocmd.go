/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chaincode

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/common"
	cliconfig "github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/config"
	"github.com/securekey/fabric-examples/fabric-cli/internal/github.com/hyperledger/fabric/core/common/ccprovider"
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
		if cliconfig.Config().ChaincodeID() == "" {
			fmt.Printf("\nMust specify the chaincode ID\n\n")
			cmd.HelpFunc()(cmd, args)
			return
		}
		action, err := newGetInfoAction(cmd.Flags())
		if err != nil {
			cliconfig.Config().Logger().Errorf("Error while initializing getAction: %v", err)
			return
		}

		defer action.Terminate()

		err = action.invoke()
		if err != nil {
			cliconfig.Config().Logger().Errorf("Error while running getAction: %v", err)
		}
	},
}

func getGetInfoCmd() *cobra.Command {
	flags := getInfoCmd.Flags()
	cliconfig.Config().InitPeerURL(flags)
	cliconfig.Config().InitChannelID(flags)
	cliconfig.Config().InitChaincodeID(flags)
	return getInfoCmd
}

type getInfoAction struct {
	common.Action
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
	channelClient, err := action.ChannelClient()
	if err != nil {
		return fmt.Errorf("Error retrieving channel client: %v", err)
	}

	var args [][]byte
	args = append(args, []byte(cliconfig.Config().ChannelID()))
	args = append(args, []byte(cliconfig.Config().ChaincodeID()))

	peer := action.Peer()

	fmt.Printf("querying chaincode info for %s on peer: %s...\n", cliconfig.Config().ChaincodeID(), peer.URL())

	response, err := channelClient.QueryWithOpts(
		apitxn.QueryRequest{
			Fcn:         "getccdata",
			Args:        args,
			ChaincodeID: lifecycleSCC,
		},
		apitxn.QueryOpts{
			ProposalProcessors: []apitxn.ProposalProcessor{peer},
		},
	)
	if err != nil {
		return fmt.Errorf("Error querying for chaincode info: %v", err)
	}

	ccData := &ccprovider.ChaincodeData{}
	err = proto.Unmarshal(response, ccData)
	if err != nil {
		return fmt.Errorf("Error unmarshalling chaincode data: %v", err)
	}

	action.Printer().PrintChaincodeData(ccData)

	return nil
}
