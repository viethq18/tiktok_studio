// Package ratelimit implements the per-user daily quotas from §163.
package ratelimit

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/tks/backend/internal/httpx"
)

type Action string

const (
	ActionGeneration  Action = "generation"
	ActionImageSearch Action = "image_search"
	ActionExport      Action = "export"
	ActionUpload      Action = "upload"
)

// Daily budgets (§163). Adjust once real usage data exists.
var dailyLimits = map[Action]int{
	ActionGeneration:  20,
	ActionImageSearch: 30,
	ActionExport:      20,
	ActionUpload:      100,
}

var friendlyName = map[Action]string{
	ActionGeneration:  "tạo carousel",
	ActionImageSearch: "tìm ảnh",
	ActionExport:      "export",
	ActionUpload:      "upload",
}

type Limiter struct{ rdb *redis.Client }

func New(rdb *redis.Client) *Limiter { return &Limiter{rdb: rdb} }

// Take consumes one unit of the user's daily budget for action.
// Exceeding the budget returns a clear error rather than failing silently.
func (l *Limiter) Take(ctx context.Context, userID string, action Action) error {
	limit, ok := dailyLimits[action]
	if !ok || l == nil || l.rdb == nil {
		return nil
	}
	key := fmt.Sprintf("ratelimit:%s:%s:%s", action, userID, time.Now().UTC().Format("2006-01-02"))
	count, err := l.rdb.Incr(ctx, key).Result()
	if err != nil {
		return nil // never block the product on a limiter outage
	}
	if count == 1 {
		l.rdb.Expire(ctx, key, 25*time.Hour)
	}
	if int(count) > limit {
		return httpx.NewError(http.StatusTooManyRequests, "rate_limited",
			fmt.Sprintf("Bạn đã đạt giới hạn %s hôm nay (%d/ngày). Hãy thử lại vào ngày mai.", friendlyName[action], limit))
	}
	return nil
}
