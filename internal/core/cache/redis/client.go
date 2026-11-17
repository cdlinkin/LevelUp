package core_redis_cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/cdlinkin/LevelUp/internal/core/domain"
	"github.com/redis/go-redis/v9"
)

type Client struct {
	client  *redis.Client
	timeout time.Duration
	ttl     time.Duration
}

func NewClient(ctx context.Context, config Config) (*Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr: config.Addr,
	})

	pingCtx, cancel := context.WithTimeout(ctx, config.Timeout)
	defer cancel()

	if err := client.Ping(pingCtx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("redis ping: %w", err)
	}

	return &Client{
		client:  client,
		timeout: config.Timeout,
		ttl:     config.TTL,
	}, nil
}

func (c *Client) Close() {
	if err := c.client.Close(); err != nil {
		fmt.Println("failed to close redis client:", err)
	}
}

func (c *Client) GetStatistics(
	ctx context.Context,
	userID int,
	from *time.Time,
	to *time.Time,
) (domain.Statistics, bool, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	version, err := c.version(ctx, userID)
	if err != nil {
		return domain.Statistics{}, false, err
	}

	payload, err := c.client.Get(
		ctx,
		statisticsKey(userID, version, from, to),
	).Bytes()
	if errors.Is(err, redis.Nil) {
		return domain.Statistics{}, false, nil
	}
	if err != nil {
		return domain.Statistics{}, false, fmt.Errorf("get statistics: %w", err)
	}

	statistics, err := statisticsDomainFromPayload(payload)
	if err != nil {
		return domain.Statistics{}, false, fmt.Errorf("decode statistics: %w", err)
	}

	return statistics, true, nil
}

func (c *Client) SetStatistics(
	ctx context.Context,
	userID int,
	from *time.Time,
	to *time.Time,
	statistics domain.Statistics,
) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	version, err := c.version(ctx, userID)
	if err != nil {
		return err
	}

	payload, err := payloadFromStatistics(statistics)
	if err != nil {
		return fmt.Errorf("encode statistics: %w", err)
	}

	if err := c.client.Set(
		ctx,
		statisticsKey(userID, version, from, to),
		payload,
		c.ttl,
	).Err(); err != nil {
		return fmt.Errorf("set statistics: %w", err)
	}

	return nil
}

func (c *Client) InvalidateStatistics(
	ctx context.Context,
	userID int,
) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	if err := c.client.Incr(ctx, versionKey(userID)).Err(); err != nil {
		return fmt.Errorf("invalidate statistics: %w", err)
	}

	return nil
}

func (c *Client) version(ctx context.Context, userID int) (int64, error) {
	value, err := c.client.Get(ctx, versionKey(userID)).Int64()
	if errors.Is(err, redis.Nil) {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("get statistics version: %w", err)
	}

	return value, nil
}
