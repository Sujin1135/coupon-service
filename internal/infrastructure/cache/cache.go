package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"log"
	"time"
)

type Cache interface {
	SetAdd(ctx context.Context, key string, value string) (bool, error)
	SetDel(ctx context.Context, key string, value string) (bool, error)
	Set(ctx context.Context, key string, value interface{}) error
	Get(ctx context.Context, key string) ([]byte, error)
	Del(ctx context.Context, key string) error
	Incr(ctx context.Context, key string) (int64, error)
	Decr(ctx context.Context, key string) (int64, error)
	ExpireAt(ctx context.Context, key string, expr time.Time) (bool, error)
}

type cache struct {
	redisClient *redis.Client
}

func (c cache) SetAdd(ctx context.Context, key string, value string) (bool, error) {
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

func (c cache) SetDel(ctx context.Context, key string, value string) (bool, error) {
	result, err := c.redisClient.SRem(ctx, key, value).Result()
	if err != nil {
		log.Println(err)
		return false, errors.New("occurred an error when deleting value from cache")
	}
	if result == 0 {
		return false, nil
	}

	return true, nil
}

func (c cache) Set(ctx context.Context, key string, value interface{}) error {
	data, marshalErr := json.Marshal(value)
	if marshalErr != nil {
		return marshalErr
	}
	err := c.redisClient.Set(ctx, key, data, 0).Err()
	if err != nil {
		fmt.Println(err)
		return errors.New("occurred an error when setting value to cache")
	}
	return nil
}

func (c cache) Get(ctx context.Context, key string) ([]byte, error) {
	data, err := c.redisClient.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			fmt.Println(err)
			return nil, errors.New("key not found")
		}
		fmt.Println(err)
		return nil, errors.New("occurred an error when getting value from cache")
	}

	return data, nil
}

func (c cache) Del(ctx context.Context, key string) error {
	err := c.redisClient.Del(ctx, key).Err()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			fmt.Println(err)
			return errors.New("key not found")
		}
		fmt.Println(err)
		return errors.New("occurred an error when deleting value from cache")
	}
	return nil
}

func (c cache) Incr(ctx context.Context, key string) (int64, error) {
	result, err := c.redisClient.Incr(ctx, key).Result()
	if err != nil {
		log.Println(err)
		return result, errors.New(fmt.Sprintf("occurred an error when try to increment by the key(%s)", key))
	}

	return result, nil
}

func (c cache) Decr(ctx context.Context, key string) (int64, error) {
	result, err := c.redisClient.Decr(ctx, key).Result()
	if err != nil {
		log.Println(err)
		return result, errors.New(fmt.Sprintf("occurred an error when try to decrement by the key(%s)", key))
	}

	return result, nil
}

func (c cache) ExpireAt(ctx context.Context, key string, expr time.Time) (bool, error) {
	result, err := c.redisClient.ExpireAt(ctx, key, expr).Result()
	if err != nil {
		log.Println(err)
		return result, errors.New(fmt.Sprintf("occurred an error when try to decrement by the key(%s)", key))
	}
	return result, nil
}

func NewCacheClient(client *redis.Client) Cache {
	return &cache{redisClient: client}
}
