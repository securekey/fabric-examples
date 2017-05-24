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

	pb "github.com/hyperledger/fabric/protos/peer"
	"github.com/securekey/fabric-examples/fabric-sdk-go/fabric-cli/common"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	txIDFlag = "tx"
)

var txID string

var listenTxCmd = &cobra.Command{
	Use:   "listentx",
	Short: "Listen to transaction events.",
	Long:  "Listen to transaction events",
	Run: func(cmd *cobra.Command, args []string) {
		if txID == "" {
			fmt.Printf("\nMust specify the transaction ID\n\n")
			cmd.HelpFunc()(cmd, args)
			return
		}
		action, err := newListenTXAction(cmd.Flags())
		if err != nil {
			common.Logger.Criticalf("Error while initializing listenTxAction: %v", err)
			return
		}

		err = action.invoke()
		if err != nil {
			common.Logger.Criticalf("Error while running listenTxAction: %v", err)
			return
		}
	},
}

// Cmd returns the listentx command
func getListenTXCmd() *cobra.Command {
	flags := listenTxCmd.Flags()
	flags.StringVar(&txID, txIDFlag, "", "The transaction ID")
	flags.String(common.PeerFlag, "", "The URL of the peer on which to listen for events, e.g. localhost:7051")
	return listenTxCmd
}

type listentxAction struct {
	common.ActionImpl
}

func newListenTXAction(flags *pflag.FlagSet) (*listentxAction, error) {
	action := &listentxAction{}
	err := action.Initialize(flags)
	return action, err
}

func (action *listentxAction) invoke() error {
	done := make(chan bool)

	eventHub := action.EventHub()

	fmt.Printf("Registering TX event for TxID [%s]\n", txID)

	eventHub.RegisterTxEvent(txID, func(txID string, code pb.TxValidationCode, err error) {
		fmt.Printf("Received TX event. TxID: %s, Code: %s, Error: %s\n", txID, code, err)
		done <- true
	})

	// Wait for the event
	<-done

	fmt.Printf("Unregistering TX event for TxID [%s]\n", txID)
	eventHub.UnregisterTxEvent(txID)

	return nil
}
