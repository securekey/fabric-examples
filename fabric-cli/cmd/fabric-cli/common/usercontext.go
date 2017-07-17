/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package common

import (
	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
)

// UserContext restores a previously saved user context
// TODO: This is a temporary construct which should be removed when the SDK introduces sessions
type UserContext struct {
	client      apifabclient.FabricClient
	restoreUser apifabclient.User
}

func newUserContext(client apifabclient.FabricClient) *UserContext {
	return &UserContext{
		client:      client,
		restoreUser: client.UserContext(),
	}
}

// Restore restores the user context
func (c *UserContext) Restore() {
	c.client.SetUserContext(c.restoreUser)
}
