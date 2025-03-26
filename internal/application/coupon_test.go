package application

import (
	"context"
	"coupon-service/internal/domain"
	"coupon-service/internal/infrastructure/entity"
	"coupon-service/internal/infrastructure/repository"
	"coupon-service/internal/test"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
	"time"
)

func TestCouponIssueConcurrencyWithContainer(t *testing.T) {
	redisContainer, ctx := test.SetupRedisForTest(t)
	mysqlContainer, ctx := test.SetupMySQLForTest(t)
	couponID := uuid.New().String()
	userStoreKey := "coupon:" + couponID + ":users"
	userID := uuid.New().String()
	couponService := NewCouponService(
		redisContainer.Client,
		repository.NewCouponRepository(mysqlContainer.DB),
		repository.NewIssuedCouponRepository(mysqlContainer.DB),
	)

	mysqlContainer.MigrateEntities(&entity.IssuedCouponEntity{})

	t.Run("동일 사용자 중복 요청 시 false와 에러가 반환 되어야 함", func(t *testing.T) {
		initCache(t, redisContainer, ctx, couponID, 10)

		_ = couponService.IssueCoupon(ctx, couponID, userID)
		err := couponService.IssueCoupon(ctx, couponID, userID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already issued", "중복 발행 오류 메시지 확인")
	})

	t.Run("동일 사용자 중복 요청 시 한개의 쿠폰만 소진 되어야 한다", func(t *testing.T) {
		initCache(t, redisContainer, ctx, couponID, 10)

		_ = couponService.IssueCoupon(ctx, couponID, userID)
		err := couponService.IssueCoupon(ctx, couponID, userID)

		count, err := redisContainer.Client.Get(ctx, genCouponIdKey(couponID)).Int()
		assert.NoError(t, err)
		assert.Equal(t, 9, count, "쿠폰이 1개만 소비되어야 함")
	})

	t.Run("총 발행량 제한 테스트", func(t *testing.T) {
		const (
			numUsers    = 100
			couponLimit = 5
		)

		initCache(t, redisContainer, ctx, couponID, couponLimit)

		var wg sync.WaitGroup
		successCount := 0
		failCount := 0
		var mu sync.Mutex

		for i := 0; i < numUsers; i++ {
			wg.Add(1)
			userID := uuid.New().String()

			go func(uid string) {
				defer wg.Done()

				err := couponService.IssueCoupon(ctx, couponID, uid)

				mu.Lock()
				defer mu.Unlock()

				if err != nil {
					failCount++
				} else {
					successCount++
				}
			}(userID)
		}

		wg.Wait()

		count, err := redisContainer.Client.Get(ctx, genCouponIdKey(couponID)).Int()
		assert.NoError(t, err)
		assert.Equal(t, 0, count, "모든 쿠폰이 소진되어야 함")
		assert.Equal(t, couponLimit, successCount, fmt.Sprintf("정확히 %d명만 쿠폰을 발급받아야 함\n", couponLimit))
		assert.Equal(t, numUsers-couponLimit, failCount, fmt.Sprintf("나머지 %d명은 실패해야 함\n", numUsers-couponLimit))

		setSize, err := redisContainer.Client.SCard(ctx, userStoreKey).Result()
		assert.NoError(t, err)
		assert.Equal(t, int64(couponLimit), setSize, fmt.Sprintf("Set에 %d명의 사용자만 저장되어야 함", couponLimit))
	})

	t.Run("롤백 로직 테스트", func(t *testing.T) {
		initCache(t, redisContainer, ctx, couponID, 0)

		userID := "rollback-test-user"

		err := couponService.IssueCoupon(ctx, couponID, userID)
		assert.Contains(t, err.Error(), "all coupons has been issued", "모든 쿠폰 소진 시 발생하는 에러")

		isMember, err := redisContainer.Client.SIsMember(ctx, userStoreKey, userID).Result()
		assert.NoError(t, err)
		assert.False(t, isMember, "사용자가 Set에 추가되지 않아야 함 (롤백 성공)")

		count, err := redisContainer.Client.Get(ctx, genCouponIdKey(couponID)).Int()
		assert.NoError(t, err)
		assert.Equal(t, 0, count, "카운터는 0을 유지해야 함")
	})
}

func TestCouponIssueWithContainer(t *testing.T) {
	redisContainer, ctx := test.SetupRedisForTest(t)
	mysqlContainer, ctx := test.SetupMySQLForTest(t)
	couponID := uuid.New().String()
	const userID = "same-user-1"
	couponService := NewCouponService(
		redisContainer.Client,
		repository.NewCouponRepository(mysqlContainer.DB),
		repository.NewIssuedCouponRepository(mysqlContainer.DB),
	)

	mysqlContainer.MigrateEntities(&entity.IssuedCouponEntity{})

	t.Run("존재하지 않은 쿠폰 발급 요청 시 에러가 발생한다", func(t *testing.T) {
		initCache(t, redisContainer, ctx, couponID, 10)

		err := couponService.IssueCoupon(ctx, "invalid-coupon", userID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "key not found", "존재하지 않은 쿠폰 발급 요청")
	})

	t.Run("쿠폰 발급 시작 전 요청 시 에러가 발생한다", func(t *testing.T) {
		now := time.Now()
		coupon := domain.Coupon{
			ID:          couponID,
			Name:        "테스트 발급 쿠폰",
			IssueAmount: 10,
			IssuedAt:    now.Add(time.Duration(5) * time.Hour),
			ExpiresAt:   now.Add(time.Duration(10) * time.Hour),
			CreatedAt:   now,
			ModifiedAt:  now,
		}

		initCache(t, redisContainer, ctx, couponID, 10)
		createCouponCache(ctx, redisContainer, couponID, coupon)

		err := couponService.IssueCoupon(ctx, couponID, userID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "coupon issuance has not started yet", "발급 시작 전 요청 시 발생")
	})

	t.Run("쿠폰 발급 종료 후 요청 시 에러가 발생한다", func(t *testing.T) {
		now := time.Now()
		coupon := domain.Coupon{
			ID:          couponID,
			Name:        "테스트 발급 쿠폰",
			IssueAmount: 10,
			IssuedAt:    now.Add(time.Duration(-15) * time.Hour),
			ExpiresAt:   now.Add(time.Duration(-10) * time.Hour),
			CreatedAt:   now,
			ModifiedAt:  now,
		}

		initCache(t, redisContainer, ctx, couponID, 10)
		createCouponCache(ctx, redisContainer, couponID, coupon)

		err := couponService.IssueCoupon(ctx, couponID, userID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "the coupon issuance period has expired", "발급 만료 후 요청 시 발생")
	})

	t.Run("쿠폰 발급 후 발급된 쿠폰이 정상적으로 조회 되어야 한다", func(t *testing.T) {
		initCache(t, redisContainer, ctx, couponID, 10)

		_ = couponService.IssueCoupon(ctx, couponID, userID)

		sut := repository.NewIssuedCouponRepository(mysqlContainer.DB).FindByCouponId(couponID)[0]
		assert.Equal(t, couponID, sut.CouponID)
	})
}

func TestCreateCouponWithContainer(t *testing.T) {
	redisContainer, ctx := test.SetupRedisForTest(t)
	mysqlContainer, ctx := test.SetupMySQLForTest(t)
	couponService := NewCouponService(
		redisContainer.Client,
		repository.NewCouponRepository(mysqlContainer.DB),
		repository.NewIssuedCouponRepository(mysqlContainer.DB),
	)

	mysqlContainer.MigrateEntities(&entity.CouponEntity{})

	t.Run("쿠폰 생성 후 반환되는 에러가 없어야 한다", func(t *testing.T) {
		now := time.Now()
		_, err := couponService.CreateCoupon(
			ctx,
			"쿠폰발급 테스트",
			10,
			now.Add(time.Duration(-5)*time.Hour),
			now.Add(time.Duration(5)*time.Hour),
		)

		assert.NoError(t, err)
	})

	t.Run("쿠폰 생성 후 캐싱 데이터가 조회 되어야 한다", func(t *testing.T) {
		now := time.Now()
		data, _ := couponService.CreateCoupon(
			ctx,
			"쿠폰발급 테스트",
			10,
			now.Add(time.Duration(-5)*time.Hour),
			now.Add(time.Duration(5)*time.Hour),
		)

		var coupon domain.Coupon
		cacheData, err2 := redisContainer.Client.Get(ctx, genCouponCacheKey(data.ID)).Bytes()
		if err2 != nil {
			assert.NoError(t, err2)
		}
		if err3 := json.Unmarshal(cacheData, &coupon); err3 != nil {
			assert.NoError(t, err3)
		}

		assert.Equal(t, data.ID, coupon.ID)
		assert.Equal(t, data.Name, coupon.Name)
		assert.True(t, data.IssuedAt.Equal(coupon.IssuedAt))
		assert.True(t, data.ExpiresAt.Equal(coupon.ExpiresAt))
		assert.Equal(t, data.IssueAmount, coupon.IssueAmount)
	})

	t.Run("쿠폰 생성 후 쿠폰 수량 캐싱 데이터가 조회 되어야 한다", func(t *testing.T) {
		now := time.Now()
		data, _ := couponService.CreateCoupon(
			ctx,
			"쿠폰발급 테스트",
			10,
			now.Add(time.Duration(-5)*time.Hour),
			now.Add(time.Duration(5)*time.Hour),
		)

		amount, err2 := redisContainer.Client.Get(ctx, genCouponIdKey(data.ID)).Int()
		if err2 != nil {
			assert.NoError(t, err2)
		}

		assert.Equal(t, data.IssueAmount, int64(amount))
	})
}

func initCache(
	t *testing.T,
	redisContainer *test.RedisContainer,
	ctx context.Context,
	couponId string,
	couponAmount int,
) {
	_ = redisContainer.FlushAll(ctx)
	err := redisContainer.Client.Set(ctx, genCouponIdKey(couponId), couponAmount, 0).Err()
	if err != nil {
		fmt.Println(err)
		return
	}
	now := time.Now()
	coupon := domain.Coupon{
		ID:          couponId,
		Name:        "테스트 발급 쿠폰",
		IssueAmount: 10,
		IssuedAt:    now.Add(time.Duration(-5) * time.Second),
		ExpiresAt:   now.Add(time.Duration(1) * time.Hour),
		CreatedAt:   now,
		ModifiedAt:  now,
	}
	createCouponCache(ctx, redisContainer, couponId, coupon)
	require.NoError(t, err)
}

func createCouponCache(
	ctx context.Context,
	redisContainer *test.RedisContainer,
	couponId string,
	data domain.Coupon,
) {
	jsonData, err := json.Marshal(data)
	err = redisContainer.Client.Set(ctx, "coupon:"+couponId+":data", jsonData, 0).Err()
	if err != nil {
		fmt.Println(err)
	}
}

func genCouponIdKey(couponID string) string {
	return fmt.Sprintf("coupon:%s:remaining", couponID)
}

func genCouponCacheKey(couponID string) string {
	return fmt.Sprintf("coupon:%s:data", couponID)
}
