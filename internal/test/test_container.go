package test

import (
	"context"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
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

func (rc *RedisContainer) FlushAll(ctx context.Context) error {
	return rc.Client.FlushAll(ctx).Err()
}

type MySQLContainer struct {
	Container testcontainers.Container
	URI       string
	Host      string
	Port      string
	DB        *gorm.DB
	Username  string
	Password  string
	Database  string
}

func NewMySQLContainer(ctx context.Context) (*MySQLContainer, error) {
	// 기본 설정
	username := "root"
	password := "test_password"
	database := "coupon_service_test"

	// MySQL 컨테이너 요청 설정
	req := testcontainers.ContainerRequest{
		Image:        "mysql:8.0",
		ExposedPorts: []string{"3306/tcp"},
		Env: map[string]string{
			"MYSQL_ROOT_PASSWORD": password,
			"MYSQL_DATABASE":      database,
		},
		WaitingFor: wait.ForLog("port: 3306  MySQL Community Server - GPL").
			WithStartupTimeout(time.Second * 60),
	}

	// 컨테이너 시작
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("MySQL 컨테이너 시작 실패: %w", err)
	}

	// 호스트 및 포트 정보 가져오기
	host, err := container.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("컨테이너 호스트 가져오기 실패: %w", err)
	}

	port, err := container.MappedPort(ctx, "3306")
	if err != nil {
		return nil, fmt.Errorf("매핑된 포트 가져오기 실패: %w", err)
	}

	portString := port.Port()
	uri := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		username, password, host, portString, database)

	var db *gorm.DB
	var connectErr error
	for i := 0; i < 6; i++ {
		db, connectErr = gorm.Open(mysql.Open(uri), &gorm.Config{})
		if connectErr == nil {
			break
		}
		time.Sleep(5 * time.Second)
	}

	if connectErr != nil {
		_ = container.Terminate(ctx)
		return nil, fmt.Errorf("MySQL 연결 실패: %w", connectErr)
	}

	return &MySQLContainer{
		Container: container,
		URI:       uri,
		Host:      host,
		Port:      portString,
		DB:        db,
		Username:  username,
		Password:  password,
		Database:  database,
	}, nil
}

func (mc *MySQLContainer) Cleanup(ctx context.Context) error {
	if mc.Container != nil {
		if err := mc.Container.Terminate(ctx); err != nil {
			return fmt.Errorf("MySQL 컨테이너 종료 실패: %w", err)
		}
	}

	return nil
}

func (mc *MySQLContainer) MigrateEntities(entities ...interface{}) error {
	return mc.DB.AutoMigrate(entities...)
}
