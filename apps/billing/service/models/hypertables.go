package models

import (
	"time"

	"github.com/antinvestor/common/timescale"
)

const (
	usageEventsChunk         = 7 * 24 * time.Hour
	usageEventsCompressAfter = 14 * 24 * time.Hour
	usageEventsRetain        = 730 * 24 * time.Hour // 2y billing audit window
)

// Hypertables returns the TimescaleDB configuration for this app's
// append-only tables. Applied idempotently by timescale.Ensure at
// service startup.
func Hypertables() []timescale.Hypertable {
	return []timescale.Hypertable{
		{
			Table:         "usage_events",
			TimeColumn:    "true_created_at",
			ChunkInterval: usageEventsChunk,
			SegmentBy:     []string{"partition_id", "subscription_id"},
			CompressAfter: usageEventsCompressAfter,
			RetainFor:     usageEventsRetain,
		},
	}
}
