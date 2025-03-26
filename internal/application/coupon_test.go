package application

import (
	"context"
	"coupon-service/internal/test"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
)

func TestCouponIssueConcurrencyWithContainer(t *testing.T) {
	redisContainer, ctx := test.SetupRedisForTest(t)
	couponID := "test-coupon-1"
	userStoreKey := "coupon:" + couponID + ":users"
	couponKey := "coupon:" + couponID + ":remaining"
	const userID = "same-user-1"
	couponService := NewCouponService(redisContainer.Client)

	t.Run("동일 사용자 중복 요청 시 false와 에러가 반환 되어야 함", func(t *testing.T) {
		initCache(t, redisContainer, ctx, couponKey, 10)

		_, err := couponService.IssueCoupon(ctx, couponID, userID)
		success2, err := couponService.IssueCoupon(ctx, couponID, userID)

		assert.Error(t, err)
		assert.False(t, success2, "동일 사용자의 두 번째 요청은 실패해야 함")
		assert.Contains(t, err.Error(), "already issued", "중복 발행 오류 메시지 확인")
	})

	t.Run("동일 사용자 중복 요청 시 한개의 쿠폰만 소진 되어야 한다", func(t *testing.T) {
		initCache(t, redisContainer, ctx, couponKey, 10)

		_, err := couponService.IssueCoupon(ctx, couponID, userID)
		_, err = couponService.IssueCoupon(ctx, couponID, userID)

		count, err := redisContainer.Client.Get(ctx, couponKey).Int()
		assert.NoError(t, err)
		assert.Equal(t, 9, count, "쿠폰이 1개만 소비되어야 함")
	})

	t.Run("총 발행량 제한 테스트", func(t *testing.T) {
		const (
			numUsers    = 100
			couponLimit = 5
		)

		initCache(t, redisContainer, ctx, couponKey, couponLimit)

		var wg sync.WaitGroup
		successCount := 0
		failCount := 0
		var mu sync.Mutex

		for i := 0; i < numUsers; i++ {
			wg.Add(1)
			userID := fmt.Sprintf("concurrent-user-%d", i)

			go func(uid string) {
				defer wg.Done()

				success, err := couponService.IssueCoupon(ctx, couponID, uid)

				mu.Lock()
				defer mu.Unlock()

				if success {
					successCount++
				} else {
					failCount++
					t.Logf("사용자 %s: 실패, 오류: %v", uid, err)
				}
			}(userID)
		}

		wg.Wait()

		count, err := redisContainer.Client.Get(ctx, couponKey).Int()
		assert.NoError(t, err)
		assert.Equal(t, 0, count, "모든 쿠폰이 소진되어야 함")
		assert.Equal(t, couponLimit, successCount, fmt.Sprintf("정확히 %d명만 쿠폰을 발급받아야 함\n", couponLimit))
		assert.Equal(t, numUsers-couponLimit, failCount, fmt.Sprintf("나머지 %d명은 실패해야 함\n", numUsers-couponLimit))

		setSize, err := redisContainer.Client.SCard(ctx, userStoreKey).Result()
		assert.NoError(t, err)
		assert.Equal(t, int64(couponLimit), setSize, fmt.Sprintf("Set에 %d명의 사용자만 저장되어야 함", couponLimit))
	})

	t.Run("롤백 로직 테스트", func(t *testing.T) {
		initCache(t, redisContainer, ctx, couponKey, 0)

		userID := "rollback-test-user"

		success, err := couponService.IssueCoupon(ctx, couponID, userID)
		assert.False(t, success, "쿠폰이 없으므로 실패해야 함")
		assert.NoError(t, err, "정상적인 실패이므로 오류가 없어야 함")

		isMember, err := redisContainer.Client.SIsMember(ctx, userStoreKey, userID).Result()
		assert.NoError(t, err)
		assert.False(t, isMember, "사용자가 Set에 추가되지 않아야 함 (롤백 성공)")

		count, err := redisContainer.Client.Get(ctx, couponKey).Int()
		assert.NoError(t, err)
		assert.Equal(t, 0, count, "카운터는 0을 유지해야 함")
	})
}

func initCache(
	t *testing.T,
	redisContainer *test.RedisContainer,
	ctx context.Context,
	couponKey string,
	couponAmount int,
) {
	_ = redisContainer.FlushAll(ctx)
	err := redisContainer.Client.Set(ctx, couponKey, couponAmount, 0).Err()
	require.NoError(t, err)
}
