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

	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/pkg/errors"
	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/action"
	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/chaincode/multitask"
	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/chaincode/querytask"
	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/chaincode/task"
	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/chaincode/utils"
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
	cliconfig.InitValidate(flags)
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
	var tasks []task.Task
	var errs []error
	var wg sync.WaitGroup
	var taskID int
	var success int
	var successDurations []time.Duration
	var failDurations []time.Duration

	for i := 0; i < cliconfig.Config().Iterations(); i++ {
		ctxt := utils.NewContext()
		multiTask := multitask.New(wg.Done)
		for _, args := range argsArray {
			taskID++
			var startTime time.Time
			cargs := args
			task := querytask.New(
				ctxt,
				strconv.Itoa(taskID), channelClient, targets, &cargs, a.Printer(),
				retry.Opts{
					Attempts:       cliconfig.Config().MaxAttempts(),
					InitialBackoff: cliconfig.Config().InitialBackoff(),
					MaxBackoff:     cliconfig.Config().MaxBackoff(),
					BackoffFactor:  cliconfig.Config().BackoffFactor(),
					RetryableCodes: retry.ChannelClientRetryableCodes,
				},
				verbose,
				cliconfig.Config().PrintPayloadOnly(),
				cliconfig.Config().Validate(),
				func() {
					startTime = time.Now()
				},
				func(err error) {
					duration := time.Since(startTime)
					mutex.Lock()
					if err != nil {
						errs = append(errs, err)
					} else {
						success++
						successDurations = append(successDurations, duration)
					}
					mutex.Unlock()
				})
			multiTask.Add(task)
		}
		tasks = append(tasks, multiTask)
	}

	wg.Add(len(tasks))

	numInvocations := len(tasks) * len(argsArray)

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
	sleepTime := time.Duration(cliconfig.Config().SleepTime()) * time.Millisecond

	for _, task := range tasks {
		if err := executor.Submit(task); err != nil {
			return errors.Errorf("error submitting task: %s", err)
		}
		if sleepTime > 0 {
			time.Sleep(sleepTime)
		}
	}

	// Wait for all tasks to complete
	wg.Wait()
	done <- true

	duration := time.Now().Sub(startTime)

	var allErrs []error
	var attempts int
	for _, task := range tasks {
		attempts = attempts + task.Attempts()
		if task.LastError() != nil {
			allErrs = append(allErrs, task.LastError())
		}
	}

	if len(errs) > 0 {
		fmt.Printf("\n*** %d errors querying chaincode:\n", len(errs))
		for _, err := range errs {
			fmt.Printf("%s\n", err)
		}
	} else if len(allErrs) > 0 {
		fmt.Printf("\n*** %d transient errors querying chaincode:\n", len(allErrs))
		for _, err := range allErrs {
			fmt.Printf("%s\n", err)
		}
	}

	if numInvocations/len(argsArray) > 1 {
		fmt.Printf("\n")
		fmt.Printf("*** ---------- Summary: ----------\n")
		fmt.Printf("***   - Queries:         %d\n", numInvocations)
		fmt.Printf("***   - Concurrency:     %d\n", cliconfig.Config().Concurrency())
		fmt.Printf("***   - Successfull:     %d\n", success)
		fmt.Printf("***   - Total attempts:  %d\n", attempts)
		fmt.Printf("***   - Duration:        %2.2fs\n", duration.Seconds())
		fmt.Printf("***   - Rate:            %2.2f/s\n", float64(numInvocations)/duration.Seconds())
		fmt.Printf("***   - Average:         %2.2fs\n", average(append(successDurations, failDurations...)))
		fmt.Printf("***   - Average Success: %2.2fs\n", average(successDurations))
		fmt.Printf("***   - Average Fail:    %2.2fs\n", average(failDurations))
		fmt.Printf("***   - Min Success:     %2.2fs\n", min(successDurations))
		fmt.Printf("***   - Max Success:     %2.2fs\n", max(successDurations))
		fmt.Printf("*** ------------------------------\n")
	}

	return nil
}
