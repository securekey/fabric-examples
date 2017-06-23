/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package common

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/hyperledger/fabric-sdk-go/api"
)

// Base64URLEncode encodes the byte array into a base64 string
func Base64URLEncode(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}

// Base64URLDecode decodes the base64 string into a byte array
func Base64URLDecode(data string) ([]byte, error) {
	//check if it has padding or not
	if strings.HasSuffix(data, "=") {
		return base64.URLEncoding.DecodeString(data)
	}
	return base64.RawURLEncoding.DecodeString(data)
}

// QueryChaincode performs a query on multiple peers. An error is returned if at least one of the peers
// returns an error. The payload response from one of the peers is returned.
func QueryChaincode(channel api.Channel, peers []api.Peer, chaincodeID string, channelID string, args []string) ([]byte, error) {
	signedProposal, err := channel.CreateTransactionProposal(chaincodeID, channelID, args, true, nil)
	if err != nil {
		return nil, fmt.Errorf("CreateTransactionProposal returned error: %v", err)
	}

	transactionProposalResponses, err := channel.SendTransactionProposal(signedProposal, 0, peers)
	if err != nil {
		return nil, fmt.Errorf("CreateTransactionProposal returned error: %v", err)
	}

	var responses []*api.TransactionProposalResponse
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
