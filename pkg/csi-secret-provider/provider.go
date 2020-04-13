package csisecretprovider

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"go.uber.org/zap"

	"golang.org/x/net/context"
	yaml "gopkg.in/yaml.v2"

	"github.com/oracle/oci-go-sdk/common/auth"
	ocisecrets "github.com/oracle/oci-go-sdk/secrets"

	"github.com/pkg/errors"
)

// Provider implements the secrets-store-csi-driver provider interface
type Provider struct {
	logger *zap.Logger
}

type SecretReference struct {
	SecretID      string `json:"secretID" yaml:"secretID"`
	VersionNumber int    `json:"versionNumber" yaml:"versionNumber"`
	FileName      string `json:"fileName" yaml:"fileName"`
}

// StringArray ...
type StringArray struct {
	Array []string `json:"array" yaml:"array"`
}

// NewProvider creates a new OCI Key Vault Provider.
func NewProvider(logger *zap.Logger) (*Provider, error) {
	logger.Sugar().Debugf("NewOCIProvider")
	return &Provider{
		logger: logger,
	}, nil
}

// MountSecretsStoreObjectContent mounts content of the secrets store object to target path
func (p *Provider) MountSecretsStoreObjectContent(ctx context.Context, attrib map[string]string, secrets map[string]string, targetPath string, permission os.FileMode) (err error) {

	objectsStrings := attrib["objects"]
	if objectsStrings == "" {
		return fmt.Errorf("objects is not set")
	}
	p.logger.Sugar().Infof("objects: %s", objectsStrings)

	var objects StringArray
	err = yaml.Unmarshal([]byte(objectsStrings), &objects)
	if err != nil {
		p.logger.Sugar().Infof("unmarshal failed for objects")
		return err
	}
	p.logger.Sugar().Debugf("objects array: %v", objects.Array)
	secretReferences := []SecretReference{}
	for i, object := range objects.Array {
		var secretRef SecretReference
		err = yaml.Unmarshal([]byte(object), &secretRef)
		if err != nil {
			p.logger.Sugar().Infof("unmarshal failed for secretReferences at index %d", i)
			return err
		}
		secretReferences = append(secretReferences, secretRef)
	}

	p.logger.Sugar().Infof("unmarshaled secretReferences: %v", secretReferences)
	p.logger.Sugar().Infof("secretReferences len: %d", len(secretReferences))

	if len(secretReferences) == 0 {
		return fmt.Errorf("objects array is empty")
	}

	// TODO: Support more options on the parameters?
	cfg, err := auth.InstancePrincipalConfigurationProvider()
	if err != nil {
		return errors.Wrap(err, "unable to create instance principal configuration provider")
	}

	secretClient, err := ocisecrets.NewSecretsClientWithConfigurationProvider(cfg)
	if err != nil {
		return errors.Wrap(err, "unable to create secret client")
	}

	// Create Secret Fetcher thing

	for _, secretRef := range secretReferences {

		resp, err := secretClient.GetSecretBundle(ctx, ocisecrets.GetSecretBundleRequest{
			SecretId: &secretRef.SecretID,
		})
		if err != nil {
			return errors.Wrapf(err, "unable to fetch secret: %q", secretRef.SecretID)
		}

		objectContent, err := base64.StdEncoding.DecodeString(*resp.SecretBundleContent.(ocisecrets.Base64SecretBundleContentDetails).Content)
		if err != nil {
			return errors.Wrapf(err, "unable to base64 decode secret: %q", secretRef.SecretID)
		}

		fileName := secretRef.FileName
		if fileName == "" {
			fileName = *resp.SecretBundle.VersionName
		}
		if err := ioutil.WriteFile(filepath.Join(targetPath, fileName), []byte(objectContent), permission); err != nil {
			return errors.Wrapf(err, "secrets store csi driver failed to mount %s at %s", fileName, targetPath)
		}
		p.logger.Sugar().Infof("secrets store csi driver mounted %s", fileName)
		p.logger.Sugar().Infof("Mount point: %s", targetPath)
	}

	return nil
}

// apiVersion: secrets-store.csi.x-k8s.io/v1alpha1
// kind: SecretProviderClass
// metadata:
//   name: oci-test
// spec:
//   provider: oci
//   parameters:
//     objects:  |
//       array:
//         - |
//           secretID: ocid1.vaultsecret.oc1.iad.amaaaaaa2ahbgkyawyssocqw7ohcpt6ejrj4z55zrg2lcw3ezj4hl2cuhktq
// 		  fileName: test

// volumes:
// - name: secrets-store-inline
// csi:
// 	driver: secrets-store.csi.k8s.io
// 	readOnly: true
// 	volumeAttributes:
// 	secretProviderClass: "oci-test"
