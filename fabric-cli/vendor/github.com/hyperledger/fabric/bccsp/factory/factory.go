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
package factory

import (
	"fmt"
	"sync"

	"github.com/hyperledger/fabric/bccsp"
	"github.com/op/go-logging"
)

var (
	// Default BCCSP
	defaultBCCSP bccsp.BCCSP

	// when InitFactories has not been called yet (should only happen
	// in test cases), use this BCCSP temporarily
	bootBCCSP bccsp.BCCSP

	// BCCSP Factories
	bccspMap map[string]bccsp.BCCSP

	// factories' Sync on Initialization
	factoriesInitOnce sync.Once
	bootBCCSPInitOnce sync.Once

	// Factories' Initialization Error
	factoriesInitError error

	logger = logging.MustGetLogger("BCCSP_FACTORY")
)

// BCCSPFactory is used to get instances of the BCCSP interface.
// A Factory has name used to address it.
type BCCSPFactory interface {

	// Name returns the name of this factory
	Name() string

	// Get returns an instance of BCCSP using opts.
	Get(opts *FactoryOpts) (bccsp.BCCSP, error)
}

// GetDefault returns a non-ephemeral (long-term) BCCSP
func GetDefault() bccsp.BCCSP {
	if defaultBCCSP == nil {
		logger.Warning("Before using BCCSP, please call InitFactories(). Falling back to bootBCCSP.")
		bootBCCSPInitOnce.Do(func() {
			var err error
			f := &SWFactory{}
			bootBCCSP, err = f.Get(&DefaultOpts)
			if err != nil {
				panic("BCCSP Internal error, failed initialization with DefaultOpts!")
			}
		})
		return bootBCCSP
	}
	return defaultBCCSP
}

// GetBCCSP returns a BCCSP created according to the options passed in input.
func GetBCCSP(name string) (bccsp.BCCSP, error) {
	return bccspMap[name], nil
}

// InitFactories must be called before using factory interfaces
// It is acceptable to call with config = nil, in which case
// some defaults will get used
// Error is returned only if defaultBCCSP cannot be found
func InitFactories(config *FactoryOpts) error {
	factoriesInitOnce.Do(func() {
		// Take some precautions on default opts
		if config == nil {
			config = &DefaultOpts
		}

		if config.ProviderName == "" {
			config.ProviderName = "SW"
		}

		if config.SwOpts == nil {
			config.SwOpts = DefaultOpts.SwOpts
		}

		// Initialize factories map
		bccspMap = make(map[string]bccsp.BCCSP)

		// Software-Based BCCSP
		if config.SwOpts != nil {
			f := &SWFactory{}
			err := initBCCSP(f, config)
			if err != nil {
				factoriesInitError = fmt.Errorf("[%s]", err)
			}
		}

		// PKCS11-Based BCCSP
		if config.Pkcs11Opts != nil {
			f := &PKCS11Factory{}
			err := initBCCSP(f, config)
			if err != nil {
				factoriesInitError = fmt.Errorf("%s\n[%s]", factoriesInitError, err)
			}
		}

		var ok bool
		defaultBCCSP, ok = bccspMap[config.ProviderName]
		if !ok {
			factoriesInitError = fmt.Errorf("%s\nCould not find default `%s` BCCSP", factoriesInitError, config.ProviderName)
		}
	})

	return factoriesInitError
}

func initBCCSP(f BCCSPFactory, config *FactoryOpts) error {
	csp, err := f.Get(config)
	if err != nil {
		return fmt.Errorf("Could not initialize BCCSP %s [%s]", f.Name(), err)
	}

	logger.Debugf("Initialize BCCSP [%s]", f.Name())
	bccspMap[f.Name()] = csp
	return nil
}
