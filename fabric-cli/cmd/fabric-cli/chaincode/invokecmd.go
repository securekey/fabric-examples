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
	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/chaincode/invoketask"
	cliconfig "github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/config"
	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/executor"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var invokeCmd = &cobra.Command{
	Use:   "invoke",
	Short: "invoke chaincode.",
	Long:  "invoke chaincode",
	Run: func(cmd *cobra.Command, args []string) {
		if cliconfig.Config().ChaincodeID() == "" {
			fmt.Printf("\nMust specify the chaincode ID\n\n")
			cmd.HelpFunc()(cmd, args)
			return
		}
		action, err := newInvokeAction(cmd.Flags())
		if err != nil {
			cliconfig.Config().Logger().Errorf("Error while initializing invokeAction: %v", err)
			return
		}

		defer action.Terminate()

		err = action.invoke()
		if err != nil {
			cliconfig.Config().Logger().Errorf("Error while running invokeAction: %v", err)
		}
	},
}

func getInvokeCmd() *cobra.Command {
	flags := invokeCmd.Flags()
	cliconfig.InitPeerURL(flags)
	cliconfig.InitChannelID(flags)
	cliconfig.InitChaincodeID(flags)
	cliconfig.InitArgs(flags)
	cliconfig.InitIterations(flags)
	cliconfig.InitSleepTime(flags)
	cliconfig.InitTimeout(flags)
	cliconfig.InitPrintPayloadOnly(flags)
	cliconfig.InitConcurrency(flags)
	cliconfig.InitMaxAttempts(flags)
	cliconfig.InitInitialBackoff(flags)
	cliconfig.InitMaxBackoff(flags)
	cliconfig.InitBackoffFactor(flags)
	cliconfig.InitVerbosity(flags)
	cliconfig.InitSelectionProvider(flags)
	return invokeCmd
}

type invokeAction struct {
	action.Action
	numInvoked uint32
	done       chan bool
}

func newInvokeAction(flags *pflag.FlagSet) (*invokeAction, error) {
	action := &invokeAction{done: make(chan bool)}
	err := action.Initialize(flags)
	return action, err
}

func (a *invokeAction) invoke() error {
	channelClient, err := a.ChannelClient()
	if err != nil {
		return errors.Errorf("Error getting channel client: %v", err)
	}

	argsArray, err := action.ArgsArray()
	if err != nil {
		return err
	}

	executor := executor.NewConcurrent("Invoke Chaincode", cliconfig.Config().Concurrency())
	executor.Start()
	defer executor.Stop(true)

	var errs []error
	success := 0

	var targets []fab.Peer
	if len(cliconfig.Config().PeerURL()) > 0 || len(cliconfig.Config().OrgIDs()) > 0 {
		targets = a.Peers()
	}

	var wg sync.WaitGroup
	var mutex sync.RWMutex
	var tasks []*invoketask.Task
	var taskID int
	for i := 0; i < cliconfig.Config().Iterations(); i++ {
		for _, args := range argsArray {
			taskID++
			task := invoketask.New(
				strconv.Itoa(taskID), channelClient, targets,
				cliconfig.Config().ChaincodeID(),
				&args, executor,
				retry.Opts{
					Attempts:       cliconfig.Config().MaxAttempts(),
					InitialBackoff: cliconfig.Config().InitialBackoff(),
					MaxBackoff:     cliconfig.Config().MaxBackoff(),
					BackoffFactor:  cliconfig.Config().BackoffFactor(),
					RetryableCodes: retry.ChannelClientRetryableCodes,
				},
				cliconfig.Config().Verbose() || cliconfig.Config().Iterations() == 1,
				cliconfig.Config().PrintPayloadOnly(), a.Printer(),

				func(err error) {
					defer wg.Done()
					mutex.Lock()
					defer mutex.Unlock()
					if err != nil {
						errs = append(errs, err)
					} else {
						success++
					}
				})
			tasks = append(tasks, task)
		}
	}

	numInvocations := len(tasks)

	wg.Add(numInvocations)

	done := make(chan bool)
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		for {
			select {
			case <-ticker.C:
				mutex.RLock()
				if len(errs) > 0 {
					fmt.Printf("*** %d failed invocation(s) out of %d\n", len(errs), numInvocations)
				}
				fmt.Printf("*** %d successfull invocation(s) out of %d\n", success, numInvocations)
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

	var allErrs []error
	var attempts int
	for _, task := range tasks {
		attempts = attempts + task.Attempts()
		if task.LastError() != nil {
			allErrs = append(allErrs, task.LastError())
		}
	}

	if len(errs) > 0 {
		fmt.Printf("\n*** %d errors invoking chaincode:\n", len(errs))
		for _, err := range errs {
			fmt.Printf("%s\n", err)
		}
	} else if len(allErrs) > 0 {
		fmt.Printf("\n*** %d transient errors invoking chaincode:\n", len(allErrs))
		for _, err := range allErrs {
			fmt.Printf("%s\n", err)
		}
	}

	if numInvocations > 1 {
		fmt.Printf("\n")
		fmt.Printf("*** ---------- Summary: ----------\n")
		fmt.Printf("***   - Invocations:     %d\n", numInvocations)
		fmt.Printf("***   - Successfull:     %d\n", success)
		fmt.Printf("***   - Total attempts:  %d\n", attempts)
		fmt.Printf("***   - Duration:        %s\n", duration)
		fmt.Printf("***   - Rate:            %2.2f/s\n", float64(numInvocations)/duration.Seconds())
		fmt.Printf("*** ------------------------------\n")
	}

	return nil
}
