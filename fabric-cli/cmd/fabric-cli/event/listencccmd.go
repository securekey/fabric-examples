/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package event

import (
	"fmt"

	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/action"
	cliconfig "github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/config"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var listenccCmd = &cobra.Command{
	Use:   "listencc",
	Short: "Listen to chaincode events.",
	Long:  "Listen to chaincode events",
	Run: func(cmd *cobra.Command, args []string) {
		if cliconfig.Config().ChaincodeID() == "" {
			fmt.Printf("\nMust specify the chaincode ID\n\n")
			cmd.HelpFunc()(cmd, args)
			return
		}
		if cliconfig.Config().ChaincodeEvent() == "" {
			fmt.Printf("\nMust specify the event name\n\n")
			cmd.HelpFunc()(cmd, args)
			return
		}

		action, err := newListenCCAction(cmd.Flags())
		if err != nil {
			fmt.Printf("\nError while initializing listenCCAction: %v\n", err)
			return
		}

		defer action.Terminate()

		err = action.invoke()
		if err != nil {
			fmt.Printf("\nError while running listenCCAction: %v\n", err)
		}
	},
}

func getListenCCCmd() *cobra.Command {
	flags := listenccCmd.Flags()
	cliconfig.InitPeerURL(flags, "", "The URL of the peer on which to listen for events, e.g. grpcs://localhost:7051")
	cliconfig.InitChaincodeID(flags)
	cliconfig.InitChaincodeEvent(flags)
	return listenccCmd
}

type listenccAction struct {
	action.Action
	inputEvent
}

func newListenCCAction(flags *pflag.FlagSet) (*listenccAction, error) {
	action := &listenccAction{inputEvent: inputEvent{done: make(chan bool)}}
	err := action.Initialize(flags)
	return action, err
}

func (a *listenccAction) invoke() error {
	eventHub, err := a.EventHub()
	if err != nil {
		return err
	}

	fmt.Printf("Registering CC event on chaincode [%s] and event [%s]\n", cliconfig.Config().ChaincodeID(), cliconfig.Config().ChaincodeEvent())

	registration := eventHub.RegisterChaincodeEvent(cliconfig.Config().ChaincodeID(), cliconfig.Config().ChaincodeEvent(), func(event *apifabclient.ChaincodeEvent) {
		a.Printer().PrintChaincodeEvent(event)
	})

	a.WaitForEnter()

	fmt.Printf("Unregistering CC event on chaincode [%s] and event [%s]\n", cliconfig.Config().ChaincodeID(), cliconfig.Config().ChaincodeEvent())
	eventHub.UnregisterChaincodeEvent(registration)

	return nil
}
