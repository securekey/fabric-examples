/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package channel

import (
	"fmt"
	"io/ioutil"

	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	fabricCommon "github.com/hyperledger/fabric/protos/common"
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

		defer action.Terminate()

		err = action.invoke()
		if err != nil {
			common.Config().Logger().Criticalf("Error while running channelCreateAction: %v", err)
		}
	},
}

func getChannelCreateCmd() *cobra.Command {
	flags := channelCreateCmd.Flags()
	common.Config().InitChannelID(flags)
	common.Config().InitOrdererURL(flags)
	common.Config().InitTxFile(flags)
	return channelCreateCmd
}

type channelCreateAction struct {
	common.Action
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

	// Sign the config with the org admin user
	context := action.SetUserContext(action.OrgAdminUser(common.Config().OrgID()))
	defer context.Restore()

	txID, err := action.Client().NewTxnID()
	if err != nil {
		return fmt.Errorf("Error creating transaction ID: %v", err)
	}

	configSignature, err := action.Client().SignChannelConfig(config)
	if err != nil {
		return fmt.Errorf("Error signing configuration: %v", err)
	}

	orderer, err := action.RandomOrderer()
	if err != nil {
		return err
	}

	// Use the Orderer Admin user to create the channel
	action.SetUserContext(action.OrgOrdererAdminUser(common.Config().OrgID()))

	fmt.Printf("Attempting to create channel: %s\n", common.Config().ChannelID())

	_, err = action.Client().CreateChannel(apifabclient.CreateChannelRequest{
		Name:       common.Config().ChannelID(),
		Orderer:    orderer,
		Config:     config,
		Signatures: []*fabricCommon.ConfigSignature{configSignature},
		TxnID:      txID,
	})
	if err != nil {
		return fmt.Errorf("Error from create channel: %s", err.Error())
	}

	fmt.Println("Channel created!")

	return nil
}
