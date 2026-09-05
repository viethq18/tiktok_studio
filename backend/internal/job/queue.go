package job

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	QueueGeneration = "tks:queue:generation"
	QueueExport     = "tks:queue:export"

	TypeCarouselGeneration = "carousel_generation"
	TypeCarouselExport     = "carousel_export"
)

// Queue is a Redis list used as a FIFO work queue. A list plus BRPOP is enough
// for MVP throughput and avoids inventing a queue protocol (§81).
type Queue struct{ rdb *redis.Client }

func NewQueue(rdb *redis.Client) *Queue { return &Queue{rdb: rdb} }

func (q *Queue) Enqueue(ctx context.Context, queue, jobID string) error {
	return q.rdb.LPush(ctx, queue, jobID).Err()
}

// Dequeue blocks until a job arrives or the timeout elapses.
func (q *Queue) Dequeue(ctx context.Context, queue string, timeout time.Duration) (string, error) {
	res, err := q.rdb.BRPop(ctx, timeout, queue).Result()
	if err == redis.Nil {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	if len(res) < 2 {
		return "", nil
	}
	return res[1], nil
}
