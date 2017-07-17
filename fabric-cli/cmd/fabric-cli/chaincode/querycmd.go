/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chaincode

import (
	"encoding/json"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/common"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var queryCmd = &cobra.Command{
	Use:   "query",
	Short: "Query chaincode.",
	Long:  "Query chaincode",
	Run: func(cmd *cobra.Command, args []string) {
		if common.Config().ChaincodeID() == "" {
			fmt.Printf("\nMust specify the chaincode ID\n\n")
			cmd.HelpFunc()(cmd, args)
			return
		}
		action, err := newQueryAction(cmd.Flags())
		if err != nil {
			common.Config().Logger().Criticalf("Error while initializing queryAction: %v", err)
			return
		}

		defer action.Terminate()

		err = action.query()
		if err != nil {
			common.Config().Logger().Criticalf("Error while running queryAction: %v", err)
		}
	},
}

func getQueryCmd() *cobra.Command {
	flags := queryCmd.Flags()
	common.Config().InitPeerURL(flags)
	common.Config().InitChannelID(flags)
	common.Config().InitChaincodeID(flags)
	common.Config().InitArgs(flags)
	common.Config().InitIterations(flags)
	common.Config().InitSleepTime(flags)
	return queryCmd
}

type queryAction struct {
	common.Action
	numInvoked uint32
	done       chan bool
}

func newQueryAction(flags *pflag.FlagSet) (*queryAction, error) {
	action := &queryAction{done: make(chan bool)}
	err := action.Initialize(flags)
	return action, err
}

func (action *queryAction) query() error {
	channelClient, err := action.ChannelClient()
	if err != nil {
		return fmt.Errorf("Error getting channel client: %v", err)
	}

	argBytes := []byte(common.Config().Args())
	args := &common.ArgStruct{}
	err = json.Unmarshal(argBytes, args)
	if err != nil {
		return fmt.Errorf("Error unmarshaling JSON arg string: %v", err)
	}

	if common.Config().Iterations() > 1 {
		go action.queryMultiple(channelClient, args.Func, args.Args, common.Config().Iterations())

		completed := false
		for !completed {
			select {
			case <-action.done:
				completed = true
			case <-time.After(5 * time.Second):
				fmt.Printf("... completed %d out of %d\n", action.numInvoked, common.Config().Iterations())
			}
		}
	} else {
		response, err := action.doQuery(channelClient, args.Func, args.Args)
		if err != nil {
			fmt.Printf("Error invoking chaincode: %v\n", err)
		} else {
			fmt.Printf("***** Response: %s\n", response)
		}
	}

	return nil
}

func (action *queryAction) queryMultiple(channel apifabclient.Channel, fctn string, args []string, iterations int) {
	fmt.Printf("Querying CC %d times ...\n", iterations)
	for i := 0; i < iterations; i++ {
		if response, err := action.doQuery(channel, fctn, args); err != nil {
			fmt.Printf("Error invoking chaincode: %v\n", err)
		} else {
			common.Config().Logger().Infof("***** Response: %s\n", response)
		}

		if (i+1) < iterations && common.Config().SleepTime() > 0 {
			time.Sleep(time.Duration(common.Config().SleepTime()) * time.Millisecond)
		}
		atomic.AddUint32(&action.numInvoked, 1)
	}
	fmt.Printf("Completed %d queries\n", iterations)
	action.done <- true
}

func (action *queryAction) doQuery(channel apifabclient.Channel, fctn string, args []string) ([]byte, error) {
	common.Config().Logger().Infof("Invoking chaincode: %s on channel: %s, function: %s, args: %v\n", common.Config().ChaincodeID(), common.Config().ChannelID(), fctn, args)

	response, err := QueryChaincode(channel, action.Peers(), common.Config().ChaincodeID(), common.Config().ChannelID(), fctn, args)
	if err != nil {
		return nil, err
	}

	return response, nil
}
