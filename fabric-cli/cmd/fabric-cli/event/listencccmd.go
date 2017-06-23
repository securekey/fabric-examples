/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package event

import (
	"fmt"

	"bufio"
	"os"

	"github.com/hyperledger/fabric-sdk-go/api"
	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/common"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var listenccCmd = &cobra.Command{
	Use:   "listencc",
	Short: "Listen to chaincode events.",
	Long:  "Listen to chaincode events",
	Run: func(cmd *cobra.Command, args []string) {
		if common.Config().ChaincodeID() == "" {
			fmt.Printf("\nMust specify the chaincode ID\n\n")
			cmd.HelpFunc()(cmd, args)
			return
		}
		if common.Config().ChaincodeEvent() == "" {
			fmt.Printf("\nMust specify the event name\n\n")
			cmd.HelpFunc()(cmd, args)
			return
		}

		action, err := newListenCCAction(cmd.Flags())
		if err != nil {
			fmt.Printf("\nError while initializing listenCCAction: %v\n", err)
			return
		}

		err = action.invoke()
		if err != nil {
			fmt.Printf("\nError while running listenCCAction: %v\n", err)
			return
		}
	},
}

// Cmd returns the listencc command
func getListenCCCmd() *cobra.Command {
	flags := listenccCmd.Flags()
	common.Config().InitPeerURL(flags, "", "The URL of the peer on which to listen for events, e.g. localhost:7051")
	common.Config().InitChaincodeID(flags)
	common.Config().InitChaincodeEvent(flags)
	return listenccCmd
}

type listenccAction struct {
	done chan bool
	common.ActionImpl
}

func newListenCCAction(flags *pflag.FlagSet) (*listenccAction, error) {
	action := &listenccAction{done: make(chan bool)}
	err := action.Initialize(flags)
	return action, err
}

func (action *listenccAction) invoke() error {
	eventHub := action.EventHub()

	fmt.Printf("Registering CC event on chaincode [%s] and event [%s]\n", common.Config().ChaincodeID(), common.Config().ChaincodeEvent())

	registration := eventHub.RegisterChaincodeEvent(common.Config().ChaincodeID(), common.Config().ChaincodeEvent(), func(event *api.ChaincodeEvent) {
		fmt.Printf("Received CC event:\n")
		fmt.Printf("- Channel ID: %s\n", event.ChannelID)
		fmt.Printf("- Chaincode ID: %s\n", event.ChaincodeID)
		fmt.Printf("- Event: %s\n", event.EventName)
		fmt.Printf("- TxID: %s\n", event.TxID)
		fmt.Printf("- Payload: %v\n", event.Payload)

		fmt.Println("Press <enter> to terminate")
	})

	go action.readFromCLI()

	<-action.done

	fmt.Printf("Unregistering CC event on chaincode [%s] and event [%s]\n", common.Config().ChaincodeID(), common.Config().ChaincodeEvent())
	eventHub.UnregisterChaincodeEvent(registration)

	return nil
}

func (action *listenccAction) readFromCLI() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Press <enter> to terminate")
	reader.ReadString('\n')
	action.done <- true
}
