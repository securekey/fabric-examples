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
	"bufio"
	"fmt"
	"os"

	fabricCommon "github.com/hyperledger/fabric/protos/common"
	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/common"
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
			common.Config().Logger().Criticalf("Error while initializing listenBlockAction: %v", err)
			return
		}

		err = action.invoke()
		if err != nil {
			common.Config().Logger().Criticalf("Error while running listenBlockAction: %v", err)
			return
		}
	},
}

// Cmd returns the listenBlock command
func getListenBlockCmd() *cobra.Command {
	common.Config().InitPeerURL(listenBlockCmd.Flags(), "", "The URL of the peer on which to listen for events, e.g. localhost:7051")
	return listenBlockCmd
}

type listenBlockAction struct {
	common.ActionImpl
	done chan bool
}

func newlistenBlockAction(flags *pflag.FlagSet) (*listenBlockAction, error) {
	action := &listenBlockAction{done: make(chan bool)}
	err := action.Initialize(flags)
	return action, err
}

func (action *listenBlockAction) invoke() error {
	eventHub := action.EventHub()

	fmt.Printf("Registering block event\n")

	callback := func(block *fabricCommon.Block) {
		action.Printer().PrintBlock(block)
		fmt.Println("Press <enter> to terminate")
	}

	eventHub.RegisterBlockEvent(callback)

	go action.readFromCLI()

	<-action.done

	fmt.Printf("Unregistering block event\n")
	eventHub.UnregisterBlockEvent(callback)

	return nil
}

func (action *listenBlockAction) readFromCLI() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Press <enter> to terminate")
	reader.ReadString('\n')
	action.done <- true
}
