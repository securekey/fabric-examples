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

package channel

import (
	"fmt"
	"io/ioutil"

	"github.com/hyperledger/fabric-sdk-go/config"
	fabricClient "github.com/hyperledger/fabric-sdk-go/fabric-client"
	"github.com/securekey/fabric-examples/fabric-cli/common"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	txFileFlag     = "txfile"
	defaultTxFile  = "fixtures/channel/testchannel.tx"
	defaultOrderer = "localhost:7050"
)

var txFile string

var channelCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create Channel",
	Long:  "Create a new channel",
	Run: func(cmd *cobra.Command, args []string) {
		action, err := newChannelCreateAction(cmd.Flags())
		if err != nil {
			common.Logger.Criticalf("Error while initializing channelCreateAction: %v", err)
			return
		}

		err = action.invoke()
		if err != nil {
			common.Logger.Criticalf("Error while running channelCreateAction: %v", err)
			return
		}
	},
}

// ChainCreateCmd returns the chainCreateAction command
func getChannelCreateCmd() *cobra.Command {
	flags := channelCreateCmd.Flags()
	flags.StringVar(&common.ChannelID, common.ChannelIDFlag, common.ChannelID, "The channel ID")
	flags.StringVar(&txFile, txFileFlag, defaultTxFile, "The path of the channel.tx file")
	flags.StringVar(&common.OrdererURL, common.OrdererFlag, defaultOrderer, "The URL of the orderer, e.g. localhost:7050")
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
	configTx, err := ioutil.ReadFile(txFile)
	if err != nil {
		return fmt.Errorf("An error occurred while reading TX file %s: %v", txFile, err)
	}

	certificate := config.GetOrdererTLSCertificate()
	serverHostOverride := "orderer0"

	orderer, err := fabricClient.NewOrderer(common.OrdererURL, certificate, serverHostOverride)
	if err != nil {
		return fmt.Errorf("CreateNewOrderer return error: %v", err)
	}

	fmt.Printf("Attempting to create channel: %s\n", common.ChannelID)

	chain, err := action.Client().CreateChannel(&fabricClient.CreateChannelRequest{
		Envelope: configTx,
		Orderer:  orderer,
		Name:     common.ChannelID,
	})
	if err != nil {
		return fmt.Errorf("Error from create channel: %s", err.Error())
	}

	if chain != nil {
		fmt.Println("Channel created!")
	}

	return nil
}
