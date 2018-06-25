/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package invoketask

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
	"github.com/pkg/errors"
	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/action"
	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/chaincode/invokeerror"
	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/chaincode/utils"
	cliconfig "github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/config"
	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/executor"
	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/printer"
)

// Task is a Task that invokes a chaincode
type Task struct {
	executor      *executor.Executor
	channelClient *channel.Client
	targets       []fab.Peer
	id            string
	ccID          string
	args          *action.ArgStruct
	retryOpts     retry.Opts
	attempt       int
	lastErr       error
	callback      func(err error)
	verbose       bool
	printer       printer.Printer
	txID          string
	payloadOnly   bool
}

// New returns a new Task
func New(id string, channelClient *channel.Client, targets []fab.Peer, ccID string, args *action.ArgStruct,
	executor *executor.Executor, retryOpts retry.Opts, verbose bool,
	payloadOnly bool, p printer.Printer, callback func(err error)) *Task {
	return &Task{
		id:            id,
		channelClient: channelClient,
		targets:       targets,
		printer:       p,
		ccID:          ccID,
		args:          args,
		executor:      executor,
		retryOpts:     retryOpts,
		callback:      callback,
		attempt:       1,
		verbose:       verbose,
		payloadOnly:   payloadOnly,
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

// Invoke invokes the task
func (t *Task) Invoke() {
	if err := t.doInvoke(); err != nil {
		t.lastErr = err
		t.callback(err)
	} else {
		cliconfig.Config().Logger().Debugf("(%s) - Successfully invoked chaincode\n", t.id)
		t.callback(nil)
	}
}

func (t *Task) doInvoke() error {
	cliconfig.Config().Logger().Debugf("(%s) - Invoking chaincode: %s, function: %s, args: %+v. Attempt #%d...\n",
		t.id, t.ccID, t.args.Func, t.args.Args, t.attempt)

	var opts []channel.RequestOption
	opts = append(opts, channel.WithRetry(t.retryOpts))
	if len(t.targets) > 0 {
		opts = append(opts, channel.WithTargets(t.targets...))
	}

	response, err := t.channelClient.Execute(
		channel.Request{
			ChaincodeID: t.ccID,
			Fcn:         t.args.Func,
			Args:        utils.AsBytes(t.args.Args),
		},
		opts...,
	)
	if err != nil {
		return invokeerror.Errorf(invokeerror.TransientError, "SendTransactionProposal return error: %v", err)
	}

	if t.verbose {
		t.printer.PrintTxProposalResponses(response.Responses, t.payloadOnly)
	}

	t.txID = string(response.TransactionID)

	switch pb.TxValidationCode(response.TxValidationCode) {
	case pb.TxValidationCode_VALID:
		cliconfig.Config().Logger().Debugf("(%s) - Successfully committed transaction [%s] ...\n", t.id, response.TransactionID)
		return nil
	case pb.TxValidationCode_DUPLICATE_TXID, pb.TxValidationCode_MVCC_READ_CONFLICT, pb.TxValidationCode_PHANTOM_READ_CONFLICT:
		cliconfig.Config().Logger().Debugf("(%s) - Transaction commit failed for [%s] with code [%s]. This is most likely a transient error.\n", t.id, response.TransactionID, response.TxValidationCode)
		return invokeerror.Wrapf(invokeerror.TransientError, errors.New("Duplicate TxID"), "invoke Error received from eventhub for TxID [%s]. Code: %s", response.TransactionID, response.TxValidationCode)
	default:
		cliconfig.Config().Logger().Debugf("(%s) - Transaction commit failed for [%s] with code [%s].\n", t.id, response.TransactionID, response.TxValidationCode)
		return invokeerror.Wrapf(invokeerror.PersistentError, errors.New("error"), "invoke Error received from eventhub for TxID [%s]. Code: %s", response.TransactionID, response.TxValidationCode)
	}

}
