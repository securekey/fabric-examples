/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chaincode

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/pkg/errors"
	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/action"
	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/chaincode/querytask"
	cliconfig "github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/config"
	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/executor"
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
	cliconfig.InitPeerURL(flags)
	cliconfig.InitChannelID(flags)
	cliconfig.InitChaincodeID(flags)
	cliconfig.InitArgs(flags)
	cliconfig.InitIterations(flags)
	cliconfig.InitSleepTime(flags)
	cliconfig.InitTimeout(flags)
	cliconfig.InitPrintPayloadOnly(flags)
	cliconfig.InitConcurrency(flags)
	cliconfig.InitVerbosity(flags)
	cliconfig.InitSelectionProvider(flags)
	return queryCmd
}

type queryAction struct {
	action.Action
	numInvoked uint32
	done       chan bool
}

func newQueryAction(flags *pflag.FlagSet) (*queryAction, error) {
	action := &queryAction{done: make(chan bool)}
	err := action.Initialize(flags)
	return action, err
}

func (a *queryAction) query() error {
	channelClient, err := a.ChannelClient()
	if err != nil {
		return errors.Errorf("Error getting channel client: %v", err)
	}

	argsArray, err := action.ArgsArray()
	if err != nil {
		return err
	}

	var targets []fab.Peer
	if len(cliconfig.Config().PeerURL()) > 0 || len(cliconfig.Config().OrgIDs()) > 0 {
		targets = a.Peers()
	}

	executor := executor.NewConcurrent("Query Chaincode", cliconfig.Config().Concurrency())
	executor.Start()
	defer executor.Stop(true)

	verbose := cliconfig.Config().Verbose() || cliconfig.Config().Iterations() == 1

	var mutex sync.RWMutex
	var tasks []*querytask.Task
	var errs []error
	var wg sync.WaitGroup
	var taskID int
	var success int
	for i := 0; i < cliconfig.Config().Iterations(); i++ {
		for _, args := range argsArray {
			taskID++
			task := querytask.New(
				strconv.Itoa(taskID), channelClient, targets, &args, a.Printer(), verbose, cliconfig.Config().PrintPayloadOnly(),

				func(err error) {
					defer wg.Done()
					mutex.Lock()
					if err != nil {
						errs = append(errs, err)
					} else {
						success++
					}
					mutex.Unlock()
				})
			tasks = append(tasks, task)
		}
	}

	numInvocations := len(tasks)
	wg.Add(numInvocations)

	done := make(chan bool)
	go func() {
		ticker := time.NewTicker(3 * time.Second)
		for {
			select {
			case <-ticker.C:
				mutex.RLock()
				if len(errs) > 0 {
					fmt.Printf("*** %d failed query(s) out of %d\n", len(errs), numInvocations)
				}
				fmt.Printf("*** %d successfull query(s) out of %d\n", success, numInvocations)
				mutex.RUnlock()
			case <-done:
				return
			}
		}
	}()

	startTime := time.Now()

	for _, task := range tasks {
		if err := executor.Submit(task); err != nil {
			return errors.Errorf("error submitting task: %s", err)
		}
	}

	// Wait for all tasks to complete
	wg.Wait()
	done <- true
	duration := time.Now().Sub(startTime)

	if len(errs) > 0 {
		fmt.Printf("\n*** %d errors querying chaincode:\n", len(errs))
		for _, err := range errs {
			fmt.Printf("%s\n", err)
		}
	}

	if numInvocations > 1 {
		fmt.Printf("\n")
		fmt.Printf("*** ---------- Summary: ----------\n")
		fmt.Printf("***   - Queries:     %d\n", numInvocations)
		fmt.Printf("***   - Successfull: %d\n", success)
		fmt.Printf("***   - Duration:    %s\n", duration)
		fmt.Printf("***   - Rate:        %2.2f/s\n", float64(numInvocations)/duration.Seconds())
		fmt.Printf("*** ------------------------------\n")
	}

	return nil
}
