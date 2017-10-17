/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package msp

import (
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/bccsp"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/bccsp/factory"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/msp"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type OrganizationalUnitIdentifiersConfiguration struct {
	Certificate                  string `yaml:"Certificate,omitempty"`
	OrganizationalUnitIdentifier string `yaml:"OrganizationalUnitIdentifier,omitempty"`
}

type Configuration struct {
	OrganizationalUnitIdentifiers []*OrganizationalUnitIdentifiersConfiguration `yaml:"OrganizationalUnitIdentifiers,omitempty"`
}

func readFile(file string) ([]byte, error) {
	fileCont, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, errors.Wrapf(err, "could not read file %s", file)
	}

	return fileCont, nil
}

func readPemFile(file string) ([]byte, error) {
	bytes, err := readFile(file)
	if err != nil {
		return nil, errors.Wrapf(err, "reading from file %s failed", file)
	}

	b, _ := pem.Decode(bytes)
	if b == nil { // TODO: also check that the type is what we expect (cert vs key..)
		return nil, errors.Errorf("no pem content for file %s", file)
	}

	return bytes, nil
}

func getPemMaterialFromDir(dir string) ([][]byte, error) {
	mspLogger.Debugf("Reading directory %s", dir)

	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return nil, err
	}

	content := make([][]byte, 0)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, errors.Wrapf(err, "could not read directory %s", err)
	}

	for _, f := range files {
		if f.IsDir() {
			continue
		}

		fullName := filepath.Join(dir, string(filepath.Separator), f.Name())
		mspLogger.Debugf("Inspecting file %s", fullName)

		item, err := readPemFile(fullName)
		if err != nil {
			mspLogger.Warningf("Failed reading file %s: %s", fullName, err)
			continue
		}

		content = append(content, item)
	}

	return content, nil
}

const (
	cacerts              = "cacerts"
	admincerts           = "admincerts"
	signcerts            = "signcerts"
	keystore             = "keystore"
	intermediatecerts    = "intermediatecerts"
	crlsfolder           = "crls"
	configfilename       = "config.yaml"
	tlscacerts           = "tlscacerts"
	tlsintermediatecerts = "tlsintermediatecerts"
)

func SetupBCCSPKeystoreConfig(bccspConfig *factory.FactoryOpts, keystoreDir string) *factory.FactoryOpts {
	if bccspConfig == nil {
		bccspConfig = factory.GetDefaultOpts()
	}

	if bccspConfig.ProviderName == "SW" {
		if bccspConfig.SwOpts == nil {
			bccspConfig.SwOpts = factory.GetDefaultOpts().SwOpts
		}

		// Only override the KeyStorePath if it was left empty
		if bccspConfig.SwOpts.FileKeystore == nil ||
			bccspConfig.SwOpts.FileKeystore.KeyStorePath == "" {
			bccspConfig.SwOpts.Ephemeral = false
			bccspConfig.SwOpts.FileKeystore = &factory.FileKeystoreOpts{KeyStorePath: keystoreDir}
		}
	}

	return bccspConfig
}

func GetLocalMspConfig(dir string, bccspConfig *factory.FactoryOpts, ID string) (*msp.MSPConfig, error) {
	signcertDir := filepath.Join(dir, signcerts)
	keystoreDir := filepath.Join(dir, keystore)
	bccspConfig = SetupBCCSPKeystoreConfig(bccspConfig, keystoreDir)

	err := factory.InitFactories(bccspConfig)
	if err != nil {
		return nil, errors.WithMessage(err, "could not initialize BCCSP Factories")
	}

	signcert, err := getPemMaterialFromDir(signcertDir)
	if err != nil || len(signcert) == 0 {
		return nil, errors.Wrapf(err, "could not load a valid signer certificate from directory %s", signcertDir)
	}

	/* FIXME: for now we're making the following assumptions
	1) there is exactly one signing cert
	2) BCCSP's KeyStore has the private key that matches SKI of
	   signing cert
	*/

	sigid := &msp.SigningIdentityInfo{PublicSigner: signcert[0], PrivateSigner: nil}

	return getMspConfig(dir, ID, sigid)
}

func GetVerifyingMspConfig(dir string, ID string) (*msp.MSPConfig, error) {
	return getMspConfig(dir, ID, nil)
}

func getMspConfig(dir string, ID string, sigid *msp.SigningIdentityInfo) (*msp.MSPConfig, error) {
	cacertDir := filepath.Join(dir, cacerts)
	admincertDir := filepath.Join(dir, admincerts)
	intermediatecertsDir := filepath.Join(dir, intermediatecerts)
	crlsDir := filepath.Join(dir, crlsfolder)
	configFile := filepath.Join(dir, configfilename)
	tlscacertDir := filepath.Join(dir, tlscacerts)
	tlsintermediatecertsDir := filepath.Join(dir, tlsintermediatecerts)

	cacerts, err := getPemMaterialFromDir(cacertDir)
	if err != nil || len(cacerts) == 0 {
		return nil, errors.WithMessage(err, fmt.Sprintf("could not load a valid ca certificate from directory %s", cacertDir))
	}

	admincert, err := getPemMaterialFromDir(admincertDir)
	if err != nil || len(admincert) == 0 {
		return nil, errors.WithMessage(err, fmt.Sprintf("could not load a valid admin certificate from directory %s", admincertDir))
	}

	intermediatecerts, err := getPemMaterialFromDir(intermediatecertsDir)
	if os.IsNotExist(err) {
		mspLogger.Debugf("Intermediate certs folder not found at [%s]. Skipping. [%s]", intermediatecertsDir, err)
	} else if err != nil {
		return nil, errors.WithMessage(err, fmt.Sprintf("failed loading intermediate ca certs at [%s]", intermediatecertsDir))
	}

	tlsCACerts, err := getPemMaterialFromDir(tlscacertDir)
	tlsIntermediateCerts := [][]byte{}
	if os.IsNotExist(err) {
		mspLogger.Debugf("TLS CA certs folder not found at [%s]. Skipping and ignoring TLS intermediate CA folder. [%s]", tlsintermediatecertsDir, err)
	} else if err != nil {
		return nil, errors.WithMessage(err, fmt.Sprintf("failed loading TLS ca certs at [%s]", tlsintermediatecertsDir))
	} else if len(tlsCACerts) != 0 {
		tlsIntermediateCerts, err = getPemMaterialFromDir(tlsintermediatecertsDir)
		if os.IsNotExist(err) {
			mspLogger.Debugf("TLS intermediate certs folder not found at [%s]. Skipping. [%s]", tlsintermediatecertsDir, err)
		} else if err != nil {
			return nil, errors.WithMessage(err, fmt.Sprintf("failed loading TLS intermediate ca certs at [%s]", tlsintermediatecertsDir))
		}
	} else {
		mspLogger.Debugf("TLS CA certs folder at [%s] is empty. Skipping.", tlsintermediatecertsDir)
	}

	crls, err := getPemMaterialFromDir(crlsDir)
	if os.IsNotExist(err) {
		mspLogger.Debugf("crls folder not found at [%s]. Skipping. [%s]", crlsDir, err)
	} else if err != nil {
		return nil, errors.WithMessage(err, fmt.Sprintf("failed loading crls at [%s]", crlsDir))
	}

	// Load configuration file
	// if the configuration file is there then load it
	// otherwise skip it
	var ouis []*msp.FabricOUIdentifier
	_, err = os.Stat(configFile)
	if err == nil {
		// load the file, if there is a failure in loading it then
		// return an error
		raw, err := ioutil.ReadFile(configFile)
		if err != nil {
			return nil, errors.Wrapf(err, "failed loading configuration file at [%s]", configFile)
		}

		configuration := Configuration{}
		err = yaml.Unmarshal(raw, &configuration)
		if err != nil {
			return nil, errors.Wrapf(err, "failed unmarshalling configuration file at [%s]", configFile)
		}

		// Prepare OrganizationalUnitIdentifiers
		if len(configuration.OrganizationalUnitIdentifiers) > 0 {
			for _, ouID := range configuration.OrganizationalUnitIdentifiers {
				f := filepath.Join(dir, ouID.Certificate)
				raw, err = ioutil.ReadFile(f)
				if err != nil {
					return nil, errors.Wrapf(err, "failed loading OrganizationalUnit certificate at [%s]", f)
				}
				oui := &msp.FabricOUIdentifier{
					Certificate:                  raw,
					OrganizationalUnitIdentifier: ouID.OrganizationalUnitIdentifier,
				}
				ouis = append(ouis, oui)
			}
		}
	} else {
		mspLogger.Debugf("MSP configuration file not found at [%s]: [%s]", configFile, err)
	}

	// Set FabricCryptoConfig
	cryptoConfig := &msp.FabricCryptoConfig{
		SignatureHashFamily:            bccsp.SHA2,
		IdentityIdentifierHashFunction: bccsp.SHA256,
	}

	// Compose FabricMSPConfig
	fmspconf := &msp.FabricMSPConfig{
		Admins:            admincert,
		RootCerts:         cacerts,
		IntermediateCerts: intermediatecerts,
		SigningIdentity:   sigid,
		Name:              ID,
		OrganizationalUnitIdentifiers: ouis,
		RevocationList:                crls,
		CryptoConfig:                  cryptoConfig,
		TlsRootCerts:                  tlsCACerts,
		TlsIntermediateCerts:          tlsIntermediateCerts,
	}

	fmpsjs, _ := proto.Marshal(fmspconf)

	mspconf := &msp.MSPConfig{Config: fmpsjs, Type: int32(FABRIC)}

	return mspconf, nil
}

// IdemixConfig is the filename of the idemix msp config file
const IdemixConfig = "idemixmspconfig"

// GetIdemixMspConfig returns the configuration for the Idemix MSP
func GetIdemixMspConfig(dir string) (*msp.MSPConfig, error) {
	confStringBytes, err := readFile(filepath.Join(dir, IdemixConfig))
	if err != nil {
		return nil, errors.Wrapf(err, "error reading idemix config file")
	}
	config := &msp.IdemixMSPConfig{}
	err = proto.UnmarshalText(string(confStringBytes), config)
	if err != nil {
		return nil, errors.Wrapf(err, "error unmarshalling idemix config")
	}
	confBytes, err := proto.Marshal(config)
	if err != nil {
		return nil, errors.Wrapf(err, "error creating idemix config")
	}

	return &msp.MSPConfig{Config: confBytes, Type: int32(IDEMIX)}, nil
}
