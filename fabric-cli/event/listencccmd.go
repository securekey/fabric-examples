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

package event

import (
	"fmt"

	"bufio"
	"os"

	"github.com/hyperledger/fabric-sdk-go/fabric-client/events"
	"github.com/securekey/fabric-examples/fabric-cli/common"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const ccEventNameFlag = "event"

var ccEventName string

var listenccCmd = &cobra.Command{
	Use:   "listencc",
	Short: "Listen to chaincode events.",
	Long:  "Listen to chaincode events",
	Run: func(cmd *cobra.Command, args []string) {
		if common.ChaincodeID == "" {
			fmt.Printf("\nMust specify the chaincode ID\n\n")
			cmd.HelpFunc()(cmd, args)
			return
		}
		if ccEventName == "" {
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
	flags.String(common.PeerFlag, "", "The URL of the peer on which to listen for events, e.g. localhost:7051")
	flags.StringVar(&common.ChaincodeID, common.ChaincodeIDFlag, common.ChaincodeID, "The chaincode ID")
	flags.StringVar(&ccEventName, ccEventNameFlag, "", "The chaincode event name")
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

	fmt.Printf("Registering CC event on chaincode [%s] and event [%s]\n", common.ChaincodeID, ccEventName)

	registration := eventHub.RegisterChaincodeEvent(common.ChaincodeID, ccEventName, func(event *events.ChaincodeEvent) {
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

	fmt.Printf("Unregistering CC event on chaincode [%s] and event [%s]\n", common.ChaincodeID, ccEventName)
	eventHub.UnregisterChaincodeEvent(registration)

	return nil
}

func (action *listenccAction) readFromCLI() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Press <enter> to terminate")
	reader.ReadString('\n')
	action.done <- true
}
