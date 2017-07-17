/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package common

import (
	"encoding/base64"
	"fmt"
	"strings"
)

// Base64URLEncode encodes the byte array into a base64 string
func Base64URLEncode(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}

// Base64URLDecode decodes the base64 string into a byte array
func Base64URLDecode(data string) ([]byte, error) {
	//check if it has padding or not
	if strings.HasSuffix(data, "=") {
		return base64.URLEncoding.DecodeString(data)
	}
	return base64.RawURLEncoding.DecodeString(data)
}

// AsURL return a URL in the form host:port
func AsURL(host string, port int) string {
	return fmt.Sprintf("%s:%d", host, port)
}
