// Copyright (c) 2018, Oracle and/or its affiliates. All rights reserved.

package ociauthz

import (
	"bytes"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"sync"
	"time"

	"oracle.com/oci/httpsigner"
)

const pkixPublicKeyType = `PUBLIC KEY`

var pkixPreamble = []byte("-----BEGIN PUBLIC KEY-----")

var (
	keyCache  = KeyCache{}
	cacheLock = sync.Mutex{}
)

const (
	cachePrefixSystemKey = "SYSTEMKEY"
	cachePrefixAPIKey    = "APIKEY"
	nullKeyFilter        = ""

	// Pattern comes from VALID_ASW_KEY_ID_FORMAT
	// authorization-sdk/authentication-token-verification/src/main/java/com/oracle/pic/identity/authentication/
	//   token/TokenVerifierImpl.java
	systemKeyPattern = `^(?i)asw[a-zA-Z0-9-_]{0,50}$`
)

// SEC-2456 NOTE:  It would be better to have the caching and filtering be composite wrappers, but in order to avoid
// breaking the API compatibility for this fix, the functionality has been added directly to KeyServiceKeySupplier. We
// will revisit for a major rev bump.

// KeyServiceKeySupplier retrieves keys from the Identity key service
// Deprecated
type KeyServiceKeySupplier struct {
	client      httpsigner.Client
	uri         string
	keyCache    *KeyCache
	cachePeriod time.Duration
	cachePrefix string
	keyFilter   *regexp.Regexp
}

// NewKeyServiceKeySupplier returns a new instance of the key service key supplier for looking up token signing keys
// (e.g. `asw`)
func NewKeyServiceKeySupplier(client httpsigner.Client, endpoint string) *KeyServiceKeySupplier {
	return newKeyServiceKeySupplier(
		client,
		fmt.Sprintf(keyServiceURITemplate, endpoint),
		time.Minute*keyServiceTokenServiceCachePeriodMinutes,
		cachePrefixSystemKey,
		systemKeyPattern,
	)
}

// NewAPIKeyServiceKeySupplier returns a new instance of the key service key supplier for looking up customer API
// public keys
func NewAPIKeyServiceKeySupplier(client httpsigner.Client, endpoint string) *KeyServiceKeySupplier {
	return newKeyServiceKeySupplier(
		client,
		fmt.Sprintf(keyServiceSRURITemplate, endpoint),
		time.Minute*keyServiceDefaultCachePeriodMinutes,
		cachePrefixAPIKey,
		nullKeyFilter,
	)
}

// newKeyServiceKeySupplier returns a new instance of the key service key supplier.  Will panic if keyPattern is
// non-empty and does not compile as a valid regexp.
func newKeyServiceKeySupplier(client httpsigner.Client, uri string, cachePeriod time.Duration, prefix, keyPattern string) *KeyServiceKeySupplier {
	var keyFilter *regexp.Regexp
	if len(keyPattern) > 0 {
		keyFilter = regexp.MustCompile(keyPattern)
	}

	return &KeyServiceKeySupplier{
		client:      client,
		uri:         uri,
		keyCache:    &keyCache,
		cachePeriod: cachePeriod,
		cachePrefix: prefix,
		keyFilter:   keyFilter,
	}
}

// Key returns a key with the given id.  If the keyID doesn't match supported keyID format, ErrUnsupportedKeyFormat is
// return.  If no key identified by keyID can be found, then ErrKeyNotFound is returned. ServiceResponseError is
// returned if there is an error fetching the key.
func (k *KeyServiceKeySupplier) Key(keyID string) (interface{}, error) {

	// validate acceptable keyID
	if k.keyFilter != nil && !k.keyFilter.MatchString(keyID) {
		return nil, ErrUnsupportedKeyFormat
	}

	// assemble cache key id to avoid collisions in the cache
	cacheKeyID := fmt.Sprintf("%s%s", k.cachePrefix, keyID)

	// Check for cached key
	publicKey, ok := k.keyCache.Get(cacheKeyID)
	if ok {
		return publicKey, nil
	}

	// Build request
	request, err := http.NewRequest("GET", fmt.Sprintf("%s/%s", k.uri, keyID), nil)
	if err != nil {
		return nil, err
	}

	// Call
	response, err := k.client.Do(request)
	if err != nil {
		return nil, err
	}

	// Check response status
	if response.StatusCode != http.StatusOK {
		if response.StatusCode == http.StatusNotFound {
			return nil, ErrKeyNotFound
		}

		return nil, &ServiceResponseError{response}
	}

	// Read response
	body, err := ioutil.ReadAll(response.Body)
	defer response.Body.Close()
	if err != nil {
		return nil, err
	}

	// parse response
	if bytes.Contains(body, pkixPreamble) {
		publicKey, err = decodePKIXPublicKey(body)
	} else {
		publicKey, err = decodeJWK(body)
	}

	if err != nil {
		return nil, err
	}

	// Store in cache
	k.keyCache.Store(cacheKeyID, publicKey, k.cachePeriod)

	return publicKey, nil
}

// decodePKIXPublicKey extracts public key material from a PKIX formatted response
func decodePKIXPublicKey(encodedKey []byte) (interface{}, error) {

	// first decode the PEM block into DER format
	block, _ := pem.Decode(encodedKey)
	if block == nil || block.Type != pkixPublicKeyType {
		return nil, ErrPEMDecodeError
	}

	// decode DER encoded key
	return x509.ParsePKIXPublicKey(block.Bytes)
}

// decodeJKW extracts key material from a JWK formatted response
func decodeJWK(encodedKey []byte) (interface{}, error) {

	// Parse response
	var jwk JWK
	err := json.Unmarshal(encodedKey, &jwk)
	if err != nil {
		return nil, err
	}

	// Extract the public key
	return jwk.PublicKey()

}

// KeyCache caches keys for a short time to avoid repeated calls to the Identity service
type KeyCache map[string]*CachedKey

// CachedKey represents an item in the key cache
type CachedKey struct {
	Key     interface{}
	Expires time.Time
}

// Get returns a cached key (if present), and a boolean to indicate whether or not the key was
// available
func (kc KeyCache) Get(keyID string) (interface{}, bool) {
	cacheLock.Lock()
	defer cacheLock.Unlock()

	cachedKey, ok := kc[keyID]
	if ok {
		if time.Now().After(cachedKey.Expires) {
			// key has expired
			delete(kc, keyID)
		} else {
			return cachedKey.Key, true
		}
	}

	return nil, false
}

// Store stores the indicated key in the cache with the specified cache period
func (kc KeyCache) Store(keyID string, key interface{}, cachePeriod time.Duration) {

	// TODO: run this only x% of times?
	kc.ClearExpired()

	cacheLock.Lock()
	defer cacheLock.Unlock()

	expires := time.Now().Add(cachePeriod)

	kc[keyID] = &CachedKey{
		Key:     key,
		Expires: expires,
	}
}

// ClearExpired clears old keys from the cache by checking their expiry
func (kc KeyCache) ClearExpired() {
	cacheLock.Lock()
	defer cacheLock.Unlock()

	now := time.Now()
	for keyID, cachedKey := range kc {
		if now.After(cachedKey.Expires) {
			delete(kc, keyID)
		}
	}
}
