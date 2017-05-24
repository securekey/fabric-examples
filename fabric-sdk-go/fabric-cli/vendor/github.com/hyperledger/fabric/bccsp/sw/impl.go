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
package sw

import (
	"crypto/ecdsa"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"

	"crypto/rsa"

	"hash"

	"crypto/x509"

	"crypto/hmac"

	"crypto/elliptic"
	"crypto/sha256"
	"crypto/sha512"

	"github.com/hyperledger/fabric/bccsp"
	"github.com/hyperledger/fabric/bccsp/utils"
	"github.com/op/go-logging"
	"golang.org/x/crypto/sha3"
)

var (
	logger = logging.MustGetLogger("SW_BCCSP")
)

// NewDefaultSecurityLevel returns a new instance of the software-based BCCSP
// at security level 256, hash family SHA2 and using FolderBasedKeyStore as KeyStore.
func NewDefaultSecurityLevel(keyStorePath string) (bccsp.BCCSP, error) {
	ks := &fileBasedKeyStore{}
	if err := ks.Init(nil, keyStorePath, false); err != nil {
		return nil, fmt.Errorf("Failed initializing key store [%s]", err)
	}

	return New(256, "SHA2", ks)
}

// NewDefaultSecurityLevel returns a new instance of the software-based BCCSP
// at security level 256, hash family SHA2 and using the passed KeyStore.
func NewDefaultSecurityLevelWithKeystore(keyStore bccsp.KeyStore) (bccsp.BCCSP, error) {
	return New(256, "SHA2", keyStore)
}

// New returns a new instance of the software-based BCCSP
// set at the passed security level, hash family and KeyStore.
func New(securityLevel int, hashFamily string, keyStore bccsp.KeyStore) (bccsp.BCCSP, error) {
	// Init config
	conf := &config{}
	err := conf.setSecurityLevel(securityLevel, hashFamily)
	if err != nil {
		return nil, fmt.Errorf("Failed initializing configuration [%s]", err)
	}

	// Check KeyStore
	if keyStore == nil {
		return nil, errors.New("Invalid bccsp.KeyStore instance. It must be different from nil.")
	}

	return &impl{conf, keyStore}, nil
}

// SoftwareBasedBCCSP is the software-based implementation of the BCCSP.
type impl struct {
	conf *config
	ks   bccsp.KeyStore
}

// KeyGen generates a key using opts.
func (csp *impl) KeyGen(opts bccsp.KeyGenOpts) (k bccsp.Key, err error) {
	// Validate arguments
	if opts == nil {
		return nil, errors.New("Invalid Opts parameter. It must not be nil.")
	}

	// Parse algorithm
	switch opts.(type) {
	case *bccsp.ECDSAKeyGenOpts:
		lowLevelKey, err := ecdsa.GenerateKey(csp.conf.ellipticCurve, rand.Reader)
		if err != nil {
			return nil, fmt.Errorf("Failed generating ECDSA key [%s]", err)
		}

		k = &ecdsaPrivateKey{lowLevelKey}

	case *bccsp.ECDSAP256KeyGenOpts:
		lowLevelKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			return nil, fmt.Errorf("Failed generating ECDSA P256 key [%s]", err)
		}

		k = &ecdsaPrivateKey{lowLevelKey}

	case *bccsp.ECDSAP384KeyGenOpts:
		lowLevelKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
		if err != nil {
			return nil, fmt.Errorf("Failed generating ECDSA P384 key [%s]", err)
		}

		k = &ecdsaPrivateKey{lowLevelKey}

	case *bccsp.AESKeyGenOpts:
		lowLevelKey, err := GetRandomBytes(csp.conf.aesBitLength)

		if err != nil {
			return nil, fmt.Errorf("Failed generating AES key [%s]", err)
		}

		k = &aesPrivateKey{lowLevelKey, false}

	case *bccsp.AES256KeyGenOpts:
		lowLevelKey, err := GetRandomBytes(32)

		if err != nil {
			return nil, fmt.Errorf("Failed generating AES 256 key [%s]", err)
		}

		k = &aesPrivateKey{lowLevelKey, false}

	case *bccsp.AES192KeyGenOpts:
		lowLevelKey, err := GetRandomBytes(24)

		if err != nil {
			return nil, fmt.Errorf("Failed generating AES 192 key [%s]", err)
		}

		k = &aesPrivateKey{lowLevelKey, false}

	case *bccsp.AES128KeyGenOpts:
		lowLevelKey, err := GetRandomBytes(16)

		if err != nil {
			return nil, fmt.Errorf("Failed generating AES 128 key [%s]", err)
		}

		k = &aesPrivateKey{lowLevelKey, false}

	case *bccsp.RSAKeyGenOpts:
		lowLevelKey, err := rsa.GenerateKey(rand.Reader, csp.conf.rsaBitLength)

		if err != nil {
			return nil, fmt.Errorf("Failed generating RSA key [%s]", err)
		}

		k = &rsaPrivateKey{lowLevelKey}

	case *bccsp.RSA1024KeyGenOpts:
		lowLevelKey, err := rsa.GenerateKey(rand.Reader, 1024)

		if err != nil {
			return nil, fmt.Errorf("Failed generating RSA 1024 key [%s]", err)
		}

		k = &rsaPrivateKey{lowLevelKey}

	case *bccsp.RSA2048KeyGenOpts:
		lowLevelKey, err := rsa.GenerateKey(rand.Reader, 2048)

		if err != nil {
			return nil, fmt.Errorf("Failed generating RSA 2048 key [%s]", err)
		}

		k = &rsaPrivateKey{lowLevelKey}

	case *bccsp.RSA3072KeyGenOpts:
		lowLevelKey, err := rsa.GenerateKey(rand.Reader, 3072)

		if err != nil {
			return nil, fmt.Errorf("Failed generating RSA 3072 key [%s]", err)
		}

		k = &rsaPrivateKey{lowLevelKey}

	case *bccsp.RSA4096KeyGenOpts:
		lowLevelKey, err := rsa.GenerateKey(rand.Reader, 4096)

		if err != nil {
			return nil, fmt.Errorf("Failed generating RSA 4096 key [%s]", err)
		}

		k = &rsaPrivateKey{lowLevelKey}

	default:
		return nil, fmt.Errorf("Unrecognized KeyGenOpts provided [%s]", opts.Algorithm())
	}

	// If the key is not Ephemeral, store it.
	if !opts.Ephemeral() {
		// Store the key
		err = csp.ks.StoreKey(k)
		if err != nil {
			return nil, fmt.Errorf("Failed storing key [%s]. [%s]", opts.Algorithm(), err)
		}
	}

	return k, nil
}

// KeyDeriv derives a key from k using opts.
// The opts argument should be appropriate for the primitive used.
func (csp *impl) KeyDeriv(k bccsp.Key, opts bccsp.KeyDerivOpts) (dk bccsp.Key, err error) {
	// Validate arguments
	if k == nil {
		return nil, errors.New("Invalid Key. It must not be nil.")
	}

	// Derive key
	switch k.(type) {
	case *ecdsaPublicKey:
		// Validate opts
		if opts == nil {
			return nil, errors.New("Invalid Opts parameter. It must not be nil.")
		}

		ecdsaK := k.(*ecdsaPublicKey)

		switch opts.(type) {

		// Re-randomized an ECDSA private key
		case *bccsp.ECDSAReRandKeyOpts:
			reRandOpts := opts.(*bccsp.ECDSAReRandKeyOpts)
			tempSK := &ecdsa.PublicKey{
				Curve: ecdsaK.pubKey.Curve,
				X:     new(big.Int),
				Y:     new(big.Int),
			}

			var k = new(big.Int).SetBytes(reRandOpts.ExpansionValue())
			var one = new(big.Int).SetInt64(1)
			n := new(big.Int).Sub(ecdsaK.pubKey.Params().N, one)
			k.Mod(k, n)
			k.Add(k, one)

			// Compute temporary public key
			tempX, tempY := ecdsaK.pubKey.ScalarBaseMult(k.Bytes())
			tempSK.X, tempSK.Y = tempSK.Add(
				ecdsaK.pubKey.X, ecdsaK.pubKey.Y,
				tempX, tempY,
			)

			// Verify temporary public key is a valid point on the reference curve
			isOn := tempSK.Curve.IsOnCurve(tempSK.X, tempSK.Y)
			if !isOn {
				return nil, errors.New("Failed temporary public key IsOnCurve check.")
			}

			reRandomizedKey := &ecdsaPublicKey{tempSK}

			// If the key is not Ephemeral, store it.
			if !opts.Ephemeral() {
				// Store the key
				err = csp.ks.StoreKey(reRandomizedKey)
				if err != nil {
					return nil, fmt.Errorf("Failed storing ECDSA key [%s]", err)
				}
			}

			return reRandomizedKey, nil

		default:
			return nil, fmt.Errorf("Unrecognized KeyDerivOpts provided [%s]", opts.Algorithm())

		}
	case *ecdsaPrivateKey:
		// Validate opts
		if opts == nil {
			return nil, errors.New("Invalid Opts parameter. It must not be nil.")
		}

		ecdsaK := k.(*ecdsaPrivateKey)

		switch opts.(type) {

		// Re-randomized an ECDSA private key
		case *bccsp.ECDSAReRandKeyOpts:
			reRandOpts := opts.(*bccsp.ECDSAReRandKeyOpts)
			tempSK := &ecdsa.PrivateKey{
				PublicKey: ecdsa.PublicKey{
					Curve: ecdsaK.privKey.Curve,
					X:     new(big.Int),
					Y:     new(big.Int),
				},
				D: new(big.Int),
			}

			var k = new(big.Int).SetBytes(reRandOpts.ExpansionValue())
			var one = new(big.Int).SetInt64(1)
			n := new(big.Int).Sub(ecdsaK.privKey.Params().N, one)
			k.Mod(k, n)
			k.Add(k, one)

			tempSK.D.Add(ecdsaK.privKey.D, k)
			tempSK.D.Mod(tempSK.D, ecdsaK.privKey.PublicKey.Params().N)

			// Compute temporary public key
			tempX, tempY := ecdsaK.privKey.PublicKey.ScalarBaseMult(k.Bytes())
			tempSK.PublicKey.X, tempSK.PublicKey.Y =
				tempSK.PublicKey.Add(
					ecdsaK.privKey.PublicKey.X, ecdsaK.privKey.PublicKey.Y,
					tempX, tempY,
				)

			// Verify temporary public key is a valid point on the reference curve
			isOn := tempSK.Curve.IsOnCurve(tempSK.PublicKey.X, tempSK.PublicKey.Y)
			if !isOn {
				return nil, errors.New("Failed temporary public key IsOnCurve check.")
			}

			reRandomizedKey := &ecdsaPrivateKey{tempSK}

			// If the key is not Ephemeral, store it.
			if !opts.Ephemeral() {
				// Store the key
				err = csp.ks.StoreKey(reRandomizedKey)
				if err != nil {
					return nil, fmt.Errorf("Failed storing ECDSA key [%s]", err)
				}
			}

			return reRandomizedKey, nil

		default:
			return nil, fmt.Errorf("Unrecognized KeyDerivOpts provided [%s]", opts.Algorithm())

		}
	case *aesPrivateKey:
		// Validate opts
		if opts == nil {
			return nil, errors.New("Invalid Opts parameter. It must not be nil.")
		}

		aesK := k.(*aesPrivateKey)

		switch opts.(type) {
		case *bccsp.HMACTruncated256AESDeriveKeyOpts:
			hmacOpts := opts.(*bccsp.HMACTruncated256AESDeriveKeyOpts)

			mac := hmac.New(csp.conf.hashFunction, aesK.privKey)
			mac.Write(hmacOpts.Argument())
			hmacedKey := &aesPrivateKey{mac.Sum(nil)[:csp.conf.aesBitLength], false}

			// If the key is not Ephemeral, store it.
			if !opts.Ephemeral() {
				// Store the key
				err = csp.ks.StoreKey(hmacedKey)
				if err != nil {
					return nil, fmt.Errorf("Failed storing ECDSA key [%s]", err)
				}
			}

			return hmacedKey, nil

		case *bccsp.HMACDeriveKeyOpts:

			hmacOpts := opts.(*bccsp.HMACDeriveKeyOpts)

			mac := hmac.New(csp.conf.hashFunction, aesK.privKey)
			mac.Write(hmacOpts.Argument())
			hmacedKey := &aesPrivateKey{mac.Sum(nil), true}

			// If the key is not Ephemeral, store it.
			if !opts.Ephemeral() {
				// Store the key
				err = csp.ks.StoreKey(hmacedKey)
				if err != nil {
					return nil, fmt.Errorf("Failed storing ECDSA key [%s]", err)
				}
			}

			return hmacedKey, nil

		default:
			return nil, fmt.Errorf("Unrecognized KeyDerivOpts provided [%s]", opts.Algorithm())

		}

	default:
		return nil, fmt.Errorf("Key type not recognized [%s]", k)
	}
}

// KeyImport imports a key from its raw representation using opts.
// The opts argument should be appropriate for the primitive used.
func (csp *impl) KeyImport(raw interface{}, opts bccsp.KeyImportOpts) (k bccsp.Key, err error) {
	// Validate arguments
	if raw == nil {
		return nil, errors.New("Invalid raw. Cannot be nil")
	}

	if opts == nil {
		return nil, errors.New("Invalid Opts parameter. It must not be nil.")
	}

	switch opts.(type) {

	case *bccsp.AES256ImportKeyOpts:
		aesRaw, ok := raw.([]byte)
		if !ok {
			return nil, errors.New("[AES256ImportKeyOpts] Invalid raw material. Expected byte array.")
		}

		if len(aesRaw) != 32 {
			return nil, fmt.Errorf("[AES256ImportKeyOpts] Invalid Key Length [%d]. Must be 32 bytes", len(aesRaw))
		}

		aesK := &aesPrivateKey{utils.Clone(aesRaw), false}

		// If the key is not Ephemeral, store it.
		if !opts.Ephemeral() {
			// Store the key
			err = csp.ks.StoreKey(aesK)
			if err != nil {
				return nil, fmt.Errorf("Failed storing AES key [%s]", err)
			}
		}

		return aesK, nil

	case *bccsp.HMACImportKeyOpts:
		aesRaw, ok := raw.([]byte)
		if !ok {
			return nil, errors.New("[HMACImportKeyOpts] Invalid raw material. Expected byte array.")
		}

		if len(aesRaw) == 0 {
			return nil, errors.New("[HMACImportKeyOpts] Invalid raw. It must not be nil.")
		}

		aesK := &aesPrivateKey{utils.Clone(aesRaw), false}

		// If the key is not Ephemeral, store it.
		if !opts.Ephemeral() {
			// Store the key
			err = csp.ks.StoreKey(aesK)
			if err != nil {
				return nil, fmt.Errorf("Failed storing AES key [%s]", err)
			}
		}

		return aesK, nil

	case *bccsp.ECDSAPKIXPublicKeyImportOpts:
		der, ok := raw.([]byte)
		if !ok {
			return nil, errors.New("[ECDSAPKIXPublicKeyImportOpts] Invalid raw material. Expected byte array.")
		}

		if len(der) == 0 {
			return nil, errors.New("[ECDSAPKIXPublicKeyImportOpts] Invalid raw. It must not be nil.")
		}

		lowLevelKey, err := utils.DERToPublicKey(der)
		if err != nil {
			return nil, fmt.Errorf("Failed converting PKIX to ECDSA public key [%s]", err)
		}

		ecdsaPK, ok := lowLevelKey.(*ecdsa.PublicKey)
		if !ok {
			return nil, errors.New("Failed casting to ECDSA public key. Invalid raw material.")
		}

		k = &ecdsaPublicKey{ecdsaPK}

		// If the key is not Ephemeral, store it.
		if !opts.Ephemeral() {
			// Store the key
			err = csp.ks.StoreKey(k)
			if err != nil {
				return nil, fmt.Errorf("Failed storing ECDSA key [%s]", err)
			}
		}

		return k, nil

	case *bccsp.ECDSAPrivateKeyImportOpts:
		der, ok := raw.([]byte)
		if !ok {
			return nil, errors.New("[ECDSADERPrivateKeyImportOpts] Invalid raw material. Expected byte array.")
		}

		if len(der) == 0 {
			return nil, errors.New("[ECDSADERPrivateKeyImportOpts] Invalid raw. It must not be nil.")
		}

		lowLevelKey, err := utils.DERToPrivateKey(der)
		if err != nil {
			return nil, fmt.Errorf("Failed converting PKIX to ECDSA public key [%s]", err)
		}

		ecdsaSK, ok := lowLevelKey.(*ecdsa.PrivateKey)
		if !ok {
			return nil, errors.New("Failed casting to ECDSA public key. Invalid raw material.")
		}

		k = &ecdsaPrivateKey{ecdsaSK}

		// If the key is not Ephemeral, store it.
		if !opts.Ephemeral() {
			// Store the key
			err = csp.ks.StoreKey(k)
			if err != nil {
				return nil, fmt.Errorf("Failed storing ECDSA key [%s]", err)
			}
		}

		return k, nil

	case *bccsp.ECDSAGoPublicKeyImportOpts:
		lowLevelKey, ok := raw.(*ecdsa.PublicKey)
		if !ok {
			return nil, errors.New("[ECDSAGoPublicKeyImportOpts] Invalid raw material. Expected *ecdsa.PublicKey.")
		}

		k = &ecdsaPublicKey{lowLevelKey}

		// If the key is not Ephemeral, store it.
		if !opts.Ephemeral() {
			// Store the key
			err = csp.ks.StoreKey(k)
			if err != nil {
				return nil, fmt.Errorf("Failed storing ECDSA key [%s]", err)
			}
		}

		return k, nil

	case *bccsp.RSAGoPublicKeyImportOpts:
		lowLevelKey, ok := raw.(*rsa.PublicKey)
		if !ok {
			return nil, errors.New("[RSAGoPublicKeyImportOpts] Invalid raw material. Expected *rsa.PublicKey.")
		}

		k = &rsaPublicKey{lowLevelKey}

		// If the key is not Ephemeral, store it.
		if !opts.Ephemeral() {
			// Store the key
			err = csp.ks.StoreKey(k)
			if err != nil {
				return nil, fmt.Errorf("Failed storing RSA publi key [%s]", err)
			}
		}

		return k, nil

	case *bccsp.X509PublicKeyImportOpts:
		x509Cert, ok := raw.(*x509.Certificate)
		if !ok {
			return nil, errors.New("[X509PublicKeyImportOpts] Invalid raw material. Expected *x509.Certificate.")
		}

		pk := x509Cert.PublicKey

		switch pk.(type) {
		case *ecdsa.PublicKey:
			return csp.KeyImport(pk, &bccsp.ECDSAGoPublicKeyImportOpts{Temporary: opts.Ephemeral()})
		case *rsa.PublicKey:
			return csp.KeyImport(pk, &bccsp.RSAGoPublicKeyImportOpts{Temporary: opts.Ephemeral()})
		default:
			return nil, errors.New("Certificate public key type not recognized. Supported keys: [ECDSA, RSA]")
		}

	default:
		return nil, errors.New("Import Key Options not recognized")
	}
}

// GetKey returns the key this CSP associates to
// the Subject Key Identifier ski.
func (csp *impl) GetKey(ski []byte) (k bccsp.Key, err error) {
	return csp.ks.GetKey(ski)
}

// Hash hashes messages msg using options opts.
func (csp *impl) Hash(msg []byte, opts bccsp.HashOpts) (digest []byte, err error) {
	var h hash.Hash
	if opts == nil {
		h = csp.conf.hashFunction()
	} else {
		switch opts.(type) {
		case *bccsp.SHAOpts:
			h = csp.conf.hashFunction()
		case *bccsp.SHA256Opts:
			h = sha256.New()
		case *bccsp.SHA384Opts:
			h = sha512.New384()
		case *bccsp.SHA3_256Opts:
			h = sha3.New256()
		case *bccsp.SHA3_384Opts:
			h = sha3.New384()
		default:
			return nil, fmt.Errorf("Algorithm not recognized [%s]", opts.Algorithm())
		}
	}

	h.Write(msg)
	return h.Sum(nil), nil
}

// GetHash returns and instance of hash.Hash using options opts.
// If opts is nil then the default hash function is returned.
func (csp *impl) GetHash(opts bccsp.HashOpts) (h hash.Hash, err error) {
	if opts == nil {
		return csp.conf.hashFunction(), nil
	}

	switch opts.(type) {
	case *bccsp.SHAOpts:
		return csp.conf.hashFunction(), nil
	case *bccsp.SHA256Opts:
		return sha256.New(), nil
	case *bccsp.SHA384Opts:
		return sha512.New384(), nil
	case *bccsp.SHA3_256Opts:
		return sha3.New256(), nil
	case *bccsp.SHA3_384Opts:
		return sha3.New384(), nil
	default:
		return nil, fmt.Errorf("Algorithm not recognized [%s]", opts.Algorithm())
	}
}

// Sign signs digest using key k.
// The opts argument should be appropriate for the primitive used.
//
// Note that when a signature of a hash of a larger message is needed,
// the caller is responsible for hashing the larger message and passing
// the hash (as digest).
func (csp *impl) Sign(k bccsp.Key, digest []byte, opts bccsp.SignerOpts) (signature []byte, err error) {
	// Validate arguments
	if k == nil {
		return nil, errors.New("Invalid Key. It must not be nil.")
	}
	if len(digest) == 0 {
		return nil, errors.New("Invalid digest. Cannot be empty.")
	}

	// Check key type
	switch k.(type) {
	case *ecdsaPrivateKey:
		return csp.signECDSA(k.(*ecdsaPrivateKey).privKey, digest, opts)
	case *rsaPrivateKey:
		if opts == nil {
			return nil, errors.New("Invalid options. Nil.")
		}

		return k.(*rsaPrivateKey).privKey.Sign(rand.Reader, digest, opts)
	default:
		return nil, fmt.Errorf("Key type not recognized [%s]", k)
	}
}

// Verify verifies signature against key k and digest
func (csp *impl) Verify(k bccsp.Key, signature, digest []byte, opts bccsp.SignerOpts) (valid bool, err error) {
	// Validate arguments
	if k == nil {
		return false, errors.New("Invalid Key. It must not be nil.")
	}
	if len(signature) == 0 {
		return false, errors.New("Invalid signature. Cannot be empty.")
	}
	if len(digest) == 0 {
		return false, errors.New("Invalid digest. Cannot be empty.")
	}

	// Check key type
	switch k.(type) {
	case *ecdsaPrivateKey:
		return csp.verifyECDSA(&(k.(*ecdsaPrivateKey).privKey.PublicKey), signature, digest, opts)
	case *ecdsaPublicKey:
		return csp.verifyECDSA(k.(*ecdsaPublicKey).pubKey, signature, digest, opts)
	case *rsaPrivateKey:
		if opts == nil {
			return false, errors.New("Invalid options. It must not be nil.")
		}
		switch opts.(type) {
		case *rsa.PSSOptions:
			err := rsa.VerifyPSS(&(k.(*rsaPrivateKey).privKey.PublicKey),
				(opts.(*rsa.PSSOptions)).Hash,
				digest, signature, opts.(*rsa.PSSOptions))

			return err == nil, err
		default:
			return false, fmt.Errorf("Opts type not recognized [%s]", opts)
		}
	case *rsaPublicKey:
		if opts == nil {
			return false, errors.New("Invalid options. It must not be nil.")
		}
		switch opts.(type) {
		case *rsa.PSSOptions:
			err := rsa.VerifyPSS(k.(*rsaPublicKey).pubKey,
				(opts.(*rsa.PSSOptions)).Hash,
				digest, signature, opts.(*rsa.PSSOptions))

			return err == nil, err
		default:
			return false, fmt.Errorf("Opts type not recognized [%s]", opts)
		}
	default:
		return false, fmt.Errorf("Key type not recognized [%s]", k)
	}
}

// Encrypt encrypts plaintext using key k.
// The opts argument should be appropriate for the primitive used.
func (csp *impl) Encrypt(k bccsp.Key, plaintext []byte, opts bccsp.EncrypterOpts) (ciphertext []byte, err error) {
	// Validate arguments
	if k == nil {
		return nil, errors.New("Invalid Key. It must not be nil.")
	}

	// Check key type
	switch k.(type) {
	case *aesPrivateKey:
		// check for mode
		switch opts.(type) {
		case *bccsp.AESCBCPKCS7ModeOpts, bccsp.AESCBCPKCS7ModeOpts:
			// AES in CBC mode with PKCS7 padding
			return AESCBCPKCS7Encrypt(k.(*aesPrivateKey).privKey, plaintext)
		default:
			return nil, fmt.Errorf("Mode not recognized [%s]", opts)
		}
	default:
		return nil, fmt.Errorf("Key type not recognized [%s]", k)
	}
}

// Decrypt decrypts ciphertext using key k.
// The opts argument should be appropriate for the primitive used.
func (csp *impl) Decrypt(k bccsp.Key, ciphertext []byte, opts bccsp.DecrypterOpts) (plaintext []byte, err error) {
	// Validate arguments
	if k == nil {
		return nil, errors.New("Invalid Key. It must not be nil.")
	}

	// Check key type
	switch k.(type) {
	case *aesPrivateKey:
		// check for mode
		switch opts.(type) {
		case *bccsp.AESCBCPKCS7ModeOpts, bccsp.AESCBCPKCS7ModeOpts:
			// AES in CBC mode with PKCS7 padding
			return AESCBCPKCS7Decrypt(k.(*aesPrivateKey).privKey, ciphertext)
		default:
			return nil, fmt.Errorf("Mode not recognized [%s]", opts)
		}
	default:
		return nil, fmt.Errorf("Key type not recognized [%s]", k)
	}
}
