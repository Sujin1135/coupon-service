package test

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// RedisContainer Redis 테스트 컨테이너를 위한 래퍼
type RedisContainer struct {
	Container testcontainers.Container
	URI       string
	Port      string
	Client    *redis.Client
}

// NewRedisContainer 새 Redis 컨테이너 시작
func NewRedisContainer(ctx context.Context) (*RedisContainer, error) {
	// Redis 컨테이너 요청 설정
	req := testcontainers.ContainerRequest{
		Image:        "redis:latest",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections"),
	}

	// 컨테이너 시작
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("redis 컨테이너 시작 실패: %w", err)
	}

	// 호스트 및 포트 정보 가져오기
	host, err := container.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("컨테이너 호스트 가져오기 실패: %w", err)
	}

	port, err := container.MappedPort(ctx, "6379")
	if err != nil {
		return nil, fmt.Errorf("매핑된 포트 가져오기 실패: %w", err)
	}

	uri := fmt.Sprintf("%s:%s", host, port.Port())

	// Redis 클라이언트 생성
	client := redis.NewClient(&redis.Options{
		Addr:     uri,
		Password: "", // 기본 비밀번호 없음
		DB:       0,  // 기본 DB
	})

	// 연결 테스트
	_, err = client.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("redis 연결 실패: %w", err)
	}

	return &RedisContainer{
		Container: container,
		URI:       uri,
		Port:      port.Port(),
		Client:    client,
	}, nil
}

// Cleanup 컨테이너 종료 및 리소스 정리
func (rc *RedisContainer) Cleanup(ctx context.Context) error {
	if rc.Client != nil {
		if err := rc.Client.Close(); err != nil {
			return fmt.Errorf("redis 클라이언트 종료 실패: %w", err)
		}
	}

	if rc.Container != nil {
		if err := rc.Container.Terminate(ctx); err != nil {
			return fmt.Errorf("컨테이너 종료 실패: %w", err)
		}
	}

	return nil
}

// FlushAll Redis 데이터베이스의 모든 데이터 삭제
func (rc *RedisContainer) FlushAll(ctx context.Context) error {
	return rc.Client.FlushAll(ctx).Err()
}

// WithTimeout 타임아웃이 있는 컨텍스트 생성
func WithTimeout(seconds int) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), time.Duration(seconds)*time.Second)
}
