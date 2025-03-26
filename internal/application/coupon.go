package application

import (
	"context"
	"coupon-service/internal/infrastructure/cache"
	"errors"
	"log"
)

type CouponService struct {
	cache cache.Cache[int]
}

func NewCouponService() *CouponService {
	return &CouponService{
		cache: cache.NewCacheClient[int](),
	}
}

func (c *CouponService) IssueCoupon(ctx context.Context, couponId string, userId string) (bool, error) {
	userStoreKey := "coupon:" + couponId + ":users"
	couponKey := "coupon:" + couponId + ":remaining"

	added, err := c.cache.SetAdd(ctx, userStoreKey, userId)
	if err != nil {
		return false, err
	}
	if added {
		return false, errors.New("coupon already issued to this user")
	}

	count, err := c.cache.Decr(ctx, couponKey)
	if err != nil {
		return false, err
	}

	if count < 0 {
		_, incrErr := c.cache.Incr(ctx, couponKey)
		if incrErr != nil {
			log.Println(incrErr)
			return false, errors.New("failed to increment coupon amount for recover")
		}

		_, delErr := c.cache.SetDel(ctx, userStoreKey)
		if delErr != nil {
			log.Println(delErr)
			return false, errors.New("failed to delete coupon for recover")
		}
		return false, nil
	}

	// 쿠폰 발급 처리 (DB 저장 등)
	// ...

	return true, nil
}
