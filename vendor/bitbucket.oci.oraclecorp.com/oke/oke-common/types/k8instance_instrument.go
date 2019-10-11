package types

import (
	"time"

	"github.com/go-kit/kit/metrics"
	"golang.org/x/net/context"
)

func NewK8InstanceInstrumentationMiddleware(s K8InstanceStoreServer, requestDuration metrics.Histogram) K8InstanceStoreServer {
	return &K8InstanceInstrumentationMiddleware{
		next:            s,
		requestDuration: requestDuration,
	}
}

type K8InstanceInstrumentationMiddleware struct {
	next            K8InstanceStoreServer
	requestDuration metrics.Histogram
}

// ensure that K8InstanceInstrumentationMiddleware implements K8InstanceStoreServer interface
var _ K8InstanceStoreServer = &K8InstanceInstrumentationMiddleware{}

func (m *K8InstanceInstrumentationMiddleware) K8InstanceList(ctx context.Context, req *K8InstanceListRequest) (*K8InstanceListResponse, error) {
	defer func(begin time.Time) {
		m.requestDuration.With("method", "K8InstanceStoreServer_K8InstanceList").Observe(
			time.Since(begin).Seconds())
	}(time.Now())
	return m.next.K8InstanceList(ctx, req)
}

func (m *K8InstanceInstrumentationMiddleware) K8InstanceNew(ctx context.Context, req *K8InstanceNewRequest) (*K8InstanceNewResponse, error) {
	defer func(begin time.Time) {
		m.requestDuration.With("method", "K8InstanceStoreServer_K8InstanceNew").Observe(
			time.Since(begin).Seconds())
	}(time.Now())
	return m.next.K8InstanceNew(ctx, req)
}

func (m *K8InstanceInstrumentationMiddleware) K8InstanceGet(ctx context.Context, req *K8InstanceRequest) (*K8InstanceResponse, error) {
	defer func(begin time.Time) {
		m.requestDuration.With("method", "K8InstanceStoreServer_K8InstanceGet").Observe(
			time.Since(begin).Seconds())
	}(time.Now())
	return m.next.K8InstanceGet(ctx, req)
}

func (m *K8InstanceInstrumentationMiddleware) K8InstanceDelete(ctx context.Context, req *K8InstanceDeleteRequest) (*K8InstanceDeleteResponse, error) {
	defer func(begin time.Time) {
		m.requestDuration.With("method", "K8InstanceStoreServer_K8InstanceDelete").Observe(
			time.Since(begin).Seconds())
	}(time.Now())
	return m.next.K8InstanceDelete(ctx, req)
}

func (m *K8InstanceInstrumentationMiddleware) K8InstanceUpdate(ctx context.Context, req *K8InstanceUpdateRequest) (*K8InstanceUpdateResponse, error) {
	defer func(begin time.Time) {
		m.requestDuration.With("method", "K8InstanceStoreServer_K8InstanceUpdate").Observe(
			time.Since(begin).Seconds())
	}(time.Now())
	return m.next.K8InstanceUpdate(ctx, req)
}

func (m *K8InstanceInstrumentationMiddleware) K8InstanceUpdateTKMState(ctx context.Context, req *K8InstanceUpdateTKMStateRequest) (*K8InstanceUpdateTKMStateResponse, error) {
	defer func(begin time.Time) {
		m.requestDuration.With("method", "K8InstanceStoreServer_K8InstanceUpdateTKMState").Observe(
			time.Since(begin).Seconds())
	}(time.Now())
	return m.next.K8InstanceUpdateTKMState(ctx, req)
}
