/*
Copyright IBM Corp. 2016 All Rights Reserved.

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

package cauthdsl

import (
	"fmt"

	"github.com/hyperledger/fabric/common/flogging"
	"github.com/hyperledger/fabric/msp"
	cb "github.com/hyperledger/fabric/protos/common"
	mb "github.com/hyperledger/fabric/protos/msp"
)

var cauthdslLogger = flogging.MustGetLogger("cauthdsl")

// compile recursively builds a go evaluatable function corresponding to the policy specified
func compile(policy *cb.SignaturePolicy, identities []*mb.MSPPrincipal, deserializer msp.IdentityDeserializer) (func([]*cb.SignedData, []bool) bool, error) {
	switch t := policy.Type.(type) {
	case *cb.SignaturePolicy_NOutOf_:
		policies := make([]func([]*cb.SignedData, []bool) bool, len(t.NOutOf.Rules))
		for i, policy := range t.NOutOf.Rules {
			compiledPolicy, err := compile(policy, identities, deserializer)
			if err != nil {
				return nil, err
			}
			policies[i] = compiledPolicy

		}
		return func(signedData []*cb.SignedData, used []bool) bool {
			cauthdslLogger.Debugf("Gate evaluation starts: (%v)", t)
			verified := int32(0)
			_used := make([]bool, len(used))
			for _, policy := range policies {
				copy(_used, used)
				if policy(signedData, _used) {
					verified++
					copy(used, _used)
				}
			}

			if verified >= t.NOutOf.N {
				cauthdslLogger.Debugf("Gate evaluation succeeds: (%v)", t)
			} else {
				cauthdslLogger.Debugf("Gate evaluation fails: (%v)", t)
			}

			return verified >= t.NOutOf.N
		}, nil
	case *cb.SignaturePolicy_SignedBy:
		if t.SignedBy < 0 || t.SignedBy >= int32(len(identities)) {
			return nil, fmt.Errorf("Identity index out of range, requested %v, but identies length is %d", t.SignedBy, len(identities))
		}
		signedByID := identities[t.SignedBy]
		return func(signedData []*cb.SignedData, used []bool) bool {
			cauthdslLogger.Debugf("Principal evaluation starts: (%v) (used %v)", t, used)
			for i, sd := range signedData {
				if used[i] {
					continue
				}
				identity, err := deserializer.DeserializeIdentity(sd.Identity)
				if err != nil {
					cauthdslLogger.Errorf("Principal deserialization failed: (%s) for identity %v", err, sd.Identity)
					continue
				}
				err = identity.SatisfiesPrincipal(signedByID)
				if err == nil {
					cauthdslLogger.Debugf("Principal matched by identity: (%v) for %v", t, sd.Identity)
					err = identity.Verify(sd.Data, sd.Signature)
					if err == nil {
						cauthdslLogger.Debugf("Principal evaluation succeeds: (%v) (used %v)", t, used)
						used[i] = true
						return true
					}
				} else {
					cauthdslLogger.Debugf("Identity (%v) does not satisfy principal: %s", sd.Identity, err)
				}
			}
			cauthdslLogger.Debugf("Principal evaluation fails: (%v) %v", t, used)
			return false
		}, nil
	default:
		return nil, fmt.Errorf("Unknown type: %T:%v", t, t)
	}
}
