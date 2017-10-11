/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package event

import (
	"fmt"

	fabricCommon "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/common"
	cliconfig "github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/config"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var listenBlockCmd = &cobra.Command{
	Use:   "listenblock",
	Short: "Listen to block events.",
	Long:  "Listen to block events",
	Run: func(cmd *cobra.Command, args []string) {
		action, err := newlistenBlockAction(cmd.Flags())
		if err != nil {
			cliconfig.Config().Logger().Errorf("Error while initializing listenBlockAction: %v", err)
			return
		}

		defer action.Terminate()

		err = action.invoke()
		if err != nil {
			cliconfig.Config().Logger().Errorf("Error while running listenBlockAction: %v", err)
		}
	},
}

func getListenBlockCmd() *cobra.Command {
	cliconfig.Config().InitPeerURL(listenBlockCmd.Flags(), "", "The URL of the peer on which to listen for events, e.g. grpcs://localhost:7051")
	return listenBlockCmd
}

type listenBlockAction struct {
	common.Action
	inputEvent
}

func newlistenBlockAction(flags *pflag.FlagSet) (*listenBlockAction, error) {
	action := &listenBlockAction{inputEvent: inputEvent{done: make(chan bool)}}
	err := action.Initialize(flags)
	return action, err
}

func (action *listenBlockAction) invoke() error {
	eventHub, err := action.EventHub()
	if err != nil {
		return err
	}

	fmt.Printf("Registering block event\n")

	callback := func(block *fabricCommon.Block) {
		action.Printer().PrintBlock(block)
		fmt.Println("Press <enter> to terminate")
	}

	eventHub.RegisterBlockEvent(callback)

	action.WaitForEnter()

	fmt.Printf("Unregistering block event\n")
	eventHub.UnregisterBlockEvent(callback)

	return nil
}
