/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package channel

import (
	"fmt"
	"io/ioutil"

	"github.com/hyperledger/fabric-sdk-go/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/orderer"
	"github.com/hyperledger/fabric/common/crypto"
	fabricCommon "github.com/hyperledger/fabric/protos/common"
	protos_utils "github.com/hyperledger/fabric/protos/utils"
	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/common"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var channelCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create Channel",
	Long:  "Create a new channel",
	Run: func(cmd *cobra.Command, args []string) {
		action, err := newChannelCreateAction(cmd.Flags())
		if err != nil {
			common.Config().Logger().Criticalf("Error while initializing channelCreateAction: %v", err)
			return
		}

		err = action.invoke()
		if err != nil {
			common.Config().Logger().Criticalf("Error while running channelCreateAction: %v", err)
			return
		}
	},
}

// ChainCreateCmd returns the chainCreateAction command
func getChannelCreateCmd() *cobra.Command {
	flags := channelCreateCmd.Flags()
	common.Config().InitChannelID(flags)
	common.Config().InitOrdererURL(flags)
	common.Config().InitTxFile(flags)
	return channelCreateCmd
}

type channelCreateAction struct {
	common.ActionImpl
}

func newChannelCreateAction(flags *pflag.FlagSet) (*channelCreateAction, error) {
	action := &channelCreateAction{}
	err := action.Initialize(flags)
	return action, err
}

func (action *channelCreateAction) invoke() error {
	configTx, err := ioutil.ReadFile(common.Config().TxFile())
	if err != nil {
		return fmt.Errorf("An error occurred while reading TX file %s: %v", common.Config().TxFile(), err)
	}

	config, err := action.Client().ExtractChannelConfig(configTx)
	if err != nil {
		return fmt.Errorf("Error extracting channel config: %v", err)
	}

	configSignature, err := action.Client().SignChannelConfig(config)
	if err != nil {
		return fmt.Errorf("Error signing configuration: %v", err)
	}

	var configSignatures []*fabricCommon.ConfigSignature
	configSignatures = append(configSignatures, configSignature)

	creator, err := action.Client().GetIdentity()
	if err != nil {
		return fmt.Errorf("Error getting creator: %v", err)
	}
	nonce, err := crypto.GetRandomNonce()
	if err != nil {
		return fmt.Errorf("Could not compute nonce: %s", err)
	}
	txID, err := protos_utils.ComputeProposalTxID(nonce, creator)
	if err != nil {
		return fmt.Errorf("Could not compute TxID: %s", err)
	}

	ordererAdmin, err := action.GetOrdererAdminUser()
	if err != nil {
		return fmt.Errorf("Error getting orderer admin user: %v", err)
	}
	action.Client().SetUserContext(ordererAdmin)

	orderer, err := orderer.NewOrderer(
		fmt.Sprintf("%s:%s", common.Config().GetOrdererHost(), common.Config().GetOrdererPort()),
		common.Config().GetOrdererTLSCertificate(), common.Config().GetOrdererTLSServerHostOverride(), common.Config())
	if err != nil {
		return fmt.Errorf("CreateNewOrderer return error: %v", err)
	}

	fmt.Printf("Attempting to create channel: %s\n", common.Config().ChannelID())

	err = action.Client().CreateChannel(&api.CreateChannelRequest{
		Name:       common.Config().ChannelID(),
		Orderer:    orderer,
		Config:     config,
		Signatures: configSignatures,
		TxID:       txID,
		Nonce:      nonce,
	})
	if err != nil {
		return fmt.Errorf("Error from create channel: %s", err.Error())
	}

	fmt.Println("Channel created!")

	return nil
}
