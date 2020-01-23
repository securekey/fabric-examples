/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package event

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/action"
	cliconfig "github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/config"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var listenTxCmd = &cobra.Command{
	Use:   "listentx",
	Short: "Listen to transaction events.",
	Long:  "Listen to transaction events",
	Run: func(cmd *cobra.Command, args []string) {
		if cliconfig.Config().TxID() == "" {
			fmt.Printf("\nMust specify the transaction ID\n\n")
			cmd.HelpFunc()(cmd, args)
			return
		}
		action, err := newListenTXAction(cmd.Flags())
		if err != nil {
			cliconfig.Config().Logger().Errorf("Error while initializing listenTxAction: %v", err)
			return
		}

		defer action.Terminate()

		err = action.invoke()
		if err != nil {
			cliconfig.Config().Logger().Errorf("Error while running listenTxAction: %v", err)
		}
	},
}

func getListenTXCmd() *cobra.Command {
	flags := listenTxCmd.Flags()
	cliconfig.InitChannelID(flags)
	cliconfig.InitTxID(flags)
	cliconfig.InitPeerURL(flags, "", "The URL of the peer on which to listen for events, e.g. grpcs://localhost:7051")
	return listenTxCmd
}

type listentxAction struct {
	action.Action
	inputEvent
}

func newListenTXAction(flags *pflag.FlagSet) (*listentxAction, error) {
	action := &listentxAction{inputEvent: inputEvent{done: make(chan bool)}}
	err := action.Initialize(flags)
	return action, err
}

func (a *listentxAction) invoke() error {

	eventHub, err := a.EventClient()
	if err != nil {
		return err
	}

	fmt.Printf("Registering TX event for TxID [%s]\n", cliconfig.Config().TxID())

	reg, eventch, err := eventHub.RegisterTxStatusEvent(cliconfig.Config().TxID())
	if err != nil {
		return errors.WithMessage(err, "Error registering for block events")
	}
	defer eventHub.Unregister(reg)

	enterch := a.WaitForEnter()
	fmt.Println("Press <enter> to terminate")

	select {
	case _, _ = <-enterch:
		return nil
	case event, ok := <-eventch:
		if !ok {
			return errors.WithMessage(err, "unexpected closed channel while waiting for tx status event")
		}
		fmt.Printf("Received TX event. TxID: %s, Code: %s, Error: %s\n", event.TxID, event.TxValidationCode, err)
	}

	return nil
}
