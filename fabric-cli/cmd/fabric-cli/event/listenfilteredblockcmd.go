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

var listenFilteredBlockCmd = &cobra.Command{
	Use:   "listenfilteredblock",
	Short: "Listen to filtered block events.",
	Long:  "Listen to filtered block events",
	Run: func(cmd *cobra.Command, args []string) {
		action, err := newlistenFilteredBlockAction(cmd.Flags())
		if err != nil {
			cliconfig.Config().Logger().Errorf("Error while initializing listenFilteredBlockAction: %v", err)
			return
		}

		defer action.Terminate()

		err = action.invoke()
		if err != nil {
			cliconfig.Config().Logger().Errorf("Error while running listenFilteredBlockAction: %v", err)
		}
	},
}

func getListenFilteredBlockCmd() *cobra.Command {
	flags := listenFilteredBlockCmd.Flags()
	cliconfig.InitChannelID(flags)
	cliconfig.InitPeerURL(flags, "", "The URL of the peer on which to listen for events, e.g. localhost:7051")
	cliconfig.InitSeekType(flags)
	cliconfig.InitBlockNum(flags)
	return listenFilteredBlockCmd
}

type listenFilteredBlockAction struct {
	action.Action
	inputEvent
}

func newlistenFilteredBlockAction(flags *pflag.FlagSet) (*listenFilteredBlockAction, error) {
	action := &listenFilteredBlockAction{inputEvent: inputEvent{done: make(chan bool)}}
	err := action.Initialize(flags)
	return action, err
}

func (a *listenFilteredBlockAction) invoke() error {
	eventClient, err := a.EventClient(event.WithSeekType(cliconfig.Config().SeekType()), event.WithBlockNum(cliconfig.Config().BlockNum()))
	if err != nil {
		return err
	}

	fmt.Printf("Registering filtered block event\n")

	breg, beventch, err := eventClient.RegisterFilteredBlockEvent()
	if err != nil {
		return errors.WithMessage(err, "Error registering for filtered block events")
	}
	defer eventClient.Unregister(breg)

	enterch := a.WaitForEnter()
	for {
		select {
		case _, _ = <-enterch:
			return nil
		case event, ok := <-beventch:
			if !ok {
				return errors.WithMessage(err, "unexpected closed channel while waiting for filtered block event")
			}
			a.Printer().PrintFilteredBlock(event.FilteredBlock)
			fmt.Println("Press <enter> to terminate")
		}
	}
}
