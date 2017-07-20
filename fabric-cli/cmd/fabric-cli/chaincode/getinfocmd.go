/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chaincode

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
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

		defer action.Terminate()

		err = action.invoke()
		if err != nil {
			common.Config().Logger().Criticalf("Error while running getAction: %v", err)
		}
	},
}

func getGetInfoCmd() *cobra.Command {
	flags := getInfoCmd.Flags()
	common.Config().InitPeerURL(flags)
	common.Config().InitChannelID(flags)
	common.Config().InitChaincodeID(flags)
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
	channel, err := action.ChannelClient()
	if err != nil {
		return fmt.Errorf("Error retrieving channel client: %v", err)
	}

	var args []string
	args = append(args, common.Config().ChannelID())
	args = append(args, common.Config().ChaincodeID())

	peer := action.Peers()[0]
	orgID, err := action.OrgOfPeer(peer.URL())
	if err != nil {
		return err
	}

	context := action.SetUserContext(action.OrgUser(orgID))
	defer context.Restore()

	fmt.Printf("querying chaincode chaincode info for %s on peer: %s...\n", common.Config().ChaincodeID(), peer.URL())

	responses, _, err := channel.SendTransactionProposal(apitxn.ChaincodeInvokeRequest{
		Targets:     []apitxn.ProposalProcessor{peer},
		Fcn:         "getccdata",
		Args:        args,
		ChaincodeID: lifecycleSCC,
	})
	if err != nil {
		return fmt.Errorf("Error querying for chaincode info: %v", err)
	}
	if len(responses) == 0 {
		return fmt.Errorf("Didn't receive a response from chaincode %s", lifecycleSCC)
	}

	response := responses[0]
	if response.Err != nil {
		return response.Err
	}

	ccData := &ccprovider.ChaincodeData{}
	err = proto.Unmarshal(response.ProposalResponse.Response.Payload, ccData)
	if err != nil {
		return fmt.Errorf("Error unmarshalling chaincode data: %v", err)
	}

	action.Printer().PrintChaincodeData(ccData)

	return nil
}
