package stats

import (
	"context"
	"sync/atomic"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

func EmitStats(ctx context.Context, stats statsProvider) {
	statsPayload := map[string]uint64{
		"scanned": atomic.LoadUint64(stats.GetScanned()),
		"renamed": atomic.LoadUint64(stats.GetRenamed()),
		"skipped": atomic.LoadUint64(stats.GetSkipped()),
		"errors":  atomic.LoadUint64(stats.GetErrors()),
	}
	wailsRuntime.EventsEmit(ctx, "statsUpdated", statsPayload)
}
