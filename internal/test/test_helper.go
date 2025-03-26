package test

import (
	"context"
	"coupon-service/internal/config"
	"testing"
	"time"
)

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

func SetupMySQLForTest(t testing.TB) (*MySQLContainer, context.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	t.Cleanup(func() { cancel() })

	// MySQL 컨테이너 시작
	mysqlContainer, err := NewMySQLContainer(ctx)
	if err != nil {
		t.Fatalf("MySQL 컨테이너 설정 실패: %v", err)
	}

	// 테스트 종료 시 컨테이너 정리
	t.Cleanup(func() {
		cleanCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := mysqlContainer.Cleanup(cleanCtx); err != nil {
			t.Logf("MySQL 컨테이너 정리 실패: %v", err)
		}
	})

	return mysqlContainer, ctx
}
