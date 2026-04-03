package common

import (
	"context"
	"kgogame/util/logs"
	"time"
)

func CallMetric(ctx context.Context, info string, fn func()) {
	start := time.Now()
	defer func() {
		logs.Debug("%s duration %s", info, time.Since(start))
	}()

	fn()
}
