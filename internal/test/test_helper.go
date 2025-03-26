package test

import (
	"context"
	"coupon-service/internal/config"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

// SetupRedisForTest 테스트를 위한 Redis 컨테이너 설정
func SetupRedisForTest(t *testing.T) (*RedisContainer, context.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(func() { cancel() })

	// Redis 컨테이너 시작
	redisContainer, err := NewRedisContainer(ctx)
	if err != nil {
		t.Fatalf("Redis 컨테이너 설정 실패: %v", err)
	}

	// 테스트 종료 시 컨테이너 정리
	t.Cleanup(func() {
		cleanCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := redisContainer.Cleanup(cleanCtx); err != nil {
			t.Logf("Redis 컨테이너 정리 실패: %v", err)
		}
	})

	// 기존 config의 CacheClient 대체 (테스트에서 사용하기 위함)
	config.CacheClient = redisContainer.Client

	return redisContainer, ctx
}

// CreateRedisClientForTest 테스트용 Redis 클라이언트 생성
func CreateRedisClientForTest(t *testing.T, redisContainer *RedisContainer) *redis.Client {
	// 컨테이너 URI로 새 클라이언트 생성
	client := redis.NewClient(&redis.Options{
		Addr:     redisContainer.URI,
		Password: "",
		DB:       0,
	})

	t.Cleanup(func() {
		if err := client.Close(); err != nil {
			t.Logf("Redis 클라이언트 종료 실패: %v", err)
		}
	})

	return client
}
