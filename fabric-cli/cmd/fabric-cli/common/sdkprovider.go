package common

import (
	"github.com/hyperledger/fabric-sdk-go/def/fabapi/context/defprovider"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/def/fabapi/context"
	"github.com/hyperledger/fabric-sdk-go/def/fabapi/opt"
)

type providerFactory struct {
	context.SDKProviderFactory
	cnfg apiconfig.Config
}

// NewProviderFactory returns the SDK provider factory
func newProviderFactory(cnfg apiconfig.Config) *providerFactory {
	return &providerFactory{
		SDKProviderFactory: defprovider.NewDefaultProviderFactory(),
		cnfg:               cnfg,
	}
}

// NewConfigProvider creates a Config using the SDK's  implementation
func (f *providerFactory) NewConfigProvider(o opt.ConfigOpts, a opt.SDKOpts) (apiconfig.Config, error) {
	return f.cnfg, nil
}
