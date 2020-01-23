/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package channel

import (
	"fmt"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/pkg/errors"
	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/action"
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
	cliconfig.InitChannelID(flags)
	cliconfig.InitOrdererURL(flags)
	cliconfig.InitTxFile(flags)
	return channelCreateCmd
}

type channelCreateAction struct {
	action.Action
}

func newChannelCreateAction(flags *pflag.FlagSet) (*channelCreateAction, error) {
	a := &channelCreateAction{}
	err := a.Initialize(flags)
	return a, err
}

func (a *channelCreateAction) invoke() error {
	user, err := a.OrgAdminUser(a.OrgID())
	if err != nil {
		return err
	}

	chMgmtClient, err := a.ResourceMgmtClient()
	if err != nil {
		return err
	}

	fmt.Printf("Attempting to create/update channel: %s\n", cliconfig.Config().ChannelID())

	req := resmgmt.SaveChannelRequest{
		ChannelID:         cliconfig.Config().ChannelID(),
		ChannelConfigPath: cliconfig.Config().TxFile(),
		SigningIdentities: []msp.SigningIdentity{user},
	}

	orderer, err := a.RandomOrderer()
	if err != nil {
		return err
	}

	_, err = chMgmtClient.SaveChannel(req, resmgmt.WithOrderer(orderer))
	if err != nil {
		return errors.Errorf("Error from save channel: %s", err.Error())
	}

	fmt.Printf("Channel created/updated: %s\n", cliconfig.Config().ChannelID())

	return nil
}
