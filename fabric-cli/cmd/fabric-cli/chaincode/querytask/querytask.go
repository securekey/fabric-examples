/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package querytask

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/action"
	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/chaincode/utils"
	cliconfig "github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/config"
	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/printer"
)

// Task is the query task
type Task struct {
	channelClient *channel.Client
	id            string
	args          *action.ArgStruct
	callback      func(err error)
	printer       printer.Printer
	verbose       bool
}

// New creates a new query Task
func New(id string, channelClient *channel.Client, args *action.ArgStruct, printer printer.Printer, verbose bool, callback func(err error)) *Task {
	return &Task{
		id:            id,
		channelClient: channelClient,
		args:          args,
		callback:      callback,
		printer:       printer,
		verbose:       verbose,
	}
}

// Invoke invokes the query task
func (t *Task) Invoke() {
	if _, err := t.channelClient.Query(
		channel.Request{
			ChaincodeID: cliconfig.Config().ChaincodeID(),
			Fcn:         t.args.Func,
			Args:        utils.AsBytes(t.args.Args),
		},
	); err != nil {
		cliconfig.Config().Logger().Debugf("(%s) - Error querying chaincode: %s\n", t.id, err)
		t.callback(err)
	} else {
		cliconfig.Config().Logger().Debugf("(%s) - Chaincode query was successful\n", t.id)
		t.callback(nil)
	}
}
