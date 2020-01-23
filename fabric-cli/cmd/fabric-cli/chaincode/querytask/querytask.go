/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package querytask

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel/invoke"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/action"
	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/chaincode/utils"
	cliconfig "github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/config"
	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/printer"
)

// Task is the query task
type Task struct {
	ctxt          utils.Context
	channelClient *channel.Client
	targets       []fab.Peer
	retryOpts     retry.Opts
	id            string
	args          *action.ArgStruct
	startedCB     func()
	completedCB   func(err error)
	printer       printer.Printer
	verbose       bool
	payloadOnly   bool
	validate      bool
	attempt       int
	lastErr       error
}

// New creates a new query Task
func New(ctxt utils.Context, id string, channelClient *channel.Client, targets []fab.Peer, args *action.ArgStruct, printer printer.Printer,
	retryOpts retry.Opts, verbose bool, payloadOnly bool, validate bool, startedCB func(), completedCB func(err error)) *Task {
	return &Task{
		ctxt:          ctxt,
		id:            id,
		channelClient: channelClient,
		targets:       targets,
		retryOpts:     retryOpts,
		args:          args,
		startedCB:     startedCB,
		completedCB:   completedCB,
		attempt:       1,
		printer:       printer,
		verbose:       verbose,
		payloadOnly:   payloadOnly,
		validate:      validate,
	}
}

// Invoke invokes the query task
func (t *Task) Invoke() {
	t.startedCB()

	var opts []channel.RequestOption
	opts = append(opts, channel.WithRetry(t.retryOpts))
	opts = append(opts, channel.WithBeforeRetry(func(err error) {
		t.attempt++
	}))
	if len(t.targets) > 0 {
		opts = append(opts, channel.WithTargets(t.targets...))
	}

	request := channel.Request{
		ChaincodeID: cliconfig.Config().ChaincodeID(),
		Fcn:         t.args.Func,
		Args:        utils.AsBytes(t.ctxt, t.args.Args),
	}

	var additionalHandlers []invoke.Handler
	if t.validate {
		// Add the validation handlers
		additionalHandlers = append(additionalHandlers,
			invoke.NewEndorsementValidationHandler(
				invoke.NewSignatureValidationHandler(),
			),
		)
	}

	response, err := t.channelClient.InvokeHandler(
		invoke.NewProposalProcessorHandler(
			invoke.NewEndorsementHandler(additionalHandlers...),
		),
		request, opts...)
	if err != nil {
		cliconfig.Config().Logger().Debugf("(%s) - Error querying chaincode: %s\n", t.id, err)
		t.lastErr = err
		t.completedCB(err)
	} else {
		cliconfig.Config().Logger().Debugf("(%s) - Chaincode query was successful\n", t.id)

		if t.verbose {
			t.printer.PrintTxProposalResponses(response.Responses, t.payloadOnly)
		}

		t.completedCB(nil)
	}
}

// Attempts returns the number of invocation attempts that were made
// in order to achieve a successful response
func (t *Task) Attempts() int {
	return t.attempt
}

// LastError returns the last error that was recorder
func (t *Task) LastError() error {
	return t.lastErr
}
