/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chaincode

import (
	"fmt"

	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
)

// QueryChaincode performs a query on multiple peers. An error is returned if at least one of the peers
// returns an error. The payload response from one of the peers is returned.
func QueryChaincode(channel apifabclient.Channel, peers []apifabclient.Peer, chaincodeID string, channelID string, fctn string, args []string) ([]byte, error) {
	targets := make([]apitxn.ProposalProcessor, len(peers))
	for i, p := range peers {
		targets[i] = p
	}

	request := apitxn.ChaincodeInvokeRequest{
		Targets:      targets,
		Fcn:          fctn,
		Args:         args,
		TransientMap: nil,
		ChaincodeID:  chaincodeID,
	}
	transactionProposalResponses, _, err := channel.SendTransactionProposal(request)
	if err != nil {
		return nil, fmt.Errorf("CreateTransactionProposal returned error: %v", err)
	}

	var responses []*apitxn.TransactionProposalResponse
	for _, v := range transactionProposalResponses {
		if v.Err != nil {
			return nil, fmt.Errorf("query Endorser %s return error: %v", v.Endorser, v.Err)
		}
		responses = append(responses, v)
	}

	if len(responses) == 0 {
		return nil, fmt.Errorf("no response from endorsers")
	}

	return responses[0].ProposalResponse.Response.Payload, nil
}
