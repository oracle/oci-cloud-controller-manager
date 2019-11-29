package client

import (
	"github.com/oracle/oci-cloud-controller-manager/pkg/cloudprovider/providers/oci/config"
	"go.uber.org/zap"
	"k8s.io/client-go/util/flowcontrol"
)

const (
	rateLimitQPSDefault    = 20.0
	rateLimitBucketDefault = 5
)

//GetClient returns the client for given Configuration
func GetClient(logger *zap.SugaredLogger, cfg *config.Config) (Interface, error) {
	cp, err := config.NewConfigurationProvider(cfg)
	if err != nil {
		logger.With(zap.Error(err)).Fatal("Unable to create client.")
		return nil, err
	}

	c, err := New(logger, cp, &RateLimiter{
		Reader: flowcontrol.NewTokenBucketRateLimiter(
			rateLimitQPSDefault,
			rateLimitBucketDefault,
		),
		Writer: flowcontrol.NewTokenBucketRateLimiter(
			rateLimitQPSDefault,
			rateLimitBucketDefault,
		),
	}, cfg.Auth.TenancyID)
	return c, err
}
