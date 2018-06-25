/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chaincode

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/core/common/ccprovider"
	"github.com/pkg/errors"
	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/action"
	cliconfig "github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/config"
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
	cliconfig.InitPeerURL(flags)
	cliconfig.InitChannelID(flags)
	cliconfig.InitChaincodeID(flags)
	return getInfoCmd
}

type getInfoAction struct {
	action.Action
}

func newGetInfoAction(flags *pflag.FlagSet) (*getInfoAction, error) {
	action := &getInfoAction{}
	err := action.Initialize(flags)
	if len(action.Peers()) == 0 {
		return nil, errors.New("a peer must be specified")
	}
	return action, err
}

func (action *getInfoAction) invoke() error {
	peer := action.Peer()
	channelClient, err := action.ChannelClient()
	if err != nil {
		return errors.Errorf("error retrieving channel client: %v", err)
	}

	var args [][]byte
	args = append(args, []byte(cliconfig.Config().ChannelID()))
	args = append(args, []byte(cliconfig.Config().ChaincodeID()))

	fmt.Printf("querying chaincode info for %s on peer: %s...\n", cliconfig.Config().ChaincodeID(), peer.URL())

	response, err := channelClient.Query(
		channel.Request{ChaincodeID: lifecycleSCC, Fcn: "getccdata", Args: args},
		channel.WithTargetEndpoints(peer.URL()))
	if err != nil {
		return errors.Errorf("error querying for chaincode info: %v", err)
	}

	ccData := &ccprovider.ChaincodeData{}
	err = proto.Unmarshal(response.Payload, ccData)
	if err != nil {
		return errors.Errorf("error unmarshalling chaincode data: %v", err)
	}

	action.Printer().PrintChaincodeData(ccData)

	return nil
}
