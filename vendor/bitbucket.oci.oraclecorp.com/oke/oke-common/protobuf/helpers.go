package protobuf

import (
	"time"

	duration "github.com/golang/protobuf/ptypes/duration"
	timestamp "github.com/golang/protobuf/ptypes/timestamp"
)

func FromTime(src time.Time) *timestamp.Timestamp {
	return &timestamp.Timestamp{
		Seconds: src.Unix(),
		Nanos:   int32(src.UnixNano()),
	}
}

func ToTime(t *timestamp.Timestamp) time.Time {
	if t == nil {
		return time.Time{}
	}
	return time.Unix(t.Seconds, int64(t.Nanos))
}

func ToDuration(m *duration.Duration) time.Duration {
	if m == nil {
		return 0
	}
	return time.Duration(m.Seconds)*time.Second + time.Duration(m.Nanos)
}

func GetZeroTimestamp() *timestamp.Timestamp {
	return &timestamp.Timestamp{
		Seconds: 0,
		Nanos:   0,
	}
}

func IsTimestampZero(t *timestamp.Timestamp) bool {
	if t == nil {
		return true
	}

	if t.Seconds == 0 && t.Nanos == 0 {
		return true
	}

	return false
}
