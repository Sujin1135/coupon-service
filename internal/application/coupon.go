package application

import (
	"context"
	"coupon-service/internal/domain"
	"coupon-service/internal/infrastructure/cache"
	"coupon-service/internal/infrastructure/repository"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"log"
	"time"
)

type CouponService struct {
	cache            cache.Cache
	couponRepository *repository.CouponRepository
}

func NewCouponService(
	cacheClient *redis.Client,
	couponRepository *repository.CouponRepository,
) *CouponService {
	return &CouponService{
		cache:            cache.NewCacheClient(cacheClient),
		couponRepository: couponRepository,
	}
}

func (c *CouponService) IssueCoupon(ctx context.Context, couponId string, userId string) (bool, error) {
	dataKey := "coupon:" + couponId + ":data"
	userStoreKey := "coupon:" + couponId + ":users"
	couponKey := "coupon:" + couponId + ":remaining"
	now := time.Now()

	err := c.validateCouponEvent(ctx, dataKey, now)
	if err != nil {
		return false, err
	}

	err = c.controlConcurrent(ctx, userStoreKey, userId, couponKey)
	if err != nil {
		return false, err
	}

	_, err = c.couponRepository.FindOne(couponId)
	if err != nil {
		fmt.Println(fmt.Sprintf("coupon was not found by id(%s)", couponId))
	}

	return true, nil
}

func (c *CouponService) validateCouponEvent(ctx context.Context, dataKey string, now time.Time) error {
	var coupon domain.Coupon
	data, err := c.cache.Get(ctx, dataKey)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, &coupon); err != nil {
		return err
	}
	if coupon.IssuedAt.After(now) {
		return errors.New("coupon issuance has not started yet")
	}
	if coupon.ExpiresAt.Before(now) {
		return errors.New("the coupon issuance period has expired")
	}
	return nil
}

func (c *CouponService) controlConcurrent(ctx context.Context, userStoreKey string, userId string, couponKey string) error {
	added, err := c.cache.SetAdd(ctx, userStoreKey, userId)
	if err != nil {
		return err
	}
	if added == false {
		return errors.New("coupon already issued to this user")
	}

	count, err := c.cache.Decr(ctx, couponKey)
	if err != nil {
		return err
	}

	if count < 0 {
		_, incrErr := c.cache.Incr(ctx, couponKey)
		if incrErr != nil {
			log.Println(incrErr)
			return errors.New("failed to increment coupon amount for recover")
		}

		_, delErr := c.cache.SetDel(ctx, userStoreKey, userId)
		if delErr != nil {
			log.Println(delErr)
			return errors.New("failed to delete coupon for recover")
		}
		return errors.New("all coupons has been issued")
	}
	return nil
}
