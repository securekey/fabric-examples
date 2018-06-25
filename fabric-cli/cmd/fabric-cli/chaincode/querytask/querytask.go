/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package querytask

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/action"
	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/chaincode/utils"
	cliconfig "github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/config"
	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/printer"
)

// Task is the query task
type Task struct {
	channelClient *channel.Client
	targets       []fab.Peer
	id            string
	args          *action.ArgStruct
	callback      func(err error)
	printer       printer.Printer
	verbose       bool
	payloadOnly   bool
}

// New creates a new query Task
func New(id string, channelClient *channel.Client, targets []fab.Peer, args *action.ArgStruct, printer printer.Printer, verbose bool, payloadOnly bool, callback func(err error)) *Task {
	return &Task{
		id:            id,
		channelClient: channelClient,
		targets:       targets,
		args:          args,
		callback:      callback,
		printer:       printer,
		verbose:       verbose,
		payloadOnly:   payloadOnly,
	}
}

// Invoke invokes the query task
func (t *Task) Invoke() {
	var opts []channel.RequestOption
	if len(t.targets) > 0 {
		opts = append(opts, channel.WithTargets(t.targets...))
	}
	if response, err := t.channelClient.Query(
		channel.Request{
			ChaincodeID: cliconfig.Config().ChaincodeID(),
			Fcn:         t.args.Func,
			Args:        utils.AsBytes(t.args.Args),
		},
		opts...,
	); err != nil {
		cliconfig.Config().Logger().Debugf("(%s) - Error querying chaincode: %s\n", t.id, err)
		t.callback(err)
	} else {
		cliconfig.Config().Logger().Debugf("(%s) - Chaincode query was successful\n", t.id)

		if t.verbose {
			t.printer.PrintTxProposalResponses(response.Responses, t.payloadOnly)
		}

		t.callback(nil)
	}
}
