/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package channel

import (
	"fmt"
	"io/ioutil"

	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	fabricCommon "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/common"
	cliconfig "github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/config"
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
			cliconfig.Config().Logger().Errorf("Error while initializing channelCreateAction: %v", err)
			return
		}

		defer action.Terminate()

		err = action.invoke()
		if err != nil {
			cliconfig.Config().Logger().Errorf("Error while running channelCreateAction: %v", err)
		}
	},
}

func getChannelCreateCmd() *cobra.Command {
	flags := channelCreateCmd.Flags()
	cliconfig.Config().InitChannelID(flags)
	cliconfig.Config().InitOrdererURL(flags)
	cliconfig.Config().InitTxFile(flags)
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
	configTx, err := ioutil.ReadFile(cliconfig.Config().TxFile())
	if err != nil {
		return fmt.Errorf("An error occurred while reading TX file %s: %v", cliconfig.Config().TxFile(), err)
	}

	user, err := action.OrgAdminUser(action.OrgID())
	if err != nil {
		return err
	}

	adminFabClient, err := action.ClientForUser(action.OrgID(), user)
	if err != nil {
		return fmt.Errorf("error getting fabric client: %s", err)
	}

	config, err := adminFabClient.ExtractChannelConfig(configTx)
	if err != nil {
		return fmt.Errorf("error extracting channel config: %v", err)
	}

	txID, err := adminFabClient.NewTxnID()
	if err != nil {
		return fmt.Errorf("Error creating transaction ID: %v", err)
	}

	configSignature, err := adminFabClient.SignChannelConfig(config)
	if err != nil {
		return fmt.Errorf("Error signing configuration: %v", err)
	}

	orderer, err := action.RandomOrderer()
	if err != nil {
		return err
	}

	// Use the Orderer Admin user to create the channel
	ordererAdminUser, err := action.OrdererAdminUser()
	if err != nil {
		return fmt.Errorf("error getting orderer admin user: %s", err)
	}

	ordererAdminFabClient, err := action.ClientForUser(action.OrgID(), ordererAdminUser)
	if err != nil {
		return fmt.Errorf("error getting fabric client: %s", err)
	}

	fmt.Printf("Attempting to create channel: %s\n", cliconfig.Config().ChannelID())

	_, err = ordererAdminFabClient.CreateChannel(apifabclient.CreateChannelRequest{
		Name:       cliconfig.Config().ChannelID(),
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
