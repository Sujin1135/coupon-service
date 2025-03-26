package cache

import (
	"context"
	"coupon-service/internal/config"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"log"
	"time"
)

type Cache[T any] interface {
	SetAdd(ctx context.Context, key string, value string) (bool, error)
	SetDel(ctx context.Context, key string) (bool, error)
	Set(ctx context.Context, key string, value T) error
	Incr(ctx context.Context, key string) (int64, error)
	Decr(ctx context.Context, key string) (int64, error)
	ExpireAt(ctx context.Context, key string, expr time.Time) (bool, error)
}

type cache[T any] struct {
	redisClient *redis.Client
}

func (c cache[T]) SetAdd(ctx context.Context, key string, value string) (bool, error) {
	added, err := c.redisClient.SAdd(ctx, key, value).Result()
	if err != nil {
		log.Println(err)
		return false, errors.New("occurred an error when adding value to cache")
	}
	if added == 0 {
		return false, nil
	}
	return true, nil
}

func (c cache[T]) SetDel(ctx context.Context, key string) (bool, error) {
	result, err := c.redisClient.SRem(ctx, key).Result()
	if err != nil {
		log.Println(err)
		return false, errors.New("occurred an error when deleting value from cache")
	}
	if result == 0 {
		return false, nil
	}

	return true, nil
}

func (c cache[T]) Set(ctx context.Context, key string, value T) error {
	err := c.redisClient.Set(ctx, key, value, 0).Err()
	if err != nil {
		fmt.Println(err)
		return errors.New("occurred an error when setting value to cache")
	}
	return nil
}

func (c cache[T]) Incr(ctx context.Context, key string) (int64, error) {
	result, err := c.redisClient.Incr(ctx, key).Result()
	if err != nil {
		log.Println(err)
		return result, errors.New(fmt.Sprintf("occurred an error when try to increment by the key(%s)", key))
	}

	return result, nil
}

func (c cache[T]) Decr(ctx context.Context, key string) (int64, error) {
	result, err := c.redisClient.Decr(ctx, key).Result()
	if err != nil {
		log.Println(err)
		return result, errors.New(fmt.Sprintf("occurred an error when try to decrement by the key(%s)", key))
	}

	return result, nil
}

func (c cache[T]) ExpireAt(ctx context.Context, key string, expr time.Time) (bool, error) {
	result, err := c.redisClient.ExpireAt(ctx, key, expr).Result()
	if err != nil {
		log.Println(err)
		return result, errors.New(fmt.Sprintf("occurred an error when try to decrement by the key(%s)", key))
	}
	return result, nil
}

func NewCacheClient[T any]() Cache[T] {
	return &cache[T]{redisClient: config.CacheClient}
}
