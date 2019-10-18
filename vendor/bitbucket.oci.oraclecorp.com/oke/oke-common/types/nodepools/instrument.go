package nodepools

import (
	"time"

	"github.com/go-kit/kit/metrics"
	"golang.org/x/net/context"
)

func NewInstrumentationMiddleware(s StoreServer, requestDuration metrics.Histogram) StoreServer {
	return &InstrumentationMiddleware{
		next:            s,
		requestDuration: requestDuration,
	}
}

type InstrumentationMiddleware struct {
	next            StoreServer
	requestDuration metrics.Histogram
}

// ensure that InstrumentationMiddleware implements NodePoolStoreServer interface
var _ StoreServer = &InstrumentationMiddleware{}

func (m *InstrumentationMiddleware) NodePoolNew(ctx context.Context, req *NewRequest) (*NewResponse, error) {
	defer func(begin time.Time) {
		m.requestDuration.With("method", "NodePoolStoreServer_NodePoolNew").Observe(
			time.Since(begin).Seconds())
	}(time.Now())
	return m.next.NodePoolNew(ctx, req)
}

func (m *InstrumentationMiddleware) NodePoolGet(ctx context.Context, req *GetRequest) (*GetResponse, error) {
	defer func(begin time.Time) {
		m.requestDuration.With("method", "NodePoolStoreServer_NodePoolGet").Observe(
			time.Since(begin).Seconds())
	}(time.Now())
	return m.next.NodePoolGet(ctx, req)
}

func (m *InstrumentationMiddleware) NodePoolDelete(ctx context.Context, req *DeleteRequest) (*DeleteResponse, error) {
	defer func(begin time.Time) {
		m.requestDuration.With("method", "NodePoolStoreServer_NodePoolDelete").Observe(
			time.Since(begin).Seconds())
	}(time.Now())
	return m.next.NodePoolDelete(ctx, req)
}

func (m *InstrumentationMiddleware) NodePoolUpdate(ctx context.Context, req *UpdateRequest) (*UpdateResponse, error) {
	defer func(begin time.Time) {
		m.requestDuration.With("method", "NodePoolStoreServer_NodePoolUpdate").Observe(
			time.Since(begin).Seconds())
	}(time.Now())
	return m.next.NodePoolUpdate(ctx, req)
}

func (m *InstrumentationMiddleware) NodePoolUpdateNodeStates(ctx context.Context, req *UpdateNodeStatesRequest) (*UpdateNodeStatesResponse, error) {
	defer func(begin time.Time) {
		m.requestDuration.With("method", "NodePoolStoreServer_NodePoolUpdateNodeStates").Observe(
			time.Since(begin).Seconds())
	}(time.Now())
	return m.next.NodePoolUpdateNodeStates(ctx, req)
}
