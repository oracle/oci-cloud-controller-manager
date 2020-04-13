package main

import (
	"context"
	"encoding/json"
	"os"

	// "github.com/docker/distribution/version"
	csisecretprovider "github.com/oracle/oci-cloud-controller-manager/pkg/csi-secret-provider"
	"go.uber.org/zap"

	// "github.com/Azure/secrets-store-csi-driver-provider-azure/pkg/version"
	"github.com/oracle/oci-cloud-controller-manager/pkg/logging"

	"github.com/spf13/pflag"
)

var (
	attributes  = pflag.String("attributes", "", "volume attributes")
	secrets     = pflag.String("secrets", "", "node publish ref secret")
	targetPath  = pflag.String("targetPath", "", "Target path to write data.")
	permission  = pflag.String("permission", "", "File permission")
	debug       = pflag.Bool("debug", false, "sets log to debug level")
	versionInfo = pflag.Bool("version", false, "prints the version information")
)

func main() {
	pflag.Parse()

	var attrib, secret map[string]string
	var filePermission os.FileMode
	var err error

	log := logging.Logger()
	defer log.Sync()
	zap.ReplaceGlobals(log)
	// TODO: Handle debug flag to set debug level log.

	// if *versionInfo {
	// 	if err = version.PrintVersion(); err != nil {
	// 		log.Fatalf("failed to print version, err: %+v", err)
	// 	}
	// 	os.Exit(0)
	// }

	err = json.Unmarshal([]byte(*attributes), &attrib)
	if err != nil {
		log.Sugar().Fatal("failed to unmarshal attributes, err: %v", err)
	}
	err = json.Unmarshal([]byte(*secrets), &secret)
	if err != nil {
		log.Sugar().Fatalf("failed to unmarshal secrets, err: %v", err)
	}
	err = json.Unmarshal([]byte(*permission), &filePermission)
	if err != nil {
		log.Sugar().Fatalf("failed to unmarshal file permission, err: %v", err)
	}

	provider, err := csisecretprovider.NewProvider(log)
	if err != nil {
		log.Sugar().Fatalf("[error] : %v", err)
	}

	ctx := context.Background()
	err = provider.MountSecretsStoreObjectContent(ctx, attrib, secret, *targetPath, filePermission)
	if err != nil {
		log.Sugar().Fatalf("[error] : %v", err)
	}

	os.Exit(0)
}
