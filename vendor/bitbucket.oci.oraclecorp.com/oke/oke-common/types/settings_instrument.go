package types

import (
	"time"

	"github.com/go-kit/kit/metrics"
	google_protobuf "github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/net/context"
)

func NewSettingsInstrumentationMiddleware(s SettingsStoreServer, requestDuration metrics.Histogram) SettingsStoreServer {
	return &SettingsInstrumentationMiddleware{
		next:            s,
		requestDuration: requestDuration,
	}
}

type SettingsInstrumentationMiddleware struct {
	next            SettingsStoreServer
	requestDuration metrics.Histogram
}

// ensure that SettingsInstrumentationMiddleware implements SettingsStoreServer interface
var _ SettingsStoreServer = &SettingsInstrumentationMiddleware{}

func (m *SettingsInstrumentationMiddleware) SettingsList(ctx context.Context, req *SettingsListRequest) (*SettingsListResponse, error) {
	defer func(begin time.Time) {
		m.requestDuration.With("method", "SettingsStoreServer_SettingsList").Observe(
			time.Since(begin).Seconds())
	}(time.Now())
	return m.next.SettingsList(ctx, req)
}

func (m *SettingsInstrumentationMiddleware) SettingsGet(ctx context.Context, req *SettingsGetRequest) (*SettingsGetResponse, error) {
	defer func(begin time.Time) {
		m.requestDuration.With("method", "SettingsStoreServer_SettingsGet").Observe(
			time.Since(begin).Seconds())
	}(time.Now())
	return m.next.SettingsGet(ctx, req)
}

func (m *SettingsInstrumentationMiddleware) SettingsNew(ctx context.Context, req *SettingsNewRequest) (*SettingsNewResponse, error) {
	defer func(begin time.Time) {
		m.requestDuration.With("method", "SettingsStoreServer_SettingsNew").Observe(
			time.Since(begin).Seconds())
	}(time.Now())
	return m.next.SettingsNew(ctx, req)
}

func (m *SettingsInstrumentationMiddleware) SettingsUpdate(ctx context.Context, req *SettingsUpdateRequest) (*SettingsUpdateResponse, error) {
	defer func(begin time.Time) {
		m.requestDuration.With("method", "SettingsStoreServer_SettingsUpdate").Observe(
			time.Since(begin).Seconds())
	}(time.Now())
	return m.next.SettingsUpdate(ctx, req)
}

func (m *SettingsInstrumentationMiddleware) SettingsDelete(ctx context.Context, req *SettingsDeleteRequest) (*google_protobuf.Empty, error) {
	defer func(begin time.Time) {
		m.requestDuration.With("method", "SettingsStoreServer_SettingsDelete").Observe(
			time.Since(begin).Seconds())
	}(time.Now())
	return m.next.SettingsDelete(ctx, req)
}
