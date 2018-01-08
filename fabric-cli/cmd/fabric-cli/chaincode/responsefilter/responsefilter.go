/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package responsefilter

import (
	"bytes"

	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/chaincode/invokeerror"
	cliconfig "github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/config"
	"github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/printer"
)

// Filter ensures that responses from all endorsers have the same status and payload
type Filter struct {
	verbose bool
	printer printer.Printer
}

// New returns a new response filter
func New(verbose bool, printer printer.Printer) *Filter {
	return &Filter{
		verbose: verbose,
		printer: printer,
	}
}

// ProcessTxProposalResponse is invoked after all responses are received from the endorsers. This function
// ensures that all responses contain the sames status and payload, otherwise an error is returned.
func (f *Filter) ProcessTxProposalResponse(responses []*apitxn.TransactionProposalResponse) ([]*apitxn.TransactionProposalResponse, error) {
	var txID string
	var lastStatus int32
	var lastEndorser string
	var lastPayload []byte

	if f.verbose {
		f.printer.PrintTxProposalResponses(responses, cliconfig.Config().PrintPayloadOnly())
	}

	for _, response := range responses {
		txID = response.Proposal.TxnID.ID
		if lastStatus == 0 {
			lastStatus = response.Status
			lastEndorser = response.Endorser
			lastPayload = response.TransactionProposalResult.ProposalResponse.Payload
			continue
		}
		if response.Status != lastStatus {
			cliconfig.Config().Logger().Debugf("Status [%d] from [%s] does not match status [%d] from endorser [%s] for TxID [%s].\n", response.Status, response.Endorser, lastStatus, lastEndorser, txID)
			return responses, invokeerror.Errorf(invokeerror.TransientError, "status [%d] from [%s] does not match status [%d] from endorser [%s] for TxID [%s]", response.Status, response.Endorser, lastStatus, lastEndorser, txID)
		}
		if bytes.Compare(response.TransactionProposalResult.ProposalResponse.Payload, lastPayload) != 0 {
			cliconfig.Config().Logger().Debugf("The payload from [%s] does not match the payload from endorser [%s] for TxID [%s].\n", response.Endorser, lastEndorser, txID)
			return responses, invokeerror.Errorf(invokeerror.TransientError, "the payload from [%s] does not match the payload from endorser [%s] for TxID [%s].\n", response.Endorser, lastEndorser, txID)
		}
	}

	if lastStatus != 200 {
		return responses, invokeerror.Errorf(invokeerror.TransientError, "error endorsing transaction [%s] - Status: [%d]", txID, lastStatus)
	}

	return responses, nil
}
