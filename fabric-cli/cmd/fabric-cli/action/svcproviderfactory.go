/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package action

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/selection/dynamicselection"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/selection/fabricselection"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/selection/staticselection"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	contextApi "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/factory/defsvc"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/provider/chpvdr"
	"github.com/pkg/errors"
	cliconfig "github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/config"
)

// serviceProviderFactory is configured with either static or dynamic selection provider
type serviceProviderFactory struct {
	defsvc.ProviderFactory
}

func newServiceProviderFactory() (*serviceProviderFactory, error) {
	return &serviceProviderFactory{}, nil
}

type fabricSelectionChannelProvider struct {
	fab.ChannelProvider
	service   fab.ChannelService
	selection fab.SelectionService
}

type fabricSelectionChannelService struct {
	fab.ChannelService
	selection fab.SelectionService
}

// CreateChannelProvider returns a new default implementation of channel provider
func (f *serviceProviderFactory) CreateChannelProvider(config fab.EndpointConfig, opts ...options.Opt) (fab.ChannelProvider, error) {
	chProvider, err := chpvdr.New(config, opts...)
	if err != nil {
		return nil, err
	}
	return &fabricSelectionChannelProvider{
		ChannelProvider: chProvider,
	}, nil
}

type closable interface {
	Close()
}

// Close frees resources and caches.
func (cp *fabricSelectionChannelProvider) Close() {
	if c, ok := cp.ChannelProvider.(closable); ok {
		c.Close()
	}
	if cp.selection != nil {
		if c, ok := cp.selection.(closable); ok {
			c.Close()
		}
	}
}

type providerInit interface {
	Initialize(providers contextApi.Providers) error
}

func (cp *fabricSelectionChannelProvider) Initialize(providers contextApi.Providers) error {
	if init, ok := cp.ChannelProvider.(providerInit); ok {
		return init.Initialize(providers)
	}
	return nil
}

// ChannelService creates a ChannelService for an identity
func (cp *fabricSelectionChannelProvider) ChannelService(ctx fab.ClientContext, channelID string) (fab.ChannelService, error) {
	chService, err := cp.ChannelProvider.ChannelService(ctx, channelID)
	if err != nil {
		return nil, err
	}

	discovery, err := chService.Discovery()
	if err != nil {
		return nil, err
	}

	if cp.selection == nil {
		switch cliconfig.Config().SelectionProvider() {
		case cliconfig.StaticSelectionProvider:
			cliconfig.Config().Logger().Debugf("Using static selection provider.")
			cp.selection, err = staticselection.NewService(discovery)
		case cliconfig.DynamicSelectionProvider:
			cliconfig.Config().Logger().Debugf("Using dynamic selection provider.")
			cp.selection, err = dynamicselection.NewService(ctx, channelID, discovery)
		case cliconfig.FabricSelectionProvider:
			cliconfig.Config().Logger().Debugf("Using Fabric selection provider.")
			cp.selection, err = fabricselection.New(ctx, channelID, discovery)
		default:
			return nil, errors.Errorf("invalid selection provider: %s", cliconfig.Config().SelectionProvider())
		}

		if err != nil {
			return nil, err
		}
	}

	return &fabricSelectionChannelService{
		ChannelService: chService,
		selection:      cp.selection,
	}, nil
}

func (cs *fabricSelectionChannelService) Selection() (fab.SelectionService, error) {
	return cs.selection, nil
}
