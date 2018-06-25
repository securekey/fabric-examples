/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package event

import (
	"fmt"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/event"
	"github.com/pkg/errors"
	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/action"
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
	flags := listenBlockCmd.Flags()
	cliconfig.InitChannelID(flags)
	cliconfig.InitPeerURL(flags, "", "The URL of the peer on which to listen for events, e.g. localhost:7051")
	cliconfig.InitSeekType(flags)
	cliconfig.InitBlockNum(flags)
	return listenBlockCmd
}

type listenBlockAction struct {
	action.Action
	inputEvent
}

func newlistenBlockAction(flags *pflag.FlagSet) (*listenBlockAction, error) {
	action := &listenBlockAction{inputEvent: inputEvent{done: make(chan bool)}}
	err := action.Initialize(flags)
	return action, err
}

func (a *listenBlockAction) invoke() error {
	eventClient, err := a.EventClient(event.WithBlockEvents(), event.WithSeekType(cliconfig.Config().SeekType()), event.WithBlockNum(cliconfig.Config().BlockNum()))
	if err != nil {
		return err
	}

	fmt.Printf("Registering block event\n")

	breg, beventch, err := eventClient.RegisterBlockEvent()
	if err != nil {
		return errors.WithMessage(err, "Error registering for block events")
	}
	defer eventClient.Unregister(breg)

	enterch := a.WaitForEnter()
	for {
		select {
		case _, _ = <-enterch:
			return nil
		case event, ok := <-beventch:
			if !ok {
				return errors.WithMessage(err, "unexpected closed channel while waiting for block event")
			}
			a.Printer().PrintBlock(event.Block)
			fmt.Println("Press <enter> to terminate")
		}
	}
}
