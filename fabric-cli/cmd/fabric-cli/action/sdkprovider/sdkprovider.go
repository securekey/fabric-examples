/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sdkprovider

import (
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/def/fabapi/context"
	"github.com/hyperledger/fabric-sdk-go/def/fabapi/context/defprovider"
	"github.com/hyperledger/fabric-sdk-go/def/fabapi/opt"
	cliconfig "github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli/config"
)

// Factory is an SDK provider factory implementation that overrides
// the COnfig profider, Selection provider, and Disacovery provider.
type Factory struct {
	context.SDKProviderFactory
	cnfg              apiconfig.Config
	selectionProvider apifabclient.SelectionProvider
	discoveryProvider apifabclient.DiscoveryProvider
}

// New creates a new Factory
func New(selectionProvider apifabclient.SelectionProvider, discoveryProvider apifabclient.DiscoveryProvider) *Factory {
	return &Factory{
		SDKProviderFactory: defprovider.NewDefaultProviderFactory(),
		cnfg:               cliconfig.Config(),
		selectionProvider:  selectionProvider,
		discoveryProvider:  discoveryProvider,
	}
}

// NewConfigProvider creates a Config using the SDK's  implementation
func (f *Factory) NewConfigProvider(o opt.ConfigOpts, a opt.SDKOpts) (apiconfig.Config, error) {
	return f.cnfg, nil
}

// NewSelectionProvider returns a new implementation of dynamic selection provider
func (f *Factory) NewSelectionProvider(config apiconfig.Config) (apifabclient.SelectionProvider, error) {
	return f.selectionProvider, nil
}

// NewDiscoveryProvider returns a new DiscoveryProvider
func (f *Factory) NewDiscoveryProvider(config apiconfig.Config) (apifabclient.DiscoveryProvider, error) {
	return f.discoveryProvider, nil
}
