/*
Copyright SecureKey Technologies Inc. All Rights Reserved.


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at


      http://www.apache.org/licenses/LICENSE-2.0


Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
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
