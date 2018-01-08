/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package channel

import (
	"fmt"

	chmgmt "github.com/hyperledger/fabric-sdk-go/api/apitxn/chmgmtclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
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
	action := &channelCreateAction{}
	err := action.Initialize(flags)
	return action, err
}

func (a *channelCreateAction) invoke() error {
	user, err := a.OrgAdminUser(a.OrgID())
	if err != nil {
		return err
	}

	chMgmtClient, err := a.ChannelMgmtClient()
	if err != nil {
		return err
	}

	fmt.Printf("Attempting to create channel: %s\n", cliconfig.Config().ChannelID())

	req := chmgmt.SaveChannelRequest{
		ChannelID:     cliconfig.Config().ChannelID(),
		ChannelConfig: cliconfig.Config().TxFile(),
		SigningUser:   user,
	}

	if err := chMgmtClient.SaveChannel(req); err != nil {
		return errors.Errorf("Error from create channel: %s", err.Error())
	}

	fmt.Println("Channel created!")

	return nil
}
