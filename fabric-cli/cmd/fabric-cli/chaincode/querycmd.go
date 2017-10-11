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

	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/common"
	cliconfig "github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/config"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var queryCmd = &cobra.Command{
	Use:   "query",
	Short: "Query chaincode.",
	Long:  "Query chaincode",
	Run: func(cmd *cobra.Command, args []string) {
		if cliconfig.Config().ChaincodeID() == "" {
			fmt.Printf("\nMust specify the chaincode ID\n\n")
			cmd.HelpFunc()(cmd, args)
			return
		}
		action, err := newQueryAction(cmd.Flags())
		if err != nil {
			cliconfig.Config().Logger().Errorf("Error while initializing queryAction: %v", err)
			return
		}

		defer action.Terminate()

		err = action.query()
		if err != nil {
			cliconfig.Config().Logger().Errorf("Error while running queryAction: %v", err)
		}
	},
}

func getQueryCmd() *cobra.Command {
	flags := queryCmd.Flags()
	cliconfig.Config().InitPeerURL(flags)
	cliconfig.Config().InitChannelID(flags)
	cliconfig.Config().InitChaincodeID(flags)
	cliconfig.Config().InitArgs(flags)
	cliconfig.Config().InitIterations(flags)
	cliconfig.Config().InitSleepTime(flags)
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

	argBytes := []byte(cliconfig.Config().Args())
	args := &common.ArgStruct{}
	err = json.Unmarshal(argBytes, args)
	if err != nil {
		return fmt.Errorf("Error unmarshaling JSON arg string: %v", err)
	}

	if cliconfig.Config().Iterations() > 1 {
		go action.queryMultiple(channelClient, args.Func, asBytes(args.Args), cliconfig.Config().Iterations())

		completed := false
		for !completed {
			select {
			case <-action.done:
				completed = true
			case <-time.After(5 * time.Second):
				fmt.Printf("... completed %d out of %d\n", action.numInvoked, cliconfig.Config().Iterations())
			}
		}
	} else {
		if response, err := action.doQuery(channelClient, args.Func, asBytes(args.Args)); err != nil {
			fmt.Printf("Error invoking chaincode: %v\n", err)
		} else {
			action.Printer().Print("%s", response)
		}
		fmt.Println("Done!")
	}

	return nil
}
func (action *queryAction) queryMultiple(channel apitxn.ChannelClient, fctn string, args [][]byte, iterations int) {
	fmt.Printf("Querying CC %d times ...\n", iterations)
	for i := 0; i < iterations; i++ {
		if response, err := action.doQuery(channel, fctn, args); err != nil {
			fmt.Printf("Error invoking chaincode: %v\n", err)
		} else {
			action.Printer().Print("%s", response)
		}

		if (i+1) < iterations && cliconfig.Config().SleepTime() > 0 {
			time.Sleep(time.Duration(cliconfig.Config().SleepTime()) * time.Millisecond)
		}
		atomic.AddUint32(&action.numInvoked, 1)
	}
	fmt.Printf("Completed %d queries\n", iterations)
	action.done <- true
}

func (action *queryAction) doQuery(channelClient apitxn.ChannelClient, fctn string, args [][]byte) ([]byte, error) {
	cliconfig.Config().Logger().Infof("Querying chaincode: %s on channel: %s, function: %s, args: %v\n", cliconfig.Config().ChaincodeID(), cliconfig.Config().ChannelID(), fctn, args)

	return channelClient.QueryWithOpts(
		apitxn.QueryRequest{
			Fcn:         fctn,
			Args:        args,
			ChaincodeID: cliconfig.Config().ChaincodeID(),
		},
		apitxn.QueryOpts{
			ProposalProcessors: action.ProposalProcessors(),
			Timeout:            cliconfig.Config().Timeout(),
		},
	)
}
