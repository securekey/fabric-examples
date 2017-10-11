package common

import (
	"github.com/hyperledger/fabric-sdk-go/def/fabapi/context/defprovider"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/def/fabapi/context"
	"github.com/hyperledger/fabric-sdk-go/def/fabapi/opt"
)

type ProviderFactory struct {
	context.SDKProviderFactory
	cnfg apiconfig.Config
}

// NewProviderFactory returns the SDK provider factory
func NewProviderFactory(cnfg apiconfig.Config) *ProviderFactory {
	return &ProviderFactory{
		SDKProviderFactory: defprovider.NewDefaultProviderFactory(),
		cnfg:               cnfg,
	}
}

// NewConfigProvider creates a Config using the SDK's  implementation
func (f *ProviderFactory) NewConfigProvider(o opt.ConfigOpts, a opt.SDKOpts) (apiconfig.Config, error) {
	return f.cnfg, nil
}
