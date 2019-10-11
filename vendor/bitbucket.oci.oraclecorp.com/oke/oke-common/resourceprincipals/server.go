package resourceprincipals

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"math"
	"net"
	"sync"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"

	"bitbucket.oci.oraclecorp.com/oke/oke-common/types"
)

var (
	resourcePrincipalExpiryTimestampSeconds = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "resource_principal_expires_at_timestamp_seconds",
			Help: "Unix timestamp for Resource Principal Session Token expiry",
		},
		[]string{"tkc_id"},
	)
)

func init() {
	prometheus.MustRegister(resourcePrincipalExpiryTimestampSeconds)
}

var _ types.ResourcePrincipalRecipientServer = (*Server)(nil)

// Server implements ResourcePrincipalsRecipient gRPC server.
type Server struct {
	// address to serve on.
	address string

	// mux guards against concurrent updates to the Resource Principal.
	mux sync.Mutex

	// securityToken is the Resource Principals Session Token (RPST).
	securityToken securityToken

	// path is the filesystem path were authn material is written.
	path string

	// clusterShortID is the short id of cluster for which we're receiving Resource Principals.
	clusterShortID string

	logger *log.Entry

	// tlsConfig for mTLS.
	tlsConfig *tls.Config
	*grpc.Server
}

// NewServer constructs a new Server.
func NewServer(logger *log.Entry, tlsConfig *tls.Config, address, path, clusterShortID string) *Server {
	return &Server{
		logger:               logger,
		address:              address,
		path:                 path,
		clusterShortID:       clusterShortID,
		tlsConfig:            tlsConfig,
	}
}

// SetResourcePrincipal accepts requests to update our Resource Principal auth
// material (RPST, PrivateKey) and writes updated tokens to the filesystem.
func (s *Server) SetResourcePrincipal(ctx context.Context, req *types.SetResourcePrincipalRequest) (*types.SetResourcePrincipalResponse, error) {
	s.mux.Lock()
	defer s.mux.Unlock()

	token, err := newToken(req.ResourcePrincipal.RPST)
	if err != nil {
		err := errors.Wrap(err, "failed to parse RPST JWT")
		s.logger.WithError(err).Error("Received invalid SetResourcePrincipal request")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	logger := s.logger.WithField("expires_at", token.ExpiresAt())

	logger.Info("Received SetResourcePrincipal request (RPST, PrivateKey)")

	// Only accept fresher Resource Principals.
	if s.securityToken != nil && token.ExpiresAt().Before(s.securityToken.ExpiresAt()) {
		err := errors.Errorf("given RPST less fresh than existing RPST: %v vs %v", token.ExpiresAt(), s.securityToken.ExpiresAt())
		logger.WithError(err).Error("Received SetResourcePrincipal request with an less fresh RPST")
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	err = s.writeToFilesystem(req.ResourcePrincipal)
	if err != nil {
		err := errors.Wrap(err, "writing resource principal to filesystem")
		logger.WithError(err).Error("Failed to persist Resource Principal")
		return nil, status.Error(codes.Unavailable, err.Error())
	}

	s.securityToken = token

	resourcePrincipalExpiryTimestampSeconds.With(prometheus.Labels{"tkc_id": s.clusterShortID}).
		Set(float64(token.ExpiresAt().UnixNano()) / 1e9)

	return &types.SetResourcePrincipalResponse{}, nil
}

func (s *Server) writeToFilesystem(rp *types.ResourcePrincipal) error {
	data, err := json.Marshal(rp)
	if err != nil {
		return errors.Wrap(err, "marshaling resource principal")
	}

	err = ioutil.WriteFile(s.path, data, 0600)
	if err != nil {
		return errors.Wrapf(err, "writing resource principal to %q", s.path)
	}

	return nil
}

// Run runs the server until the given context is canceled.
func (s *Server) Run(ctx context.Context) error {
	listener, err := net.Listen("tcp", s.address)
	if err != nil {
		log.WithError(err).Fatalf("Error running server")
	}

	opts := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(math.MaxInt32),
		grpc.MaxSendMsgSize(math.MaxInt32),
		grpc_middleware.WithUnaryServerChain(
			grpc_prometheus.UnaryServerInterceptor,
			grpc_recovery.UnaryServerInterceptor(),
		),
		grpc_middleware.WithStreamServerChain(
			grpc_prometheus.StreamServerInterceptor,
			grpc_recovery.StreamServerInterceptor(),
		),
	}
	if s.tlsConfig != nil {
		opts = append(opts, grpc.Creds(credentials.NewTLS(s.tlsConfig)))
	}

	s.Server = grpc.NewServer(opts...)
	types.RegisterResourcePrincipalRecipientServer(s.Server, s)

	errCh := make(chan error, 1)
	go func() {
		err := s.Serve(listener)
		errCh <- errors.Wrap(err, "failed to run server")
	}()

	select {
	case err := <-errCh:
		if err == nil {
			s.logger.Infof("Server shutdown")
		}
		return err
	case <-ctx.Done():
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

		go func() {
			s.GracefulStop()
			cancel()
		}()

		<-ctx.Done()

		s.Stop()
		return nil
	}
}
