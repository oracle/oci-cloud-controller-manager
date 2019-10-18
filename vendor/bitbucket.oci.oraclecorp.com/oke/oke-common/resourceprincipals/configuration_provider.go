package resourceprincipals

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"io/ioutil"
	"sync"
	"time"

	"github.com/oracle/oci-go-sdk/common"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/wait"

	"bitbucket.oci.oraclecorp.com/oke/oke-common/types"
)

var _ common.ConfigurationProvider = (*resourcePrincipalConfigurationProvider)(nil)

type resourcePrincipalConfigurationProvider struct {
	mux  sync.Mutex // protects rpst and ekPr
	rpst securityToken
	ekPr *rsa.PrivateKey

	region common.Region

	path string // path to json serialised types.ResourcePrincipal

	logger *log.Entry
}

// NewConfigurationProvider constructs a ConfigurationProvider for
// authenticating with OCI using a Resource Principal v2.2.
// Blocks until there is a token present on the filesystem.
func NewConfigurationProvider(ctx context.Context, logger *log.Entry, region common.Region, path string) (common.ConfigurationProvider, error) {
	cp := &resourcePrincipalConfigurationProvider{
		region: region,
		path:   path,
		logger: logger,
	}

	// Block until we have a resource principal token or the context is canceled.
	err := wait.PollImmediateUntil(2*time.Second, func() (bool, error) {
		err := cp.refreshFromFilesystem()
		if err != nil {
			log.WithError(err).Info("Waiting for initial Resource Principal")
		}
		return err == nil, nil

	}, ctx.Done())
	if err != nil {
		return nil, errors.Wrap(err, "awaiting initial authentication material")
	}

	return cp, nil
}

func (cp *resourcePrincipalConfigurationProvider) PrivateRSAKey() (*rsa.PrivateKey, error) {
	cp.mux.Lock()
	defer cp.mux.Unlock()

	err := cp.refreshFromFilesystemIfNotValid()
	if err != nil {
		return nil, err
	}

	c := *cp.ekPr
	return &c, nil
}

func (cp *resourcePrincipalConfigurationProvider) KeyID() (string, error) {
	cp.mux.Lock()
	defer cp.mux.Unlock()

	err := cp.refreshFromFilesystemIfNotValid()
	if err != nil {
		return "", err
	}

	return "ST$" + cp.rpst.String(), nil
}

func (cp *resourcePrincipalConfigurationProvider) refreshFromFilesystemIfNotValid() error {
	// Attempt refreshing the token if we don't have one, it has expired, or
	// will expire within the next 5m.
	if !cp.hasAuthnMaterial() || !cp.rpst.Valid() {
		err := cp.refreshFromFilesystem()
		if err != nil {
			if !cp.hasAuthnMaterial() {
				return err
			}
			// If we have authn material (which is potentially invalid) log the error but
			// use the existing authn material as it's the best we can do (we'll retry 401s).
			cp.logger.WithError(err).Error("Failed to refresh resource principal from file. Using existing token.")
		}
	}

	return nil
}

// hasAuthnMaterial returns true if we have both an RPST and a ephemeral
// private key.
func (cp *resourcePrincipalConfigurationProvider) hasAuthnMaterial() bool {
	return cp.rpst != nil && cp.ekPr != nil
}

// refreshFromFilesystem reads the resource principal from the given filesystem
// path and updates the cached token and key if the contents are fresher than
func (cp *resourcePrincipalConfigurationProvider) refreshFromFilesystem() error {
	b, err := ioutil.ReadFile(cp.path)
	if err != nil {
		return errors.Wrapf(err, "reading %q", cp.path)
	}

	var rp types.ResourcePrincipal
	err = json.Unmarshal(b, &rp)
	if err != nil {
		return errors.Wrap(err, "unmarshaling resource principal")
	}

	rpst, err := newToken(rp.RPST)
	if err != nil {
		return errors.Wrap(err, "parsing RPST")
	}

	// Verify that token read from file is fresher than the existing one if present.
	if cp.rpst != nil && !rpst.ExpiresAt().After(cp.rpst.ExpiresAt()) {
		return errors.Errorf("no newer RPST at %q", cp.path)
	}

	ekPr, err := common.PrivateKeyFromBytes([]byte(rp.PrivateKey), nil)
	if err != nil {
		return errors.Wrap(err, "parsing ephemeral private key PEM")
	}

	cp.rpst = rpst
	cp.ekPr = ekPr

	return nil
}

func (cp *resourcePrincipalConfigurationProvider) TenancyOCID() (string, error) {
	return "", nil
}

func (cp *resourcePrincipalConfigurationProvider) UserOCID() (string, error) {
	return "", nil
}

func (cp *resourcePrincipalConfigurationProvider) KeyFingerprint() (string, error) {
	return "", nil
}

func (cp *resourcePrincipalConfigurationProvider) Region() (string, error) {
	return string(cp.region), nil
}
