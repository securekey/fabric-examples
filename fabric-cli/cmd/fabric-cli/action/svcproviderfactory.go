/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package action

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/selection/dynamicselection"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/selection/staticselection"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/factory/defsvc"
	"github.com/pkg/errors"
	cliconfig "github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/config"
)

// serviceProviderFactory is configured with either static or dynamic selection provider
type serviceProviderFactory struct {
	defsvc.ProviderFactory
	selectionProvider fab.SelectionProvider
	channelUsers      []dynamicselection.ChannelUser
}

func newServiceProviderFactory() (*serviceProviderFactory, error) {
	sdkConfig, err := cliconfig.Provider()()
	if err != nil {
		return nil, errors.Errorf("Error initializing SDK config: %s", err)
	}

	org := cliconfig.Config().OrgID()
	if org == "" {
		orgValue, ok := sdkConfig.Lookup("client.organization")
		if !ok {
			return nil, errors.New("client.organization not found in SDK configuration")
		}
		org = orgValue.(string)
	}

	username := cliconfig.Config().UserName()
	if username == "" {
		username = defaultUser
	}

	f := serviceProviderFactory{}
	f.channelUsers = []dynamicselection.ChannelUser{
		{
			ChannelID: cliconfig.Config().ChannelID(),
			Username:  username,
			OrgName:   org,
		},
	}
	return &f, nil
}

// CreateSelectionProvider returns a new implementation of selection provider
func (f *serviceProviderFactory) CreateSelectionProvider(config fab.EndpointConfig) (fab.SelectionProvider, error) {

	var err error
	if f.selectionProvider == nil {
		switch cliconfig.Config().SelectionProvider() {

		case cliconfig.StaticSelectionProvider:
			cliconfig.Config().Logger().Debugf("Using static selection provider.\n")
			f.selectionProvider, err = staticselection.New(config)

		case cliconfig.DynamicSelectionProvider:
			cliconfig.Config().Logger().Debugf("Using dynamic selection provider.\n")
			f.selectionProvider, err = dynamicselection.New(config, f.channelUsers)

		default:
			return nil, errors.Errorf("invalid selection provider: %s", cliconfig.Config().SelectionProvider())
		}
	}
	return f.selectionProvider, err
}
