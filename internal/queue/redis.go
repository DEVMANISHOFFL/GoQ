package queue

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
)

type RedisQueue struct {
	client *redis.Client
	queue  string
}

func NewRedisQueue(addr, queue string) *RedisQueue {
	rdb := redis.NewClient(&redis.Options{Addr: addr})
	return &RedisQueue{client: rdb, queue: queue}
}

func (q *RedisQueue) Enqueue(ctx context.Context, job string) error {
	return q.client.LPush(ctx, q.queue, job).Err()
}

func (q *RedisQueue) Dequeue(ctx context.Context) (string, error) {

	vals, err := q.client.BRPop(ctx, 0, q.queue).Result()
	if err != nil {
		return "", err
	}
	if len(vals) < 2 {
		return "", fmt.Errorf("unexpected BRPop response: %v", vals)
	}
	return vals[1], nil
}
